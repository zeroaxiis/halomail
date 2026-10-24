package postgres

import (
	"context"

	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
)

// The same account lock as ApplyPayment makes capture/refund races deterministic.
// Refund IDs are immutable and unique, so repeated snapshots cannot inflate totals.
func (r *Users) ApplyRefunds(ctx context.Context, orgID, orderID, paymentID, currency string, paymentAmount int64, refunds []domain.BillingRefund) (domain.RefundResult, error) {
	var outcome domain.RefundResult
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return outcome, err
	}
	defer tx.Rollback(ctx)
	var locked string
	if err := tx.QueryRow(ctx, "SELECT id FROM orgs WHERE id=$1 FOR UPDATE", orgID).Scan(&locked); err != nil {
		return outcome, err
	}
	var conflict bool
	err = tx.QueryRow(ctx, `SELECT EXISTS (
 SELECT 1 FROM billing_payments WHERE (payment_id=$2 OR order_id=$3)
 AND (org_id<>$1 OR payment_id<>$2 OR order_id<>$3)
 ) OR EXISTS (
 SELECT 1 FROM billing_refunds WHERE (payment_id=$2 OR order_id=$3)
 AND (org_id<>$1 OR payment_id<>$2 OR order_id<>$3 OR currency<>$4 OR payment_amount<>$5)
 )`, orgID, paymentID, orderID, currency, paymentAmount).Scan(&conflict)
	if err != nil {
		return outcome, err
	}
	if conflict {
		return outcome, errs.Conflict("refund does not match the recorded purchase")
	}
	for _, refund := range refunds {
		result, err := tx.Exec(ctx, `INSERT INTO billing_refunds
 (refund_id,payment_id,order_id,org_id,payment_amount,amount,currency,created_at)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(refund_id) DO NOTHING`,
			refund.ID, paymentID, orderID, orgID, paymentAmount, refund.Amount, currency, refund.CreatedAt)
		if err != nil {
			return outcome, err
		}
		if result.RowsAffected() == 0 {
			var matches bool
			err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM billing_refunds
 WHERE refund_id=$1 AND payment_id=$2 AND order_id=$3 AND org_id=$4
 AND payment_amount=$5 AND amount=$6 AND currency=$7 AND created_at=$8)`,
				refund.ID, paymentID, orderID, orgID, paymentAmount, refund.Amount, currency, refund.CreatedAt).Scan(&matches)
			if err != nil {
				return outcome, err
			}
			if !matches {
				return outcome, errs.Conflict("refund ID already belongs to another refund")
			}
		}
	}
	var total int64
	if err := tx.QueryRow(ctx, "SELECT COALESCE(SUM(amount),0) FROM billing_refunds WHERE payment_id=$1", paymentID).Scan(&total); err != nil {
		return outcome, err
	}
	if total > paymentAmount {
		return outcome, errs.Conflict("completed refunds exceed the payment amount")
	}
	outcome.FullyRefunded = total == paymentAmount
	if outcome.FullyRefunded {
		// Preserve billing_paid_at: an older, delayed purchase must not restore access.
		result, err := tx.Exec(ctx, `UPDATE orgs SET tier='free', rzp_subscription_id=NULL, billing_cycle_end=NULL
 WHERE id=$1 AND rzp_subscription_id=$2`, orgID, orderID)
		if err != nil {
			return outcome, err
		}
		outcome.Downgraded = result.RowsAffected() > 0
	}
	return outcome, tx.Commit(ctx)
}
