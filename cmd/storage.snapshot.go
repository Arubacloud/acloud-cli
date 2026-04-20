package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	storageCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotGetCmd)
	snapshotCmd.AddCommand(snapshotUpdateCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	snapshotCmd.AddCommand(snapshotListCmd)

	// Add flags for snapshot commands
	snapshotCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotCreateCmd.Flags().String("name", "", "Name for the snapshot (required)")
	snapshotCreateCmd.Flags().String("region", "", "Region code (required)")
	snapshotCreateCmd.Flags().String("volume-uri", "", "Source volume URI (required)")
	snapshotCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	snapshotCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	snapshotCreateCmd.MarkFlagRequired("name")
	snapshotCreateCmd.MarkFlagRequired("region")
	snapshotCreateCmd.MarkFlagRequired("volume-uri")

	snapshotGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	snapshotUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotUpdateCmd.Flags().String("name", "", "New name for the snapshot")
	snapshotUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	snapshotDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	snapshotDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	snapshotListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotListCmd.Flags().String("volume-uri", "", "Block storage volume URI (required)")
	snapshotListCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	snapshotListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	snapshotListCmd.Flags().Int32("offset", 0, "Number of results to skip")
	snapshotListCmd.MarkFlagRequired("volume-uri")

	snapshotGetCmd.ValidArgsFunction = completeSnapshotID
	snapshotUpdateCmd.ValidArgsFunction = completeSnapshotID
	snapshotDeleteCmd.ValidArgsFunction = completeSnapshotID

}

// Completion functions for storage resources

func completeSnapshotID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	response, err := client.FromStorage().Snapshots().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, snapshot := range response.Data.Values {
			if snapshot.Metadata.ID != nil && snapshot.Metadata.Name != nil {
				id := *snapshot.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *snapshot.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Snapshot subcommands
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots",
	Long:  `Perform CRUD operations on snapshots in Aruba Cloud.`,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new snapshot",
	Long: `Create a point-in-time snapshot of an existing block storage volume.

The volume URI is required. Snapshots can later be used to create new volumes
or restore data with 'acloud storage blockstorage create --snapshot-uri'.`,
	Example: `  acloud storage snapshot create --name my-snap --region IT-BG \
    --volume-uri /projects/<proj-id>/providers/Aruba.Storage/blockStorages/<vol-id>`,
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
		volumeURI, _ := cmd.Flags().GetString("volume-uri")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build the create request
		createRequest := types.SnapshotRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.SnapshotPropertiesRequest{
				Volume: types.ReferenceResource{
					URI: volumeURI,
				},
			},
		}

		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating snapshot with the following parameters:")
			fmt.Printf("  Name:       %s\n", name)
			fmt.Printf("  Region:     %s\n", region)
			fmt.Printf("  Volume URI: %s\n", volumeURI)
			fmt.Printf("  Tags:       %v\n", tags)
			fmt.Println()
		}

		// Create the snapshot using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromStorage().Snapshots().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating snapshot: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("Snapshot", name))
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			}
			if response.Data.Metadata.CreationDate != nil && !response.Data.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", response.Data.Metadata.CreationDate.Format(DateLayout))
			}
		} else {
			fmt.Println(msgCreatedAsync("Snapshot", name))
		}
		return nil
	},
}

var snapshotGetCmd = &cobra.Command{
	Use:   "get [snapshot-id]",
	Short: "Get snapshot details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

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

		// Get snapshot details using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
		if err != nil {
			return fmt.Errorf("getting snapshot details: %w", err)
		}
		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			snapshot := response.Data

			// Display snapshot details
			fmt.Println("\nSnapshot Details:")
			fmt.Println("=================")

			if snapshot.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *snapshot.Metadata.ID)
			}

			if snapshot.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *snapshot.Metadata.URI)
			}

			if snapshot.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *snapshot.Metadata.Name)
			}

			fmt.Printf("Size (GB):       %d\n", snapshot.Properties.SizeGB)

			if snapshot.Properties.Volume != nil && snapshot.Properties.Volume.URI != nil {
				fmt.Printf("Source Volume:   %s\n", *snapshot.Properties.Volume.URI)
			}

			if snapshot.Metadata.LocationResponse != nil {
				if snapshot.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", snapshot.Metadata.LocationResponse.Value)
				}
			}

			status := ""
			if snapshot.Status.State != nil {
				status = *snapshot.Status.State
			}
			fmt.Printf("Status:          %s\n", status)

			if snapshot.Metadata.CreationDate != nil && !snapshot.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", snapshot.Metadata.CreationDate.Format(DateLayout))
			}

			if snapshot.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *snapshot.Metadata.CreatedBy)
			}

			if len(snapshot.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", snapshot.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			fmt.Println()
		} else {
			fmt.Println("Snapshot not found")
		}
		return nil
	},
}

var snapshotUpdateCmd = &cobra.Command{
	Use:   "update [snapshot-id]",
	Short: "Update a snapshot (name and/or tags only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

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
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// First, get the current snapshot details to preserve existing values
		ctx, cancel := newCtx()
		defer cancel()
		getResponse, err := client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
		if err != nil {
			return fmt.Errorf("getting snapshot details: %w", err)
		}

		if getResponse == nil || getResponse.Data == nil {
			return fmt.Errorf("snapshot not found")
		}

		currentSnapshot := getResponse.Data

		// Get region value
		regionValue := ""
		if currentSnapshot.Metadata.LocationResponse != nil {
			regionValue = currentSnapshot.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for snapshot")
		}

		// Build the update request with current values as defaults
		volumeURI := ""
		if currentSnapshot.Properties.Volume != nil && currentSnapshot.Properties.Volume.URI != nil {
			volumeURI = *currentSnapshot.Properties.Volume.URI
		}

		currentName := ""
		if currentSnapshot.Metadata.Name != nil {
			currentName = *currentSnapshot.Metadata.Name
		}
		updateRequest := types.SnapshotRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: currentName,
					Tags: currentSnapshot.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.SnapshotPropertiesRequest{
				Volume: types.ReferenceResource{
					URI: volumeURI,
				},
			},
		}

		// Update only the fields that were provided
		if name != "" {
			updateRequest.Metadata.ResourceMetadataRequest.Name = name
		}

		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.ResourceMetadataRequest.Tags = tags
		}

		// Update the snapshot using the SDK
		response, err := client.FromStorage().Snapshots().Update(ctx, projectID, snapshotID, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating snapshot: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("Snapshot", snapshotID))
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			}
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("Snapshot", snapshotID))
		}
		return nil
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete [snapshot-id]",
	Short: "Delete a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get flags
		confirm, _ := cmd.Flags().GetBool("yes")

		// If not confirmed, ask for confirmation
		if !confirm {
			ok, err := confirmDelete("snapshot", snapshotID)
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
			_, err = client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: snapshot not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("snapshot", snapshotID))
			return nil
		}

		// Delete the snapshot using the SDK
		deleteResp, err := client.FromStorage().Snapshots().Delete(ctx, projectID, snapshotID, nil)
		if err != nil {
			return fmt.Errorf("deleting snapshot: %w", err)
		}
		if deleteResp != nil && deleteResp.IsError() && deleteResp.Error != nil {
			return fmtAPIError(deleteResp.StatusCode, deleteResp.Error.Title, deleteResp.Error.Detail)
		}

		fmt.Println(msgDeleted("Snapshot", snapshotID))
		return nil
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots for a block storage volume",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get block storage URI flag
		volumeURI, _ := cmd.Flags().GetString("volume-uri")

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// List snapshots using the SDK (filter by volume URI on client side)
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromStorage().Snapshots().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing snapshots: %w", err)
		}
		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		// Check verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			fmt.Println("=== Full API Response ===")
			fmt.Printf("%+v\n\n", response)
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Filter snapshots by volume URI
			var filteredSnapshots []types.SnapshotResponse
			for _, snapshot := range response.Data.Values {
				if snapshot.Properties.Volume != nil && snapshot.Properties.Volume.URI != nil && *snapshot.Properties.Volume.URI == volumeURI {
					filteredSnapshots = append(filteredSnapshots, snapshot)
				}
			}

			if len(filteredSnapshots) == 0 {
				fmt.Printf("No snapshots found for volume: %s\n", volumeURI)
				return nil
			}

			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "SIZE(GB)", Width: 12},
				{Header: "STATUS", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, snapshot := range filteredSnapshots {
				name := ""
				if snapshot.Metadata.Name != nil && *snapshot.Metadata.Name != "" {
					name = *snapshot.Metadata.Name
				}

				id := ""
				if snapshot.Metadata.ID != nil && *snapshot.Metadata.ID != "" {
					id = *snapshot.Metadata.ID
				}

				size := fmt.Sprintf("%d", snapshot.Properties.SizeGB)

				status := ""
				if snapshot.Status.State != nil {
					status = *snapshot.Status.State
				}

				rows = append(rows, []string{name, id, size, status})
			}

			// Print the table
			PrintTable(headers, rows)
		} else {
			fmt.Println("No snapshots found")
		}
		return nil
	},
}
