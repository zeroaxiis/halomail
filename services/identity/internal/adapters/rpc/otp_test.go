package rpc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOTPDeliveryEndpointsRequireServerSecret(t *testing.T) {
	for _, secret := range []string{"", "short", strings.Repeat("s", 32)} {
		mux := http.NewServeMux()
		NewHandlers(nil).MountOTP(mux, secret)
		for _, action := range []string{"request", "cancel"} {
			r := httptest.NewRequest("POST", "/v1/auth/otp/"+action, strings.NewReader(`{}`))
			r.Header.Set("X-OTP-Delivery-Secret", "wrong")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("unprotected %s: %d", action, w.Code)
			}
		}
	}
}

func TestOTPRejectsMalformedBodies(t *testing.T) {
	mux := http.NewServeMux()
	NewHandlers(nil).MountOTP(mux, strings.Repeat("s", 32))
	for _, body := range []string{`{`, `{"unknown":1}`, `{} {}`, `{"code":"` + strings.Repeat("1", 9000) + `"}`} {
		r := httptest.NewRequest("POST", "/v1/auth/otp/verify", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("invalid body accepted: %d", w.Code)
		}
	}
}
