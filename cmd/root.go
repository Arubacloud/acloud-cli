package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// clientState encapsulates the cached SDK client and its configuration (TD-018).
// Grouping them in a struct makes it easy to reset atomically in tests and
// prevents parallel-test races caused by separate package-level variables.
type clientState struct {
	mu          sync.Mutex
	client      aruba.Client
	clientID    string
	secret      string
	debug       bool
	baseURL     string
	tokenIssuer string
	override    aruba.Client // tests only: bypasses config loading when non-nil
}

var state = &clientState{}

// resetClientState resets the cached client and its configuration to zero values.
// Intended for use in tests to prevent state leaking between test cases.
func resetClientState() {
	state = &clientState{}
}

// setClientForTesting injects a mock client that GetArubaClient returns directly,
// bypassing config file loading entirely. Only for use in tests.
func setClientForTesting(c aruba.Client) {
	state.mu.Lock()
	defer state.mu.Unlock()
	state.override = c
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "acloud",
	Short:        "CLI for Aruba Cloud APIs",
	SilenceUsage: true, // Don't print usage on runtime errors
	Long: `acloud is a command-line interface for interacting with Aruba Cloud APIs.
It provides a simple and intuitive way to manage your Aruba Cloud resources
directly from your terminal.`,
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.acloud.yaml)")

	// Add global debug flag (TD-012: description warns about credential exposure)
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging (WARNING: may expose credentials and tokens in HTTP headers)")
	// Add global output format flag (TD-016)
	rootCmd.PersistentFlags().StringP("output", "o", "table", "Output format: table or json")
}

// GetArubaClient creates and returns an Aruba Cloud SDK client using stored credentials
// It automatically checks for the --debug flag from the root command to enable verbose logging
// The client is cached to avoid recreating it on every call, but is invalidated if credentials or debug flag change
func GetArubaClient() (aruba.Client, error) {
	// Short-circuit for tests: return the injected mock without loading config.
	state.mu.Lock()
	if state.override != nil {
		c := state.override
		state.mu.Unlock()
		return c, nil
	}
	state.mu.Unlock()

	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w. Please run 'acloud config set' to configure credentials", err)
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("client ID or client secret not configured. Please run 'acloud config set --client-id YOUR_CLIENT_ID --client-secret YOUR_CLIENT_SECRET'")
	}

	// Get base URL and token issuer URL from config, or use defaults
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	tokenIssuerURL := config.TokenIssuerURL
	if tokenIssuerURL == "" {
		tokenIssuerURL = DefaultTokenIssuerURL
	}

	// Check if debug flag is set on root command
	debugEnabled := false
	if rootCmd != nil {
		debugEnabled, _ = rootCmd.PersistentFlags().GetBool("debug")
	}

	// Check if we can reuse the cached client
	state.mu.Lock()
	defer state.mu.Unlock()

	// Reuse cached client if credentials, URLs, and debug flag haven't changed
	if state.client != nil &&
		state.clientID == config.ClientID &&
		state.secret == config.ClientSecret &&
		state.debug == debugEnabled &&
		state.baseURL == baseURL &&
		state.tokenIssuer == tokenIssuerURL {
		return state.client, nil
	}

	// Create SDK client with credentials using DefaultOptions
	// Only enable native logger when debug is enabled
	options := aruba.DefaultOptions(config.ClientID, config.ClientSecret)

	// Apply base URL if provided
	if baseURL != "" {
		options = options.WithBaseURL(baseURL)
	}

	// Apply token issuer URL if provided
	if tokenIssuerURL != "" {
		options = options.WithTokenIssuerURL(tokenIssuerURL)
	}

	if debugEnabled {
		options = options.WithNativeLogger()
		// Configure logger output for debug mode
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		log.SetPrefix("[ArubaSDK] ")
	} else {
		// Ensure logger is disabled when debug is off
		options = options.WithDefaultLogger()
	}

	client, err := aruba.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create Aruba Cloud client: %w", err)
	}

	// Cache the client and its configuration
	state.client = client
	state.clientID = config.ClientID
	state.secret = config.ClientSecret
	state.debug = debugEnabled
	state.baseURL = baseURL
	state.tokenIssuer = tokenIssuerURL

	return client, nil
}

// GetProjectID returns the project ID from the flag or current context
func GetProjectID(cmd *cobra.Command) (string, error) {
	// Try to get from flag first
	projectID, _ := cmd.Flags().GetString("project-id")
	if projectID != "" {
		return projectID, nil
	}

	// Try to get from context
	projectID, err := GetCurrentProjectID()
	if err != nil {
		return "", fmt.Errorf("project ID not specified. Use --project-id flag or set a context with 'acloud context use <name>'")
	}

	return projectID, nil
}

// newCtx returns a context with a 30-second timeout for SDK calls (TD-006).
func newCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// fmtAPIError formats an SDK API error response into a Go error (TD-001/TD-003).
func fmtAPIError(statusCode int, title, detail *string) error {
	msg := fmt.Sprintf("API error (status %d)", statusCode)
	if title != nil {
		msg += ": " + *title
	}
	if detail != nil {
		msg += " — " + *detail
	}
	return fmt.Errorf("%s", msg)
}

// confirmDelete prompts the user for confirmation before a destructive operation.
// Returns true if the user confirmed, false if they declined or stdin is non-interactive (TD-005).
func confirmDelete(resourceType, id string) (bool, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return false, fmt.Errorf("delete requires --yes/-y in non-interactive mode")
	}
	fmt.Printf("Are you sure you want to delete %s %s? (yes/no): ", resourceType, id)
	var response string
	fmt.Scanln(&response)
	if response != "yes" && response != "y" {
		fmt.Println("Delete cancelled")
		return false, nil
	}
	return true, nil
}

// readSecret prompts the user for a secret value with echo disabled (TD-011).
// Returns an error if stdin is not an interactive terminal.
func readSecret(prompt string) (string, error) {
	fi, err := os.Stdin.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		return "", fmt.Errorf("cannot read secret interactively: stdin is not a terminal; pass the flag explicitly instead")
	}
	fmt.Fprint(os.Stderr, prompt)
	secret, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // newline after hidden input
	if err != nil {
		return "", fmt.Errorf("reading secret: %w", err)
	}
	return string(secret), nil
}

// msgCreated returns a consistent success message for synchronous create operations (TD-020).
func msgCreated(kind, name string) string {
	return fmt.Sprintf("%s '%s' created successfully.", kind, name)
}

// msgCreatedAsync returns a consistent message for async create operations (TD-020).
func msgCreatedAsync(kind, name string) string {
	return fmt.Sprintf("%s '%s' creation initiated. Use 'get' to check status.", kind, name)
}

// msgUpdated returns a consistent success message for synchronous update operations (TD-020).
func msgUpdated(kind, name string) string {
	return fmt.Sprintf("%s '%s' updated successfully.", kind, name)
}

// msgUpdatedAsync returns a consistent message for async update operations (TD-020).
func msgUpdatedAsync(kind, name string) string {
	return fmt.Sprintf("%s '%s' update initiated. Use 'get' to check status.", kind, name)
}

// msgDeleted returns a consistent success message for delete operations (TD-020).
func msgDeleted(kind, name string) string {
	return fmt.Sprintf("%s '%s' deleted successfully.", kind, name)
}

// msgAction returns a consistent success message for arbitrary actions (TD-020).
func msgAction(kind, name, verb string) string {
	return fmt.Sprintf("%s '%s' %s successfully.", kind, name, verb)
}

// msgDryRun returns the dry-run output for a would-be delete (TD-019).
func msgDryRun(kind, id string) string {
	return fmt.Sprintf("[dry-run] Would delete %s '%s'. Resource exists and is accessible.", kind, id)
}

// listParams builds pagination RequestParameters from --limit and --offset flags (TD-017).
// Returns nil when neither flag is set, preserving the existing nil-means-no-options contract.
func listParams(cmd *cobra.Command) *types.RequestParameters {
	limit, _ := cmd.Flags().GetInt32("limit")
	offset, _ := cmd.Flags().GetInt32("offset")
	if limit == 0 && offset == 0 {
		return nil
	}
	params := &types.RequestParameters{}
	if limit > 0 {
		params.Limit = &limit
	}
	if offset > 0 {
		params.Offset = &offset
	}
	return params
}

// TableColumn represents a column definition for the table printer
type TableColumn struct {
	Header string // Column header name
	Width  int    // Column width for formatting
}

// PrintTable prints data in the format requested by the global --output flag (TD-016).
// When --output=json the rows are serialised as a JSON array of objects keyed by column header.
// When --output=table (the default) the existing fixed-width table format is used.
func PrintTable(headers []TableColumn, rows [][]string) {
	// Check global --output flag via rootCmd (avoids changing signature across all call sites)
	format := "table"
	if rootCmd != nil {
		format, _ = rootCmd.PersistentFlags().GetString("output")
	}

	if format == "json" {
		result := make([]map[string]string, 0, len(rows))
		for _, row := range rows {
			obj := make(map[string]string, len(headers))
			for i, col := range headers {
				if i < len(row) {
					obj[col.Header] = row[i]
				}
			}
			result = append(result, obj)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return
	}

	// Default: fixed-width table output
	formatStr := ""
	headerValues := make([]interface{}, len(headers))
	for i, col := range headers {
		formatStr += fmt.Sprintf("%%-%ds ", col.Width)
		headerValues[i] = col.Header
	}
	formatStr += "\n"
	fmt.Printf(formatStr, headerValues...)

	for _, row := range rows {
		rowValues := make([]interface{}, len(row))
		for i, val := range row {
			if len(headers) > i && len(val) > headers[i].Width {
				val = val[:headers[i].Width-3] + "..."
			}
			rowValues[i] = val
		}
		fmt.Printf(formatStr, rowValues...)
	}
}
