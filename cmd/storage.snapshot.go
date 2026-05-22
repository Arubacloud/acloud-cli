package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func snapshotRef(projectID, snapshotID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/snapshots/" + snapshotID)
}

func snapshotFromRaw(s *aruba.Snapshot) *types.SnapshotResponse {
	if s == nil {
		return nil
	}
	return s.Raw()
}

func snapshotListPayload(l *aruba.List[*aruba.Snapshot]) any {
	if r, ok := l.Raw().(*types.Response[types.SnapshotList]); ok && r != nil {
		return r.Data
	}
	return nil
}

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
	snapshotListCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
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
	list, err := client.FromStorage().Snapshots().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, s := range list.Items() {
			id := s.ID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, s.Name()))
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
    --volume-uri /projects/<proj-id>/providers/Aruba.Storage/blockStorages/<vol-id>`,
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
		verbose, _ := cmd.Flags().GetBool("verbose")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		if verbose {
			fmt.Println("Creating snapshot with the following parameters:")
			fmt.Printf("  Name:       %s\n", name)
			fmt.Printf("  Region:     %s\n", region)
			fmt.Printf("  Volume URI: %s\n", volumeURI)
			fmt.Printf("  Tags:       %v\n", tags)
			fmt.Println()
		}

		snap := aruba.NewSnapshot().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			FromVolume(aruba.URI(volumeURI))
		if len(tags) > 0 {
			snap.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromStorage().Snapshots().Create(ctx, snap)
		if err != nil {
			return fmt.Errorf("creating snapshot: %w", apiErrFromV2(err))
		}

		resource := snapshotFromRaw(created)
		if resource != nil {
			fmt.Printf("\n%s\n", msgCreated("Snapshot", name))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
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
		got, err := client.FromStorage().Snapshots().Get(ctx, snapshotRef(projectID, snapshotID))
		if err != nil {
			return fmt.Errorf("getting snapshot details: %w", apiErrFromV2(err))
		}

		snapshot := snapshotFromRaw(got)
		if snapshot != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(snapshot, nil, nil)
				return nil
			}

			fmt.Println("\nSnapshot Details:")
			fmt.Println("=================")

			if snapshot.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *snapshot.Metadata.ID)
			}
			if snapshot.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *snapshot.Metadata.URI)
			}
			if snapshot.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *snapshot.Metadata.Name)
			}
			if snapshot.Properties.SizeGB != nil {
				fmt.Printf("Size (GB):       %d\n", *snapshot.Properties.SizeGB)
			}
			if snapshot.Properties.Volume != nil && snapshot.Properties.Volume.URI != nil {
				fmt.Printf("Source Volume:   %s\n", *snapshot.Properties.Volume.URI)
			}
			if snapshot.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", snapshot.Metadata.LocationResponse.Value)
			}
			status := ""
			if snapshot.Status.State != nil {
				status = *snapshot.Status.State
			}
			fmt.Printf("Status:          %s\n", status)
			if snapshot.Metadata.CreationDate != nil && !snapshot.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", snapshot.Metadata.CreationDate.Format(DateLayout))
			}
			if snapshot.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *snapshot.Metadata.CreatedBy)
			}
			if len(snapshot.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", snapshot.Metadata.Tags)
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
		current, err := client.FromStorage().Snapshots().Get(ctx, snapshotRef(projectID, snapshotID))
		if err != nil {
			return fmt.Errorf("getting snapshot details: %w", apiErrFromV2(err))
		}

		if current.Raw() == nil {
			return fmt.Errorf("snapshot not found")
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}

		updated, err := client.FromStorage().Snapshots().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating snapshot: %w", apiErrFromV2(err))
		}

		resource := snapshotFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("Snapshot", snapshotID))
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
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
			_, err = client.FromStorage().Snapshots().Get(ctx, snapshotRef(projectID, snapshotID))
			if err != nil {
				return fmt.Errorf("dry-run: snapshot not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("snapshot", snapshotID))
			return nil
		}

		if err := client.FromStorage().Snapshots().Delete(ctx, snapshotRef(projectID, snapshotID)); err != nil {
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
		list, err := client.FromStorage().Snapshots().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing snapshots: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			var filtered []*aruba.Snapshot
			for _, s := range list.Items() {
				if s.VolumeURI() == volumeURI {
					filtered = append(filtered, s)
				}
			}

			if len(filtered) == 0 {
				fmt.Printf("No snapshots found for volume: %s\n", volumeURI)
				return nil
			}

			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "SIZE(GB)", Width: 12},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, s := range filtered {
				sizeStr := "0"
				r := snapshotFromRaw(s)
				if r != nil && r.Properties.SizeGB != nil {
					sizeStr = fmt.Sprintf("%d", *r.Properties.SizeGB)
				}
				rows = append(rows, []string{s.Name(), s.ID(), sizeStr, s.State()})
			}

			PrintOutput(snapshotListPayload(list), headers, rows)
		} else {
			fmt.Println("No snapshots found")
		}
		return nil
	},
}
