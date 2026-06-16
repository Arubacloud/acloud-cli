package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

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

	// Verify file exists at XDG path
	contextPath := filepath.Join(tmpDir, "acloud", "context.yaml")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		t.Fatal("SaveContext() did not create context file at XDG path")
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
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Write invalid YAML to the XDG path, creating the directory first.
	xdgPath := filepath.Join(tmpDir, "acloud", "context.yaml")
	if err := os.MkdirAll(filepath.Dir(xdgPath), 0700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(xdgPath, []byte("this is not valid yaml: ["), 0600); err != nil {
		t.Fatalf("write: %v", err)
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

// ─── Layer-1: Validate ────────────────────────────────────────────────────────

func TestContextSetArgs_Validate(t *testing.T) {
	cases := []struct {
		name    string
		args    ContextSetArgs
		wantErr bool
	}{
		{"project-id set", ContextSetArgs{Name: "ctx", ProjectID: "pid"}, false},
		{"project-id empty", ContextSetArgs{Name: "ctx", ProjectID: ""}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if tc.wantErr && err == nil {
				t.Error("Validate() = nil, want error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestNewContextSetArgsFromCobraCommand_ValidationError(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	// cobra.ExactArgs(1) is already enforced by the command, but we test the
	// constructor directly — pass a valid positional arg and omit --project-id.
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "", "")
	// Do not set project-id, so it remains empty.
	_, err := NewContextSetArgsFromCobraCommand(cmd, []string{"myctx"})
	if err == nil {
		t.Fatal("expected ErrValidationFailed, got nil")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed, got: %v", err)
	}
}

func TestContextUseArgs_Validate(t *testing.T) {
	// Validate is a no-op for ContextUseArgs; verify it never returns an error.
	if err := (&ContextUseArgs{Name: "any"}).Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

func TestContextDeleteArgs_Validate(t *testing.T) {
	// Validate is a no-op for ContextDeleteArgs; verify it never returns an error.
	if err := (&ContextDeleteArgs{Name: "any"}).Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

// ─── Layer-2: Operation ───────────────────────────────────────────────────────

func TestContextSet_Operation(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	out := captureStdout(func() {
		err := ContextSet(context.Background(), ContextSetArgs{Name: "op-ctx", ProjectID: "op-pid"})
		if err != nil {
			t.Errorf("ContextSet() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "op-ctx") {
		t.Errorf("expected context name in output, got: %s", out)
	}
	if !strings.Contains(out, "op-pid") {
		t.Errorf("expected project ID in output, got: %s", out)
	}

	// Verify the context was persisted.
	loaded, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if info, ok := loaded.Contexts["op-ctx"]; !ok || info.ProjectID != "op-pid" {
		t.Errorf("context not persisted correctly: %+v", loaded.Contexts)
	}
}

func TestContextUse_Operation_NotFound(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	if err := SaveContext(&Context{
		Contexts: map[string]CtxInfo{"existing": {ProjectID: "p1"}},
	}); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	err := ContextUse(context.Background(), ContextUseArgs{Name: "missing"})
	if err == nil {
		t.Fatal("expected error for missing context")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestContextUse_Operation_SwitchesContext(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	if err := SaveContext(&Context{
		CurrentContext: "first",
		Contexts: map[string]CtxInfo{
			"first":  {ProjectID: "p1"},
			"second": {ProjectID: "p2"},
		},
	}); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	out := captureStdout(func() {
		err := ContextUse(context.Background(), ContextUseArgs{Name: "second"})
		if err != nil {
			t.Errorf("ContextUse() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "second") {
		t.Errorf("expected 'second' in output, got: %s", out)
	}

	loaded, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if loaded.CurrentContext != "second" {
		t.Errorf("CurrentContext = %q, want second", loaded.CurrentContext)
	}
}

func TestContextList_Operation_Empty(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	out := captureStdout(func() {
		err := ContextList(context.Background(), ContextListArgs{})
		if err != nil {
			t.Errorf("ContextList() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No contexts found") {
		t.Errorf("expected 'No contexts found', got: %s", out)
	}
}

func TestContextList_Operation_WithContexts(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	if err := SaveContext(&Context{
		CurrentContext: "alpha",
		Contexts: map[string]CtxInfo{
			"alpha": {ProjectID: "pa"},
			"beta":  {ProjectID: "pb"},
		},
	}); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	out := captureStdout(func() {
		err := ContextList(context.Background(), ContextListArgs{})
		if err != nil {
			t.Errorf("ContextList() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "alpha") {
		t.Errorf("expected 'alpha' in output, got: %s", out)
	}
	if !strings.Contains(out, "beta") {
		t.Errorf("expected 'beta' in output, got: %s", out)
	}
	// Active context should be marked with *.
	if !strings.Contains(out, "*") {
		t.Errorf("expected '*' marker for current context, got: %s", out)
	}
}

func TestContextCurrent_Operation_NoContext(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	out := captureStdout(func() {
		err := ContextCurrent(context.Background(), ContextCurrentArgs{})
		if err != nil {
			t.Errorf("ContextCurrent() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No current context") {
		t.Errorf("expected 'No current context', got: %s", out)
	}
}

func TestContextCurrent_Operation_WithContext(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	if err := SaveContext(&Context{
		CurrentContext: "myctx",
		Contexts:       map[string]CtxInfo{"myctx": {ProjectID: "mypid"}},
	}); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	out := captureStdout(func() {
		err := ContextCurrent(context.Background(), ContextCurrentArgs{})
		if err != nil {
			t.Errorf("ContextCurrent() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "myctx") {
		t.Errorf("expected context name, got: %s", out)
	}
	if !strings.Contains(out, "mypid") {
		t.Errorf("expected project ID, got: %s", out)
	}
}

func TestContextDelete_Operation_Success(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	if err := SaveContext(&Context{
		CurrentContext: "del-ctx",
		Contexts:       map[string]CtxInfo{"del-ctx": {ProjectID: "dp"}},
	}); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	out := captureStdout(func() {
		err := ContextDelete(context.Background(), ContextDeleteArgs{Name: "del-ctx"})
		if err != nil {
			t.Errorf("ContextDelete() unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "del-ctx") {
		t.Errorf("expected deleted context name in output, got: %s", out)
	}

	loaded, err := LoadContext()
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if _, exists := loaded.Contexts["del-ctx"]; exists {
		t.Error("context was not removed from file")
	}
	// Deleting the active context must clear CurrentContext.
	if loaded.CurrentContext != "" {
		t.Errorf("CurrentContext = %q, want empty after deleting active context", loaded.CurrentContext)
	}
}

func TestContextDelete_Operation_NotFound(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	if err := SaveContext(&Context{
		Contexts: map[string]CtxInfo{"other": {ProjectID: "p"}},
	}); err != nil {
		t.Fatalf("SaveContext: %v", err)
	}

	err := ContextDelete(context.Background(), ContextDeleteArgs{Name: "missing"})
	if err == nil {
		t.Fatal("expected error for missing context")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error, got: %v", err)
	}
}
