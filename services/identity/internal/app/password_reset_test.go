package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/crypto"
	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
)

type mockUserRepoForReset struct {
	UserRepo
	users map[string]*domain.User
}

func (m *mockUserRepoForReset) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, errs.NotFound("user not found")
}

func (m *mockUserRepoForReset) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, errs.NotFound("user not found")
}

type mockAuditRepoForReset struct {
	AuditRepo
}

func (m *mockAuditRepoForReset) Insert(ctx context.Context, l *domain.AuditLog) error {
	return nil
}

type mockPasswordResetRepo struct {
	createdReset     *domain.PasswordReset
	outboxID         string
	emailHTML        string
	resets           map[string]*domain.PasswordReset
	completedResetID string
	completedUserID  string
	completedHash    string
}

func newMockPasswordResetRepo() *mockPasswordResetRepo {
	return &mockPasswordResetRepo{
		resets: make(map[string]*domain.PasswordReset),
	}
}

func (m *mockPasswordResetRepo) CreateReset(ctx context.Context, reset *domain.PasswordReset, outboxID, emailHTML string) error {
	m.createdReset = reset
	m.outboxID = outboxID
	m.emailHTML = emailHTML
	m.resets[reset.TokenHash] = reset
	return nil
}

func (m *mockPasswordResetRepo) GetResetByTokenHash(ctx context.Context, hash string) (*domain.PasswordReset, error) {
	if r, ok := m.resets[hash]; ok {
		return r, nil
	}
	return nil, errs.NotFound("reset not found")
}

func (m *mockPasswordResetRepo) CompleteReset(ctx context.Context, resetID, userID, newPasswordHash string) error {
	m.completedResetID = resetID
	m.completedUserID = userID
	m.completedHash = newPasswordHash
	if m.createdReset != nil {
		if r, ok := m.resets[m.createdReset.TokenHash]; ok && r.ID == resetID {
			now := time.Now()
			r.UsedAt = &now
		}
	}
	return nil
}

func TestForgotPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid email returns error", func(t *testing.T) {
		svc := New(Repos{}, Config{}, nil)
		err := svc.ForgotPassword(ctx, "not-an-email")
		if err == nil {
			t.Fatalf("expected error for invalid email")
		}
	})

	t.Run("nonexistent user returns nil to prevent enumeration", func(t *testing.T) {
		mockUsers := &mockUserRepoForReset{users: map[string]*domain.User{}}
		svc := New(Repos{Users: mockUsers}, Config{}, nil)
		err := svc.ForgotPassword(ctx, "ghost@example.com")
		if err != nil {
			t.Fatalf("expected nil for nonexistent user, got: %v", err)
		}
	})

	t.Run("valid user generates reset token and outbox email", func(t *testing.T) {
		mockUsers := &mockUserRepoForReset{
			users: map[string]*domain.User{
				"usr_1": {ID: "usr_1", Email: "alex@example.com", Name: "Alex", OrgID: "org_1"},
			},
		}
		mockResets := newMockPasswordResetRepo()
		svc := New(Repos{
			Users:          mockUsers,
			PasswordResets: mockResets,
			Audit:          &mockAuditRepoForReset{},
		}, Config{WebURL: "http://localhost:3000"}, nil)

		err := svc.ForgotPassword(ctx, "alex@example.com")
		if err != nil {
			t.Fatalf("expected success, got: %v", err)
		}
		if mockResets.createdReset == nil {
			t.Fatalf("expected password reset record to be created")
		}
		if mockResets.createdReset.Email != "alex@example.com" {
			t.Errorf("expected email alex@example.com, got %s", mockResets.createdReset.Email)
		}
		if mockResets.outboxID == "" || mockResets.emailHTML == "" {
			t.Errorf("expected outbox mail to be queued")
		}
	})
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("invalid token returns error", func(t *testing.T) {
		svc := New(Repos{}, Config{}, nil)
		err := svc.ResetPassword(ctx, "short", "validpassword123")
		if err == nil {
			t.Fatalf("expected error for short token")
		}
	})

	t.Run("short password returns error", func(t *testing.T) {
		svc := New(Repos{}, Config{}, nil)
		token, _ := crypto.RandomToken(32)
		err := svc.ResetPassword(ctx, token, "short")
		if err == nil {
			t.Fatalf("expected error for short password")
		}
	})

	t.Run("unknown token returns unauthorized", func(t *testing.T) {
		mockResets := newMockPasswordResetRepo()
		svc := New(Repos{PasswordResets: mockResets}, Config{}, nil)
		token, _ := crypto.RandomToken(32)
		err := svc.ResetPassword(ctx, token, "newpassword123")
		if err == nil || errs.KindOf(err) != errs.KindUnauthorized {
			t.Fatalf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("expired token returns unauthorized", func(t *testing.T) {
		mockResets := newMockPasswordResetRepo()
		rawToken, _ := crypto.RandomToken(32)
		tokenHash := crypto.SHA256Hex(rawToken)
		mockResets.resets[tokenHash] = &domain.PasswordReset{
			ID:        "rst_1",
			UserID:    "usr_1",
			TokenHash: tokenHash,
			ExpiresAt: now.Add(-10 * time.Minute), // expired
		}

		svc := New(Repos{PasswordResets: mockResets}, Config{}, nil)
		err := svc.ResetPassword(ctx, rawToken, "newpassword123")
		if err == nil || errs.KindOf(err) != errs.KindUnauthorized {
			t.Fatalf("expected unauthorized error for expired token, got %v", err)
		}
	})

	t.Run("used token returns unauthorized", func(t *testing.T) {
		mockResets := newMockPasswordResetRepo()
		rawToken, _ := crypto.RandomToken(32)
		tokenHash := crypto.SHA256Hex(rawToken)
		usedTime := now.Add(-5 * time.Minute)
		mockResets.resets[tokenHash] = &domain.PasswordReset{
			ID:        "rst_1",
			UserID:    "usr_1",
			TokenHash: tokenHash,
			ExpiresAt: now.Add(50 * time.Minute),
			UsedAt:    &usedTime, // already used
		}

		svc := New(Repos{PasswordResets: mockResets}, Config{}, nil)
		err := svc.ResetPassword(ctx, rawToken, "newpassword123")
		if err == nil || errs.KindOf(err) != errs.KindUnauthorized {
			t.Fatalf("expected unauthorized error for used token, got %v", err)
		}
	})

	t.Run("valid token resets password and marks token used", func(t *testing.T) {
		mockResets := newMockPasswordResetRepo()
		rawToken, _ := crypto.RandomToken(32)
		tokenHash := crypto.SHA256Hex(rawToken)
		mockResets.resets[tokenHash] = &domain.PasswordReset{
			ID:        "rst_1",
			UserID:    "usr_1",
			TokenHash: tokenHash,
			ExpiresAt: now.Add(50 * time.Minute),
		}
		mockResets.createdReset = mockResets.resets[tokenHash]

		svc := New(Repos{
			PasswordResets: mockResets,
			Audit:          &mockAuditRepoForReset{},
		}, Config{}, nil)

		err := svc.ResetPassword(ctx, rawToken, "strongnewpassword123")
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if mockResets.completedResetID != "rst_1" || mockResets.completedUserID != "usr_1" {
			t.Errorf("expected complete reset for rst_1/usr_1, got %s/%s", mockResets.completedResetID, mockResets.completedUserID)
		}
		if mockResets.completedHash == "" {
			t.Errorf("expected new password hash to be generated")
		}
	})
}
