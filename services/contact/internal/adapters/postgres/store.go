// Package postgres implements the contact repository ports over PostgreSQL.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aashishrajdev/halomail/services/contact/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/outbox"
	"github.com/aashishrajdev/halomail/services/shared/usage"
)

// ---- Forms ---------------------------------------------------------------

type Forms struct{ pool *pgxpool.Pool }

func NewForms(pool *pgxpool.Pool) *Forms { return &Forms{pool: pool} }

func (r *Forms) Inbox(ctx context.Context, ownerID string) (*domain.Form, error) {
	id := "inbox_" + ownerID
	_, err := r.pool.Exec(ctx, `INSERT INTO forms(id,owner_id,name,slug,target_email)
	SELECT $1,id,'Website forms',$1,email FROM users WHERE id=$2 ON CONFLICT(id) DO NOTHING`, id, ownerID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

const formColumns = `id, owner_id, name, slug, target_email, spam_protection, redirect_url, fields, active, created_at`

func (r *Forms) Create(ctx context.Context, f *domain.Form) error {
	fields, err := json.Marshal(f.Fields)
	if err != nil {
		return err
	}
	if f.Fields == nil {
		fields = []byte("[]")
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO forms (`+formColumns+`)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9,$10)`,
		f.ID, f.OwnerID, f.Name, f.Slug, f.TargetEmail, f.SpamProtection, f.RedirectURL, string(fields), f.Active, f.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.Conflict("a form with that slug already exists")
		}
	}
	return err
}

func (r *Forms) GetByID(ctx context.Context, id string) (*domain.Form, error) {
	return scanForm(r.pool.QueryRow(ctx, `SELECT `+formColumns+` FROM forms WHERE id=$1`, id))
}

func (r *Forms) GetBySlug(ctx context.Context, slug string) (*domain.Form, error) {
	return scanForm(r.pool.QueryRow(ctx, `SELECT `+formColumns+` FROM forms WHERE slug=$1`, slug))
}

func (r *Forms) ListByOwner(ctx context.Context, ownerID string) ([]domain.Form, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+formColumns+` FROM forms WHERE owner_id=$1 ORDER BY created_at`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Form
	for rows.Next() {
		f, err := scanForm(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *f)
	}
	return out, rows.Err()
}

func (r *Forms) Update(ctx context.Context, f *domain.Form) error {
	fields, err := json.Marshal(f.Fields)
	if err != nil {
		return err
	}
	if f.Fields == nil {
		fields = []byte("[]")
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE forms SET name=$2, target_email=$3, spam_protection=$4, redirect_url=$5, fields=$6::jsonb, active=$7 WHERE id=$1`,
		f.ID, f.Name, f.TargetEmail, f.SpamProtection, f.RedirectURL, string(fields), f.Active)
	return err
}

func (r *Forms) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM forms WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("form not found")
	}
	return nil
}

func scanForm(row pgx.Row) (*domain.Form, error) {
	var (
		f      domain.Form
		fields []byte
	)
	if err := row.Scan(&f.ID, &f.OwnerID, &f.Name, &f.Slug, &f.TargetEmail, &f.SpamProtection,
		&f.RedirectURL, &fields, &f.Active, &f.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("form not found")
		}
		return nil, err
	}
	if len(fields) > 0 {
		_ = json.Unmarshal(fields, &f.Fields)
	}
	return &f, nil
}

// ---- Messages ------------------------------------------------------------

type Messages struct {
	pool   *pgxpool.Pool
	policy usage.Policy
}

func NewMessages(pool *pgxpool.Pool, policies ...usage.Policy) *Messages {
	policy := usage.Policy{Forms: 25, Meetings: 10, Monthly: true}
	if len(policies) > 0 {
		policy = policies[0]
	}
	return &Messages{pool: pool, policy: policy}
}

const msgColumns = `id, form_id, owner_id, sender_name, sender_email, data, ip, user_agent, spam_score, is_spam, read, created_at`

func (r *Messages) Create(ctx context.Context, m *domain.Message) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if !m.IsSpam {
		if err = usage.Consume(ctx, tx, r.policy, m.OwnerID, "forms"); err != nil {
			return err
		}
	}
	data, err := json.Marshal(m.Data)
	if err != nil {
		return err
	}
	if m.Data == nil {
		data = []byte("{}")
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO messages (`+msgColumns+`)
		 VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.FormID, m.OwnerID, m.SenderName, m.SenderEmail, string(data),
		m.IP, m.UserAgent, m.SpamScore, m.IsSpam, m.Read, m.CreatedAt)
	if err != nil {
		return err
	}
	if !m.IsSpam {
		var recipient string
		if err = tx.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, m.OwnerID).Scan(&recipient); err != nil {
			return err
		}
		fields := make(map[string]string, len(m.Data)+2)
		for key, value := range m.Data {
			fields[key] = value
		}
		fields["Sender name"] = m.SenderName
		fields["Sender email"] = m.SenderEmail
		if err = outbox.Queue(ctx, tx, m.ID, recipient, m.SenderEmail, "New website form submission", outbox.Fields("New form submission", fields)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Messages) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	return scanMessage(r.pool.QueryRow(ctx, `SELECT `+msgColumns+` FROM messages WHERE id=$1`, id))
}

func (r *Messages) List(ctx context.Context, ownerID, formID string, unreadOnly bool, limit, offset int) ([]domain.Message, error) {
	args := []any{ownerID}
	q := `SELECT ` + msgColumns + ` FROM messages WHERE owner_id=$1`
	if formID != "" {
		args = append(args, formID)
		q += fmt.Sprintf(" AND form_id=$%d", len(args))
	}
	if unreadOnly {
		q += " AND read=false"
	}
	args = append(args, limit)
	q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args))
	args = append(args, offset)
	q += fmt.Sprintf(" OFFSET $%d", len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Message
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *Messages) MarkRead(ctx context.Context, id, ownerID string, read bool) error {
	tag, err := r.pool.Exec(ctx, `UPDATE messages SET read=$3 WHERE id=$1 AND owner_id=$2`, id, ownerID, read)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("message not found")
	}
	return nil
}

func (r *Messages) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM messages WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("message not found")
	}
	return nil
}

func (r *Messages) GetUsageStats(ctx context.Context, ownerID string) (*domain.UsageStats, error) {
	var stats domain.UsageStats
	err := r.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COALESCE(COUNT(*) FILTER (WHERE read = false), 0),
			COALESCE(COUNT(*) FILTER (WHERE is_spam = true), 0)
		FROM messages 
		WHERE owner_id = $1
	`, ownerID).Scan(&stats.TotalMessages, &stats.UnreadMessages, &stats.SpamPrevented)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func scanMessage(row pgx.Row) (*domain.Message, error) {
	var (
		m    domain.Message
		data []byte
	)
	if err := row.Scan(&m.ID, &m.FormID, &m.OwnerID, &m.SenderName, &m.SenderEmail, &data,
		&m.IP, &m.UserAgent, &m.SpamScore, &m.IsSpam, &m.Read, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("message not found")
		}
		return nil, err
	}
	if len(data) > 0 {
		_ = json.Unmarshal(data, &m.Data)
	}
	return &m, nil
}
