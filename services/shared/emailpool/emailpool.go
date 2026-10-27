// Package emailpool provides a multi-account, round-robin, and failover
// pool for transactional email delivery via Resend.
//
// Features:
// - Round-robin dispatch across multiple Resend accounts (e.g. email, nts, notify)
// - Triples burst throughput (bypassing Resend's 2 req/sec limit)
// - Automatically detects HTTP 429 daily_quota_exceeded and marks the exhausted account
//   inactive until 00:01 UTC midnight, immediately retrying with the next account in the pool.
// - Automatically detects HTTP 429 rate_limit_exceeded and temporarily backs off the account.
// - Zero hardcoded secrets: reads strictly from environment variables or explicit config.
package emailpool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Message is a transactional email request.
type Message struct {
	ID      string
	To      []string
	From    string
	ReplyTo string
	Subject string
	HTML    string
	Text    string
}

// Account represents a single Resend account credentials and sending identity.
type Account struct {
	APIKey string
	From   string
	Domain string

	mu               sync.RWMutex
	exhaustedUntil   time.Time // Daily quota 429
	rateLimitedUntil time.Time // 2 req/s burst 429
}

func (a *Account) IsActive() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	now := time.Now()
	if !a.exhaustedUntil.IsZero() && now.Before(a.exhaustedUntil) {
		return false
	}
	if !a.rateLimitedUntil.IsZero() && now.Before(a.rateLimitedUntil) {
		return false
	}
	return true
}

func (a *Account) MarkDailyExhausted() {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now().UTC()
	// Next UTC midnight + 1 minute grace
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 1, 0, 0, time.UTC)
	a.exhaustedUntil = nextMidnight
}

func (a *Account) MarkRateLimited(duration time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.rateLimitedUntil = time.Now().Add(duration)
}

// Pool coordinates dispatching across multiple Resend accounts.
type Pool struct {
	accounts []*Account
	counter  uint64
	client   *http.Client
	baseURL  string
}

// Option configures the pool.
type Option func(*Pool)

// WithHTTPClient sets a custom HTTP client (useful for testing).
func WithHTTPClient(c *http.Client) Option {
	return func(p *Pool) {
		if c != nil {
			p.client = c
		}
	}
}

// WithBaseURL sets a custom API endpoint (useful for testing with mock servers).
func WithBaseURL(url string) Option {
	return func(p *Pool) {
		if url != "" {
			p.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// Clean helper
func cleanVal(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"'`)
	return strings.TrimSpace(s)
}

// New creates a new pool with the given accounts.
func New(accounts []*Account, opts ...Option) *Pool {
	p := &Pool{
		accounts: accounts,
		client:   &http.Client{Timeout: 15 * time.Second},
		baseURL:  "https://api.resend.com",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// NewFromEnv constructs a pool by scanning standard environment variables.
// It checks:
// 1. RESEND_POOL (comma-separated key:from pairs, e.g. "re_1:HaloMail <no-reply@email.zeroaxiis.tech>,re_2:HaloMail <no-reply@nts.zeroaxiis.tech>")
// 2. RESEND_API_KEY / EMAIL_FROM
// 3. NTS_RESEND_API_KEY / NTS_EMAIL
// 4. NOTIFY_RESEND_API_KEY / NOTIFY_EMAIL
// 5. RESEND_KEY_1..5 / RESEND_FROM_1..5
func NewFromEnv(opts ...Option) *Pool {
	seenKeys := make(map[string]bool)
	var accounts []*Account

	addAccount := func(apiKey, from string) {
		apiKey = cleanVal(apiKey)
		from = cleanVal(from)
		if apiKey == "" || seenKeys[apiKey] {
			return
		}
		if from == "" || strings.Contains(from, "halomail.dev") {
			from = "HaloMail <no-reply@email.zeroaxiis.tech>"
		}
		seenKeys[apiKey] = true
		accounts = append(accounts, &Account{
			APIKey: apiKey,
			From:   from,
		})
	}

	// 1. Explicit pool string: key1:from1,key2:from2
	if poolStr := cleanVal(os.Getenv("RESEND_POOL")); poolStr != "" {
		for _, entry := range strings.Split(poolStr, ",") {
			parts := strings.SplitN(strings.TrimSpace(entry), ":", 2)
			if len(parts) == 2 {
				addAccount(parts[0], parts[1])
			} else if len(parts) == 1 {
				addAccount(parts[0], "")
			}
		}
	}

	// 2. Primary: RESEND_API_KEY + EMAIL_FROM
	addAccount(os.Getenv("RESEND_API_KEY"), os.Getenv("EMAIL_FROM"))

	// 3. Secondary: NTS_RESEND_API_KEY + NTS_EMAIL
	addAccount(os.Getenv("NTS_RESEND_API_KEY"), os.Getenv("NTS_EMAIL"))
	if _, ok := seenKeys[cleanVal(os.Getenv("NTS_RESEND_API_KEY"))]; !ok && cleanVal(os.Getenv("NTS_RESEND_API_KEY")) != "" {
		addAccount(os.Getenv("NTS_RESEND_API_KEY"), "HaloMail <no-reply@nts.zeroaxiis.tech>")
	}

	// 4. Tertiary: NOTIFY_RESEND_API_KEY + NOTIFY_EMAIL
	addAccount(os.Getenv("NOTIFY_RESEND_API_KEY"), os.Getenv("NOTIFY_EMAIL"))
	if _, ok := seenKeys[cleanVal(os.Getenv("NOTIFY_RESEND_API_KEY"))]; !ok && cleanVal(os.Getenv("NOTIFY_RESEND_API_KEY")) != "" {
		addAccount(os.Getenv("NOTIFY_RESEND_API_KEY"), "HaloMail <no-reply@notify.zeroaxiis.tech>")
	}

	// 5. Dynamic indexed keys: RESEND_KEY_1..5
	for i := 1; i <= 5; i++ {
		k := cleanVal(os.Getenv(fmt.Sprintf("RESEND_KEY_%d", i)))
		f := cleanVal(os.Getenv(fmt.Sprintf("RESEND_FROM_%d", i)))
		if k != "" {
			addAccount(k, f)
		}
	}

	return New(accounts, opts...)
}

// Len returns the number of accounts configured in the pool.
func (p *Pool) Len() int {
	return len(p.accounts)
}

// ActiveCount returns how many accounts are currently active (not rate limited or exhausted).
func (p *Pool) ActiveCount() int {
	count := 0
	for _, a := range p.accounts {
		if a.IsActive() {
			count++
		}
	}
	return count
}

// Send attempts delivery across the pool using round-robin and automatic failover.
func (p *Pool) Send(ctx context.Context, msg Message) (id, provider string, err error) {
	if len(p.accounts) == 0 {
		return "", "resend", fmt.Errorf("emailpool: no resend accounts configured")
	}

	numAccounts := len(p.accounts)
	startIndex := int((atomic.AddUint64(&p.counter, 1) - 1) % uint64(numAccounts))

	var lastErr error

	// Try each account once in round-robin sequence starting from startIndex
	for i := 0; i < numAccounts; i++ {
		idx := (startIndex + i) % numAccounts
		account := p.accounts[idx]

		if !account.IsActive() {
			continue
		}

		sendID, tryErr := p.sendWithAccount(ctx, account, msg, i > 0)
		if tryErr == nil {
			return sendID, "resend", nil
		}

		lastErr = tryErr
		// If tryErr indicates daily quota or rate limit, sendWithAccount already marked the account
		// so the loop immediately falls over to the next active account in the pool.
	}

	// If all accounts were inactive or exhausted, try the first account directly as final fallback
	if lastErr == nil && len(p.accounts) > 0 {
		sendID, tryErr := p.sendWithAccount(ctx, p.accounts[0], msg, true)
		return sendID, "resend", tryErr
	}

	return "", "resend", lastErr
}

func (p *Pool) sendWithAccount(ctx context.Context, acc *Account, msg Message, isRetry bool) (string, error) {
	from := msg.From
	if from == "" || strings.Contains(from, "halomail.dev") {
		from = acc.From
	}

	payload := map[string]any{
		"from":    from,
		"to":      msg.To,
		"subject": msg.Subject,
	}
	if msg.HTML != "" {
		payload["html"] = msg.HTML
	}
	if msg.Text != "" {
		payload["text"] = msg.Text
	}
	if msg.ReplyTo != "" {
		payload["reply_to"] = msg.ReplyTo
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	reqURL := p.baseURL + "/emails"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+acc.APIKey)
	req.Header.Set("Content-Type", "application/json")
	if msg.ID != "" {
		idempotencyKey := msg.ID
		if isRetry {
			idempotencyKey += "-pool"
		}
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))

	if resp.StatusCode < 300 {
		var out struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(respBody, &out); err == nil && out.ID != "" {
			return out.ID, nil
		}
		return "msg_sent", nil
	}

	// Handle 429 Rate Limits / Quota Exceeded
	if resp.StatusCode == http.StatusTooManyRequests {
		bodyStr := strings.ToLower(string(respBody))
		if strings.Contains(bodyStr, "daily_quota_exceeded") || strings.Contains(bodyStr, "daily email sending limit") {
			acc.MarkDailyExhausted()
			return "", fmt.Errorf("daily quota reached on %s: %s", acc.From, string(respBody))
		}
		// Burst rate limit (2 req/s)
		acc.MarkRateLimited(2 * time.Second)
		return "", fmt.Errorf("rate limited on %s: %s", acc.From, string(respBody))
	}

	// Handle Domain Verification error (403)
	if resp.StatusCode == http.StatusForbidden && strings.Contains(string(respBody), "not verified") {
		acc.MarkDailyExhausted()
		return "", fmt.Errorf("domain unverified for %s: %s", acc.From, string(respBody))
	}

	return "", fmt.Errorf("resend status %d: %s", resp.StatusCode, string(respBody))
}
