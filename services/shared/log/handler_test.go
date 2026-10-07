package log

import (
	"bytes"
	"context"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"strings"
	"testing"
)

func TestTraceHandlerAddsTraceAndSpanIDs(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(&traceHandler{Handler: slog.NewJSONHandler(&buf, nil)})

	traceID, _ := trace.TraceIDFromHex("0af7651916cd43dd8448eb211c80319c")
	spanID, _ := trace.SpanIDFromHex("b7ad6b7169203331")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	// With must keep the trace decoration, not unwrap to the base handler.
	logger.With("component", "test").InfoContext(ctx, "traced")
	out := buf.String()
	if !strings.Contains(out, `"trace_id":"0af7651916cd43dd8448eb211c80319c"`) || !strings.Contains(out, `"span_id":"b7ad6b7169203331"`) {
		t.Fatalf("trace correlation missing: %s", out)
	}

	buf.Reset()
	logger.Info("untraced")
	if strings.Contains(buf.String(), "trace_id") {
		t.Fatalf("trace id added without an active span: %s", buf.String())
	}
}
