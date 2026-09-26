package config

import (
	"strings"
	"testing"
)

func TestHelpers(t *testing.T) {
	if got := (HTTP{Host: "127.0.0.1", Port: 9090}).Addr(); got != "127.0.0.1:9090" {
		t.Fatalf("Addr = %q", got)
	}
	if !(App{Env: "production"}).IsProd() || (App{Env: "development"}).IsProd() {
		t.Fatal("IsProd should be true only for production")
	}
	if (Email{}).UseResend() || !(Email{ResendAPIKey: "re_123"}).UseResend() {
		t.Fatal("UseResend should follow the presence of an API key")
	}
}

func TestLoadRejectsInvalidSettings(t *testing.T) {
	t.Run("short jwt secret", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "too-short")
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Fatalf("err = %v, want JWT_SECRET length error", err)
		}
	})
	t.Run("negative free limit", func(t *testing.T) {
		t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
		t.Setenv("FREE_FORM_LIMIT", "-1")
		if _, err := Load(); err == nil {
			t.Fatal("negative free limit accepted")
		}
	})
}
