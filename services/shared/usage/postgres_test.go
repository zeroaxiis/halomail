package usage_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aashishrajdev/halomail/services/shared/accesskey"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/aashishrajdev/halomail/services/shared/usage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPostgresQuotaAndKeyLifecycle(test *testing.T) {
	address := os.Getenv("TEST_DATABASE_URL")
	if address == "" {
		test.Skip("set TEST_DATABASE_URL to run isolated PostgreSQL migration/concurrency tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, address)
	if err != nil {
		test.Fatal(err)
	}
	defer admin.Close()
	schema := pgx.Identifier{fmt.Sprintf("halomail_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		test.Fatal(err)
	}
	defer func() {
		if _, cleanupErr := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); cleanupErr != nil {
			test.Error(cleanupErr)
		}
	}()
	settings, err := pgxpool.ParseConfig(address)
	if err != nil {
		test.Fatal(err)
	}
	settings.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, settings)
	if err != nil {
		test.Fatal(err)
	}
	defer pool.Close()
	for _, service := range []string{"identity", "scheduling", "contact", "notification"} {
		paths, err := filepath.Glob(filepath.Join("..", "..", service, "migrations", "*.sql"))
		if err != nil || len(paths) == 0 {
			test.Fatalf("missing migrations for %s", service)
		}
		for _, path := range paths {
			content, err := os.ReadFile(path)
			if err != nil {
				test.Fatal(err)
			}
			up, _, _ := strings.Cut(string(content), "-- +goose Down")
			if _, err = pool.Exec(ctx, up); err != nil {
				test.Fatalf("%s: %v", path, err)
			}
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO orgs(id,name,slug) VALUES('org','Test','test'); INSERT INTO users(id,org_id,email,handle,password_hash) VALUES('owner','org','owner@example.com','owner','unused')`); err != nil {
		test.Fatal(err)
	}
	secret := strings.Repeat("a", 48)
	hash := sha256.Sum256([]byte(secret))
	insert := `INSERT INTO api_keys(id,org_id,user_id,name,prefix,last_four,secret_hash,scopes) VALUES($1,'org','owner','forms','hl_forms','aaaa',$2,ARRAY['forms:submit'])`
	if _, err = pool.Exec(ctx, insert, "first", hex.EncodeToString(hash[:])); err != nil {
		test.Fatal(err)
	}
	if _, err = pool.Exec(ctx, insert, "second", "different-hash"); err == nil {
		test.Fatal("allowed a second active forms key")
	}
	if _, err = accesskey.Verify(ctx, pool, secret, "forms"); err != nil {
		test.Fatal(err)
	}
	if _, err = accesskey.Verify(ctx, pool, secret, "meetings"); err == nil {
		test.Fatal("forms key allowed meeting bookings")
	}
	var successes atomic.Int32
	var pending sync.WaitGroup
	for attempt := 0; attempt < 40; attempt++ {
		pending.Add(1)
		go func() {
			defer pending.Done()
			transaction, err := pool.Begin(ctx)
			if err != nil {
				test.Error(err)
				return
			}
			defer transaction.Rollback(ctx)
			err = usage.Consume(ctx, transaction, usage.Policy{Forms: 25, Monthly: true}, "owner", "forms")
			if err != nil {
				if errs.KindOf(err) != errs.KindRateLimited {
					test.Error(err)
				}
				return
			}
			if err = transaction.Commit(ctx); err != nil {
				test.Error(err)
				return
			}
			successes.Add(1)
		}()
	}
	pending.Wait()
	if successes.Load() != 25 {
		test.Fatalf("quota allowed %d concurrent submissions, want 25", successes.Load())
	}
	if _, err = pool.Exec(ctx, `UPDATE api_keys SET revoked=true WHERE id='first'`); err != nil {
		test.Fatal(err)
	}
	if _, err = accesskey.Verify(ctx, pool, secret, "forms"); err == nil {
		test.Fatal("deleted key accepted")
	}
	if _, err = pool.Exec(ctx, insert, "second", "different-hash"); err != nil {
		test.Fatal("could not replace revoked key:", err)
	}
	stats, err := usage.Read(ctx, pool, usage.Policy{Forms: 25, Monthly: true}, "owner", "forms")
	if err != nil || stats.Used != 25 || stats.Remaining != 0 {
		test.Fatalf("key rotation reset quota: %+v %v", stats, err)
	}
}
