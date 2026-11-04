package adminwebhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// Event is the envelope dispatched to the external admin dashboard.
type Event struct {
	Event     string `json:"event"`
	Timestamp string `json:"timestamp"`
	Data      any    `json:"data"`
}

// Dispatcher manages outbound HTTP webhooks to the external admin dashboard.
type Dispatcher struct {
	url    string
	secret string
	client *http.Client
}

// New creates a new Dispatcher. If url is empty, Dispatch will be a no-op.
func New(url, secret string) *Dispatcher {
	return NewWithClient(url, secret, &http.Client{Timeout: 10 * time.Second})
}

// NewWithClient creates a Dispatcher with a custom HTTP client (useful for tests or proxies).
func NewWithClient(url, secret string, client *http.Client) *Dispatcher {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &Dispatcher{
		url:    strings.TrimSpace(url),
		secret: strings.TrimSpace(secret),
		client: client,
	}
}

// Dispatch sends an event payload asynchronously in a background goroutine.
func (d *Dispatcher) Dispatch(event string, data any) {
	if d == nil || d.url == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		payload := Event{
			Event:     event,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Data:      data,
		}

		raw, err := json.Marshal(payload)
		if err != nil {
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.url, bytes.NewReader(raw))
		if err != nil {
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "HaloMail-Admin-Webhook/1.0")

		if d.secret != "" {
			mac := hmac.New(sha256.New, []byte(d.secret))
			mac.Write(raw)
			sig := hex.EncodeToString(mac.Sum(nil))
			req.Header.Set("X-HaloMail-Signature", sig)
		}

		resp, err := d.client.Do(req)
		if err != nil {
			slog.Default().Warn("admin webhook delivery failed", "url", d.url, "event", event, "err", err.Error())
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			slog.Default().Warn("admin webhook non-2xx response", "url", d.url, "event", event, "status", resp.StatusCode)
		} else {
			slog.Default().Debug("admin webhook delivered", "url", d.url, "event", event, "status", resp.StatusCode)
		}
	}()
}
