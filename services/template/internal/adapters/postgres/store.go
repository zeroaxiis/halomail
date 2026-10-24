// Package postgres implements the template repository over PostgreSQL.
package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zeroaxiis/halomail/services/template/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
)

type Templates struct{ pool *pgxpool.Pool }

func NewTemplates(pool *pgxpool.Pool) *Templates { return &Templates{pool: pool} }

const cols = `id, owner_id, name, theme, subject, custom_html, created_at, updated_at`

func (r *Templates) Create(ctx context.Context, t *domain.Template) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO templates (`+cols+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		t.ID, t.OwnerID, t.Name, t.Theme, t.Subject, t.CustomHTML, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Templates) GetByID(ctx context.Context, id string) (*domain.Template, error) {
	return scan(r.pool.QueryRow(ctx, `SELECT `+cols+` FROM templates WHERE id=$1`, id))
}

func (r *Templates) ListByOwner(ctx context.Context, ownerID string) ([]domain.Template, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+cols+` FROM templates WHERE owner_id=$1 ORDER BY created_at DESC`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Template
	for rows.Next() {
		t, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *Templates) Update(ctx context.Context, t *domain.Template) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE templates SET name=$2, theme=$3, subject=$4, custom_html=$5, updated_at=$6 WHERE id=$1`,
		t.ID, t.Name, t.Theme, t.Subject, t.CustomHTML, t.UpdatedAt)
	return err
}

func (r *Templates) Delete(ctx context.Context, id, ownerID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM templates WHERE id=$1 AND owner_id=$2`, id, ownerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("template not found")
	}
	return nil
}

func scan(row pgx.Row) (*domain.Template, error) {
	var t domain.Template
	if err := row.Scan(&t.ID, &t.OwnerID, &t.Name, &t.Theme, &t.Subject, &t.CustomHTML, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("template not found")
		}
		return nil, err
	}
	return &t, nil
}
