package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestFromFallsBackToDefault(t *testing.T) {
	if From(context.Background()) != slog.Default() {
		t.Fatal("context without a logger should yield slog.Default")
	}
}

func TestWithRequestIDTagsContextLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	ctx := Into(context.Background(), logger)
	if From(ctx) != logger {
		t.Fatal("From did not return the logger stored by Into")
	}

	From(WithRequestID(ctx, "req_1")).Info("handled")
	if !strings.Contains(buf.String(), `"request_id":"req_1"`) {
		t.Fatalf("request id missing from output: %s", buf.String())
	}
}
