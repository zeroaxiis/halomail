package app

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeroaxiis/halomail/services/contact/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/emailpool"
	"github.com/zeroaxiis/halomail/services/shared/outbox"
)

// ResendForwarder dispatches incoming form submissions to the target email
// immediately upon receipt, providing high-reliability real-time delivery with
// multi-account pooling, round-robin, and automatic failover.
type ResendForwarder struct {
	pool      *pgxpool.Pool
	emailPool *emailpool.Pool
}

func NewResendForwarder(pool *pgxpool.Pool, apiKey, from string) *ResendForwarder {
	return &ResendForwarder{
		pool:      pool,
		emailPool: emailpool.NewFromEnv(),
	}
}

func (f *ResendForwarder) Forward(ctx context.Context, form *domain.Form, msg *domain.Message) error {
	target := form.TargetEmail
	if target == "" {
		return nil
	}

	fields := make(map[string]string, len(msg.Data)+2)
	for k, v := range msg.Data {
		fields[k] = v
	}
	if msg.SenderName != "" {
		fields["Sender name"] = msg.SenderName
	}
	if msg.SenderEmail != "" {
		fields["Sender email"] = msg.SenderEmail
	}
	htmlBody := outbox.Fields("New form submission: "+form.Name, fields)

	// Dispatch immediately in background via the multi-account pool
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		_, _, err := f.emailPool.Send(bgCtx, emailpool.Message{
			ID:      "dispatch-" + msg.ID,
			To:      []string{target},
			ReplyTo: msg.SenderEmail,
			Subject: "New website form submission: " + form.Name,
			HTML:    htmlBody,
		})
		if err == nil {
			// Mark as delivered in mail_outbox so background worker does not re-send
			if f.pool != nil {
				_, _ = f.pool.Exec(bgCtx, `UPDATE mail_outbox SET delivered_at=now() WHERE id=$1`, msg.ID)
			}
		}
	}()

	return nil
}
