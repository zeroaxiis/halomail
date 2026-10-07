// Package health exposes liveness and readiness probes. Liveness reports the
// process is up; readiness runs registered dependency checks (db, redis, ...).
package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Check returns nil when the dependency is healthy.
type Check func(ctx context.Context) error

// Checker holds the registered readiness probes. It is safe for concurrent use.
type Checker struct {
	mu     sync.RWMutex
	checks map[string]Check
}

// New returns a Checker with no probes registered.
func New() *Checker { return &Checker{checks: make(map[string]Check)} }

// Register adds a named readiness probe.
func (c *Checker) Register(name string, fn Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = fn
}

type response struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// Liveness always returns 200 — the binary is running.
func (c *Checker) Liveness() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, response{Status: "ok"})
	}
}

// Readiness runs every registered check; 503 if any fail.
func (c *Checker) Readiness() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		c.mu.RLock()
		defer c.mu.RUnlock()

		out := response{Status: "ok", Checks: make(map[string]string, len(c.checks))}
		code := http.StatusOK
		for name, fn := range c.checks {
			if err := fn(ctx); err != nil {
				out.Checks[name] = err.Error()
				out.Status = "unavailable"
				code = http.StatusServiceUnavailable
				continue
			}
			out.Checks[name] = "ok"
		}
		writeJSON(w, code, out)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
