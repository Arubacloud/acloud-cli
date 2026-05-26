package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	databaseCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupGetCmd)
	backupCmd.AddCommand(backupDeleteCmd)
	backupCmd.AddCommand(backupListCmd)

	backupCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	backupCreateCmd.Flags().String("name", "", "Backup name (required)")
	backupCreateCmd.Flags().String("region", "", "Region code (required)")
	backupCreateCmd.Flags().String("dbaas-id", "", "DBaaS instance ID (required)")
	backupCreateCmd.Flags().String("database-name", "", "Database name (required)")
	backupCreateCmd.Flags().String("zone", "", "Availability zone (e.g. ITBG-1); defaults to region if unset")
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

	backupGetCmd.ValidArgsFunction = completeDatabaseBackupID
	backupDeleteCmd.ValidArgsFunction = completeDatabaseBackupID
}

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
	list, err := client.FromDatabase().Backups().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, backup := range list.Items() {
			raw := backup.Raw()
			if raw != nil && raw.Metadata.ID != nil && raw.Metadata.Name != nil {
				id := *raw.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *raw.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

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

		dbaasURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID
		databaseURI := dbaasURI + "/databases/" + databaseName

		bkp := aruba.NewDBaaSBackup().
			InProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			FromDBaaS(aruba.URI(dbaasURI)).
			FromDatabase(aruba.URI(databaseURI)).
			BilledBy(aruba.BillingPeriod(billingPeriod)).
			RetaggedAs(tags...)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Backups().Create(ctx, bkp)
		if err != nil {
			return fmt.Errorf("creating backup: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "REGION", Width: 20},
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
			regionVal := ""
			if raw.Metadata.LocationResponse != nil {
				regionVal = string(raw.Metadata.LocationResponse.Value)
			}
			statusVal := ""
			if raw.Status.State != nil {
				statusVal = string(*raw.Status.State)
			}
			PrintOutput(raw, headers, [][]string{{id, nameVal, regionVal, statusVal}})
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
		backup, err := client.FromDatabase().Backups().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/backups/"+backupID))
		if err != nil {
			return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
		}

		if backup != nil && backup.Raw() != nil {
			raw := backup.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nBackup Details:")
			fmt.Println("==============")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
			}
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", string(*raw.Status.State))
			}
			if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", raw.Metadata.CreationDate.Format(DateLayout))
			}
			if raw.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
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
		list, err := client.FromDatabase().Backups().List(ctx, aruba.URI("/projects/"+projectID))
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
			for _, backup := range list.Items() {
				raw := backup.Raw()
				if raw == nil {
					continue
				}
				name := ""
				if raw.Metadata.Name != nil {
					name = *raw.Metadata.Name
				}
				id := ""
				if raw.Metadata.ID != nil {
					id = *raw.Metadata.ID
				}
				region := ""
				if raw.Metadata.LocationResponse != nil {
					region = string(raw.Metadata.LocationResponse.Value)
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, region, status})
			}
			PrintOutput(list, headers, rows)
		} else {
			fmt.Println("No backups found")
		}
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
			_, err = client.FromDatabase().Backups().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/backups/"+backupID))
			if err != nil {
				return fmt.Errorf("dry-run: database backup not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("database backup", backupID))
			return nil
		}

		err = client.FromDatabase().Backups().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/backups/"+backupID))
		if err != nil {
			return fmt.Errorf("deleting backup: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Backup", backupID))
		return nil
	},
}
