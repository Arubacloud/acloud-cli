package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func restoreRef(projectID, backupID, restoreID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID + "/restores/" + restoreID)
}

func restoreFromRaw(r *aruba.StorageRestore) *types.StorageRestoreResponse {
	if r == nil {
		return nil
	}
	return r.Raw()
}

func restoreListPayload(l *aruba.List[*aruba.StorageRestore]) any {
	if r, ok := l.Raw().(*types.Response[types.StorageRestoreList]); ok && r != nil {
		return r.Data
	}
	return nil
}

func init() {
	storageCmd.AddCommand(storageRestoreCmd)
	storageRestoreCmd.AddCommand(storageRestoreListCmd)
	storageRestoreCmd.AddCommand(storageRestoreGetCmd)
	storageRestoreCmd.AddCommand(storageRestoreUpdateCmd)
	storageRestoreCmd.AddCommand(storageRestoreDeleteCmd)

	storageRestoreCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreCmd.Flags().String("name", "", "Name for the restore operation (required)")
	storageRestoreCmd.Flags().String("region", "ITBG-Bergamo", "Region code")
	storageRestoreCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	storageRestoreCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	storageRestoreCmd.MarkFlagRequired("name")

	storageRestoreListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	storageRestoreListCmd.Flags().Int32("offset", 0, "Number of results to skip")
	storageRestoreGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreUpdateCmd.Flags().String("name", "", "New name for the restore operation")
	storageRestoreUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	storageRestoreDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	storageRestoreDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	storageRestoreDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	storageRestoreGetCmd.ValidArgsFunction = completeRestoreID
	storageRestoreUpdateCmd.ValidArgsFunction = completeRestoreID
	storageRestoreDeleteCmd.ValidArgsFunction = completeRestoreID
	storageRestoreListCmd.ValidArgsFunction = completeBackupID
}

func completeRestoreID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeBackupID(cmd, args, toComplete)
	}
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

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
	list, err := client.FromStorage().Restores().List(ctx, backupRef(projectID, backupID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, r := range list.Items() {
			id := r.ID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, r.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var storageRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id] [volume-id]",
	Short: "Restore a block storage volume from a backup",
	Long: `Create a restore operation to copy backup data back to a block storage volume.

Both the backup and the target volume must already exist. The restore writes
backup data into the specified volume; ensure the volume is detached or otherwise
idle before starting a restore to avoid data corruption.`,
	Example: `  acloud storage restore <backup-id> <volume-id> --name my-restore`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]
		volumeID := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		verbose, _ := cmd.Flags().GetBool("verbose")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		bk, err := client.FromStorage().Backups().Get(ctx, backupRef(projectID, backupID))
		if err != nil {
			return fmt.Errorf("getting backup details: %w", apiErrFromV2(err))
		}

		target, err := client.FromStorage().Volumes().Get(ctx, volumeRef(projectID, volumeID))
		if err != nil {
			return fmt.Errorf("getting volume details: %w", apiErrFromV2(err))
		}

		if verbose {
			fmt.Println("Creating restore operation with the following parameters:")
			fmt.Printf("  Name:      %s\n", name)
			fmt.Printf("  Region:    %s\n", region)
			fmt.Printf("  Backup ID: %s\n", backupID)
			fmt.Printf("  Volume ID: %s\n", volumeID)
			if len(tags) > 0 {
				fmt.Printf("  Tags:      %v\n", tags)
			}
			fmt.Println()
		}

		rs := aruba.NewStorageRestore().
			IntoBackup(bk).
			Named(name).
			InRegion(aruba.Region(region)).
			ToVolume(target)
		if len(tags) > 0 {
			rs.ReplaceTags(tags...)
		}

		created, err := client.FromStorage().Restores().Create(ctx, rs)
		if err != nil {
			return fmt.Errorf("creating restore: %w", apiErrFromV2(err))
		}

		resource := restoreFromRaw(created)
		if resource != nil {
			fmt.Println(msgCreated("Restore operation", name))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
			if resource.Status.State != nil {
				fmt.Printf("Status:          %s\n", *resource.Status.State)
			}
		} else {
			fmt.Println(msgCreatedAsync("Restore operation", name))
		}
		return nil
	},
}

var storageRestoreListCmd = &cobra.Command{
	Use:   "list [backup-id]",
	Short: "List restore operations for a backup",
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
		list, err := client.FromStorage().Restores().List(ctx, backupRef(projectID, backupID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing restores: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, r := range list.Items() {
				rows = append(rows, []string{r.Name(), r.ID(), r.State()})
			}

			PrintOutput(restoreListPayload(list), headers, rows)
		} else {
			fmt.Println("No restores found for this backup")
		}
		return nil
	},
}

var storageRestoreGetCmd = &cobra.Command{
	Use:   "get [backup-id] [restore-id]",
	Short: "Get restore operation details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]
		restoreID := args[1]

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
		got, err := client.FromStorage().Restores().Get(ctx, restoreRef(projectID, backupID, restoreID))
		if err != nil {
			return fmt.Errorf("getting restore details: %w", apiErrFromV2(err))
		}

		restore := restoreFromRaw(got)
		if restore != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(restore, nil, nil)
				return nil
			}

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
			if restore.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", restore.Metadata.LocationResponse.Value)
			}
			if restore.Status.State != nil {
				fmt.Printf("Status:          %s\n", *restore.Status.State)
			}
			if restore.Metadata.CreationDate != nil && !restore.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", restore.Metadata.CreationDate.Format(DateLayout))
			}
			if restore.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *restore.Metadata.CreatedBy)
			}
			if len(restore.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", restore.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			fmt.Println()
		} else {
			fmt.Println("Restore operation not found")
		}
		return nil
	},
}

var storageRestoreUpdateCmd = &cobra.Command{
	Use:   "update [backup-id] [restore-id]",
	Short: "Update a restore operation (name and/or tags)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]
		restoreID := args[1]

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
		current, err := client.FromStorage().Restores().Get(ctx, restoreRef(projectID, backupID, restoreID))
		if err != nil {
			return fmt.Errorf("getting restore details: %w", apiErrFromV2(err))
		}

		if current.Raw() == nil {
			return fmt.Errorf("restore operation not found")
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}

		updated, err := client.FromStorage().Restores().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating restore: %w", apiErrFromV2(err))
		}

		resource := restoreFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("Restore operation", restoreID))
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
			fmt.Println(msgUpdatedAsync("Restore operation", restoreID))
		}
		return nil
	},
}

var storageRestoreDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id] [restore-id]",
	Short: "Delete a restore operation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupID := args[0]
		restoreID := args[1]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("restore operation", restoreID)
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
			_, err = client.FromStorage().Restores().Get(ctx, restoreRef(projectID, backupID, restoreID))
			if err != nil {
				return fmt.Errorf("dry-run: restore operation not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("restore operation", restoreID))
			return nil
		}

		if err := client.FromStorage().Restores().Delete(ctx, restoreRef(projectID, backupID, restoreID)); err != nil {
			return fmt.Errorf("deleting restore: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Restore operation", restoreID))
		return nil
	},
}
