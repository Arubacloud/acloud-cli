package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func snapshotRef(projectID, snapshotID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/snapshots/" + snapshotID)
}

func init() {
	storageCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotGetCmd)
	snapshotCmd.AddCommand(snapshotUpdateCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	snapshotCmd.AddCommand(snapshotListCmd)

	snapshotCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotCreateCmd.Flags().String("name", "", "Name for the snapshot (required)")
	snapshotCreateCmd.Flags().String("region", "", "Region code (required)")
	snapshotCreateCmd.Flags().String("volume-id", "", "Source volume ID (required)")
	snapshotCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	snapshotCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	snapshotCreateCmd.MarkFlagRequired("name")
	snapshotCreateCmd.MarkFlagRequired("region")
	snapshotCreateCmd.MarkFlagRequired("volume-id")

	snapshotGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	snapshotUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotUpdateCmd.Flags().String("name", "", "New name for the snapshot")
	snapshotUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	snapshotDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	snapshotDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	snapshotListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotListCmd.Flags().String("volume-id", "", "Block storage volume ID to filter by (optional)")
	snapshotListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	snapshotListCmd.Flags().Int32("offset", 0, "Number of results to skip")
	snapshotListCmd.MarkFlagRequired("volume-id")

	snapshotGetCmd.ValidArgsFunction = completeSnapshotID
	snapshotUpdateCmd.ValidArgsFunction = completeSnapshotID
	snapshotDeleteCmd.ValidArgsFunction = completeSnapshotID
}

func completeSnapshotID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("snapshot", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Snapshots().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, snap := range list.Items() {
			id := snap.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, snap.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots",
	Long:  `Perform CRUD operations on snapshots in Aruba Cloud.`,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new snapshot",
	Long: `Create a point-in-time snapshot of an existing block storage volume.

The volume ID is required. Snapshots can later be used to create new volumes
or restore data with 'acloud storage blockstorage create --snapshot-id'.`,
	Example: `  acloud storage snapshot create --name my-snap --region ITBG-Bergamo \
    --volume-id <vol-id>`,
	Args: cobra.NoArgs,
	RunE: StorageSnapshotCreateRun,
}

var snapshotGetCmd = &cobra.Command{
	Use:   "get [snapshot-id]",
	Short: "Get snapshot details",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageSnapshotGetRun,
}

var snapshotUpdateCmd = &cobra.Command{
	Use:   "update [snapshot-id]",
	Short: "Update a snapshot (name and/or tags only)",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageSnapshotUpdateRun,
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete [snapshot-id]",
	Short: "Delete a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageSnapshotDeleteRun,
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots for a block storage volume",
	Args:  cobra.NoArgs,
	RunE:  StorageSnapshotListRun,
}

// =============================================================================
// Args structs
// =============================================================================

// StorageSnapshotCreateArgs holds the typed arguments for creating a snapshot.
type StorageSnapshotCreateArgs struct {
	ProjectID string
	Name      string
	Region    aruba.Region
	VolumeID  string
	Tags      []string
}

// StorageSnapshotGetArgs holds the typed arguments for getting a snapshot.
type StorageSnapshotGetArgs struct {
	ProjectID string
	ID        string
}

// StorageSnapshotUpdateArgs holds the typed arguments for updating a snapshot.
type StorageSnapshotUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
}

// StorageSnapshotDeleteArgs holds the typed arguments for deleting a snapshot.
type StorageSnapshotDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// StorageSnapshotListArgs holds the typed arguments for listing snapshots.
type StorageSnapshotListArgs struct {
	ProjectID string
	VolumeID  string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewStorageSnapshotCreateArgsFromCobraCommand parses and validates args for create.
func NewStorageSnapshotCreateArgsFromCobraCommand(cmd *cobra.Command) (*StorageSnapshotCreateArgs, error) {
	args := &StorageSnapshotCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageSnapshotGetArgsFromCobraCommand parses and validates args for get.
func NewStorageSnapshotGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageSnapshotGetArgs, error) {
	args := &StorageSnapshotGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageSnapshotUpdateArgsFromCobraCommand parses and validates args for update.
func NewStorageSnapshotUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageSnapshotUpdateArgs, error) {
	args := &StorageSnapshotUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageSnapshotDeleteArgsFromCobraCommand parses and validates args for delete.
func NewStorageSnapshotDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageSnapshotDeleteArgs, error) {
	args := &StorageSnapshotDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageSnapshotListArgsFromCobraCommand parses and validates args for list.
func NewStorageSnapshotListArgsFromCobraCommand(cmd *cobra.Command) (*StorageSnapshotListArgs, error) {
	args := &StorageSnapshotListArgs{}
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
func (a *StorageSnapshotCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if a.VolumeID, err = cmd.Flags().GetString("volume-id"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *StorageSnapshotGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *StorageSnapshotUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
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
func (a *StorageSnapshotDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *StorageSnapshotListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.VolumeID, err = cmd.Flags().GetString("volume-id"); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *StorageSnapshotCreateArgs) Validate() error {
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
		errs = append(errs, errors.New("--volume-id is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *StorageSnapshotGetArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("snapshot ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *StorageSnapshotUpdateArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("snapshot ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *StorageSnapshotDeleteArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("snapshot ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *StorageSnapshotListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// StorageSnapshotCreate creates a new snapshot.
func StorageSnapshotCreate(ctx context.Context, client aruba.Client, args StorageSnapshotCreateArgs) error {
	snap := aruba.NewSnapshot().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		FromVolume(volumeRef(args.ProjectID, args.VolumeID)).
		RetaggedAs(args.Tags...)

	created, err := client.FromStorage().Snapshots().Create(ctx, snap)
	if err != nil {
		return fmt.Errorf("creating snapshot: %w", apiErrFromV2(err))
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
		fmt.Println(msgCreatedAsync("Snapshot", args.Name))
	}
	return nil
}

// StorageSnapshotGet retrieves snapshot details.
func StorageSnapshotGet(ctx context.Context, client aruba.Client, args StorageSnapshotGetArgs) error {
	snap, err := client.FromStorage().Snapshots().Get(ctx, snapshotRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting snapshot: %w", apiErrFromV2(err))
	}

	if snap != nil && snap.Raw() != nil {
		raw := snap.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(snap, nil, nil)
			return nil
		}

		fmt.Println("\nSnapshot Details:")
		fmt.Println("=================")
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Properties.SizeGB != nil {
			fmt.Printf("Size (GB):       %d\n", *raw.Properties.SizeGB)
		}
		if raw.Properties.Volume != nil && raw.Properties.Volume.URI != nil {
			fmt.Printf("Source Volume:   %s\n", *raw.Properties.Volume.URI)
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
		fmt.Println("Snapshot not found")
	}
	return nil
}

// StorageSnapshotUpdate updates a snapshot.
func StorageSnapshotUpdate(ctx context.Context, client aruba.Client, args StorageSnapshotUpdateArgs) error {
	snap, err := client.FromStorage().Snapshots().Get(ctx, snapshotRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting snapshot: %w", apiErrFromV2(err))
	}
	if snap == nil || snap.Raw() == nil {
		return fmt.Errorf("snapshot not found")
	}

	if args.Name != "" {
		snap.Named(args.Name)
	}
	if args.TagsChanged {
		snap.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromStorage().Snapshots().Update(ctx, snap)
	if err != nil {
		return fmt.Errorf("updating snapshot: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Snapshot", args.ID))
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
		fmt.Println(msgUpdatedAsync("Snapshot", args.ID))
	}
	return nil
}

// StorageSnapshotDelete deletes a snapshot.
func StorageSnapshotDelete(ctx context.Context, client aruba.Client, args StorageSnapshotDeleteArgs) error {
	err := client.FromStorage().Snapshots().Delete(ctx, snapshotRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting snapshot: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Snapshot", args.ID))
	return nil
}

// StorageSnapshotList lists snapshots for a block storage volume.
func StorageSnapshotList(ctx context.Context, client aruba.Client, args StorageSnapshotListArgs) error {
	list, err := client.FromStorage().Snapshots().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing snapshots: %w", apiErrFromV2(err))
	}

	headers := []TableColumn{
		{Header: "NAME", Width: 30},
		{Header: "ID", Width: 26},
		{Header: "SIZE(GB)", Width: 12},
		{Header: "STATUS", Width: 15},
	}

	var rows [][]string
	if list != nil {
		for _, snap := range list.Items() {
			if snap.VolumeURI() != volumeRef(args.ProjectID, args.VolumeID).URI() {
				continue
			}
			raw := snap.Raw()
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
			size := "0"
			if raw.Properties.SizeGB != nil {
				size = fmt.Sprintf("%d", *raw.Properties.SizeGB)
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, size, status})
		}
	}

	if len(rows) == 0 {
		fmt.Printf("No snapshots found for volume: %s\n", args.VolumeID)
		return nil
	}

	PrintOutput(list, headers, rows)
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// StorageSnapshotCreateRun is the RunE wiring for snapshot create.
func StorageSnapshotCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewStorageSnapshotCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageSnapshotCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageSnapshotGetRun is the RunE wiring for snapshot get.
func StorageSnapshotGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageSnapshotGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageSnapshotGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageSnapshotUpdateRun is the RunE wiring for snapshot update.
func StorageSnapshotUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageSnapshotUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageSnapshotUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageSnapshotDeleteRun is the RunE wiring for snapshot delete.
func StorageSnapshotDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewStorageSnapshotDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("snapshot", a.ID)
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
		_, err = client.FromStorage().Snapshots().Get(ctx, snapshotRef(a.ProjectID, a.ID))
		if err != nil {
			return fmt.Errorf("dry-run: snapshot not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("snapshot", a.ID))
		return nil
	}

	if err := StorageSnapshotDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageSnapshotListRun is the RunE wiring for snapshot list.
func StorageSnapshotListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewStorageSnapshotListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageSnapshotList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
