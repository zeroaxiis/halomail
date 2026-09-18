package errs

import (
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
