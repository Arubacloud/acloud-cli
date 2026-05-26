package cmd

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func vpcPeeringListPayload(l *aruba.List[*aruba.VPCPeering]) any {
	if r, ok := l.Raw().(*types.Response[types.VPCPeeringList]); ok && r != nil {
		return r.Data
	}
	return nil
}

func init() {
	// Peering
	networkCmd.AddCommand(vpcpeeringCmd)

	vpcpeeringCmd.AddCommand(vpcpeeringCreateCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringGetCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringUpdateCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringDeleteCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringListCmd)

	// VPC Peering flags
	vpcpeeringCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringCreateCmd.Flags().String("name", "", "VPC Peering name (required)")
	vpcpeeringCreateCmd.Flags().String("peer-vpc-id", "", "Peer VPC ID or URI (required)")
	vpcpeeringCreateCmd.Flags().String("region", "", "Region code (e.g., ITBG-Bergamo) (required)")
	vpcpeeringCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcpeeringCreateCmd.MarkFlagRequired("name")
	vpcpeeringCreateCmd.MarkFlagRequired("peer-vpc-id")
	vpcpeeringCreateCmd.MarkFlagRequired("region")

	vpcpeeringGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpcpeeringUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringUpdateCmd.Flags().String("name", "", "New name for the VPC peering")
	vpcpeeringUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	vpcpeeringDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcpeeringDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	vpcpeeringListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpcpeeringListCmd.Flags().Int32("offset", 0, "Number of results to skip")
}

// Peering subcommands
var vpcpeeringCmd = &cobra.Command{
	Use:   "vpcpeering",
	Short: "Manage VPC peering",
	Long:  `Perform CRUD operations on VPC peering in Aruba Cloud.`,
}

var vpcpeeringCreateCmd = &cobra.Command{
	Use:   "create [vpc-id]",
	Short: "Create a new VPC peering",
	Long: `Create a VPC peering connection between two VPCs.

Provide the ID of the peer VPC with --peer-vpc-id. Both VPCs must be in the
same region. Add routes via 'acloud network vpcpeeringroute create' after the
peering is established.`,
	Example: `  acloud network vpcpeering create <vpc-id> \
    --name my-peering --region IT-BG --peer-vpc-id <peer-vpc-id>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		name, _ := cmd.Flags().GetString("name")
		peerVPCID, _ := cmd.Flags().GetString("peer-vpc-id")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
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

		peering := aruba.NewVPCPeering().
			IntoVPC(aruba.VPCRef(projectID, vpcID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithRemoteVPC(aruba.URI(peerVPCID)).
			ReplaceTags(tags...)

		resp, err := client.FromNetwork().VPCPeerings().Create(ctx, peering)
		if err != nil {
			return fmt.Errorf("creating VPC peering: %w", apiErrFromV2(err))
		}
		if resp != nil && resp.Raw() != nil {
			raw := resp.Raw()
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			peerVPC := ""
			if raw.Properties.RemoteVPC != nil {
				peerVPC = raw.Properties.RemoteVPC.URI
			}
			regionVal := ""
			if raw.Metadata.LocationResponse != nil {
				regionVal = string(raw.Metadata.LocationResponse.Value)
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			row := []string{name, id, peerVPC, regionVal, status}
			PrintOutput(raw, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("VPC peering", name))
		}
		return nil
	},
}

var vpcpeeringGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [peering-id]",
	Short: "Get VPC peering details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		peeringID := args[1]
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
		peering, err := client.FromNetwork().VPCPeerings().Get(ctx, aruba.VPCPeeringRef(projectID, vpcID, peeringID))
		if err != nil {
			return fmt.Errorf("getting VPC peering: %w", apiErrFromV2(err))
		}
		if peering != nil && peering.Raw() != nil {
			raw := peering.Raw()
			fmt.Println("\nVPC Peering Details:")
			fmt.Println("====================")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Properties.RemoteVPC != nil {
				fmt.Printf("Peer VPC:        %s\n", raw.Properties.RemoteVPC.URI)
			}
			if raw.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
			}
			if raw.Metadata.CreationDate != nil {
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
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", *raw.Status.State)
			}
		} else {
			fmt.Println("VPC peering not found or no data returned.")
		}
		return nil
	},
}

var vpcpeeringListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List VPC peerings for a VPC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
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
		list, err := client.FromNetwork().VPCPeerings().List(ctx, aruba.VPCRef(projectID, vpcID))
		if err != nil {
			return fmt.Errorf("listing VPC peerings: %w", apiErrFromV2(err))
		}
		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, peering := range list.Items() {
				raw := peering.Raw()
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
				peerVPC := ""
				if raw.Properties.RemoteVPC != nil {
					peerVPC = raw.Properties.RemoteVPC.URI
				}
				region := ""
				if raw.Metadata.LocationResponse != nil {
					region = string(raw.Metadata.LocationResponse.Value)
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, peerVPC, region, status})
			}
			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No VPC peerings found.")
		}
		return nil
	},
}

var vpcpeeringUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [peering-id]",
	Short: "Update a VPC peering",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		peeringID := args[1]
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
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

		peering, err := client.FromNetwork().VPCPeerings().Get(ctx, aruba.VPCPeeringRef(projectID, vpcID, peeringID))
		if err != nil || peering == nil || peering.Raw() == nil {
			return fmt.Errorf("fetching current VPC peering: %w", err)
		}

		if peering.Raw().Status.State != nil && *peering.Raw().Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC peering while it is in 'InCreation' state. Please wait until the VPC peering is fully created")
		}

		if name != "" {
			peering.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			peering.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().VPCPeerings().Update(ctx, peering)
		if err != nil {
			return fmt.Errorf("updating VPC peering: %w", apiErrFromV2(err))
		}
		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			nameVal := ""
			if raw.Metadata.Name != nil {
				nameVal = *raw.Metadata.Name
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			peerVPC := ""
			if raw.Properties.RemoteVPC != nil {
				peerVPC = raw.Properties.RemoteVPC.URI
			}
			regionVal := ""
			if raw.Metadata.LocationResponse != nil {
				regionVal = string(raw.Metadata.LocationResponse.Value)
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			row := []string{nameVal, id, peerVPC, regionVal, status}
			PrintOutput(raw, headers, [][]string{row})
		} else {
			fmt.Println(msgUpdatedAsync("VPC peering", peeringID))
		}
		return nil
	},
}

var vpcpeeringDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [peering-id]",
	Short: "Delete a VPC peering",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		peeringID := args[1]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("VPC peering", peeringID)
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
			_, err = client.FromNetwork().VPCPeerings().Get(ctx, aruba.VPCPeeringRef(projectID, vpcID, peeringID))
			if err != nil {
				return fmt.Errorf("dry-run: VPC peering not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPC peering", peeringID))
			return nil
		}

		err = client.FromNetwork().VPCPeerings().Delete(ctx, aruba.VPCPeeringRef(projectID, vpcID, peeringID))
		if err != nil {
			return fmt.Errorf("deleting VPC peering: %w", err)
		}
		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		result := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{peeringID, status}
		PrintOutput(result, headers, [][]string{{peeringID, status}})
		return nil
	},
}
