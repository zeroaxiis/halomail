package app

import (
	"context"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/zeroaxiis/halomail/services/identity/internal/crypto"
	"github.com/zeroaxiis/halomail/services/identity/internal/domain"
	"github.com/zeroaxiis/halomail/services/shared/errs"
	"github.com/zeroaxiis/halomail/services/shared/idgen"
)

const resetTokenTTL = 1 * time.Hour

func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return errs.Invalid("invalid email address")
	}

	// Rate limiting: if redis is enabled, limit requests per email to 5 in 15 minutes
	if s.redis != nil {
		key := "auth:pwd_reset:rate:" + crypto.SHA256Hex(email)
		count, err := s.redis.Incr(ctx, key).Result()
		if err == nil {
			if count == 1 {
				s.redis.Expire(ctx, key, 15*time.Minute)
			}
			if count > 5 {
				return errs.RateLimited("too many password reset requests; please wait 15 minutes")
			}
		}
	}

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		// Prevent user enumeration: return success even if user not found
		return nil
	}

	rawToken, err := crypto.RandomToken(32)
	if err != nil {
		return errs.Internal("generate reset token")
	}

	tokenHash := crypto.SHA256Hex(rawToken)
	resetID := idgen.Prefixed("rst_")
	outboxID := idgen.Prefixed("mail_")

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
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", webURL, rawToken)

	emailHTML := renderPasswordResetEmail(user.Name, resetLink)

	reset := &domain.PasswordReset{
		ID:        resetID,
		UserID:    user.ID,
		Email:     user.Email,
		TokenHash: tokenHash,
		ExpiresAt: s.now().Add(resetTokenTTL),
		CreatedAt: s.now(),
	}

	if s.passwordResets != nil {
		if err := s.passwordResets.CreateReset(ctx, reset, outboxID, emailHTML); err != nil {
			return err
		}
	}

	s.record(ctx, user.OrgID, user.ID, "user.forgot_password_requested", "user", user.ID)
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" || len(rawToken) < 32 {
		return errs.Invalid("invalid or missing reset token")
	}
	if len(newPassword) < 8 {
		return errs.Invalid("password must be at least 8 characters")
	}
	if len(newPassword) > 1024 {
		return errs.Invalid("password is too long")
	}

	if s.passwordResets == nil {
		return errs.Internal("password reset repository not configured")
	}

	tokenHash := crypto.SHA256Hex(rawToken)
	reset, err := s.passwordResets.GetResetByTokenHash(ctx, tokenHash)
	if err != nil {
		return errs.Unauthorized("invalid or expired password reset link")
	}

	if !reset.Valid(s.now()) {
		return errs.Unauthorized("password reset link has expired or has already been used")
	}

	newHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return errs.Internal("hash password")
	}

	if err := s.passwordResets.CompleteReset(ctx, reset.ID, reset.UserID, newHash); err != nil {
		return err
	}

	s.record(ctx, "", reset.UserID, "user.password_reset", "user", reset.UserID)
	return nil
}

func renderPasswordResetEmail(name, resetLink string) string {
	greeting := "Hello,"
	if strings.TrimSpace(name) != "" {
		greeting = fmt.Sprintf("Hello %s,", name)
	}
	return fmt.Sprintf(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>Reset your password</title></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f9fafb; padding: 40px 20px; color: #111827;">
  <div style="max-width: 520px; margin: 0 auto; background: #ffffff; border: 1px solid #e5e7eb; border-radius: 8px; padding: 32px;">
    <h2 style="font-size: 20px; font-weight: 600; margin-top: 0; color: #111827;">Reset your password</h2>
    <p style="font-size: 15px; line-height: 1.5; color: #374151;">%s</p>
    <p style="font-size: 15px; line-height: 1.5; color: #374151;">We received a request to reset the password for your HaloMail account. Click the button below to choose a new password:</p>
    <div style="margin: 28px 0;">
      <a href="%s" style="background-color: #111827; color: #ffffff; text-decoration: none; padding: 12px 24px; border-radius: 6px; font-weight: 500; font-size: 14px; display: inline-block;">Reset Password</a>
    </div>
    <p style="font-size: 13px; line-height: 1.5; color: #6b7280;">This link is valid for 1 hour. If you didn't request a password reset, you can safely ignore this email — your password will not change.</p>
    <hr style="border: none; border-top: 1px solid #e5e7eb; margin: 24px 0;" />
    <p style="font-size: 12px; color: #9ca3af; word-break: break-all;">Button not working? Copy and paste this URL into your browser:<br /><a href="%s" style="color: #4b5563;">%s</a></p>
  </div>
</body>
</html>`, greeting, resetLink, resetLink, resetLink)
}
