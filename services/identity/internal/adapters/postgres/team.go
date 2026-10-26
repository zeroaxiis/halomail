package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
	"github.com/zeroaxiis/halomail/services/shared/outbox"
)

type Team struct {
	pool *pgxpool.Pool
}

func NewTeam(pool *pgxpool.Pool) *Team {
	return &Team{pool: pool}
}

func (r *Team) ListMembersByOrg(ctx context.Context, orgID string) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+userColumns+` FROM users WHERE org_id = $1 ORDER BY created_at ASC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Handle, &u.AvatarURL,
			&u.Timezone, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &u.Role,
			&u.DeletionScheduledAt, &u.ScheduledDeleteAt); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	return members, rows.Err()
}

func (r *Team) GetOrgTier(ctx context.Context, orgID string) (string, error) {
	var tier string
	err := r.pool.QueryRow(ctx,
		`SELECT CASE WHEN billing_cycle_end IS NOT NULL AND billing_cycle_end <= now() THEN 'free' ELSE tier END
		 FROM orgs WHERE id = $1`, orgID).Scan(&tier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "free", nil
		}
		return "free", err
	}
	if tier == "" {
		tier = "free"
	}
	return tier, nil
}

func (r *Team) GetOrgName(ctx context.Context, orgID string) (string, error) {
	var name string
	err := r.pool.QueryRow(ctx, `SELECT name FROM orgs WHERE id = $1`, orgID).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return name, nil
}

func (r *Team) GetMemberCount(ctx context.Context, orgID string) (int, error) {
	var userCount, inviteCount int
	err := r.pool.QueryRow(ctx,
		`SELECT 
			(SELECT count(*) FROM users WHERE org_id = $1),
			(SELECT count(*) FROM org_invitations WHERE org_id = $1 AND status = 'pending' AND expires_at > now())`,
		orgID).Scan(&userCount, &inviteCount)
	if err != nil {
		return 0, err
	}
	return userCount + inviteCount, nil
}

func (r *Team) CreateInvitation(ctx context.Context, inv *domain.Invitation, outboxID, emailHTML, subject string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Invalidate any existing pending invitations for this email in this org
	if _, err := tx.Exec(ctx,
		`UPDATE org_invitations SET status = 'revoked' WHERE org_id = $1 AND lower(email) = lower($2) AND status = 'pending'`,
		inv.OrgID, inv.Email,
	); err != nil {
		return mapWrite(err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO org_invitations (id, org_id, email, role, token, invited_by, status, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		inv.ID, inv.OrgID, inv.Email, inv.Role, inv.Token, inv.InvitedBy, inv.Status, inv.ExpiresAt, inv.CreatedAt,
	); err != nil {
		return mapWrite(err)
	}

	if outboxID != "" && emailHTML != "" {
		if subject == "" {
			subject = "You've been invited to join a team on HaloMail"
		}
		if err := outbox.Queue(ctx, tx, outboxID, inv.Email, "", subject, emailHTML); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *Team) ListInvitationsByOrg(ctx context.Context, orgID string) ([]domain.Invitation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT i.id, i.org_id, i.email, i.role, i.token, i.invited_by, COALESCE(NULLIF(u.name, ''), u.email, ''), i.status, i.expires_at, i.created_at
		 FROM org_invitations i
		 LEFT JOIN users u ON i.invited_by = u.id
		 WHERE i.org_id = $1 AND i.status = 'pending' AND i.expires_at > now()
		 ORDER BY i.created_at DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []domain.Invitation
	for rows.Next() {
		var inv domain.Invitation
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.Token,
			&inv.InvitedBy, &inv.InvitedByName, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, rows.Err()
}

func (r *Team) GetInvitationByToken(ctx context.Context, token string) (*domain.Invitation, error) {
	var inv domain.Invitation
	err := r.pool.QueryRow(ctx,
		`SELECT i.id, i.org_id, i.email, i.role, i.token, i.invited_by, COALESCE(NULLIF(u.name, ''), u.email, ''), i.status, i.expires_at, i.created_at
		 FROM org_invitations i
		 LEFT JOIN users u ON i.invited_by = u.id
		 WHERE i.token = $1`, token).Scan(&inv.ID, &inv.OrgID, &inv.Email, &inv.Role, &inv.Token,
		&inv.InvitedBy, &inv.InvitedByName, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.NotFound("invitation not found")
		}
		return nil, err
	}
	return &inv, nil
}

func (r *Team) RevokeInvitation(ctx context.Context, orgID, invitationID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE org_invitations SET status = 'revoked' WHERE id = $1 AND org_id = $2 AND status = 'pending'`,
		invitationID, orgID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("invitation not found or already processed")
	}
	return nil
}

func (r *Team) UpdateUserRole(ctx context.Context, orgID, userID, newRole string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET role = $3, updated_at = now() WHERE id = $2 AND org_id = $1`,
		orgID, userID, newRole)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.NotFound("user not found in this organization")
	}
	return nil
}

func (r *Team) RemoveUserFromOrg(ctx context.Context, orgID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var u domain.User
	err = tx.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1 AND org_id = $2`, userID, orgID).
		Scan(&u.ID, &u.OrgID, &u.Email, &u.Name, &u.Handle, &u.AvatarURL, &u.Timezone, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt, &u.Role, &u.DeletionScheduledAt, &u.ScheduledDeleteAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.NotFound("user not found in this organization")
		}
		return err
	}

	if u.Role == domain.RoleOwner {
		return errs.Forbidden("the workspace owner cannot be removed")
	}

	// Create a new individual workspace for the evicted user
	newOrgID := idgen.Prefixed("org_")
	newSlug := idgen.Prefixed("org-")
	_, err = tx.Exec(ctx, `INSERT INTO orgs (id, name, slug) VALUES ($1, $2, $3)`, newOrgID, u.Name+"'s Workspace", newSlug)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE users SET org_id = $1, role = 'owner', updated_at = now() WHERE id = $2`, newOrgID, userID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Team) AcceptInvitation(ctx context.Context, inv *domain.Invitation, user *domain.User) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update user to belong to the new org with the invited role
	_, err = tx.Exec(ctx,
		`UPDATE users SET org_id = $1, role = $2, updated_at = now() WHERE id = $3`,
		inv.OrgID, inv.Role, user.ID)
	if err != nil {
		return err
	}

	// Mark invitation as accepted
	_, err = tx.Exec(ctx,
		`UPDATE org_invitations SET status = 'accepted' WHERE id = $1`,
		inv.ID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Team) LeaveTeam(ctx context.Context, orgID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var role string
	err = tx.QueryRow(ctx, `SELECT role FROM users WHERE id = $1 AND org_id = $2`, userID, orgID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.NotFound("user not found in this organization")
		}
		return err
	}

	if role == domain.RoleOwner {
		return errs.Forbidden("the workspace owner cannot leave the workspace; transfer ownership or delete the organization instead")
	}

	// Delete the member user record, which frees their email so they can create a new account or accept another invite
	if _, err := tx.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
