package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func backupRef(projectID, backupID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID)
}

func backupFromRaw(b *aruba.StorageBackup) *types.StorageBackupResponse {
	if b == nil {
		return nil
	}
	return b.Raw()
}

func backupListPayload(l *aruba.List[*aruba.StorageBackup]) any {
	if r, ok := l.Raw().(*types.Response[types.StorageBackupList]); ok && r != nil {
		return r.Data
	}
	return nil
}

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
	list, err := client.FromStorage().Backups().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, b := range list.Items() {
			id := b.ID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, b.Name()))
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
		verbose, _ := cmd.Flags().GetBool("verbose")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		vol, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
		if err != nil {
			return fmt.Errorf("getting volume details: %w", apiErrFromV2(err))
		}

		if verbose {
			fmt.Println("Creating storage backup with the following parameters:")
			fmt.Printf("  Name:           %s\n", name)
			fmt.Printf("  Type:           %s\n", backupType)
			fmt.Printf("  Region:         %s\n", region)
			fmt.Printf("  Volume ID:      %s\n", volumeID)
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

		bk := aruba.NewStorageBackup().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfType(aruba.StorageBackupType(backupType)).
			FromVolume(vol)
		if retentionDays > 0 {
			bk.WithRetentionDays(retentionDays)
		}
		if billingPeriod != "" {
			bk.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if len(tags) > 0 {
			bk.ReplaceTags(tags...)
		}

		created, err := client.FromStorage().Backups().Create(ctx, bk)
		if err != nil {
			return fmt.Errorf("creating backup: %w", apiErrFromV2(err))
		}

		resource := backupFromRaw(created)
		if resource != nil {
			fmt.Println(msgCreated("Storage backup", name))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			fmt.Printf("Type:            %s\n", resource.Properties.Type)
			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
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
		list, err := client.FromStorage().Backups().List(ctx, projectRef(projectID), listOpts(cmd)...)
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
			for _, b := range list.Items() {
				backupTypeStr := string(b.Type())
				rows = append(rows, []string{b.Name(), b.ID(), backupTypeStr, b.State()})
			}

			PrintOutput(backupListPayload(list), headers, rows)
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
		got, err := client.FromStorage().Backups().Get(ctx, backupRef(projectID, backupID))
		if err != nil {
			return fmt.Errorf("getting backup details: %w", apiErrFromV2(err))
		}

		backup := backupFromRaw(got)
		if backup != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(backup, nil, nil)
				return nil
			}

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
			if backup.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", backup.Metadata.LocationResponse.Value)
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
		current, err := client.FromStorage().Backups().Get(ctx, backupRef(projectID, backupID))
		if err != nil {
			return fmt.Errorf("getting backup details: %w", apiErrFromV2(err))
		}

		if current.Raw() == nil {
			return fmt.Errorf("backup not found")
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}

		updated, err := client.FromStorage().Backups().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating backup: %w", apiErrFromV2(err))
		}

		resource := backupFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("Backup", backupID))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", resource.Metadata.Tags)
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
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
			_, err = client.FromStorage().Backups().Get(ctx, backupRef(projectID, backupID))
			if err != nil {
				return fmt.Errorf("dry-run: storage backup not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("storage backup", backupID))
			return nil
		}

		if err := client.FromStorage().Backups().Delete(ctx, backupRef(projectID, backupID)); err != nil {
			return fmt.Errorf("deleting backup: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Backup", backupID))
		return nil
	},
}
