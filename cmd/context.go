package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Context represents the CLI context configuration
type Context struct {
	CurrentContext string             `yaml:"current-context"`
	Contexts       map[string]CtxInfo `yaml:"contexts"`
}

// CtxInfo represents a single context
type CtxInfo struct {
	ProjectID string `yaml:"project-id"`
	Name      string `yaml:"name,omitempty"`
}

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage CLI contexts",
	Long:  `Manage CLI contexts to avoid passing --project-id repeatedly.`,
}

var contextSetCmd = &cobra.Command{
	Use:   "set [context-name]",
	Short: "Set a context with a project ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextSet,
}

var contextUseCmd = &cobra.Command{
	Use:   "use [context-name]",
	Short: "Switch to a different context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextUse,
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	Args:  cobra.NoArgs,
	RunE:  runContextList,
}

var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current context",
	RunE:  runContextCurrent,
}

var contextDeleteCmd = &cobra.Command{
	Use:   "delete [context-name]",
	Short: "Delete a context",
	Args:  cobra.ExactArgs(1),
	RunE:  runContextDelete,
}

// LoadContext loads the context configuration
func LoadContext() (*Context, error) {
	contextFile, err := getContextFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(contextFile)
	if err != nil {
		return nil, err
	}

	var ctx Context
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("context file %s is corrupted (%w). Delete it and run 'acloud context set' to reconfigure", contextFile, err)
	}

	return &ctx, nil
}

// SaveContext saves the context configuration
func SaveContext(ctx *Context) error {
	contextFile, err := getContextFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(contextFile)
	if err := os.MkdirAll(dir, FilePermDirAll); err != nil {
		return err
	}

	data, err := yaml.Marshal(ctx)
	if err != nil {
		return err
	}

	return os.WriteFile(contextFile, data, FilePermConfig)
}

// GetCurrentProjectID returns the project ID from the current context
func GetCurrentProjectID() (string, error) {
	ctx, err := LoadContext()
	if err != nil {
		return "", err
	}

	if ctx.CurrentContext == "" {
		return "", fmt.Errorf("no current context set")
	}

	info, exists := ctx.Contexts[ctx.CurrentContext]
	if !exists {
		return "", fmt.Errorf("current context '%s' not found", ctx.CurrentContext)
	}

	return info.ProjectID, nil
}

// getContextFilePath returns the path to the context file (TD-007).
func getContextFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".acloud-context.yaml"), nil
}

func init() {
	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextSetCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	contextCmd.AddCommand(contextDeleteCmd)

	// Flags
	contextSetCmd.Flags().String("project-id", "", "Project ID (required)")
	contextSetCmd.MarkFlagRequired("project-id")
}

func runContextSet(cmd *cobra.Command, args []string) error {
	contextName := args[0]
	projectID, err := cmd.Flags().GetString("project-id")
	if err != nil {
		return err
	}

	if projectID == "" {
		return fmt.Errorf("--project-id is required")
	}

	// Load existing context or create new
	ctx, err := LoadContext()
	if err != nil {
		// Create new context
		ctx = &Context{
			Contexts: make(map[string]CtxInfo),
		}
	}

	// Add or update context
	ctx.Contexts[contextName] = CtxInfo{
		ProjectID: projectID,
	}

	// Save context
	if err := SaveContext(ctx); err != nil {
		return fmt.Errorf("saving context: %w", err)
	}

	fmt.Printf("Context '%s' set with project ID: %s\n", contextName, projectID)
	return nil
}

func runContextUse(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	// Load context
	ctx, err := LoadContext()
	if err != nil {
		return fmt.Errorf("loading context: %w", err)
	}

	// Check if context exists
	if _, exists := ctx.Contexts[contextName]; !exists {
		available := make([]string, 0, len(ctx.Contexts))
		for name := range ctx.Contexts {
			available = append(available, name)
		}
		return fmt.Errorf("context '%s' not found; available contexts: %v", contextName, available)
	}

	// Set current context
	ctx.CurrentContext = contextName

	// Save context
	if err := SaveContext(ctx); err != nil {
		return fmt.Errorf("saving context: %w", err)
	}

	fmt.Printf("Switched to context '%s'\n", contextName)
	fmt.Printf("Project ID: %s\n", ctx.Contexts[contextName].ProjectID)
	return nil
}

func runContextList(cmd *cobra.Command, args []string) error {
	// Load context
	ctx, err := LoadContext()
	if err != nil {
		fmt.Println("No contexts found")
		return nil
	}

	if len(ctx.Contexts) == 0 {
		fmt.Println("No contexts found")
		return nil
	}

	fmt.Println("\nContexts:")
	fmt.Println("=========")
	for name, info := range ctx.Contexts {
		current := ""
		if name == ctx.CurrentContext {
			current = " *"
		}
		fmt.Printf("%-20s Project ID: %s%s\n", name, info.ProjectID, current)
	}
	if ctx.CurrentContext != "" {
		fmt.Printf("\n* = current context\n")
	}
	return nil
}

func runContextCurrent(cmd *cobra.Command, args []string) error {
	// Load context
	ctx, err := LoadContext()
	if err != nil || ctx.CurrentContext == "" {
		fmt.Println("No current context set")
		return nil
	}

	info, exists := ctx.Contexts[ctx.CurrentContext]
	if !exists {
		fmt.Println("Current context not found")
		return nil
	}

	fmt.Printf("Current context: %s\n", ctx.CurrentContext)
	fmt.Printf("Project ID:      %s\n", info.ProjectID)
	return nil
}

func runContextDelete(cmd *cobra.Command, args []string) error {
	contextName := args[0]

	// Load context
	ctx, err := LoadContext()
	if err != nil {
		return fmt.Errorf("loading context: %w", err)
	}

	// Check if context exists
	if _, exists := ctx.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	// Delete context
	delete(ctx.Contexts, contextName)

	// If this was the current context, clear it
	if ctx.CurrentContext == contextName {
		ctx.CurrentContext = ""
	}

	// Save context
	if err := SaveContext(ctx); err != nil {
		return fmt.Errorf("saving context: %w", err)
	}

	fmt.Printf("Context '%s' deleted\n", contextName)
	return nil
}
