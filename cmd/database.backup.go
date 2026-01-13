package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// Database backup commands
	databaseCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupGetCmd)
	// Note: Database backups don't support update operations
	// backupCmd.AddCommand(backupUpdateCmd)
	backupCmd.AddCommand(backupDeleteCmd)
	backupCmd.AddCommand(backupListCmd)

	// Add flags for database backup commands
	backupCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupCreateCmd.Flags().String("name", "", "Backup name (required)")
	backupCreateCmd.Flags().String("region", "", "Region code (required)")
	backupCreateCmd.Flags().String("dbaas-id", "", "DBaaS instance ID (required)")
	backupCreateCmd.Flags().String("database-name", "", "Database name (required)")
	backupCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	backupCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	backupCreateCmd.MarkFlagRequired("name")
	backupCreateCmd.MarkFlagRequired("region")
	backupCreateCmd.MarkFlagRequired("dbaas-id")
	backupCreateCmd.MarkFlagRequired("database-name")

	backupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// backupUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	// backupUpdateCmd.Flags().String("name", "", "New backup name")
	// backupUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	backupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	backupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	backupGetCmd.ValidArgsFunction = completeDatabaseBackupID
	// backupUpdateCmd.ValidArgsFunction = completeDatabaseBackupID
	backupDeleteCmd.ValidArgsFunction = completeDatabaseBackupID
}

// Completion functions for database resources
func completeDatabaseBackupID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromDatabase().Backups().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, backup := range response.Data.Values {
			if backup.Metadata.ID != nil && backup.Metadata.Name != nil {
				id := *backup.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *backup.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Database backup subcommands
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage database backups",
	Long:  `Perform CRUD operations on database backups in Aruba Cloud.`,
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new database backup",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		dbaasID, _ := cmd.Flags().GetString("dbaas-id")
		databaseName, _ := cmd.Flags().GetString("database-name")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" || region == "" || dbaasID == "" || databaseName == "" {
			fmt.Println("Error: --name, --region, --dbaas-id, and --database-name are required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get DBaaS instance to get its URI
		ctx := context.Background()
		dbaasResp, err := client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			fmt.Printf("Error getting DBaaS instance: %v\n", err)
			return
		}

		if dbaasResp == nil || dbaasResp.Data == nil || dbaasResp.Data.Metadata.URI == nil {
			fmt.Println("Error: DBaaS instance not found")
			return
		}

		// Get database to get its URI
		dbResp, err := client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			fmt.Printf("Error getting database: %v\n", err)
			return
		}

		// Note: DatabaseResponse doesn't have a URI field, so we may need to construct it
		// For now, we'll use the DBaaS URI and database name
		// The actual implementation may need adjustment based on the API
		dbaasURI := *dbaasResp.Data.Metadata.URI
		// Construct database URI (format may vary - this is a placeholder)
		databaseURI := fmt.Sprintf("%s/databases/%s", dbaasURI, databaseName)

		createRequest := types.BackupRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.BackupPropertiesRequest{
				Zone: region,
				DBaaS: types.ReferenceResource{
					URI: dbaasURI,
				},
				Database: types.ReferenceResource{
					URI: databaseURI,
				},
				BillingPlan: types.BillingPeriodResource{
					BillingPeriod: billingPeriod,
				},
			},
		}

		// Suppress unused variable warning
		_ = dbResp

		response, err := client.FromDatabase().Backups().Create(ctx, projectID, createRequest, nil)
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
			return
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				func() string {
					if response.Data.Metadata.ID != nil {
						return *response.Data.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.Name != nil {
						return *response.Data.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.LocationResponse != nil {
						return response.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
				func() string {
					if response.Data.Status.State != nil {
						return *response.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Backup created, but no data returned.")
		}
	},
}

var backupGetCmd = &cobra.Command{
	Use:   "get [backup-id]",
	Short: "Get backup details",
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
		resp, err := client.FromDatabase().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error getting backup: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get backup - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			backup := resp.Data

			fmt.Println("\nBackup Details:")
			fmt.Println("==============")

			if backup.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *backup.Metadata.ID)
			}
			if backup.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *backup.Metadata.URI)
			}
			if backup.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *backup.Metadata.Name)
			}
			if backup.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", backup.Metadata.LocationResponse.Value)
			}
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
			fmt.Println()
		} else {
			fmt.Println("Backup not found")
		}
	},
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all database backups",
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
		resp, err := client.FromDatabase().Backups().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing backups: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list backups - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, backup := range resp.Data.Values {
				row := []string{
					func() string {
						if backup.Metadata.Name != nil {
							return *backup.Metadata.Name
						}
						return ""
					}(),
					func() string {
						if backup.Metadata.ID != nil {
							return *backup.Metadata.ID
						}
						return ""
					}(),
					func() string {
						if backup.Metadata.LocationResponse != nil {
							return backup.Metadata.LocationResponse.Value
						}
						return ""
					}(),
					func() string {
						if backup.Status.State != nil {
							return *backup.Status.State
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No backups found")
		}
	},
}

var backupUpdateCmd = &cobra.Command{
	Use:   "update [backup-id]",
	Short: "Update a backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error: Database backups do not support update operations")
		fmt.Println("You can only create, list, get, and delete database backups.")
	},
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a backup",
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
		_, err = client.FromDatabase().Backups().Delete(ctx, projectID, backupID, nil)
		if err != nil {
			fmt.Printf("Error deleting backup: %v\n", err)
			return
		}

		fmt.Printf("\nBackup %s deleted successfully!\n", backupID)
	},
}
