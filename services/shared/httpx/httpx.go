// Package httpx holds helpers for the plain HTTP (non-Connect) endpoints:
// JSON responses, domain error mapping, and bearer authentication.
package httpx

import (
	"encoding/json"
	"github.com/aashishrajdev/halomail/services/shared/authn"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"net"
	"net/http"
	"strings"
)

// JSON writes body as an uncacheable JSON response with the given status.
func JSON(writer http.ResponseWriter, status int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(body)
}

// Error writes err as a JSON failure. Unclassified errors become a generic
// 500 so internal details are not exposed.
func Error(writer http.ResponseWriter, err error) {
	status, message := http.StatusInternalServerError, "internal server error"
	switch errs.KindOf(err) {
	case errs.KindInvalid:
		status, message = 400, err.Error()
	case errs.KindUnauthorized:
		status, message = 401, err.Error()
	case errs.KindForbidden:
		status, message = 403, err.Error()
	case errs.KindNotFound:
		status, message = 404, err.Error()
	case errs.KindConflict:
		status, message = 409, err.Error()
	case errs.KindRateLimited:
		status, message = 429, err.Error()
	}
	JSON(writer, status, map[string]any{"success": false, "message": message})
}

// Owner returns the user id from the request bearer token.
func Owner(request *http.Request, verifier authn.Verifier) (string, error) {
	auth := request.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", errs.Unauthorized("sign in required")
	}
	owner, _, err := verifier.Verify(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return "", errs.Unauthorized("sign in required")
	}
	return owner, nil
}

// Peer strips the port from a RemoteAddr, returning it unchanged if it has none.
func Peer(address string) string {
	peer, _, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}
	return peer
}
