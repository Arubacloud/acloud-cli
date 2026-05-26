package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	storageCmd.AddCommand(storageBackupCmd)
	storageBackupCmd.AddCommand(storageBackupListCmd)
	storageBackupCmd.AddCommand(storageBackupGetCmd)
	storageBackupCmd.AddCommand(storageBackupUpdateCmd)
	storageBackupCmd.AddCommand(storageBackupDeleteCmd)

	storageBackupCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupCmd.Flags().String("name", "", "Name for the backup (required)")
	storageBackupCmd.Flags().String("region", "ITBG-Bergamo", "Region code")
	storageBackupCmd.Flags().String("type", "Full", "Backup type: Full or Incremental")
	storageBackupCmd.Flags().Int("retention-days", 0, "Number of days to retain the backup")
	storageBackupCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")
	storageBackupCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	storageBackupCmd.MarkFlagRequired("name")

	storageBackupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	storageBackupListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	storageBackupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	storageBackupUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupUpdateCmd.Flags().String("name", "", "New name for the backup")
	storageBackupUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	storageBackupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageBackupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	storageBackupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	storageBackupGetCmd.ValidArgsFunction = completeBackupID
	storageBackupUpdateCmd.ValidArgsFunction = completeBackupID
	storageBackupDeleteCmd.ValidArgsFunction = completeBackupID
}

func completeBackupID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Backups().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, bkp := range list.Items() {
			raw := bkp.Raw()
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

var storageBackupCmd = &cobra.Command{
	Use:   "backup [volume-id]",
	Short: "Create a storage backup of a block storage volume",
	Long: `Create a backup of the specified block storage volume.

Backup types: Full or Incremental (default).
Use --retention-days to set how many days the backup is retained.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud storage backup <volume-id> --name my-backup
  acloud storage backup <volume-id> --name weekly-full --type Full --retention-days 30`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		backupType, _ := cmd.Flags().GetString("type")
		retentionDays, _ := cmd.Flags().GetInt("retention-days")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		volumeURI := "/projects/" + projectID + "/providers/Aruba.Storage/blockstorages/" + volumeID

		ctx, cancel := newCtx()
		defer cancel()

		_, err = client.FromStorage().Volumes().Get(ctx, aruba.URI(volumeURI))
		if err != nil {
			return fmt.Errorf("getting volume: %w", apiErrFromV2(err))
		}

		bkp := aruba.NewStorageBackup().
			IntoProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfType(aruba.StorageBackupType(backupType)).
			FromVolume(aruba.URI(volumeURI)).
			ReplaceTags(tags...)

		if retentionDays > 0 {
			bkp.WithRetentionDays(retentionDays)
		}
		if billingPeriod != "" {
			bkp.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}

		created, err := client.FromStorage().Backups().Create(ctx, bkp)
		if err != nil {
			return fmt.Errorf("creating backup: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "TYPE", Width: 15},
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
			typeVal := string(raw.Properties.Type)
			statusVal := ""
			if raw.Status.State != nil {
				statusVal = string(*raw.Status.State)
			}
			PrintOutput(raw, headers, [][]string{{id, nameVal, typeVal, statusVal}})
		} else {
			fmt.Println(msgCreatedAsync("Storage backup", name))
		}
		return nil
	},
}

var storageBackupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage backups",
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
		list, err := client.FromStorage().Backups().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing backups: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "TYPE", Width: 12},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, bkp := range list.Items() {
				raw := bkp.Raw()
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
				backupType := string(raw.Properties.Type)
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, backupType, status})
			}

			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No backups found")
		}
		return nil
	},
}

var storageBackupGetCmd = &cobra.Command{
	Use:   "get [backup-id]",
	Short: "Get storage backup details",
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
		bkp, err := client.FromStorage().Backups().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID))
		if err != nil {
			return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
		}

		if bkp != nil && bkp.Raw() != nil {
			raw := bkp.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nStorage Backup Details:")
			fmt.Println("=======================")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			fmt.Printf("Type:            %s\n", string(raw.Properties.Type))
			if raw.Properties.Origin.URI != "" {
				fmt.Printf("Source Volume:   %s\n", raw.Properties.Origin.URI)
			}
			if raw.Properties.RetentionDays != nil {
				fmt.Printf("Retention Days:  %d\n", *raw.Properties.RetentionDays)
			}
			if raw.Properties.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", string(*raw.Properties.BillingPeriod))
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
		} else {
			fmt.Println("Backup not found")
		}
		return nil
	},
}

var storageBackupUpdateCmd = &cobra.Command{
	Use:   "update [backup-id]",
	Short: "Update a storage backup (name and/or tags)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		bkp, err := client.FromStorage().Backups().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID))
		if err != nil {
			return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
		}
		if bkp == nil || bkp.Raw() == nil {
			return fmt.Errorf("backup not found")
		}

		if name != "" {
			bkp.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			bkp.ReplaceTags(tags...)
		}

		updated, err := client.FromStorage().Backups().Update(ctx, bkp)
		if err != nil {
			return fmt.Errorf("updating backup: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("Backup", backupID))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("Backup", backupID))
		}
		return nil
	},
}

var storageBackupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a storage backup",
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
			_, err = client.FromStorage().Backups().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID))
			if err != nil {
				return fmt.Errorf("dry-run: storage backup not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("storage backup", backupID))
			return nil
		}

		err = client.FromStorage().Backups().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID))
		if err != nil {
			return fmt.Errorf("deleting backup: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Backup", backupID))
		return nil
	},
}
