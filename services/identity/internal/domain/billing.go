package domain

import "time"

// BillingRefund is a completed provider refund. Pending and failed refunds never enter the ledger.
type BillingRefund struct {
	ID        string
	Amount    int64
	CreatedAt time.Time
}

type RefundResult struct {
	FullyRefunded bool
	Downgraded    bool
}
