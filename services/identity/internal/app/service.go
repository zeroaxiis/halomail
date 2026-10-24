package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/zeroaxiis/halomail/services/identity/internal/crypto"
	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
	"github.com/redis/go-redis/v9"
)

// accessTokenTTL is the lifetime of an access JWT. Refresh tokens live longer
// (Config.RefreshTTL) and mint new access tokens.
const accessTokenTTL = 15 * time.Minute

// Config carries the auth knobs the service needs.
type Config struct {
	JWTSecret    string
	RefreshTTL   time.Duration
	APIKeyPrefix string
	WebURL       string
}

// Service implements the identity use cases over a set of repositories.
type Service struct {
	users          UserRepo
	sessions       SessionRepo
	apiKeys        APIKeyRepo
	audit          AuditRepo
	team           TeamRepo
	passwordResets PasswordResetRepo
	tokens         crypto.TokenIssuer
	cfg            Config
	redis          *redis.Client
	now            func() time.Time
}

func New(r Repos, cfg Config, redisClient *redis.Client) *Service {
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	if cfg.APIKeyPrefix == "" {
		cfg.APIKeyPrefix = "hl_"
	}
	return &Service{
		users:          r.Users,
		sessions:       r.Sessions,
		apiKeys:        r.APIKeys,
		audit:          r.Audit,
		team:           r.Team,
		passwordResets: r.PasswordResets,
		tokens:         crypto.NewTokenIssuer(cfg.JWTSecret),
		cfg:            cfg,
		redis:          redisClient,
		now:            time.Now,
	}
}

// IssuedSession is a freshly minted token pair. The plaintext refresh token
// exists only here — only its hash is stored.
type IssuedSession struct {
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
}

// AuthResult carries an account and, only after OTP verification, a session.
type AuthResult struct {
	User    *domain.User
	Session IssuedSession
}

// ---- Auth ----------------------------------------------------------------

func (s *Service) Register(ctx context.Context, email, password, name string) (*AuthResult, error) {
	email = normalizeEmail(email)
	if !validEmail(email) {
		return nil, errs.Invalid("a valid email is required")
	}
	if len(password) < 8 || len(password) > 1024 || len(name) > 200 {
		return nil, errs.Invalid("use an 8 to 1024 character password and a name up to 200 characters")
	}

	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, errs.Wrap(err, errs.KindInternal, "hash password")
	}

	now := s.now()
	org := &domain.Org{
		ID:        idgen.Prefixed("org_"),
		Name:      defaultOrgName(name, email),
		Slug:      idgen.Prefixed("org-"),
		CreatedAt: now,
	}
	user := &domain.User{
		ID:           idgen.Prefixed("usr_"),
		OrgID:        org.ID,
		Email:        email,
		Name:         name,
		Handle:       deriveHandle(email),
		Timezone:     "UTC",
		PasswordHash: hash,
		Role:         domain.RoleOwner,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.CreateOrgAndUser(ctx, org, user); err != nil {
		return nil, err // repo maps unique violation → Conflict
	}

	s.record(ctx, org.ID, user.ID, "user.registered", "user", user.ID)
	return &AuthResult{User: user}, nil
}

// Login rejects legacy password-only clients; VerifyLoginOTP issues new sessions.
func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	return nil, errs.Unauthorized("email verification required; request and verify a login code")
}

func (s *Service) checkCredentials(ctx context.Context, email, password string) (*domain.User, error) {
	if len(password) > 1024 || len(email) > 254 {
		return nil, errs.Unauthorized("invalid email or password")
	}
	user, err := s.users.GetUserByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errs.KindOf(err) == errs.KindNotFound {
			return nil, errs.Unauthorized("invalid email or password")
		}
		return nil, err
	}
	ok, err := crypto.VerifyPassword(user.PasswordHash, password)
	if err != nil || !ok {
		return nil, errs.Unauthorized("invalid email or password")
	}

	return user, nil
}

// Refresh rotates the session: the old refresh token is revoked and a new pair
// is issued.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (IssuedSession, error) {
	sess, err := s.sessions.GetByRefreshHash(ctx, crypto.SHA256Hex(refreshToken))
	if err != nil {
		if errs.KindOf(err) == errs.KindNotFound {
			return IssuedSession{}, errs.Unauthorized("invalid session")
		}
		return IssuedSession{}, err
	}
	if !sess.Active(s.now()) {
		return IssuedSession{}, errs.Unauthorized("session expired")
	}
	user, err := s.users.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return IssuedSession{}, err
	}
	if err = s.sessions.Consume(ctx, sess.ID); err != nil {
		return IssuedSession{}, err
	}
	return s.issueSession(ctx, user)
}

// Logout revokes the session for the given refresh token. It is idempotent.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	sess, err := s.sessions.GetByRefreshHash(ctx, crypto.SHA256Hex(refreshToken))
	if err != nil {
		if errs.KindOf(err) == errs.KindNotFound {
			return nil
		}
		return err
	}
	return s.sessions.Revoke(ctx, sess.ID)
}

// VerifyToken validates an access token and returns the principal.
func (s *Service) VerifyToken(_ context.Context, accessToken string) (userID, orgID string, err error) {
	claims, err := s.tokens.Parse(accessToken)
	if err != nil {
		return "", "", errs.Unauthorized("invalid or expired token")
	}
	return claims.Subject, claims.OrgID, nil
}

func (s *Service) issueSession(ctx context.Context, user *domain.User) (IssuedSession, error) {
	access, exp, err := s.tokens.Issue(user.ID, user.OrgID, accessTokenTTL)
	if err != nil {
		return IssuedSession{}, errs.Wrap(err, errs.KindInternal, "issue access token")
	}
	refresh, err := crypto.RandomToken(32)
	if err != nil {
		return IssuedSession{}, errs.Wrap(err, errs.KindInternal, "generate refresh token")
	}
	now := s.now()
	if err := s.sessions.Create(ctx, &domain.Session{
		ID:               idgen.Prefixed("ses_"),
		UserID:           user.ID,
		OrgID:            user.OrgID,
		RefreshTokenHash: crypto.SHA256Hex(refresh),
		ExpiresAt:        now.Add(s.cfg.RefreshTTL),
		CreatedAt:        now,
	}); err != nil {
		return IssuedSession{}, err
	}
	return IssuedSession{AccessToken: access, RefreshToken: refresh, AccessExpiresAt: exp}, nil
}

// ---- Users ---------------------------------------------------------------

func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.users.GetUserByID(ctx, id)
}

func (s *Service) GetUserByHandle(ctx context.Context, handle string) (*domain.User, error) {
	return s.users.GetUserByHandle(ctx, strings.ToLower(strings.TrimSpace(handle)))
}

func (s *Service) UpdateUser(ctx context.Context, userID, name, handle, avatarURL, timezone string) (*domain.User, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if name != "" {
		user.Name = name
	}
	if handle != "" {
		user.Handle = strings.ToLower(strings.TrimSpace(handle))
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if timezone != "" {
		user.Timezone = timezone
	}
	user.UpdatedAt = s.now()
	if err := s.users.UpdateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) ProcessSubscription(ctx context.Context, orgID, customerID, subID, tier, status string, cycleEnd time.Time) error {
	if status == "cancelled" || status == "halted" || status == "refunded" {
		tier = "free" // Downgrade on cancellation or refund
	}
	return s.users.UpdateOrgBilling(ctx, orgID, customerID, subID, tier, cycleEnd)
}

// ---- API keys ------------------------------------------------------------

func (s *Service) CreateAPIKey(ctx context.Context, userID, orgID, name string, scopes []string) (*domain.APIKey, string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "forms":
		scopes = []string{"forms:submit"}
	case "meetings":
		scopes = []string{"meetings:book"}
	default:
		return nil, "", errs.Invalid("choose forms or meetings for the access key")
	}
	raw, err := crypto.RandomToken(24) // 48 hex chars
	if err != nil {
		return nil, "", errs.Wrap(err, errs.KindInternal, "generate api key")
	}
	secret := s.cfg.APIKeyPrefix + name + "_" + raw

	// api_keys.scopes is NOT NULL; a nil slice would be written as NULL and
	// skip the column default.
	if scopes == nil {
		scopes = []string{}
	}

	now := s.now()
	key := &domain.APIKey{
		ID:         idgen.Prefixed("key_"),
		OrgID:      orgID,
		UserID:     userID,
		Name:       name,
		Prefix:     s.cfg.APIKeyPrefix + name,
		LastFour:   raw[len(raw)-4:],
		SecretHash: crypto.SHA256Hex(secret),
		Scopes:     scopes,
		CreatedAt:  now,
	}
	if err := s.apiKeys.Create(ctx, key); err != nil {
		return nil, "", err
	}
	s.record(ctx, orgID, userID, "apikey.created", "api_key", key.ID)
	return key, secret, nil
}

func (s *Service) ListAPIKeys(ctx context.Context, userID string) ([]domain.APIKey, error) {
	return s.apiKeys.ListByUser(ctx, userID)
}

func (s *Service) RevokeAPIKey(ctx context.Context, id, userID, orgID string) error {
	// Security fix: Invalidate cache when revoking
	keys, err := s.apiKeys.ListByUser(ctx, userID)
	if err == nil {
		for _, k := range keys {
			if k.ID == id {
				if s.redis != nil {
					_ = s.redis.Del(ctx, "apikey:"+k.SecretHash).Err()
				}
				break
			}
		}
	}

	if err := s.apiKeys.Revoke(ctx, id, userID); err != nil {
		return err
	}
	s.record(ctx, orgID, userID, "apikey.revoked", "api_key", id)
	return nil
}

func (s *Service) VerifyAPIKey(ctx context.Context, secret string) (userID, orgID string, scopes []string, err error) {
	hash := crypto.SHA256Hex(secret)
	cacheKey := "apikey:" + hash

	type cachedKey struct {
		UserID string   `json:"user_id"`
		OrgID  string   `json:"org_id"`
		Scopes []string `json:"scopes"`
	}

	if s.redis != nil {
		if val, err := s.redis.Get(ctx, cacheKey).Result(); err == nil {
			var ck cachedKey
			if json.Unmarshal([]byte(val), &ck) == nil {
				return ck.UserID, ck.OrgID, ck.Scopes, nil
			}
		}
	}

	key, err := s.apiKeys.GetBySecretHash(ctx, hash)
	if err != nil {
		if errs.KindOf(err) == errs.KindNotFound {
			return "", "", nil, errs.Unauthorized("invalid api key")
		}
		return "", "", nil, err
	}
	if key.Revoked {
		return "", "", nil, errs.Unauthorized("api key revoked")
	}
	_ = s.apiKeys.TouchLastUsed(ctx, key.ID, s.now())

	if s.redis != nil {
		ck := cachedKey{UserID: key.UserID, OrgID: key.OrgID, Scopes: key.Scopes}
		if b, err := json.Marshal(ck); err == nil {
			_ = s.redis.Set(ctx, cacheKey, b, 15*time.Minute).Err()
		}
	}

	return key.UserID, key.OrgID, key.Scopes, nil
}

// ---- Audit ---------------------------------------------------------------

func (s *Service) ListAudit(ctx context.Context, orgID string, limit, offset int) ([]domain.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.audit.ListByOrg(ctx, orgID, limit, offset)
}

// RecordAuditInput is the explicit (RPC) audit entry.
type RecordAuditInput struct {
	OrgID, ActorID, Action, TargetType, TargetID, Metadata, IP string
}

func (s *Service) RecordAudit(ctx context.Context, in RecordAuditInput) (string, error) {
	if in.Action == "" {
		return "", errs.Invalid("action is required")
	}
	l := &domain.AuditLog{
		ID:         idgen.Prefixed("aud_"),
		OrgID:      in.OrgID,
		ActorID:    in.ActorID,
		Action:     in.Action,
		TargetType: in.TargetType,
		TargetID:   in.TargetID,
		Metadata:   emptyJSON(in.Metadata),
		IP:         in.IP,
		CreatedAt:  s.now(),
	}
	if err := s.audit.Insert(ctx, l); err != nil {
		return "", err
	}
	return l.ID, nil
}

// record is a best-effort internal audit append (errors are swallowed).
func (s *Service) record(ctx context.Context, orgID, actorID, action, targetType, targetID string) {
	_ = s.audit.Insert(ctx, &domain.AuditLog{
		ID:         idgen.Prefixed("aud_"),
		OrgID:      orgID,
		ActorID:    actorID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   "{}",
		CreatedAt:  s.now(),
	})
}

// ---- helpers -------------------------------------------------------------

func normalizeEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

func validEmail(e string) bool {
	address, err := mail.ParseAddress(e)
	return err == nil && address.Address == e && len(e) <= 254 && !strings.ContainsAny(e, "\r\n")
}

func defaultOrgName(name, email string) string {
	if n := strings.TrimSpace(name); n != "" {
		return n + "'s workspace"
	}
	return localPart(email) + "'s workspace"
}

func deriveHandle(email string) string {
	base := sanitizeHandle(localPart(email))
	if base == "" {
		base = "user"
	}
	suffix, err := crypto.RandomToken(3)
	if err != nil {
		suffix = "x"
	}
	return base + "-" + suffix
}

func localPart(email string) string {
	if at := strings.IndexByte(email, '@'); at > 0 {
		return email[:at]
	}
	return email
}

func sanitizeHandle(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func emptyJSON(s string) string {
	if strings.TrimSpace(s) == "" {
		return "{}"
	}
	return s
}

func (s *Service) ProcessPayment(ctx context.Context, userID, orgID, orderID, paymentID, tier string, paidAt, cycleEnd time.Time) (*domain.User, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if orgID != "" && orgID != user.OrgID {
		return nil, errs.Forbidden("payment account mismatch")
	}
	if err := s.users.ApplyPayment(ctx, user.OrgID, orderID, paymentID, tier, paidAt, cycleEnd); err != nil {
		return nil, err
	}
	return user, nil
}

// ---- Team Management -----------------------------------------------------

type TeamData struct {
	Members         []domain.User
	Invitations     []domain.Invitation
	MaxSeats        int
	UsedSeats       int
	Tier            string
	CurrentUserRole string
}

func MaxSeatsForTier(tier string) int {
	switch strings.ToLower(tier) {
	case "pro":
		return 4
	case "business":
		return 12
	case "enterprise":
		return 9999
	default:
		return 0 // free plan has no team seats
	}
}

func (s *Service) ListTeamMembers(ctx context.Context, userID, orgID string) (*TeamData, error) {
	caller, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if caller.OrgID != orgID {
		return nil, errs.Forbidden("not a member of this organization")
	}

	tier, err := s.team.GetOrgTier(ctx, orgID)
	if err != nil {
		tier = "free"
	}
	maxSeats := MaxSeatsForTier(tier)

	members, err := s.team.ListMembersByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	invitations, err := s.team.ListInvitationsByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &TeamData{
		Members:         members,
		Invitations:     invitations,
		MaxSeats:        maxSeats,
		UsedSeats:       len(members) + len(invitations),
		Tier:            tier,
		CurrentUserRole: caller.Role,
	}, nil
}

func (s *Service) InviteTeamMember(ctx context.Context, callerID, orgID, email, role string) (*domain.Invitation, error) {
	caller, err := s.users.GetUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}
	if caller.OrgID != orgID {
		return nil, errs.Forbidden("not a member of this organization")
	}
	if caller.Role != domain.RoleOwner && caller.Role != domain.RoleAdmin {
		return nil, errs.Forbidden("only organization owners and admins can invite team members")
	}

	role = strings.ToLower(strings.TrimSpace(role))
	if role != domain.RoleAdmin && role != domain.RoleManager && role != domain.RoleViewer {
		return nil, errs.Invalid("invalid role: must be admin, manager, or viewer")
	}
	if caller.Role == domain.RoleAdmin && role == domain.RoleAdmin {
		return nil, errs.Forbidden("only the workspace owner can invite an administrator")
	}

	email = normalizeEmail(email)
	if !validEmail(email) {
		return nil, errs.Invalid("a valid email address is required")
	}

	tier, err := s.team.GetOrgTier(ctx, orgID)
	if err != nil {
		tier = "free"
	}
	maxSeats := MaxSeatsForTier(tier)
	if maxSeats <= 0 {
		return nil, errs.Forbidden("team management is not available on the Free plan. Upgrade to Pro (up to 4 seats) or Business (up to 12 seats) to invite teammates.")
	}

	usedSeats, err := s.team.GetMemberCount(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if usedSeats >= maxSeats {
		return nil, errs.Forbidden("seat limit reached for the %s plan (%d seats). Upgrade your plan to invite more team members.", tier, maxSeats)
	}

	existingMembers, err := s.team.ListMembersByOrg(ctx, orgID)
	if err == nil {
		for _, m := range existingMembers {
			if strings.EqualFold(m.Email, email) {
				return nil, errs.Conflict("this user is already a member of your workspace")
			}
		}
	}

	token, err := crypto.RandomToken(24)
	if err != nil {
		return nil, err
	}

	inv := &domain.Invitation{
		ID:            idgen.Prefixed("inv_"),
		OrgID:         orgID,
		Email:         email,
		Role:          role,
		Token:         token,
		InvitedBy:     caller.ID,
		InvitedByName: caller.Name,
		Status:        domain.InvitePending,
		ExpiresAt:     s.now().Add(7 * 24 * time.Hour),
		CreatedAt:     s.now(),
	}

	orgName, _ := s.team.GetOrgName(ctx, orgID)
	if strings.TrimSpace(orgName) == "" {
		orgName = "Workspace"
	}

	inviterName := strings.TrimSpace(caller.Name)
	if inviterName == "" {
		inviterName = caller.Email
	}

	webURL := strings.TrimRight(s.cfg.WebURL, "/")
	if webURL == "" {
		webURL = strings.TrimRight(os.Getenv("PUBLIC_WEB_URL"), "/")
	}
	if webURL == "" || (strings.Contains(webURL, "localhost") && (os.Getenv("APP_ENV") == "production" || os.Getenv("NODE_ENV") == "production")) {
		if os.Getenv("APP_ENV") == "production" || os.Getenv("NODE_ENV") == "production" {
			webURL = "https://halomail.zeroaxiis.tech"
		} else if webURL == "" {
			webURL = "http://localhost:3000"
		}
	}
	inviteLink := fmt.Sprintf("%s/invite?token=%s", webURL, token)

	roleLabel := "Team Member"
	switch role {
	case domain.RoleAdmin:
		roleLabel = "Administrator"
	case domain.RoleManager:
		roleLabel = "Manager"
	case domain.RoleViewer:
		roleLabel = "Viewer"
	}

	subject := fmt.Sprintf("%s invited you to join %s on HaloMail", inviterName, orgName)
	emailHTML := renderTeamInviteEmail(inviterName, orgName, roleLabel, inviteLink)
	outboxID := idgen.Prefixed("mail_")

	if err := s.team.CreateInvitation(ctx, inv, outboxID, emailHTML, subject); err != nil {
		return nil, err
	}

	s.record(ctx, orgID, caller.ID, "team.member.invited", "invitation", inv.ID)
	return inv, nil
}

func renderTeamInviteEmail(inviterName, orgName, roleLabel, inviteLink string) string {
	return fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>Invitation to join %[2]s</title></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f9fafb; padding: 40px 20px; color: #111827;">
  <div style="max-width: 520px; margin: 0 auto; background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 32px;">
    <h2 style="font-size: 20px; font-weight: 600; margin-top: 0; color: #111827;">Join %[2]s on HaloMail</h2>
    <p style="font-size: 15px; line-height: 1.5; color: #374151;">Hello,</p>
    <p style="font-size: 15px; line-height: 1.5; color: #374151;"><strong>%[1]s</strong> has invited you to join <strong>%[2]s</strong> as a <strong>%[3]s</strong> on HaloMail.</p>
    <div style="margin: 28px 0;">
      <a href="%[4]s" style="background-color: #111827; color: #ffffff; text-decoration: none; padding: 12px 24px; border-radius: 6px; font-weight: 500; font-size: 14px; display: inline-block;">Accept Invitation</a>
    </div>
    <p style="font-size: 13px; line-height: 1.5; color: #6b7280;">This invitation is valid for 7 days. If you weren't expecting this invitation, you can safely ignore this email.</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;" />
    <p style="font-size: 12px; color: #9ca3af; word-break: break-all;">Button not working? Copy and paste this URL into your browser:<br /><a href="%[4]s" style="color: #4b5563;">%[4]s</a></p>
  </div>
</body>
</html>`, inviterName, orgName, roleLabel, inviteLink)
}

func (s *Service) UpdateMemberRole(ctx context.Context, callerID, orgID, targetUserID, newRole string) (*domain.User, error) {
	caller, err := s.users.GetUserByID(ctx, callerID)
	if err != nil {
		return nil, err
	}
	if caller.OrgID != orgID {
		return nil, errs.Forbidden("not a member of this organization")
	}
	if caller.Role != domain.RoleOwner && caller.Role != domain.RoleAdmin {
		return nil, errs.Forbidden("only organization owners and admins can update member roles")
	}

	newRole = strings.ToLower(strings.TrimSpace(newRole))
	if newRole != domain.RoleAdmin && newRole != domain.RoleManager && newRole != domain.RoleViewer {
		return nil, errs.Invalid("invalid role: must be admin, manager, or viewer")
	}

	target, err := s.users.GetUserByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	if target.OrgID != orgID {
		return nil, errs.NotFound("member not found in this organization")
	}
	if target.Role == domain.RoleOwner {
		return nil, errs.Forbidden("cannot modify the workspace owner's role")
	}
	if caller.Role == domain.RoleAdmin {
		if target.Role == domain.RoleAdmin || newRole == domain.RoleAdmin {
			return nil, errs.Forbidden("only the workspace owner can promote or demote administrators")
		}
	}

	if err := s.team.UpdateUserRole(ctx, orgID, targetUserID, newRole); err != nil {
		return nil, err
	}

	target.Role = newRole
	s.record(ctx, orgID, caller.ID, "team.member.role_updated", "user", targetUserID)
	return target, nil
}

func (s *Service) RemoveTeamMember(ctx context.Context, callerID, orgID, targetUserID string) error {
	caller, err := s.users.GetUserByID(ctx, callerID)
	if err != nil {
		return err
	}
	if caller.OrgID != orgID {
		return errs.Forbidden("not a member of this organization")
	}
	if caller.Role != domain.RoleOwner && caller.Role != domain.RoleAdmin {
		return errs.Forbidden("only organization owners and admins can remove team members")
	}

	target, err := s.users.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if target.OrgID != orgID {
		return errs.NotFound("member not found in this organization")
	}
	if target.Role == domain.RoleOwner {
		return errs.Forbidden("the workspace owner cannot be removed")
	}
	if caller.Role == domain.RoleAdmin && target.Role == domain.RoleAdmin {
		return errs.Forbidden("only the workspace owner can remove an administrator")
	}

	if err := s.team.RemoveUserFromOrg(ctx, orgID, targetUserID); err != nil {
		return err
	}

	s.record(ctx, orgID, caller.ID, "team.member.removed", "user", targetUserID)
	return nil
}

func (s *Service) RevokeInvitation(ctx context.Context, callerID, orgID, invitationID string) error {
	caller, err := s.users.GetUserByID(ctx, callerID)
	if err != nil {
		return err
	}
	if caller.OrgID != orgID {
		return errs.Forbidden("not a member of this organization")
	}
	if caller.Role != domain.RoleOwner && caller.Role != domain.RoleAdmin {
		return errs.Forbidden("only organization owners and admins can revoke invitations")
	}

	if err := s.team.RevokeInvitation(ctx, orgID, invitationID); err != nil {
		return err
	}

	s.record(ctx, orgID, caller.ID, "team.invitation.revoked", "invitation", invitationID)
	return nil
}

func (s *Service) AcceptInvitation(ctx context.Context, token string, currentUserID string) (*domain.User, IssuedSession, error) {
	inv, err := s.team.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, IssuedSession{}, err
	}
	if inv.Status != domain.InvitePending {
		return nil, IssuedSession{}, errs.Invalid("this invitation has already been processed or revoked")
	}
	if s.now().After(inv.ExpiresAt) {
		return nil, IssuedSession{}, errs.Invalid("this invitation has expired")
	}

	var user *domain.User
	if currentUserID != "" {
		user, err = s.users.GetUserByID(ctx, currentUserID)
		if err != nil {
			return nil, IssuedSession{}, err
		}
		if !strings.EqualFold(user.Email, inv.Email) {
			return nil, IssuedSession{}, errs.Forbidden("you are signed in as %s, but this invitation was sent to %s. Please switch accounts.", user.Email, inv.Email)
		}
	} else {
		user, err = s.users.GetUserByEmail(ctx, inv.Email)
		if err != nil {
			return nil, IssuedSession{}, errs.NotFound("no account found for %s. Please register or sign in first, then accept the invite.", inv.Email)
		}
	}

	if err := s.team.AcceptInvitation(ctx, inv, user); err != nil {
		return nil, IssuedSession{}, err
	}

	user.OrgID = inv.OrgID
	user.Role = inv.Role

	session, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, IssuedSession{}, err
	}

	s.record(ctx, inv.OrgID, user.ID, "team.invitation.accepted", "invitation", inv.ID)
	return user, session, nil
}

