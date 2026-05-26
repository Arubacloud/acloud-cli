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
	list, err := client.FromStorage().Restores().List(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, r := range list.Items() {
			raw := r.Raw()
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

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		backupURI := "/projects/" + projectID + "/providers/Aruba.Storage/backups/" + backupID
		volumeURI := "/projects/" + projectID + "/providers/Aruba.Storage/blockstorages/" + volumeID

		ctx, cancel := newCtx()
		defer cancel()

		_, err = client.FromStorage().Backups().Get(ctx, aruba.URI(backupURI))
		if err != nil {
			return fmt.Errorf("getting backup: %w", apiErrFromV2(err))
		}

		_, err = client.FromStorage().Volumes().Get(ctx, aruba.URI(volumeURI))
		if err != nil {
			return fmt.Errorf("getting volume: %w", apiErrFromV2(err))
		}

		restore := aruba.NewStorageRestore().
			IntoBackup(aruba.URI(backupURI)).
			Named(name).
			InRegion(aruba.Region(region)).
			ToVolume(aruba.URI(volumeURI)).
			ReplaceTags(tags...)

		created, err := client.FromStorage().Restores().Create(ctx, restore)
		if err != nil {
			return fmt.Errorf("creating restore: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
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
			statusVal := ""
			if raw.Status.State != nil {
				statusVal = string(*raw.Status.State)
			}
			PrintOutput(raw, headers, [][]string{{id, nameVal, statusVal}})
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
		list, err := client.FromStorage().Restores().List(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID))
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
				raw := r.Raw()
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
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, status})
			}

			PrintOutput(list.Raw(), headers, rows)
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
		restore, err := client.FromStorage().Restores().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID+"/restores/"+restoreID))
		if err != nil {
			return fmt.Errorf("getting restore: %w", apiErrFromV2(err))
		}

		if restore != nil && restore.Raw() != nil {
			raw := restore.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nRestore Operation Details:")
			fmt.Println("==========================")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Properties.Destination.URI != "" {
				fmt.Printf("Target Volume:   %s\n", raw.Properties.Destination.URI)
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
		restore, err := client.FromStorage().Restores().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID+"/restores/"+restoreID))
		if err != nil {
			return fmt.Errorf("getting restore: %w", apiErrFromV2(err))
		}
		if restore == nil || restore.Raw() == nil {
			return fmt.Errorf("restore operation not found")
		}

		if name != "" {
			restore.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			restore.ReplaceTags(tags...)
		}

		updated, err := client.FromStorage().Restores().Update(ctx, restore)
		if err != nil {
			return fmt.Errorf("updating restore: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("Restore operation", restoreID))
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

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
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
			_, err = client.FromStorage().Restores().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID+"/restores/"+restoreID))
			if err != nil {
				return fmt.Errorf("dry-run: restore operation not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("restore operation", restoreID))
			return nil
		}

		err = client.FromStorage().Restores().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/backups/"+backupID+"/restores/"+restoreID))
		if err != nil {
			return fmt.Errorf("deleting restore: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Restore operation", restoreID))
		return nil
	},
}
