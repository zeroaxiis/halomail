// Package app holds notification use cases: transactional email and webhooks.
package app

import (
	"context"

	"github.com/zeroaxiis/halomail/services/notification/internal/domain"
)

type WebhookRepo interface {
	Create(ctx context.Context, w *domain.Webhook) error
	GetByID(ctx context.Context, id string) (*domain.Webhook, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Webhook, error)
	ListSubscribed(ctx context.Context, ownerID, event string) ([]domain.Webhook, error)
	UpdateSecret(ctx context.Context, id, ownerID, secret string) error
	Delete(ctx context.Context, id, ownerID string) error
}

type DeliveryRepo interface {
	Create(ctx context.Context, d *domain.Delivery) error
	// ClaimPending returns up to limit deliveries that still need attempts.
	ClaimPending(ctx context.Context, limit int) ([]domain.Delivery, error)
	MarkResult(ctx context.Context, id, status string, responseCode, attempts int) error
}

type Repos struct {
	Webhooks   WebhookRepo
	Deliveries DeliveryRepo
}
