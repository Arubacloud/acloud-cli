package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	networkCmd.AddCommand(vpnrouteCmd)

	vpnrouteCmd.AddCommand(vpnrouteCreateCmd)
	vpnrouteCmd.AddCommand(vpnrouteGetCmd)
	vpnrouteCmd.AddCommand(vpnrouteUpdateCmd)
	vpnrouteCmd.AddCommand(vpnrouteDeleteCmd)
	vpnrouteCmd.AddCommand(vpnrouteListCmd)

	// VPN Route flags
	vpnrouteCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpnrouteCreateCmd.Flags().String("name", "", "VPN Route name (required)")
	vpnrouteCreateCmd.Flags().String("region", "", "Region code (required)")
	vpnrouteCreateCmd.Flags().String("cloud-subnet", "", "CIDR of the cloud subnet (required)")
	vpnrouteCreateCmd.Flags().String("onprem-subnet", "", "CIDR of the on-prem subnet (required)")
	vpnrouteCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpnrouteCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")

	vpnrouteGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpnrouteUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpnrouteUpdateCmd.Flags().String("name", "", "New name for the VPN route")
	vpnrouteUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpnrouteUpdateCmd.Flags().String("cloud-subnet", "", "CIDR of the cloud subnet")
	vpnrouteUpdateCmd.Flags().String("onprem-subnet", "", "CIDR of the on-prem subnet")

	vpnrouteDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpnrouteDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	vpnrouteListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	vpnrouteGetCmd.ValidArgsFunction = completeVPNRouteID
	vpnrouteUpdateCmd.ValidArgsFunction = completeVPNRouteID
	vpnrouteDeleteCmd.ValidArgsFunction = completeVPNRouteID
}

// Completion functions for network resources
func completeVPNRouteID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 1 {
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

	vpnTunnelID := args[0]

	ctx := context.Background()
	response, err := client.FromNetwork().VPNRoutes().List(ctx, projectID, vpnTunnelID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, route := range response.Data.Values {
			if route.Metadata.ID != nil && route.Metadata.Name != nil {
				id := *route.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *route.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var vpnrouteCmd = &cobra.Command{
	Use:   "vpnroute",
	Short: "Manage VPN tunnel routes",
	Long:  `Perform CRUD operations on VPN tunnel routes in Aruba Cloud.`,
}

var vpnrouteCreateCmd = &cobra.Command{
	Use:   "create [vpn-tunnel-id]",
	Short: "Create a new VPN tunnel route",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		cloudSubnet, _ := cmd.Flags().GetString("cloud-subnet")
		onPremSubnet, _ := cmd.Flags().GetString("onprem-subnet")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Validate required fields
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}
		if region == "" {
			fmt.Println("Error: --region is required")
			return
		}
		if cloudSubnet == "" {
			fmt.Println("Error: --cloud-subnet is required")
			return
		}
		if onPremSubnet == "" {
			fmt.Println("Error: --onprem-subnet is required")
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
		req := types.VPNRouteRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.VPNRoutePropertiesRequest{
				CloudSubnet:  cloudSubnet,
				OnPremSubnet: onPremSubnet,
			},
		}

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating VPN route with the following parameters:")
			fmt.Printf("  Name:          %s\n", name)
			fmt.Printf("  Region:        %s\n", region)
			fmt.Printf("  Cloud Subnet:  %s\n", cloudSubnet)
			fmt.Printf("  OnPrem Subnet: %s\n", onPremSubnet)
			if len(tags) > 0 {
				fmt.Printf("  Tags:          %v\n", tags)
			}
			fmt.Println()
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().VPNRoutes().Create(ctx, projectID, vpnTunnelID, req, nil)
		if err != nil {
			fmt.Printf("Error creating VPN route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to create VPN route - Status: %d\n", resp.StatusCode)
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
				{Header: "CLOUD SUBNET", Width: 18},
				{Header: "ONPREM SUBNET", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				name,
				*resp.Data.Metadata.ID,
				cloudSubnet,
				onPremSubnet,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("VPN route created, but no ID returned.")
		}
	},
}

var vpnrouteGetCmd = &cobra.Command{
	Use:   "get [vpn-tunnel-id] [route-id]",
	Short: "Get VPN tunnel route details",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]
		routeID := args[1]

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
		resp, err := client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil {
			fmt.Printf("Error getting VPN route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get VPN route - Status: %d\n", resp.StatusCode)
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
			fmt.Println("\nVPN Route Details:")
			fmt.Println("==================")
			if route.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *route.Metadata.ID)
			}
			if route.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *route.Metadata.URI)
			}
			if route.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *route.Metadata.Name)
			}
			if route.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", route.Metadata.LocationResponse.Value)
			}
			fmt.Printf("Cloud Subnet:    %s\n", route.Properties.CloudSubnet)
			fmt.Printf("OnPrem Subnet:   %s\n", route.Properties.OnPremSubnet)
			if route.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", route.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if route.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *route.Metadata.CreatedBy)
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
			fmt.Println("VPN route not found or no data returned.")
		}
	},
}

var vpnrouteListCmd = &cobra.Command{
	Use:   "list [vpn-tunnel-id]",
	Short: "List VPN tunnel routes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]

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
		resp, err := client.FromNetwork().VPNRoutes().List(ctx, projectID, vpnTunnelID, nil)
		if err != nil {
			fmt.Printf("Error listing VPN routes: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list VPN routes - Status: %d\n", resp.StatusCode)
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
				{Header: "CLOUD SUBNET", Width: 18},
				{Header: "ONPREM SUBNET", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, route := range resp.Data.Values {
				name := ""
				if route.Metadata.Name != nil {
					name = *route.Metadata.Name
				}
				id := ""
				if route.Metadata.ID != nil {
					id = *route.Metadata.ID
				}
				cloudSubnet := route.Properties.CloudSubnet
				onPremSubnet := route.Properties.OnPremSubnet
				status := ""
				if route.Status.State != nil {
					status = *route.Status.State
				}
				rows = append(rows, []string{name, id, cloudSubnet, onPremSubnet, status})
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No VPN routes found.")
		}
	},
}

var vpnrouteUpdateCmd = &cobra.Command{
	Use:   "update [vpn-tunnel-id] [route-id]",
	Short: "Update a VPN tunnel route",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]
		routeID := args[1]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		cloudSubnet, _ := cmd.Flags().GetString("cloud-subnet")
		onPremSubnet, _ := cmd.Flags().GetString("onprem-subnet")

		// At least one field must be provided
		if name == "" && !cmd.Flags().Changed("tags") && cloudSubnet == "" && onPremSubnet == "" {
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

		// Fetch current VPN route details
		getResp, err := client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			fmt.Printf("Error fetching current VPN route: %v\n", err)
			return
		}

		current := getResp.Data

		// Block update if VPN route is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update VPN route while it is in 'InCreation' state. Please wait until the VPN route is fully created.")
			return
		}

		// Normalize region code if needed
		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for VPN route")
			return
		}

		// Build update request by merging user input with current values
		req := types.VPNRouteRequest{
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
					Value: regionValue,
				},
			},
			Properties: types.VPNRoutePropertiesRequest{
				CloudSubnet: func() string {
					if cloudSubnet != "" {
						return cloudSubnet
					}
					return current.Properties.CloudSubnet
				}(),
				OnPremSubnet: func() string {
					if onPremSubnet != "" {
						return onPremSubnet
					}
					return current.Properties.OnPremSubnet
				}(),
			},
		}

		resp, err := client.FromNetwork().VPNRoutes().Update(ctx, projectID, vpnTunnelID, routeID, req, nil)
		if err != nil {
			fmt.Printf("Error updating VPN route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to update VPN route - Status: %d\n", resp.StatusCode)
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
				{Header: "CLOUD SUBNET", Width: 18},
				{Header: "ONPREM SUBNET", Width: 18},
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
				resp.Data.Properties.CloudSubnet,
				resp.Data.Properties.OnPremSubnet,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Printf("VPN route '%s' updated.\n", routeID)
		}
	},
}

var vpnrouteDeleteCmd = &cobra.Command{
	Use:   "delete [vpn-tunnel-id] [route-id]",
	Short: "Delete a VPN tunnel route",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]
		routeID := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete VPN route %s? This action cannot be undone.\n", routeID)
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
		resp, err := client.FromNetwork().VPNRoutes().Delete(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil {
			fmt.Printf("Error deleting VPN route: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to delete VPN route - Status: %d\n", resp.StatusCode)
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
