package log

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
		" DEBUG ": slog.LevelDebug,
		"":        slog.LevelInfo,
		"verbose": slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestRedactMasksSensitiveKeys(t *testing.T) {
	for _, key := range []string{"password", "Authorization", "refresh_token", "X-Api_Key", "jwt", "client_secret"} {
		got := redact(nil, slog.String(key, "hunter2"))
		if got.Key != key || got.Value.String() != "[REDACTED]" {
			t.Errorf("%s was not redacted: %v", key, got)
		}
	}
	if got := redact(nil, slog.String("email", "grace@example.com")); got.Value.String() != "grace@example.com" {
		t.Errorf("non-sensitive value changed: %v", got)
	}
}
