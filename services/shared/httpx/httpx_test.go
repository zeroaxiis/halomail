package httpx

import (
	"encoding/json"
	"errors"
	"github.com/aashishrajdev/halomail/services/shared/errs"
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

func TestErrorMapsKindsToStatus(t *testing.T) {
	cases := []struct {
		err     error
		status  int
		message string
	}{
		{errs.Invalid("email required"), http.StatusBadRequest, "email required"},
		{errs.Unauthorized("sign in required"), http.StatusUnauthorized, "sign in required"},
		{errs.Forbidden("not yours"), http.StatusForbidden, "not yours"},
		{errs.NotFound("form not found"), http.StatusNotFound, "form not found"},
		{errs.Conflict("slug taken"), http.StatusConflict, "slug taken"},
		{errs.RateLimited("slow down"), http.StatusTooManyRequests, "slow down"},
		{errors.New("pq: syntax error"), http.StatusInternalServerError, "internal server error"},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		Error(rec, tc.err)
		var body struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if rec.Code != tc.status || body.Message != tc.message || body.Success {
			t.Errorf("%v: got %d %+v, want %d %q", tc.err, rec.Code, body, tc.status, tc.message)
		}
	}
}
