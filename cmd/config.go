/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Config represents the runtime application configuration.
type Config struct {
	ClientID       string
	ClientSecret   string
	BaseURL        string
	TokenIssuerURL string
}

// configFile is the persisted representation stored in ~/.acloud.yaml.
type configFile struct {
	ClientID       string `yaml:"clientId"`
	ClientSecret   string `yaml:"clientSecret"`
	BaseURL        string `yaml:"baseUrl,omitempty"`
	TokenIssuerURL string `yaml:"tokenIssuerUrl,omitempty"`
}

type configShowDisplay struct {
	ClientID       string `json:"clientId" yaml:"clientId"`
	ClientSecret   string `json:"clientSecret" yaml:"clientSecret"`
	BaseURL        string `json:"baseUrl" yaml:"baseUrl"`
	TokenIssuerURL string `json:"tokenIssuerUrl" yaml:"tokenIssuerUrl"`
}

const (
	DefaultBaseURL        = "https://api.arubacloud.com"
	DefaultTokenIssuerURL = "https://mylogin.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
	maskedClientSecret    = "***"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage acloud configuration",
	Long:  `Configure acloud with your Aruba Cloud API credentials (clientId and clientSecret).`,
}

// configSetCmd represents the config set command
var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values",
	Long:  `Set configuration values for acloud. Pass clientId as a flag; clientSecret is read from ACLOUD_CLIENT_SECRET or prompted securely.`,
	RunE:  ConfigSetRun,
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current acloud configuration.`,
	RunE:  ConfigShowRun,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)

	// Flags for config set command
	configSetCmd.Flags().String("client-id", "", "Aruba Cloud API client ID (required)")
	configSetCmd.Flags().String("base-url", "", "Base URL for Aruba Cloud API (optional, default: https://api.arubacloud.com)")
	configSetCmd.Flags().String("token-issuer-url", "", "Token issuer URL for authentication (optional, default: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token)")
}

// xdgConfigDir returns the XDG config home directory (#176).
// Uses $XDG_CONFIG_HOME if set, otherwise falls back to $HOME/.config.
func xdgConfigDir() (string, error) {
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

// GetConfigPath returns the XDG-compliant path to the config file (#176).
// New path: $XDG_CONFIG_HOME/acloud/config.yaml (default: ~/.config/acloud/config.yaml).
func GetConfigPath() (string, error) {
	dir, err := xdgConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "acloud", "config.yaml"), nil
}

// legacyConfigPath returns the pre-XDG config path (~/.acloud.yaml).
func legacyConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acloud.yaml"), nil
}

// migrateLegacyConfig copies ~/.acloud.yaml → XDG path when the new file is
// absent. Prints a one-time notice to stderr so the user is informed (#176).
func migrateLegacyConfig() {
	newPath, err := GetConfigPath()
	if err != nil {
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		return // new file already exists; nothing to do
	}
	oldPath, err := legacyConfigPath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return // old file absent; normal first-run case
	}
	if err := os.MkdirAll(filepath.Dir(newPath), FilePermDirAll); err != nil {
		return
	}
	if err := os.WriteFile(newPath, data, FilePermConfig); err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "notice: config migrated from %s to %s\n", oldPath, newPath)
}

// LoadConfig loads the configuration from the config file, with env var overrides.
// Env vars ACLOUD_CLIENT_ID, ACLOUD_CLIENT_SECRET, ACLOUD_BASE_URL, and
// ACLOUD_TOKEN_ISSUER_URL take precedence over the config file when set.
// If the config file is missing but credentials are supplied via env vars,
// the file is not required and no error is returned.
// Migrates legacy ~/.acloud.yaml to the XDG path on first load (#176).
func LoadConfig() (*Config, error) {
	migrateLegacyConfig()
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	config := &Config{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		// File missing: only succeeds if env vars supply the credentials.
		envID := os.Getenv("ACLOUD_CLIENT_ID")
		envSecret := os.Getenv("ACLOUD_CLIENT_SECRET")
		if envID == "" && envSecret == "" {
			return nil, err
		}
		config.ClientID = envID
		config.ClientSecret = envSecret
		if v := os.Getenv("ACLOUD_BASE_URL"); v != "" {
			config.BaseURL = v
		}
		if v := os.Getenv("ACLOUD_TOKEN_ISSUER_URL"); v != "" {
			config.TokenIssuerURL = v
		}
		return config, nil
	}

	fileCfg := &configFile{}
	if err2 := yaml.Unmarshal(data, fileCfg); err2 != nil {
		return nil, fmt.Errorf("config file %s is corrupted (%w). Delete it and run 'acloud config set' to reconfigure", configPath, err2)
	}
	*config = configFromFile(fileCfg)

	// Env vars override file values — useful for CI/CD and e2e tests.
	if v := os.Getenv("ACLOUD_CLIENT_ID"); v != "" {
		config.ClientID = v
	}
	if v := os.Getenv("ACLOUD_CLIENT_SECRET"); v != "" {
		config.ClientSecret = v
	}
	if v := os.Getenv("ACLOUD_BASE_URL"); v != "" {
		config.BaseURL = v
	}
	if v := os.Getenv("ACLOUD_TOKEN_ISSUER_URL"); v != "" {
		config.TokenIssuerURL = v
	}

	return config, nil
}

// SaveConfig saves the configuration to the XDG config file, creating
// the parent directory if it does not exist (#176).
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), FilePermDirAll); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(fileFromConfig(config))
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, FilePermConfig)
}

func configFromFile(fileCfg *configFile) Config {
	if fileCfg == nil {
		return Config{}
	}
	return Config{
		ClientID:       fileCfg.ClientID,
		ClientSecret:   fileCfg.ClientSecret,
		BaseURL:        fileCfg.BaseURL,
		TokenIssuerURL: fileCfg.TokenIssuerURL,
	}
}

func fileFromConfig(config *Config) configFile {
	if config == nil {
		return configFile{}
	}
	return configFile{
		ClientID:       config.ClientID,
		ClientSecret:   config.ClientSecret,
		BaseURL:        config.BaseURL,
		TokenIssuerURL: config.TokenIssuerURL,
	}
}

// ─── Config Set ───────────────────────────────────────────────────────────────

// ConfigSetArgs holds the parsed and validated arguments for the config set command.
type ConfigSetArgs struct {
	ClientID       string
	ClientSecret   string
	BaseURL        string
	TokenIssuerURL string
}

// NewConfigSetArgsFromCobraCommand parses flags and validates them.
func NewConfigSetArgsFromCobraCommand(cmd *cobra.Command) (*ConfigSetArgs, error) {
	args := &ConfigSetArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// ParseFromCobraCommand reads flag values into the args struct.
func (a *ConfigSetArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error
	if a.ClientID, err = cmd.Flags().GetString("client-id"); err != nil {
		errs = append(errs, err)
	}
	if a.BaseURL, err = cmd.Flags().GetString("base-url"); err != nil {
		errs = append(errs, err)
	}
	if a.TokenIssuerURL, err = cmd.Flags().GetString("token-issuer-url"); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Validate performs pure validation on the parsed args.
// Note: client-id and client-secret emptiness is validated against the
// existing config inside ConfigSet, because an existing config may supply them.
func (a *ConfigSetArgs) Validate() error {
	return nil
}

// ConfigSet saves new configuration values, merging with any existing config.
func ConfigSet(_ context.Context, args ConfigSetArgs) error {
	// Load existing config or start fresh.
	config, err := LoadConfig()
	if err != nil {
		config = &Config{}
	}

	// If no client-id is available from either the existing config or the flag, fail.
	if config.ClientID == "" && args.ClientID == "" {
		return fmt.Errorf("--client-id is required")
	}
	// ACLOUD_CLIENT_SECRET is accepted for automation and takes precedence over prompt.
	if args.ClientSecret == "" {
		args.ClientSecret = os.Getenv("ACLOUD_CLIENT_SECRET")
	}

	// If no client-secret is available, prompt interactively.
	if config.ClientSecret == "" && args.ClientSecret == "" {
		prompted, err := readSecret("Enter client secret: ")
		if err != nil {
			return fmt.Errorf("client secret is required (set ACLOUD_CLIENT_SECRET or provide interactive input): %w", err)
		}
		args.ClientSecret = prompted
	}

	// Apply provided values on top of existing config.
	if args.ClientID != "" {
		config.ClientID = args.ClientID
	}
	if args.ClientSecret != "" {
		config.ClientSecret = args.ClientSecret
	}
	if args.BaseURL != "" {
		config.BaseURL = args.BaseURL
	}
	if args.TokenIssuerURL != "" {
		config.TokenIssuerURL = args.TokenIssuerURL
	}

	// Final check: both fields must be non-empty after merging.
	if config.ClientID == "" || config.ClientSecret == "" {
		return fmt.Errorf("both client-id and client-secret are required")
	}

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	fmt.Println("Configuration updated successfully")
	if args.ClientID != "" {
		fmt.Printf("  Client ID: %s\n", args.ClientID)
	}
	if args.ClientSecret != "" {
		fmt.Println("  Client Secret: ********")
	}
	if args.BaseURL != "" {
		fmt.Printf("  Base URL: %s\n", args.BaseURL)
	}
	if args.TokenIssuerURL != "" {
		fmt.Printf("  Token Issuer URL: %s\n", args.TokenIssuerURL)
	}
	return nil
}

// ConfigSetRun is the RunE wiring for the config set command.
func ConfigSetRun(cmd *cobra.Command, _ []string) error {
	args, err := NewConfigSetArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ConfigSet(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ─── Config Show ──────────────────────────────────────────────────────────────

// ConfigShowArgs holds the (empty) arguments for the config show command.
type ConfigShowArgs struct{}

// NewConfigShowArgsFromCobraCommand parses and validates args for config show.
func NewConfigShowArgsFromCobraCommand(_ *cobra.Command) (*ConfigShowArgs, error) {
	return &ConfigShowArgs{}, nil
}

// ConfigShow prints the current configuration.
func ConfigShow(_ context.Context, _ ConfigShowArgs) error {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("No configuration found. Please run 'acloud config set' to create one.")
		return nil
	}
	display := configShowDisplayFromConfig(config)

	if resolveOutputFormat() == OutputFormatJSON || resolveOutputFormat() == OutputFormatYAML {
		PrintOutput(display, nil, nil)
		return nil
	}
	if resolveOutputFormat() == OutputFormatTableJSON || resolveOutputFormat() == OutputFormatTableYAML {
		headers := []TableColumn{
			{Header: "CLIENT_ID", Width: 24},
			{Header: "CLIENT_SECRET", Width: 16},
			{Header: "BASE_URL", Width: 40},
			{Header: "TOKEN_ISSUER_URL", Width: 40},
		}
		rows := [][]string{{display.ClientID, display.ClientSecret, display.BaseURL, display.TokenIssuerURL}}
		PrintOutput(nil, headers, rows)
		return nil
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  Client ID: %s\n", config.ClientID)
	if config.ClientSecret != "" {
		fmt.Printf("  Client Secret: %s\n", maskedClientSecret)
	} else {
		fmt.Println("  Client Secret: (not set)")
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL + " (default)"
	}
	fmt.Printf("  Base URL: %s\n", baseURL)
	tokenIssuerURL := config.TokenIssuerURL
	if tokenIssuerURL == "" {
		tokenIssuerURL = DefaultTokenIssuerURL + " (default)"
	}
	fmt.Printf("  Token Issuer URL: %s\n", tokenIssuerURL)
	return nil
}

func configShowDisplayFromConfig(config *Config) configShowDisplay {
	secret := "(not set)"
	if config.ClientSecret != "" {
		secret = maskedClientSecret
	}
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	tokenIssuerURL := config.TokenIssuerURL
	if tokenIssuerURL == "" {
		tokenIssuerURL = DefaultTokenIssuerURL
	}

	return configShowDisplay{
		ClientID:       config.ClientID,
		ClientSecret:   secret,
		BaseURL:        baseURL,
		TokenIssuerURL: tokenIssuerURL,
	}
}

// ConfigShowRun is the RunE wiring for the config show command.
func ConfigShowRun(cmd *cobra.Command, _ []string) error {
	args, err := NewConfigShowArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ConfigShow(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
