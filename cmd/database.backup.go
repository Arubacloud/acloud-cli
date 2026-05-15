package cmd

import (
	"context"
	"fmt"
	"strings"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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

	backupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	backupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	backupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	backupListCmd.Flags().Int("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	backupGetCmd.ValidArgsFunction = completeDatabaseBackupID
	backupDeleteCmd.ValidArgsFunction = completeDatabaseBackupID
}

// File-local Ref helpers

func databaseBackupRef(projectID, backupID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/backups/" + backupID)
}

func databaseBackupFromRaw(b *aruba.DBaaSBackup) *types.BackupResponse {
	if b == nil {
		return nil
	}
	return b.Raw()
}

func databaseBackupListPayload(l *aruba.List[*aruba.DBaaSBackup]) any {
	if r, ok := l.Raw().(*types.Response[types.BackupList]); ok && r != nil {
		return r.Data
	}
	return nil
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
	list, err := client.FromDatabase().Backups().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, b := range list.Items() {
			id := b.DBaaSBackupID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, b.Name()))
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

		bk := aruba.NewDBaaSBackup().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			FromDBaaS(dbaasRef(projectID, dbaasID)).
			FromDatabase(databaseRef(projectID, dbaasID, databaseName)).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		if len(tags) > 0 {
			bk.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Backups().Create(ctx, bk)
		if err != nil {
			return fmt.Errorf("creating backup: %w", apiErrFromV2(err))
		}

		resource := databaseBackupFromRaw(created)
		if resource != nil {
			fmt.Printf("\n%s\n", msgCreated("Backup", name))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if resource.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", resource.Metadata.LocationResponse.Value)
			}
			if resource.Status.State != nil {
				fmt.Printf("Status:          %s\n", *resource.Status.State)
			}
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
		got, err := client.FromDatabase().Backups().Get(ctx, databaseBackupRef(projectID, backupID))
		if err != nil {
			return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
		}

		resource := databaseBackupFromRaw(got)
		if resource != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(resource, nil, nil)
				return nil
			}

			fmt.Println("\nBackup Details:")
			fmt.Println("==============")

			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if resource.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", resource.Metadata.LocationResponse.Value)
			}
			if resource.Status.State != nil {
				fmt.Printf("Status:          %s\n", *resource.Status.State)
			}
			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
			if resource.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *resource.Metadata.CreatedBy)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", resource.Metadata.Tags)
			}
			if resource.Properties.DBaaS.URI != "" {
				fmt.Printf("DBaaS URI:       %s\n", resource.Properties.DBaaS.URI)
			}
			if resource.Properties.Database.URI != "" {
				fmt.Printf("Database URI:    %s\n", resource.Properties.Database.URI)
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
		list, err := client.FromDatabase().Backups().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing backups: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, b := range list.Items() {
				row := []string{
					b.Name(),
					b.DBaaSBackupID(),
					string(b.Region()),
					b.State(),
				}
				rows = append(rows, row)
			}
			PrintOutput(databaseBackupListPayload(list), headers, rows)
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")

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
			if _, err := client.FromDatabase().Backups().Get(ctx, databaseBackupRef(projectID, backupID)); err != nil {
				return fmt.Errorf("dry-run: database backup not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("database backup", backupID))
			return nil
		}

		if !skipConfirm {
			ok, err := confirmDelete("backup", backupID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromDatabase().Backups().Delete(ctx, databaseBackupRef(projectID, backupID)); err != nil {
			return fmt.Errorf("deleting backup: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Backup", backupID))
		return nil
	},
}
