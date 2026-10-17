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

	"github.com/aashishrajdev/halomail/services/identity/internal/domain"
	"github.com/aashishrajdev/halomail/services/shared/usage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func refundTestDB(t *testing.T) (*pgxpool.Pool, context.Context) {
	t.Helper()
	address := os.Getenv("TEST_DATABASE_URL")
	if address == "" {
		t.Skip("set TEST_DATABASE_URL to test refunds in an isolated schema")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(admin.Close)
	schema := pgx.Identifier{fmt.Sprintf("refund_test_%d", time.Now().UnixNano())}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	config, err := pgxpool.ParseConfig(address)
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	paths, err := filepath.Glob("../../../migrations/*.sql")
	if err != nil || len(paths) == 0 {
		t.Fatal("missing migrations")
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
	if _, err := pool.Exec(ctx, `INSERT INTO orgs(id,name,slug) VALUES('org','Test','test');
 INSERT INTO users(id,org_id,email,handle,password_hash) VALUES('user','org','refund@example.com','test','unused')`); err != nil {
		t.Fatal(err)
	}
	return pool, ctx
}

func assertPlan(t *testing.T, ctx context.Context, pool *pgxpool.Pool, want string) {
	t.Helper()
	var tier string
	if err := pool.QueryRow(ctx, "SELECT tier FROM orgs WHERE id='org'").Scan(&tier); err != nil {
		t.Fatal(err)
	}
	if tier != want {
		t.Fatalf("plan = %s, want %s", tier, want)
	}
	if tier := usage.GetTier(ctx, pool, "user"); tier != want {
		t.Fatalf("effective plan = %s, want %s", tier, want)
	}
}

func TestFullRefundDowngradesWithoutLosingUsageOrRevivingOldPayments(t *testing.T) {
	pool, ctx := refundTestDB(t)
	repo := NewUsers(pool)
	paid := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	if err := repo.ApplyPayment(ctx, "org", "order_one", "pay_one", "pro", paid, paid.AddDate(0, 1, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO feature_usage(owner_id,feature,period,used) VALUES('user','forms',date_trunc('month',now()),7)`); err != nil {
		t.Fatal(err)
	}
	first := domain.BillingRefund{ID: "rfnd_one", Amount: 40, CreatedAt: paid.Add(time.Minute)}
	second := domain.BillingRefund{ID: "rfnd_two", Amount: 60, CreatedAt: paid.Add(2 * time.Minute)}
	result, err := repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{first})
	if err != nil || result.FullyRefunded || result.Downgraded {
		t.Fatalf("partial refund changed plan: %+v %v", result, err)
	}
	assertPlan(t, ctx, pool, "pro")
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{first}); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	var count, total int64
	if err := pool.QueryRow(ctx, "SELECT count(*),sum(amount) FROM billing_refunds").Scan(&count, &total); err != nil || count != 1 || total != 40 {
		t.Fatalf("duplicate refund counted: %d %d %v", count, total, err)
	}
	result, err = repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{first, second})
	if err != nil || !result.FullyRefunded || !result.Downgraded {
		t.Fatalf("full refund not applied: %+v %v", result, err)
	}
	assertPlan(t, ctx, pool, "free")
	var cleared bool
	var watermark time.Time
	if err := pool.QueryRow(ctx, "SELECT rzp_subscription_id IS NULL AND billing_cycle_end IS NULL,billing_paid_at FROM orgs WHERE id='org'").Scan(&cleared, &watermark); err != nil || !cleared || !watermark.Equal(paid) {
		t.Fatalf("billing state not cleared correctly: %v %v %v", cleared, watermark, err)
	}
	stats, err := usage.Read(ctx, pool, usage.Policy{Forms: 25, Monthly: true}, "user", "forms")
	if err != nil || stats.Used != 7 || stats.Limit != 25 || stats.Remaining != 18 {
		t.Fatalf("wrong free allowance: %+v %v", stats, err)
	}
	for _, p := range []struct {
		order, id string
		at        time.Time
	}{{"order_one", "pay_one", paid}, {"order_old", "pay_old", paid.Add(-time.Hour)}} {
		if err := repo.ApplyPayment(ctx, "org", p.order, p.id, "pro", p.at, p.at.AddDate(0, 1, 0)); err != nil {
			t.Fatal(err)
		}
	}
	assertPlan(t, ctx, pool, "free")
	result, err = repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{first})
	if err != nil || !result.FullyRefunded || result.Downgraded {
		t.Fatalf("stale snapshot regressed refund state: %+v %v", result, err)
	}
	newer := paid.Add(30 * time.Minute)
	if err := repo.ApplyPayment(ctx, "org", "order_new", "pay_new", "business", newer, newer.AddDate(0, 1, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{first, second}); err != nil {
		t.Fatal(err)
	}
	assertPlan(t, ctx, pool, "business")
	for _, args := range []struct {
		payment, order, currency string
		amount                   int64
		refunds                  []domain.BillingRefund
	}{
		{"pay_one", "order_one", "USD", 100, []domain.BillingRefund{first}},
		{"pay_one", "order_one", "INR", 101, []domain.BillingRefund{first}},
		{"pay_new", "order_new", "INR", 100, []domain.BillingRefund{first}},
		{"pay_other", "order_one", "INR", 100, []domain.BillingRefund{first}},
		{"pay_one", "order_one", "INR", 100, []domain.BillingRefund{{ID: "rfnd_over", Amount: 1, CreatedAt: paid}}},
	} {
		if _, err := repo.ApplyRefunds(ctx, "org", args.order, args.payment, args.currency, args.amount, args.refunds); err == nil {
			t.Fatalf("conflicting refund accepted: %+v", args)
		}
	}
	assertPlan(t, ctx, pool, "business")
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM billing_refunds").Scan(&count); err != nil || count != 2 {
		t.Fatalf("failed refund was not rolled back: %d %v", count, err)
	}
}

func TestRefundBeforeCaptureAndConcurrentCallbacks(t *testing.T) {
	for _, partial := range []bool{false, true} {
		t.Run(fmt.Sprintf("partial_%t", partial), func(t *testing.T) {
			pool, ctx := refundTestDB(t)
			repo := NewUsers(pool)
			paid := time.Now().UTC().Truncate(time.Second)
			amount := int64(100)
			if partial {
				amount = 40
			}
			refund := domain.BillingRefund{ID: "rfnd_before", Amount: amount, CreatedAt: paid}
			if _, err := repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{refund}); err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			for i := 0; i < 12; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					var err error
					if i%2 == 0 {
						err = repo.ApplyPayment(ctx, "org", "order_one", "pay_one", "pro", paid, paid.AddDate(0, 1, 0))
					} else {
						_, err = repo.ApplyRefunds(ctx, "org", "order_one", "pay_one", "INR", 100, []domain.BillingRefund{refund})
					}
					if err != nil {
						t.Error(err)
					}
				}(i)
			}
			wg.Wait()
			want := "free"
			if partial {
				want = "pro"
			}
			assertPlan(t, ctx, pool, want)
		})
	}
}
