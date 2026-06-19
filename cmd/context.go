package cmd

import (
	"context"
	"errors"
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
	RunE:  ContextSetRun,
}

var contextUseCmd = &cobra.Command{
	Use:   "use [context-name]",
	Short: "Switch to a different context",
	Args:  cobra.ExactArgs(1),
	RunE:  ContextUseRun,
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	Args:  cobra.NoArgs,
	RunE:  ContextListRun,
}

var contextCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current context",
	RunE:  ContextCurrentRun,
}

var contextDeleteCmd = &cobra.Command{
	Use:   "delete [context-name]",
	Short: "Delete a context",
	Args:  cobra.ExactArgs(1),
	RunE:  ContextDeleteRun,
}

// LoadContext loads the context configuration.
// Migrates legacy ~/.acloud-context.yaml to the XDG path on first load (#176).
func LoadContext() (*Context, error) {
	migrateLegacyContext()
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

// ClearAllContexts removes all saved contexts and unsets the current context.
func ClearAllContexts() error {
	return SaveContext(&Context{Contexts: make(map[string]CtxInfo)})
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

// GetCurrentProjectID returns the project ID from the current context.
// When no current context is set but exactly one context exists, that context
// is used automatically — so `acloud context set` is enough without a follow-up
// `acloud context use`.
func GetCurrentProjectID() (string, error) {
	ctx, err := LoadContext()
	if err != nil {
		return "", err
	}

	if ctx.CurrentContext == "" {
		if len(ctx.Contexts) == 1 {
			for _, info := range ctx.Contexts {
				return info.ProjectID, nil
			}
		}
		return "", fmt.Errorf("no current context set")
	}

	info, exists := ctx.Contexts[ctx.CurrentContext]
	if !exists {
		return "", fmt.Errorf("current context '%s' not found", ctx.CurrentContext)
	}

	return info.ProjectID, nil
}

// getContextFilePath returns the XDG-compliant path to the context file (#176, TD-007).
// New path: $XDG_CONFIG_HOME/acloud/context.yaml (default: ~/.config/acloud/context.yaml).
func getContextFilePath() (string, error) {
	dir, err := xdgConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine XDG config directory: %w", err)
	}
	return filepath.Join(dir, "acloud", "context.yaml"), nil
}

// legacyContextPath returns the pre-XDG context path (~/.acloud-context.yaml).
func legacyContextPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".acloud-context.yaml"), nil
}

// migrateLegacyContext copies ~/.acloud-context.yaml → XDG path when absent (#176).
func migrateLegacyContext() {
	newPath, err := getContextFilePath()
	if err != nil {
		return
	}
	if _, err := os.Stat(newPath); err == nil {
		return
	}
	oldPath, err := legacyContextPath()
	if err != nil {
		return
	}
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(newPath), FilePermDirAll); err != nil {
		return
	}
	if err := os.WriteFile(newPath, data, FilePermConfig); err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "notice: context migrated from %s to %s\n", oldPath, newPath)
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

// ─── Context Set ──────────────────────────────────────────────────────────────

// ContextSetArgs holds parsed arguments for the context set command.
type ContextSetArgs struct {
	Name      string
	ProjectID string
}

// NewContextSetArgsFromCobraCommand parses and validates args for context set.
func NewContextSetArgsFromCobraCommand(cmd *cobra.Command, cmdArgs []string) (*ContextSetArgs, error) {
	args := &ContextSetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cmdArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// ParseFromCobraCommand reads positional arg and flags into the struct.
func (a *ContextSetArgs) ParseFromCobraCommand(cmd *cobra.Command, cmdArgs []string) error {
	a.Name = cmdArgs[0]
	var err error
	if a.ProjectID, err = cmd.Flags().GetString("project-id"); err != nil {
		return err
	}
	return nil
}

// Validate performs pure validation on the parsed args.
func (a *ContextSetArgs) Validate() error {
	var errs []error
	if a.ProjectID == "" {
		errs = append(errs, errors.New("--project-id is required"))
	}
	return errors.Join(errs...)
}

// ContextSet creates or updates a named context.
func ContextSet(_ context.Context, args ContextSetArgs) error {
	ctx, err := LoadContext()
	if err != nil {
		ctx = &Context{
			Contexts: make(map[string]CtxInfo),
		}
	}

	ctx.Contexts[args.Name] = CtxInfo{
		ProjectID: args.ProjectID,
	}

	if err := SaveContext(ctx); err != nil {
		return fmt.Errorf("saving context: %w", err)
	}

	fmt.Printf("Context '%s' set with project ID: %s\n", args.Name, args.ProjectID)
	return nil
}

// ContextSetRun is the RunE wiring for the context set command.
func ContextSetRun(cmd *cobra.Command, cmdArgs []string) error {
	args, err := NewContextSetArgsFromCobraCommand(cmd, cmdArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContextSet(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ─── Context Use ──────────────────────────────────────────────────────────────

// ContextUseArgs holds parsed arguments for the context use command.
type ContextUseArgs struct {
	Name string
}

// NewContextUseArgsFromCobraCommand parses and validates args for context use.
func NewContextUseArgsFromCobraCommand(_ *cobra.Command, cmdArgs []string) (*ContextUseArgs, error) {
	args := &ContextUseArgs{}
	if err := args.ParseFromCobraCommand(cmdArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// ParseFromCobraCommand reads the positional arg into the struct.
func (a *ContextUseArgs) ParseFromCobraCommand(cmdArgs []string) error {
	a.Name = cmdArgs[0]
	return nil
}

// Validate performs pure validation on the parsed args.
func (a *ContextUseArgs) Validate() error {
	return nil
}

// ContextUse switches the active context to the named context.
func ContextUse(_ context.Context, args ContextUseArgs) error {
	ctx, err := LoadContext()
	if err != nil {
		return fmt.Errorf("loading context: %w", err)
	}

	if _, exists := ctx.Contexts[args.Name]; !exists {
		available := make([]string, 0, len(ctx.Contexts))
		for name := range ctx.Contexts {
			available = append(available, name)
		}
		return fmt.Errorf("context '%s' not found; available contexts: %v", args.Name, available)
	}

	ctx.CurrentContext = args.Name

	if err := SaveContext(ctx); err != nil {
		return fmt.Errorf("saving context: %w", err)
	}

	fmt.Printf("Switched to context '%s'\n", args.Name)
	fmt.Printf("Project ID: %s\n", ctx.Contexts[args.Name].ProjectID)
	return nil
}

// ContextUseRun is the RunE wiring for the context use command.
func ContextUseRun(cmd *cobra.Command, cmdArgs []string) error {
	args, err := NewContextUseArgsFromCobraCommand(cmd, cmdArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContextUse(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ─── Context List ─────────────────────────────────────────────────────────────

// ContextListArgs holds the (empty) arguments for the context list command.
type ContextListArgs struct{}

// NewContextListArgsFromCobraCommand parses and validates args for context list.
func NewContextListArgsFromCobraCommand(_ *cobra.Command) (*ContextListArgs, error) {
	return &ContextListArgs{}, nil
}

// ContextList prints all configured contexts.
func ContextList(_ context.Context, _ ContextListArgs) error {
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

// ContextListRun is the RunE wiring for the context list command.
func ContextListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewContextListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContextList(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ─── Context Current ──────────────────────────────────────────────────────────

// ContextCurrentArgs holds the (empty) arguments for the context current command.
type ContextCurrentArgs struct{}

// NewContextCurrentArgsFromCobraCommand parses and validates args for context current.
func NewContextCurrentArgsFromCobraCommand(_ *cobra.Command) (*ContextCurrentArgs, error) {
	return &ContextCurrentArgs{}, nil
}

// ContextCurrent prints the active context name and its project ID.
func ContextCurrent(_ context.Context, _ ContextCurrentArgs) error {
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

// ContextCurrentRun is the RunE wiring for the context current command.
func ContextCurrentRun(cmd *cobra.Command, _ []string) error {
	args, err := NewContextCurrentArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContextCurrent(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ─── Context Delete ───────────────────────────────────────────────────────────

// ContextDeleteArgs holds parsed arguments for the context delete command.
type ContextDeleteArgs struct {
	Name string
}

// NewContextDeleteArgsFromCobraCommand parses and validates args for context delete.
func NewContextDeleteArgsFromCobraCommand(_ *cobra.Command, cmdArgs []string) (*ContextDeleteArgs, error) {
	args := &ContextDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmdArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// ParseFromCobraCommand reads the positional arg into the struct.
func (a *ContextDeleteArgs) ParseFromCobraCommand(cmdArgs []string) error {
	a.Name = cmdArgs[0]
	return nil
}

// Validate performs pure validation on the parsed args.
func (a *ContextDeleteArgs) Validate() error {
	return nil
}

// ContextDelete removes a named context from the context file.
func ContextDelete(_ context.Context, args ContextDeleteArgs) error {
	ctx, err := LoadContext()
	if err != nil {
		return fmt.Errorf("loading context: %w", err)
	}

	if _, exists := ctx.Contexts[args.Name]; !exists {
		return fmt.Errorf("context '%s' not found", args.Name)
	}

	delete(ctx.Contexts, args.Name)

	// If this was the active context, clear the pointer.
	if ctx.CurrentContext == args.Name {
		ctx.CurrentContext = ""
	}

	if err := SaveContext(ctx); err != nil {
		return fmt.Errorf("saving context: %w", err)
	}

	fmt.Printf("Context '%s' deleted\n", args.Name)
	return nil
}

// ContextDeleteRun is the RunE wiring for the context delete command.
func ContextDeleteRun(cmd *cobra.Command, cmdArgs []string) error {
	args, err := NewContextDeleteArgsFromCobraCommand(cmd, cmdArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContextDelete(ctx, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
