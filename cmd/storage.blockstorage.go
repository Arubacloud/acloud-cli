package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

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

// Completion functions for storage resources

func completeBlockStorageID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Allow completion even if args exist - user might be completing a partial ID

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromStorage().Volumes().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, volume := range response.Data.Values {
			if volume.Metadata.ID != nil && volume.Metadata.Name != nil {
				id := *volume.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *volume.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Blockstorage subcommands
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
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get flags
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

		// Validate required fields
		if size <= 0 {
			return fmt.Errorf("--size must be greater than 0")
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build the create request
		createRequest := types.BlockStorageRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.BlockStoragePropertiesRequest{
				SizeGB:        size,
				BillingPeriod: billingPeriod,
				Type:          types.BlockStorageType(volumeType),
			},
		}

		// Add zone only if provided
		if zone != "" {
			createRequest.Properties.Zone = &zone
		}

		// Add snapshot if provided
		if snapshotURI != "" {
			createRequest.Properties.Snapshot = &types.ReferenceResource{URI: snapshotURI}
		}

		// Add bootable if --set-bootable is provided (always true)
		if setBootable {
			bootable := true
			createRequest.Properties.Bootable = &bootable
		}

		// Add image if provided
		if image != "" {
			createRequest.Properties.Image = &image
		}

		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Debug output if verbose
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

		// Create the block storage using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromStorage().Volumes().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating block storage: %w", err)
		}

		if response != nil && response.IsError() {
			if response.Error != nil {
				return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
			}
			return fmt.Errorf("API error (status %d)", response.StatusCode)
		}

		if response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("Block storage", name))
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			fmt.Printf("Size (GB):       %d\n", response.Data.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", response.Data.Properties.Type)
			fmt.Printf("Zone:            %s\n", response.Data.Properties.Zone)
			if response.Data.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", response.Data.Metadata.LocationResponse.Value)
			}
			if response.Data.Status.State != nil {
				fmt.Printf("Status:          %s\n", *response.Data.Status.State)
			}
			if response.Data.Metadata.CreationDate != nil && !response.Data.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", response.Data.Metadata.CreationDate.Format(DateLayout))
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

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get block storage details using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if err != nil {
			return fmt.Errorf("getting block storage details: %w", err)
		}
		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		if response != nil && response.Data != nil {
			volume := response.Data

			// Display volume details
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

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one field must be provided
		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			fmt.Println("Note: Size update is not supported by the API yet")
			return nil
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// First, get the current volume details to preserve existing values
		ctx, cancel := newCtx()
		defer cancel()
		getResponse, err := client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if err != nil {
			return fmt.Errorf("getting block storage details: %w", err)
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Block storage not found")
			return nil
		}

		currentVolume := getResponse.Data

		// Check if the volume status allows updates
		if currentVolume.Status.State != nil {
			status := *currentVolume.Status.State
			if status != "Used" && status != "NotUsed" {
				return fmt.Errorf("cannot update block storage with status '%s': block storage can only be updated when status is 'Used' or 'NotUsed'", status)
			}
		}

		// Get region value
		regionValue := ""
		if currentVolume.Metadata.LocationResponse != nil {
			regionValue = currentVolume.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for block storage")
		}

		// Handle zone - if empty, set to nil
		var zone *string
		if currentVolume.Properties.Zone != "" {
			zone = &currentVolume.Properties.Zone
		}

		// Build the update request with current values as defaults
		updateRequest := types.BlockStorageRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *currentVolume.Metadata.Name,
					Tags: currentVolume.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.BlockStoragePropertiesRequest{
				SizeGB:        currentVolume.Properties.SizeGB,
				BillingPeriod: "Hour",
				Zone:          zone,
				Type:          currentVolume.Properties.Type,
			},
		}

		// Update only the fields that were provided
		if name != "" {
			updateRequest.Metadata.ResourceMetadataRequest.Name = name
		}

		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.ResourceMetadataRequest.Tags = tags
		}

		// Update the block storage using the SDK
		response, err := client.FromStorage().Volumes().Update(ctx, projectID, volumeID, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating block storage: %w", err)
		}

		if response != nil && response.IsError() {
			if response.Error != nil {
				return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
			}
			return fmt.Errorf("API error (status %d)", response.StatusCode)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("Block storage", volumeID))
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
			fmt.Printf("Size (GB):       %d\n", response.Data.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", response.Data.Properties.Type)
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

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get flags
		confirm, _ := cmd.Flags().GetBool("yes")

		// If not confirmed, ask for confirmation
		if !confirm {
			ok, err := confirmDelete("block storage", volumeID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: block storage not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("block storage", volumeID))
			return nil
		}

		// Delete the block storage using the SDK
		deleteResp, err := client.FromStorage().Volumes().Delete(ctx, projectID, volumeID, nil)
		if err != nil {
			return fmt.Errorf("deleting block storage: %w", err)
		}
		if deleteResp != nil && deleteResp.IsError() {
			return apiErrFromResp(deleteResp.StatusCode, deleteResp.Error)
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
		// Get flags
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// List block storage using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromStorage().Volumes().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing block storage: %w", err)
		}
		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		// Debug output
		if verbose {
			fmt.Println("\n=== DEBUG: Raw Response ===")
			fmt.Printf("Status Code: %d\n", response.StatusCode)
			fmt.Printf("Success: %v\n", response.IsSuccess())
			if response.Data != nil {
				fmt.Printf("Number of volumes: %d\n", len(response.Data.Values))
				if len(response.Data.Values) > 0 {
					fmt.Println("\n=== Volumes Detail ===")
					for i, vol := range response.Data.Values {
						fmt.Printf("\n--- Volume %d ---\n", i+1)
						if vol.Metadata.ID != nil {
							fmt.Printf("  ID: %s\n", *vol.Metadata.ID)
						}
						if vol.Metadata.Name != nil {
							fmt.Printf("  Name: %s\n", *vol.Metadata.Name)
						}
						fmt.Printf("  Size: %d GB\n", vol.Properties.SizeGB)
						fmt.Printf("  Type: %v\n", vol.Properties.Type)
						fmt.Printf("  Zone: %s\n", vol.Properties.Zone)
						fmt.Printf("  Region: %s\n", vol.Metadata.LocationResponse.Value)
						if vol.Status.State != nil {
							fmt.Printf("  Status State: %s\n", *vol.Status.State)
						} else {
							fmt.Printf("  Status State: <nil>\n")
						}
						fmt.Printf("  Full Status: %+v\n", vol.Status)
					}
				}
			}
			fmt.Println("\n=== End Debug ===")
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "SIZE(GB)", Width: 12},
				{Header: "REGION", Width: 15},
				{Header: "ZONE", Width: 15},
				{Header: "TYPE", Width: 15},
				{Header: "STATUS", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, volume := range response.Data.Values {
				name := ""
				if volume.Metadata.Name != nil && *volume.Metadata.Name != "" {
					name = *volume.Metadata.Name
				}

				id := ""
				if volume.Metadata.ID != nil {
					id = *volume.Metadata.ID
				}

				size := fmt.Sprintf("%d", volume.Properties.SizeGB)

				region := ""
				if volume.Metadata.LocationResponse != nil {
					region = volume.Metadata.LocationResponse.Value
				}
				zone := volume.Properties.Zone

				volumeType := fmt.Sprintf("%v", volume.Properties.Type)

				status := ""
				if volume.Status.State != nil {
					status = *volume.Status.State
				}

				rows = append(rows, []string{name, id, size, region, zone, volumeType, status})
			}

			// Print the table
			PrintTable(headers, rows)
		} else {
			fmt.Println("No block storage found")
		}
		return nil
	},
}
