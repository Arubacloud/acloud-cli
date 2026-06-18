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

// configFile is the persisted representation of a single credential set.
type configFile struct {
	ClientID       string `yaml:"clientId"`
	ClientSecret   string `yaml:"clientSecret"`
	BaseURL        string `yaml:"baseUrl,omitempty"`
	TokenIssuerURL string `yaml:"tokenIssuerUrl,omitempty"`
}

// multiProfileConfigFile is the envelope used when the config contains named
// profiles. Detection: if the YAML has a top-level "profiles" mapping the file
// is in multi-profile format; otherwise it is treated as single-profile (#180).
type multiProfileConfigFile struct {
	ActiveProfile string                `yaml:"activeProfile,omitempty"`
	Profiles      map[string]configFile `yaml:"profiles"`
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
	configCmd.AddCommand(configProfileCmd)
	configProfileCmd.AddCommand(configProfileListCmd)
	configProfileCmd.AddCommand(configProfileSetCmd)
	configProfileCmd.AddCommand(configProfileDeleteCmd)
	configProfileCmd.AddCommand(configProfileUseCmd)

	// Flags for config set command
	configSetCmd.Flags().String("client-id", "", "Aruba Cloud API client ID (required)")
	configSetCmd.Flags().String("base-url", "", "Base URL for Aruba Cloud API (optional, default: https://api.arubacloud.com)")
	configSetCmd.Flags().String("token-issuer-url", "", "Token issuer URL for authentication (optional, default: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token)")

	// Flags for config profile set
	configProfileSetCmd.Flags().String("client-id", "", "Aruba Cloud API client ID")
	configProfileSetCmd.Flags().String("client-secret", "", "Aruba Cloud API client secret (or use ACLOUD_CLIENT_SECRET)")
	configProfileSetCmd.Flags().String("base-url", "", "Base URL for Aruba Cloud API")
	configProfileSetCmd.Flags().String("token-issuer-url", "", "Token issuer URL for authentication")
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

// ─── Multi-profile support (#180) ────────────────────────────────────────────

// getActiveProfile returns the profile name for the current invocation.
// Priority: --profile flag > ACLOUD_PROFILE env var > activeProfile in config file > "default".
func getActiveProfile() string {
	if rootCmd != nil {
		if p, _ := rootCmd.PersistentFlags().GetString("profile"); p != "" {
			return p
		}
	}
	if p := os.Getenv("ACLOUD_PROFILE"); p != "" {
		return p
	}
	if p := loadActiveProfileFromFile(); p != "" {
		return p
	}
	return "default"
}

// loadActiveProfileFromFile reads the activeProfile field from the config file.
// Returns empty string if the file is missing or has no activeProfile set.
func loadActiveProfileFromFile() string {
	configPath, err := GetConfigPath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	var mpf multiProfileConfigFile
	if yaml.Unmarshal(data, &mpf) == nil {
		return mpf.ActiveProfile
	}
	return ""
}

// setActiveProfile persists the given profile name as the active profile in the
// config file. Flag (--profile) and env var (ACLOUD_PROFILE) still take precedence
// at runtime.
func setActiveProfile(name string) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), FilePermDirAll); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	var mpf multiProfileConfigFile
	if data, err := os.ReadFile(configPath); err == nil {
		_ = yaml.Unmarshal(data, &mpf)
	}
	if mpf.Profiles == nil {
		mpf.Profiles = make(map[string]configFile)
	}
	mpf.ActiveProfile = name
	data, err := yaml.Marshal(mpf)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, FilePermConfig)
}

// loadAllProfiles reads the config file and returns a map of all profiles.
// Single-profile files are wrapped as {"default": singleProfile} transparently.
func loadAllProfiles() (map[string]configFile, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Try multi-profile format first.
	var mpf multiProfileConfigFile
	if yaml.Unmarshal(data, &mpf) == nil && len(mpf.Profiles) > 0 {
		return mpf.Profiles, nil
	}

	// Fall back to single-profile format.
	var fileCfg configFile
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("config file %s is corrupted: %w", configPath, err)
	}
	return map[string]configFile{"default": fileCfg}, nil
}

// saveAllProfiles writes all profiles to the config file in multi-profile format,
// preserving the activeProfile field if one was already stored.
func saveAllProfiles(profiles map[string]configFile) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), FilePermDirAll); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	mpf := multiProfileConfigFile{Profiles: profiles}
	if existing, err := os.ReadFile(configPath); err == nil {
		var existingMpf multiProfileConfigFile
		if yaml.Unmarshal(existing, &existingMpf) == nil {
			mpf.ActiveProfile = existingMpf.ActiveProfile
		}
	}
	data, err := yaml.Marshal(mpf)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, FilePermConfig)
}

// sortedProfileNames returns profile names in deterministic order.
func sortedProfileNames(profiles map[string]configFile) []string {
	names := make([]string, 0, len(profiles))
	for k := range profiles {
		names = append(names, k)
	}
	// "default" first, then alphabetical.
	result := make([]string, 0, len(names))
	for _, n := range names {
		if n == "default" {
			result = append([]string{"default"}, result...)
		} else {
			result = append(result, n)
		}
	}
	return result
}

// LoadConfig loads the configuration from the config file, with env var overrides.
// Env vars ACLOUD_CLIENT_ID, ACLOUD_CLIENT_SECRET, ACLOUD_BASE_URL, and
// ACLOUD_TOKEN_ISSUER_URL take precedence over the config file when set.
// If the config file is missing but credentials are supplied via env vars,
// the file is not required and no error is returned.
// Migrates legacy ~/.acloud.yaml to the XDG path on first load (#176).
// Supports multi-profile config via --profile flag or ACLOUD_PROFILE env var (#180).
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

	// Try multi-profile format.
	var mpf multiProfileConfigFile
	if yaml.Unmarshal(data, &mpf) == nil && len(mpf.Profiles) > 0 {
		profileName := getActiveProfile()
		fileCfg, ok := mpf.Profiles[profileName]
		if !ok {
			return nil, fmt.Errorf("profile %q not found in %s. Run 'acloud config profile list' to see available profiles", profileName, configPath)
		}
		*config = configFromFile(&fileCfg)
	} else {
		// Single-profile format (backward compat).
		fileCfg := &configFile{}
		if err2 := yaml.Unmarshal(data, fileCfg); err2 != nil {
			return nil, fmt.Errorf("config file %s is corrupted (%w). Delete it and run 'acloud config set' to reconfigure", configPath, err2)
		}
		*config = configFromFile(fileCfg)
	}

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

// SaveConfig saves the configuration to the active profile (#176, #180).
// On first write it produces multi-profile format; subsequent calls preserve
// all existing profiles and update only the active one.
func SaveConfig(config *Config) error {
	profiles, err := loadAllProfiles()
	if err != nil {
		// New file or unreadable: start fresh.
		profiles = make(map[string]configFile)
	}
	profiles[getActiveProfile()] = fileFromConfig(config)
	return saveAllProfiles(profiles)
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

// ─── config profile ───────────────────────────────────────────────────────────

var configProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage named credential profiles",
	Long:  `Manage named credential profiles for acloud (dev/staging/prod accounts).`,
}

var configProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all credential profiles",
	Args:  cobra.NoArgs,
	RunE:  ConfigProfileListRun,
}

var configProfileSetCmd = &cobra.Command{
	Use:   "set <profile-name>",
	Short: "Create or update a credential profile",
	Args:  cobra.ExactArgs(1),
	RunE:  ConfigProfileSetRun,
}

var configProfileDeleteCmd = &cobra.Command{
	Use:   "delete <profile-name>",
	Short: "Delete a credential profile",
	Args:  cobra.ExactArgs(1),
	RunE:  ConfigProfileDeleteRun,
}

var configProfileUseCmd = &cobra.Command{
	Use:   "use <profile-name>",
	Short: "Switch the active profile",
	Long:  `Switch the active credential profile. The selection is persisted in the config file and overridden by --profile flag or ACLOUD_PROFILE env var.`,
	Args:  cobra.ExactArgs(1),
	RunE:  ConfigProfileUseRun,
}

// ConfigProfileListRun lists all profiles, marking the active one with *.
func ConfigProfileListRun(_ *cobra.Command, _ []string) error {
	profiles, err := loadAllProfiles()
	if err != nil {
		fmt.Println("No profiles configured. Run 'acloud config set' to create one.")
		return nil
	}
	active := getActiveProfile()
	headers := []TableColumn{
		{Header: "PROFILE", Width: 20},
		{Header: "CLIENT_ID", Width: 30},
		{Header: "BASE_URL", Width: 40},
	}
	var rows [][]string
	for _, name := range sortedProfileNames(profiles) {
		p := profiles[name]
		marker := ""
		if name == active {
			marker = "* "
		}
		rows = append(rows, []string{marker + name, p.ClientID, p.BaseURL})
	}
	PrintOutput(nil, headers, rows)
	return nil
}

// ConfigProfileSetRun creates or updates a named profile.
func ConfigProfileSetRun(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	clientID, _ := cmd.Flags().GetString("client-id")
	clientSecret, _ := cmd.Flags().GetString("client-secret")
	baseURL, _ := cmd.Flags().GetString("base-url")
	tokenIssuerURL, _ := cmd.Flags().GetString("token-issuer-url")

	// Accept client secret from env var if not provided as flag.
	if clientSecret == "" {
		clientSecret = os.Getenv("ACLOUD_CLIENT_SECRET")
	}

	profiles, err := loadAllProfiles()
	if err != nil {
		profiles = make(map[string]configFile)
	}

	// Merge with existing profile if present.
	existing := profiles[profileName]
	if clientID != "" {
		existing.ClientID = clientID
	}
	if clientSecret != "" {
		existing.ClientSecret = clientSecret
	}
	if baseURL != "" {
		existing.BaseURL = baseURL
	}
	if tokenIssuerURL != "" {
		existing.TokenIssuerURL = tokenIssuerURL
	}

	if existing.ClientID == "" {
		return fmt.Errorf("--client-id is required when creating a new profile")
	}
	if existing.ClientSecret == "" {
		prompted, err := readSecret(fmt.Sprintf("Enter client secret for profile %q: ", profileName))
		if err != nil {
			return fmt.Errorf("client secret is required (set ACLOUD_CLIENT_SECRET or provide interactive input): %w", err)
		}
		existing.ClientSecret = prompted
	}

	profiles[profileName] = existing
	if err := saveAllProfiles(profiles); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	fmt.Printf("Profile %q saved.\n", profileName)
	return nil
}

// ConfigProfileDeleteRun removes a named profile.
func ConfigProfileDeleteRun(_ *cobra.Command, args []string) error {
	profileName := args[0]

	profiles, err := loadAllProfiles()
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}
	if _, ok := profiles[profileName]; !ok {
		return fmt.Errorf("profile %q not found", profileName)
	}
	delete(profiles, profileName)

	if err := saveAllProfiles(profiles); err != nil {
		return fmt.Errorf("saving profiles: %w", err)
	}
	fmt.Printf("Profile %q deleted.\n", profileName)
	return nil
}

// ConfigProfileUseRun switches the active profile persisted in the config file.
func ConfigProfileUseRun(_ *cobra.Command, args []string) error {
	profileName := args[0]

	profiles, err := loadAllProfiles()
	if err != nil {
		return fmt.Errorf("loading profiles: %w", err)
	}
	if _, ok := profiles[profileName]; !ok {
		return fmt.Errorf("profile %q not found. Run 'acloud config profile list' to see available profiles", profileName)
	}

	if err := setActiveProfile(profileName); err != nil {
		return fmt.Errorf("switching profile: %w", err)
	}
	fmt.Printf("Switched to profile %q.\n", profileName)
	return nil
}
