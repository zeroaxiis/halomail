package connectutil

import (
	"connectrpc.com/connect"
	"context"
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
