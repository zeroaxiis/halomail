package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decode(t *testing.T, rec *httptest.ResponseRecorder) response {
	t.Helper()
	var out response
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return out
}

func TestLivenessAlwaysOK(t *testing.T) {
	rec := httptest.NewRecorder()
	New().Liveness()(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q", got)
	}
	if out := decode(t, rec); out.Status != "ok" || len(out.Checks) != 0 {
		t.Fatalf("body = %+v", out)
	}
}

func TestReadinessReportsEachCheck(t *testing.T) {
	checker := New()
	checker.Register("postgres", func(context.Context) error { return nil })
	checker.Register("redis", func(context.Context) error { return nil })

	rec := httptest.NewRecorder()
	checker.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	out := decode(t, rec)
	if out.Status != "ok" || out.Checks["postgres"] != "ok" || out.Checks["redis"] != "ok" {
		t.Fatalf("body = %+v", out)
	}
}

func TestReadinessUnavailableWhenAnyCheckFails(t *testing.T) {
	checker := New()
	checker.Register("postgres", func(context.Context) error { return nil })
	checker.Register("redis", func(context.Context) error { return errors.New("dial tcp: refused") })

	rec := httptest.NewRecorder()
	checker.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	out := decode(t, rec)
	if out.Status != "unavailable" || out.Checks["redis"] != "dial tcp: refused" || out.Checks["postgres"] != "ok" {
		t.Fatalf("body = %+v", out)
	}
}
