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
	Run: func(cmd *cobra.Command, args []string) {
		contextName := args[0]
		projectID, _ := cmd.Flags().GetString("project-id")

		if projectID == "" {
			fmt.Println("Error: --project-id is required")
			return
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
			fmt.Printf("Error saving context: %v\n", err)
			return
		}

		fmt.Printf("Context '%s' set with project ID: %s\n", contextName, projectID)
	},
}

var contextUseCmd = &cobra.Command{
	Use:   "use [context-name]",
	Short: "Switch to a different context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contextName := args[0]

		// Load context
		ctx, err := LoadContext()
		if err != nil {
			fmt.Printf("Error loading context: %v\n", err)
			return
		}

		// Check if context exists
		if _, exists := ctx.Contexts[contextName]; !exists {
			fmt.Printf("Context '%s' not found. Available contexts: ", contextName)
			for name := range ctx.Contexts {
				fmt.Printf("%s ", name)
			}
			fmt.Println()
			return
		}

		// Set current context
		ctx.CurrentContext = contextName

		// Save context
		if err := SaveContext(ctx); err != nil {
			fmt.Printf("Error saving context: %v\n", err)
			return
		}

		fmt.Printf("Switched to context '%s'\n", contextName)
		fmt.Printf("Project ID: %s\n", ctx.Contexts[contextName].ProjectID)
	},
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	Run: func(cmd *cobra.Command, args []string) {
		// Load context
		ctx, err := LoadContext()
		if err != nil {
			fmt.Println("No contexts found")
			return
		}

		if len(ctx.Contexts) == 0 {
			fmt.Println("No contexts found")
			return
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
	},
}

var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current context",
	Run: func(cmd *cobra.Command, args []string) {
		// Load context
		ctx, err := LoadContext()
		if err != nil || ctx.CurrentContext == "" {
			fmt.Println("No current context set")
			return
		}

		info, exists := ctx.Contexts[ctx.CurrentContext]
		if !exists {
			fmt.Println("Current context not found")
			return
		}

		fmt.Printf("Current context: %s\n", ctx.CurrentContext)
		fmt.Printf("Project ID:      %s\n", info.ProjectID)
	},
}

var contextDeleteCmd = &cobra.Command{
	Use:   "delete [context-name]",
	Short: "Delete a context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		contextName := args[0]

		// Load context
		ctx, err := LoadContext()
		if err != nil {
			fmt.Printf("Error loading context: %v\n", err)
			return
		}

		// Check if context exists
		if _, exists := ctx.Contexts[contextName]; !exists {
			fmt.Printf("Context '%s' not found\n", contextName)
			return
		}

		// Delete context
		delete(ctx.Contexts, contextName)

		// If this was the current context, clear it
		if ctx.CurrentContext == contextName {
			ctx.CurrentContext = ""
		}

		// Save context
		if err := SaveContext(ctx); err != nil {
			fmt.Printf("Error saving context: %v\n", err)
			return
		}

		fmt.Printf("Context '%s' deleted\n", contextName)
	},
}

// LoadContext loads the context configuration
func LoadContext() (*Context, error) {
	contextFile := getContextFilePath()
	data, err := os.ReadFile(contextFile)
	if err != nil {
		return nil, err
	}

	var ctx Context
	if err := yaml.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

// SaveContext saves the context configuration
func SaveContext(ctx *Context) error {
	contextFile := getContextFilePath()

	// Ensure directory exists
	dir := filepath.Dir(contextFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(ctx)
	if err != nil {
		return err
	}

	return os.WriteFile(contextFile, data, 0600)
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

// getContextFilePath returns the path to the context file
func getContextFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".acloud-context.yaml"
	}
	return filepath.Join(home, ".acloud-context.yaml")
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
