package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestConstructorsSetKindAndFormatMessage(t *testing.T) {
	cases := []struct {
		name string
		err  *Error
		kind Kind
	}{
		{"invalid", Invalid("bad %s", "input"), KindInvalid},
		{"not found", NotFound("bad %s", "input"), KindNotFound},
		{"conflict", Conflict("bad %s", "input"), KindConflict},
		{"unauthorized", Unauthorized("bad %s", "input"), KindUnauthorized},
		{"forbidden", Forbidden("bad %s", "input"), KindForbidden},
		{"rate limited", RateLimited("bad %s", "input"), KindRateLimited},
		{"internal", Internal("bad %s", "input"), KindInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Kind != tc.kind {
				t.Fatalf("kind = %d, want %d", tc.err.Kind, tc.kind)
			}
			if got := tc.err.Error(); got != "bad input" {
				t.Fatalf("message = %q, want %q", got, "bad input")
			}
		})
	}
}

func TestErrorIncludesWrappedCause(t *testing.T) {
	cause := errors.New("connection refused")
	err := Wrap(cause, KindInternal, "load user")
	if got := err.Error(); got != "load user: connection refused" {
		t.Fatalf("message = %q", got)
	}
	if !errors.Is(err, cause) {
		t.Fatal("wrapped cause is not reachable through errors.Is")
	}
	if Invalid("no cause").Unwrap() != nil {
		t.Fatal("constructor errors should not carry a cause")
	}
}

func TestKindOfWalksErrorChain(t *testing.T) {
	wrapped := fmt.Errorf("handler: %w", NotFound("form %s", "f1"))
	if got := KindOf(wrapped); got != KindNotFound {
		t.Fatalf("kind of wrapped error = %d, want %d", got, KindNotFound)
	}
	if got := KindOf(errors.New("plain")); got != KindUnknown {
		t.Fatalf("kind of plain error = %d, want %d", got, KindUnknown)
	}
	if got := KindOf(nil); got != KindUnknown {
		t.Fatalf("kind of nil = %d, want %d", got, KindUnknown)
	}
}
