package app

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/crypto"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/redis/go-redis/v9"
)

const otpTTL = 10 * time.Minute

var challengePattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
var codePattern = regexp.MustCompile(`^[0-9]{6}$`)

// OTPDelivery is returned only to the authenticated mail-delivery server.
type OTPDelivery struct {
	ChallengeID string `json:"challengeId"`
	Email       string `json:"email"`
	Code        string `json:"code"`
	ExpiresIn   int    `json:"expiresIn"`
}

var limitOTP = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then redis.call('EXPIRE', KEYS[1], 600) end
return count
`)

var storeOTP = redis.NewScript(`
local old = redis.call('GET', KEYS[2])
if old then redis.call('DEL', 'auth:otp:challenge:' .. old) end
redis.call('HSET', KEYS[1], 'hash', ARGV[1], 'user', ARGV[2], 'attempts', 0)
redis.call('EXPIRE', KEYS[1], 600)
redis.call('SET', KEYS[2], ARGV[3], 'EX', 600)
return 1
`)

var consumeOTP = redis.NewScript(`
local expected = redis.call('HGET', KEYS[1], 'hash')
if not expected then return '' end
local attempts = redis.call('HINCRBY', KEYS[1], 'attempts', 1)
if expected ~= ARGV[1] then
  if attempts >= 5 then redis.call('DEL', KEYS[1]) end
  return ''
end
local user = redis.call('HGET', KEYS[1], 'user')
redis.call('DEL', KEYS[1])
return user
`)

func (s *Service) otpHash(challenge, code string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.JWTSecret))
	mac.Write([]byte("login-otp:" + challenge + ":" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Service) RequestLoginOTP(ctx context.Context, email, password string) (*OTPDelivery, error) {
	if s.redis == nil {
		return nil, errs.Internal("login verification unavailable")
	}
	email = normalizeEmail(email)
	if !validEmail(email) || len(password) > 1024 {
		return nil, errs.Unauthorized("invalid email or password")
	}
	// Bound password checks and mail sends per account, including across resends.
	key := "auth:otp:request:" + crypto.SHA256Hex(email)
	count, err := limitOTP.Run(ctx, s.redis, []string{key}).Int()
	if err != nil {
		return nil, errs.Internal("login verification unavailable")
	}
	if count > 10 {
		return nil, errs.RateLimited("too many login requests; try again in 10 minutes")
	}
	user, err := s.checkCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}
	allowed, err := s.redis.SetNX(ctx, key+":cooldown", "1", time.Minute).Result()
	if err != nil {
		return nil, errs.Internal("login verification unavailable")
	}
	if !allowed {
		return nil, errs.RateLimited("wait 60 seconds before requesting another code")
	}
	challenge, err := crypto.RandomToken(32)
	if err != nil {
		return nil, errs.Internal("generate login challenge")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return nil, errs.Internal("generate login code")
	}
	code := fmt.Sprintf("%06d", n.Int64())
	_, err = storeOTP.Run(ctx, s.redis, []string{"auth:otp:challenge:" + challenge, "auth:otp:user:" + user.ID}, s.otpHash(challenge, code), user.ID, challenge).Result()
	if err != nil {
		return nil, errs.Internal("store login code")
	}
	return &OTPDelivery{ChallengeID: challenge, Email: user.Email, Code: code, ExpiresIn: int(otpTTL.Seconds())}, nil
}

func (s *Service) CancelLoginOTP(ctx context.Context, challenge string) error {
	if s.redis == nil {
		return errs.Internal("login verification unavailable")
	}
	if !challengePattern.MatchString(challenge) {
		return errs.Invalid("invalid challenge")
	}
	return s.redis.Del(ctx, "auth:otp:challenge:"+challenge).Err()
}

func (s *Service) VerifyLoginOTP(ctx context.Context, challenge, code string) (*AuthResult, error) {
	if s.redis == nil {
		return nil, errs.Internal("login verification unavailable")
	}
	if !challengePattern.MatchString(challenge) || !codePattern.MatchString(code) {
		return nil, errs.Unauthorized("invalid or expired code; request a new code if needed")
	}
	// Check, count attempts, and consume atomically so concurrent requests cannot reuse a code.
	userID, err := consumeOTP.Run(ctx, s.redis, []string{"auth:otp:challenge:" + challenge}, s.otpHash(challenge, code)).Text()
	if err != nil {
		return nil, errs.Internal("login verification unavailable")
	}
	if userID == "" {
		return nil, errs.Unauthorized("invalid or expired code; request a new code if needed")
	}
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	session, err := s.issueSession(ctx, user)
	if err != nil {
		return nil, err
	}
	s.record(ctx, user.OrgID, user.ID, "user.login", "user", user.ID)
	return &AuthResult{User: user, Session: session}, nil
}
