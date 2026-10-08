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
		Payment struct {
			Entity struct {
				Email string `json:"email"`
			} `json:"entity"`
		} `json:"payment"`
	} `json:"payload"`
}

func (h *Handlers) MountBilling(mux *http.ServeMux, webhookSecret string) {
	mux.HandleFunc("POST /v1/billing/razorpay/webhook", func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// Verify Signature
		signature := r.Header.Get("X-Razorpay-Signature")
		if !verifyRazorpaySignature(bodyBytes, signature, webhookSecret) {
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}

		var payload RazorpayWebhookPayload
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			httpx.Error(w, err)
			return
		}

		// We only care about subscription events
		sub := payload.Payload.Subscription.Entity
		if sub.ID == "" {
			httpx.JSON(w, http.StatusOK, map[string]bool{"success": true, "ignored": true})
			return
		}

		orgID := sub.Notes.OrgID
		tier := sub.Notes.Tier
		if orgID == "" {
			// If org_id is missing, we can't upgrade anyone.
			httpx.JSON(w, http.StatusOK, map[string]bool{"success": true, "ignored": true})
			return
		}

		if err := h.app.ProcessSubscription(r.Context(), orgID, sub.CustomerID, sub.ID, tier, sub.Status, time.Unix(sub.CurrentEnd, 0)); err != nil {
			httpx.Error(w, err)
			return
		}

		if payload.Event == "subscription.charged" {
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
