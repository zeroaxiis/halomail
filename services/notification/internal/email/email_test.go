package email

import (
	"testing"
)

func TestAddressOnly(t *testing.T) {
	cases := map[string]string{
		"HaloMail <noreply@halomail.dev>": "noreply@halomail.dev",
		"noreply@halomail.dev":            "noreply@halomail.dev",
		"not an address":                  "not an address",
	}
	for in, want := range cases {
		if got := addressOnly(in); got != want {
			t.Errorf("addressOnly(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("  ", "", "a@example.com", "b@example.com"); got != "a@example.com" {
		t.Errorf("firstNonEmpty = %q", got)
	}
}
