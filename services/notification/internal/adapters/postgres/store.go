// Package postgres implements the notification repository ports over PostgreSQL.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zeroaxiis/halomail/services/notification/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
)

// ---- Webhooks ------------------------------------------------------------

type Webhooks struct{ pool *pgxpool.Pool }

func NewWebhooks(pool *pgxpool.Pool) *Webhooks { return &Webhooks{pool: pool} }

const whCols = `id, owner_id, url, events, secret, active, created_at`

func (r *Webhooks) Create(ctx context.Context, w *domain.Webhook) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhooks (`+whCols+`) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		w.ID, w.OwnerID, w.URL, w.Events, w.Secret, w.Active, w.CreatedAt)
	return err
}

func (r *Webhooks) GetByID(ctx context.Context, id string) (*domain.Webhook, error) {
	return scanWebhook(r.pool.QueryRow(ctx, `SELECT `+whCols+` FROM webhooks WHERE id=$1`, id))
}

func (r *Webhooks) ListByOwner(ctx context.Context, ownerID string) ([]domain.Webhook, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+whCols+` FROM webhooks WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectWebhooks(rows)
}

func (r *Webhooks) ListSubscribed(ctx context.Context, ownerID, event string) ([]domain.Webhook, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+whCols+` FROM webhooks WHERE owner_id=$1 AND active AND $2 = ANY(events)`, ownerID, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectWebhooks(rows)
}

func (r *Webhooks) UpdateSecret(ctx context.Context, id, ownerID, secret string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE webhooks SET secret=$3 WHERE id=$1 AND owner_id=$2`, id, ownerID, secret)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("webhook not found")
	}
	return nil
}

func (r *Webhooks) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM webhooks WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("webhook not found")
	}
	return nil
}

func collectWebhooks(rows pgx.Rows) ([]domain.Webhook, error) {
	var out []domain.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	return out, rows.Err()
}

func scanWebhook(row pgx.Row) (*domain.Webhook, error) {
	var w domain.Webhook
	if err := row.Scan(&w.ID, &w.OwnerID, &w.URL, &w.Events, &w.Secret, &w.Active, &w.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("webhook not found")
		}
		return nil, err
	}
	return &w, nil
}

// ---- Deliveries ----------------------------------------------------------

type Deliveries struct{ pool *pgxpool.Pool }

func NewDeliveries(pool *pgxpool.Pool) *Deliveries { return &Deliveries{pool: pool} }

const dCols = `id, webhook_id, owner_id, event, status, response_code, attempts, payload, created_at`

func (r *Deliveries) Create(ctx context.Context, d *domain.Delivery) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webhook_deliveries (id, webhook_id, owner_id, event, status, response_code, attempts, payload, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		d.ID, d.WebhookID, d.OwnerID, d.Event, d.Status, d.ResponseCode, d.Attempts, d.Payload, d.CreatedAt)
	return err
}

func (r *Deliveries) ClaimPending(ctx context.Context, limit int) ([]domain.Delivery, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+dCols+` FROM webhook_deliveries WHERE status=$1 ORDER BY created_at LIMIT $2`,
		domain.DeliveryPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Delivery
	for rows.Next() {
		var d domain.Delivery
		if err := rows.Scan(&d.ID, &d.WebhookID, &d.OwnerID, &d.Event, &d.Status,
			&d.ResponseCode, &d.Attempts, &d.Payload, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *Deliveries) MarkResult(ctx context.Context, id, status string, responseCode, attempts int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE webhook_deliveries SET status=$2, response_code=$3, attempts=$4, updated_at=now() WHERE id=$1`,
		id, status, responseCode, attempts)
	return err
}
