package health

import (
	"encoding/json"
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
