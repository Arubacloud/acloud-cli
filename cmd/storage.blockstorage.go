package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func volumeRef(projectID, volumeID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/blockstorages/" + volumeID)
}

func volumeFromRaw(b *aruba.BlockStorage) *types.BlockStorageResponse {
	if b == nil {
		return nil
	}
	return b.Raw()
}

func volumeListPayload(l *aruba.List[*aruba.BlockStorage]) any {
	if r, ok := l.Raw().(*types.Response[types.BlockStorageList]); ok && r != nil {
		return r.Data
	}
	return nil
}

func init() {
	blockstorageCreateCmd.Flags().String("snapshot-uri", "", "URI of the snapshot to use (optional)")
	blockstorageCreateCmd.Flags().Bool("set-bootable", false, "Set block storage as bootable (optional)")
	blockstorageCreateCmd.Flags().String("image", "", "Image string to use for the block storage (optional)")
	// Block storage commands
	storageCmd.AddCommand(blockstorageCmd)
	blockstorageCmd.AddCommand(blockstorageCreateCmd)
	blockstorageCmd.AddCommand(blockstorageGetCmd)
	blockstorageCmd.AddCommand(blockstorageUpdateCmd)
	blockstorageCmd.AddCommand(blockstorageDeleteCmd)
	blockstorageCmd.AddCommand(blockstorageListCmd)

	// Add flags for blockstorage commands
	blockstorageCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageCreateCmd.Flags().String("name", "", "Name for the block storage (required)")
	blockstorageCreateCmd.Flags().String("region", "ITBG-Bergamo", "Region code (required)")
	blockstorageCreateCmd.Flags().String("zone", "", "Zone/datacenter (optional, only for zonal block storage)")
	blockstorageCreateCmd.MarkFlagRequired("region")
	blockstorageCreateCmd.Flags().Int("size", 0, "Size in GB (required)")
	blockstorageCreateCmd.Flags().String("type", "Standard", "Type: Standard or Performance")
	blockstorageCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	blockstorageCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	blockstorageCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	blockstorageCreateCmd.MarkFlagRequired("name")
	blockstorageCreateCmd.MarkFlagRequired("size")

	blockstorageGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	blockstorageUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageUpdateCmd.Flags().String("name", "", "New name for the block storage")
	blockstorageUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	blockstorageDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	blockstorageDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	blockstorageListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageListCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	blockstorageListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	blockstorageListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	blockstorageGetCmd.ValidArgsFunction = completeBlockStorageID
	blockstorageUpdateCmd.ValidArgsFunction = completeBlockStorageID
	blockstorageDeleteCmd.ValidArgsFunction = completeBlockStorageID
}

func completeBlockStorageID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Volumes().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, v := range list.Items() {
			id := v.ID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, v.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
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
The volume can be initialised from a snapshot with --snapshot-uri, or from an
image with --image. Pass --set-bootable to mark the volume as a boot disk.

Billing period: Hour (default), Month, or Year.`,
	Example: `  # Create an empty 50 GB Standard volume
  acloud storage blockstorage create --name my-volume --size 50 --region IT-BG

  # Create a Performance volume from a snapshot
  acloud storage blockstorage create --name fast-vol --size 100 --region IT-BG \
    --type Performance \
    --snapshot-uri /projects/<proj-id>/providers/Aruba.Storage/blockStorages/<vol-id>/snapshots/<snap-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		zone, _ := cmd.Flags().GetString("zone")
		size, _ := cmd.Flags().GetInt("size")
		volumeType, _ := cmd.Flags().GetString("type")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		snapshotURI, _ := cmd.Flags().GetString("snapshot-uri")
		setBootable, _ := cmd.Flags().GetBool("set-bootable")
		image, _ := cmd.Flags().GetString("image")
		verbose, _ := cmd.Flags().GetBool("verbose")

		if size <= 0 {
			return fmt.Errorf("--size must be greater than 0")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		if verbose {
			fmt.Println("\nCreating block storage with the following parameters:")
			fmt.Printf("  Name:           %s\n", name)
			fmt.Printf("  Region:         %s\n", region)
			if zone != "" {
				fmt.Printf("  Zone:           %s\n", zone)
			}
			fmt.Printf("  Size:           %d GB\n", size)
			fmt.Printf("  Type:           %s\n", volumeType)
			fmt.Printf("  Billing Period: %s\n", billingPeriod)
			if snapshotURI != "" {
				fmt.Printf("  Snapshot URI:   %s\n", snapshotURI)
			}
			if setBootable {
				fmt.Printf("  Bootable:       true\n")
			}
			if image != "" {
				fmt.Printf("  Image:          %s\n", image)
			}
			fmt.Printf("  Project ID:     %s\n", projectID)
			fmt.Println()
		}

		vol := aruba.NewBlockStorage().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfType(aruba.BlockStorageType(volumeType)).
			WithSizeGB(size).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		if zone != "" {
			vol.InZone(aruba.Zone(zone))
		}
		if len(tags) > 0 {
			vol.ReplaceTags(tags...)
		}
		if setBootable {
			vol.SetBootable()
		}
		if image != "" {
			vol.FromImage(image)
		}
		if snapshotURI != "" {
			vol.FromSnapshot(aruba.URI(snapshotURI))
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromStorage().Volumes().Create(ctx, vol)
		if err != nil {
			return fmt.Errorf("creating block storage: %w", apiErrFromV2(err))
		}

		resource := volumeFromRaw(created)
		if resource != nil {
			fmt.Printf("\n%s\n", msgCreated("Block storage", name))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			fmt.Printf("Size (GB):       %d\n", resource.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", resource.Properties.Type)
			fmt.Printf("Zone:            %s\n", resource.Properties.Zone)
			if resource.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", resource.Metadata.LocationResponse.Value)
			}
			if resource.Status.State != nil {
				fmt.Printf("Status:          %s\n", *resource.Status.State)
			}
			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
		} else {
			fmt.Println(msgCreatedAsync("Block storage", name))
		}
		return nil
	},
}

var blockstorageGetCmd = &cobra.Command{
	Use:   "get [volume-id]",
	Short: "Get block storage details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		got, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
		if err != nil {
			return fmt.Errorf("getting block storage details: %w", apiErrFromV2(err))
		}

		volume := volumeFromRaw(got)
		if volume != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(volume, nil, nil)
				return nil
			}

			fmt.Println("\nBlock Storage Details:")
			fmt.Println("======================")

			if volume.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *volume.Metadata.ID)
			}
			if volume.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *volume.Metadata.URI)
			}
			if volume.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *volume.Metadata.Name)
			}
			fmt.Printf("Size (GB):       %d\n", volume.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", volume.Properties.Type)
			fmt.Printf("Zone:            %s\n", volume.Properties.Zone)
			if volume.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", volume.Metadata.LocationResponse.Value)
			}
			if volume.Properties.Bootable != nil {
				fmt.Printf("Bootable:        %t\n", *volume.Properties.Bootable)
			}
			if volume.Status.State != nil {
				fmt.Printf("Status:          %s\n", *volume.Status.State)
			}
			if volume.Metadata.CreationDate != nil && !volume.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", volume.Metadata.CreationDate.Format(DateLayout))
			}
			if volume.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *volume.Metadata.CreatedBy)
			}
			if len(volume.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", volume.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			fmt.Println()
		} else {
			fmt.Println("Block storage not found")
		}
		return nil
	},
}

var blockstorageUpdateCmd = &cobra.Command{
	Use:   "update [volume-id]",
	Short: "Update block storage (name and/or tags)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			fmt.Println("Note: Size update is not supported by the API yet")
			return nil
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		current, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
		if err != nil {
			return fmt.Errorf("getting block storage details: %w", apiErrFromV2(err))
		}

		if current.Raw() == nil {
			fmt.Println("Block storage not found")
			return nil
		}

		if state := current.State(); state != "" && state != StateUsed && state != StateNotUsed {
			return fmt.Errorf("cannot update block storage with status '%s': block storage can only be updated when status is 'Used' or 'NotUsed'", state)
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}

		updated, err := client.FromStorage().Volumes().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating block storage: %w", apiErrFromV2(err))
		}

		resource := volumeFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("Block storage", volumeID))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", resource.Metadata.Tags)
			}
			fmt.Printf("Size (GB):       %d\n", resource.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", resource.Properties.Type)
		} else {
			fmt.Println(msgUpdatedAsync("Block storage", volumeID))
		}
		return nil
	},
}

var blockstorageDeleteCmd = &cobra.Command{
	Use:   "delete [volume-id]",
	Short: "Delete block storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("block storage", volumeID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
			if err != nil {
				return fmt.Errorf("dry-run: block storage not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("block storage", volumeID))
			return nil
		}

		if err := client.FromStorage().Volumes().Delete(ctx, volumeRef(projectID, volumeID)); err != nil {
			return fmt.Errorf("deleting block storage: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Block storage", volumeID))
		return nil
	},
}

var blockstorageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all block storage",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromStorage().Volumes().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing block storage: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			if verbose {
				fmt.Printf("\n=== Block Storage (count: %d) ===\n\n", len(list.Items()))
			}

			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "SIZE(GB)", Width: 12},
				{Header: "REGION", Width: 15},
				{Header: "ZONE", Width: 15},
				{Header: "TYPE", Width: 15},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, v := range list.Items() {
				r := volumeFromRaw(v)
				sizeStr := "0"
				region := string(v.Region())
				zone := string(v.Zone())
				typeStr := string(v.Type())
				if r != nil {
					sizeStr = fmt.Sprintf("%d", r.Properties.SizeGB)
				}
				rows = append(rows, []string{v.Name(), v.ID(), sizeStr, region, zone, typeStr, v.State()})
			}

			PrintOutput(volumeListPayload(list), headers, rows)
		} else {
			fmt.Println("No block storage found")
		}
		return nil
	},
}
