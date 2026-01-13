package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_InvalidYAML(t *testing.T) {
	// Save original home dir
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create invalid YAML file
	configPath := filepath.Join(tmpDir, ".acloud.yaml")
	invalidYAML := "this is not valid yaml: ["
	err := os.WriteFile(configPath, []byte(invalidYAML), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	// Should return error
	config, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error for invalid YAML")
	}
	if config != nil {
		t.Error("LoadConfig() should return nil config for invalid YAML")
	}
}

func TestSaveConfig_EmptyConfig(t *testing.T) {
	// Save original home dir
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create empty config
	emptyConfig := &Config{}

	// Save config
	err := SaveConfig(emptyConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Load and verify
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
	// Save original home dir
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create config with only ClientID
	partialConfig := &Config{
		ClientID: "test-client-id-only",
	}

	// Save config
	err := SaveConfig(partialConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Load and verify
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
	// Save original home dir
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

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
