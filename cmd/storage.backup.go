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

	// Backup commands
	storageCmd.AddCommand(storageBackupCmd)
	storageBackupCmd.AddCommand(storageBackupListCmd)
	storageBackupCmd.AddCommand(storageBackupGetCmd)
	storageBackupCmd.AddCommand(storageBackupUpdateCmd)
	storageBackupCmd.AddCommand(storageBackupDeleteCmd)

	// Add flags for backup command
	storageBackupCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupCmd.Flags().String("name", "", "Name for the backup (required)")
	storageBackupCmd.Flags().String("region", "ITBG-Bergamo", "Region code")
	storageBackupCmd.Flags().String("type", "Full", "Backup type: Full or Incremental")
	storageBackupCmd.Flags().Int("retention-days", 0, "Number of days to retain the backup")
	storageBackupCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")
	storageBackupCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	storageBackupCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	storageBackupCmd.MarkFlagRequired("name")

	storageBackupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupUpdateCmd.Flags().String("name", "", "New name for the backup")
	storageBackupUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	storageBackupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	storageBackupGetCmd.ValidArgsFunction = completeBackupID
	storageBackupUpdateCmd.ValidArgsFunction = completeBackupID
	storageBackupDeleteCmd.ValidArgsFunction = completeBackupID
}

// Completion functions for storage resources

func completeBackupID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	response, err := client.FromStorage().Backups().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, backup := range response.Data.Values {
			if backup.Metadata.ID != nil && backup.Metadata.Name != nil {
				id := *backup.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *backup.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Backup command - creates an Aruba.Storage/backup resource
var storageBackupCmd = &cobra.Command{
	Use:   "backup [volume-id]",
	Short: "Create a storage backup of a block storage volume",
	Long:  `Create a storage backup resource (Aruba.Storage/backup) for a block storage volume.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		volumeID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		backupType, _ := cmd.Flags().GetString("type")
		retentionDays, _ := cmd.Flags().GetInt("retention-days")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
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

		// First, get the volume details to get the full URI
		ctx := context.Background()
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

		// Build the backup create request
		createRequest := types.StorageBackupRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.StorageBackupPropertiesRequest{
				StorageBackupType: types.StorageBackupType(backupType),
				Origin: types.ReferenceResource{
					URI: volumeURI,
				},
			},
		}

		// Add optional fields
		if retentionDays > 0 {
			createRequest.Properties.RetentionDays = &retentionDays
		}
		if billingPeriod != "" {
			createRequest.Properties.BillingPeriod = &billingPeriod
		}

		// Get verbose flag
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating storage backup with the following parameters:")
			fmt.Printf("  Name:           %s\n", name)
			fmt.Printf("  Type:           %s\n", backupType)
			fmt.Printf("  Region:         %s\n", region)
			fmt.Printf("  Volume ID:      %s\n", volumeID)
			fmt.Printf("  Volume URI:     %s\n", volumeURI)
			if retentionDays > 0 {
				fmt.Printf("  Retention Days: %d\n", retentionDays)
			}
			if billingPeriod != "" {
				fmt.Printf("  Billing Period: %s\n", billingPeriod)
			}
			if len(tags) > 0 {
				fmt.Printf("  Tags:           %v\n", tags)
			}
			fmt.Println()
		}

		// Create the backup using the SDK
		response, err := client.FromStorage().Backups().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating backup: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create backup - Status: %d\n", response.StatusCode)
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
			fmt.Println("Storage backup created successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			fmt.Printf("Type:            %s\n", response.Data.Properties.Type)
			if !response.Data.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", response.Data.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
		}
	},
}

// Backup List command
var storageBackupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage backups",
	Run: func(cmd *cobra.Command, args []string) {
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
		response, err := client.FromStorage().Backups().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing backups: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "TYPE", Width: 12},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, backup := range response.Data.Values {
				name := ""
				if backup.Metadata.Name != nil {
					name = *backup.Metadata.Name
				}

				id := ""
				if backup.Metadata.ID != nil {
					id = *backup.Metadata.ID
				}

				backupType := string(backup.Properties.Type)

				status := ""
				if backup.Status.State != nil {
					status = *backup.Status.State
				}

				rows = append(rows, []string{name, id, backupType, status})
			}

			PrintTable(headers, rows)
		} else {
			fmt.Println("No backups found")
		}
	},
}

// Backup Get command
var storageBackupGetCmd = &cobra.Command{
	Use:   "get [backup-id]",
	Short: "Get storage backup details",
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
		response, err := client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error getting backup details: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			backup := response.Data

			fmt.Println("\nStorage Backup Details:")
			fmt.Println("=======================")

			if backup.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *backup.Metadata.ID)
			}
			if backup.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *backup.Metadata.URI)
			}
			if backup.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *backup.Metadata.Name)
			}

			fmt.Printf("Type:            %s\n", backup.Properties.Type)

			if backup.Properties.Origin.URI != "" {
				fmt.Printf("Source Volume:   %s\n", backup.Properties.Origin.URI)
			}

			if backup.Properties.RetentionDays != nil {
				fmt.Printf("Retention Days:  %d\n", *backup.Properties.RetentionDays)
			}

			if backup.Properties.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", *backup.Properties.BillingPeriod)
			}

			fmt.Printf("Region:          %s\n", backup.Metadata.LocationResponse.Value)

			if backup.Status.State != nil {
				fmt.Printf("Status:          %s\n", *backup.Status.State)
			}

			if !backup.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", backup.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}

			if backup.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *backup.Metadata.CreatedBy)
			}

			if len(backup.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", backup.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
		} else {
			fmt.Println("Backup not found")
		}
	},
}

// Backup Update command
var storageBackupUpdateCmd = &cobra.Command{
	Use:   "update [backup-id]",
	Short: "Update a storage backup (name and/or tags)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]

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

		// First, get the current backup details
		ctx := context.Background()
		getResponse, err := client.FromStorage().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error getting backup details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Backup not found")
			return
		}

		currentBackup := getResponse.Data

		// Get region value
		regionValue := ""
		if currentBackup.Metadata.LocationResponse != nil {
			regionValue = currentBackup.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for backup")
			return
		}

		// Build the update request with current values as defaults
		updateRequest := types.StorageBackupRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *currentBackup.Metadata.Name,
					Tags: currentBackup.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.StorageBackupPropertiesRequest{
				StorageBackupType: currentBackup.Properties.Type,
				Origin: types.ReferenceResource{
					URI: currentBackup.Properties.Origin.URI,
				},
			},
		}

		// Add optional fields if present
		if currentBackup.Properties.RetentionDays != nil {
			updateRequest.Properties.RetentionDays = currentBackup.Properties.RetentionDays
		}
		if currentBackup.Properties.BillingPeriod != nil {
			updateRequest.Properties.BillingPeriod = currentBackup.Properties.BillingPeriod
		}

		// Update only the fields that were provided
		if name != "" {
			updateRequest.Metadata.ResourceMetadataRequest.Name = name
		}

		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.ResourceMetadataRequest.Tags = tags
		}

		// Update the backup using the SDK
		response, err := client.FromStorage().Backups().Update(ctx, projectID, backupID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating backup: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update backup - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nBackup updated successfully!")
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

// Backup Delete command
var storageBackupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a storage backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		backupID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			fmt.Printf("Are you sure you want to delete backup %s? (yes/no): ", backupID)
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
		_, err = client.FromStorage().Backups().Delete(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error deleting backup: %v\n", err)
			return
		}

		fmt.Printf("\nBackup %s deleted successfully!\n", backupID)
	},
}
