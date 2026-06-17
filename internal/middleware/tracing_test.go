package middleware

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// traceMockTransport is a self-contained RoundTripper for tracing tests.
type traceMockTransport struct {
	resp *http.Response
	err  error
}

func (m *traceMockTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func traceResp(status int) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}
}

func TestTracingTransport_RecordsSpanOnSuccess(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	tracer := tp.Tracer("test")

	mock := &traceMockTransport{resp: traceResp(200)}
	rt := &TracingTransport{Base: mock, Tracer: tracer}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/vpcs", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	span := spans[0]
	if !strings.Contains(span.Name, "GET") {
		t.Errorf("expected span name to contain 'GET', got: %s", span.Name)
	}
	if span.Status.Code != codes.Ok {
		t.Errorf("expected Ok status, got: %v", span.Status.Code)
	}
}

func TestTracingTransport_RecordsErrorSpanOn4xx(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	tracer := tp.Tracer("test")

	mock := &traceMockTransport{resp: traceResp(404)}
	rt := &TracingTransport{Base: mock, Tracer: tracer}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/missing", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Errorf("expected Error status for 404, got: %v", spans[0].Status.Code)
	}
}

func TestTracingTransport_RecordsTransportError(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	tracer := tp.Tracer("test")

	netErr := errors.New("connection refused")
	mock := &traceMockTransport{err: netErr}
	rt := &TracingTransport{Base: mock, Tracer: tracer}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/", nil)
	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error")
	}

	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Errorf("expected Error status for transport error, got: %v", spans[0].Status.Code)
	}
	if len(spans[0].Events) == 0 {
		t.Error("expected error event recorded on span")
	}
}
