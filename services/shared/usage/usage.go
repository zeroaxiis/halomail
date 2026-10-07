// Package usage enforces the free-tier allowance per owner and feature.
package usage

import (
	"context"
	"errors"
	"time"

	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Policy is the free-tier allowance. With Monthly set the counters reset on
// the first of each month (UTC); otherwise they are lifetime limits.
type Policy struct {
	Forms    int  `env:"FREE_FORM_LIMIT" envDefault:"25"`
	Meetings int  `env:"FREE_MEETING_LIMIT" envDefault:"10"`
	Monthly  bool `env:"FREE_LIMITS_MONTHLY" envDefault:"true"`
}

// Stats is the usage of one feature in the current window.
type Stats struct {
	Used      int        `json:"used"`
	Limit     int        `json:"limit"`
	Remaining int        `json:"remaining"`
	ResetsAt  *time.Time `json:"resetsAt"`
}

// Window returns the limit for feature, the start of the current period and,
// for monthly policies, when the allowance resets.
func (policy Policy) Window(feature string, now time.Time) (int, time.Time, *time.Time) {
	limit := policy.Forms
	if feature == "meetings" {
		limit = policy.Meetings
	}
	period := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	var reset *time.Time
	if policy.Monthly {
		current := now.UTC()
		period = time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, time.UTC)
		next := period.AddDate(0, 1, 0)
		reset = &next
	}
	return limit, period, reset
}

// Consume spends one unit of the allowance inside tx, returning a RateLimited
// error once the limit is reached.
func Consume(ctx context.Context, tx pgx.Tx, policy Policy, ownerID, feature string) error {
	limit, period, _ := policy.Window(feature, time.Now())
	if limit <= 0 {
		return errs.RateLimited("free %s allowance exhausted", feature)
	}
	var used int
	err := tx.QueryRow(ctx, `INSERT INTO feature_usage (owner_id,feature,period,used) VALUES ($1,$2,$3,1)
	ON CONFLICT (owner_id,feature,period) DO UPDATE SET used=feature_usage.used+1
	WHERE feature_usage.used < $4 RETURNING used`, ownerID, feature, period, limit).Scan(&used)
	if errors.Is(err, pgx.ErrNoRows) {
		return errs.RateLimited("free %s allowance exhausted", feature)
	}
	return err
}

// Read reports the usage for the current window without consuming any.
func Read(ctx context.Context, pool *pgxpool.Pool, policy Policy, ownerID, feature string) (Stats, error) {
	limit, period, reset := policy.Window(feature, time.Now())
	stats := Stats{Limit: limit, ResetsAt: reset}
	err := pool.QueryRow(ctx, `SELECT used FROM feature_usage WHERE owner_id=$1 AND feature=$2 AND period=$3`, ownerID, feature, period).Scan(&stats.Used)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return stats, err
	}
	stats.Remaining = max(0, limit-stats.Used)
	return stats, nil
}
