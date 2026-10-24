// Package domain holds the identity entities and business invariants. It is
// pure: no database, no transport, no framework imports.
package domain

import "time"

// User is an authenticated principal, scoped to an Org.
type User struct {
	ID           string
	OrgID        string
	Email        string
	Name         string
	Handle       string
	AvatarURL    string
	Timezone     string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

const (
	RoleOwner   = "owner"
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleViewer  = "viewer"
)

// Invitation represents a pending or historical invite to join an Org.
type Invitation struct {
	ID            string
	OrgID         string
	Email         string
	Role          string
	Token         string
	InvitedBy     string
	InvitedByName string
	Status        string // pending, accepted, revoked, expired
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

const (
	InvitePending  = "pending"
	InviteAccepted = "accepted"
	InviteRevoked  = "revoked"
	InviteExpired  = "expired"
)

// Org is the tenant boundary. A user belongs to exactly one org.
type Org struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
}

// Session is a refresh-token-backed login. Only the hash of the refresh token
// is persisted; the plaintext is shown to the client exactly once.
type Session struct {
	ID               string
	UserID           string
	OrgID            string
	RefreshTokenHash string
	ExpiresAt        time.Time
	Revoked          bool
	CreatedAt        time.Time
}

// Expired reports whether the session is past its lifetime as of now.
func (s Session) Expired(now time.Time) bool { return now.After(s.ExpiresAt) }

// Active reports whether the session can still be used.
func (s Session) Active(now time.Time) bool { return !s.Revoked && !s.Expired(now) }

// APIKey is a developer credential. Only the hash of the secret is stored.
type APIKey struct {
	ID         string
	OrgID      string
	UserID     string
	Name       string
	Prefix     string
	LastFour   string
	SecretHash string
	Scopes     []string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	Revoked    bool
}

// AuditLog is an immutable record of a security-relevant action.
type AuditLog struct {
	ID         string
	OrgID      string
	ActorID    string
	Action     string
	TargetType string
	TargetID   string
	Metadata   string // JSON
	IP         string
	CreatedAt  time.Time
}

// PasswordReset represents a single-use token issued to recover account access.
type PasswordReset struct {
	ID        string
	UserID    string
	Email     string
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func (p PasswordReset) Valid(now time.Time) bool {
	return p.UsedAt == nil && now.Before(p.ExpiresAt)
}

