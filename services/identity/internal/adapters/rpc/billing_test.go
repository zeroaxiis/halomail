package rpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/zeroaxiis/halomail/services/identity/internal/app"
	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type billingRepo struct {
	app.UserRepo
	calls              int
	org, tier, payment string
	end                time.Time
	err                error
	refundCalls        int
	refunds            []domain.BillingRefund
	refundResult       domain.RefundResult
}

func (r *billingRepo) GetUserByID(context.Context, string) (*domain.User, error) {
	return &domain.User{ID: "user", OrgID: "org"}, nil
}
func (r *billingRepo) ApplyPayment(_ context.Context, org, order, payment, tier string, paid, end time.Time) error {
	r.calls++
	r.org = org
	r.tier = tier
	r.payment = payment
	r.end = end
	return r.err
}
func (r *billingRepo) ApplyRefunds(_ context.Context, org, order, payment, currency string, amount int64, refunds []domain.BillingRefund) (domain.RefundResult, error) {
	r.refundCalls++
	r.org = org
	r.payment = payment
	r.refunds = refunds
	return r.refundResult, r.err
}

func refundPayload(amounts ...int64) RazorpayWebhookPayload {
	var payload RazorpayWebhookPayload
	json.Unmarshal([]byte(paidPayload("monthly")), &payload)
	payload.Event = "refund.processed"
	payload.Payload.Order.Entity.Notes.Product = "halomail"
	for i, amount := range amounts {
		payload.Payload.Refunds.Items = append(payload.Payload.Refunds.Items, refundEntity{
			ID: "rfnd_" + strings.Repeat("x", i+1), PaymentID: "pay_1", Status: "processed",
			Amount: amount, Currency: "INR", CreatedAt: 1791676900,
		})
	}
	return payload
}

func TestBillingRefundUsesCanonicalAccountAndCompletedAmounts(t *testing.T) {
	for _, full := range []bool{false, true} {
		payload := refundPayload(40)
		if full {
			payload = refundPayload(40, 60)
		}
		body, _ := json.Marshal(payload)
		repo := &billingRepo{refundResult: domain.RefundResult{FullyRefunded: full, Downgraded: full}}
		response := billingRequest(repo, string(body), "secret")
		if response.Code != 200 || repo.refundCalls != 1 || repo.calls != 0 || repo.org != "org" || repo.payment != "pay_1" {
			t.Fatalf("refund was not processed: %d %s %+v", response.Code, response.Body.String(), repo)
		}
		var result map[string]any
		json.Unmarshal(response.Body.Bytes(), &result)
		if result["fullyRefunded"] != full || result["downgraded"] != full {
			t.Fatalf("wrong refund response: %v", result)
		}
		if repo.refunds[0].Amount != 40 {
			t.Fatal("wrong refund amount")
		}
	}
}

func TestBillingRefundRejectsInvalidAndConflictingDetails(t *testing.T) {
	for _, mutate := range []func(*RazorpayWebhookPayload){
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].Status = "pending" },
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].Status = "failed" },
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].PaymentID = "pay_other" },
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].Currency = "USD" },
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].Amount = 0 },
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].Amount = 101 },
		func(p *RazorpayWebhookPayload) { p.Payload.Refunds.Items[0].CreatedAt = 0 },
		func(p *RazorpayWebhookPayload) {
			p.Payload.Refunds.Items = append(p.Payload.Refunds.Items, p.Payload.Refunds.Items[0])
		},
		func(p *RazorpayWebhookPayload) { p.Payload.Payment.Entity.OrderID = "order_other" },
		func(p *RazorpayWebhookPayload) { p.Payload.Payment.Entity.Status = "authorized" },
		func(p *RazorpayWebhookPayload) { p.Payload.Order.Entity.Notes.Product = "other" },
		func(p *RazorpayWebhookPayload) { p.Payload.Order.Entity.Notes.OrgID = "other_org" },
	} {
		payload := refundPayload(100)
		mutate(&payload)
		body, _ := json.Marshal(payload)
		repo := &billingRepo{}
		response := billingRequest(repo, string(body), "secret")
		if response.Code < 400 || repo.refundCalls != 0 {
			t.Fatalf("invalid refund accepted: %d %s", response.Code, body)
		}
	}
	for _, event := range []string{"refund.created", "refund.failed"} {
		payload := refundPayload(100)
		payload.Event = event
		body, _ := json.Marshal(payload)
		repo := &billingRepo{}
		if response := billingRequest(repo, string(body), "secret"); response.Code != 200 || repo.refundCalls != 0 {
			t.Fatal("unprocessed refund applied")
		}
	}
}

func TestBillingRefundFailuresAreNotAcknowledged(t *testing.T) {
	payload := refundPayload(100)
	body, _ := json.Marshal(payload)
	if response := billingRequest(&billingRepo{}, string(body), ""); response.Code != 401 {
		t.Fatal("unsigned refund accepted")
	}
	if response := billingRequest(&billingRepo{err: errors.New("database down")}, string(body), "secret"); response.Code < 500 {
		t.Fatal("database failure acknowledged")
	}
}
func billingRequest(repo *billingRepo, raw, secret string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	NewHandlers(app.New(app.Repos{Users: repo}, app.Config{}, nil)).MountBilling(mux, secret)
	req := httptest.NewRequest("POST", "/v1/billing/razorpay/webhook", strings.NewReader(raw))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(raw))
	req.Header.Set("X-Razorpay-Signature", hex.EncodeToString(mac.Sum(nil)))
	req.Header.Set("X-Internal-Secret", "otp-secret")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}
func paidPayload(cycle string) string {
	return `{"event":"order.paid","payload":{"order":{"entity":{"id":"order_1","status":"paid","amount":100,"currency":"INR","notes":{"user_id":"user","tier":"business","billing_cycle":"` + cycle + `"}}},"payment":{"entity":{"id":"pay_1","order_id":"order_1","status":"captured","amount":100,"currency":"INR","created_at":1791676800,"notes":[]}}}}`
}
func TestBillingResolvesLegacyAccountAndYearlyCycle(t *testing.T) {
	repo := &billingRepo{}
	res := billingRequest(repo, paidPayload("yearly"), "secret")
	if res.Code != 200 {
		t.Fatalf("%d: %s", res.Code, res.Body.String())
	}
	if repo.org != "org" || repo.tier != "business" || repo.payment != "pay_1" || repo.end != billingCycleEnd(time.Unix(1791676800, 0).UTC(), 12) {
		t.Fatalf("wrong activation: %+v", repo)
	}
	var body map[string]any
	json.Unmarshal(res.Body.Bytes(), &body)
	if body["orgId"] != "org" {
		t.Fatal("canonical account missing")
	}
}
func TestBillingRejectsInvalidPaymentsAndConfiguration(t *testing.T) {
	for _, raw := range []string{
		strings.Replace(paidPayload("monthly"), "captured", "authorized", 1),
		strings.Replace(paidPayload("monthly"), "business", "invalid", 1),
		strings.Replace(paidPayload("monthly"), "\"user_id\":\"user\"", "\"org_id\":\"org\"", 1),
		paidPayload("weekly"),
	} {
		repo := &billingRepo{}
		res := billingRequest(repo, raw, "secret")
		if res.Code < 400 || repo.calls != 0 {
			t.Fatalf("accepted %s: %d", raw, res.Code)
		}
	}
	if res := billingRequest(&billingRepo{}, paidPayload("monthly"), ""); res.Code != 401 {
		t.Fatal("empty signing secret accepted")
	}
	repo := &billingRepo{err: errors.New("database down")}
	if res := billingRequest(repo, paidPayload("monthly"), "secret"); res.Code < 500 {
		t.Fatal("database failure acknowledged")
	}
}
func TestBillingCycleClampsCalendarEnds(t *testing.T) {
	for _, tt := range []struct {
		start  string
		months int
		want   string
	}{
		{"2026-01-31", 1, "2026-02-28"}, {"2024-02-29", 12, "2025-02-28"}, {"2026-10-11", 12, "2027-10-11"},
	} {
		start, _ := time.Parse("2006-01-02", tt.start)
		if got := billingCycleEnd(start, tt.months).Format("2006-01-02"); got != tt.want {
			t.Fatalf("%s: %s", tt.start, got)
		}
	}
}
