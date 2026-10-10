package rpc

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/aashishrajdev/halomail/services/shared/httpx"
)

// RazorpayWebhookPayload is a partial struct for the events we care about.
type RazorpayWebhookPayload struct {
	Event   string `json:"event"`
	Payload struct {
		Subscription struct {
			Entity struct {
				ID         string `json:"id"`
				CustomerID string `json:"customer_id"`
				Status     string `json:"status"`
				Notes      struct {
					OrgID string `json:"org_id"`
					Tier  string `json:"tier"`
				} `json:"notes"`
				CurrentEnd int64 `json:"current_end"`
			} `json:"entity"`
		} `json:"subscription"`
		Order struct {
			Entity struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Notes  struct {
					OrgID string `json:"org_id"`
					Tier  string `json:"tier"`
				} `json:"notes"`
			} `json:"entity"`
		} `json:"order"`
		Payment struct {
			Entity struct {
				ID     string `json:"id"`
				Status string `json:"status"`
				Email  string `json:"email"`
				Notes  struct {
					OrgID string `json:"org_id"`
					Tier  string `json:"tier"`
				} `json:"notes"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

func (h *Handlers) MountBilling(mux *http.ServeMux, webhookSecret string, internalSecret string) {
	mux.HandleFunc("POST /v1/billing/razorpay/webhook", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// Verify Signature (Bypass if requested internally by Next.js using internal secret)
		isInternal := r.Header.Get("X-Internal-Secret") == internalSecret && internalSecret != ""
		if !isInternal {
			signature := r.Header.Get("X-Razorpay-Signature")
			if !verifyRazorpaySignature(bodyBytes, signature, webhookSecret) {
				http.Error(w, "Invalid signature", http.StatusUnauthorized)
				return
			}
		}

		var payload RazorpayWebhookPayload
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			httpx.Error(w, err)
			return
		}

		var orgID, tier, customerID, entityID, status string
		var cycleEnd time.Time

		if payload.Event == "subscription.charged" || payload.Event == "subscription.halted" || payload.Event == "subscription.cancelled" {
			sub := payload.Payload.Subscription.Entity
			orgID = sub.Notes.OrgID
			tier = sub.Notes.Tier
			customerID = sub.CustomerID
			entityID = sub.ID
			status = sub.Status
			cycleEnd = time.Unix(sub.CurrentEnd, 0)
		} else if payload.Event == "order.paid" || payload.Event == "order.refunded" {
			order := payload.Payload.Order.Entity
			orgID = order.Notes.OrgID
			tier = order.Notes.Tier
			customerID = ""
			entityID = order.ID
			status = order.Status
			if payload.Event == "order.refunded" {
				status = "refunded"
			}
			cycleEnd = time.Now().AddDate(0, 1, 0) // Default 1 month for orders
		} else if payload.Event == "payment.captured" || payload.Event == "payment.refunded" {
			payment := payload.Payload.Payment.Entity
			orgID = payment.Notes.OrgID
			tier = payment.Notes.Tier
			customerID = ""
			entityID = payment.ID
			status = payment.Status
			if status == "" {
				status = "captured"
			}
			if payload.Event == "payment.refunded" {
				status = "refunded"
			}
			cycleEnd = time.Now().AddDate(0, 1, 0) // Default 1 month for payments
		}

		if orgID == "" {
			// If org_id is missing, we can't upgrade anyone.
			httpx.JSON(w, http.StatusOK, map[string]bool{"success": true, "ignored": true})
			return
		}

		if err := h.app.ProcessSubscription(r.Context(), orgID, customerID, entityID, tier, status, cycleEnd); err != nil {
			httpx.Error(w, err)
			return
		}

		if payload.Event == "subscription.charged" || payload.Event == "order.paid" {
			email := payload.Payload.Payment.Entity.Email
			if email != "" {
				go sendInvoiceEmail(email, tier) // run async
			}
		}

		httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
	})
}

func sendInvoiceEmail(to, tier string) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return
	}

	body := map[string]any{
		"from":    "HaloMail <no-reply@email.zeroaxiis.tech>",
		"to":      []string{to},
		"subject": "Payment Receipt & Invoice - HaloMail",
		"html":    fmt.Sprintf("<h1>Payment Successful</h1><p>Thank you for subscribing to the <strong>%s</strong> plan!</p><p>Your features are now unlocked.</p><p>HaloMail Team</p>", tier),
	}
	bodyBytes, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	client.Do(req)
}

func verifyRazorpaySignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
