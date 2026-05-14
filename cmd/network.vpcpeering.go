package cmd

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// vpcPeeringRef is shared with vpcpeeringroute.go.
func vpcPeeringRef(projectID, vpcID, peeringID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/vpcs/" + vpcID + "/vpcPeerings/" + peeringID)
}

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
			IntoVPC(vpcRef(projectID, vpcID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithRemoteVPC(aruba.URI(peerVPCID))
		if len(tags) > 0 {
			peering.ReplaceTags(tags...)
		}

		created, err := client.FromNetwork().VPCPeerings().Create(ctx, peering)
		if err != nil {
			return fmt.Errorf("creating VPC peering: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil && r.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			peerVPC := ""
			if r.Properties.RemoteVPC != nil {
				peerVPC = r.Properties.RemoteVPC.URI
			}
			regionVal := ""
			if r.Metadata.LocationResponse != nil {
				regionVal = string(r.Metadata.LocationResponse.Value)
			}
			status := ""
			if r.Status.State != nil {
				status = *r.Status.State
			}
			PrintOutput(r, headers, [][]string{{name, *r.Metadata.ID, peerVPC, regionVal, status}})
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
		got, err := client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(projectID, vpcID, peeringID))
		if err != nil {
			return fmt.Errorf("getting VPC peering: %w", apiErrFromV2(err))
		}

		peering := got.Raw()
		if peering != nil {
			fmt.Println("\nVPC Peering Details:")
			fmt.Println("====================")
			if peering.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *peering.Metadata.ID)
			}
			if peering.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *peering.Metadata.Name)
			}
			if peering.Properties.RemoteVPC != nil {
				fmt.Printf("Peer VPC:        %s\n", peering.Properties.RemoteVPC.URI)
			}
			if peering.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", peering.Metadata.LocationResponse.Value)
			}
			if peering.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", peering.Metadata.CreationDate.Format(DateLayout))
			}
			if peering.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *peering.Metadata.CreatedBy)
			}
			if len(peering.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", peering.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			if peering.Status.State != nil {
				fmt.Printf("Status:          %s\n", *peering.Status.State)
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
		list, err := client.FromNetwork().VPCPeerings().List(ctx, vpcRef(projectID, vpcID), listOpts(cmd)...)
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
			for _, p := range list.Items() {
				r := p.Raw()
				name := p.Name()
				id := p.VPCPeeringID()
				peerVPC := ""
				region := ""
				status := ""
				if r != nil {
					if r.Properties.RemoteVPC != nil {
						peerVPC = r.Properties.RemoteVPC.URI
					}
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}
				rows = append(rows, []string{name, id, peerVPC, region, status})
			}
			PrintOutput(vpcPeeringListPayload(list), headers, rows)
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

		cur, err := client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(projectID, vpcID, peeringID))
		if err != nil {
			return fmt.Errorf("fetching current VPC peering: %w", apiErrFromV2(err))
		}
		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("VPC peering not found")
		}
		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC peering while it is in 'InCreation' state. Please wait until the VPC peering is fully created")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().VPCPeerings().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating VPC peering: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			updName := ""
			if ur.Metadata.Name != nil {
				updName = *ur.Metadata.Name
			}
			updID := ""
			if ur.Metadata.ID != nil {
				updID = *ur.Metadata.ID
			}
			peerVPC := ""
			if ur.Properties.RemoteVPC != nil {
				peerVPC = ur.Properties.RemoteVPC.URI
			}
			regionVal := ""
			if ur.Metadata.LocationResponse != nil {
				regionVal = string(ur.Metadata.LocationResponse.Value)
			}
			updStatus := ""
			if ur.Status.State != nil {
				updStatus = *ur.Status.State
			}
			PrintOutput(ur, headers, [][]string{{updName, updID, peerVPC, regionVal, updStatus}})
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
			_, err = client.FromNetwork().VPCPeerings().Get(ctx, vpcPeeringRef(projectID, vpcID, peeringID))
			if err != nil {
				return fmt.Errorf("dry-run: VPC peering not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPC peering", peeringID))
			return nil
		}

		if err := client.FromNetwork().VPCPeerings().Delete(ctx, vpcPeeringRef(projectID, vpcID, peeringID)); err != nil {
			return fmt.Errorf("deleting VPC peering: %w", apiErrFromV2(err))
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
