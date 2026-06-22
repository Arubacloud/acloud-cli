package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	dbaasCmd.AddCommand(dbaasDatabaseCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseCreateCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseGetCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseUpdateCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseDeleteCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseListCmd)

	dbaasDatabaseCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseCreateCmd.Flags().String("name", "", "Database name (required)")
	dbaasDatabaseCreateCmd.MarkFlagRequired("name")

	dbaasDatabaseGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasDatabaseUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseUpdateCmd.Flags().String("name", "", "New database name")

	dbaasDatabaseDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasDatabaseDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasDatabaseListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasDatabaseListCmd.Flags().Int("offset", 0, "Number of results to skip")

	dbaasDatabaseCreateCmd.ValidArgsFunction = completeDBaasDatabaseCreate
	dbaasDatabaseListCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseGetCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseUpdateCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseDeleteCmd.ValidArgsFunction = completeDBaaSDatabaseID
}

func completeDBaasDatabaseCreate(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeDBaaSID(cmd, args, toComplete)
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func completeDBaaSDatabaseID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeDBaaSID(cmd, args, toComplete)
	}
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	dbaasID := args[0]

	key := cacheKey("database", projectID, dbaasID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromDatabase().Databases().List(ctx, dbaasRef(projectID, dbaasID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, db := range list.Items() {
			raw := db.Raw()
			if raw != nil && raw.Name != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", raw.Name, raw.Name))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

type databaseGetView struct {
	Name, CreatedAt, CreatedBy string
}

var dbaasDatabaseCmd = &cobra.Command{
	Use:   "database [dbaas-id]",
	Short: "Manage databases in DBaaS",
	Long:  `Perform CRUD operations on databases in DBaaS.`,
}

var dbaasDatabaseCreateCmd = &cobra.Command{
	Use:   "create [dbaas-id]",
	Short: "Create a new database in DBaaS",
	Long: `Create a new database schema inside an existing DBaaS instance.

The DBaaS instance must already exist and be in a ready state.
Use 'acloud database dbaas get <dbaas-id>' to check its status.`,
	Example: `  acloud database dbaas database create <dbaas-id> --name myapp_db`,
	Args:    cobra.ExactArgs(1),
	RunE:    DatabaseDBaaSDatabaseCreateRun,
}

var dbaasDatabaseGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [database-name]",
	Short: "Get database details",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSDatabaseGetRun,
}

var dbaasDatabaseListCmd = &cobra.Command{
	Use:   "list [dbaas-id]",
	Short: "List all databases in DBaaS",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSDatabaseListRun,
}

var dbaasDatabaseUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id] [database-name]",
	Short: "Update a database",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSDatabaseUpdateRun,
}

var dbaasDatabaseDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [database-name]",
	Short: "Delete a database",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSDatabaseDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// DatabaseDBaaSDatabaseCreateArgs holds the typed arguments for creating a database.
type DatabaseDBaaSDatabaseCreateArgs struct {
	ProjectID string
	DBaaSID   string
	Name      string
}

// DatabaseDBaaSDatabaseGetArgs holds the typed arguments for getting a database.
type DatabaseDBaaSDatabaseGetArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
}

// DatabaseDBaaSDatabaseUpdateArgs holds the typed arguments for updating a database.
type DatabaseDBaaSDatabaseUpdateArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
	Name         string
}

// DatabaseDBaaSDatabaseDeleteArgs holds the typed arguments for deleting a database.
type DatabaseDBaaSDatabaseDeleteArgs struct {
	ProjectID    string
	DBaaSID      string
	DatabaseName string
	DryRun       bool
	SkipConfirm  bool
}

// DatabaseDBaaSDatabaseListArgs holds the typed arguments for listing databases.
type DatabaseDBaaSDatabaseListArgs struct {
	ProjectID string
	DBaaSID   string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewDatabaseDBaaSDatabaseCreateArgsFromCobraCommand parses and validates args for create.
func NewDatabaseDBaaSDatabaseCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSDatabaseCreateArgs, error) {
	args := &DatabaseDBaaSDatabaseCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSDatabaseGetArgsFromCobraCommand parses and validates args for get.
func NewDatabaseDBaaSDatabaseGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSDatabaseGetArgs, error) {
	args := &DatabaseDBaaSDatabaseGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSDatabaseUpdateArgsFromCobraCommand parses and validates args for update.
func NewDatabaseDBaaSDatabaseUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSDatabaseUpdateArgs, error) {
	args := &DatabaseDBaaSDatabaseUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSDatabaseDeleteArgsFromCobraCommand parses and validates args for delete.
func NewDatabaseDBaaSDatabaseDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSDatabaseDeleteArgs, error) {
	args := &DatabaseDBaaSDatabaseDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSDatabaseListArgsFromCobraCommand parses and validates args for list.
func NewDatabaseDBaaSDatabaseListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSDatabaseListArgs, error) {
	args := &DatabaseDBaaSDatabaseListArgs{}
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
func (a *DatabaseDBaaSDatabaseCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *DatabaseDBaaSDatabaseGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *DatabaseDBaaSDatabaseUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *DatabaseDBaaSDatabaseDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the list args struct.
func (a *DatabaseDBaaSDatabaseListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *DatabaseDBaaSDatabaseCreateArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *DatabaseDBaaSDatabaseGetArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *DatabaseDBaaSDatabaseUpdateArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}
	if a.Name == "" {
		errs = append(errs, errors.New("--name is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *DatabaseDBaaSDatabaseDeleteArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("database name is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *DatabaseDBaaSDatabaseListArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}

	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// DatabaseDBaaSDatabaseCreate creates a new database inside a DBaaS instance.
func DatabaseDBaaSDatabaseCreate(ctx context.Context, client aruba.Client, args DatabaseDBaaSDatabaseCreateArgs) error {
	db := aruba.NewDatabase().
		InDBaaS(dbaasRef(args.ProjectID, args.DBaaSID)).
		Named(args.Name)

	created, err := client.FromDatabase().Databases().Create(ctx, db)
	if err != nil {
		return fmt.Errorf("creating database: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		fmt.Printf("\n%s\n", msgCreated("Database", args.Name))
		fmt.Printf("Name:            %s\n", raw.Name)
		if raw.CreationDate != nil {
			fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
		}
	} else {
		fmt.Println(msgCreatedAsync("Database", args.Name))
	}
	return nil
}

// DatabaseDBaaSDatabaseGet retrieves and displays a database's details.
func DatabaseDBaaSDatabaseGet(ctx context.Context, client aruba.Client, args DatabaseDBaaSDatabaseGetArgs) error {
	db, err := client.FromDatabase().Databases().Get(ctx, databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName))
	if err != nil {
		return fmt.Errorf("getting database: %w", apiErrFromV2(err))
	}

	if db == nil || db.Raw() == nil {
		fmt.Println("Database not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(db, nil, nil)
		return nil
	}
	raw := db.Raw()
	view := databaseGetView{Name: raw.Name}
	if raw.CreationDate != nil {
		view.CreatedAt = raw.CreationDate.Format(DateLayout)
	}
	if raw.CreatedBy != nil {
		view.CreatedBy = *raw.CreatedBy
	}
	return renderGet(databaseGetTmpl, view)
}

// DatabaseDBaaSDatabaseUpdate updates a database's name.
func DatabaseDBaaSDatabaseUpdate(ctx context.Context, client aruba.Client, args DatabaseDBaaSDatabaseUpdateArgs) error {
	db, err := client.FromDatabase().Databases().Get(ctx, databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName))
	if err != nil {
		return fmt.Errorf("getting database: %w", apiErrFromV2(err))
	}
	if db == nil || db.Raw() == nil {
		return fmt.Errorf("database not found")
	}

	db.Named(args.Name)

	updated, err := client.FromDatabase().Databases().Update(ctx, db)
	if err != nil {
		return fmt.Errorf("updating database: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		fmt.Printf("\n%s\n", msgUpdated("Database", args.DatabaseName))
		fmt.Printf("Name:            %s\n", updated.Raw().Name)
	} else {
		fmt.Println(msgUpdatedAsync("Database", args.DatabaseName))
	}
	return nil
}

// DatabaseDBaaSDatabaseDelete deletes a database from a DBaaS instance.
func DatabaseDBaaSDatabaseDelete(ctx context.Context, client aruba.Client, args DatabaseDBaaSDatabaseDeleteArgs) error {
	err := client.FromDatabase().Databases().Delete(ctx, databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName))
	if err != nil {
		return fmt.Errorf("deleting database: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Database", args.DatabaseName))
	return nil
}

// DatabaseDBaaSDatabaseList lists all databases in a DBaaS instance.
func DatabaseDBaaSDatabaseList(ctx context.Context, client aruba.Client, args DatabaseDBaaSDatabaseListArgs) error {
	list, err := client.FromDatabase().Databases().List(ctx, dbaasRef(args.ProjectID, args.DBaaSID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing databases: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No databases found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.Database]{
		{TableColumn: TableColumn{Header: "NAME", Width: 40}, Value: func(db *aruba.Database) string {
			if r := db.Raw(); r != nil {
				return r.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "CREATION DATE", Width: 25}, Value: func(db *aruba.Database) string {
			if r := db.Raw(); r != nil && r.CreationDate != nil {
				return r.CreationDate.Format(DateLayout)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "CREATED BY", Width: 30}, Value: func(db *aruba.Database) string {
			if r := db.Raw(); r != nil && r.CreatedBy != nil {
				return *r.CreatedBy
			}
			return ""
		}},
	}, list.Items(), func(db *aruba.Database) bool { return db.Raw() != nil })
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// DatabaseDBaaSDatabaseCreateRun is the Cobra RunE handler for database create.
func DatabaseDBaaSDatabaseCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSDatabaseCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSDatabaseCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSDatabaseGetRun is the Cobra RunE handler for database get.
func DatabaseDBaaSDatabaseGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSDatabaseGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSDatabaseGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSDatabaseUpdateRun is the Cobra RunE handler for database update.
func DatabaseDBaaSDatabaseUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSDatabaseUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSDatabaseUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSDatabaseDeleteRun is the Cobra RunE handler for database delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func DatabaseDBaaSDatabaseDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSDatabaseDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete(fmt.Sprintf("database '%s' in DBaaS instance", args.DatabaseName), args.DBaaSID)
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
		if _, err := client.FromDatabase().Databases().Get(ctx, databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName)); err != nil {
			return fmt.Errorf("dry-run: database not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("database", args.DatabaseName))
		return nil
	}

	if err := DatabaseDBaaSDatabaseDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSDatabaseListRun is the Cobra RunE handler for database list.
func DatabaseDBaaSDatabaseListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSDatabaseListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSDatabaseList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
