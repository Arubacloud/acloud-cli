package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "acloud",
	Short:        "CLI for Aruba Cloud APIs",
	SilenceUsage: true,
	Long: `acloud is a command-line interface for interacting with Aruba Cloud APIs.
It provides a simple and intuitive way to manage your Aruba Cloud resources
directly from your terminal.`,
}

// SetVersion sets the CLI version string shown by --version.
// Pass the value injected via ldflags at build time; pass "" to show "dev".
func SetVersion(v string) {
	if v == "" {
		v = "dev"
	}
	rootCmd.Version = v
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging (WARNING: may expose credentials and tokens in HTTP headers)")
	rootCmd.PersistentFlags().StringP("output", "o", OutputFormatTable, "Output format: table|std|standard, table-json|std-json, table-yaml|std-yaml, json, yaml")
	rootCmd.PersistentFlags().String("timeout", "30s", "Timeout for API calls (e.g. 30s, 2m, 5m)")
	rootCmd.PersistentFlags().String("profile", "", "Credential profile to use (overrides ACLOUD_PROFILE env var)")
	rootCmd.PersistentFlags().Bool("telemetry", false, "Enable OpenTelemetry tracing (set OTEL_EXPORTER_OTLP_ENDPOINT to configure the collector)")
}

// GetProjectID returns the project ID from the flag or current context.
func GetProjectID(cmd *cobra.Command) (string, error) {
	projectID, _ := cmd.Flags().GetString("project-id")
	if projectID != "" {
		return projectID, nil
	}
	projectID, err := GetCurrentProjectID()
	if err != nil {
		return "", fmt.Errorf("project ID not specified. Use --project-id flag or set a context with 'acloud context use <name>'")
	}
	return projectID, nil
}

// newCtx returns a context whose timeout is governed by the global --timeout flag
// (default 30s). When a root tracing span is active it is used as the parent
// context so that SDK calls nest under the command span.
func newCtx() (context.Context, context.CancelFunc) {
	d := 30 * time.Second
	if rootCmd != nil {
		if raw, err := rootCmd.PersistentFlags().GetString("timeout"); err == nil && raw != "" {
			if parsed, err := time.ParseDuration(raw); err == nil && parsed > 0 {
				d = parsed
			}
		}
	}
	return context.WithTimeout(activeSpanCtx, d)
}
