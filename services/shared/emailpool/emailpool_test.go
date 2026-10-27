package emailpool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestPool_RoundRobin(t *testing.T) {
	var callCountAccount1 int32
	var callCountAccount2 int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "Bearer key_1" {
			atomic.AddInt32(&callCountAccount1, 1)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "res_1"})
		} else if auth == "Bearer key_2" {
			atomic.AddInt32(&callCountAccount2, 1)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "res_2"})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer server.Close()

	accounts := []*Account{
		{APIKey: "key_1", From: "Sender 1 <s1@example.com>"},
		{APIKey: "key_2", From: "Sender 2 <s2@example.com>"},
	}

	p := New(accounts, WithBaseURL(server.URL), WithHTTPClient(server.Client()))

	for i := 0; i < 4; i++ {
		_, _, err := p.Send(context.Background(), Message{
			To:      []string{"user@example.com"},
			Subject: "Test",
			Text:    "Hello",
		})
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}

	if atomic.LoadInt32(&callCountAccount1) != 2 || atomic.LoadInt32(&callCountAccount2) != 2 {
		t.Fatalf("expected 2 calls per account, got a1=%d, a2=%d", callCountAccount1, callCountAccount2)
	}
}

func TestPool_FailoverOnDailyQuota(t *testing.T) {
	var callCountAccount1 int32
	var callCountAccount2 int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "Bearer key_1" {
			atomic.AddInt32(&callCountAccount1, 1)
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"statusCode": 429,
				"message":    "You have reached your daily email sending limit.",
				"name":       "daily_quota_exceeded",
			})
		} else if auth == "Bearer key_2" {
			atomic.AddInt32(&callCountAccount2, 1)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"id": "res_success_2"})
		}
	}))
	defer server.Close()

	acc1 := &Account{APIKey: "key_1", From: "Sender 1 <s1@example.com>"}
	acc2 := &Account{APIKey: "key_2", From: "Sender 2 <s2@example.com>"}

	p := New([]*Account{acc1, acc2}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))

	// First call hits acc1 -> gets 429 daily quota -> fails over to acc2
	id, _, err := p.Send(context.Background(), Message{
		To:      []string{"user@example.com"},
		Subject: "Test Quota",
	})
	if err != nil {
		t.Fatalf("expected failover to succeed, got error: %v", err)
	}
	if id != "res_success_2" {
		t.Fatalf("expected res_success_2, got %s", id)
	}

	// Verify acc1 is now marked inactive (exhausted)
	if acc1.IsActive() {
		t.Fatal("expected acc1 to be marked inactive after daily_quota_exceeded")
	}

	// Subsequent call should skip acc1 directly and go to acc2 without failing
	id2, _, err2 := p.Send(context.Background(), Message{
		To:      []string{"user@example.com"},
		Subject: "Test Quota 2",
	})
	if err2 != nil {
		t.Fatalf("expected subsequent call to succeed, got %v", err2)
	}
	if id2 != "res_success_2" {
		t.Fatalf("expected res_success_2 on 2nd call, got %s", id2)
	}
}

func TestPool_NewFromEnv(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "key_test_1")
	t.Setenv("EMAIL_FROM", "HaloMail <no-reply@email.zeroaxiis.tech>")
	t.Setenv("NTS_RESEND_API_KEY", "key_test_2")
	t.Setenv("NTS_EMAIL", "HaloMail <no-reply@nts.zeroaxiis.tech>")
	t.Setenv("NOTIFY_RESEND_API_KEY", "key_test_3")
	t.Setenv("NOTIFY_EMAIL", "HaloMail <no-reply@notify.zeroaxiis.tech>")

	pool := NewFromEnv()
	if pool.Len() != 3 {
		t.Fatalf("expected 3 accounts discovered, got %d", pool.Len())
	}
	if pool.ActiveCount() != 3 {
		t.Fatalf("expected 3 active accounts, got %d", pool.ActiveCount())
	}
}
