package middleware

import (
	"log/slog"
	"net/http"
	"strings"
)

// LoggingTransport is an http.RoundTripper that logs each request and response
// at DEBUG level via the provided slog.Logger. The Authorization header value
// is always replaced with "Bearer [REDACTED]" before logging to prevent
// credential exposure.
type LoggingTransport struct {
	Base   http.RoundTripper
	Logger *slog.Logger
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	t.Logger.Debug("→ request",
		"method", req.Method,
		"url", req.URL.String(),
		"authorization", redactAuth(req.Header.Get("Authorization")),
	)

	resp, err := base.RoundTrip(req)
	if err != nil {
		t.Logger.Debug("← error", "err", err)
		return nil, err
	}

	t.Logger.Debug("← response",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
	)
	return resp, nil
}

// redactAuth replaces the token value in an Authorization header with
// "[REDACTED]", preserving the scheme prefix (e.g. "Bearer").
func redactAuth(v string) string {
	if v == "" {
		return ""
	}
	parts := strings.SplitN(v, " ", 2)
	if len(parts) == 2 {
		return parts[0] + " [REDACTED]"
	}
	return "[REDACTED]"
}
