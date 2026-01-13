package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestGetProjectID_FromFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "", "Project ID")

	// Set project ID via flag
	cmd.Flags().Set("project-id", "test-project-123")

	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Fatalf("GetProjectID() error = %v", err)
	}

	if projectID != "test-project-123" {
		t.Errorf("GetProjectID() = %v, want test-project-123", projectID)
	}
}

func TestGetProjectID_FromContext(t *testing.T) {
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

	// Create context with current context set
	testContext := &Context{
		CurrentContext: "test-context",
		Contexts: map[string]CtxInfo{
			"test-context": {
				ProjectID: "context-project-456",
			},
		},
	}

	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Create command without project-id flag set
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "", "Project ID")

	// Should get from context
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Fatalf("GetProjectID() error = %v", err)
	}

	if projectID != "context-project-456" {
		t.Errorf("GetProjectID() = %v, want context-project-456", projectID)
	}
}

func TestGetProjectID_FlagTakesPrecedence(t *testing.T) {
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

	// Create context with current context set
	testContext := &Context{
		CurrentContext: "test-context",
		Contexts: map[string]CtxInfo{
			"test-context": {
				ProjectID: "context-project-456",
			},
		},
	}

	err := SaveContext(testContext)
	if err != nil {
		t.Fatalf("SaveContext() error = %v", err)
	}

	// Create command with both flag and context available
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "", "Project ID")
	cmd.Flags().Set("project-id", "flag-project-789")

	// Flag should take precedence
	projectID, err := GetProjectID(cmd)
	if err != nil {
		t.Fatalf("GetProjectID() error = %v", err)
	}

	if projectID != "flag-project-789" {
		t.Errorf("GetProjectID() = %v, want flag-project-789", projectID)
	}
}

func TestGetProjectID_NoFlagNoContext(t *testing.T) {
	// Save original home dir
	originalHome := os.Getenv("HOME")
	originalUserProfile := os.Getenv("USERPROFILE")
	if originalHome == "" {
		originalHome = originalUserProfile // Windows
	}

	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Set both HOME and USERPROFILE to temp directory (for cross-platform compatibility)
	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		if originalUserProfile != "" {
			os.Setenv("USERPROFILE", originalUserProfile)
		}
		// Clear client cache to avoid interference
		clientCacheLock.Lock()
		clientCache = nil
		cachedClientID = ""
		cachedSecret = ""
		cachedDebug = false
		clientCacheLock.Unlock()
	}()

	// Ensure no context file exists in temp dir
	contextPath := filepath.Join(tmpDir, ".acloud-context.yaml")
	_ = os.Remove(contextPath) // Ignore error if file doesn't exist

	// Also ensure no context file exists in actual home (in case of test pollution)
	actualHome, _ := os.UserHomeDir()
	if actualHome != "" && actualHome != tmpDir {
		actualContextPath := filepath.Join(actualHome, ".acloud-context.yaml")
		_ = os.Remove(actualContextPath) // Ignore error
	}

	// Create command without project-id flag set
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "", "Project ID")

	// Should return error
	projectID, err := GetProjectID(cmd)
	if err == nil {
		t.Error("GetProjectID() should return error when no flag and no context")
	}
	if projectID != "" {
		t.Errorf("GetProjectID() = %v, want empty string", projectID)
	}
}
