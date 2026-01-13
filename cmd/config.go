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
	ClientID     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

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
	Run: func(cmd *cobra.Command, args []string) {
		clientID, _ := cmd.Flags().GetString("client-id")
		clientSecret, _ := cmd.Flags().GetString("client-secret")

		if clientID == "" && clientSecret == "" {
			fmt.Println("Error: At least one of --client-id or --client-secret must be provided")
			os.Exit(1)
		}

		// Load existing config or create new one
		config, err := LoadConfig()
		if err != nil {
			// If config doesn't exist, create a new one
			config = &Config{}
		}

		// Update only provided values
		if clientID != "" {
			config.ClientID = clientID
		}
		if clientSecret != "" {
			config.ClientSecret = clientSecret
		}

		// Save config
		if err := SaveConfig(config); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configuration updated successfully")
		if clientID != "" {
			fmt.Printf("  Client ID: %s\n", clientID)
		}
		if clientSecret != "" {
			fmt.Println("  Client Secret: ********")
		}
	},
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current acloud configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := LoadConfig()
		if err != nil {
			fmt.Println("No configuration found. Please run 'acloud config set' to create one.")
			return
		}

		fmt.Println("Current configuration:")
		fmt.Printf("  Client ID: %s\n", config.ClientID)
		if config.ClientSecret != "" {
			fmt.Println("  Client Secret: ********")
		} else {
			fmt.Println("  Client Secret: (not set)")
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)

	// Flags for config set command
	configSetCmd.Flags().String("client-id", "", "Aruba Cloud API client ID")
	configSetCmd.Flags().String("client-secret", "", "Aruba Cloud API client secret")
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".acloud.yaml"), nil
}

// LoadConfig loads the configuration from the config file
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
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

	return os.WriteFile(configPath, data, 0600)
}
