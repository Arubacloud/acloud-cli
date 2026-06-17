package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	managementCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add completion for project IDs
	projectGetCmd.ValidArgsFunction = completeProjectID
	projectUpdateCmd.ValidArgsFunction = completeProjectID
	projectDeleteCmd.ValidArgsFunction = completeProjectID

	// Add flags for project create command
	projectCreateCmd.Flags().String("name", "", "Name for the project (required)")
	projectCreateCmd.Flags().String("description", "", "Description for the project")
	projectCreateCmd.Flags().StringSlice("tags", []string{}, "Tags for the project (comma-separated)")
	projectCreateCmd.Flags().Bool("default", false, "Set as default project")
	projectCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	projectCreateCmd.MarkFlagRequired("name")

	// Add flags for project delete command
	projectDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	projectDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	// Add flags for project update command
	projectUpdateCmd.Flags().String("description", "", "New description for the project")
	projectUpdateCmd.Flags().StringSlice("tags", []string{}, "Tags for the project (comma-separated)")

	// Add flags for project list command
	projectListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	projectListCmd.Flags().Int32("offset", 0, "Number of results to skip")
}

// completeProjectID provides completion for project IDs
func completeProjectID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Allow completion even if args exist - user might be completing a partial ID

	key := cacheKey("project")
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	// Get SDK client
	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// List projects
	ctx := context.Background()
	list, err := client.FromProject().List(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, p := range list.Items() {
			id := p.ID()
			name := p.Name()
			if id != "" {
				// Format: "id\tname" - the tab separates the completion from the description
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

type projectGetView struct {
	ID, Name, Description, Default, CreatedAt, UpdatedAt, CreatedBy, UpdatedBy, Tags string
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Perform CRUD operations on projects in Aruba Cloud.`,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long: `Create a new project in your Aruba Cloud account.

Pass --default to make this project the active default for subsequent commands.
Use 'acloud context set-project' to switch the active project at any time.`,
	Example: `  acloud management project create --name my-project --description "Production workloads"
  acloud management project create --name dev-project --default`,
	Args: cobra.NoArgs,
	RunE: ManagementProjectCreateRun,
}

var projectGetCmd = &cobra.Command{
	Use:   "get [project-id]",
	Short: "Get project details",
	Args:  cobra.ExactArgs(1),
	RunE:  ManagementProjectGetRun,
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project-id]",
	Short: "Update a project (description and/or tags only)",
	Args:  cobra.ExactArgs(1),
	RunE:  ManagementProjectUpdateRun,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project-id]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE:  ManagementProjectDeleteRun,
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Args:  cobra.NoArgs,
	RunE:  ManagementProjectListRun,
}

// =============================================================================
// Args structs
// =============================================================================

// ManagementProjectCreateArgs holds the typed arguments for creating a project.
type ManagementProjectCreateArgs struct {
	Name        string
	Description string
	Tags        []string
	SetDefault  bool
	Verbose     bool
}

// ManagementProjectGetArgs holds the typed arguments for getting a project.
type ManagementProjectGetArgs struct {
	ID string
}

// ManagementProjectUpdateArgs holds the typed arguments for updating a project.
type ManagementProjectUpdateArgs struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	TagsChanged bool
}

// ManagementProjectDeleteArgs holds the typed arguments for deleting a project.
type ManagementProjectDeleteArgs struct {
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// ManagementProjectListArgs holds the typed arguments for listing projects.
type ManagementProjectListArgs struct {
	CallOpts []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewManagementProjectCreateArgsFromCobraCommand parses and validates args for create.
func NewManagementProjectCreateArgsFromCobraCommand(cmd *cobra.Command) (*ManagementProjectCreateArgs, error) {
	args := &ManagementProjectCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewManagementProjectGetArgsFromCobraCommand parses and validates args for get.
func NewManagementProjectGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ManagementProjectGetArgs, error) {
	args := &ManagementProjectGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewManagementProjectUpdateArgsFromCobraCommand parses and validates args for update.
func NewManagementProjectUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ManagementProjectUpdateArgs, error) {
	args := &ManagementProjectUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewManagementProjectDeleteArgsFromCobraCommand parses and validates args for delete.
func NewManagementProjectDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ManagementProjectDeleteArgs, error) {
	args := &ManagementProjectDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewManagementProjectListArgsFromCobraCommand parses and validates args for list.
func NewManagementProjectListArgsFromCobraCommand(cmd *cobra.Command) (*ManagementProjectListArgs, error) {
	args := &ManagementProjectListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// =============================================================================
// ParseFromCobraCommand methods
// =============================================================================

// ParseFromCobraCommand reads Cobra flags into the create args struct.
func (a *ManagementProjectCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Description, err = cmd.Flags().GetString("description"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.SetDefault, err = cmd.Flags().GetBool("default"); err != nil {
		errs = append(errs, err)
	}
	if a.Verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra positional args into the get args struct.
func (a *ManagementProjectGetArgs) ParseFromCobraCommand(_ *cobra.Command, cobraArgs []string) error {
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	return nil
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *ManagementProjectUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.Description, err = cmd.Flags().GetString("description"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *ManagementProjectDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *ManagementProjectListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	a.CallOpts = listOpts(cmd)
	return nil
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *ManagementProjectCreateArgs) Validate() error {
	var errs []error

	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *ManagementProjectGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *ManagementProjectUpdateArgs) Validate() error {
	if a.Description == "" && !a.TagsChanged {
		return errors.New("at least one of --description or --tags must be provided")
	}
	return nil
}

// Validate checks the delete args for correctness.
func (a *ManagementProjectDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *ManagementProjectListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// ManagementProjectCreate creates a new project.
func ManagementProjectCreate(ctx context.Context, client aruba.Client, args ManagementProjectCreateArgs) error {
	// Build the create request
	proj := aruba.NewProject().Named(args.Name).RetaggedAs(args.Tags...)
	if args.Description != "" {
		proj.DescribedAs(args.Description)
	}
	if args.SetDefault {
		proj.AsDefault()
	}

	// Debug output if verbose
	if args.Verbose {
		fmt.Println("Creating project with the following parameters:")
		fmt.Printf("  Name:        %s\n", args.Name)
		if args.Description != "" {
			fmt.Printf("  Description: %s\n", args.Description)
		}
		fmt.Printf("  Default:     %t\n", args.SetDefault)
		if len(args.Tags) > 0 {
			fmt.Printf("  Tags:        %v\n", args.Tags)
		}
		fmt.Println()
	}

	created, err := client.FromProject().Create(ctx, proj)
	if err != nil {
		return fmt.Errorf("creating project: %w", apiErrFromV2(err))
	}

	if created != nil && created.ID() != "" {
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "DEFAULT", Width: 10},
		}
		defaultVal := "No"
		if created.IsDefault() {
			defaultVal = "Yes"
		}
		row := []string{created.ID(), created.Name(), defaultVal}
		PrintOutput(created, headers, [][]string{row})
	} else {
		fmt.Println(msgCreatedAsync("Project", args.Name))
	}
	return nil
}

// ManagementProjectGet retrieves project details.
func ManagementProjectGet(ctx context.Context, client aruba.Client, args ManagementProjectGetArgs) error {
	p, err := client.FromProject().Get(ctx, projectRef(args.ID))
	if err != nil {
		return fmt.Errorf("getting project: %w", apiErrFromV2(err))
	}

	if p == nil || p.ID() == "" {
		fmt.Println("Project not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(p, nil, nil)
		return nil
	}
	view := projectGetView{
		ID:          p.ID(),
		Name:        p.Name(),
		Description: p.Description(),
		Default:     fmt.Sprintf("%t", p.IsDefault()),
		CreatedBy:   p.CreatedBy(),
		UpdatedBy:   p.UpdatedBy(),
		Tags:        "[]",
	}
	if !p.CreatedAt().IsZero() {
		view.CreatedAt = p.CreatedAt().Format(DateLayout)
	}
	if !p.UpdatedAt().IsZero() {
		view.UpdatedAt = p.UpdatedAt().Format(DateLayout)
	}
	if tags := p.Tags(); len(tags) > 0 {
		view.Tags = fmt.Sprintf("%v", tags)
	}
	return renderGet(projectGetTmpl, view)
}

// ManagementProjectUpdate updates a project's description and/or tags.
func ManagementProjectUpdate(ctx context.Context, client aruba.Client, args ManagementProjectUpdateArgs) error {
	// First, get the current project details to preserve existing values.
	p, err := client.FromProject().Get(ctx, projectRef(args.ID))
	if err != nil {
		return fmt.Errorf("fetching current project: %w", apiErrFromV2(err))
	}

	if p == nil || p.ID() == "" {
		return fmt.Errorf("project not found or no data returned")
	}

	// Apply updates
	if args.Description != "" {
		p.DescribedAs(args.Description)
	}
	if args.TagsChanged {
		p.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromProject().Update(ctx, p)
	if err != nil {
		return fmt.Errorf("updating project: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.ID() != "" {
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "DEFAULT", Width: 10},
		}
		defaultVal := "No"
		if updated.IsDefault() {
			defaultVal = "Yes"
		}
		row := []string{updated.ID(), updated.Name(), defaultVal}
		PrintOutput(updated, headers, [][]string{row})
	} else {
		fmt.Println(msgUpdatedAsync("Project", args.ID))
	}
	return nil
}

// ManagementProjectDelete deletes a project.
func ManagementProjectDelete(ctx context.Context, client aruba.Client, args ManagementProjectDeleteArgs) error {
	if args.DryRun {
		_, err := client.FromProject().Get(ctx, projectRef(args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: project not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("project", args.ID))
		return nil
	}

	err := client.FromProject().Delete(ctx, projectRef(args.ID))
	if err != nil {
		return fmt.Errorf("deleting project: %w", apiErrFromV2(err))
	}

	headers := []TableColumn{
		{Header: "ID", Width: 30},
		{Header: "STATUS", Width: 15},
	}
	status := "deleted"
	result := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{args.ID, status}
	PrintOutput(result, headers, [][]string{{args.ID, status}})
	return nil
}

// ManagementProjectList lists all projects.
func ManagementProjectList(ctx context.Context, client aruba.Client, args ManagementProjectListArgs) error {
	list, err := client.FromProject().List(ctx, args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing projects: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No projects found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.Project]{
		{TableColumn: TableColumn{Header: "NAME", Width: 40}, Value: func(p *aruba.Project) string { return p.Name() }},
		{TableColumn: TableColumn{Header: "ID", Width: 30}, Value: func(p *aruba.Project) string { return p.ID() }},
		{TableColumn: TableColumn{Header: "CREATION DATE", Width: 15}, Value: func(p *aruba.Project) string {
			if p.CreatedAt().IsZero() {
				return "N/A"
			}
			return p.CreatedAt().Format("02-01-2006")
		}},
	}, list.Items())
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// ManagementProjectCreateRun is the RunE wiring for project create.
func ManagementProjectCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewManagementProjectCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ManagementProjectCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ManagementProjectGetRun is the RunE wiring for project get.
func ManagementProjectGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewManagementProjectGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ManagementProjectGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ManagementProjectUpdateRun is the RunE wiring for project update.
func ManagementProjectUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewManagementProjectUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ManagementProjectUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ManagementProjectDeleteRun is the RunE wiring for project delete.
func ManagementProjectDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewManagementProjectDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("project", a.ID)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ManagementProjectDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ManagementProjectListRun is the RunE wiring for project list.
func ManagementProjectListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewManagementProjectListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ManagementProjectList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
