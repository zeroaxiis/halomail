// Package app holds identity use cases. It depends on the domain and on
// repository *interfaces* (ports) implemented by adapters — never on a concrete
// database. This keeps the business logic testable and storage-agnostic.
package app

import (
	"context"
	"time"

	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
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
	ApplyPayment(ctx context.Context, orgID, orderID, paymentID, tier string, paidAt, cycleEnd time.Time) error
	ApplyRefunds(ctx context.Context, orgID, orderID, paymentID, currency string, paymentAmount int64, refunds []domain.BillingRefund) (domain.RefundResult, error)
	UpdateOrgBilling(ctx context.Context, orgID, customerID, subID, tier string, cycleEnd time.Time) error
	ScheduleAccountDeletion(ctx context.Context, userID string, scheduledAt time.Time) error
	CancelAccountDeletion(ctx context.Context, userID string) error
	HardDeleteUser(ctx context.Context, userID string) error
	PurgeExpiredUserByEmail(ctx context.Context, email string) (bool, error)
}

// SessionRepo persists refresh-token sessions.
type SessionRepo interface {
	Create(ctx context.Context, s *domain.Session) error
	GetByRefreshHash(ctx context.Context, hash string) (*domain.Session, error)
	Consume(ctx context.Context, id string) error
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
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

// TeamRepo persists team members and invitations.
type TeamRepo interface {
	ListMembersByOrg(ctx context.Context, orgID string) ([]domain.User, error)
	GetOrgTier(ctx context.Context, orgID string) (string, error)
	GetOrgName(ctx context.Context, orgID string) (string, error)
	GetMemberCount(ctx context.Context, orgID string) (int, error)
	CreateInvitation(ctx context.Context, inv *domain.Invitation, outboxID, emailHTML, subject string) error
	ListInvitationsByOrg(ctx context.Context, orgID string) ([]domain.Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*domain.Invitation, error)
	RevokeInvitation(ctx context.Context, orgID, invitationID string) error
	UpdateUserRole(ctx context.Context, orgID, userID, newRole string) error
	RemoveUserFromOrg(ctx context.Context, orgID, userID string) error
	LeaveTeam(ctx context.Context, orgID, userID string) error
	AcceptInvitation(ctx context.Context, inv *domain.Invitation, user *domain.User) error
}

// PasswordResetRepo persists password recovery tokens and handles password updates.
type PasswordResetRepo interface {
	CreateReset(ctx context.Context, reset *domain.PasswordReset, outboxID, emailHTML string) error
	GetResetByTokenHash(ctx context.Context, hash string) (*domain.PasswordReset, error)
	CompleteReset(ctx context.Context, resetID, userID, newPasswordHash string) error
}

// Repos toplevel groups the ports for wiring convenience.
type Repos struct {
	Users          UserRepo
	Sessions       SessionRepo
	APIKeys        APIKeyRepo
	Audit          AuditRepo
	Team           TeamRepo
	PasswordResets PasswordResetRepo
}


