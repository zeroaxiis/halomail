package config

import (
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
