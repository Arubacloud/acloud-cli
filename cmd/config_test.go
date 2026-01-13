package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
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
