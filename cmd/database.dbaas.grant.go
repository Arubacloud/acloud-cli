package cmd

import (
	"context"
	"errors"
	"fmt"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// databaseRef returns a Ref for a specific database inside a DBaaS instance.
func databaseRef(projectID, dbaasID, dbName string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Database/dbaas/" + dbaasID +
		"/databases/" + dbName)
}

func init() {
	dbaasCmd.AddCommand(dbaasGrantCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantCreateCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantListCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantGetCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantDeleteCmd)

	dbaasGrantCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasGrantCreateCmd.Flags().String("username", "", "Username to grant access to (required)")
	dbaasGrantCreateCmd.Flags().String("role", "", "Role to grant (e.g. liteadmin, readonly, readwrite) (required)")
	dbaasGrantCreateCmd.MarkFlagRequired("username")
	dbaasGrantCreateCmd.MarkFlagRequired("role")

	dbaasGrantListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasGrantListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasGrantListCmd.Flags().Int("offset", 0, "Number of results to skip")

	dbaasGrantGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasGrantDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasGrantDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasGrantDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
}

// grantRef returns a Ref for a specific grant inside a database inside a DBaaS instance.
func grantRef(projectID, dbaasID, dbName, grantID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + dbName + "/grants/" + grantID)
}

func completeGrantID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	dbaasID, dbName := args[0], args[1]

	key := cacheKey("grant", projectID, dbaasID, dbName)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ctx := context.Background()
	list, err := client.FromDatabase().Grants().List(ctx, databaseRef(projectID, dbaasID, dbName))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	if list != nil {
		for _, g := range list.Items() {
			id := g.ID()
			completions = append(completions, fmt.Sprintf("%s\t%s→%s", id, g.Username(), g.RoleName()))
		}
	}
	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var dbaasGrantCmd = &cobra.Command{
	Use:   "grant [dbaas-id] [database-name]",
	Short: "Manage grants on a database",
	Long:  `Assign or revoke user access roles on a database inside a DBaaS instance.`,
}

var dbaasGrantCreateCmd = &cobra.Command{
	Use:   "create [dbaas-id] [database-name]",
	Short: "Grant a user access to a database",
	Long: `Grant a user a role on a specific database inside a DBaaS instance.

Common roles: liteadmin, readonly, readwrite.`,
	Example: `  acloud database dbaas grant create <dbaas-id> mydb --username myuser --role liteadmin`,
	Args:    cobra.ExactArgs(2),
	RunE:    DatabaseDBaaSGrantCreateRun,
}

var dbaasGrantListCmd = &cobra.Command{
	Use:   "list [dbaas-id] [database-name]",
	Short: "List all grants on a database",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSGrantListRun,
}

var dbaasGrantGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [database-name] [grant-id]",
	Short: "Get grant details",
	Args:  cobra.ExactArgs(3),
	RunE:  DatabaseDBaaSGrantGetRun,
}

var dbaasGrantDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [database-name] [grant-id]",
	Short: "Delete a grant",
	Args:  cobra.ExactArgs(3),
	RunE:  DatabaseDBaaSGrantDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// DatabaseDBaaSGrantCreateArgs holds the typed arguments for creating a grant.
type DatabaseDBaaSGrantCreateArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
	Username     string
	Role         string
}

// DatabaseDBaaSGrantListArgs holds the typed arguments for listing grants.
type DatabaseDBaaSGrantListArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
	CallOpts     []aruba.CallOption
}

// DatabaseDBaaSGrantGetArgs holds the typed arguments for getting a grant.
type DatabaseDBaaSGrantGetArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
	GrantID      string
}

// DatabaseDBaaSGrantDeleteArgs holds the typed arguments for deleting a grant.
type DatabaseDBaaSGrantDeleteArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
	GrantID      string
	DryRun       bool
	SkipConfirm  bool
}

// =============================================================================
// Constructors
// =============================================================================

// NewDatabaseDBaaSGrantCreateArgsFromCobraCommand parses and validates args for create.
func NewDatabaseDBaaSGrantCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSGrantCreateArgs, error) {
	args := &DatabaseDBaaSGrantCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSGrantListArgsFromCobraCommand parses and validates args for list.
func NewDatabaseDBaaSGrantListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSGrantListArgs, error) {
	args := &DatabaseDBaaSGrantListArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSGrantGetArgsFromCobraCommand parses and validates args for get.
func NewDatabaseDBaaSGrantGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSGrantGetArgs, error) {
	args := &DatabaseDBaaSGrantGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSGrantDeleteArgsFromCobraCommand parses and validates args for delete.
func NewDatabaseDBaaSGrantDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSGrantDeleteArgs, error) {
	args := &DatabaseDBaaSGrantDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the create args struct.
func (a *DatabaseDBaaSGrantCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.DatabaseName = cobraArgs[1]
	}
	if a.Username, err = cmd.Flags().GetString("username"); err != nil {
		errs = append(errs, err)
	}
	if a.Role, err = cmd.Flags().GetString("role"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the list args struct.
func (a *DatabaseDBaaSGrantListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.DatabaseName = cobraArgs[1]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *DatabaseDBaaSGrantGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.DatabaseName = cobraArgs[1]
	}
	if len(cobraArgs) > 2 {
		a.GrantID = cobraArgs[2]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *DatabaseDBaaSGrantDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.DatabaseName = cobraArgs[1]
	}
	if len(cobraArgs) > 2 {
		a.GrantID = cobraArgs[2]
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *DatabaseDBaaSGrantCreateArgs) Validate() error {
	var errs []error
	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}
	if a.Username == "" {
		errs = append(errs, errors.New("--username is required"))
	}
	if a.Role == "" {
		errs = append(errs, errors.New("--role is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *DatabaseDBaaSGrantListArgs) Validate() error {
	var errs []error
	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *DatabaseDBaaSGrantGetArgs) Validate() error {
	var errs []error
	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}
	if a.GrantID == "" {
		errs = append(errs, errors.New("grant ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *DatabaseDBaaSGrantDeleteArgs) Validate() error {
	var errs []error
	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}
	if a.GrantID == "" {
		errs = append(errs, errors.New("grant ID is required"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// DatabaseDBaaSGrantCreate creates a new grant on a database.
func DatabaseDBaaSGrantCreate(ctx context.Context, client aruba.Client, args DatabaseDBaaSGrantCreateArgs) error {
	g := aruba.NewGrant().
		InDatabase(databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName)).
		ForUser(args.Username).
		OfRole(args.Role)

	created, err := client.FromDatabase().Grants().Create(ctx, g)
	if err != nil {
		return fmt.Errorf("creating grant: %w", apiErrFromV2(err))
	}

	if created != nil {
		fmt.Printf("\n%s\n", msgCreated("Grant", args.Username))
		fmt.Printf("Username:        %s\n", created.Username())
		fmt.Printf("Role:            %s\n", created.RoleName())
		fmt.Printf("Database:        %s\n", created.DatabaseName())
	} else {
		fmt.Printf("\n%s\n", msgCreatedAsync("Grant", args.Username))
	}
	return nil
}

// DatabaseDBaaSGrantList lists all grants on a database.
func DatabaseDBaaSGrantList(ctx context.Context, client aruba.Client, args DatabaseDBaaSGrantListArgs) error {
	list, err := client.FromDatabase().Grants().List(ctx, databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing grants: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No grants found")
		return nil
	}
	renderList(nil, []ListColumn[*aruba.Grant]{
		{TableColumn: TableColumn{Header: "GRANT ID", Width: 30}, Value: func(g *aruba.Grant) string { return g.ID() }},
		{TableColumn: TableColumn{Header: "USERNAME", Width: 30}, Value: func(g *aruba.Grant) string { return g.Username() }},
		{TableColumn: TableColumn{Header: "ROLE", Width: 20}, Value: func(g *aruba.Grant) string { return g.RoleName() }},
		{TableColumn: TableColumn{Header: "DATABASE", Width: 25}, Value: func(g *aruba.Grant) string { return g.DatabaseName() }},
	}, list.Items())
	return nil
}

// DatabaseDBaaSGrantGet retrieves and displays a grant's details.
func DatabaseDBaaSGrantGet(ctx context.Context, client aruba.Client, args DatabaseDBaaSGrantGetArgs) error {
	got, err := client.FromDatabase().Grants().Get(ctx, grantRef(args.ProjectID, args.DBaaSID, args.DatabaseName, args.GrantID))
	if err != nil {
		return fmt.Errorf("getting grant: %w", apiErrFromV2(err))
	}

	if got != nil {
		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(got, nil, nil)
			return nil
		}
		fmt.Println("\nGrant Details:")
		fmt.Println("==============")
		fmt.Printf("Username:        %s\n", got.Username())
		fmt.Printf("Role:            %s\n", got.RoleName())
		fmt.Printf("Database:        %s\n", got.DatabaseName())
		if t := got.CreatedAt(); !t.IsZero() {
			fmt.Printf("Creation Date:   %s\n", t.Format(DateLayout))
		}
		if s := got.CreatedBy(); s != "" {
			fmt.Printf("Created By:      %s\n", s)
		}
		fmt.Println()
	} else {
		fmt.Println("Grant not found")
	}
	return nil
}

// DatabaseDBaaSGrantDelete deletes a grant.
func DatabaseDBaaSGrantDelete(ctx context.Context, client aruba.Client, args DatabaseDBaaSGrantDeleteArgs) error {
	if err := client.FromDatabase().Grants().Delete(ctx, grantRef(args.ProjectID, args.DBaaSID, args.DatabaseName, args.GrantID)); err != nil {
		return fmt.Errorf("deleting grant: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Grant", args.GrantID))
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// DatabaseDBaaSGrantCreateRun is the Cobra RunE handler for grant create.
func DatabaseDBaaSGrantCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSGrantCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSGrantCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSGrantListRun is the Cobra RunE handler for grant list.
func DatabaseDBaaSGrantListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSGrantListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSGrantList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSGrantGetRun is the Cobra RunE handler for grant get.
func DatabaseDBaaSGrantGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSGrantGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSGrantGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSGrantDeleteRun is the Cobra RunE handler for grant delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func DatabaseDBaaSGrantDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSGrantDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete(fmt.Sprintf("grant '%s' on database '%s'", args.GrantID, args.DatabaseName), args.DBaaSID)
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

	if args.DryRun {
		if _, err := client.FromDatabase().Grants().Get(ctx, grantRef(args.ProjectID, args.DBaaSID, args.DatabaseName, args.GrantID)); err != nil {
			return fmt.Errorf("dry-run: grant not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("grant", args.GrantID))
		return nil
	}

	if err := DatabaseDBaaSGrantDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
