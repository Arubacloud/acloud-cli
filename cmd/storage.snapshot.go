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

	snapshotListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotListCmd.Flags().String("volume-uri", "", "Block storage volume URI (required)")
	snapshotListCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
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
	Run: func(cmd *cobra.Command, args []string) {
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		volumeURI, _ := cmd.Flags().GetString("volume-uri")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Validate required fields
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}
		if region == "" {
			fmt.Println("Error: --region is required")
			return
		}
		if volumeURI == "" {
			fmt.Println("Error: --volume-uri is required")
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
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
		ctx := context.Background()
		response, err := client.FromStorage().Snapshots().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating snapshot: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nSnapshot created successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if !response.Data.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", response.Data.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
		} else {
			fmt.Println("Warning: Snapshot may have been created but response is empty")
		}
	},
}

var snapshotGetCmd = &cobra.Command{
	Use:   "get [snapshot-id]",
	Short: "Get snapshot details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapshotID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get snapshot details using the SDK
		ctx := context.Background()
		response, err := client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
		if err != nil {
			fmt.Printf("Error getting snapshot details: %v\n", err)
			return
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
				fmt.Printf("Region:          %s\n", snapshot.Metadata.LocationResponse.Value)
			}

			status := ""
			if snapshot.Status.State != nil {
				status = *snapshot.Status.State
			}
			fmt.Printf("Status:          %s\n", status)

			if !snapshot.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", snapshot.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}

var snapshotUpdateCmd = &cobra.Command{
	Use:   "update [snapshot-id]",
	Short: "Update a snapshot (name and/or tags only)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapshotID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one field must be provided
		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// First, get the current snapshot details to preserve existing values
		ctx := context.Background()
		getResponse, err := client.FromStorage().Snapshots().Get(ctx, projectID, snapshotID, nil)
		if err != nil {
			fmt.Printf("Error getting snapshot details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Snapshot not found")
			return
		}

		currentSnapshot := getResponse.Data

		// Get region value
		regionValue := ""
		if currentSnapshot.Metadata.LocationResponse != nil {
			regionValue = currentSnapshot.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for snapshot")
			return
		}

		// Build the update request with current values as defaults
		volumeURI := ""
		if currentSnapshot.Properties.Volume != nil && currentSnapshot.Properties.Volume.URI != nil {
			volumeURI = *currentSnapshot.Properties.Volume.URI
		}

		updateRequest := types.SnapshotRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *currentSnapshot.Metadata.Name,
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
			fmt.Printf("Error updating snapshot: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update snapshot - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nSnapshot updated successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("Warning: Update may have succeeded but response is empty")
		}
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete [snapshot-id]",
	Short: "Delete a snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		snapshotID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get flags
		confirm, _ := cmd.Flags().GetBool("yes")

		// If not confirmed, ask for confirmation
		if !confirm {
			fmt.Printf("Are you sure you want to delete snapshot %s? (yes/no): ", snapshotID)
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Delete the snapshot using the SDK
		ctx := context.Background()
		_, err = client.FromStorage().Snapshots().Delete(ctx, projectID, snapshotID, nil)
		if err != nil {
			fmt.Printf("Error deleting snapshot: %v\n", err)
			return
		}

		fmt.Printf("\nSnapshot %s deleted successfully!\n", snapshotID)
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots for a block storage volume",
	Run: func(cmd *cobra.Command, args []string) {
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get block storage URI flag
		volumeURI, _ := cmd.Flags().GetString("volume-uri")
		if volumeURI == "" {
			fmt.Println("Error: --volume-uri is required")
			fmt.Println("Use: acloud storage snapshot list --volume-uri <block-storage-uri>")
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// List snapshots using the SDK (filter by volume URI on client side)
		ctx := context.Background()
		response, err := client.FromStorage().Snapshots().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing snapshots: %v\n", err)
			return
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
				return
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
	},
}
