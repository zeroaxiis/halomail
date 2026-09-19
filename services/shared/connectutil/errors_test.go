package connectutil

import (
	"connectrpc.com/connect"
	"errors"
	"github.com/aashishrajdev/halomail/services/shared/errs"
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
