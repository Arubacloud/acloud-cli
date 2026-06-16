package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func backupRef(projectID, backupID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID)
}

func init() {
	storageCmd.AddCommand(storageBackupCmd)
	storageBackupCmd.AddCommand(storageBackupListCmd)
	storageBackupCmd.AddCommand(storageBackupGetCmd)
	storageBackupCmd.AddCommand(storageBackupUpdateCmd)
	storageBackupCmd.AddCommand(storageBackupDeleteCmd)

	storageBackupCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupCmd.Flags().String("name", "", "Name for the backup (required)")
	storageBackupCmd.Flags().String("region", "ITBG-Bergamo", "Region code")
	storageBackupCmd.Flags().String("type", "Full", "Backup type: Full or Incremental")
	storageBackupCmd.Flags().Int("retention-days", 0, "Number of days to retain the backup")
	storageBackupCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")
	storageBackupCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	storageBackupCmd.MarkFlagRequired("name")

	storageBackupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	storageBackupListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	storageBackupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	storageBackupUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupUpdateCmd.Flags().String("name", "", "New name for the backup")
	storageBackupUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	storageBackupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	storageBackupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	storageBackupGetCmd.ValidArgsFunction = completeBackupID
	storageBackupUpdateCmd.ValidArgsFunction = completeBackupID
	storageBackupDeleteCmd.ValidArgsFunction = completeBackupID
}

func completeBackupID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("backup", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Backups().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, bkp := range list.Items() {
			id := bkp.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, bkp.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var storageBackupCmd = &cobra.Command{
	Use:   "backup [volume-id]",
	Short: "Create a storage backup of a block storage volume",
	Long: `Create a backup of the specified block storage volume.

Backup types: Full or Incremental (default).
Use --retention-days to set how many days the backup is retained.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud storage backup <volume-id> --name my-backup
  acloud storage backup <volume-id> --name weekly-full --type Full --retention-days 30`,
	Args: cobra.ExactArgs(1),
	RunE: StorageBackupCreateRun,
}

var storageBackupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage backups",
	Args:  cobra.NoArgs,
	RunE:  StorageBackupListRun,
}

var storageBackupGetCmd = &cobra.Command{
	Use:   "get [backup-id]",
	Short: "Get storage backup details",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageBackupGetRun,
}

var storageBackupUpdateCmd = &cobra.Command{
	Use:   "update [backup-id]",
	Short: "Update a storage backup (name and/or tags)",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageBackupUpdateRun,
}

var storageBackupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a storage backup",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageBackupDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// StorageBackupCreateArgs holds the typed arguments for creating a storage backup.
type StorageBackupCreateArgs struct {
	ProjectID     string
	VolumeID      string
	Name          string
	Region        aruba.Region
	BackupType    aruba.StorageBackupType
	BillingPeriod aruba.BillingPeriod
	RetentionDays int
	Tags          []string
}

// StorageBackupGetArgs holds the typed arguments for getting a storage backup.
type StorageBackupGetArgs struct {
	ProjectID string
	ID        string
}

// StorageBackupUpdateArgs holds the typed arguments for the backup update stub.
type StorageBackupUpdateArgs struct {
	ProjectID string
	ID        string
}

// StorageBackupDeleteArgs holds the typed arguments for deleting a storage backup.
type StorageBackupDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// StorageBackupListArgs holds the typed arguments for listing storage backups.
type StorageBackupListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewStorageBackupCreateArgsFromCobraCommand parses and validates args for create.
func NewStorageBackupCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBackupCreateArgs, error) {
	args := &StorageBackupCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBackupGetArgsFromCobraCommand parses and validates args for get.
func NewStorageBackupGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBackupGetArgs, error) {
	args := &StorageBackupGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBackupUpdateArgsFromCobraCommand parses and validates args for the update stub.
func NewStorageBackupUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBackupUpdateArgs, error) {
	args := &StorageBackupUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBackupDeleteArgsFromCobraCommand parses and validates args for delete.
func NewStorageBackupDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBackupDeleteArgs, error) {
	args := &StorageBackupDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBackupListArgsFromCobraCommand parses and validates args for list.
func NewStorageBackupListArgsFromCobraCommand(cmd *cobra.Command) (*StorageBackupListArgs, error) {
	args := &StorageBackupListArgs{}
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

// ParseFromCobraCommand reads Cobra flags and positional args into the create args struct.
func (a *StorageBackupCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VolumeID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("type"); err == nil {
		a.BackupType = aruba.StorageBackupType(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.RetentionDays, err = cmd.Flags().GetInt("retention-days"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *StorageBackupGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update stub args struct.
func (a *StorageBackupUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *StorageBackupDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
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
func (a *StorageBackupListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *StorageBackupCreateArgs) Validate() error {
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
	if a.VolumeID == "" {
		errs = append(errs, errors.New("volume ID is required"))
	}
	if !slices.Contains(validStorageBackupTypes, a.BackupType) {
		errs = append(errs, fmt.Errorf("--type %q: must be one of %v", a.BackupType, validStorageBackupTypes))
	}
	if a.BillingPeriod != "" && !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *StorageBackupGetArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("backup ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the update stub args for correctness.
func (a *StorageBackupUpdateArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("backup ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *StorageBackupDeleteArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("backup ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *StorageBackupListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// StorageBackupCreate creates a new storage backup. It pre-validates the source
// volume via a cross-family Get before issuing the Create request.
func StorageBackupCreate(ctx context.Context, client aruba.Client, args StorageBackupCreateArgs) error {
	volumeURI := "/projects/" + args.ProjectID + "/providers/Aruba.Storage/blockStorages/" + args.VolumeID

	_, err := client.FromStorage().Volumes().Get(ctx, aruba.URI(volumeURI))
	if err != nil {
		return fmt.Errorf("getting volume: %w", apiErrFromV2(err))
	}

	bkp := aruba.NewStorageBackup().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		OfType(args.BackupType).
		FromVolume(aruba.URI(volumeURI)).
		RetaggedAs(args.Tags...)

	if args.RetentionDays > 0 {
		bkp.RetainedForDays(args.RetentionDays)
	}
	if args.BillingPeriod != "" {
		bkp.BilledBy(args.BillingPeriod)
	}

	created, err := client.FromStorage().Backups().Create(ctx, bkp)
	if err != nil {
		return fmt.Errorf("creating backup: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "TYPE", Width: 15},
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
		typeVal := string(raw.Properties.Type)
		statusVal := ""
		if raw.Status.State != nil {
			statusVal = string(*raw.Status.State)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, typeVal, statusVal}})
	} else {
		fmt.Println(msgCreatedAsync("Storage backup", args.Name))
	}
	return nil
}

// StorageBackupGet retrieves storage backup details.
func StorageBackupGet(ctx context.Context, client aruba.Client, args StorageBackupGetArgs) error {
	bkp, err := client.FromStorage().Backups().Get(ctx, backupRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
	}

	if bkp != nil && bkp.Raw() != nil {
		raw := bkp.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(bkp, nil, nil)
			return nil
		}

		fmt.Println("\nStorage Backup Details:")
		fmt.Println("=======================")
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		fmt.Printf("Type:            %s\n", string(raw.Properties.Type))
		if raw.Properties.Origin.URI != "" {
			fmt.Printf("Source Volume:   %s\n", raw.Properties.Origin.URI)
		}
		if raw.Properties.RetentionDays != nil {
			fmt.Printf("Retention Days:  %d\n", *raw.Properties.RetentionDays)
		}
		if raw.Properties.BillingPeriod != nil {
			fmt.Printf("Billing Period:  %s\n", string(*raw.Properties.BillingPeriod))
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
	} else {
		fmt.Println("Backup not found")
	}
	return nil
}

// StorageBackupUpdate is a stub — the API does not support backup updates.
func StorageBackupUpdate(_ context.Context, _ aruba.Client, _ StorageBackupUpdateArgs) error {
	return fmt.Errorf("storage backup update is not supported by the API")
}

// StorageBackupDelete deletes a storage backup.
func StorageBackupDelete(ctx context.Context, client aruba.Client, args StorageBackupDeleteArgs) error {
	err := client.FromStorage().Backups().Delete(ctx, backupRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting backup: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Backup", args.ID))
	return nil
}

// StorageBackupList lists all storage backups in a project.
func StorageBackupList(ctx context.Context, client aruba.Client, args StorageBackupListArgs) error {
	list, err := client.FromStorage().Backups().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing backups: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "TYPE", Width: 12},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, bkp := range list.Items() {
			raw := bkp.Raw()
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
			backupType := string(raw.Properties.Type)
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, backupType, status})
		}

		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No backups found")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// StorageBackupCreateRun is the RunE wiring for backup create.
func StorageBackupCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageBackupCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBackupCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBackupGetRun is the RunE wiring for backup get.
func StorageBackupGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageBackupGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBackupGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBackupUpdateRun is the RunE wiring for the backup update stub (not supported).
func StorageBackupUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageBackupUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	// Update is a stub — no client needed.
	return StorageBackupUpdate(context.Background(), nil, *args)
}

// StorageBackupDeleteRun is the RunE wiring for backup delete.
func StorageBackupDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewStorageBackupDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("backup", a.ID)
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

	if a.DryRun {
		_, err = client.FromStorage().Backups().Get(ctx, backupRef(a.ProjectID, a.ID))
		if err != nil {
			return fmt.Errorf("dry-run: storage backup not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("storage backup", a.ID))
		return nil
	}

	if err := StorageBackupDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBackupListRun is the RunE wiring for backup list.
func StorageBackupListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewStorageBackupListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBackupList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
