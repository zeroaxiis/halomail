package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	"github.com/zeroaxiis/halomail/services/shared/outbox"
)

type PasswordResets struct {
	pool *pgxpool.Pool
}

func NewPasswordResets(pool *pgxpool.Pool) *PasswordResets {
	return &PasswordResets{pool: pool}
}

func (r *PasswordResets) CreateReset(ctx context.Context, reset *domain.PasswordReset, outboxID, emailHTML string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Invalidate any existing active resets for this user
	if _, err := tx.Exec(ctx,
		`UPDATE password_resets SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`,
		reset.UserID,
	); err != nil {
		return mapWrite(err)
	}

	// Insert the new reset token record
	if _, err := tx.Exec(ctx,
		`INSERT INTO password_resets (id, user_id, email, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		reset.ID, reset.UserID, reset.Email, reset.TokenHash, reset.ExpiresAt, reset.CreatedAt,
	); err != nil {
		return mapWrite(err)
	}

	// Enqueue transactional email to outbox if content provided
	if outboxID != "" && emailHTML != "" {
		if err := outbox.Queue(ctx, tx, outboxID, reset.Email, "", "Reset your HaloMail password", emailHTML); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PasswordResets) GetResetByTokenHash(ctx context.Context, hash string) (*domain.PasswordReset, error) {
	var pr domain.PasswordReset
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, email, token_hash, expires_at, used_at, created_at
		 FROM password_resets WHERE token_hash = $1`, hash).
		Scan(&pr.ID, &pr.UserID, &pr.Email, &pr.TokenHash, &pr.ExpiresAt, &pr.UsedAt, &pr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("invalid or expired password reset link")
		}
		return nil, err
	}
	return &pr, nil
}

func (r *PasswordResets) CompleteReset(ctx context.Context, resetID, userID, newPasswordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Update user password hash
	tag, err := tx.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = now() WHERE id = $2`,
		newPasswordHash, userID,
	)
	if err != nil {
		return mapWrite(err)
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("user not found")
	}

	// Mark reset token as used
	if _, err := tx.Exec(ctx,
		`UPDATE password_resets SET used_at = now() WHERE id = $1`,
		resetID,
	); err != nil {
		return mapWrite(err)
	}

	// Invalidate all existing sessions for this user for security
	if _, err := tx.Exec(ctx,
		`UPDATE sessions SET revoked = true WHERE user_id = $1`,
		userID,
	); err != nil {
		return mapWrite(err)
	}

	return tx.Commit(ctx)
}
