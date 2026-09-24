package ratelimit

import (
	"context"
	"testing"
)

func TestMemoryAllowsBurstThenDenies(t *testing.T) {
	limiter := NewMemory(Config{RPS: 0.001, Burst: 3})
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if ok, err := limiter.Allow(ctx, "203.0.113.7"); err != nil || !ok {
			t.Fatalf("request %d inside burst denied (err=%v)", i+1, err)
		}
	}
	if ok, _ := limiter.Allow(ctx, "203.0.113.7"); ok {
		t.Fatal("request beyond burst allowed")
	}
}

func TestMemoryTracksKeysIndependently(t *testing.T) {
	limiter := NewMemory(Config{RPS: 0.001, Burst: 1})
	ctx := context.Background()
	if ok, _ := limiter.Allow(ctx, "a"); !ok {
		t.Fatal("first request for a denied")
	}
	if ok, _ := limiter.Allow(ctx, "a"); ok {
		t.Fatal("second request for a allowed")
	}
	if ok, _ := limiter.Allow(ctx, "b"); !ok {
		t.Fatal("b was throttled by a's usage")
	}
}

func TestNewFallsBackToMemoryWithoutRedis(t *testing.T) {
	if _, ok := New(nil, Config{RPS: 1, Burst: 1}).(*memoryLimiter); !ok {
		t.Fatal("nil redis client should select the in-memory limiter")
	}
}
