package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/zeroaxiis/halomail/services/notification/internal/domain"
	"github.com/zeroaxiis/halomail/services/notification/internal/email"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
)

type Service struct {
	webhooks   WebhookRepo
	deliveries DeliveryRepo
	sender     email.Sender
	now        func() time.Time
}

func New(r Repos, sender email.Sender) *Service {
	return &Service{webhooks: r.Webhooks, deliveries: r.Deliveries, sender: sender, now: time.Now}
}

// ---- Email ---------------------------------------------------------------

func (s *Service) SendEmail(ctx context.Context, to []string, from, replyTo, subject, html, text string) (string, string, error) {
	if len(to) == 0 {
		return "", "", errs.Invalid("at least one recipient is required")
	}
	if strings.TrimSpace(subject) == "" {
		return "", "", errs.Invalid("subject is required")
	}
	id, provider, err := s.sender.Send(ctx, email.Message{
		To: to, From: from, ReplyTo: replyTo, Subject: subject, HTML: html, Text: text,
	})
	if err != nil {
		return "", "", errs.Wrap(err, errs.KindInternal, "send email")
	}
	return id, provider, nil
}

// ---- Webhooks ------------------------------------------------------------

func (s *Service) CreateWebhook(ctx context.Context, ownerID, url string, events []string) (*domain.Webhook, string, error) {
	if !validURL(url) {
		return nil, "", errs.Invalid("a valid https URL is required")
	}
	if len(events) == 0 {
		return nil, "", errs.Invalid("subscribe to at least one event")
	}
	secret := newSecret()
	w := &domain.Webhook{
		ID:        idgen.Prefixed("whk_"),
		OwnerID:   ownerID,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Active:    true,
		CreatedAt: s.now(),
	}
	if err := s.webhooks.Create(ctx, w); err != nil {
		return nil, "", err
	}
	return w, secret, nil
}

func (s *Service) ListWebhooks(ctx context.Context, ownerID string) ([]domain.Webhook, error) {
	return s.webhooks.ListByOwner(ctx, ownerID)
}

func (s *Service) DeleteWebhook(ctx context.Context, ownerID, id string) error {
	return s.webhooks.Delete(ctx, id, ownerID)
}

func (s *Service) RotateSecret(ctx context.Context, ownerID, id string) (string, error) {
	secret := newSecret()
	if err := s.webhooks.UpdateSecret(ctx, id, ownerID, secret); err != nil {
		return "", err
	}
	return secret, nil
}

// Dispatch fans an event out to every subscribed webhook by creating a pending
// delivery for each; the worker delivers them asynchronously.
func (s *Service) Dispatch(ctx context.Context, ownerID, event, payload string) (int, error) {
	if event == "" {
		return 0, errs.Invalid("event is required")
	}
	subs, err := s.webhooks.ListSubscribed(ctx, ownerID, event)
	if err != nil {
		return 0, err
	}
	queued := 0
	for _, w := range subs {
		if !w.Active {
			continue
		}
		d := &domain.Delivery{
			ID:        idgen.Prefixed("whd_"),
			WebhookID: w.ID,
			OwnerID:   ownerID,
			Event:     event,
			Status:    domain.DeliveryPending,
			Payload:   payload,
			CreatedAt: s.now(),
		}
		if err := s.deliveries.Create(ctx, d); err != nil {
			return queued, err
		}
		queued++
	}
	return queued, nil
}

// ---- helpers -------------------------------------------------------------

func newSecret() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return "whsec_" + hex.EncodeToString(b)
}

func validURL(u string) bool {
	return strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")
}
