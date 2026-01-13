package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadContext(t *testing.T) {
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

	// Test with non-existent context
	ctx, err := LoadContext()
	if err == nil {
		t.Error("LoadContext() should return error when context file doesn't exist")
	}
	if ctx != nil {
		t.Error("LoadContext() should return nil context when file doesn't exist")
	}
}

func TestSaveContext(t *testing.T) {
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

	// Create test context
	testContext := &Context{
		CurrentContext: "test-context",
		Contexts: map[string]CtxInfo{
			"test-context": {
				ProjectID: "test-project-id",
			},
		},
	}

	// Save context
	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Verify file exists
	contextPath := filepath.Join(tmpDir, ".acloud-context.yaml")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Fatal("SaveContext() did not create context file")
	}

	// Load and verify
	loadedContext, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext() after SaveContext() error = %v", err)
	}

	if loadedContext.CurrentContext != testContext.CurrentContext {
		t.Errorf("LoadContext() CurrentContext = %v, want %v", loadedContext.CurrentContext, testContext.CurrentContext)
	}

	if loadedContext.Contexts["test-context"].ProjectID != testContext.Contexts["test-context"].ProjectID {
		t.Errorf("LoadContext() ProjectID = %v, want %v", loadedContext.Contexts["test-context"].ProjectID, testContext.Contexts["test-context"].ProjectID)
	}
}

func TestGetCurrentProjectID(t *testing.T) {
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

	// Test with no context
	projectID, err := GetCurrentProjectID()
	if err == nil {
		t.Error("GetCurrentProjectID() should return error when no context is set")
	}
	if projectID != "" {
		t.Errorf("GetCurrentProjectID() = %v, want empty string", projectID)
	}

	// Create context with current context set
	testContext := &Context{
		CurrentContext: "test-context",
		Contexts: map[string]CtxInfo{
			"test-context": {
				ProjectID: "test-project-id",
			},
		},
	}

	err = SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Now should return project ID
	projectID, err = GetCurrentProjectID()
	if err != nil {
		t.Fatalf("GetCurrentProjectID() error = %v", err)
	}

	if projectID != "test-project-id" {
		t.Errorf("GetCurrentProjectID() = %v, want test-project-id", projectID)
	}
}
