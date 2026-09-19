package connectutil

import (
	"connectrpc.com/connect"
	"errors"
	"github.com/aashishrajdev/halomail/services/shared/errs"
	"strings"
	"testing"
)

func TestToConnectMapsKinds(t *testing.T) {
	cases := map[errs.Kind]connect.Code{
		errs.KindInvalid:      connect.CodeInvalidArgument,
		errs.KindNotFound:     connect.CodeNotFound,
		errs.KindConflict:     connect.CodeAlreadyExists,
		errs.KindUnauthorized: connect.CodeUnauthenticated,
		errs.KindForbidden:    connect.CodePermissionDenied,
		errs.KindRateLimited:  connect.CodeResourceExhausted,
	}
	for kind, want := range cases {
		err := ToConnect(&errs.Error{Kind: kind, Message: "boom"})
		if got := connect.CodeOf(err); got != want {
			t.Errorf("kind %d mapped to %s, want %s", kind, got, want)
		}
	}
}

func TestToConnectPassesThroughNilAndConnectErrors(t *testing.T) {
	if ToConnect(nil) != nil {
		t.Fatal("nil error should stay nil")
	}
	original := connect.NewError(connect.CodeAborted, errors.New("retry"))
	if got := ToConnect(original); got != original {
		t.Fatalf("connect error was re-wrapped: %v", got)
	}
}

func TestToConnectHidesUnclassifiedErrors(t *testing.T) {
	for _, in := range []error{
		errors.New("pq: password authentication failed"),
		errs.Internal("pq: password authentication failed"),
	} {
		err := ToConnect(in)
		if got := connect.CodeOf(err); got != connect.CodeInternal {
			t.Fatalf("code = %s, want internal", got)
		}
		if strings.Contains(err.Error(), "password") {
			t.Fatalf("internal detail leaked to the client: %v", err)
		}
	}
}
