package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aashishrajdev/halomail/services/shared/usage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPaymentLedgerReplayAndOrdering(t *testing.T) {
	address := os.Getenv("TEST_DATABASE_URL")
	if address == "" {
		t.Skip("set TEST_DATABASE_URL to test billing transactions in an isolated schema")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, address)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	schema := pgx.Identifier{fmt.Sprintf("billing_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
	config, err := pgxpool.ParseConfig(address)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	paths, err := filepath.Glob("../../../migrations/*.sql")
	if err != nil || len(paths) == 0 {
		t.Fatal("missing identity migrations")
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		up, _, _ := strings.Cut(string(content), "-- +goose Down")
		if _, err := pool.Exec(ctx, up); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
	}
	_, err = pool.Exec(ctx, `INSERT INTO orgs(id,name,slug) VALUES('org','Test','test');
 INSERT INTO users(id,org_id,email,handle,password_hash) VALUES('user','org','test@example.com','test','unused')`)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewUsers(pool)
	paid := time.Now().UTC().Truncate(time.Second)
	end := paid.AddDate(0, 1, 0)
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := repo.ApplyPayment(ctx, "org", "order_one", "pay_one", "pro", paid, end); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM billing_payments").Scan(&count); err != nil || count != 1 {
		t.Fatalf("duplicates: %d %v", count, err)
	}
	newer := paid.Add(time.Hour)
	if err := repo.ApplyPayment(ctx, "org", "order_two", "pay_two", "business", newer, newer.AddDate(0, 12, 0)); err != nil {
		t.Fatal(err)
	}
	if err := repo.ApplyPayment(ctx, "org", "order_one", "pay_one", "pro", paid, end); err != nil {
		t.Fatal(err)
	}
	if err := repo.ApplyPayment(ctx, "org", "order_old", "pay_old", "pro", paid.Add(-time.Hour), end); err != nil {
		t.Fatal(err)
	}
	var tier string
	var actualEnd time.Time
	if err := pool.QueryRow(ctx, "SELECT tier,billing_cycle_end FROM orgs WHERE id='org'").Scan(&tier, &actualEnd); err != nil {
		t.Fatal(err)
	}
	if tier != "business" || !actualEnd.Equal(newer.AddDate(0, 12, 0)) {
		t.Fatalf("old payment overwrote current subscription: %s %v", tier, actualEnd)
	}
	if err := repo.ApplyPayment(ctx, "org", "order_one", "pay_conflict", "business", paid, end); err == nil {
		t.Fatal("conflicting payment/order accepted")
	}
	stats, err := usage.Read(ctx, pool, usage.Policy{Forms: 25, Meetings: 10, Monthly: true}, "user", "forms")
	if err != nil || stats.Limit != 2500 {
		t.Fatalf("paid quota: %+v %v", stats, err)
	}
	if _, err := pool.Exec(ctx, "UPDATE orgs SET billing_cycle_end=now()-interval '1 second' WHERE id='org'"); err != nil {
		t.Fatal(err)
	}
	if got := usage.GetTier(ctx, pool, "user"); got != "free" {
		t.Fatalf("expired tier %s", got)
	}
	stats, err = usage.Read(ctx, pool, usage.Policy{Forms: 25, Monthly: true}, "user", "forms")
	if err != nil || stats.Limit != 25 {
		t.Fatalf("expired quota: %+v %v", stats, err)
	}
}
