package domain

import (
	"testing"
)

func TestSecretLastFour(t *testing.T) {
	cases := map[string]string{
		"whsec_abcd1234": "1234",
		"abcd":           "abcd",
		"abc":            "abc",
		"":               "",
	}
	for secret, want := range cases {
		if got := (Webhook{Secret: secret}).SecretLastFour(); got != want {
			t.Errorf("SecretLastFour(%q) = %q, want %q", secret, got, want)
		}
	}
}
