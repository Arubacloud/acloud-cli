package cmd

import (
	"os"
	"path/filepath"
	"runtime"
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
