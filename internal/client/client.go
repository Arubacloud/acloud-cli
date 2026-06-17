package client

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/Arubacloud/acloud-cli/internal/logging"
	"github.com/Arubacloud/acloud-cli/internal/middleware"
	"github.com/Arubacloud/sdk-go/pkg/aruba"
)

// Params holds the fully-resolved configuration used to build or look up a cached SDK client.
type Params struct {
	ClientID       string
	ClientSecret   string
	BaseURL        string
	TokenIssuerURL string
	UserAgent      string
	Debug          bool
	LogLevel       string
	LogFormat      string
}

type cachedState struct {
	mu          sync.Mutex
	client      aruba.Client
	clientID    string
	secret      string
	debug       bool
	baseURL     string
	tokenIssuer string
	logLevel    string
	logFormat   string
	override    aruba.Client // set by SetForTesting only
}

var s = &cachedState{}

// Get returns the cached aruba.Client when Params match the previous call;
// otherwise it constructs a new client, caches it, and returns it.
func Get(p Params) (aruba.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil &&
		s.clientID == p.ClientID &&
		s.secret == p.ClientSecret &&
		s.debug == p.Debug &&
		s.baseURL == p.BaseURL &&
		s.tokenIssuer == p.TokenIssuerURL &&
		s.logLevel == p.LogLevel &&
		s.logFormat == p.LogFormat {
		return s.client, nil
	}

	options := aruba.DefaultOptions(p.ClientID, p.ClientSecret).WithDefaultLogger()
	if p.BaseURL != "" {
		options = options.WithBaseURL(p.BaseURL)
	}
	if p.TokenIssuerURL != "" {
		options = options.WithTokenIssuerURL(p.TokenIssuerURL)
	}
	if p.UserAgent != "" {
		options = options.WithUserAgent(p.UserAgent)
	}

	// Inject a logging transport when debug-level logging is active so that
	// HTTP requests and responses are visible with Authorization headers
	// redacted. This replaces the former WithNativeLogger + global log.*
	// side-effects approach.
	if p.LogLevel == "debug" {
		options = options.WithCustomHTTPClient(&http.Client{
			Transport: &middleware.LoggingTransport{
				Base:   http.DefaultTransport,
				Logger: logging.L(),
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
	s.logLevel = p.LogLevel
	s.logFormat = p.LogFormat

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
