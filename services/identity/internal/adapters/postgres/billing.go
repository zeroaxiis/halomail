package postgres

import (
	"context"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"time"
)

// The account lock serializes simultaneous callbacks. The ledger deduplicates both
// Razorpay events and checkout retries; older payments cannot overwrite a newer plan.
func (r *Users) ApplyPayment(ctx context.Context, orgID, orderID, paymentID, tier string, paidAt, cycleEnd time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var exists string
	if err := tx.QueryRow(ctx, "SELECT id FROM orgs WHERE id=$1 FOR UPDATE", orgID).Scan(&exists); err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `INSERT INTO billing_payments (payment_id,order_id,org_id,tier,paid_at,cycle_end)
 VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, paymentID, orderID, orgID, tier, paidAt, cycleEnd)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		var matches bool
		err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM billing_payments WHERE payment_id=$1 AND order_id=$2 AND org_id=$3 AND tier=$4)`, paymentID, orderID, orgID, tier).Scan(&matches)
		if err != nil {
			return err
		}
		if !matches {
			return errs.Conflict("payment already belongs to another purchase")
		}
		return tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `UPDATE orgs SET tier=$2, rzp_subscription_id=$3, billing_cycle_end=$4, billing_paid_at=$5
 WHERE id=$1 AND (billing_paid_at IS NULL OR billing_paid_at <= $5)
 AND NOT EXISTS (
   SELECT 1 FROM billing_refunds WHERE payment_id=$6 OR order_id=$3
   GROUP BY payment_id HAVING SUM(amount) >= MAX(payment_amount)
 )`, orgID, tier, orderID, cycleEnd, paidAt, paymentID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
