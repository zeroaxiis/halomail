package usage

import (
	"context"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/jackc/pgx/v5"
	"testing"
	"time"
)

func TestWindow(test *testing.T) {
	policy := Policy{Forms: 25, Meetings: 10, Monthly: true}
	now := time.Date(2026, 1, 1, 1, 0, 0, 0, time.FixedZone("IST", 19800))
	limit, period, reset := policy.Window("forms", now)
	if limit != 25 || period.Format("2006-01-02") != "2025-12-01" || reset == nil || reset.Format("2006-01-02") != "2026-01-01" {
		test.Fatalf("wrong UTC monthly window: %d %v %v", limit, period, reset)
	}
	policy.Monthly = false
	limit, period, reset = policy.Window("meetings", now)
	if limit != 10 || period.Year() != 1970 || reset != nil {
		test.Fatal("wrong lifetime window")
	}
}

type resultRow struct{ err error }

func (row resultRow) Scan(...any) error { return row.err }

type quotaTransaction struct {
	pgx.Tx
	row pgx.Row
}

func (transaction quotaTransaction) QueryRow(context.Context, string, ...any) pgx.Row {
	return transaction.row
}

func TestConsumeExhaustedOrDisabled(test *testing.T) {
	for _, limit := range []int{0, 25} {
		err := Consume(context.Background(), quotaTransaction{row: resultRow{err: pgx.ErrNoRows}}, Policy{Forms: limit}, "owner", "forms")
		if errs.KindOf(err) != errs.KindRateLimited {
			test.Fatalf("allowance failure not enforced: %v", err)
		}
	}
}
