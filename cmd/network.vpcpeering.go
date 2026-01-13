package cmd

import (
	"context"
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

	vpcpeeringGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpcpeeringUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringUpdateCmd.Flags().String("name", "", "New name for the VPC peering")
	vpcpeeringUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	vpcpeeringDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	vpcpeeringListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		name, _ := cmd.Flags().GetString("name")
		peerVPCID, _ := cmd.Flags().GetString("peer-vpc-id")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if name == "" || peerVPCID == "" || region == "" {
			fmt.Println("Error: --name, --peer-vpc-id, and --region are required")
			return
		}
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}
		ctx := context.Background()
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
			fmt.Printf("Error creating VPC peering: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to create VPC peering - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
			fmt.Println("VPC peering created, but no ID returned.")
		}
	},
}

var vpcpeeringGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [peering-id]",
	Short: "Get VPC peering details",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}
		ctx := context.Background()
		resp, err := client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
		if err != nil {
			fmt.Printf("Error getting VPC peering: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get VPC peering - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
				fmt.Printf("Region:          %s\n", peering.Metadata.LocationResponse.Value)
			}
			if peering.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", peering.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}

var vpcpeeringListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List VPC peerings for a VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}
		ctx := context.Background()
		resp, err := client.FromNetwork().VPCPeerings().List(ctx, projectID, vpcID, nil)
		if err != nil {
			fmt.Printf("Error listing VPC peerings: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list VPC peerings - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
	},
}

var vpcpeeringUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [peering-id]",
	Short: "Update a VPC peering",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}
		ctx := context.Background()
		// Fetch current peering details
		getResp, err := client.FromNetwork().VPCPeerings().Get(ctx, projectID, vpcID, peeringID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			fmt.Printf("Error fetching current VPC peering: %v\n", err)
			return
		}
		current := getResp.Data
		// Block update if peering is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update VPC peering while it is in 'InCreation' state. Please wait until the VPC peering is fully created.")
			return
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
			fmt.Printf("Error updating VPC peering: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to update VPC peering - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
			fmt.Printf("VPC peering '%s' updated.\n", peeringID)
		}
	},
}

var vpcpeeringDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [peering-id]",
	Short: "Delete a VPC peering",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete VPC peering %s? This action cannot be undone.\n", peeringID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}
		ctx := context.Background()
		resp, err := client.FromNetwork().VPCPeerings().Delete(ctx, projectID, vpcID, peeringID, nil)
		if err != nil {
			fmt.Printf("Error deleting VPC peering: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to delete VPC peering - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}
		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		PrintTable(headers, [][]string{{peeringID, status}})
	},
}
