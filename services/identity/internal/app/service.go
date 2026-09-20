package app

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/crypto"
	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/idgen"
)

// accessTokenTTL is the lifetime of an access JWT. Refresh tokens live longer
// (Config.RefreshTTL) and mint new access tokens.
const accessTokenTTL = 15 * time.Minute

// Config carries the auth knobs the service needs.
type Config struct {
	JWTSecret    string
	RefreshTTL   time.Duration
	APIKeyPrefix string
}

// Service implements the identity use cases over a set of repositories.
type Service struct {
	users    UserRepo
	sessions SessionRepo
	apiKeys  APIKeyRepo
	audit    AuditRepo
	tokens   crypto.TokenIssuer
	cfg      Config
	now      func() time.Time
}

func New(r Repos, cfg Config) *Service {
	if cfg.RefreshTTL <= 0 {
		cfg.RefreshTTL = 30 * 24 * time.Hour
	}
	if cfg.APIKeyPrefix == "" {
		cfg.APIKeyPrefix = "hl_"
	}
	return &Service{
		users:    r.Users,
		sessions: r.Sessions,
		apiKeys:  r.APIKeys,
		audit:    r.Audit,
		tokens:   crypto.NewTokenIssuer(cfg.JWTSecret),
		cfg:      cfg,
		now:      time.Now,
	}
}

// IssuedSession is a freshly minted token pair. The plaintext refresh token
// exists only here — only its hash is stored.
type IssuedSession struct {
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
}

// AuthResult is the output of Register/Login.
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
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.CreateOrgAndUser(ctx, org, user); err != nil {
		return nil, err // repo maps unique violation → Conflict
	}

	sess, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
	}
	s.record(ctx, org.ID, user.ID, "user.registered", "user", user.ID)
	return &AuthResult{User: user, Session: sess}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
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

	sess, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
	}
	s.record(ctx, user.OrgID, user.ID, "user.login", "user", user.ID)
	return &AuthResult{User: user, Session: sess}, nil
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
	if err := s.apiKeys.Revoke(ctx, id, userID); err != nil {
		return err
	}
	s.record(ctx, orgID, userID, "apikey.revoked", "api_key", id)
	return nil
}

func (s *Service) VerifyAPIKey(ctx context.Context, secret string) (userID, orgID string, scopes []string, err error) {
	key, err := s.apiKeys.GetBySecretHash(ctx, crypto.SHA256Hex(secret))
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
