package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
)

type mockTeamRepo struct {
	TeamRepo
	tier         string
	members      []domain.User
	invitations  []domain.Invitation
	createdInv   *domain.Invitation
	memberCounts int
	revokedID    string
	removedID    string
	updatedRole  string
}

func (m *mockTeamRepo) GetOrgTier(ctx context.Context, orgID string) (string, error) {
	if m.tier == "" {
		return "free", nil
	}
	return m.tier, nil
}

func (m *mockTeamRepo) GetMemberCount(ctx context.Context, orgID string) (int, error) {
	return m.memberCounts, nil
}

func (m *mockTeamRepo) ListMembersByOrg(ctx context.Context, orgID string) ([]domain.User, error) {
	return m.members, nil
}

func (m *mockTeamRepo) ListInvitationsByOrg(ctx context.Context, orgID string) ([]domain.Invitation, error) {
	return m.invitations, nil
}

func (m *mockTeamRepo) CreateInvitation(ctx context.Context, inv *domain.Invitation) error {
	m.createdInv = inv
	return nil
}

func (m *mockTeamRepo) GetInvitationByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	for i := range m.invitations {
		if m.invitations[i].Token == token {
			return &m.invitations[i], nil
		}
	}
	if m.createdInv != nil && m.createdInv.Token == token {
		return m.createdInv, nil
	}
	return nil, errs.NotFound("invitation not found")
}

func (m *mockTeamRepo) RevokeInvitation(ctx context.Context, orgID, invitationID string) error {
	m.revokedID = invitationID
	return nil
}

func (m *mockTeamRepo) UpdateUserRole(ctx context.Context, orgID, userID, newRole string) error {
	m.updatedRole = newRole
	return nil
}

func (m *mockTeamRepo) RemoveUserFromOrg(ctx context.Context, orgID, userID string) error {
	m.removedID = userID
	return nil
}

func (m *mockTeamRepo) AcceptInvitation(ctx context.Context, inv *domain.Invitation, user *domain.User) error {
	inv.Status = domain.InviteAccepted
	user.OrgID = inv.OrgID
	user.Role = inv.Role
	return nil
}

type mockUserRepoForTeam struct {
	UserRepo
	users map[string]*domain.User
}

func (m *mockUserRepoForTeam) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errs.NotFound("user not found")
	}
	return u, nil
}

func (m *mockUserRepoForTeam) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) {
			return u, nil
		}
	}
	return nil, errs.NotFound("user not found")
}

func TestTeamManagementTierLimits(t *testing.T) {
	ctx := context.Background()

	owner := &domain.User{
		ID:    "usr_owner",
		OrgID: "org_1",
		Email: "owner@company.com",
		Name:  "Owner",
		Role:  domain.RoleOwner,
	}

	userMap := map[string]*domain.User{
		owner.ID: owner,
	}

	// 1. Free plan cannot invite
	t.Run("free plan cannot invite teammates", func(t *testing.T) {
		teamRepo := &mockTeamRepo{tier: "free", memberCounts: 1}
		svc := New(Repos{
			Users: &mockUserRepoForTeam{users: userMap},
			Team:  teamRepo,
			Audit: auditRepository{},
		}, Config{JWTSecret: strings.Repeat("s", 32)}, nil)

		_, err := svc.InviteTeamMember(ctx, owner.ID, owner.OrgID, "colleague@company.com", domain.RoleManager)
		if err == nil || !strings.Contains(err.Error(), "Free plan") {
			t.Fatalf("expected Free plan error, got %v", err)
		}
	})

	// 2. Pro plan allows up to 4 members
	t.Run("pro plan allows up to 4 members and rejects 5th", func(t *testing.T) {
		teamRepo := &mockTeamRepo{tier: "pro", memberCounts: 3}
		svc := New(Repos{
			Users: &mockUserRepoForTeam{users: userMap},
			Team:  teamRepo,
			Audit: auditRepository{},
		}, Config{JWTSecret: strings.Repeat("s", 32)}, nil)

		inv, err := svc.InviteTeamMember(ctx, owner.ID, owner.OrgID, "teammate4@company.com", domain.RoleManager)
		if err != nil {
			t.Fatalf("unexpected error inviting 4th member: %v", err)
		}
		if inv.Role != domain.RoleManager || inv.Email != "teammate4@company.com" {
			t.Fatalf("invalid invitation created: %+v", inv)
		}

		// When 4 seats are already taken
		teamRepo.memberCounts = 4
		_, err = svc.InviteTeamMember(ctx, owner.ID, owner.OrgID, "teammate5@company.com", domain.RoleViewer)
		if err == nil || !strings.Contains(err.Error(), "seat limit reached") {
			t.Fatalf("expected seat limit error, got %v", err)
		}
	})

	// 3. Business plan allows up to 12 members
	t.Run("business plan allows up to 12 members", func(t *testing.T) {
		teamRepo := &mockTeamRepo{tier: "business", memberCounts: 11}
		svc := New(Repos{
			Users: &mockUserRepoForTeam{users: userMap},
			Team:  teamRepo,
			Audit: auditRepository{},
		}, Config{JWTSecret: strings.Repeat("s", 32)}, nil)

		_, err := svc.InviteTeamMember(ctx, owner.ID, owner.OrgID, "teammate12@company.com", domain.RoleAdmin)
		if err != nil {
			t.Fatalf("unexpected error on 12th member: %v", err)
		}

		teamRepo.memberCounts = 12
		_, err = svc.InviteTeamMember(ctx, owner.ID, owner.OrgID, "teammate13@company.com", domain.RoleViewer)
		if err == nil || !strings.Contains(err.Error(), "seat limit reached") {
			t.Fatalf("expected seat limit error, got %v", err)
		}
	})
}

func TestTeamRolePermissions(t *testing.T) {
	ctx := context.Background()

	owner := &domain.User{ID: "usr_owner", OrgID: "org_1", Email: "owner@company.com", Role: domain.RoleOwner}
	admin := &domain.User{ID: "usr_admin", OrgID: "org_1", Email: "admin@company.com", Role: domain.RoleAdmin}
	manager := &domain.User{ID: "usr_manager", OrgID: "org_1", Email: "manager@company.com", Role: domain.RoleManager}
	viewer := &domain.User{ID: "usr_viewer", OrgID: "org_1", Email: "viewer@company.com", Role: domain.RoleViewer}

	userMap := map[string]*domain.User{
		owner.ID:   owner,
		admin.ID:   admin,
		manager.ID: manager,
		viewer.ID:  viewer,
	}

	teamRepo := &mockTeamRepo{tier: "pro", memberCounts: 2}
	svc := New(Repos{
		Users: &mockUserRepoForTeam{users: userMap},
		Team:  teamRepo,
		Audit: auditRepository{},
	}, Config{JWTSecret: strings.Repeat("s", 32)}, nil)

	// Manager and Viewer cannot invite
	_, err := svc.InviteTeamMember(ctx, manager.ID, manager.OrgID, "new@company.com", domain.RoleViewer)
	if err == nil || !strings.Contains(err.Error(), "owners and admins") {
		t.Fatalf("manager should not be able to invite, got %v", err)
	}

	_, err = svc.InviteTeamMember(ctx, viewer.ID, viewer.OrgID, "new@company.com", domain.RoleViewer)
	if err == nil || !strings.Contains(err.Error(), "owners and admins") {
		t.Fatalf("viewer should not be able to invite, got %v", err)
	}

	// Admin cannot invite another Admin
	_, err = svc.InviteTeamMember(ctx, admin.ID, admin.OrgID, "newadmin@company.com", domain.RoleAdmin)
	if err == nil || !strings.Contains(err.Error(), "workspace owner can invite an administrator") {
		t.Fatalf("admin should not be able to invite admin, got %v", err)
	}

	// Owner can invite Admin
	_, err = svc.InviteTeamMember(ctx, owner.ID, owner.OrgID, "newadmin@company.com", domain.RoleAdmin)
	if err != nil {
		t.Fatalf("owner should be able to invite admin: %v", err)
	}

	// Admin cannot remove Admin
	err = svc.RemoveTeamMember(ctx, admin.ID, admin.OrgID, admin.ID)
	// (trying to remove another admin)
	otherAdmin := &domain.User{ID: "usr_admin2", OrgID: "org_1", Email: "admin2@company.com", Role: domain.RoleAdmin}
	userMap[otherAdmin.ID] = otherAdmin
	err = svc.RemoveTeamMember(ctx, admin.ID, admin.OrgID, otherAdmin.ID)
	if err == nil || !strings.Contains(err.Error(), "workspace owner can remove an administrator") {
		t.Fatalf("admin should not be able to remove another admin, got %v", err)
	}

	// Owner cannot be removed
	err = svc.RemoveTeamMember(ctx, admin.ID, admin.OrgID, owner.ID)
	if err == nil || !strings.Contains(err.Error(), "workspace owner cannot be removed") {
		t.Fatalf("owner cannot be removed, got %v", err)
	}
}

func TestAcceptInvitation(t *testing.T) {
	ctx := context.Background()
	owner := &domain.User{ID: "usr_owner", OrgID: "org_1", Email: "owner@company.com", Role: domain.RoleOwner}
	invitee := &domain.User{ID: "usr_invitee", OrgID: "org_solo", Email: "invitee@company.com", Role: domain.RoleOwner}

	userMap := map[string]*domain.User{
		owner.ID:   owner,
		invitee.ID: invitee,
	}

	invitation := domain.Invitation{
		ID:        "inv_1",
		OrgID:     "org_1",
		Email:     "invitee@company.com",
		Role:      domain.RoleManager,
		Token:     "secret-token-123",
		Status:    domain.InvitePending,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	teamRepo := &mockTeamRepo{
		tier:        "pro",
		invitations: []domain.Invitation{invitation},
	}

	svc := New(Repos{
		Users:    &mockUserRepoForTeam{users: userMap},
		Team:     teamRepo,
		Sessions: &mockSessionRepo{},
		Audit:    auditRepository{},
	}, Config{JWTSecret: strings.Repeat("s", 32)}, nil)

	user, session, err := svc.AcceptInvitation(ctx, "secret-token-123", invitee.ID)
	if err != nil {
		t.Fatalf("failed to accept invitation: %v", err)
	}
	if user.OrgID != "org_1" || user.Role != domain.RoleManager {
		t.Fatalf("user was not joined with manager role in org_1: %+v", user)
	}
	if session.AccessToken == "" {
		t.Fatal("session was not issued upon accepting invite")
	}
}

type mockSessionRepo struct {
	SessionRepo
}

func (mockSessionRepo) Create(context.Context, *domain.Session) error { return nil }
