package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

// logginMockTransport is a simple RoundTripper for logging tests.
type loggingMockTransport struct {
	resp *http.Response
	err  error
}

func (m *loggingMockTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func newLogLogger() (*slog.Logger, *strings.Builder) {
	buf := &strings.Builder{}
	h := slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return slog.New(h), buf
}

func logResp(status int) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(""))}
}

func TestLoggingTransport_LogsRequestAndResponse(t *testing.T) {
	logger, buf := newLogLogger()
	lt := &LoggingTransport{
		Base:   &loggingMockTransport{resp: logResp(200)},
		Logger: logger,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com/path", nil)
	req.Header.Set("Authorization", "Bearer secret-token")

	resp, err := lt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	out := buf.String()
	if !strings.Contains(out, "request") {
		t.Errorf("expected 'request' in log, got: %s", out)
	}
	if !strings.Contains(out, "response") {
		t.Errorf("expected 'response' in log, got: %s", out)
	}
	if !strings.Contains(out, "200") {
		t.Errorf("expected status code in log, got: %s", out)
	}
}

func TestLoggingTransport_RedactsAuthorizationHeader(t *testing.T) {
	logger, buf := newLogLogger()
	lt := &LoggingTransport{
		Base:   &loggingMockTransport{resp: logResp(200)},
		Logger: logger,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	req.Header.Set("Authorization", "Bearer super-secret-token-12345")

	lt.RoundTrip(req) //nolint:errcheck

	out := buf.String()
	if strings.Contains(out, "super-secret-token-12345") {
		t.Errorf("secret token must not appear in logs, got: %s", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in logs, got: %s", out)
	}
	if !strings.Contains(out, "Bearer") {
		t.Errorf("expected scheme prefix 'Bearer' preserved, got: %s", out)
	}
}

func TestLoggingTransport_LogsTransportError(t *testing.T) {
	logger, buf := newLogLogger()
	lt := &LoggingTransport{
		Base:   &loggingMockTransport{err: io.ErrUnexpectedEOF},
		Logger: logger,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	_, err := lt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error")
	}

	out := buf.String()
	if !strings.Contains(out, "error") {
		t.Errorf("expected 'error' in log, got: %s", out)
	}
}

func TestLoggingTransport_NoAuthHeader(t *testing.T) {
	logger, buf := newLogLogger()
	lt := &LoggingTransport{
		Base:   &loggingMockTransport{resp: logResp(201)},
		Logger: logger,
	}

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	resp, err := lt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	out := buf.String()
	if strings.Contains(out, "[REDACTED]") {
		t.Errorf("expected no [REDACTED] when no auth header, got: %s", out)
	}
}

func TestRedactAuth(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Bearer abc123", "Bearer [REDACTED]"},
		{"Token xyz", "Token [REDACTED]"},
		{"Basic dXNlcjpwYXNz", "Basic [REDACTED]"},
		{"plain-token", "[REDACTED]"},
		{"", ""},
	}
	for _, tc := range tests {
		got := redactAuth(tc.input)
		if got != tc.want {
			t.Errorf("redactAuth(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
