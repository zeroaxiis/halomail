package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProtectLimitsBodiesAndLoginAttempts(test *testing.T) {
	handler := Protect(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := io.ReadAll(request.Body); err != nil {
			writer.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest("POST", "/submit", strings.NewReader(strings.Repeat("x", 65537)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		test.Fatal("oversized body accepted")
	}
	for attempt := 0; attempt < 11; attempt++ {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("POST", "/halomail.identity.v1.AuthService/Login", strings.NewReader("{}")))
	}
	if response.Code != http.StatusTooManyRequests {
		test.Fatal("login attempts were not throttled")
	}
}

func TestProtectSetsSecurityHeaders(test *testing.T) {
	handler := Protect(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
	if response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Header().Get("Referrer-Policy") != "no-referrer" {
		test.Fatalf("security headers missing: %v", response.Header())
	}
}

func TestProtectThrottlesOnlyPosts(test *testing.T) {
	handler := Protect(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	for attempt := 0; attempt < 100; attempt++ {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/healthz", nil))
		if response.Code != http.StatusNoContent {
			test.Fatalf("GET %d throttled", attempt+1)
		}
	}
	for attempt := 0; attempt < 40; attempt++ {
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("POST", "/submit", strings.NewReader("{}")))
	}
	if response.Code != http.StatusTooManyRequests || response.Header().Get("Retry-After") != "2" {
		test.Fatalf("POST burst not throttled: %d %v", response.Code, response.Header())
	}
}
