package calendar

import (
	"encoding/base64"
	"github.com/aashishrajdev/halomail/services/scheduling/internal/domain"
	"golang.org/x/oauth2"
	"strings"
	"testing"
	"time"
)

func TestCredentialEncryption(test *testing.T) {
	client := &Google{key: base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))}
	token := &oauth2.Token{AccessToken: "private-access", RefreshToken: "private-refresh"}
	encrypted, err := client.seal("owner", token)
	if err != nil {
		test.Fatal(err)
	}
	decoded, err := client.open("owner", encrypted)
	if err != nil || decoded.RefreshToken != token.RefreshToken {
		test.Fatalf("round trip failed: %v", err)
	}
	if _, err = client.open("someone-else", encrypted); err == nil {
		test.Fatal("credential usable by another owner")
	}
	raw, _ := base64.StdEncoding.DecodeString(encrypted)
	raw[len(raw)-1] ^= 1
	if _, err = client.open("owner", base64.StdEncoding.EncodeToString(raw)); err == nil {
		test.Fatal("tampered credential accepted")
	}
}

func TestSameTimeUsesInstants(test *testing.T) {
	start := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	booking := domain.Booking{Start: start, End: start.Add(30 * time.Minute)}
	event := eventResponse{Start: eventTime{DateTime: "2026-10-01T15:30:00+05:30"}, End: eventTime{DateTime: "2026-10-01T16:00:00+05:30"}}
	if !sameTime(event, booking) {
		test.Fatal("retry would modify an unchanged event")
	}
	booking.Start = booking.Start.Add(time.Hour)
	if sameTime(event, booking) {
		test.Fatal("reschedule not detected")
	}
}
