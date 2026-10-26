package app

import (
	"context"
	"testing"

	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
)

type googleMockUserRepo struct {
	UserRepo
	usersByEmail map[string]*domain.User
	createdOrg   *domain.Org
	createdUser  *domain.User
	updatedUser  *domain.User
}

func (m *googleMockUserRepo) PurgeExpiredUserByEmail(ctx context.Context, email string) (bool, error) {
	return false, nil
}
func (m *googleMockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.usersByEmail[email]; ok {
		return u, nil
	}
	return nil, errs.NotFound("user not found")
}
func (m *googleMockUserRepo) CreateOrgAndUser(ctx context.Context, org *domain.Org, user *domain.User) error {
	m.createdOrg = org
	m.createdUser = user
	if m.usersByEmail == nil {
		m.usersByEmail = make(map[string]*domain.User)
	}
	m.usersByEmail[user.Email] = user
	return nil
}
func (m *googleMockUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	m.updatedUser = user
	return nil
}
func (m *googleMockUserRepo) CancelAccountDeletion(ctx context.Context, userID string) error {
	return nil
}

type googleMockSessionRepo struct {
	SessionRepo
	created *domain.Session
}

func (m *googleMockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	m.created = s
	return nil
}

func TestGoogleLoginNewUser(t *testing.T) {
	userRepo := &googleMockUserRepo{usersByEmail: make(map[string]*domain.User)}
	sessRepo := &googleMockSessionRepo{}
	svc := New(Repos{
		Users:    userRepo,
		Sessions: sessRepo,
		Audit:    auditRepository{},
	}, Config{JWTSecret: "test-secret-at-least-32-chars-long-1234567890"}, nil)

	res, err := svc.GoogleLogin(context.Background(), "newuser@example.com", "New User", "https://avatar.url/1", "google-sub-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.User == nil || res.User.Email != "newuser@example.com" {
		t.Fatalf("unexpected user: %+v", res.User)
	}
	if res.User.Role != domain.RoleOwner {
		t.Fatalf("expected role owner, got %s", res.User.Role)
	}
	if res.User.AvatarURL != "https://avatar.url/1" {
		t.Fatalf("expected avatar url, got %s", res.User.AvatarURL)
	}
	if res.Session.AccessToken == "" || res.Session.RefreshToken == "" {
		t.Fatalf("expected non-empty session")
	}
	if userRepo.createdOrg == nil || userRepo.createdUser == nil {
		t.Fatalf("expected org and user to be created")
	}
}

func TestGoogleLoginExistingUser(t *testing.T) {
	existing := &domain.User{
		ID:        "usr_existing",
		OrgID:     "org_existing",
		Email:     "existing@example.com",
		Name:      "Existing",
		AvatarURL: "",
		Role:      domain.RoleOwner,
	}
	userRepo := &googleMockUserRepo{
		usersByEmail: map[string]*domain.User{
			"existing@example.com": existing,
		},
	}
	sessRepo := &googleMockSessionRepo{}
	svc := New(Repos{
		Users:    userRepo,
		Sessions: sessRepo,
		Audit:    auditRepository{},
	}, Config{JWTSecret: "test-secret-at-least-32-chars-long-1234567890"}, nil)

	res, err := svc.GoogleLogin(context.Background(), "existing@example.com", "Existing Updated", "https://new-avatar.url", "google-sub-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.User.ID != "usr_existing" {
		t.Fatalf("expected existing user id, got %s", res.User.ID)
	}
	if userRepo.createdOrg != nil {
		t.Fatalf("should not create new org for existing user")
	}
	if res.User.AvatarURL != "https://new-avatar.url" {
		t.Fatalf("expected avatar to update, got %s", res.User.AvatarURL)
	}
	if res.Session.AccessToken == "" || res.Session.RefreshToken == "" {
		t.Fatalf("expected session to be issued")
	}
}

func TestGoogleLoginRejectsInvalidEmail(t *testing.T) {
	svc := New(Repos{}, Config{JWTSecret: "test-secret-at-least-32-chars-long-1234567890"}, nil)
	_, err := svc.GoogleLogin(context.Background(), "invalid-email", "Name", "", "")
	if err == nil {
		t.Fatalf("expected error for invalid email")
	}
}
