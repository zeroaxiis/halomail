// Package postgres implements the identity repository ports over PostgreSQL
// using pgx. Each aggregate gets its own repo type. Database errors are
// translated into shared/errs kinds so the app and RPC layers stay
// storage-agnostic.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
)

// ---- Users ---------------------------------------------------------------

type Users struct{ pool *pgxpool.Pool }

func NewUsers(pool *pgxpool.Pool) *Users { return &Users{pool: pool} }

const userColumns = `id, org_id, email, name, handle, avatar_url, timezone, password_hash, created_at, updated_at`

func (r *Users) CreateOrgAndUser(ctx context.Context, org *domain.Org, user *domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	if _, err := tx.Exec(ctx,
		`INSERT INTO orgs (id, name, slug, created_at) VALUES ($1, $2, $3, $4)`,
		org.ID, org.Name, org.Slug, org.CreatedAt,
	); err != nil {
		return mapWrite(err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO users (id, org_id, email, name, handle, avatar_url, timezone, password_hash, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		user.ID, user.OrgID, user.Email, user.Name, user.Handle, user.AvatarURL,
		user.Timezone, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	); err != nil {
		return mapWrite(err)
	}

	return tx.Commit(ctx)
}

func (r *Users) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE lower(email) = lower($1)`, email))
}

func (r *Users) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id))
}

func (r *Users) GetUserByHandle(ctx context.Context, handle string) (*domain.User, error) {
	return scanUser(r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE lower(handle) = lower($1)`, handle))
}

func (r *Users) UpdateUser(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET name=$2, handle=$3, avatar_url=$4, timezone=$5, updated_at=$6 WHERE id=$1`,
		u.ID, u.Name, u.Handle, u.AvatarURL, u.Timezone, u.UpdatedAt,
	)
	return mapWrite(err)
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Handle, &u.AvatarURL,
		&u.Timezone, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("user not found")
		}
		return nil, err
	}
	return &u, nil
}

// ---- Sessions ------------------------------------------------------------

type Sessions struct{ pool *pgxpool.Pool }

func NewSessions(pool *pgxpool.Pool) *Sessions { return &Sessions{pool: pool} }

func (r *Sessions) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, org_id, refresh_token_hash, expires_at, revoked, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		s.ID, s.UserID, s.OrgID, s.RefreshTokenHash, s.ExpiresAt, s.Revoked, s.CreatedAt,
	)
	return mapWrite(err)
}

func (r *Sessions) GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error) {
	var s domain.Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, org_id, refresh_token_hash, expires_at, revoked, created_at
		 FROM sessions WHERE refresh_token_hash = $1`, hash,
	).Scan(&s.ID, &s.UserID, &s.OrgID, &s.RefreshTokenHash, &s.ExpiresAt, &s.Revoked, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("session not found")
		}
		return nil, err
	}
	return &s, nil
}

func (r *Sessions) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked = true WHERE id = $1`, id)
	return err
}

func (r *Sessions) Consume(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked=true WHERE id=$1 AND NOT revoked AND expires_at>now()`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errs.Unauthorized("session expired or already refreshed")
	}
	return nil
}

// ---- API keys ------------------------------------------------------------

type APIKeys struct{ pool *pgxpool.Pool }

func NewAPIKeys(pool *pgxpool.Pool) *APIKeys { return &APIKeys{pool: pool} }

func (r *APIKeys) Create(ctx context.Context, k *domain.APIKey) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO api_keys (id, org_id, user_id, name, prefix, last_four, secret_hash, scopes, created_at, revoked)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		k.ID, k.OrgID, k.UserID, k.Name, k.Prefix, k.LastFour, k.SecretHash, k.Scopes, k.CreatedAt, k.Revoked,
	)
	return mapWrite(err)
}

func (r *APIKeys) ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, user_id, name, prefix, last_four, secret_hash, scopes, created_at, last_used_at, revoked
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, *k)
	}
	return keys, rows.Err()
}

func (r *APIKeys) GetBySecretHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	k, err := scanAPIKey(r.pool.QueryRow(ctx,
		`SELECT id, org_id, user_id, name, prefix, last_four, secret_hash, scopes, created_at, last_used_at, revoked
		 FROM api_keys WHERE secret_hash = $1`, hash))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("api key not found")
		}
		return nil, err
	}
	return k, nil
}

func (r *APIKeys) Revoke(ctx context.Context, id, userID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET revoked = true WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("api key not found")
	}
	return nil
}

func (r *APIKeys) TouchLastUsed(ctx context.Context, id string, t time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE api_keys SET last_used_at = $2 WHERE id = $1`, id, t)
	return err
}

func scanAPIKey(row pgx.Row) (*domain.APIKey, error) {
	var (
		k        domain.APIKey
		lastUsed *time.Time
	)
	if err := row.Scan(&k.ID, &k.OrgID, &k.UserID, &k.Name, &k.Prefix, &k.LastFour,
		&k.SecretHash, &k.Scopes, &k.CreatedAt, &lastUsed, &k.Revoked); err != nil {
		return nil, err
	}
	k.LastUsedAt = lastUsed
	return &k, nil
}

// ---- Audit ---------------------------------------------------------------

type Audit struct{ pool *pgxpool.Pool }

func NewAudit(pool *pgxpool.Pool) *Audit { return &Audit{pool: pool} }

func (r *Audit) Insert(ctx context.Context, l *domain.AuditLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_logs (id, org_id, actor_id, action, target_type, target_id, metadata, ip, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8,$9)`,
		l.ID, l.OrgID, l.ActorID, l.Action, l.TargetType, l.TargetID, l.Metadata, l.IP, l.CreatedAt,
	)
	return err
}

func (r *Audit) ListByOrg(ctx context.Context, orgID string, limit, offset int) ([]domain.AuditLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, org_id, actor_id, action, target_type, target_id, metadata::text, ip, created_at
		 FROM audit_logs WHERE org_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		if err := rows.Scan(&l.ID, &l.OrgID, &l.ActorID, &l.Action, &l.TargetType,
			&l.TargetID, &l.Metadata, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// mapWrite translates unique-constraint violations into a Conflict.
func mapWrite(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return errs.Conflict("a record with these details already exists")
	}
	return err
}
