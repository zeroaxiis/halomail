package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zeroaxiis/halomail/services/identity/internal/crypto"
	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
)

type mockUserRepoForAccountDeletion struct {
	UserRepo
	users   map[string]*domain.User
	orgs    map[string]*domain.Org
	deleted []string
}

func newMockUserRepoForAccountDeletion() *mockUserRepoForAccountDeletion {
	return &mockUserRepoForAccountDeletion{
		users: make(map[string]*domain.User),
		orgs:  make(map[string]*domain.Org),
	}
}

func (m *mockUserRepoForAccountDeletion) CreateOrgAndUser(ctx context.Context, org *domain.Org, user *domain.User) error {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, user.Email) {
			return errs.Conflict("email exists")
		}
	}
	m.orgs[org.ID] = org
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepoForAccountDeletion) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errs.NotFound("user not found")
	}
	return u, nil
}

func (m *mockUserRepoForAccountDeletion) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, errs.NotFound("user not found")
}

func (m *mockUserRepoForAccountDeletion) ScheduleAccountDeletion(ctx context.Context, userID string, scheduledAt time.Time) error {
	u, ok := m.users[userID]
	if !ok {
		return errs.NotFound("user not found")
	}
	now := time.Now()
	u.DeletionScheduledAt = &now
	u.ScheduledDeleteAt = &scheduledAt
	return nil
}

func (m *mockUserRepoForAccountDeletion) CancelAccountDeletion(ctx context.Context, userID string) error {
	u, ok := m.users[userID]
	if !ok {
		return errs.NotFound("user not found")
	}
	u.DeletionScheduledAt = nil
	u.ScheduledDeleteAt = nil
	return nil
}

func (m *mockUserRepoForAccountDeletion) HardDeleteUser(ctx context.Context, userID string) error {
	delete(m.users, userID)
	m.deleted = append(m.deleted, userID)
	return nil
}

func (m *mockUserRepoForAccountDeletion) PurgeExpiredUserByEmail(ctx context.Context, email string) (bool, error) {
	for id, u := range m.users {
		if strings.EqualFold(u.Email, email) && u.ScheduledDeleteAt != nil && time.Now().After(*u.ScheduledDeleteAt) {
			delete(m.users, id)
			m.deleted = append(m.deleted, id)
			return true, nil
		}
	}
	return false, nil
}

type mockSessionRepoForAccountDeletion struct {
	SessionRepo
	revokedAll []string
}

func (m *mockSessionRepoForAccountDeletion) Create(ctx context.Context, s *domain.Session) error {
	return nil
}

func (m *mockSessionRepoForAccountDeletion) RevokeAllForUser(ctx context.Context, userID string) error {
	m.revokedAll = append(m.revokedAll, userID)
	return nil
}

type mockTeamRepoForAccountDeletion struct {
	TeamRepo
	users *mockUserRepoForAccountDeletion
}

func (m *mockTeamRepoForAccountDeletion) LeaveTeam(ctx context.Context, orgID, userID string) error {
	u, ok := m.users.users[userID]
	if !ok {
		return errs.NotFound("user not found")
	}
	if u.Role == domain.RoleOwner {
		return errs.Forbidden("owner cannot leave")
	}
	delete(m.users.users, userID)
	return nil
}

func (m *mockTeamRepoForAccountDeletion) ListMembersByOrg(ctx context.Context, orgID string) ([]domain.User, error) {
	var list []domain.User
	for _, u := range m.users.users {
		if u.OrgID == orgID {
			list = append(list, *u)
		}
	}
	return list, nil
}

func TestTeamMemberCannotCreateAccountUntilLeavingOrDeleting(t *testing.T) {
	userRepo := newMockUserRepoForAccountDeletion()
	sessRepo := &mockSessionRepoForAccountDeletion{}
	teamRepo := &mockTeamRepoForAccountDeletion{users: userRepo}

	// Create user Bob who was invited to Alice's org (role: viewer)
	bob := &domain.User{
		ID:           "usr_bob",
		OrgID:        "org_alice",
		Email:        "bob@example.com",
		Name:         "Bob",
		Role:         domain.RoleViewer,
		PasswordHash: "somehash",
	}
	userRepo.users[bob.ID] = bob

	svc := New(Repos{
		Users:    userRepo,
		Sessions: sessRepo,
		Team:     teamRepo,
	}, Config{JWTSecret: "test-secret-at-least-32-bytes-long!"}, nil)

	ctx := context.Background()

	// Bob tries to Register on halomail
	_, err := svc.Register(ctx, "bob@example.com", "password123!", "Bob Solo")
	if err == nil {
		t.Fatal("expected registration to fail for existing team member")
	}
	if !strings.Contains(err.Error(), "already a member of a team workspace with this email") {
		t.Fatalf("expected specific team membership error, got: %v", err)
	}

	// Now Bob leaves the team
	err = svc.LeaveTeam(ctx, "usr_bob")
	if err != nil {
		t.Fatalf("failed to leave team: %v", err)
	}

	// Now Bob can create his own account!
	res, err := svc.Register(ctx, "bob@example.com", "password123!", "Bob Solo")
	if err != nil {
		t.Fatalf("expected registration to succeed after leaving team, got: %v", err)
	}
	if res.User.Role != domain.RoleOwner {
		t.Fatalf("expected new account to be owner, got: %s", res.User.Role)
	}
}

func TestAccountDeletion24HourGracePeriodAndPurge(t *testing.T) {
	userRepo := newMockUserRepoForAccountDeletion()
	sessRepo := &mockSessionRepoForAccountDeletion{}
	teamRepo := &mockTeamRepoForAccountDeletion{users: userRepo}

	hash, _ := crypto.HashPassword("securepass123")
	user := &domain.User{
		ID:           "usr_charlie",
		OrgID:        "org_charlie",
		Email:        "charlie@example.com",
		Name:         "Charlie",
		Role:         domain.RoleOwner,
		PasswordHash: hash,
	}
	userRepo.users[user.ID] = user

	svc := New(Repos{
		Users:    userRepo,
		Sessions: sessRepo,
		Team:     teamRepo,
	}, Config{JWTSecret: "test-secret-at-least-32-bytes-long!"}, nil)

	ctx := context.Background()

	// Schedule deletion
	scheduledAt, err := svc.ScheduleDeleteAccount(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to schedule deletion: %v", err)
	}
	if scheduledAt.Before(time.Now().Add(23 * time.Hour)) {
		t.Fatalf("scheduledAt should be ~24h in the future, got: %v", scheduledAt)
	}
	if len(sessRepo.revokedAll) == 0 || sessRepo.revokedAll[0] != user.ID {
		t.Fatalf("expected sessions to be revoked on scheduling deletion")
	}

	// Within 24h, user can cancel deletion
	err = svc.CancelDeleteAccount(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to cancel deletion: %v", err)
	}
	if user.ScheduledDeleteAt != nil {
		t.Fatalf("expected ScheduledDeleteAt to be cleared after cancellation")
	}

	// Schedule again
	_, _ = svc.ScheduleDeleteAccount(ctx, user.ID)

	// Simulate 24 hours passing
	past := time.Now().Add(-1 * time.Hour)
	user.ScheduledDeleteAt = &past

	// Now registering with charlie@example.com should auto-purge expired account and succeed!
	res, err := svc.Register(ctx, "charlie@example.com", "newpassword123!", "Charlie New")
	if err != nil {
		t.Fatalf("expected registration to succeed after 24h deletion, got: %v", err)
	}
	if res.User.ID == "usr_charlie" {
		t.Fatalf("expected a new user ID, got old user ID")
	}
}
