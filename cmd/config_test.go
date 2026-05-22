package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

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
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

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

	// Verify file exists
	configPath := filepath.Join(tmpDir, ".acloud.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("SaveConfig() did not create config file")
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
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

	// Get config path
	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".acloud.yaml")
	if configPath != expectedPath {
		t.Errorf("GetConfigPath() = %v, want %v", configPath, expectedPath)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("os.UserHomeDir() on Windows uses Win32 API, not HOME env var")
	}
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configPath := filepath.Join(tmpDir, ".acloud.yaml")
	invalidYAML := "this is not valid yaml: ["
	err := os.WriteFile(configPath, []byte(invalidYAML), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
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
	if runtime.GOOS == "windows" {
		t.Skip("os.UserHomeDir() on Windows uses Win32 API, not HOME env var")
	}
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	configPath, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".acloud.yaml")
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
	os.Unsetenv("ACLOUD_CLIENT_SECRET")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("ACLOUD_CLIENT_ID", origID)
		os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)
	}()

	out, err := runCmdCapture(nil, []string{"config", "set", "--client-id", "test-id", "--client-secret", "test-secret"})
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
	os.Unsetenv("ACLOUD_CLIENT_ID")
	os.Unsetenv("ACLOUD_CLIENT_SECRET")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("ACLOUD_CLIENT_ID", origID)
		os.Setenv("ACLOUD_CLIENT_SECRET", origSecret)
	}()

	_, err := runCmdCapture(nil, []string{"config", "set", "--client-secret", "test-secret"})
	if err == nil {
		t.Fatal("expected error for missing --client-id")
	}
	if !strings.Contains(err.Error(), "client-id") {
		t.Errorf("expected client-id mention in error, got: %s", err.Error())
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
	if !strings.Contains(out, "********") {
		t.Errorf("expected redacted secret in output, got: %s", out)
	}
}

func TestConfigShowCmd_NoConfig(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
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
