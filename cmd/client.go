package cmd

import (
	"fmt"

	"github.com/Arubacloud/acloud-cli/internal/client"
	"github.com/Arubacloud/sdk-go/pkg/aruba"
)

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

	return client.Get(client.Params{
		ClientID:       cfg.ClientID,
		ClientSecret:   cfg.ClientSecret,
		BaseURL:        baseURL,
		TokenIssuerURL: tokenIssuer,
		Debug:          debug,
		UserAgent:      "acloud-cli@" + rootCmd.Version,
	})
}

func resetClientState() {
	client.Reset()
	completionCacheReset()
}

func setClientForTesting(c aruba.Client) {
	client.SetForTesting(c)
}
