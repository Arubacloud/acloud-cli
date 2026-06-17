package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const instrumentationName = "github.com/Arubacloud/acloud-cli"

// Setup initialises an OTLP/HTTP tracer provider and registers it globally.
// endpoint is the base URL of the OTLP collector (e.g. "http://localhost:4318").
// If endpoint is empty the function reads OTEL_EXPORTER_OTLP_ENDPOINT from the
// environment via the standard OTel SDK defaults.
//
// Returns a shutdown function that flushes and closes the exporter.
// The caller must invoke it when the command finishes.
func Setup(ctx context.Context, endpoint string) (shutdown func(context.Context), err error) {
	opts := []otlptracehttp.Option{otlptracehttp.WithInsecure()}
	if endpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(endpoint))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	return func(ctx context.Context) {
		_ = tp.Shutdown(ctx)
	}, nil
}

// NoopSetup registers a no-op tracer provider globally. Called when
// --telemetry is not set to ensure otel.Tracer() returns a safe no-op.
func NoopSetup() {
	otel.SetTracerProvider(noop.NewTracerProvider())
}

// StartSpan starts a root span for the given command path (e.g.
// "acloud network vpc list") and returns the derived context and span.
// The span name follows the pattern acloud.<family>.<resource>.<action>.
func StartSpan(ctx context.Context, commandPath string, attrs ...attribute.KeyValue) (context.Context, oteltrace.Span) {
	name := SpanName(commandPath)
	tracer := otel.Tracer(instrumentationName)
	ctx, span := tracer.Start(ctx, name,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(attrs...),
	)
	return ctx, span
}

// SpanName converts a cobra command path to the dot-separated span name
// convention used by acloud (e.g. "acloud network vpc list" →
// "acloud.network.vpc.list").
func SpanName(commandPath string) string {
	return strings.ReplaceAll(strings.TrimSpace(commandPath), " ", ".")
}
