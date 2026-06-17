package middleware

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// RetryTransport is an http.RoundTripper that retries failed requests with
// exponential backoff and jitter. It buffers the request body once so it can
// be replayed on each attempt.
//
// Retry policy:
//   - GET: retried on any 5xx status, 429, 503, or transport error.
//   - POST/PUT/DELETE: retried only on 429 or 503 — prevents silent double-creates on
//     general 5xx failures.
//
// All retry warnings are written to stderr so they never corrupt -o json/yaml output.
type RetryTransport struct {
	Base       http.RoundTripper
	MaxRetries int
	BaseDelay  time.Duration
	Debug      bool
}

func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	// Buffer the request body once so it can be replayed on each attempt.
	var bodyBuf []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		bodyBuf, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	var resp *http.Response
	var err error

	for attempt := 0; attempt <= t.MaxRetries; attempt++ {
		r := req.Clone(req.Context())
		if bodyBuf != nil {
			r.Body = io.NopCloser(bytes.NewReader(bodyBuf))
			r.ContentLength = int64(len(bodyBuf))
		}

		resp, err = base.RoundTrip(r)

		if attempt == t.MaxRetries || !shouldRetry(req.Method, resp, err) {
			return resp, err
		}

		// Drain and close the failed response body before retrying.
		if resp != nil {
			io.Copy(io.Discard, resp.Body) //nolint:errcheck
			resp.Body.Close()
		}

		t.warnRetry(resp, err, attempt+1)

		delay := t.backoffDelay(attempt)
		if t.Debug {
			fmt.Fprintf(os.Stderr, "[debug] retry %d/%d: sleeping %s\n", attempt+1, t.MaxRetries, delay)
		}

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(delay):
		}
	}

	return resp, err
}

func (t *RetryTransport) warnRetry(resp *http.Response, err error, attempt int) {
	if resp != nil {
		fmt.Fprintf(os.Stderr, "[warn] API call failed (status %d), retrying (%d/%d)...\n",
			resp.StatusCode, attempt, t.MaxRetries)
	} else {
		fmt.Fprintf(os.Stderr, "[warn] API call failed (%v), retrying (%d/%d)...\n",
			err, attempt, t.MaxRetries)
	}
}

// backoffDelay returns the wait duration for the given attempt (0-indexed),
// using exponential backoff capped at 30 s plus up to 25% random jitter.
func (t *RetryTransport) backoffDelay(attempt int) time.Duration {
	delay := t.BaseDelay * time.Duration(1<<uint(attempt))
	const maxDelay = 30 * time.Second
	if delay > maxDelay || delay <= 0 {
		delay = maxDelay
	}
	if delay > 0 {
		jitter := time.Duration(rand.Int63n(int64(delay / 4))) //nolint:gosec
		delay += jitter
	}
	return delay
}

// shouldRetry reports whether the failed attempt should be retried given the
// HTTP method, response, and transport-level error.
func shouldRetry(method string, resp *http.Response, err error) bool {
	if err != nil {
		// Retry transport errors only for GET (safe to replay with no side effects).
		return method == http.MethodGet
	}
	if resp == nil {
		return false
	}
	// 429 Too Many Requests and 503 Service Unavailable: retry all methods.
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
		return true
	}
	// Other 5xx: only retry GET.
	if resp.StatusCode >= 500 {
		return method == http.MethodGet
	}
	return false
}
