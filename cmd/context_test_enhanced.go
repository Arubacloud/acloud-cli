package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadContext_InvalidYAML(t *testing.T) {
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
	contextPath := filepath.Join(tmpDir, ".acloud-context.yaml")
	invalidYAML := "this is not valid yaml: ["
	err := os.WriteFile(contextPath, []byte(invalidYAML), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	// Should return error
	ctx, err := LoadContext()
	if err == nil {
		t.Error("LoadContext() should return error for invalid YAML")
	}
	if ctx != nil {
		t.Error("LoadContext() should return nil context for invalid YAML")
	}
}

func TestSaveContext_EmptyContext(t *testing.T) {
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

	// Create empty context
	emptyContext := &Context{
		Contexts: make(map[string]CtxInfo),
	}

	// Save context
	err := SaveContext(emptyContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Load and verify
	loadedContext, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext() after SaveContext() error = %v", err)
	}

	if len(loadedContext.Contexts) != 0 {
		t.Errorf("LoadContext() Contexts length = %v, want 0", len(loadedContext.Contexts))
	}
}

func TestGetCurrentProjectID_NoCurrentContext(t *testing.T) {
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

	// Create context without current context set
	testContext := &Context{
		Contexts: map[string]CtxInfo{
			"test-context": {
				ProjectID: "test-project-id",
			},
		},
	}

	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Should return error
	projectID, err := GetCurrentProjectID()
	if err == nil {
		t.Error("GetCurrentProjectID() should return error when no current context is set")
	}
	if projectID != "" {
		t.Errorf("GetCurrentProjectID() = %v, want empty string", projectID)
	}
}

func TestGetCurrentProjectID_ContextNotFound(t *testing.T) {
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

	// Create context with current context pointing to non-existent context
	testContext := &Context{
		CurrentContext: "non-existent-context",
		Contexts: map[string]CtxInfo{
			"test-context": {
				ProjectID: "test-project-id",
			},
		},
	}

	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Should return error
	projectID, err := GetCurrentProjectID()
	if err == nil {
		t.Error("GetCurrentProjectID() should return error when current context not found")
	}
	if projectID != "" {
		t.Errorf("GetCurrentProjectID() = %v, want empty string", projectID)
	}
}

func TestSaveContext_MultipleContexts(t *testing.T) {
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

	// Create context with multiple contexts
	testContext := &Context{
		CurrentContext: "context1",
		Contexts: map[string]CtxInfo{
			"context1": {
				ProjectID: "project-1",
			},
			"context2": {
				ProjectID: "project-2",
			},
			"context3": {
				ProjectID: "project-3",
			},
		},
	}

	// Save context
	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Load and verify
	loadedContext, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext() after SaveContext() error = %v", err)
	}

	if len(loadedContext.Contexts) != 3 {
		t.Errorf("LoadContext() Contexts length = %v, want 3", len(loadedContext.Contexts))
	}

	if loadedContext.CurrentContext != "context1" {
		t.Errorf("LoadContext() CurrentContext = %v, want context1", loadedContext.CurrentContext)
	}

	if loadedContext.Contexts["context1"].ProjectID != "project-1" {
		t.Errorf("LoadContext() context1 ProjectID = %v, want project-1", loadedContext.Contexts["context1"].ProjectID)
	}
}
