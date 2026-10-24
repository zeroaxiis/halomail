package app

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/zeroaxiis/halomail/services/notification/internal/domain"
)

// Worker delivers pending webhook deliveries with HMAC-signed payloads and
// bounded retries. One instance runs per process.
type Worker struct {
	webhooks   WebhookRepo
	deliveries DeliveryRepo
	client     *http.Client
	logger     *slog.Logger
	interval   time.Duration
	batch      int
}

func NewWorker(r Repos, logger *slog.Logger) *Worker {
	return &Worker{
		webhooks:   r.Webhooks,
		deliveries: r.Deliveries,
		client:     &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
		interval:   2 * time.Second,
		batch:      20,
	}
}

// Run polls for pending deliveries until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	pending, err := w.deliveries.ClaimPending(ctx, w.batch)
	if err != nil {
		w.logger.ErrorContext(ctx, "claim pending deliveries", "error", err.Error())
		return
	}
	for i := range pending {
		w.deliver(ctx, pending[i])
	}
}

func (w *Worker) deliver(ctx context.Context, d domain.Delivery) {
	attempts := d.Attempts + 1

	wh, err := w.webhooks.GetByID(ctx, d.WebhookID)
	if err != nil {
		_ = w.deliveries.MarkResult(ctx, d.ID, domain.DeliveryFailed, 0, attempts)
		return
	}

	body := []byte(d.Payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		w.finish(ctx, d.ID, 0, attempts)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "HaloMail-Webhooks/1")
	req.Header.Set("X-HaloMail-Event", d.Event)
	req.Header.Set("X-HaloMail-Signature", "sha256="+sign(wh.Secret, body))

	resp, err := w.client.Do(req)
	code := 0
	if resp != nil {
		code = resp.StatusCode
		resp.Body.Close()
	}
	if err != nil || code >= 300 {
		w.finish(ctx, d.ID, code, attempts)
		w.logger.WarnContext(ctx, "webhook delivery failed",
			"delivery", d.ID, "webhook", wh.ID, "code", code, "attempt", attempts)
		return
	}

	_ = w.deliveries.MarkResult(ctx, d.ID, domain.DeliverySucceeded, code, attempts)
	w.logger.InfoContext(ctx, "webhook delivered", "delivery", d.ID, "webhook", wh.ID, "code", code)
}

// finish marks a failed attempt, retrying until the attempt ceiling.
func (w *Worker) finish(ctx context.Context, id string, code, attempts int) {
	status := domain.DeliveryPending
	if attempts >= domain.MaxAttempts {
		status = domain.DeliveryFailed
	}
	_ = w.deliveries.MarkResult(ctx, id, status, code, attempts)
}

func sign(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}
