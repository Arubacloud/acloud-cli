package middleware

import (
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// TracingTransport is an http.RoundTripper that records an OpenTelemetry client
// span for each HTTP call. It propagates the span context from the incoming
// request context so child spans nest correctly under the command root span.
type TracingTransport struct {
	Base   http.RoundTripper
	Tracer oteltrace.Tracer
}

func (t *TracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	spanName := req.Method + " " + req.URL.Path
	ctx, span := t.Tracer.Start(req.Context(), spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
		oteltrace.WithAttributes(
			attribute.String("http.request.method", req.Method),
			attribute.String("url.full", req.URL.String()),
		),
	)
	defer span.End()

	resp, err := base.RoundTrip(req.WithContext(ctx))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))
	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, http.StatusText(resp.StatusCode))
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return resp, nil
}
