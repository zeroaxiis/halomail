package rpc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/aashishrajdev/halomail/services/shared/httpx"
)

type billingNotes struct {
	Product      string `json:"product"`
	UserID       string `json:"user_id"`
	OrgID        string `json:"org_id"`
	Tier         string `json:"tier"`
	BillingCycle string `json:"billing_cycle"`
}

// Razorpay represents empty notes as [] in payment payloads.
func (n *billingNotes) UnmarshalJSON(data []byte) error {
	if string(data) == "[]" || string(data) == "null" {
		return nil
	}
	type plain billingNotes
	return json.Unmarshal(data, (*plain)(n))
}

type billingEntity struct {
	ID        string       `json:"id"`
	OrderID   string       `json:"order_id"`
	Status    string       `json:"status"`
	Amount    int64        `json:"amount"`
	Currency  string       `json:"currency"`
	CreatedAt int64        `json:"created_at"`
	Notes     billingNotes `json:"notes"`
}
type refundEntity struct {
	ID        string `json:"id"`
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	CreatedAt int64  `json:"created_at"`
}
type RazorpayWebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Order struct {
			Entity billingEntity `json:"entity"`
		} `json:"order"`
		Payment struct {
			Entity billingEntity `json:"entity"`
		} `json:"payment"`
		Refunds struct {
			Items []refundEntity `json:"items"`
		} `json:"refunds"`
	} `json:"payload"`
}

func (h *Handlers) MountBilling(mux *http.ServeMux, webhookSecret string) {
	mux.HandleFunc("POST /v1/billing/razorpay/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}
		// Billing always requires its own signature, never the OTP delivery secret.
		if !verifyRazorpaySignature(body, r.Header.Get("X-Razorpay-Signature"), webhookSecret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
		var payload RazorpayWebhookPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if payload.Event == "refund.processed" {
			h.processBillingRefunds(w, r, payload)
			return
		}
		if payload.Event != "order.paid" && payload.Event != "payment.captured" {
			httpx.JSON(w, http.StatusOK, map[string]bool{"ignored": true})
			return
		}
		order, payment := payload.Payload.Order.Entity, payload.Payload.Payment.Entity
		notes := order.Notes
		if order.ID == "" || payment.ID == "" || payment.OrderID != order.ID || order.Status != "paid" || payment.Status != "captured" ||
			order.Amount <= 0 || order.Amount != payment.Amount || order.Currency == "" || order.Currency != payment.Currency ||
			notes.UserID == "" || (notes.Tier != "pro" && notes.Tier != "business") || payment.CreatedAt <= 0 {
			http.Error(w, "Missing or invalid paid order details", http.StatusUnprocessableEntity)
			return
		}
		months := 1
		switch notes.BillingCycle {
		case "", "monthly":
		case "yearly":
			months = 12
		default:
			http.Error(w, "Invalid billing cycle", http.StatusUnprocessableEntity)
			return
		}
		paidAt := time.Unix(payment.CreatedAt, 0).UTC()
		cycleEnd := billingCycleEnd(paidAt, months)
		user, err := h.app.ProcessPayment(r.Context(), notes.UserID, notes.OrgID, order.ID, payment.ID, notes.Tier, paidAt, cycleEnd)
		if err != nil {
			httpx.Error(w, err)
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]any{"success": true, "orgId": user.OrgID, "userId": user.ID})
	})
}

// Clamp month ends (January 31 -> February 28), including leap-year annual plans.
func billingCycleEnd(start time.Time, months int) time.Time {
	first := time.Date(start.Year(), start.Month()+time.Month(months), 1, start.Hour(), start.Minute(), start.Second(), 0, time.UTC)
	lastDay := first.AddDate(0, 1, -1).Day()
	return first.AddDate(0, 0, min(start.Day(), lastDay)-1)
}
func verifyRazorpaySignature(body []byte, signature, secret string) bool {
	if secret == "" {
		return false
	}
	decoded, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), decoded)
}
