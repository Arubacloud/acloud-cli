package cmd

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

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
		req := types.VPCPeeringRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{Value: region},
			},
			Properties: types.VPCPeeringPropertiesRequest{
				RemoteVPC: &types.ReferenceResource{URI: peerVPCID},
			},
		}
		resp, err := client.FromNetwork().VPCPeerings().Create(ctx, projectID, vpcID, req, nil)
		if err != nil {
			return fmt.Errorf("creating VPC peering: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil && resp.Data.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				name,
				*resp.Data.Metadata.ID,
				func() string {
					if resp.Data.Properties.RemoteVPC != nil {
						return resp.Data.Properties.RemoteVPC.URI
					}
					return ""
				}(),
				func() string {
					if resp.Data.Metadata.LocationResponse != nil {
						return resp.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					} else {
						return ""
					}
				}(),
			}
			PrintTable(headers, [][]string{row})
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
		resp, err := client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
		if err != nil {
			return fmt.Errorf("getting VPC peering: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil {
			peering := resp.Data
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
				if peering.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", peering.Metadata.LocationResponse.Value)
				}
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
		resp, err := client.FromNetwork().VPCPeerings().List(ctx, projectID, vpcID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing VPC peerings: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, peering := range resp.Data.Values {
				name := ""
				if peering.Metadata.Name != nil {
					name = *peering.Metadata.Name
				}
				id := ""
				if peering.Metadata.ID != nil {
					id = *peering.Metadata.ID
				}
				peerVPC := ""
				if peering.Properties.RemoteVPC != nil {
					peerVPC = peering.Properties.RemoteVPC.URI
				}
				region := ""
				if peering.Metadata.LocationResponse != nil {
					region = peering.Metadata.LocationResponse.Value
				}
				status := ""
				if peering.Status.State != nil {
					status = *peering.Status.State
				}
				rows = append(rows, []string{name, id, peerVPC, region, status})
			}
			PrintTable(headers, rows)
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
		// Fetch current peering details
		getResp, err := client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			return fmt.Errorf("fetching current VPC peering: %w", err)
		}
		current := getResp.Data
		// Block update if peering is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC peering while it is in 'InCreation' state. Please wait until the VPC peering is fully created")
		}
		// Build update request by merging user input with all current valid fields
		req := types.VPCPeeringRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: func() string {
						if name != "" {
							return name
						}
						if current.Metadata.Name != nil {
							return *current.Metadata.Name
						}
						return ""
					}(),
					Tags: func() []string {
						if cmd.Flags().Changed("tags") {
							return tags
						}
						if current.Metadata.Tags != nil {
							return current.Metadata.Tags
						}
						return []string{}
					}(),
				},
				Location: types.LocationRequest{
					Value: func() string {
						if current.Metadata.LocationResponse != nil {
							return current.Metadata.LocationResponse.Value
						}
						return ""
					}(),
				},
			},
			Properties: types.VPCPeeringPropertiesRequest{
				RemoteVPC: current.Properties.RemoteVPC,
			},
		}
		resp, err := client.FromNetwork().VPCPeerings().Update(ctx, projectID, vpcID, peeringID, req, nil)
		if err != nil {
			return fmt.Errorf("updating VPC peering: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "PEER VPC", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				func() string {
					if resp.Data.Metadata.Name != nil {
						return *resp.Data.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if resp.Data.Metadata.ID != nil {
						return *resp.Data.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if resp.Data.Properties.RemoteVPC != nil {
						return resp.Data.Properties.RemoteVPC.URI
					}
					return ""
				}(),
				func() string {
					if resp.Data.Metadata.LocationResponse != nil {
						return resp.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
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

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
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
			_, err = client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: VPC peering not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("VPC peering", peeringID))
			return nil
		}

		resp, err := client.FromNetwork().VPCPeerings().Delete(ctx, projectID, vpcID, peeringID, nil)
		if err != nil {
			return fmt.Errorf("deleting VPC peering: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		PrintTable(headers, [][]string{{peeringID, status}})
		return nil
	},
}
