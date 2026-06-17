package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	databaseCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupGetCmd)
	backupCmd.AddCommand(backupDeleteCmd)
	backupCmd.AddCommand(backupListCmd)

	backupCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupCreateCmd.Flags().String("name", "", "Backup name (required)")
	backupCreateCmd.Flags().String("region", "", "Region code (required)")
	backupCreateCmd.Flags().String("zone", "", "Availability zone (e.g. ITBG-1); defaults to region when omitted")
	backupCreateCmd.Flags().String("dbaas-id", "", "DBaaS instance ID (required)")
	backupCreateCmd.Flags().String("database-name", "", "Database name (required)")
	backupCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
	backupCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	backupCreateCmd.MarkFlagRequired("name")
	backupCreateCmd.MarkFlagRequired("region")
	backupCreateCmd.MarkFlagRequired("dbaas-id")
	backupCreateCmd.MarkFlagRequired("database-name")

	backupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	backupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	backupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	backupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	backupListCmd.Flags().Int("offset", 0, "Number of results to skip")

	backupGetCmd.ValidArgsFunction = completeDatabaseBackupID
	backupDeleteCmd.ValidArgsFunction = completeDatabaseBackupID
}

// databaseBackupRef returns a Ref for a specific database backup.
type databaseBackupGetView struct {
	ID, URI, Name, Region, Status, CreatedAt, CreatedBy, Tags string
}

func databaseBackupRef(projectID, backupID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/backups/" + backupID)
}

func completeDatabaseBackupID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("database-backup", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromDatabase().Backups().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, backup := range list.Items() {
			id := backup.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, backup.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage database backups",
	Long:  `Perform CRUD operations on database backups in Aruba Cloud.`,
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new database backup",
	Long: `Create a backup of a database inside a DBaaS instance.

Provide the DBaaS instance ID (--dbaas-id) and the database name (--database-name).
The backup will be stored and can be restored later.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud database backup create \
    --name my-backup --region ITBG-Bergamo --zone ITBG-1 \
    --dbaas-id <dbaas-id> \
    --database-name myapp_db`,
	Args: cobra.NoArgs,
	RunE: DatabaseDBaaSBackupCreateRun,
}

var backupGetCmd = &cobra.Command{
	Use:   "get [backup-id]",
	Short: "Get backup details",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSBackupGetRun,
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all database backups",
	Args:  cobra.NoArgs,
	RunE:  DatabaseDBaaSBackupListRun,
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a backup",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSBackupDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// DatabaseDBaaSBackupCreateArgs holds the typed arguments for creating a database backup.
type DatabaseDBaaSBackupCreateArgs struct {
	ProjectID     string
	Name          string
	Region        aruba.Region
	Zone          string
	DBaaSID       string
	DatabaseName  string
	BillingPeriod aruba.BillingPeriod
	Tags          []string
}

// DatabaseDBaaSBackupGetArgs holds the typed arguments for getting a database backup.
type DatabaseDBaaSBackupGetArgs struct {
	ProjectID string
	BackupID  string
}

// DatabaseDBaaSBackupListArgs holds the typed arguments for listing database backups.
type DatabaseDBaaSBackupListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// DatabaseDBaaSBackupDeleteArgs holds the typed arguments for deleting a database backup.
type DatabaseDBaaSBackupDeleteArgs struct {
	ProjectID   string
	BackupID    string
	DryRun      bool
	SkipConfirm bool
}

// =============================================================================
// Constructors
// =============================================================================

// NewDatabaseDBaaSBackupCreateArgsFromCobraCommand parses and validates args for create.
func NewDatabaseDBaaSBackupCreateArgsFromCobraCommand(cmd *cobra.Command) (*DatabaseDBaaSBackupCreateArgs, error) {
	args := &DatabaseDBaaSBackupCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSBackupGetArgsFromCobraCommand parses and validates args for get.
func NewDatabaseDBaaSBackupGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSBackupGetArgs, error) {
	args := &DatabaseDBaaSBackupGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSBackupListArgsFromCobraCommand parses and validates args for list.
func NewDatabaseDBaaSBackupListArgsFromCobraCommand(cmd *cobra.Command) (*DatabaseDBaaSBackupListArgs, error) {
	args := &DatabaseDBaaSBackupListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSBackupDeleteArgsFromCobraCommand parses and validates args for delete.
func NewDatabaseDBaaSBackupDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSBackupDeleteArgs, error) {
	args := &DatabaseDBaaSBackupDeleteArgs{}
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

// ParseFromCobraCommand reads Cobra flags into the create args struct.
func (a *DatabaseDBaaSBackupCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if a.Zone, err = cmd.Flags().GetString("zone"); err != nil {
		errs = append(errs, err)
	}
	if a.DBaaSID, err = cmd.Flags().GetString("dbaas-id"); err != nil {
		errs = append(errs, err)
	}
	if a.DatabaseName, err = cmd.Flags().GetString("database-name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *DatabaseDBaaSBackupGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *DatabaseDBaaSBackupListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *DatabaseDBaaSBackupDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
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
func (a *DatabaseDBaaSBackupCreateArgs) Validate() error {
	var errs []error

	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}
	if !slices.Contains(validRegions, a.Region) {
		errs = append(errs, fmt.Errorf("--region %q: must be one of %v", a.Region, validRegions))
	}
	if !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	if a.DBaaSID == "" {
		errs = append(errs, errors.New("--dbaas-id is required"))
	}
	if a.DatabaseName == "" {
		errs = append(errs, errors.New("--database-name is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *DatabaseDBaaSBackupGetArgs) Validate() error {
	if a.BackupID == "" {
		return errors.New("backup ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *DatabaseDBaaSBackupListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the delete args for correctness.
func (a *DatabaseDBaaSBackupDeleteArgs) Validate() error {
	if a.BackupID == "" {
		return errors.New("backup ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// DatabaseDBaaSBackupCreate creates a new database backup.
func DatabaseDBaaSBackupCreate(ctx context.Context, client aruba.Client, args DatabaseDBaaSBackupCreateArgs) error {
	bkp := aruba.NewDBaaSBackup().
		InProject(projectRef(args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		FromDBaaS(dbaasRef(args.ProjectID, args.DBaaSID)).
		FromDatabase(databaseRef(args.ProjectID, args.DBaaSID, args.DatabaseName)).
		BilledBy(args.BillingPeriod).
		RetaggedAs(args.Tags...)

	if args.Zone != "" {
		bkp.InZone(aruba.Zone(args.Zone))
	}

	created, err := client.FromDatabase().Backups().Create(ctx, bkp)
	if err != nil {
		return fmt.Errorf("creating backup: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		statusVal := ""
		if raw.Status.State != nil {
			statusVal = string(*raw.Status.State)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, regionVal, statusVal}})
	} else {
		fmt.Println(msgCreatedAsync("Backup", args.Name))
	}
	return nil
}

// DatabaseDBaaSBackupGet retrieves and displays a database backup's details.
func DatabaseDBaaSBackupGet(ctx context.Context, client aruba.Client, args DatabaseDBaaSBackupGetArgs) error {
	backup, err := client.FromDatabase().Backups().Get(ctx, databaseBackupRef(args.ProjectID, args.BackupID))
	if err != nil {
		return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
	}

	if backup == nil || backup.Raw() == nil {
		fmt.Println("Backup not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(backup, nil, nil)
		return nil
	}
	raw := backup.Raw()
	view := databaseBackupGetView{Tags: "[]"}
	if raw.Metadata.ID != nil {
		view.ID = *raw.Metadata.ID
	}
	if raw.Metadata.URI != nil {
		view.URI = *raw.Metadata.URI
	}
	if raw.Metadata.Name != nil {
		view.Name = *raw.Metadata.Name
	}
	if raw.Metadata.LocationResponse != nil {
		view.Region = string(raw.Metadata.LocationResponse.Value)
	}
	if raw.Status.State != nil {
		view.Status = string(*raw.Status.State)
	}
	if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
		view.CreatedAt = raw.Metadata.CreationDate.Format(DateLayout)
	}
	if raw.Metadata.CreatedBy != nil {
		view.CreatedBy = *raw.Metadata.CreatedBy
	}
	if len(raw.Metadata.Tags) > 0 {
		view.Tags = fmt.Sprintf("%v", raw.Metadata.Tags)
	}
	return renderGet(databaseBackupGetTmpl, view)
}

// DatabaseDBaaSBackupList lists all database backups in a project.
func DatabaseDBaaSBackupList(ctx context.Context, client aruba.Client, args DatabaseDBaaSBackupListArgs) error {
	list, err := client.FromDatabase().Backups().List(ctx, aruba.URI("/projects/"+args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing backups: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No backups found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.DBaaSBackup]{
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(b *aruba.DBaaSBackup) string {
			if r := b.Raw(); r != nil && r.Metadata.Name != nil {
				return *r.Metadata.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ID", Width: 30}, Value: func(b *aruba.DBaaSBackup) string {
			if r := b.Raw(); r != nil && r.Metadata.ID != nil {
				return *r.Metadata.ID
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "REGION", Width: 20}, Value: func(b *aruba.DBaaSBackup) string {
			if r := b.Raw(); r != nil && r.Metadata.LocationResponse != nil {
				return string(r.Metadata.LocationResponse.Value)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(b *aruba.DBaaSBackup) string {
			if r := b.Raw(); r != nil && r.Status.State != nil {
				return string(*r.Status.State)
			}
			return ""
		}},
	}, list.Items(), func(b *aruba.DBaaSBackup) bool { return b.Raw() != nil })
	return nil
}

// DatabaseDBaaSBackupDelete deletes a database backup.
func DatabaseDBaaSBackupDelete(ctx context.Context, client aruba.Client, args DatabaseDBaaSBackupDeleteArgs) error {
	err := client.FromDatabase().Backups().Delete(ctx, databaseBackupRef(args.ProjectID, args.BackupID))
	if err != nil {
		return fmt.Errorf("deleting backup: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Backup", args.BackupID))
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// DatabaseDBaaSBackupCreateRun is the Cobra RunE handler for backup create.
func DatabaseDBaaSBackupCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewDatabaseDBaaSBackupCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSBackupCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSBackupGetRun is the Cobra RunE handler for backup get.
func DatabaseDBaaSBackupGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSBackupGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSBackupGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSBackupListRun is the Cobra RunE handler for backup list.
func DatabaseDBaaSBackupListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewDatabaseDBaaSBackupListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSBackupList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSBackupDeleteRun is the Cobra RunE handler for backup delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func DatabaseDBaaSBackupDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSBackupDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("backup", args.BackupID)
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
		if _, err := client.FromDatabase().Backups().Get(ctx, databaseBackupRef(args.ProjectID, args.BackupID)); err != nil {
			return fmt.Errorf("dry-run: database backup not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("database backup", args.BackupID))
		return nil
	}

	if err := DatabaseDBaaSBackupDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
