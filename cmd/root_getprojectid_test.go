package cmd

import (
	"os"
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
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "") // clear so HOME/.config is used, not CI's XDG dir
	t.Cleanup(resetClientState)

	// Create command without project-id flag set
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "", "Project ID")

	// Should return error: no --project-id and no context file in empty tmpDir
	projectID, err := GetProjectID(cmd)
	if err == nil {
		t.Error("GetProjectID() should return error when no flag and no context")
	}
	if projectID != "" {
		t.Errorf("GetProjectID() = %v, want empty string", projectID)
	}
}
