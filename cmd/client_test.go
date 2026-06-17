package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSkipClientInit(t *testing.T) {
	cases := []struct {
		name string
		path []string // command path from root to leaf (root first)
		want bool
	}{
		// Commands that MUST skip client init
		{"config root", []string{"acloud", "config"}, true},
		{"config set", []string{"acloud", "config", "set"}, true},
		{"config show", []string{"acloud", "config", "show"}, true},
		{"config profile list", []string{"acloud", "config", "profile", "list"}, true},
		{"context root", []string{"acloud", "context"}, true},
		{"context use", []string{"acloud", "context", "use"}, true},
		{"context list", []string{"acloud", "context", "list"}, true},
		{"completion", []string{"acloud", "completion"}, true},
		{"__complete", []string{"acloud", "__complete"}, true},
		{"__completeNoDesc", []string{"acloud", "__completeNoDesc"}, true},
		{"help", []string{"acloud", "help"}, true},
		// Resource commands that MUST NOT skip
		{"network vpc list", []string{"acloud", "network", "vpc", "list"}, false},
		{"compute cloudserver get", []string{"acloud", "compute", "cloudserver", "get"}, false},
		{"storage blockstorage list", []string{"acloud", "storage", "blockstorage", "list"}, false},
		{"security kms create", []string{"acloud", "security", "kms", "create"}, false},
		{"management project list", []string{"acloud", "management", "project", "list"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Build a chain of commands matching the path.
			var leaf *cobra.Command
			for i, name := range tc.path {
				cmd := &cobra.Command{Use: name}
				if i > 0 {
					leaf.AddCommand(cmd)
				}
				leaf = cmd
			}
			if got := skipClientInit(leaf); got != tc.want {
				t.Errorf("skipClientInit(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestPersistentPreRunE_SkipsForConfig(t *testing.T) {
	// config show must work without a valid credentials file.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Cleanup(resetClientState)

	// Running config show without credentials should not return a credential error.
	err := runCmd(nil, []string{"config", "show"})
	if err != nil {
		t.Errorf("config show should not require credentials, got: %v", err)
	}
}

func TestPersistentPreRunE_SkipsForContext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("USERPROFILE", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Cleanup(resetClientState)

	err := runCmd(nil, []string{"context", "list"})
	if err != nil {
		t.Errorf("context list should not require credentials, got: %v", err)
	}
}
