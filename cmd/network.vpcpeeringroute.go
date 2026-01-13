package cmd

import (
	"context"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	networkCmd.AddCommand(vpcpeeringrouteCmd)

	vpcpeeringrouteCmd.AddCommand(vpcpeeringrouteCreateCmd)
	vpcpeeringrouteCmd.AddCommand(vpcpeeringrouteGetCmd)
	vpcpeeringrouteCmd.AddCommand(vpcpeeringrouteUpdateCmd)
	vpcpeeringrouteCmd.AddCommand(vpcpeeringrouteDeleteCmd)
	vpcpeeringrouteCmd.AddCommand(vpcpeeringrouteListCmd)

	// VPC Peering Route flags
	vpcpeeringrouteCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringrouteCreateCmd.Flags().String("name", "", "VPC Peering Route name (required)")
	vpcpeeringrouteCreateCmd.Flags().String("local-network", "", "Local network address in CIDR notation (required)")
	vpcpeeringrouteCreateCmd.Flags().String("remote-network", "", "Remote network address in CIDR notation (required)")
	vpcpeeringrouteCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	vpcpeeringrouteCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcpeeringrouteCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")

	vpcpeeringrouteGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpcpeeringrouteUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringrouteUpdateCmd.Flags().String("name", "", "New name for the VPC peering route")
	vpcpeeringrouteUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpcpeeringrouteUpdateCmd.Flags().String("local-network", "", "Local network address in CIDR notation")
	vpcpeeringrouteUpdateCmd.Flags().String("remote-network", "", "Remote network address in CIDR notation")
	vpcpeeringrouteUpdateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")

	vpcpeeringrouteDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringrouteDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	vpcpeeringrouteListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	vpcpeeringrouteGetCmd.ValidArgsFunction = completeVPCPeeringRouteID
	vpcpeeringrouteUpdateCmd.ValidArgsFunction = completeVPCPeeringRouteID
	vpcpeeringrouteDeleteCmd.ValidArgsFunction = completeVPCPeeringRouteID
}

// Completion functions for network resources
func completeVPCPeeringRouteID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	vpcID := args[0]
	vpcPeeringID := args[1]

	ctx := context.Background()
	response, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, projectID, vpcID, vpcPeeringID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Note: VPCPeeringRouteResponse.Metadata doesn't have ID field
	// We can't provide completion without ID, so return empty
	// The user will need to type the route ID manually
	var completions []string
	if response != nil && response.Data != nil {
		// Completion not available for VPC peering routes due to missing ID in metadata
		_ = response.Data.Values
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var vpcpeeringrouteCmd = &cobra.Command{
	Use:   "vpcpeeringroute",
	Short: "Manage VPC peering routes",
	Long:  `Perform CRUD operations on VPC peering routes in Aruba Cloud.`,
}

var vpcpeeringrouteCreateCmd = &cobra.Command{
	Use:   "create [vpc-id] [peering-id]",
	Short: "Create a new VPC peering route",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]

		name, _ := cmd.Flags().GetString("name")
		localNetwork, _ := cmd.Flags().GetString("local-network")
		remoteNetwork, _ := cmd.Flags().GetString("remote-network")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Validate required fields
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}
		if localNetwork == "" {
			fmt.Println("Error: --local-network is required")
			return
		}
		if remoteNetwork == "" {
			fmt.Println("Error: --remote-network is required")
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

		// Build the create request
		req := types.VPCPeeringRouteRequest{
			Metadata: types.ResourceMetadataRequest{
				Name: name,
				Tags: tags,
			},
			Properties: types.VPCPeeringRoutePropertiesRequest{
				LocalNetworkAddress:  localNetwork,
				RemoteNetworkAddress: remoteNetwork,
				BillingPlan: types.BillingPeriodResource{
					BillingPeriod: billingPeriod,
				},
			},
		}

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating VPC peering route with the following parameters:")
			fmt.Printf("  Name:            %s\n", name)
			fmt.Printf("  Local Network:   %s\n", localNetwork)
			fmt.Printf("  Remote Network:  %s\n", remoteNetwork)
			fmt.Printf("  Billing Period:  %s\n", billingPeriod)
			if len(tags) > 0 {
				fmt.Printf("  Tags:            %v\n", tags)
			}
			fmt.Println()
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().VPCPeeringRoutes().Create(ctx, projectID, vpcID, peeringID, req, nil)
		if err != nil {
			fmt.Printf("Error creating VPC peering route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to create VPC peering route - Status: %d\n", resp.StatusCode)
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
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				resp.Data.Metadata.Name,
				localNetwork,
				remoteNetwork,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("VPC peering route created, but no data returned.")
		}
	},
}

var vpcpeeringrouteGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [peering-id] [route-id]",
	Short: "Get VPC peering route details",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]
		routeID := args[2]

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
		resp, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil {
			fmt.Printf("Error getting VPC peering route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get VPC peering route - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			route := resp.Data
			fmt.Println("\nVPC Peering Route Details:")
			fmt.Println("==========================")
			fmt.Printf("ID:              %s\n", routeID)
			fmt.Printf("Name:            %s\n", route.Metadata.Name)
			fmt.Printf("Local Network:    %s\n", route.Properties.LocalNetworkAddress)
			fmt.Printf("Remote Network:   %s\n", route.Properties.RemoteNetworkAddress)
			if route.Properties.BillingPlan.BillingPeriod != "" {
				fmt.Printf("Billing Period:  %s\n", route.Properties.BillingPlan.BillingPeriod)
			}
			if len(route.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", route.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			if route.Status.State != nil {
				fmt.Printf("Status:          %s\n", *route.Status.State)
			}
		} else {
			fmt.Println("VPC peering route not found or no data returned.")
		}
	},
}

var vpcpeeringrouteListCmd = &cobra.Command{
	Use:   "list [vpc-id] [peering-id]",
	Short: "List VPC peering routes",
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
		resp, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, projectID, vpcID, peeringID, nil)
		if err != nil {
			fmt.Printf("Error listing VPC peering routes: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list VPC peering routes - Status: %d\n", resp.StatusCode)
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
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, route := range resp.Data.Values {
				name := route.Metadata.Name
				localNetwork := route.Properties.LocalNetworkAddress
				remoteNetwork := route.Properties.RemoteNetworkAddress
				status := ""
				if route.Status.State != nil {
					status = *route.Status.State
				}
				rows = append(rows, []string{name, localNetwork, remoteNetwork, status})
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No VPC peering routes found.")
		}
	},
}

var vpcpeeringrouteUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [peering-id] [route-id]",
	Short: "Update a VPC peering route",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]
		routeID := args[2]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		localNetwork, _ := cmd.Flags().GetString("local-network")
		remoteNetwork, _ := cmd.Flags().GetString("remote-network")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		// At least one field must be provided
		if name == "" && !cmd.Flags().Changed("tags") && localNetwork == "" && remoteNetwork == "" && billingPeriod == "" {
			fmt.Println("Error: at least one field must be provided for update")
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

		// Fetch current VPC peering route details
		getResp, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			fmt.Printf("Error fetching current VPC peering route: %v\n", err)
			return
		}

		current := getResp.Data

		// Block update if VPC peering route is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update VPC peering route while it is in 'InCreation' state. Please wait until the VPC peering route is fully created.")
			return
		}

		// Build update request by merging user input with current values
		req := types.VPCPeeringRouteRequest{
			Metadata: types.ResourceMetadataRequest{
				Name: func() string {
					if name != "" {
						return name
					}
					return current.Metadata.Name
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
			Properties: types.VPCPeeringRoutePropertiesRequest{
				LocalNetworkAddress: func() string {
					if localNetwork != "" {
						return localNetwork
					}
					return current.Properties.LocalNetworkAddress
				}(),
				RemoteNetworkAddress: func() string {
					if remoteNetwork != "" {
						return remoteNetwork
					}
					return current.Properties.RemoteNetworkAddress
				}(),
				BillingPlan: types.BillingPeriodResource{
					BillingPeriod: func() string {
						if billingPeriod != "" {
							return billingPeriod
						}
						return current.Properties.BillingPlan.BillingPeriod
					}(),
				},
			},
		}

		resp, err := client.FromNetwork().VPCPeeringRoutes().Update(ctx, projectID, vpcID, peeringID, routeID, req, nil)
		if err != nil {
			fmt.Printf("Error updating VPC peering route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to update VPC peering route - Status: %d\n", resp.StatusCode)
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
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				resp.Data.Metadata.Name,
				resp.Data.Properties.LocalNetworkAddress,
				resp.Data.Properties.RemoteNetworkAddress,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Printf("VPC peering route '%s' updated.\n", routeID)
		}
	},
}

var vpcpeeringrouteDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [peering-id] [route-id]",
	Short: "Delete a VPC peering route",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		peeringID := args[1]
		routeID := args[2]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete VPC peering route %s? This action cannot be undone.\n", routeID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().VPCPeeringRoutes().Delete(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil {
			fmt.Printf("Error deleting VPC peering route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to delete VPC peering route - Status: %d\n", resp.StatusCode)
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
		PrintTable(headers, [][]string{{routeID, status}})
	},
}
