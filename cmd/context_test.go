package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadContext(t *testing.T) {
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

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
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

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
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("USERPROFILE", originalUserProfile)
	}()

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

func TestLoadContext_InvalidYAML(t *testing.T) {
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

	contextPath := filepath.Join(tmpDir, ".acloud-context.yaml")
	invalidYAML := "this is not valid yaml: ["
	err := os.WriteFile(contextPath, []byte(invalidYAML), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML: %v", err)
	}

	ctx, err := LoadContext()
	if err == nil {
		t.Error("LoadContext() should return error for invalid YAML")
	}
	if ctx != nil {
		t.Error("LoadContext() should return nil context for invalid YAML")
	}
}

func TestSaveContext_EmptyContext(t *testing.T) {
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	emptyContext := &Context{
		Contexts: make(map[string]CtxInfo),
	}

	err := SaveContext(emptyContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	loadedContext, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext() after SaveContext() error = %v", err)
	}

	if len(loadedContext.Contexts) != 0 {
		t.Errorf("LoadContext() Contexts length = %v, want 0", len(loadedContext.Contexts))
	}
}

func TestGetCurrentProjectID_NoCurrentContext(t *testing.T) {
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

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

	projectID, err := GetCurrentProjectID()
	if err == nil {
		t.Error("GetCurrentProjectID() should return error when no current context is set")
	}
	if projectID != "" {
		t.Errorf("GetCurrentProjectID() = %v, want empty string", projectID)
	}
}

func TestGetCurrentProjectID_ContextNotFound(t *testing.T) {
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

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

	projectID, err := GetCurrentProjectID()
	if err == nil {
		t.Error("GetCurrentProjectID() should return error when current context not found")
	}
	if projectID != "" {
		t.Errorf("GetCurrentProjectID() = %v, want empty string", projectID)
	}
}

func TestSaveContext_MultipleContexts(t *testing.T) {
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()

	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	testContext := &Context{
		CurrentContext: "context1",
		Contexts: map[string]CtxInfo{
			"context1": {ProjectID: "project-1"},
			"context2": {ProjectID: "project-2"},
			"context3": {ProjectID: "project-3"},
		},
	}

	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

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
}
