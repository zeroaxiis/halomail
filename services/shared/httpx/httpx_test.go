package httpx

import (
	"encoding/json"
	"errors"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestOwnerRequiresValidBearerToken(t *testing.T) {
	secret := strings.Repeat("k", 32)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, authn.Claims{
		OrgID: "org_1",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "usr_1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	verifier := authn.NewVerifier(secret)

	for _, header := range []string{"", "Basic dXNlcjpwYXNz", "Bearer not-a-jwt", token} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Authorization", header)
		if _, err := Owner(request, verifier); errs.KindOf(err) != errs.KindUnauthorized {
			t.Errorf("header %q: err = %v, want unauthorized", header, err)
		}
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	owner, err := Owner(request, verifier)
	if err != nil || owner != "usr_1" {
		t.Fatalf("owner = %q, err = %v", owner, err)
	}
}

func TestPeer(t *testing.T) {
	cases := map[string]string{
		"203.0.113.7:51234": "203.0.113.7",
		"[::1]:8080":        "::1",
		"203.0.113.7":       "203.0.113.7",
		"":                  "",
	}
	for in, want := range cases {
		if got := Peer(in); got != want {
			t.Errorf("Peer(%q) = %q, want %q", in, got, want)
		}
	}
}
