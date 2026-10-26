package rpc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoogleAuthRequiresInternalSecret(t *testing.T) {
	for _, secret := range []string{"", "short", strings.Repeat("s", 32)} {
		mux := http.NewServeMux()
		NewHandlers(nil).MountGoogleAuth(mux, secret)
		r := httptest.NewRequest("POST", "/v1/auth/google", strings.NewReader(`{}`))
		r.Header.Set("X-OTP-Delivery-Secret", "wrong")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("unprotected /v1/auth/google: %d", w.Code)
		}
	}
}

func TestGoogleAuthRejectsMalformedBodies(t *testing.T) {
	secret := strings.Repeat("s", 32)
	mux := http.NewServeMux()
	NewHandlers(nil).MountGoogleAuth(mux, secret)
	for _, body := range []string{`{`, `{"unknown":1}`, `{} {}`, `{"email":"` + strings.Repeat("a", 9000) + `"}`} {
		r := httptest.NewRequest("POST", "/v1/auth/google", strings.NewReader(body))
		r.Header.Set("X-OTP-Delivery-Secret", secret)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("invalid body accepted: %d", w.Code)
		}
	}
}
