package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// Restore commands
	storageCmd.AddCommand(storageRestoreCmd)
	storageRestoreCmd.AddCommand(storageRestoreListCmd)
	storageRestoreCmd.AddCommand(storageRestoreGetCmd)
	storageRestoreCmd.AddCommand(storageRestoreUpdateCmd)
	storageRestoreCmd.AddCommand(storageRestoreDeleteCmd)

	// Add flags for restore command
	storageRestoreCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreCmd.Flags().String("name", "", "Name for the restore operation (required)")
	storageRestoreCmd.Flags().String("region", "ITBG-Bergamo", "Region code")
	storageRestoreCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	storageRestoreCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	storageRestoreCmd.MarkFlagRequired("name")

	storageRestoreListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreUpdateCmd.Flags().String("name", "", "New name for the restore operation")
	storageRestoreUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	storageRestoreDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	storageRestoreGetCmd.ValidArgsFunction = completeRestoreID
	storageRestoreUpdateCmd.ValidArgsFunction = completeRestoreID
	storageRestoreDeleteCmd.ValidArgsFunction = completeRestoreID
	storageRestoreListCmd.ValidArgsFunction = completeBackupID
}

// Completion functions for storage resources

func completeRestoreID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// For restore, we need backup-id as first arg, restore-id as second
	if len(args) == 0 {
		// First arg is backup-id, use backup completion
		return completeBackupID(cmd, args, toComplete)
	}
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Second arg is restore-id
	backupID := args[0]
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromStorage().Restores().List(ctx, projectID, backupID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, restore := range response.Data.Values {
			if restore.Metadata.ID != nil && restore.Metadata.Name != nil {
				id := *restore.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *restore.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Restore command - creates an Aruba.Storage/restore resource
var storageRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id] [volume-id]",
	Short: "Restore a block storage volume from a backup",
	Long:  `Create a restore operation (Aruba.Storage/restore) to restore a volume from a backup.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]
		volumeID := args[1]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
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

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()

		// Get the backup details
		backupResponse, err := client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error getting backup details: %v\n", err)
			return
		}

		if backupResponse == nil || backupResponse.Data == nil {
			fmt.Println("Backup not found")
			return
		}

		backupURI := *backupResponse.Data.Metadata.URI

		// Get the volume details
		volumeResponse, err := client.FromStorage().Volumes().Get(ctx, projectID, volumeID, nil)
		if err != nil {
			fmt.Printf("Error getting volume details: %v\n", err)
			return
		}

		if volumeResponse == nil || volumeResponse.Data == nil {
			fmt.Println("Volume not found")
			return
		}

		volumeURI := *volumeResponse.Data.Metadata.URI

		// Build the restore request
		createRequest := types.RestoreRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.RestorePropertiesRequest{
				Target: types.ReferenceResource{
					URI: volumeURI,
				},
			},
		}

		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating restore operation with the following parameters:")
			fmt.Printf("  Name:       %s\n", name)
			fmt.Printf("  Region:     %s\n", region)
			fmt.Printf("  Backup ID:  %s\n", backupID)
			fmt.Printf("  Backup URI: %s\n", backupURI)
			fmt.Printf("  Volume ID:  %s\n", volumeID)
			fmt.Printf("  Volume URI: %s\n", volumeURI)
			if len(tags) > 0 {
				fmt.Printf("  Tags:       %v\n", tags)
			}
			fmt.Println()
		}

		// Create the restore using the SDK
		response, err := client.FromStorage().Restores().Create(ctx, projectID, backupID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating restore: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create restore - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			if verbose {
				if response.RawBody != nil {
					var errorDetail map[string]interface{}
					if err := json.Unmarshal(response.RawBody, &errorDetail); err == nil {
						fmt.Printf("Full Error Response: %+v\n", errorDetail)
					}
				}
			}
			return
		}

		if response.Data != nil {
			fmt.Println("Restore operation created successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if !response.Data.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", response.Data.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if response.Data.Status.State != nil {
				fmt.Printf("Status:          %s\n", *response.Data.Status.State)
			}
		}
	},
}

// Restore List command
var storageRestoreListCmd = &cobra.Command{
	Use:   "list [backup-id]",
	Short: "List restore operations for a backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		response, err := client.FromStorage().Restores().List(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error listing restores: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, restore := range response.Data.Values {
				name := ""
				if restore.Metadata.Name != nil {
					name = *restore.Metadata.Name
				}

				id := ""
				if restore.Metadata.ID != nil {
					id = *restore.Metadata.ID
				}

				status := ""
				if restore.Status.State != nil {
					status = *restore.Status.State
				}

				rows = append(rows, []string{name, id, status})
			}

			PrintTable(headers, rows)
		} else {
			fmt.Println("No restores found for this backup")
		}
	},
}

// Restore Get command
var storageRestoreGetCmd = &cobra.Command{
	Use:   "get [backup-id] [restore-id]",
	Short: "Get restore operation details",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]
		restoreID := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		response, err := client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
		if err != nil {
			fmt.Printf("Error getting restore details: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			restore := response.Data

			fmt.Println("\nRestore Operation Details:")
			fmt.Println("==========================")

			if restore.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *restore.Metadata.ID)
			}
			if restore.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *restore.Metadata.URI)
			}
			if restore.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *restore.Metadata.Name)
			}

			if restore.Properties.Destination.URI != "" {
				fmt.Printf("Target Volume:   %s\n", restore.Properties.Destination.URI)
			}

			fmt.Printf("Region:          %s\n", restore.Metadata.LocationResponse.Value)

			if restore.Status.State != nil {
				fmt.Printf("Status:          %s\n", *restore.Status.State)
			}

			if !restore.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", restore.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}

			if restore.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *restore.Metadata.CreatedBy)
			}

			if len(restore.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", restore.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
		} else {
			fmt.Println("Restore operation not found")
		}
	},
}

// Restore Update command
var storageRestoreUpdateCmd = &cobra.Command{
	Use:   "update [backup-id] [restore-id]",
	Short: "Update a restore operation (name and/or tags)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]
		restoreID := args[1]

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

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// First, get the current restore details
		ctx := context.Background()
		getResponse, err := client.FromStorage().Restores().Get(ctx, projectID, backupID, restoreID, nil)
		if err != nil {
			fmt.Printf("Error getting restore details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Restore operation not found")
			return
		}

		currentRestore := getResponse.Data

		// Get region value
		regionValue := ""
		if currentRestore.Metadata.LocationResponse != nil {
			regionValue = currentRestore.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for restore operation")
			return
		}

		// Build the update request with current values as defaults
		updateRequest := types.RestoreRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *currentRestore.Metadata.Name,
					Tags: currentRestore.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.RestorePropertiesRequest{
				Target: types.ReferenceResource{
					URI: currentRestore.Properties.Destination.URI,
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

		// Update the restore using the SDK
		response, err := client.FromStorage().Restores().Update(ctx, projectID, backupID, restoreID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating restore: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update restore - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nRestore operation updated successfully!")
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

// Restore Delete command
var storageRestoreDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id] [restore-id]",
	Short: "Delete a restore operation",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]
		restoreID := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			fmt.Printf("Are you sure you want to delete restore operation %s? (yes/no): ", restoreID)
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		_, err = client.FromStorage().Restores().Delete(ctx, projectID, backupID, restoreID, nil)
		if err != nil {
			fmt.Printf("Error deleting restore: %v\n", err)
			return
		}

		fmt.Printf("\nRestore operation %s deleted successfully!\n", restoreID)
	},
}
