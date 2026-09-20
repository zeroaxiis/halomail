package app

import (
	"context"
	"github.com/aashishrajdev/halomail/services/identity/internal/crypto"
	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"strings"
	"testing"
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
		service := New(Repos{APIKeys: repo, Audit: auditRepository{}}, Config{})
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
	service := New(Repos{}, Config{})
	if _, _, err := service.CreateAPIKey(context.Background(), "owner", "org", "admin", nil); err == nil {
		test.Fatal("unknown key feature accepted")
	}
}
