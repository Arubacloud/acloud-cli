package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "") // use HOME/.config, not CI's XDG dir

	// Test with non-existent config
	config, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error when config file doesn't exist")
	}
	if config != nil {
		t.Error("LoadConfig() should return nil config when file doesn't exist")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create test config
	testConfig := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	// Save config
	err := SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file exists at XDG path
	configPath := filepath.Join(tmpDir, "acloud", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("SaveConfig() did not create config file at XDG path")
	}

	// Load and verify
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after SaveConfig() error = %v", err)
	}

	if loadedConfig.ClientID != testConfig.ClientID {
		t.Errorf("LoadConfig() ClientID = %v, want %v", loadedConfig.ClientID, testConfig.ClientID)
	}

	if loadedConfig.ClientSecret != testConfig.ClientSecret {
		t.Errorf("LoadConfig() ClientSecret = %v, want %v", loadedConfig.ClientSecret, testConfig.ClientSecret)
	}
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Get config path
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	// XDG path: $XDG_CONFIG_HOME/acloud/config.yaml
	expectedPath := filepath.Join(tmpDir, "acloud", "config.yaml")
	if configPath != expectedPath {
		t.Errorf("GetConfigPath() = %v, want %v", configPath, expectedPath)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Write invalid YAML to the XDG path, creating the directory first.
	xdgPath := filepath.Join(tmpDir, "acloud", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(xdgPath), 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(xdgPath, []byte("this is not valid yaml: ["), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	config, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error for invalid YAML")
	}
	if config != nil {
		t.Error("LoadConfig() should return nil config for invalid YAML")
	}
}

func TestSaveConfig_EmptyConfig(t *testing.T) {
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	emptyConfig := &Config{}

	err := SaveConfig(emptyConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after SaveConfig() error = %v", err)
	}

	if loadedConfig.ClientID != "" {
		t.Errorf("LoadConfig() ClientID = %v, want empty string", loadedConfig.ClientID)
	}

	if loadedConfig.ClientSecret != "" {
		t.Errorf("LoadConfig() ClientSecret = %v, want empty string", loadedConfig.ClientSecret)
	}
}

func TestSaveConfig_PartialConfig(t *testing.T) {
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	partialConfig := &Config{
		ClientID: "test-client-id-only",
	}

	err := SaveConfig(partialConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after SaveConfig() error = %v", err)
	}

	if loadedConfig.ClientID != "test-client-id-only" {
		t.Errorf("LoadConfig() ClientID = %v, want test-client-id-only", loadedConfig.ClientID)
	}

	if loadedConfig.ClientSecret != "" {
		t.Errorf("LoadConfig() ClientSecret = %v, want empty string", loadedConfig.ClientSecret)
	}
}

func TestGetConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "acloud", "config.yaml")
	if configPath != expectedPath {
		t.Errorf("GetConfigPath() = %v, want %v", configPath, expectedPath)
	}
}

func TestLoadConfig_EnvVarsNoFile(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	origID := os.Getenv("ACLOUD_CLIENT_ID")
	origSecret := os.Getenv("ACLOUD_CLIENT_SECRET")
	origBase := os.Getenv("ACLOUD_BASE_URL")
	origToken := os.Getenv("ACLOUD_TOKEN_ISSUER_URL")
	defer func() {
		os.Setenv("ACLOUD_CLIENT_ID", origID)
		os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)
		os.Setenv("ACLOUD_BASE_URL", origBase)
		os.Setenv("ACLOUD_TOKEN_ISSUER_URL", origToken)
	}()

	os.Setenv("ACLOUD_CLIENT_ID", "env-id")
	os.Setenv("ACLOUD_CLIENT_SECRET", "env-secret")
	os.Setenv("ACLOUD_BASE_URL", "https://custom.example.com")
	os.Setenv("ACLOUD_TOKEN_ISSUER_URL", "https://token.example.com")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() should succeed with env vars: %v", err)
	}
	if config.ClientID != "env-id" {
		t.Errorf("ClientID = %q, want env-id", config.ClientID)
	}
	if config.ClientSecret != "env-secret" {
		t.Errorf("ClientSecret = %q, want env-secret", config.ClientSecret)
	}
	if config.BaseURL != "https://custom.example.com" {
		t.Errorf("BaseURL = %q, want https://custom.example.com", config.BaseURL)
	}
	if config.TokenIssuerURL != "https://token.example.com" {
		t.Errorf("TokenIssuerURL = %q, want https://token.example.com", config.TokenIssuerURL)
	}
}

func TestLoadConfig_EnvVarOverridesFile(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	if err := SaveConfig(&Config{ClientID: "file-id", ClientSecret: "file-secret"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	origID := os.Getenv("ACLOUD_CLIENT_ID")
	origSecret := os.Getenv("ACLOUD_CLIENT_SECRET")
	origBase := os.Getenv("ACLOUD_BASE_URL")
	origToken := os.Getenv("ACLOUD_TOKEN_ISSUER_URL")
	defer func() {
		os.Setenv("ACLOUD_CLIENT_ID", origID)
		os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)
		os.Setenv("ACLOUD_BASE_URL", origBase)
		os.Setenv("ACLOUD_TOKEN_ISSUER_URL", origToken)
	}()

	os.Setenv("ACLOUD_CLIENT_ID", "override-id")
	os.Setenv("ACLOUD_CLIENT_SECRET", "override-secret")
	os.Setenv("ACLOUD_BASE_URL", "https://override.api.com")
	os.Setenv("ACLOUD_TOKEN_ISSUER_URL", "https://override.token.com")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig(): %v", err)
	}
	if config.ClientID != "override-id" {
		t.Errorf("ClientID = %q, want override-id", config.ClientID)
	}
	if config.BaseURL != "https://override.api.com" {
		t.Errorf("BaseURL = %q, want https://override.api.com", config.BaseURL)
	}
	if config.TokenIssuerURL != "https://override.token.com" {
		t.Errorf("TokenIssuerURL = %q, want https://override.token.com", config.TokenIssuerURL)
	}
}

func TestConfigSetCmd_Success(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origID := os.Getenv("ACLOUD_CLIENT_ID")
	origSecret := os.Getenv("ACLOUD_CLIENT_SECRET")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Setenv("ACLOUD_CLIENT_SECRET", "env-secret")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("ACLOUD_CLIENT_ID", origID)
		os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)
	}()

	out, err := runCmdCapture(nil, []string{"config", "set", "--client-id", "test-id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "updated successfully") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestConfigSetCmd_MissingClientID(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origID := os.Getenv("ACLOUD_CLIENT_ID")
	origSecret := os.Getenv("ACLOUD_CLIENT_SECRET")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "") // use HOME/.config, not CI's XDG dir
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Setenv("ACLOUD_CLIENT_SECRET", "env-secret")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("ACLOUD_CLIENT_ID", origID)
		os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)
	}()

	_, err := runCmdCapture(nil, []string{"config", "set"})
	if err == nil {
		t.Fatal("expected error for missing --client-id")
	}
	if !strings.Contains(err.Error(), "client-id") {
		t.Errorf("expected client-id mention in error, got: %s", err.Error())
	}
}

func TestConfigSetCmd_ClientSecretFlagRejected(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	_, err := runCmdCapture(nil, []string{"config", "set", "--client-id", "test-id", "--client-secret", "test-secret"})
	if err == nil {
		t.Fatal("expected error for unknown --client-secret flag")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("expected unknown flag error, got: %s", err.Error())
	}
}

func TestConfigShowCmd(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	if err := SaveConfig(&Config{ClientID: "show-id", ClientSecret: "show-secret"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	out, err := runCmdCapture(nil, []string{"config", "show"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "show-id") {
		t.Errorf("expected client ID in output, got: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("expected redacted secret in output, got: %s", out)
	}
	if strings.Contains(out, "show-secret") {
		t.Errorf("plaintext secret leaked in output: %s", out)
	}
}

func TestConfigShowCmd_JSONMasksClientSecret(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	if err := SaveConfig(&Config{ClientID: "show-id", ClientSecret: "show-secret"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	out, err := runCmdCapture(nil, []string{"config", "show", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "show-secret") {
		t.Fatalf("plaintext secret leaked in JSON output: %s", out)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("expected valid JSON output, got error: %v\noutput: %s", err, out)
	}
	if got["clientSecret"] != maskedClientSecret {
		t.Fatalf("clientSecret = %v, want %q", got["clientSecret"], maskedClientSecret)
	}
}

func TestConfigShowCmd_YAMLMasksClientSecret(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	if err := SaveConfig(&Config{ClientID: "show-id", ClientSecret: "show-secret"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	out, err := runCmdCapture(nil, []string{"config", "show", "--output", "yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "show-secret") {
		t.Fatalf("plaintext secret leaked in YAML output: %s", out)
	}

	var got map[string]any
	if err := yaml.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("expected valid YAML output, got error: %v\noutput: %s", err, out)
	}
	if got["clientSecret"] != maskedClientSecret {
		t.Fatalf("clientSecret = %v, want %q", got["clientSecret"], maskedClientSecret)
	}
}

func TestConfigShowCmd_NoConfig(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "") // use HOME/.config, not CI's XDG dir
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}()

	out, err := runCmdCapture(nil, []string{"config", "show"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No configuration found") {
		t.Errorf("expected no-config message, got: %s", out)
	}
}

// ─── Layer-1: Validate ────────────────────────────────────────────────────────

func TestConfigSetArgs_Validate(t *testing.T) {
	// ConfigSetArgs.Validate is intentionally no-op: required-field checks are
	// deferred to ConfigSet because they depend on the existing config on disk.
	// Verify that Validate always returns nil regardless of field values.
	cases := []struct {
		name string
		args ConfigSetArgs
	}{
		{"all empty", ConfigSetArgs{}},
		{"only client-id", ConfigSetArgs{ClientID: "id"}},
		{"only client-secret", ConfigSetArgs{ClientSecret: "secret"}},
		{"all set", ConfigSetArgs{ClientID: "id", ClientSecret: "secret", BaseURL: "https://x", TokenIssuerURL: "https://y"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.args.Validate(); err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestNewConfigSetArgsFromCobraCommand_ParseError(t *testing.T) {
	// A cobra.Command with none of the expected flags registered causes
	// GetString to return an error for every flag, which should surface as
	// ErrParsingFailed from the constructor.
	bare := &cobra.Command{}
	_, err := NewConfigSetArgsFromCobraCommand(bare)
	if err == nil {
		t.Fatal("expected ErrParsingFailed for command with no flags, got nil")
	}
	if !errors.Is(err, ErrParsingFailed) {
		t.Errorf("expected ErrParsingFailed, got: %v", err)
	}
}

// ─── Layer-2: Operation ───────────────────────────────────────────────────────

func withTempHome(t *testing.T) func() {
	t.Helper()
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME") // ensure HOME/.config is used, not CI's XDG dir
	return func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}
}

func TestConfigSet_Operation_Success(t *testing.T) {
	defer withTempHome(t)()
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Unsetenv("ACLOUD_CLIENT_SECRET")

	out := captureStdout(func() {
		err := ConfigSet(context.Background(), ConfigSetArgs{
			ClientID:     "op-id",
			ClientSecret: "op-secret",
		})
		if err != nil {
			t.Errorf("ConfigSet() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "updated successfully") {
		t.Errorf("expected success message, got: %s", out)
	}
	if !strings.Contains(out, "op-id") {
		t.Errorf("expected client ID in output, got: %s", out)
	}

	// Verify file was written correctly.
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig after ConfigSet: %v", err)
	}
	if loaded.ClientID != "op-id" {
		t.Errorf("ClientID = %q, want op-id", loaded.ClientID)
	}
	if loaded.ClientSecret != "op-secret" {
		t.Errorf("ClientSecret = %q, want op-secret", loaded.ClientSecret)
	}
}

func TestConfigSet_Operation_MissingClientID(t *testing.T) {
	defer withTempHome(t)()
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Unsetenv("ACLOUD_CLIENT_SECRET")

	err := ConfigSet(context.Background(), ConfigSetArgs{ClientSecret: "secret"})
	if err == nil {
		t.Fatal("expected error for missing client-id")
	}
	if !strings.Contains(err.Error(), "client-id") {
		t.Errorf("expected client-id mention, got: %v", err)
	}
}

func TestConfigSet_Operation_UpdatesExistingConfig(t *testing.T) {
	defer withTempHome(t)()
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Unsetenv("ACLOUD_CLIENT_SECRET")

	// Write initial config.
	if err := SaveConfig(&Config{ClientID: "old-id", ClientSecret: "old-secret"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// Update only the base URL — existing credentials are preserved.
	err := ConfigSet(context.Background(), ConfigSetArgs{
		BaseURL: "https://custom.example.com",
	})
	if err != nil {
		t.Fatalf("ConfigSet() error: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.ClientID != "old-id" {
		t.Errorf("ClientID = %q, want old-id", loaded.ClientID)
	}
	if loaded.BaseURL != "https://custom.example.com" {
		t.Errorf("BaseURL = %q, want https://custom.example.com", loaded.BaseURL)
	}
}

func TestConfigSet_Operation_UsesEnvClientSecret(t *testing.T) {
	defer withTempHome(t)()
	origSecret := os.Getenv("ACLOUD_CLIENT_SECRET")
	os.Setenv("ACLOUD_CLIENT_SECRET", "env-op-secret")
	defer os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)

	err := ConfigSet(context.Background(), ConfigSetArgs{ClientID: "env-op-id"})
	if err != nil {
		t.Fatalf("ConfigSet() error: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if loaded.ClientSecret != "env-op-secret" {
		t.Errorf("ClientSecret = %q, want env-op-secret", loaded.ClientSecret)
	}
}

func TestConfigSet_Operation_MissingClientSecret_NonInteractive(t *testing.T) {
	defer withTempHome(t)()
	os.Unsetenv("ACLOUD_CLIENT_SECRET")

	err := ConfigSet(context.Background(), ConfigSetArgs{ClientID: "id-without-secret"})
	if err == nil {
		t.Fatal("expected error when no secret is available in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "set ACLOUD_CLIENT_SECRET") {
		t.Errorf("expected ACLOUD_CLIENT_SECRET guidance, got: %v", err)
	}
}

func TestConfigShow_Operation_WithConfig(t *testing.T) {
	defer withTempHome(t)()
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Unsetenv("ACLOUD_CLIENT_SECRET")

	if err := SaveConfig(&Config{ClientID: "show-op-id", ClientSecret: "show-secret"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	out := captureStdout(func() {
		err := ConfigShow(context.Background(), ConfigShowArgs{})
		if err != nil {
			t.Errorf("ConfigShow() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "show-op-id") {
		t.Errorf("expected client ID in output, got: %s", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("expected redacted secret in output, got: %s", out)
	}
	if strings.Contains(out, "show-secret") {
		t.Errorf("plaintext secret leaked in output: %s", out)
	}
	if !strings.Contains(out, DefaultBaseURL+" (default)") {
		t.Errorf("expected default base URL label, got: %s", out)
	}
}

func TestConfigShow_Operation_NoConfig(t *testing.T) {
	defer withTempHome(t)()
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Unsetenv("ACLOUD_CLIENT_SECRET")

	out := captureStdout(func() {
		err := ConfigShow(context.Background(), ConfigShowArgs{})
		if err != nil {
			t.Errorf("ConfigShow() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No configuration found") {
		t.Errorf("expected no-config message, got: %s", out)
	}
}

func TestMigrateLegacyConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}
	tmpHome := t.TempDir()
	tmpXDG := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", tmpXDG)

	// Write a legacy config at ~/.acloud.yaml
	legacyPath := filepath.Join(tmpHome, ".acloud.yaml")
	legacyYAML := "clientId: migrated-id\nclientSecret: migrated-secret\n"
	if err := os.WriteFile(legacyPath, []byte(legacyYAML), 0600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	// Calling LoadConfig should trigger migration to XDG path
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after migration: %v", err)
	}
	if cfg.ClientID != "migrated-id" {
		t.Errorf("ClientID = %q, want migrated-id", cfg.ClientID)
	}

	// New XDG file must exist
	newPath := filepath.Join(tmpXDG, "acloud", "config.yaml")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("migrateLegacyConfig() did not create XDG config file")
	}
}

func TestGetConfigPath_XDGEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	got, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath(): %v", err)
	}
	want := filepath.Join(tmpDir, "acloud", "config.yaml")
	if got != want {
		t.Errorf("GetConfigPath() = %q, want %q", got, want)
	}
}
