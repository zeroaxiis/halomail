package connectutil

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/aashishrajdev/halomail/services/shared/errs"
)

// ToConnect maps a domain error to a ConnectRPC error with the appropriate
// status code. Handlers should `return nil, connectutil.ToConnect(err)`.
func ToConnect(err error) error {
	if err == nil {
		return nil
	}
	// Already a connect error — pass through untouched.
	var ce *connect.Error
	if errors.As(err, &ce) {
		return err
	}

	var code connect.Code
	switch errs.KindOf(err) {
	case errs.KindInvalid:
		code = connect.CodeInvalidArgument
	case errs.KindNotFound:
		code = connect.CodeNotFound
	case errs.KindConflict:
		code = connect.CodeAlreadyExists
	case errs.KindUnauthorized:
		code = connect.CodeUnauthenticated
	case errs.KindForbidden:
		code = connect.CodePermissionDenied
	case errs.KindRateLimited:
		code = connect.CodeResourceExhausted
	default:
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}
	return connect.NewError(code, err)
}
