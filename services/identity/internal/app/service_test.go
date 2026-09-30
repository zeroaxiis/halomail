package app

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/aashishrajdev/halomail/services/identity/internal/crypto"
	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type keyRepository struct {
	APIKeyRepo
	saved *domain.APIKey
}

func (repo *keyRepository) Create(_ context.Context, key *domain.APIKey) error {
	repo.saved = key
	return nil
}

type auditRepository struct{ AuditRepo }

func (auditRepository) Insert(context.Context, *domain.AuditLog) error { return nil }

func TestFeatureKeysCannotRequestManagementScopes(test *testing.T) {
	for _, feature := range []string{"forms", "meetings"} {
		repo := &keyRepository{}
		service := New(Repos{APIKeys: repo, Audit: auditRepository{}}, Config{}, nil)
		key, secret, err := service.CreateAPIKey(context.Background(), "owner", "org", feature, []string{"admin", "users:write"})
		if err != nil {
			test.Fatal(err)
		}
		want := "forms:submit"
		if feature == "meetings" {
			want = "meetings:book"
		}
		if len(key.Scopes) != 1 || key.Scopes[0] != want {
			test.Fatalf("unsafe scopes: %v", key.Scopes)
		}
		if repo.saved.SecretHash != crypto.SHA256Hex(secret) || strings.Contains(repo.saved.SecretHash, secret) {
			test.Fatal("plaintext key persisted")
		}
		if !strings.HasPrefix(secret, "hl_"+feature+"_") {
			test.Fatal("wrong feature prefix")
		}
	}
	service := New(Repos{}, Config{}, nil)
	if _, _, err := service.CreateAPIKey(context.Background(), "owner", "org", "admin", nil); err == nil {
		test.Fatal("unknown key feature accepted")
	}
}

type mockAPIKeyRepo struct {
	APIKeyRepo
	getCalls int
	key      *domain.APIKey
}

func (m *mockAPIKeyRepo) GetBySecretHash(ctx context.Context, hash string) (*domain.APIKey, error) {
	m.getCalls++
	return m.key, nil
}

func (m *mockAPIKeyRepo) TouchLastUsed(ctx context.Context, id string, t time.Time) error {
	return nil
}

func TestVerifyAPIKeyRedisCaching(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	repo := &mockAPIKeyRepo{
		key: &domain.APIKey{
			ID:         "key_123",
			UserID:     "user_123",
			OrgID:      "org_123",
			Scopes:     []string{"admin"},
			SecretHash: crypto.SHA256Hex("test-secret"),
		},
	}

	service := New(Repos{APIKeys: repo}, Config{}, rdb)

	// First call should query the repo and cache the result
	userID, orgID, scopes, err := service.VerifyAPIKey(context.Background(), "test-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.getCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d", repo.getCalls)
	}
	if userID != "user_123" || orgID != "org_123" || len(scopes) != 1 || scopes[0] != "admin" {
		t.Fatalf("unexpected values returned")
	}

	// Second call should fetch from cache, repo getCalls should remain 1
	userID2, orgID2, scopes2, err := service.VerifyAPIKey(context.Background(), "test-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.getCalls != 1 {
		t.Fatalf("expected 1 repo call, got %d. Cache missed!", repo.getCalls)
	}

	if userID != userID2 || orgID != orgID2 || len(scopes) != len(scopes2) {
		t.Fatal("cache result mismatch")
	}
}
