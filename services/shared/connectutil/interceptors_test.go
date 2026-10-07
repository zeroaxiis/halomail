package connectutil

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
)

func discardLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestRecoveryConvertsPanicToInternal(t *testing.T) {
	handler := Recovery(discardLogger())(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		panic("boom")
	})
	_, err := handler(context.Background(), connect.NewRequest(&struct{}{}))
	if got := connect.CodeOf(err); got != connect.CodeInternal {
		t.Fatalf("code = %s, want internal", got)
	}
}

func TestLoggingPassesThroughResultAndError(t *testing.T) {
	want := connect.NewResponse(&struct{}{})
	ok := Logging(discardLogger())(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		return want, nil
	})
	got, err := ok(context.Background(), connect.NewRequest(&struct{}{}))
	if err != nil || got != want {
		t.Fatalf("response changed by logging: %v %v", got, err)
	}

	failing := Logging(discardLogger())(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("missing"))
	})
	if _, err = failing(context.Background(), connect.NewRequest(&struct{}{})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("error changed by logging: %v", err)
	}
}
