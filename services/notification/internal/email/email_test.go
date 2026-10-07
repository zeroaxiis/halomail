package email

import (
	"context"
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

// Addresses with line breaks are refused before the sender dials out, so the
// unroutable address below is never contacted.
func TestSMTPSendRejectsHeaderInjection(t *testing.T) {
	sender := NewSMTP("127.0.0.1", 1, "HaloMail <noreply@halomail.dev>")
	cases := map[string]Message{
		"recipient": {To: []string{"a@example.com\r\nBcc: evil@example.com"}},
		"reply-to":  {To: []string{"a@example.com"}, ReplyTo: "b@example.com\nBcc: evil@example.com"},
		"from":      {To: []string{"a@example.com"}, From: "c@example.com\rBcc: evil@example.com"},
	}
	for name, msg := range cases {
		_, provider, err := sender.Send(context.Background(), msg)
		if err == nil || err.Error() != "invalid mail address" {
			t.Errorf("%s: err = %v, want invalid mail address", name, err)
		}
		if provider != "smtp" {
			t.Errorf("%s: provider = %q", name, provider)
		}
	}
}
