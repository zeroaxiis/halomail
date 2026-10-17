package rpc

import (
	"net/http"
	"strings"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/httpx"
)

// The web adapter fetches and signs the order, payment, and completed refund list.
func (h *Handlers) processBillingRefunds(w http.ResponseWriter, r *http.Request, payload RazorpayWebhookPayload) {
	order, payment := payload.Payload.Order.Entity, payload.Payload.Payment.Entity
	notes := order.Notes
	if order.ID == "" || payment.ID == "" || payment.OrderID != order.ID || order.Status != "paid" ||
		(payment.Status != "captured" && payment.Status != "refunded") || order.Amount <= 0 || order.Amount != payment.Amount ||
		order.Currency == "" || order.Currency != payment.Currency || notes.Product != "halomail" || notes.UserID == "" ||
		(notes.Tier != "pro" && notes.Tier != "business") || len(payload.Payload.Refunds.Items) == 0 {
		http.Error(w, "Missing or invalid refunded order details", http.StatusUnprocessableEntity)
		return
	}
	refunds := make([]domain.BillingRefund, 0, len(payload.Payload.Refunds.Items))
	seen := make(map[string]bool)
	var total int64
	for _, refund := range payload.Payload.Refunds.Items {
		if !strings.HasPrefix(refund.ID, "rfnd_") || len(refund.ID) <= 5 || seen[refund.ID] || refund.PaymentID != payment.ID ||
			refund.Status != "processed" || refund.Currency != payment.Currency || refund.Amount <= 0 ||
			refund.Amount > payment.Amount-total || refund.CreatedAt <= 0 {
			http.Error(w, "Invalid completed refund details", http.StatusUnprocessableEntity)
			return
		}
		seen[refund.ID] = true
		total += refund.Amount
		refunds = append(refunds, domain.BillingRefund{ID: refund.ID, Amount: refund.Amount, CreatedAt: time.Unix(refund.CreatedAt, 0).UTC()})
	}
	user, result, err := h.app.ProcessRefunds(r.Context(), notes.UserID, notes.OrgID, order.ID, payment.ID, payment.Currency, payment.Amount, refunds)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"success": true, "orgId": user.OrgID, "userId": user.ID,
		"fullyRefunded": result.FullyRefunded, "downgraded": result.Downgraded,
	})
}
