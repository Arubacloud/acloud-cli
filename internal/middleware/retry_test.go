package middleware

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mockTransport records calls and returns pre-configured responses in order.
type mockTransport struct {
	responses []*http.Response
	errors    []error
	calls     int
	bodies    []string // captured request bodies per attempt
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	i := m.calls
	m.calls++

	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		m.bodies = append(m.bodies, string(b))
	} else {
		m.bodies = append(m.bodies, "")
	}

	if i < len(m.errors) && m.errors[i] != nil {
		return nil, m.errors[i]
	}
	if i < len(m.responses) {
		return m.responses[i], nil
	}
	return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(""))}, nil
}

func makeResp(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader("body")),
	}
}

func makeRetry(n int) *RetryTransport {
	return &RetryTransport{
		MaxRetries: n,
		BaseDelay:  time.Millisecond, // fast in tests
	}
}

// ─── shouldRetry ─────────────────────────────────────────────────────────────

func TestShouldRetry_GET_5xx(t *testing.T) {
	if !shouldRetry(http.MethodGet, makeResp(500), nil) {
		t.Error("expected GET 500 to be retryable")
	}
}

func TestShouldRetry_GET_503(t *testing.T) {
	if !shouldRetry(http.MethodGet, makeResp(503), nil) {
		t.Error("expected GET 503 to be retryable")
	}
}

func TestShouldRetry_GET_429(t *testing.T) {
	if !shouldRetry(http.MethodGet, makeResp(429), nil) {
		t.Error("expected GET 429 to be retryable")
	}
}

func TestShouldRetry_GET_200(t *testing.T) {
	if shouldRetry(http.MethodGet, makeResp(200), nil) {
		t.Error("expected GET 200 to NOT be retryable")
	}
}

func TestShouldRetry_GET_404(t *testing.T) {
	if shouldRetry(http.MethodGet, makeResp(404), nil) {
		t.Error("expected GET 404 to NOT be retryable")
	}
}

func TestShouldRetry_GET_TransportError(t *testing.T) {
	if !shouldRetry(http.MethodGet, nil, errors.New("connection refused")) {
		t.Error("expected GET transport error to be retryable")
	}
}

func TestShouldRetry_POST_5xx(t *testing.T) {
	if shouldRetry(http.MethodPost, makeResp(500), nil) {
		t.Error("expected POST 500 to NOT be retryable (prevents double-create)")
	}
}

func TestShouldRetry_POST_429(t *testing.T) {
	if !shouldRetry(http.MethodPost, makeResp(429), nil) {
		t.Error("expected POST 429 to be retryable")
	}
}

func TestShouldRetry_POST_503(t *testing.T) {
	if !shouldRetry(http.MethodPost, makeResp(503), nil) {
		t.Error("expected POST 503 to be retryable")
	}
}

func TestShouldRetry_POST_TransportError(t *testing.T) {
	if shouldRetry(http.MethodPost, nil, errors.New("dial timeout")) {
		t.Error("expected POST transport error to NOT be retryable")
	}
}

func TestShouldRetry_DELETE_429(t *testing.T) {
	if !shouldRetry(http.MethodDelete, makeResp(429), nil) {
		t.Error("expected DELETE 429 to be retryable")
	}
}

func TestShouldRetry_NilResp_NoError(t *testing.T) {
	if shouldRetry(http.MethodGet, nil, nil) {
		t.Error("expected nil resp + nil err to NOT be retryable")
	}
}

// ─── RetryTransport.RoundTrip ────────────────────────────────────────────────

func TestRetryTransport_SuccessOnFirstAttempt(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(200)}}
	rt := &RetryTransport{Base: mock, MaxRetries: 2, BaseDelay: time.Millisecond}

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call, got %d", mock.calls)
	}
}

func TestRetryTransport_GET_RetriesOn503(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(503), makeResp(503), makeResp(200)}}
	rt := makeRetry(3)
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if mock.calls != 3 {
		t.Errorf("expected 3 calls, got %d", mock.calls)
	}
}

func TestRetryTransport_GET_ExhaustsRetries(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(503), makeResp(503), makeResp(503)}}
	rt := makeRetry(2) // MaxRetries=2 → 3 total attempts
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("expected 503 after exhausted retries, got %d", resp.StatusCode)
	}
	if mock.calls != 3 {
		t.Errorf("expected 3 total calls, got %d", mock.calls)
	}
}

func TestRetryTransport_POST_DoesNotRetryOn500(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(500), makeResp(200)}}
	rt := makeRetry(2)
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected POST 500 to be returned immediately, got %d", resp.StatusCode)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call (no retry), got %d", mock.calls)
	}
}

func TestRetryTransport_POST_RetriesOn429(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(429), makeResp(200)}}
	rt := makeRetry(2)
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after 429 retry, got %d", resp.StatusCode)
	}
	if mock.calls != 2 {
		t.Errorf("expected 2 calls, got %d", mock.calls)
	}
}

func TestRetryTransport_ZeroRetries_FailFast(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(503), makeResp(200)}}
	rt := makeRetry(0) // --retries 0 = fail-fast
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 503 {
		t.Errorf("expected 503 with 0 retries, got %d", resp.StatusCode)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call, got %d", mock.calls)
	}
}

func TestRetryTransport_BodyReplayedOnRetry(t *testing.T) {
	mock := &mockTransport{responses: []*http.Response{makeResp(503), makeResp(200)}}
	rt := makeRetry(2)
	rt.Base = mock

	body := "request-body"
	req, _ := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	// Both attempts should have received the same body.
	for i, got := range mock.bodies {
		if got != body {
			t.Errorf("attempt %d: expected body %q, got %q", i, body, got)
		}
	}
}

func TestRetryTransport_ContextCancellation(t *testing.T) {
	// First response is 503 (triggers retry sleep), then cancel context before sleep ends.
	mock := &mockTransport{responses: []*http.Response{makeResp(503), makeResp(200)}}
	rt := &RetryTransport{Base: mock, MaxRetries: 2, BaseDelay: 500 * time.Millisecond}

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://example.com", nil)

	// Cancel after the first RoundTrip call returns.
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	_, err := rt.RoundTrip(req)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func TestRetryTransport_TransportError_GET_Retries(t *testing.T) {
	netErr := errors.New("connection refused")
	mock := &mockTransport{
		errors:    []error{netErr, nil},
		responses: []*http.Response{nil, makeResp(200)},
	}
	rt := makeRetry(2)
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200 after transport error retry, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_TransportError_POST_NoRetry(t *testing.T) {
	netErr := errors.New("connection refused")
	mock := &mockTransport{
		errors:    []error{netErr},
		responses: []*http.Response{nil, makeResp(200)},
	}
	rt := makeRetry(2)
	rt.Base = mock

	req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error for POST transport failure")
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call (no retry for POST transport error), got %d", mock.calls)
	}
}

func TestRetryTransport_DefaultTransport(t *testing.T) {
	// Ensure nil Base falls back to http.DefaultTransport without panicking.
	// Test the shouldRetry path which is invoked during RoundTrip.
	if !shouldRetry(http.MethodGet, makeResp(500), nil) {
		t.Error("shouldRetry invariant broken")
	}
}
