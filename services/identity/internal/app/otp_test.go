package app

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/crypto"
	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type otpUsers struct {
	UserRepo
	user *domain.User
}

func (r *otpUsers) GetUserByEmail(_ context.Context, email string) (*domain.User, error) {
	if email != r.user.Email {
		return nil, errs.NotFound("user")
	}
	return r.user, nil
}
func (r *otpUsers) GetUserByID(_ context.Context, id string) (*domain.User, error) {
	if id != r.user.ID {
		return nil, errs.NotFound("user")
	}
	return r.user, nil
}
func (r *otpUsers) CreateOrgAndUser(_ context.Context, _ *domain.Org, u *domain.User) error {
	r.user = u
	return nil
}

type otpSessions struct {
	SessionRepo
	created atomic.Int32
}

func (r *otpSessions) Create(context.Context, *domain.Session) error {
	r.created.Add(1)
	return nil
}

func otpFixture(t *testing.T) (*Service, *miniredis.Miniredis, *otpSessions) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr(), MaxRetries: -1, DialTimeout: 100 * time.Millisecond})
	t.Cleanup(func() { _ = client.Close() })
	hash, err := crypto.HashPassword("test-password")
	if err != nil {
		t.Fatal(err)
	}
	sessions := &otpSessions{}
	s := New(Repos{
		Users:    &otpUsers{user: &domain.User{ID: "user", OrgID: "org", Email: "user@example.com", PasswordHash: hash}},
		Sessions: sessions, Audit: auditRepository{},
	}, Config{JWTSecret: strings.Repeat("s", 32)}, client)
	return s, mr, sessions
}

func requestOTP(t *testing.T, s *Service) *OTPDelivery {
	t.Helper()
	d, err := s.RequestLoginOTP(context.Background(), " User@Example.com ", "test-password")
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestOTPExpiresAndStoresOnlyHash(t *testing.T) {
	s, mr, sessions := otpFixture(t)
	d := requestOTP(t, s)
	key := "auth:otp:challenge:" + d.ChallengeID
	if mr.TTL(key) != 10*time.Minute || d.ExpiresIn != 600 {
		t.Fatal("wrong expiry")
	}
	if mr.HGet(key, "hash") != s.otpHash(d.ChallengeID, d.Code) {
		t.Fatal("missing OTP hash")
	}
	if sessions.created.Load() != 0 {
		t.Fatal("session created before verification")
	}
	mr.FastForward(10 * time.Minute)
	if mr.Exists(key) {
		t.Fatal("Redis did not expire OTP")
	}
	if _, err := s.VerifyLoginOTP(context.Background(), d.ChallengeID, d.Code); errs.KindOf(err) != errs.KindUnauthorized {
		t.Fatalf("expired OTP accepted: %v", err)
	}
	if sessions.created.Load() != 0 {
		t.Fatal("expired OTP created a session")
	}
}

func TestOTPConsumedExactlyOnceConcurrently(t *testing.T) {
	s, mr, sessions := otpFixture(t)
	d := requestOTP(t, s)
	var successes atomic.Int32
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := s.VerifyLoginOTP(context.Background(), d.ChallengeID, d.Code)
			if err == nil {
				successes.Add(1)
				if res.Session.AccessToken == "" || res.Session.RefreshToken == "" {
					t.Error("missing tokens")
				}
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 || sessions.created.Load() != 1 {
		t.Fatal("OTP reused concurrently")
	}
	if mr.Exists("auth:otp:challenge:" + d.ChallengeID) {
		t.Fatal("used OTP remains in Redis")
	}
}

func TestOTPWrongAttemptsDoNotExtendTTLAndLockOut(t *testing.T) {
	s, mr, sessions := otpFixture(t)
	d := requestOTP(t, s)
	wrong := "000000"
	if d.Code == wrong {
		wrong = "111111"
	}
	mr.FastForward(time.Minute)
	for range 4 {
		if _, err := s.VerifyLoginOTP(context.Background(), d.ChallengeID, wrong); errs.KindOf(err) != errs.KindUnauthorized {
			t.Fatal("wrong OTP accepted")
		}
	}
	if mr.TTL("auth:otp:challenge:"+d.ChallengeID) != 9*time.Minute {
		t.Fatal("attempts extended expiry")
	}
	_, _ = s.VerifyLoginOTP(context.Background(), d.ChallengeID, wrong)
	if _, err := s.VerifyLoginOTP(context.Background(), d.ChallengeID, d.Code); err == nil {
		t.Fatal("locked OTP accepted")
	}
	if sessions.created.Load() != 0 {
		t.Fatal("wrong OTP created session")
	}
}

func TestOTPResendCooldownReplacementAndCancellation(t *testing.T) {
	s, mr, _ := otpFixture(t)
	first := requestOTP(t, s)
	if _, err := s.RequestLoginOTP(context.Background(), "user@example.com", "test-password"); errs.KindOf(err) != errs.KindRateLimited {
		t.Fatal("missing resend cooldown")
	}
	mr.FastForward(time.Minute)
	second := requestOTP(t, s)
	if _, err := s.VerifyLoginOTP(context.Background(), first.ChallengeID, first.Code); err == nil {
		t.Fatal("old code survived resend")
	}
	if err := s.CancelLoginOTP(context.Background(), second.ChallengeID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.VerifyLoginOTP(context.Background(), second.ChallengeID, second.Code); err == nil {
		t.Fatal("cancelled code accepted")
	}
}

func TestOTPCredentialLimitAndNoBypass(t *testing.T) {
	s, _, sessions := otpFixture(t)
	ctx := context.Background()
	for range 10 {
		if _, err := s.RequestLoginOTP(ctx, "user@example.com", "wrong"); errs.KindOf(err) != errs.KindUnauthorized {
			t.Fatalf("bad password: %v", err)
		}
	}
	if _, err := s.RequestLoginOTP(ctx, "user@example.com", "test-password"); errs.KindOf(err) != errs.KindRateLimited {
		t.Fatal("missing request limit")
	}
	if _, err := s.Login(ctx, "user@example.com", "test-password"); err == nil {
		t.Fatal("legacy login bypass")
	}
	res, err := s.Register(ctx, "new@example.com", "test-password", "New user")
	if err != nil {
		t.Fatal(err)
	}
	if res.Session.AccessToken != "" || res.Session.RefreshToken != "" || sessions.created.Load() != 0 {
		t.Fatal("registration bypass")
	}
}

func TestOTPUnavailableRedisFailsClosed(t *testing.T) {
	s, mr, sessions := otpFixture(t)
	d := requestOTP(t, s)
	mr.Close()
	if _, err := s.VerifyLoginOTP(context.Background(), d.ChallengeID, d.Code); err == nil {
		t.Fatal("verified with unavailable Redis")
	}
	s.redis = nil
	if _, err := s.RequestLoginOTP(context.Background(), "user@example.com", "test-password"); err == nil {
		t.Fatal("requested with no Redis")
	}
	if _, err := s.VerifyLoginOTP(context.Background(), d.ChallengeID, d.Code); err == nil {
		t.Fatal("verified with no Redis")
	}
	if sessions.created.Load() != 0 {
		t.Fatal("session issued with unavailable Redis")
	}
}
