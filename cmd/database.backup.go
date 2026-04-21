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
	backupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	backupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	backupListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a backup of a database inside a DBaaS instance.

Provide the DBaaS instance ID (--dbaas-id) and the database name (--database-name).
The backup will be stored and can be restored later.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud database backup create \
    --name my-backup --region IT-BG \
    --dbaas-id <dbaas-id> \
    --database-name myapp_db`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		dbaasID, _ := cmd.Flags().GetString("dbaas-id")
		databaseName, _ := cmd.Flags().GetString("database-name")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get DBaaS instance to get its URI
		ctx, cancel := newCtx()
		defer cancel()
		dbaasResp, err := client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			return fmt.Errorf("getting DBaaS instance: %w", err)
		}

		if dbaasResp == nil || dbaasResp.Data == nil || dbaasResp.Data.Metadata.URI == nil {
			return fmt.Errorf("DBaaS instance not found")
		}

		// Get database to get its URI
		dbResp, err := client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			return fmt.Errorf("getting database: %w", err)
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
			return fmt.Errorf("creating backup: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
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
			PrintOutput(response.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Backup", name))
		}
		return nil
	},
}

var backupGetCmd = &cobra.Command{
	Use:   "get [backup-id]",
	Short: "Get backup details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]

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
		resp, err := client.FromDatabase().Backups().Get(ctx, projectID, backupID, nil)
		if err != nil {
			return fmt.Errorf("getting backup: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil {
			backup := resp.Data

			format := resolveOutputFormat()
			if format == "json" || format == "yaml" {
				PrintOutput(backup, nil, nil)
				return nil
			}

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
				if backup.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", backup.Metadata.LocationResponse.Value)
				}
			}
			if backup.Status.State != nil {
				fmt.Printf("Status:          %s\n", *backup.Status.State)
			}
			if backup.Metadata.CreationDate != nil && !backup.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", backup.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all database backups",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		resp, err := client.FromDatabase().Backups().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing backups: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
			PrintOutput(resp.Data, headers, rows)
		} else {
			fmt.Println("No backups found")
		}
		return nil
	},
}

var backupUpdateCmd = &cobra.Command{
	Use:   "update [backup-id]",
	Short: "Update a backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Error: Database backups do not support update operations")
		fmt.Println("You can only create, list, get, and delete database backups.")
		return nil
	},
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			ok, err := confirmDelete("backup", backupID)
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
			_, err = client.FromDatabase().Backups().Get(ctx, projectID, backupID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: database backup not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("database backup", backupID))
			return nil
		}

		_, err = client.FromDatabase().Backups().Delete(ctx, projectID, backupID, nil)
		if err != nil {
			return fmt.Errorf("deleting backup: %w", err)
		}

		fmt.Println(msgDeleted("Backup", backupID))
		return nil
	},
}
