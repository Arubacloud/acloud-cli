/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	ClientID       string `yaml:"clientId"`
	ClientSecret   string `yaml:"clientSecret"`
	BaseURL        string `yaml:"baseUrl,omitempty"`
	TokenIssuerURL string `yaml:"tokenIssuerUrl,omitempty"`
}

const (
	DefaultBaseURL        = "https://api.arubacloud.com"
	DefaultTokenIssuerURL = "https://mylogin.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token"
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
	Long:  `Set configuration values for acloud, such as clientId and clientSecret.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID, err := cmd.Flags().GetString("client-id")
		if err != nil {
			return err
		}
		clientSecret, err := cmd.Flags().GetString("client-secret")
		if err != nil {
			return err
		}
		baseURL, err := cmd.Flags().GetString("base-url")
		if err != nil {
			return err
		}
		tokenIssuerURL, err := cmd.Flags().GetString("token-issuer-url")
		if err != nil {
			return err
		}

		// Load existing config or create new one
		config, err := LoadConfig()
		if err != nil {
			// If config doesn't exist, create a new one
			config = &Config{}
		}

		// Validate required fields
		if config.ClientID == "" && clientID == "" {
			return fmt.Errorf("--client-id is required")
		}
		// If --client-secret was not provided and there is no existing secret, prompt interactively
		if config.ClientSecret == "" && clientSecret == "" {
			prompted, err := readSecret("Enter client secret: ")
			if err != nil {
				return fmt.Errorf("--client-secret is required: %w", err)
			}
			clientSecret = prompted
		}

		// Update only provided values
		if clientID != "" {
			config.ClientID = clientID
		}
		if clientSecret != "" {
			config.ClientSecret = clientSecret
		}
		if baseURL != "" {
			config.BaseURL = baseURL
		}
		if tokenIssuerURL != "" {
			config.TokenIssuerURL = tokenIssuerURL
		}

		// Final validation: both clientID and clientSecret must be set
		if config.ClientID == "" || config.ClientSecret == "" {
			return fmt.Errorf("both client-id and client-secret are required")
		}

		// Save config
		if err := SaveConfig(config); err != nil {
			return fmt.Errorf("saving configuration: %w", err)
		}

		fmt.Println("Configuration updated successfully")
		if clientID != "" {
			fmt.Printf("  Client ID: %s\n", clientID)
		}
		if clientSecret != "" {
			fmt.Println("  Client Secret: ********")
		}
		if baseURL != "" {
			fmt.Printf("  Base URL: %s\n", baseURL)
		}
		if tokenIssuerURL != "" {
			fmt.Printf("  Token Issuer URL: %s\n", tokenIssuerURL)
		}
		return nil
	},
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current acloud configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig()
		if err != nil {
			fmt.Println("No configuration found. Please run 'acloud config set' to create one.")
			return nil
		}

		fmt.Println("Current configuration:")
		fmt.Printf("  Client ID: %s\n", config.ClientID)
		if config.ClientSecret != "" {
			fmt.Println("  Client Secret: ********")
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
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)

	// Flags for config set command
	configSetCmd.Flags().String("client-id", "", "Aruba Cloud API client ID (required)")
	configSetCmd.Flags().String("client-secret", "", "Aruba Cloud API client secret (required)")
	configSetCmd.Flags().String("base-url", "", "Base URL for Aruba Cloud API (optional, default: https://api.arubacloud.com)")
	configSetCmd.Flags().String("token-issuer-url", "", "Token issuer URL for authentication (optional, default: https://login.aruba.it/auth/realms/cmp-new-apikey/protocol/openid-connect/token)")
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".acloud.yaml"), nil
}

// LoadConfig loads the configuration from the config file, with env var overrides.
// Env vars ACLOUD_CLIENT_ID, ACLOUD_CLIENT_SECRET, ACLOUD_BASE_URL, and
// ACLOUD_TOKEN_ISSUER_URL take precedence over the config file when set.
// If the config file is missing but credentials are supplied via env vars,
// the file is not required and no error is returned.
func LoadConfig() (*Config, error) {
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

	if err2 := yaml.Unmarshal(data, config); err2 != nil {
		return nil, fmt.Errorf("config file %s is corrupted (%w). Delete it and run 'acloud config set' to reconfigure", configPath, err2)
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

// SaveConfig saves the configuration to the config file
func SaveConfig(config *Config) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, FilePermConfig)
}
