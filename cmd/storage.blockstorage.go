package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

type blockStorageGetView struct {
	ID, URI, Name, Size, Type, Zone, Region, Bootable, Status, CreatedAt, CreatedBy, Tags string
}

const blockStorageGetTmpl = `
Block Storage Details:
======================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Size (GB):       {{.Size}}
Type:            {{.Type}}
Zone:            {{.Zone}}
Region:          {{.Region}}
Bootable:        {{.Bootable}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

func volumeRef(projectID, volumeID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/blockStorages/" + volumeID)
}

func init() {
	storageCmd.AddCommand(blockstorageCmd)
	blockstorageCmd.AddCommand(blockstorageCreateCmd)
	blockstorageCmd.AddCommand(blockstorageGetCmd)
	blockstorageCmd.AddCommand(blockstorageUpdateCmd)
	blockstorageCmd.AddCommand(blockstorageDeleteCmd)
	blockstorageCmd.AddCommand(blockstorageListCmd)

	blockstorageCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageCreateCmd.Flags().String("name", "", "Name for the block storage (required)")
	blockstorageCreateCmd.Flags().String("region", "ITBG-Bergamo", "Region code (required)")
	blockstorageCreateCmd.Flags().String("zone", "", "Zone/datacenter (optional, only for zonal block storage)")
	blockstorageCreateCmd.Flags().Int("size", 0, "Size in GB (required)")
	blockstorageCreateCmd.Flags().String("type", "Standard", "Type: Standard or Performance")
	blockstorageCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
	blockstorageCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	blockstorageCreateCmd.Flags().String("snapshot-id", "", "ID of the snapshot to use (optional)")
	blockstorageCreateCmd.Flags().Bool("set-bootable", false, "Set block storage as bootable (optional)")
	blockstorageCreateCmd.Flags().String("image", "", "Image string to use for the block storage (optional)")
	blockstorageCreateCmd.MarkFlagRequired("name")
	blockstorageCreateCmd.MarkFlagRequired("region")
	blockstorageCreateCmd.MarkFlagRequired("size")
	blockstorageCreateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")

	blockstorageGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	blockstorageUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageUpdateCmd.Flags().String("name", "", "New name for the block storage")
	blockstorageUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	blockstorageUpdateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")

	blockstorageDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	blockstorageDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	blockstorageListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	blockstorageListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	blockstorageGetCmd.ValidArgsFunction = completeBlockStorageID
	blockstorageUpdateCmd.ValidArgsFunction = completeBlockStorageID
	blockstorageDeleteCmd.ValidArgsFunction = completeBlockStorageID
}

func completeBlockStorageID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("blockstorage", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Volumes().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vol := range list.Items() {
			id := vol.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, vol.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var blockstorageCmd = &cobra.Command{
	Use:   "blockstorage",
	Short: "Manage block storage",
	Long:  `Perform CRUD operations on block storage in Aruba Cloud.`,
}

var blockstorageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new block storage",
	Long: `Create a new block storage volume in the specified region.

Volume size is specified in GB with --size. Type options: Standard or Performance.
The volume can be initialised from a snapshot with --snapshot-id, or from an
image with --image. Pass --set-bootable to mark the volume as a boot disk.

Billing period: Hour (default), Month, or Year.`,
	Example: `  # Create an empty 50 GB Standard volume
  acloud storage blockstorage create --name my-volume --size 50 --region ITBG-Bergamo

  # Create a Performance volume from a snapshot
  acloud storage blockstorage create --name fast-vol --size 100 --region ITBG-Bergamo \
    --type Performance \
    --snapshot-id <snap-id>`,
	Args: cobra.NoArgs,
	RunE: StorageBlockStorageCreateRun,
}

var blockstorageGetCmd = &cobra.Command{
	Use:   "get [volume-id]",
	Short: "Get block storage details",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageBlockStorageGetRun,
}

var blockstorageUpdateCmd = &cobra.Command{
	Use:   "update [volume-id]",
	Short: "Update block storage (name and/or tags)",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageBlockStorageUpdateRun,
}

var blockstorageDeleteCmd = &cobra.Command{
	Use:   "delete [volume-id]",
	Short: "Delete block storage",
	Args:  cobra.ExactArgs(1),
	RunE:  StorageBlockStorageDeleteRun,
}

var blockstorageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all block storage",
	Args:  cobra.NoArgs,
	RunE:  StorageBlockStorageListRun,
}

// =============================================================================
// Args structs
// =============================================================================

// StorageBlockStorageCreateArgs holds the typed arguments for creating a block storage.
type StorageBlockStorageCreateArgs struct {
	ProjectID     string
	Name          string
	Region        aruba.Region
	Zone          aruba.Zone
	SizeGB        int
	VolumeType    aruba.BlockStorageType
	BillingPeriod aruba.BillingPeriod
	Tags          []string
	SnapshotID    string
	SetBootable   bool
	Image         string
	Wait          bool
}

// StorageBlockStorageGetArgs holds the typed arguments for getting a block storage.
type StorageBlockStorageGetArgs struct {
	ProjectID string
	ID        string
}

// StorageBlockStorageUpdateArgs holds the typed arguments for updating a block storage.
type StorageBlockStorageUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
	Wait        bool
}

// StorageBlockStorageDeleteArgs holds the typed arguments for deleting a block storage.
type StorageBlockStorageDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// StorageBlockStorageListArgs holds the typed arguments for listing block storage.
type StorageBlockStorageListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewStorageBlockStorageCreateArgsFromCobraCommand parses and validates args for create.
func NewStorageBlockStorageCreateArgsFromCobraCommand(cmd *cobra.Command) (*StorageBlockStorageCreateArgs, error) {
	args := &StorageBlockStorageCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBlockStorageGetArgsFromCobraCommand parses and validates args for get.
func NewStorageBlockStorageGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBlockStorageGetArgs, error) {
	args := &StorageBlockStorageGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBlockStorageUpdateArgsFromCobraCommand parses and validates args for update.
func NewStorageBlockStorageUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBlockStorageUpdateArgs, error) {
	args := &StorageBlockStorageUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBlockStorageDeleteArgsFromCobraCommand parses and validates args for delete.
func NewStorageBlockStorageDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*StorageBlockStorageDeleteArgs, error) {
	args := &StorageBlockStorageDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewStorageBlockStorageListArgsFromCobraCommand parses and validates args for list.
func NewStorageBlockStorageListArgsFromCobraCommand(cmd *cobra.Command) (*StorageBlockStorageListArgs, error) {
	args := &StorageBlockStorageListArgs{}
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
func (a *StorageBlockStorageCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if s, err := cmd.Flags().GetString("zone"); err == nil {
		a.Zone = aruba.Zone(s)
	} else {
		errs = append(errs, err)
	}
	if a.SizeGB, err = cmd.Flags().GetInt("size"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("type"); err == nil {
		a.VolumeType = aruba.BlockStorageType(s)
	} else {
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
	if a.SnapshotID, err = cmd.Flags().GetString("snapshot-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SetBootable, err = cmd.Flags().GetBool("set-bootable"); err != nil {
		errs = append(errs, err)
	}
	if a.Image, err = cmd.Flags().GetString("image"); err != nil {
		errs = append(errs, err)
	}
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *StorageBlockStorageGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *StorageBlockStorageUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *StorageBlockStorageDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *StorageBlockStorageListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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

// Validate checks the create args for semantic validity.
func (a *StorageBlockStorageCreateArgs) Validate() error {
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
	if a.SizeGB <= 0 {
		errs = append(errs, errors.New("--size must be greater than 0"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for semantic validity.
func (a *StorageBlockStorageGetArgs) Validate() error {
	return nil
}

// Validate checks the update args for semantic validity.
func (a *StorageBlockStorageUpdateArgs) Validate() error {
	if a.Name == "" && !a.TagsChanged {
		return errors.New("at least one of --name or --tags must be provided")
	}
	return nil
}

// Validate checks the delete args for semantic validity.
func (a *StorageBlockStorageDeleteArgs) Validate() error {
	return nil
}

// Validate checks the list args for semantic validity.
func (a *StorageBlockStorageListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// StorageBlockStorageCreate creates a new block storage volume.
func StorageBlockStorageCreate(ctx context.Context, client aruba.Client, args StorageBlockStorageCreateArgs) error {
	vol := aruba.NewBlockStorage().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		SizedGB(args.SizeGB).
		OfType(args.VolumeType).
		BilledBy(args.BillingPeriod).
		RetaggedAs(args.Tags...)

	if args.Zone != "" {
		vol.InZone(args.Zone)
	}
	if args.SnapshotID != "" {
		vol.FromSnapshot(snapshotRef(args.ProjectID, args.SnapshotID))
	}
	if args.SetBootable {
		vol.AsBootable()
	}
	if args.Image != "" {
		vol.FromImage(args.Image)
	}

	created, err := client.FromStorage().Volumes().Create(ctx, vol)
	if err != nil {
		return fmt.Errorf("creating block storage: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "SIZE(GB)", Width: 12},
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
		sizeVal := fmt.Sprintf("%d", raw.Properties.SizeGB)
		typeVal := string(raw.Properties.Type)
		statusVal := ""
		if raw.Status.State != nil {
			statusVal = string(*raw.Status.State)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, sizeVal, typeVal, statusVal}})
		if args.Wait && id != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromStorage().Volumes().Get(ctx, volumeRef(args.ProjectID, id))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "Block storage", args.Name); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgCreatedAsync("Block storage", args.Name))
	}
	return nil
}

// StorageBlockStorageGet fetches and displays a block storage volume.
func StorageBlockStorageGet(ctx context.Context, client aruba.Client, args StorageBlockStorageGetArgs) error {
	vol, err := client.FromStorage().Volumes().Get(ctx, volumeRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting block storage: %w", apiErrFromV2(err))
	}

	if vol == nil || vol.Raw() == nil {
		fmt.Println("Block storage not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(vol, nil, nil)
		return nil
	}
	raw := vol.Raw()
	view := blockStorageGetView{
		Size: fmt.Sprintf("%d", raw.Properties.SizeGB),
		Type: string(raw.Properties.Type),
		Zone: string(raw.Properties.Zone),
		Tags: "[]",
	}
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
	if raw.Properties.Bootable != nil {
		view.Bootable = fmt.Sprintf("%t", *raw.Properties.Bootable)
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
	return renderGet(blockStorageGetTmpl, view)
}

// StorageBlockStorageUpdate mutates and persists a block storage volume.
func StorageBlockStorageUpdate(ctx context.Context, client aruba.Client, args StorageBlockStorageUpdateArgs) error {
	vol, err := client.FromStorage().Volumes().Get(ctx, volumeRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting block storage: %w", apiErrFromV2(err))
	}
	if vol == nil || vol.Raw() == nil {
		return fmt.Errorf("block storage not found")
	}

	if args.Name != "" {
		vol.Named(args.Name)
	}
	if args.TagsChanged {
		vol.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromStorage().Volumes().Update(ctx, vol)
	if err != nil {
		return fmt.Errorf("updating block storage: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Block storage", args.ID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
		fmt.Printf("Size (GB):       %d\n", raw.Properties.SizeGB)
		fmt.Printf("Type:            %s\n", string(raw.Properties.Type))
		if args.Wait && args.ID != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromStorage().Volumes().Get(ctx, volumeRef(args.ProjectID, args.ID))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "Block storage", args.ID); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgUpdatedAsync("Block storage", args.ID))
	}
	return nil
}

// StorageBlockStorageDelete removes a block storage volume.
func StorageBlockStorageDelete(ctx context.Context, client aruba.Client, args StorageBlockStorageDeleteArgs) error {
	err := client.FromStorage().Volumes().Delete(ctx, volumeRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting block storage: %w", apiErrFromV2(err))
	}

	fmt.Println(msgDeleted("Block storage", args.ID))
	return nil
}

// StorageBlockStorageList lists all block storage volumes in a project.
func StorageBlockStorageList(ctx context.Context, client aruba.Client, args StorageBlockStorageListArgs) error {
	list, err := client.FromStorage().Volumes().List(ctx, aruba.URI("/projects/"+args.ProjectID))
	if err != nil {
		return fmt.Errorf("listing block storage: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No block storage found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.BlockStorage]{
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil && r.Metadata.Name != nil {
				return *r.Metadata.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ID", Width: 26}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil && r.Metadata.ID != nil {
				return *r.Metadata.ID
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "SIZE(GB)", Width: 12}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil {
				return fmt.Sprintf("%d", r.Properties.SizeGB)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "REGION", Width: 15}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil && r.Metadata.LocationResponse != nil {
				return string(r.Metadata.LocationResponse.Value)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ZONE", Width: 15}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil {
				return string(r.Properties.Zone)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "TYPE", Width: 15}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil {
				return string(r.Properties.Type)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(v *aruba.BlockStorage) string {
			if r := v.Raw(); r != nil && r.Status.State != nil {
				return string(*r.Status.State)
			}
			return ""
		}},
	}, list.Items(), func(v *aruba.BlockStorage) bool { return v.Raw() != nil })
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// StorageBlockStorageCreateRun is the RunE wiring for blockstorage create.
func StorageBlockStorageCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewStorageBlockStorageCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBlockStorageCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBlockStorageGetRun is the RunE wiring for blockstorage get.
func StorageBlockStorageGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageBlockStorageGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBlockStorageGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBlockStorageUpdateRun is the RunE wiring for blockstorage update.
func StorageBlockStorageUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageBlockStorageUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBlockStorageUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBlockStorageDeleteRun is the RunE wiring for blockstorage delete.
func StorageBlockStorageDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewStorageBlockStorageDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("block storage", args.ID)
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
		_, err = client.FromStorage().Volumes().Get(ctx, volumeRef(args.ProjectID, args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: block storage not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("block storage", args.ID))
		return nil
	}

	if err := StorageBlockStorageDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// StorageBlockStorageListRun is the RunE wiring for blockstorage list.
func StorageBlockStorageListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewStorageBlockStorageListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := StorageBlockStorageList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
