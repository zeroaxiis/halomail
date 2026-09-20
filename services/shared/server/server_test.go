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
