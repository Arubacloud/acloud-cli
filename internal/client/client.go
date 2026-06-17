package client

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"go.opentelemetry.io/otel/trace"

	"github.com/Arubacloud/acloud-cli/internal/middleware"
	"github.com/Arubacloud/sdk-go/pkg/aruba"
)

// Params holds the fully-resolved configuration used to build or look up a cached SDK client.
type Params struct {
	ClientID        string
	ClientSecret    string
	BaseURL         string
	TokenIssuerURL  string
	UserAgent       string
	Debug           bool
	TelemetryTracer trace.Tracer // non-nil when --telemetry is active
}

type cachedState struct {
	mu              sync.Mutex
	client          aruba.Client
	clientID        string
	secret          string
	debug           bool
	baseURL         string
	tokenIssuer     string
	telemetryActive bool
	override        aruba.Client // set by SetForTesting only
}

var s = &cachedState{}

// Get returns the cached aruba.Client when Params match the previous call;
// otherwise it constructs a new client, caches it, and returns it.
func Get(p Params) (aruba.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	telemetryActive := p.TelemetryTracer != nil
	if s.client != nil &&
		s.clientID == p.ClientID &&
		s.secret == p.ClientSecret &&
		s.debug == p.Debug &&
		s.baseURL == p.BaseURL &&
		s.tokenIssuer == p.TokenIssuerURL &&
		s.telemetryActive == telemetryActive {
		return s.client, nil
	}

	options := aruba.DefaultOptions(p.ClientID, p.ClientSecret)
	if p.BaseURL != "" {
		options = options.WithBaseURL(p.BaseURL)
	}
	if p.TokenIssuerURL != "" {
		options = options.WithTokenIssuerURL(p.TokenIssuerURL)
	}
	if p.Debug {
		options = options.WithNativeLogger()
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		log.SetPrefix("[ArubaSDK] ")
	} else {
		options = options.WithDefaultLogger()
	}
	if p.UserAgent != "" {
		options = options.WithUserAgent(p.UserAgent)
	}

	if p.TelemetryTracer != nil {
		options = options.WithCustomHTTPClient(&http.Client{
			Transport: &middleware.TracingTransport{
				Base:   http.DefaultTransport,
				Tracer: p.TelemetryTracer,
			},
		})
	}

	c, err := aruba.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Aruba Cloud client: %w", err)
	}

	s.client = c
	s.clientID = p.ClientID
	s.secret = p.ClientSecret
	s.debug = p.Debug
	s.baseURL = p.BaseURL
	s.tokenIssuer = p.TokenIssuerURL
	s.telemetryActive = telemetryActive

	return c, nil
}

// Override returns the test-injected client and true when one is set, otherwise nil/false.
func Override() (aruba.Client, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.override != nil {
		return s.override, true
	}
	return nil, false
}

// SetForTesting injects a mock client returned directly by Override without loading config.
// Only for use in tests.
func SetForTesting(c aruba.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.override = c
}

// Reset replaces the cached state with a fresh zero value.
// Matches the original resetClientState() pattern; not safe for concurrent use with Get.
func Reset() {
	s = &cachedState{}
}
