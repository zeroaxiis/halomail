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

func TestSubscribes(t *testing.T) {
	hook := Webhook{Events: []string{EventBookingCreated, EventMessageReceived}}
	if !hook.Subscribes(EventBookingCreated) || !hook.Subscribes(EventMessageReceived) {
		t.Fatal("subscribed event not matched")
	}
	if hook.Subscribes(EventBookingCancelled) {
		t.Fatal("unsubscribed event matched")
	}
	if (Webhook{}).Subscribes(EventBookingCreated) {
		t.Fatal("webhook without events matched")
	}
}
