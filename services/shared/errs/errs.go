// Package errs defines transport-agnostic domain errors. Adapters (e.g. the
// ConnectRPC layer) translate a Kind into the appropriate wire status code,
// so the domain never imports a transport package.
package errs

import (
	"errors"
	"fmt"
)

// Kind classifies a domain error so adapters can pick a wire status code.
type Kind int

const (
	KindUnknown Kind = iota
	KindInvalid
	KindNotFound
	KindConflict
	KindUnauthorized
	KindForbidden
	KindRateLimited
	KindInternal
)

// Error is a domain error carrying a classification Kind.
type Error struct {
	Kind    Kind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func newf(k Kind, msg string, args ...any) *Error {
	return &Error{Kind: k, Message: fmt.Sprintf(msg, args...)}
}

// One constructor per Kind. msg is a fmt format string.
func Invalid(msg string, a ...any) *Error      { return newf(KindInvalid, msg, a...) }
func NotFound(msg string, a ...any) *Error     { return newf(KindNotFound, msg, a...) }
func Conflict(msg string, a ...any) *Error     { return newf(KindConflict, msg, a...) }
func Unauthorized(msg string, a ...any) *Error { return newf(KindUnauthorized, msg, a...) }
func Forbidden(msg string, a ...any) *Error    { return newf(KindForbidden, msg, a...) }
func RateLimited(msg string, a ...any) *Error  { return newf(KindRateLimited, msg, a...) }
func Internal(msg string, a ...any) *Error     { return newf(KindInternal, msg, a...) }

// Wrap attaches a Kind and message to an underlying error.
func Wrap(err error, k Kind, msg string) *Error {
	return &Error{Kind: k, Message: msg, Err: err}
}

// KindOf extracts the Kind from any error in the chain, defaulting to Unknown.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return KindUnknown
}
