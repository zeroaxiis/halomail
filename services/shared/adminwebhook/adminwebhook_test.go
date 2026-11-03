package adminwebhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

type mockTransport struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestDispatcherSendsSignedPayload(t *testing.T) {
	received := make(chan []byte, 1)
	receivedSig := make(chan string, 1)

	client := &http.Client{
		Transport: &mockTransport{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				receivedSig <- req.Header.Get("X-HaloMail-Signature")
				received <- body
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"received":true}`))),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	secret := "test-secret-key-12345"
	targetURL := "https://admin.example.com/v1/webhooks/halomail"
	dispatcher := NewWithClient(targetURL, secret, client)

	testData := map[string]string{"userId": "usr_test123", "email": "test@example.com"}
	dispatcher.Dispatch("user.registered", testData)

	select {
	case body := <-received:
		sig := <-receivedSig

		var evt Event
		if err := json.Unmarshal(body, &evt); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if evt.Event != "user.registered" {
			t.Errorf("expected event user.registered, got %s", evt.Event)
		}

		// Verify signature
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSig := hex.EncodeToString(mac.Sum(nil))
		if sig != expectedSig {
			t.Errorf("expected signature %s, got %s", expectedSig, sig)
		}

	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for webhook delivery")
	}
}

func TestNilOrEmptyURLDoesNotPanic(t *testing.T) {
	var nilDispatcher *Dispatcher
	nilDispatcher.Dispatch("test", nil)

	emptyDispatcher := New("", "")
	emptyDispatcher.Dispatch("test", nil)
}
