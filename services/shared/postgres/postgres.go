// Package postgres provides a pooled pgx connection.
package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// New opens and verifies a pgx pool against the given connection URL.
func New(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, fmt.Errorf("postgres: DATABASE_URL is empty")
	}

	// Neon's PgBouncer pooler endpoint (-pooler) operates in transaction pooling mode
	// which does not preserve session parameters like search_path and causes
	// intermittent "relation does not exist" errors. Persistent Go servers already
	// run native connection pools (pgxpool), so we automatically route Neon connections
	// to the direct unpooled endpoint.
	if strings.Contains(url, ".neon.tech") && strings.Contains(url, "-pooler.") {
		url = strings.Replace(url, "-pooler.", ".", 1)
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = time.Hour
	
	// CRITICAL FIX: Neon uses PgBouncer in transaction mode.
	// Prepared statements (which pgx uses by default) crash with "prepared statement name is already in use".
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = make(map[string]string)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = "public"

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
}
