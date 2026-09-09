// Package app holds contact use cases over repository ports.
package app

import (
	"context"

	"github.com/aashishrajdev/halomail/services/contact/internal/domain"
)

type FormRepo interface {
	Create(ctx context.Context, f *domain.Form) error
	GetByID(ctx context.Context, id string) (*domain.Form, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Form, error)
	ListByOwner(ctx context.Context, ownerID string) ([]domain.Form, error)
	Update(ctx context.Context, f *domain.Form) error
	Delete(ctx context.Context, id, ownerID string) error
}

type MessageRepo interface {
	Create(ctx context.Context, m *domain.Message) error
	GetByID(ctx context.Context, id string) (*domain.Message, error)
	List(ctx context.Context, ownerID, formID string, unreadOnly bool, limit, offset int) ([]domain.Message, error)
	MarkRead(ctx context.Context, id, ownerID string, read bool) error
	Delete(ctx context.Context, id, ownerID string) error
	GetUsageStats(ctx context.Context, ownerID string) (*domain.UsageStats, error)
}

// Forwarder delivers a received message to the form's target inbox (and/or
// webhooks). The default implementation logs; the real one calls notification.
type Forwarder interface {
	Forward(ctx context.Context, form *domain.Form, msg *domain.Message) error
}

type Repos struct {
	Forms    FormRepo
	Messages MessageRepo
}
