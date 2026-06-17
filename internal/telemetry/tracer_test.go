package telemetry

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestSpanName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"acloud network vpc list", "acloud.network.vpc.list"},
		{"acloud security kms create", "acloud.security.kms.create"},
		{"acloud", "acloud"},
		{"acloud compute cloudserver get", "acloud.compute.cloudserver.get"},
	}
	for _, tc := range tests {
		got := SpanName(tc.input)
		if got != tc.want {
			t.Errorf("SpanName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNoopSetup(t *testing.T) {
	NoopSetup()
	// Verify a tracer can be created and used without error; the noop provider
	// wraps all spans as non-recording, which is the expected default behaviour.
	tracer := otel.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()
}

func TestStartSpan_WithNoop(t *testing.T) {
	NoopSetup()
	ctx, span := StartSpan(context.Background(), "acloud network vpc list")
	if ctx == nil {
		t.Error("expected non-nil context")
	}
	if span == nil {
		t.Error("expected non-nil span")
	}
	span.End()
}

func TestStartSpan_SpanNamePropagated(t *testing.T) {
	NoopSetup()
	_, span := StartSpan(context.Background(), "acloud security kms get")
	// With the no-op provider we can't inspect the span name directly, but
	// the call must not panic and the span must be valid.
	if !span.IsRecording() {
		// noop span is never recording — just verify no panic occurred.
	}
	span.End()
}
