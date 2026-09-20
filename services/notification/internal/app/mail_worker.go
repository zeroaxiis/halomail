package app

import (
	"context"
	"errors"
	"github.com/aashishrajdev/halomail/services/notification/internal/email"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

func RunMailWorker(ctx context.Context, pool *pgxpool.Pool, sender email.Sender, logger *slog.Logger) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := deliverMail(ctx, pool, sender); err != nil {
				logger.ErrorContext(ctx, "email delivery failed; queued for retry", "error", err.Error())
			}
		}
	}
}

func deliverMail(ctx context.Context, pool *pgxpool.Pool, sender email.Sender) error {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id, recipient, replyTo, subject, body string
	var attempts int
	err = tx.QueryRow(ctx, `SELECT id,recipient,reply_to,subject,html,attempts FROM mail_outbox
	WHERE delivered_at IS NULL AND available_at<=now() ORDER BY available_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(&id, &recipient, &replyTo, &subject, &body, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, _, sendErr := sender.Send(ctx, email.Message{ID: id, To: []string{recipient}, ReplyTo: replyTo, Subject: subject, HTML: body})
	if sendErr == nil {
		_, err = tx.Exec(ctx, `UPDATE mail_outbox SET delivered_at=now(),attempts=attempts+1 WHERE id=$1`, id)
	} else {
		delay := min(3600, 10<<min(attempts, 8))
		_, err = tx.Exec(ctx, `UPDATE mail_outbox SET attempts=attempts+1,available_at=now()+$2*interval '1 second' WHERE id=$1`, id, delay)
	}
	if err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return sendErr
}
