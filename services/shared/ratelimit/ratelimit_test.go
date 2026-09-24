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
