package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zeroaxiis/halomail/services/identity/internal/adapters/postgres"
	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
)

func TestTeamCreateInvitationQueuesOutbox(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/halomail?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping postgres integration test: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("postgres ping failed: %v", err)
	}

	team := postgres.NewTeam(pool)
	users := postgres.NewUsers(pool)

	orgID := idgen.Prefixed("org_")
	userID := idgen.Prefixed("usr_")
	now := time.Now()

	org := &domain.Org{ID: orgID, Name: "Test Org", Slug: idgen.Prefixed("slug-"), CreatedAt: now}
	user := &domain.User{ID: userID, OrgID: orgID, Email: idgen.Prefixed("owner_") + "@test.com", Role: domain.RoleOwner, CreatedAt: now, UpdatedAt: now}

	if err := users.CreateOrgAndUser(ctx, org, user); err != nil {
		t.Fatalf("failed to create org/user: %v", err)
	}

	invEmail := idgen.Prefixed("member_") + "@test.com"
	inv := &domain.Invitation{
		ID:        idgen.Prefixed("inv_"),
		OrgID:     orgID,
		Email:     invEmail,
		Role:      "manager",
		Token:     idgen.Prefixed("tok_"),
		InvitedBy: userID,
		Status:    "pending",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
	}

	outboxID := idgen.Prefixed("mail_")
	subject := "You have been invited to Test Org"
	html := "<p>Click here to join</p>"

	if err := team.CreateInvitation(ctx, inv, outboxID, html, subject); err != nil {
		t.Fatalf("CreateInvitation failed: %v", err)
	}

	// Verify outbox entry
	var recipient, gotSub, gotHTML string
	err = pool.QueryRow(ctx, `SELECT recipient, subject, html FROM mail_outbox WHERE id = $1`, outboxID).
		Scan(&recipient, &gotSub, &gotHTML)
	if err != nil {
		t.Fatalf("expected mail_outbox record: %v", err)
	}

	if recipient != invEmail || gotSub != subject || gotHTML != html {
		t.Fatalf("mismatched outbox record: recipient=%s sub=%s", recipient, gotSub)
	}

	// Verify Invitation can be fetched by token
	fetched, err := team.GetInvitationByToken(ctx, inv.Token)
	if err != nil {
		t.Fatalf("failed to fetch by token: %v", err)
	}
	if fetched.ID != inv.ID || fetched.Token != inv.Token {
		t.Fatalf("mismatched fetched invitation: %+v", fetched)
	}
}
