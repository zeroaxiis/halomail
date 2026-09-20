// Package app holds identity use cases. It depends on the domain and on
// repository *interfaces* (ports) implemented by adapters — never on a concrete
// database. This keeps the business logic testable and storage-agnostic.
package app

import (
	"context"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
)

// UserRepo persists users and their org.
type UserRepo interface {
	// CreateOrgAndUser inserts the org and user atomically. It must return an
	// errs.Conflict when the email or handle already exists.
	CreateOrgAndUser(ctx context.Context, org *domain.Org, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByHandle(ctx context.Context, handle string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
}

// SessionRepo persists refresh-token sessions.
type SessionRepo interface {
	Create(ctx context.Context, s *domain.Session) error
	GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error)
	Consume(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
}

// APIKeyRepo persists API keys.
type APIKeyRepo interface {
	Create(ctx context.Context, k *domain.APIKey) error
	ListByUser(ctx context.Context, userID string) ([]domain.APIKey, error)
	GetBySecretHash(ctx context.Context, hash string) (*domain.APIKey, error)
	Revoke(ctx context.Context, id, userID string) error
	TouchLastUsed(ctx context.Context, id string, t time.Time) error
}

// AuditRepo appends and lists audit entries.
type AuditRepo interface {
	Insert(ctx context.Context, l *domain.AuditLog) error
	ListByOrg(ctx context.Context, orgID string, limit, offset int) ([]domain.AuditLog, error)
}

// Repos toplevel groups the ports for wiring convenience.
type Repos struct {
	Users    UserRepo
	Sessions SessionRepo
	APIKeys  APIKeyRepo
	Audit    AuditRepo
}
