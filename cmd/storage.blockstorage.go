package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func volumeRef(projectID, volumeID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Storage/blockStorages/" + volumeID)
}

func init() {
	storageCmd.AddCommand(blockstorageCmd)
	blockstorageCmd.AddCommand(blockstorageCreateCmd)
	blockstorageCmd.AddCommand(blockstorageGetCmd)
	blockstorageCmd.AddCommand(blockstorageUpdateCmd)
	blockstorageCmd.AddCommand(blockstorageDeleteCmd)
	blockstorageCmd.AddCommand(blockstorageListCmd)

	blockstorageCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageCreateCmd.Flags().String("name", "", "Name for the block storage (required)")
	blockstorageCreateCmd.Flags().String("region", "ITBG-Bergamo", "Region code (required)")
	blockstorageCreateCmd.Flags().String("zone", "", "Zone/datacenter (optional, only for zonal block storage)")
	blockstorageCreateCmd.Flags().Int("size", 0, "Size in GB (required)")
	blockstorageCreateCmd.Flags().String("type", "Standard", "Type: Standard or Performance")
	blockstorageCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	blockstorageCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	blockstorageCreateCmd.Flags().String("snapshot-uri", "", "URI of the snapshot to use (optional)")
	blockstorageCreateCmd.Flags().Bool("set-bootable", false, "Set block storage as bootable (optional)")
	blockstorageCreateCmd.Flags().String("image", "", "Image string to use for the block storage (optional)")
	blockstorageCreateCmd.MarkFlagRequired("name")
	blockstorageCreateCmd.MarkFlagRequired("region")
	blockstorageCreateCmd.MarkFlagRequired("size")

	blockstorageGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	blockstorageUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageUpdateCmd.Flags().String("name", "", "New name for the block storage")
	blockstorageUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	blockstorageDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	blockstorageDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	blockstorageListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	blockstorageListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	blockstorageListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	blockstorageGetCmd.ValidArgsFunction = completeBlockStorageID
	blockstorageUpdateCmd.ValidArgsFunction = completeBlockStorageID
	blockstorageDeleteCmd.ValidArgsFunction = completeBlockStorageID
}

func completeBlockStorageID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromStorage().Volumes().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vol := range list.Items() {
			raw := vol.Raw()
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

var blockstorageCmd = &cobra.Command{
	Use:   "blockstorage",
	Short: "Manage block storage",
	Long:  `Perform CRUD operations on block storage in Aruba Cloud.`,
}

var blockstorageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new block storage",
	Long: `Create a new block storage volume in the specified region.

Volume size is specified in GB with --size. Type options: Standard or Performance.
The volume can be initialised from a snapshot with --snapshot-uri, or from an
image with --image. Pass --set-bootable to mark the volume as a boot disk.

Billing period: Hour (default), Month, or Year.`,
	Example: `  # Create an empty 50 GB Standard volume
  acloud storage blockstorage create --name my-volume --size 50 --region IT-BG

  # Create a Performance volume from a snapshot
  acloud storage blockstorage create --name fast-vol --size 100 --region IT-BG \
    --type Performance \
    --snapshot-uri /projects/<proj-id>/providers/Aruba.Storage/snapshots/<snap-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		zone, _ := cmd.Flags().GetString("zone")
		size, _ := cmd.Flags().GetInt("size")
		volumeType, _ := cmd.Flags().GetString("type")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		snapshotURI, _ := cmd.Flags().GetString("snapshot-uri")
		setBootable, _ := cmd.Flags().GetBool("set-bootable")
		image, _ := cmd.Flags().GetString("image")

		if size <= 0 {
			return fmt.Errorf("--size must be greater than 0")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		vol := aruba.NewBlockStorage().
			InProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			SizedGB(size).
			OfType(aruba.BlockStorageType(volumeType)).
			BilledBy(aruba.BillingPeriod(billingPeriod)).
			RetaggedAs(tags...)

		if zone != "" {
			vol.InZone(aruba.Zone(zone))
		}
		if snapshotURI != "" {
			vol.FromSnapshot(aruba.URI(snapshotURI))
		}
		if setBootable {
			vol.AsBootable()
		}
		if image != "" {
			vol.FromImage(image)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromStorage().Volumes().Create(ctx, vol)
		if err != nil {
			return fmt.Errorf("creating block storage: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "SIZE(GB)", Width: 12},
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
			sizeVal := fmt.Sprintf("%d", raw.Properties.SizeGB)
			typeVal := string(raw.Properties.Type)
			statusVal := ""
			if raw.Status.State != nil {
				statusVal = string(*raw.Status.State)
			}
			PrintOutput(raw, headers, [][]string{{id, nameVal, sizeVal, typeVal, statusVal}})
		} else {
			fmt.Println(msgCreatedAsync("Block storage", name))
		}
		return nil
	},
}

var blockstorageGetCmd = &cobra.Command{
	Use:   "get [volume-id]",
	Short: "Get block storage details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

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
		vol, err := client.FromStorage().Volumes().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/blockStorages/"+volumeID))
		if err != nil {
			return fmt.Errorf("getting block storage: %w", apiErrFromV2(err))
		}

		if vol != nil && vol.Raw() != nil {
			raw := vol.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nBlock Storage Details:")
			fmt.Println("======================")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			fmt.Printf("Size (GB):       %d\n", raw.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", string(raw.Properties.Type))
			fmt.Printf("Zone:            %s\n", string(raw.Properties.Zone))
			if raw.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
			}
			if raw.Properties.Bootable != nil {
				fmt.Printf("Bootable:        %t\n", *raw.Properties.Bootable)
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
			fmt.Println("Block storage not found")
		}
		return nil
	},
}

var blockstorageUpdateCmd = &cobra.Command{
	Use:   "update [volume-id]",
	Short: "Update block storage (name and/or tags)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

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
		vol, err := client.FromStorage().Volumes().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/blockStorages/"+volumeID))
		if err != nil {
			return fmt.Errorf("getting block storage: %w", apiErrFromV2(err))
		}
		if vol == nil || vol.Raw() == nil {
			return fmt.Errorf("block storage not found")
		}

		if name != "" {
			vol.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			vol.RetaggedAs(tags...)
		}

		updated, err := client.FromStorage().Volumes().Update(ctx, vol)
		if err != nil {
			return fmt.Errorf("updating block storage: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("Block storage", volumeID))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
			}
			fmt.Printf("Size (GB):       %d\n", raw.Properties.SizeGB)
			fmt.Printf("Type:            %s\n", string(raw.Properties.Type))
		} else {
			fmt.Println(msgUpdatedAsync("Block storage", volumeID))
		}
		return nil
	},
}

var blockstorageDeleteCmd = &cobra.Command{
	Use:   "delete [volume-id]",
	Short: "Delete block storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volumeID := args[0]

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			ok, err := confirmDelete("block storage", volumeID)
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
			_, err = client.FromStorage().Volumes().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/blockStorages/"+volumeID))
			if err != nil {
				return fmt.Errorf("dry-run: block storage not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("block storage", volumeID))
			return nil
		}

		err = client.FromStorage().Volumes().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Storage/blockStorages/"+volumeID))
		if err != nil {
			return fmt.Errorf("deleting block storage: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Block storage", volumeID))
		return nil
	},
}

var blockstorageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all block storage",
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
		list, err := client.FromStorage().Volumes().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing block storage: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "SIZE(GB)", Width: 12},
				{Header: "REGION", Width: 15},
				{Header: "ZONE", Width: 15},
				{Header: "TYPE", Width: 15},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, vol := range list.Items() {
				raw := vol.Raw()
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
				size := fmt.Sprintf("%d", raw.Properties.SizeGB)
				region := ""
				if raw.Metadata.LocationResponse != nil {
					region = string(raw.Metadata.LocationResponse.Value)
				}
				zone := string(raw.Properties.Zone)
				volumeType := string(raw.Properties.Type)
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, size, region, zone, volumeType, status})
			}

			PrintOutput(list, headers, rows)
		} else {
			fmt.Println("No block storage found")
		}
		return nil
	},
}
