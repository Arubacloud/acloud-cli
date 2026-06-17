package cmd

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/Arubacloud/acloud-cli/internal/client"
	"github.com/Arubacloud/acloud-cli/internal/telemetry"
	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// activeSpanCtx carries the root span context created in PersistentPreRunE so
// that newCtx() in root.go can derive child contexts from it. Package-level
// because cobra provides no other cross-hook context propagation mechanism.
var activeSpanCtx = context.Background()

// traceShutdown is the OTLP provider shutdown function set when --telemetry is
// active. Called in PersistentPostRunE to flush spans before the process exits.
var traceShutdown func(context.Context)

// activeSpan is the root command span; ended in PersistentPostRunE.
var activeSpan oteltrace.Span

func init() {
	telemetry.NoopSetup() // default no-op until --telemetry enables the real provider

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if skipClientInit(cmd) {
			return nil
		}

		// Telemetry: initialise OTLP provider and start root span when --telemetry is set.
		if enabled, _ := rootCmd.PersistentFlags().GetBool("telemetry"); enabled {
			endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
			shutdown, err := telemetry.Setup(context.Background(), endpoint)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[warn] telemetry setup failed: %v\n", err)
			} else {
				traceShutdown = shutdown
				ctx, span := telemetry.StartSpan(context.Background(), cmd.CommandPath())
				activeSpan = span
				activeSpanCtx = ctx
			}
		}

		_, err := GetArubaClient()
		return err
	}

	rootCmd.PersistentPostRunE = func(cmd *cobra.Command, args []string) error {
		if activeSpan != nil {
			activeSpan.End()
			activeSpan = nil
		}
		if traceShutdown != nil {
			traceShutdown(context.Background())
			traceShutdown = nil
		}
		activeSpanCtx = context.Background()
		return nil
	}
}

// skipClientInit returns true for commands that work without an SDK client:
// config/context management, shell completion helpers, and built-in help.
func skipClientInit(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "config", "context", "completion",
			"__complete", "__completeNoDesc", "help":
			return true
		}
	}
	return false
}

// GetArubaClient returns a cached aruba.Client built from the active profile config.
// The client is re-created whenever credentials, URLs, or the --debug flag change.
func GetArubaClient() (aruba.Client, error) {
	if c, ok := client.Override(); ok {
		return c, nil
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w. Please run 'acloud config set' to configure credentials", err)
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client ID or client secret not configured. Set ACLOUD_CLIENT_ID / ACLOUD_CLIENT_SECRET env vars or run 'acloud config set --client-id YOUR_CLIENT_ID'")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	tokenIssuer := cfg.TokenIssuerURL
	if tokenIssuer == "" {
		tokenIssuer = DefaultTokenIssuerURL
	}

	debug, _ := rootCmd.PersistentFlags().GetBool("debug")
	telemetryEnabled, _ := rootCmd.PersistentFlags().GetBool("telemetry")

	return client.Get(client.Params{
		ClientID:        cfg.ClientID,
		ClientSecret:    cfg.ClientSecret,
		BaseURL:         baseURL,
		TokenIssuerURL:  tokenIssuer,
		Debug:           debug,
		UserAgent:       "acloud-cli@" + rootCmd.Version,
		TelemetryTracer: tracerWhenEnabled(telemetryEnabled),
	})
}

// tracerWhenEnabled returns the global OTel tracer when telemetry is active,
// nil otherwise. A nil tracer in client.Params means no TracingTransport is injected.
func tracerWhenEnabled(enabled bool) oteltrace.Tracer {
	if !enabled {
		return nil
	}
	return otel.Tracer("github.com/Arubacloud/acloud-cli")
}

func resetClientState() {
	client.Reset()
	completionCacheReset()
}

func setClientForTesting(c aruba.Client) {
	client.SetForTesting(c)
}
