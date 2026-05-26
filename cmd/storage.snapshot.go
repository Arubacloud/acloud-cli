package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	storageCmd.AddCommand(snapshotCmd)
	snapshotCmd.AddCommand(snapshotCreateCmd)
	snapshotCmd.AddCommand(snapshotGetCmd)
	snapshotCmd.AddCommand(snapshotUpdateCmd)
	snapshotCmd.AddCommand(snapshotDeleteCmd)
	snapshotCmd.AddCommand(snapshotListCmd)

	snapshotCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotCreateCmd.Flags().String("name", "", "Name for the snapshot (required)")
	snapshotCreateCmd.Flags().String("region", "", "Region code (required)")
	snapshotCreateCmd.Flags().String("volume-uri", "", "Source volume URI (required)")
	snapshotCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	snapshotCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	snapshotCreateCmd.MarkFlagRequired("name")
	snapshotCreateCmd.MarkFlagRequired("region")
	snapshotCreateCmd.MarkFlagRequired("volume-uri")

	snapshotGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	snapshotUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotUpdateCmd.Flags().String("name", "", "New name for the snapshot")
	snapshotUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	snapshotDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	snapshotDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	snapshotListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	snapshotListCmd.Flags().String("volume-uri", "", "Block storage volume URI (required)")
	snapshotListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	snapshotListCmd.Flags().Int32("offset", 0, "Number of results to skip")
	snapshotListCmd.MarkFlagRequired("volume-uri")

	snapshotGetCmd.ValidArgsFunction = completeSnapshotID
	snapshotUpdateCmd.ValidArgsFunction = completeSnapshotID
	snapshotDeleteCmd.ValidArgsFunction = completeSnapshotID
}

func completeSnapshotID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Snapshots().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, snap := range list.Items() {
			raw := snap.Raw()
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

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage snapshots",
	Long:  `Perform CRUD operations on snapshots in Aruba Cloud.`,
}

var snapshotCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new snapshot",
	Long: `Create a point-in-time snapshot of an existing block storage volume.

The volume URI is required. Snapshots can later be used to create new volumes
or restore data with 'acloud storage blockstorage create --snapshot-uri'.`,
	Example: `  acloud storage snapshot create --name my-snap --region IT-BG \
    --volume-uri /projects/<proj-id>/providers/Aruba.Storage/blockstorages/<vol-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		volumeURI, _ := cmd.Flags().GetString("volume-uri")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		snap := aruba.NewSnapshot().
			InProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			FromVolume(aruba.URI(volumeURI)).
			RetaggedAs(tags...)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromStorage().Snapshots().Create(ctx, snap)
		if err != nil {
			return fmt.Errorf("creating snapshot: %w", apiErrFromV2(err))
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
			fmt.Println(msgCreatedAsync("Snapshot", name))
		}
		return nil
	},
}

var snapshotGetCmd = &cobra.Command{
	Use:   "get [snapshot-id]",
	Short: "Get snapshot details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

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
		snap, err := client.FromStorage().Snapshots().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/snapshots/"+snapshotID))
		if err != nil {
			return fmt.Errorf("getting snapshot: %w", apiErrFromV2(err))
		}

		if snap != nil && snap.Raw() != nil {
			raw := snap.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nSnapshot Details:")
			fmt.Println("=================")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Properties.SizeGB != nil {
				fmt.Printf("Size (GB):       %d\n", *raw.Properties.SizeGB)
			}
			if raw.Properties.Volume != nil && raw.Properties.Volume.URI != nil {
				fmt.Printf("Source Volume:   %s\n", *raw.Properties.Volume.URI)
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
			fmt.Println("Snapshot not found")
		}
		return nil
	},
}

var snapshotUpdateCmd = &cobra.Command{
	Use:   "update [snapshot-id]",
	Short: "Update a snapshot (name and/or tags only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

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
		snap, err := client.FromStorage().Snapshots().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/snapshots/"+snapshotID))
		if err != nil {
			return fmt.Errorf("getting snapshot: %w", apiErrFromV2(err))
		}
		if snap == nil || snap.Raw() == nil {
			return fmt.Errorf("snapshot not found")
		}

		if name != "" {
			snap.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			snap.RetaggedAs(tags...)
		}

		updated, err := client.FromStorage().Snapshots().Update(ctx, snap)
		if err != nil {
			return fmt.Errorf("updating snapshot: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("Snapshot", snapshotID))
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
			fmt.Println(msgUpdatedAsync("Snapshot", snapshotID))
		}
		return nil
	},
}

var snapshotDeleteCmd = &cobra.Command{
	Use:   "delete [snapshot-id]",
	Short: "Delete a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		snapshotID := args[0]

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			ok, err := confirmDelete("snapshot", snapshotID)
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
			_, err = client.FromStorage().Snapshots().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/snapshots/"+snapshotID))
			if err != nil {
				return fmt.Errorf("dry-run: snapshot not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("snapshot", snapshotID))
			return nil
		}

		err = client.FromStorage().Snapshots().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/snapshots/"+snapshotID))
		if err != nil {
			return fmt.Errorf("deleting snapshot: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Snapshot", snapshotID))
		return nil
	},
}

var snapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List snapshots for a block storage volume",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		volumeURI, _ := cmd.Flags().GetString("volume-uri")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromStorage().Snapshots().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing snapshots: %w", apiErrFromV2(err))
		}

		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "SIZE(GB)", Width: 12},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		if list != nil {
			for _, snap := range list.Items() {
				if snap.VolumeURI() != volumeURI {
					continue
				}
				raw := snap.Raw()
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
				size := "0"
				if raw.Properties.SizeGB != nil {
					size = fmt.Sprintf("%d", *raw.Properties.SizeGB)
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, size, status})
			}
		}

		if len(rows) == 0 {
			fmt.Printf("No snapshots found for volume: %s\n", volumeURI)
			return nil
		}

		PrintOutput(list, headers, rows)
		return nil
	},
}
