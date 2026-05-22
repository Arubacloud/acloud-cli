package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func withTempHomeDir(t *testing.T) (cleanup func()) {
	t.Helper()
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	return func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
	}
}

func TestContextSetCmd(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"context", "set", "myctx", "--project-id", "proj-abc"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myctx") {
					t.Errorf("expected context name in output, got: %s", out)
				}
				if !strings.Contains(out, "proj-abc") {
					t.Errorf("expected project ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing --project-id",
			args:        []string{"context", "set", "myctx"},
			wantErr:     true,
			errContains: "project-id",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCmdCapture(nil, tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestContextUseCmd(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	// Set up a context first
	if _, err := runCmdCapture(nil, []string{"context", "set", "myctx", "--project-id", "proj-abc"}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"context", "use", "myctx"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myctx") {
					t.Errorf("expected context name in output, got: %s", out)
				}
			},
		},
		{
			name:        "non-existent context",
			args:        []string{"context", "use", "noexist"},
			wantErr:     true,
			errContains: "not found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCmdCapture(nil, tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestContextListCmd(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	// Test list with no contexts
	out, err := runCmdCapture(nil, []string{"context", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No contexts found") {
		t.Errorf("expected 'No contexts found', got: %s", out)
	}

	// Add a context and list again
	if _, err := runCmdCapture(nil, []string{"context", "set", "myctx", "--project-id", "proj-abc"}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	out, err = runCmdCapture(nil, []string{"context", "list"})
	if err != nil {
		t.Fatalf("unexpected error after set: %v", err)
	}
	if !strings.Contains(out, "myctx") {
		t.Errorf("expected context in list, got: %s", out)
	}
}

func TestContextDeleteCmd(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	// Set up a context first
	if _, err := runCmdCapture(nil, []string{"context", "set", "myctx", "--project-id", "proj-abc"}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"context", "delete", "myctx"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myctx") {
					t.Errorf("expected context name in output, got: %s", out)
				}
			},
		},
		{
			name:        "non-existent context",
			args:        []string{"context", "delete", "noexist"},
			wantErr:     true,
			errContains: "not found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCmdCapture(nil, tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestContextCurrentCmd(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	// No current context
	out, err := runCmdCapture(nil, []string{"context", "current"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "No current context") {
		t.Errorf("expected 'No current context', got: %s", out)
	}

	// Set and use a context
	if _, err := runCmdCapture(nil, []string{"context", "set", "myctx", "--project-id", "proj-abc"}); err != nil {
		t.Fatalf("setup set failed: %v", err)
	}
	if _, err := runCmdCapture(nil, []string{"context", "use", "myctx"}); err != nil {
		t.Fatalf("setup use failed: %v", err)
	}

	out, err = runCmdCapture(nil, []string{"context", "current"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "myctx") {
		t.Errorf("expected current context name, got: %s", out)
	}
	if !strings.Contains(out, "proj-abc") {
		t.Errorf("expected project ID, got: %s", out)
	}
}
