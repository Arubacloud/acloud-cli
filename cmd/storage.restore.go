package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func restoreRef(projectID, backupID, restoreID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID + "/restores/" + restoreID)
}

func init() {
	storageCmd.AddCommand(storageRestoreCmd)
	storageRestoreCmd.AddCommand(storageRestoreListCmd)
	storageRestoreCmd.AddCommand(storageRestoreGetCmd)
	storageRestoreCmd.AddCommand(storageRestoreUpdateCmd)
	storageRestoreCmd.AddCommand(storageRestoreDeleteCmd)

	// Create flags live on the parent (storageRestoreCmd is the create command).
	storageRestoreCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreCmd.Flags().String("name", "", "Name for the restore operation (required)")
	storageRestoreCmd.Flags().String("region", string(aruba.RegionITBGBergamo), "Region code (default: ITBG-Bergamo)")
	storageRestoreCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	storageRestoreCmd.MarkFlagRequired("name")

	storageRestoreListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	storageRestoreListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	storageRestoreGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	storageRestoreUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreUpdateCmd.Flags().String("name", "", "New name for the restore operation")
	storageRestoreUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	storageRestoreDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	storageRestoreDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	storageRestoreGetCmd.ValidArgsFunction = completeRestoreID
	storageRestoreUpdateCmd.ValidArgsFunction = completeRestoreID
	storageRestoreDeleteCmd.ValidArgsFunction = completeRestoreID
	storageRestoreListCmd.ValidArgsFunction = completeBackupID
}

func completeRestoreID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeBackupID(cmd, args, toComplete)
	}
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	backupID := args[0]
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("restore", projectID, backupID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Restores().List(ctx, backupRef(projectID, backupID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, r := range list.Items() {
			id := r.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, r.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var storageRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id] [volume-id]",
	Short: "Restore a block storage volume from a backup",
	Long: `Create a restore operation to copy backup data back to a block storage volume.

Both the backup and the target volume must already exist.`,
	Example: `  acloud storage restore <backup-id> <volume-id> --name my-restore --region ITBG-Bergamo`,
	Args:    cobra.ExactArgs(2),
	RunE:    StorageRestoreCreateRun,
}

var storageRestoreListCmd = &cobra.Command{
	Use:   "list [backup-id]",
	Short: "List restore operations for a backup",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageRestoreListRun,
}

var storageRestoreGetCmd = &cobra.Command{
	Use:   "get [backup-id] [restore-id]",
	Short: "Get restore operation details",
	Args:  cobra.ExactArgs(2),
	RunE:  StorageRestoreGetRun,
}

var storageRestoreUpdateCmd = &cobra.Command{
	Use:   "update [backup-id] [restore-id]",
	Short: "Update a restore operation (name and/or tags)",
	Args:  cobra.ExactArgs(2),
	RunE:  StorageRestoreUpdateRun,
}

var storageRestoreDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id] [restore-id]",
	Short: "Delete a restore operation",
	Args:  cobra.ExactArgs(2),
	RunE:  StorageRestoreDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// StorageRestoreCreateArgs holds the typed arguments for creating a restore operation.
type StorageRestoreCreateArgs struct {
	ProjectID string
	Name      string
	Region    aruba.Region
	BackupID  string
	VolumeID  string
	Tags      []string
}

// StorageRestoreGetArgs holds the typed arguments for getting a restore operation.
type StorageRestoreGetArgs struct {
	ProjectID string
	BackupID  string
	ID        string
}

// StorageRestoreUpdateArgs holds the typed arguments for updating a restore operation.
type StorageRestoreUpdateArgs struct {
	ProjectID   string
	BackupID    string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
}

// StorageRestoreDeleteArgs holds the typed arguments for deleting a restore operation.
type StorageRestoreDeleteArgs struct {
	ProjectID   string
	BackupID    string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// StorageRestoreListArgs holds the typed arguments for listing restore operations.
type StorageRestoreListArgs struct {
	ProjectID string
	BackupID  string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewStorageRestoreCreateArgsFromCobraCommand parses and validates args for create.
func NewStorageRestoreCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageRestoreCreateArgs, error) {
	args := &StorageRestoreCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageRestoreGetArgsFromCobraCommand parses and validates args for get.
func NewStorageRestoreGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageRestoreGetArgs, error) {
	args := &StorageRestoreGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageRestoreUpdateArgsFromCobraCommand parses and validates args for update.
func NewStorageRestoreUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageRestoreUpdateArgs, error) {
	args := &StorageRestoreUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageRestoreDeleteArgsFromCobraCommand parses and validates args for delete.
func NewStorageRestoreDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageRestoreDeleteArgs, error) {
	args := &StorageRestoreDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageRestoreListArgsFromCobraCommand parses and validates args for list.
func NewStorageRestoreListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageRestoreListArgs, error) {
	args := &StorageRestoreListArgs{}
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
// cobraArgs[0] = backupID, cobraArgs[1] = volumeID (positional, matching original CLI surface).
func (a *StorageRestoreCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.VolumeID = cobraArgs[1]
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *StorageRestoreGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.ID = cobraArgs[1]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *StorageRestoreUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.ID = cobraArgs[1]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *StorageRestoreDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.ID = cobraArgs[1]
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
func (a *StorageRestoreListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.BackupID = cobraArgs[0]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *StorageRestoreCreateArgs) Validate() error {
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
	if a.BackupID == "" {
		errs = append(errs, errors.New("--backup-id is required"))
	}
	if a.VolumeID == "" {
		errs = append(errs, errors.New("--volume-id is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *StorageRestoreGetArgs) Validate() error {
	var errs []error
	if a.BackupID == "" {
		errs = append(errs, errors.New("backup ID is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("restore ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *StorageRestoreUpdateArgs) Validate() error {
	if a.Name == "" && !a.TagsChanged {
		return errors.New("at least one of --name or --tags must be provided")
	}
	return nil
}

// Validate checks the delete args for correctness.
func (a *StorageRestoreDeleteArgs) Validate() error {
	var errs []error
	if a.BackupID == "" {
		errs = append(errs, errors.New("backup ID is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("restore ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *StorageRestoreListArgs) Validate() error {
	if a.BackupID == "" {
		return errors.New("--backup-id is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// StorageRestoreCreate creates a restore operation.
// It performs dual cross-family pre-validation: fetches the backup and volume first.
func StorageRestoreCreate(ctx context.Context, client aruba.Client, args StorageRestoreCreateArgs) error {
	bk, err := client.FromStorage().Backups().Get(ctx, backupRef(args.ProjectID, args.BackupID))
	if err != nil {
		return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
	}

	target, err := client.FromStorage().Volumes().Get(ctx, volumeRef(args.ProjectID, args.VolumeID))
	if err != nil {
		return fmt.Errorf("getting volume: %w", apiErrFromV2(err))
	}

	rs := aruba.NewStorageRestore().
		FromBackup(bk).
		Named(args.Name).
		InRegion(args.Region).
		ToVolume(target).
		RetaggedAs(args.Tags...)

	created, err := client.FromStorage().Restores().Create(ctx, rs)
	if err != nil {
		return fmt.Errorf("creating restore: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
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
		statusVal := ""
		if raw.Status.State != nil {
			statusVal = string(*raw.Status.State)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, statusVal}})
	} else {
		fmt.Println(msgCreatedAsync("Restore operation", args.Name))
	}
	return nil
}

// StorageRestoreGet retrieves restore operation details.
func StorageRestoreGet(ctx context.Context, client aruba.Client, args StorageRestoreGetArgs) error {
	restore, err := client.FromStorage().Restores().Get(ctx, restoreRef(args.ProjectID, args.BackupID, args.ID))
	if err != nil {
		return fmt.Errorf("getting restore: %w", apiErrFromV2(err))
	}

	if restore != nil && restore.Raw() != nil {
		raw := restore.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(restore, nil, nil)
			return nil
		}

		fmt.Println("\nRestore Operation Details:")
		fmt.Println("==========================")
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Properties.Destination.URI != "" {
			fmt.Printf("Target Volume:   %s\n", raw.Properties.Destination.URI)
		}
		if raw.Metadata.LocationResponse != nil {
			fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
		}
		if raw.Status.State != nil {
			fmt.Printf("Status:          %s\n", string(*raw.Status.State))
		}
		if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
			fmt.Printf("Creation Date:   %s\n", raw.Metadata.CreationDate.Format(DateLayout))
		}
		if raw.Metadata.CreatedBy != nil {
			fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		} else {
			fmt.Printf("Tags:            []\n")
		}
		fmt.Println()
	} else {
		fmt.Println("Restore operation not found")
	}
	return nil
}

// StorageRestoreUpdate updates a restore operation's name and/or tags.
func StorageRestoreUpdate(ctx context.Context, client aruba.Client, args StorageRestoreUpdateArgs) error {
	restore, err := client.FromStorage().Restores().Get(ctx, restoreRef(args.ProjectID, args.BackupID, args.ID))
	if err != nil {
		return fmt.Errorf("getting restore: %w", apiErrFromV2(err))
	}
	if restore == nil || restore.Raw() == nil {
		return fmt.Errorf("restore operation not found")
	}

	if args.Name != "" {
		restore.Named(args.Name)
	}
	if args.TagsChanged {
		restore.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromStorage().Restores().Update(ctx, restore)
	if err != nil {
		return fmt.Errorf("updating restore: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Restore operation", args.ID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgUpdatedAsync("Restore operation", args.ID))
	}
	return nil
}

// StorageRestoreDelete deletes a restore operation.
func StorageRestoreDelete(ctx context.Context, client aruba.Client, args StorageRestoreDeleteArgs) error {
	if args.DryRun {
		_, err := client.FromStorage().Restores().Get(ctx, restoreRef(args.ProjectID, args.BackupID, args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: restore operation not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("restore operation", args.ID))
		return nil
	}

	if err := client.FromStorage().Restores().Delete(ctx, restoreRef(args.ProjectID, args.BackupID, args.ID)); err != nil {
		return fmt.Errorf("deleting restore: %w", apiErrFromV2(err))
	}

	fmt.Println(msgDeleted("Restore operation", args.ID))
	return nil
}

// StorageRestoreList lists restore operations scoped to a specific backup.
func StorageRestoreList(ctx context.Context, client aruba.Client, args StorageRestoreListArgs) error {
	list, err := client.FromStorage().Restores().List(ctx, backupRef(args.ProjectID, args.BackupID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing restores: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, r := range list.Items() {
			raw := r.Raw()
			if raw == nil {
				continue
			}
			name := ""
			if raw.Metadata.Name != nil {
				name = *raw.Metadata.Name
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, status})
		}

		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No restores found for this backup")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// StorageRestoreCreateRun is the RunE wiring for restore create.
func StorageRestoreCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageRestoreCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageRestoreCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageRestoreGetRun is the RunE wiring for restore get.
func StorageRestoreGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageRestoreGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageRestoreGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageRestoreUpdateRun is the RunE wiring for restore update.
func StorageRestoreUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageRestoreUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageRestoreUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageRestoreDeleteRun is the RunE wiring for restore delete.
func StorageRestoreDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewStorageRestoreDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("restore operation", a.ID)
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

	if err := StorageRestoreDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageRestoreListRun is the RunE wiring for restore list.
func StorageRestoreListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageRestoreListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageRestoreList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
