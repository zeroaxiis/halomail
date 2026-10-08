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

		// TODO: Call app.Service to upgrade/downgrade the org!
		// For now, we will just log it.
		// h.app.ProcessSubscription(r.Context(), orgID, tier, sub.CustomerID, sub.ID, sub.Status, time.Unix(sub.CurrentEnd, 0))

		// TODO: Send Invoice via Resend if event == "subscription.charged"

		httpx.JSON(w, http.StatusOK, map[string]bool{"success": true})
	})
}

func verifyRazorpaySignature(body []byte, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}
