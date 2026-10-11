package rpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/aashishrajdev/halomail/services/identity/internal/app"
	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
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
