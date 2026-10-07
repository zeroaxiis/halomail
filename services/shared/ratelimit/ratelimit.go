// Package ratelimit provides a small rate limiter with two interchangeable
// backends: Redis (shared across instances) and in-memory (per instance).
//
// This is what lets HaloMail run without Redis on free tiers: when no Redis
// client is supplied, New falls back to the in-memory limiter.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// Limiter decides whether an action identified by key may proceed now.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

// Config is the allowed steady-state rate and burst.
type Config struct {
	RPS   float64 // requests per second
	Burst int     // bucket size / window allowance
}

// New returns a Redis-backed limiter when client is non-nil, otherwise an
// in-memory one. Callers don't care which they get.
func New(client *redis.Client, cfg Config) Limiter {
	if client != nil {
		return NewRedis(client, cfg)
	}
	return NewMemory(cfg)
}

// ---------------------------------------------------------------------------
// In-memory token bucket (per key), suitable for single-instance / free tier.
// ---------------------------------------------------------------------------

type memoryLimiter struct {
	cfg     Config
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	lim  *rate.Limiter
	seen time.Time
}

// NewMemory returns a per-process token-bucket limiter. It starts a background
// goroutine that evicts keys idle for ten minutes.
func NewMemory(cfg Config) *memoryLimiter {
	m := &memoryLimiter{cfg: cfg, buckets: make(map[string]*bucket)}
	go m.gc()
	return m
}

func (m *memoryLimiter) Allow(_ context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, ok := m.buckets[key]
	if !ok {
		b = &bucket{lim: rate.NewLimiter(rate.Limit(m.cfg.RPS), m.cfg.Burst)}
		m.buckets[key] = b
	}
	b.seen = time.Now()
	return b.lim.Allow(), nil
}

// gc evicts idle buckets so the map doesn't grow unbounded.
func (m *memoryLimiter) gc() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		m.mu.Lock()
		for k, b := range m.buckets {
			if b.seen.Before(cutoff) {
				delete(m.buckets, k)
			}
		}
		m.mu.Unlock()
	}
}

// ---------------------------------------------------------------------------
// Redis fixed-window counter, shared across instances.
// ---------------------------------------------------------------------------

type redisLimiter struct {
	client    *redis.Client
	limit     int64
	windowSec int64
}

// NewRedis returns a fixed one-second-window limiter shared across instances.
// Burst is the allowance per window, falling back to RPS when unset.
func NewRedis(client *redis.Client, cfg Config) *redisLimiter {
	limit := int64(cfg.Burst)
	if limit <= 0 {
		limit = int64(cfg.RPS)
	}
	return &redisLimiter{client: client, limit: limit, windowSec: 1}
}

func (r *redisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	window := time.Now().Unix() / r.windowSec
	rk := fmt.Sprintf("rl:%s:%d", key, window)

	n, err := r.client.Incr(ctx, rk).Result()
	if err != nil {
		// Fail open: a cache outage must not lock everyone out.
		return true, err
	}
	if n == 1 {
		_ = r.client.Expire(ctx, rk, time.Duration(r.windowSec)*time.Second).Err()
	}
	return n <= r.limit, nil
}
