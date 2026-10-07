package domain

import (
	"testing"
	"time"
)

func TestSessionLifetime(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	session := Session{ExpiresAt: now.Add(time.Hour)}

	if session.Expired(now) || !session.Active(now) {
		t.Fatal("fresh session should be active")
	}
	if session.Expired(session.ExpiresAt) {
		t.Fatal("session should still be valid at the exact expiry instant")
	}
	if later := now.Add(2 * time.Hour); !session.Expired(later) || session.Active(later) {
		t.Fatal("session past its expiry should be inactive")
	}

	session.Revoked = true
	if session.Active(now) {
		t.Fatal("revoked session should be inactive")
	}
}
