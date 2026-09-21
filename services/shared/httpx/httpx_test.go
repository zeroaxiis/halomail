package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONSetsHeadersAndStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	JSON(rec, http.StatusCreated, map[string]string{"id": "f1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("cache control = %q", got)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"id":"f1"}` {
		t.Fatalf("body = %s", got)
	}
}
