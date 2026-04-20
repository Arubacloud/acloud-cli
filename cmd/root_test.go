package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// setupMockConfig creates a temporary config file for testing
func setupMockConfig(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".acloud.yaml")

	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}

	// Create mock config
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("Failed to create mock config: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}

	return configPath, cleanup
}

func TestGetArubaClient(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	// Clear cache before test
	resetClientState()

	client, err := GetArubaClient()
	if err != nil {
		t.Fatalf("GetArubaClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("GetArubaClient() returned nil client")
	}
}

func TestGetArubaClient_Caching(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	// Clear cache
	resetClientState()

	client1, err1 := GetArubaClient()
	if err1 != nil {
		t.Fatalf("GetArubaClient() error = %v", err1)
	}

	// Second call should use cached client
	client2, err2 := GetArubaClient()
	if err2 != nil {
		t.Fatalf("GetArubaClient() second call error = %v", err2)
	}

	// Should be the same instance (cached)
	if client1 != client2 {
		t.Error("GetArubaClient() should return cached client on second call")
	}
}

func TestGetArubaClient_NoConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	// Create temporary directory without config
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}()

	client, err := GetArubaClient()
	if err == nil {
		t.Error("GetArubaClient() should return error when config doesn't exist")
	}
	if client != nil {
		t.Error("GetArubaClient() should return nil client when config doesn't exist")
	}

	// Verify error message
	errMsg := err.Error()
	if !contains(errMsg, "failed to load configuration") {
		t.Errorf("Expected error about failed to load configuration, got: %v", err)
	}
}

func TestGetArubaClient_EmptyCredentials(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}()

	// Create config with empty credentials
	config := &Config{
		ClientID:     "",
		ClientSecret: "",
	}
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	client, err := GetArubaClient()
	if err == nil {
		t.Error("GetArubaClient() should return error when credentials are empty")
	}
	if client != nil {
		t.Error("GetArubaClient() should return nil client when credentials are empty")
	}

	// Verify error message
	errMsg := err.Error()
	if !contains(errMsg, "client ID or client secret not configured") {
		t.Errorf("Expected error about credentials not configured, got: %v", err)
	}
}

func TestPrintTable(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 10},
		{Header: "ID", Width: 10},
		{Header: "STATUS", Width: 10},
	}

	rows := [][]string{
		{"test-name", "test-id-123", "Active"},
		{"another-name", "another-id-456", "Inactive"},
	}

	// Should not panic
	PrintTable(headers, rows)
}

func TestPrintTable_EmptyRows(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 10},
		{Header: "ID", Width: 10},
	}

	rows := [][]string{}

	// Should not panic with empty rows
	PrintTable(headers, rows)
}

func TestPrintTable_LongValues(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 5},
	}

	rows := [][]string{
		{"very-long-name-that-exceeds-width"},
	}

	// Should truncate long values
	PrintTable(headers, rows)
}

func TestPrintTable_MismatchedColumns(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 10},
		{Header: "ID", Width: 10},
	}

	rows := [][]string{
		{"test-name"}, // Missing second column
		{"another-name", "id-123", "extra-column"}, // Extra column
	}

	// Should not panic with mismatched columns
	PrintTable(headers, rows)
}

func TestNormalizeHeaderKey(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"NAME", "name"},
		{"ID", "id"},
		{"PUBLIC_KEY", "public_key"},
		{"CREATION DATE", "creation_date"},
		{"RAM(GB)", "ram_gb"},
		{"HD(GB)", "hd_gb"},
		{"CPU", "cpu"},
		{"  spaced  ", "spaced"},
		{"multi  space__key", "multi_space_key"},
		{"abc123", "abc123"},
	}
	for _, c := range cases {
		got := normalizeHeaderKey(c.in)
		if got != c.want {
			t.Errorf("normalizeHeaderKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPrintTable_JSONFormat(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 20},
		{Header: "ID", Width: 30},
		{Header: "RAM(GB)", Width: 10},
	}
	rows := [][]string{
		{"server-01", "abc123", "4"},
		{"server-02", "def456", "8"},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.PersistentFlags().Set("output", "json")
	PrintTable(headers, rows)
	rootCmd.PersistentFlags().Set("output", "table")

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
	if result[0]["name"] != "server-01" {
		t.Errorf("expected name=server-01, got %q", result[0]["name"])
	}
	if result[0]["ram_gb"] != "4" {
		t.Errorf("expected ram_gb=4, got %q", result[0]["ram_gb"])
	}
	if result[1]["id"] != "def456" {
		t.Errorf("expected id=def456, got %q", result[1]["id"])
	}
}

func TestPrintTable_JSONFormat_Empty(t *testing.T) {
	headers := []TableColumn{{Header: "NAME", Width: 10}}
	rows := [][]string{}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.PersistentFlags().Set("output", "json")
	PrintTable(headers, rows)
	rootCmd.PersistentFlags().Set("output", "table")

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	var result []map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("empty output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestPrintTable_YAMLFormat(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 20},
		{Header: "ID", Width: 30},
		{Header: "CREATION DATE", Width: 15},
	}
	rows := [][]string{
		{"project-a", "id-001", "26-03-2026"},
		{"project-b", "id-002", "27-03-2026"},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.PersistentFlags().Set("output", "yaml")
	PrintTable(headers, rows)
	rootCmd.PersistentFlags().Set("output", "table")

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	out := buf.String()
	if !strings.Contains(out, "name: project-a") {
		t.Errorf("YAML output missing 'name: project-a':\n%s", out)
	}
	if !strings.Contains(out, "id: id-001") {
		t.Errorf("YAML output missing 'id: id-001':\n%s", out)
	}
	if !strings.Contains(out, "creation_date: 26-03-2026") {
		t.Errorf("YAML output missing 'creation_date: 26-03-2026':\n%s", out)
	}
	if !strings.Contains(out, "- ") {
		t.Errorf("YAML output missing list indicators:\n%s", out)
	}
}

func TestGetArubaClient_DebugFlagChange(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	// Clear cache
	resetClientState()

	// First call without debug
	rootCmd.PersistentFlags().Set("debug", "false")
	client1, err1 := GetArubaClient()
	if err1 != nil {
		t.Fatalf("GetArubaClient() error = %v", err1)
	}

	// Change debug flag
	rootCmd.PersistentFlags().Set("debug", "true")

	// Second call should create new client (cache invalidated)
	client2, err2 := GetArubaClient()
	if err2 != nil {
		t.Fatalf("GetArubaClient() second call error = %v", err2)
	}

	// Should be different instances (cache invalidated due to debug flag change)
	if client1 == client2 {
		t.Error("GetArubaClient() should return new client when debug flag changes")
	}
}

func TestGetArubaClient_CredentialChange(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}()

	// Create initial config
	config1 := &Config{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}
	err := SaveConfig(config1)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Clear cache
	resetClientState()

	// First call
	client1, err1 := GetArubaClient()
	if err1 != nil {
		t.Fatalf("GetArubaClient() error = %v", err1)
	}

	// Change credentials
	config2 := &Config{
		ClientID:     "client-2",
		ClientSecret: "secret-2",
	}
	err = SaveConfig(config2)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Second call should create new client (cache invalidated due to credential change)
	client2, err2 := GetArubaClient()
	if err2 != nil {
		t.Fatalf("GetArubaClient() second call error = %v", err2)
	}

	// Should be different instances (cache invalidated due to credential change)
	if client1 == client2 {
		t.Error("GetArubaClient() should return new client when credentials change")
	}
}
