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
	vpnrouteCreateCmd.MarkFlagRequired("name")
	vpnrouteCreateCmd.MarkFlagRequired("region")
	vpnrouteCreateCmd.MarkFlagRequired("cloud-subnet")
	vpnrouteCreateCmd.MarkFlagRequired("onprem-subnet")

	vpnrouteGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpnrouteUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpnrouteUpdateCmd.Flags().String("name", "", "New name for the VPN route")
	vpnrouteUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpnrouteUpdateCmd.Flags().String("cloud-subnet", "", "CIDR of the cloud subnet")
	vpnrouteUpdateCmd.Flags().String("onprem-subnet", "", "CIDR of the on-prem subnet")

	vpnrouteDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpnrouteDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpnrouteDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	vpnrouteListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpnrouteListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpnrouteListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a route that directs traffic through the specified VPN tunnel.

Specify the cloud-side subnet with --cloud-subnet and the on-premises subnet
with --onprem-subnet. Both values should be valid CIDR blocks.`,
	Example: `  acloud network vpnroute create <vpn-tunnel-id> \
    --name my-route --region IT-BG \
    --cloud-subnet 10.0.0.0/24 \
    --onprem-subnet 192.168.1.0/24`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		cloudSubnet, _ := cmd.Flags().GetString("cloud-subnet")
		onPremSubnet, _ := cmd.Flags().GetString("onprem-subnet")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		verbose, _ := cmd.Flags().GetBool("verbose")

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
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

		ctx, cancel := newCtx()
		defer cancel()
		resp, err := client.FromNetwork().VPNRoutes().Create(ctx, projectID, vpnTunnelID, req, nil)
		if err != nil {
			return fmt.Errorf("creating VPN route: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
			PrintOutput(resp.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("VPN route", name))
		}
		return nil
	},
}

var vpnrouteGetCmd = &cobra.Command{
	Use:   "get [vpn-tunnel-id] [route-id]",
	Short: "Get VPN tunnel route details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]
		routeID := args[1]

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
		resp, err := client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil {
			return fmt.Errorf("getting VPN route: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
				if route.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", route.Metadata.LocationResponse.Value)
				}
			}
			fmt.Printf("Cloud Subnet:    %s\n", route.Properties.CloudSubnet)
			fmt.Printf("OnPrem Subnet:   %s\n", route.Properties.OnPremSubnet)
			if route.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", route.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var vpnrouteListCmd = &cobra.Command{
	Use:   "list [vpn-tunnel-id]",
	Short: "List VPN tunnel routes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]

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
		resp, err := client.FromNetwork().VPNRoutes().List(ctx, projectID, vpnTunnelID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing VPN routes: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
			PrintOutput(resp.Data, headers, rows)
		} else {
			fmt.Println("No VPN routes found.")
		}
		return nil
	},
}

var vpnrouteUpdateCmd = &cobra.Command{
	Use:   "update [vpn-tunnel-id] [route-id]",
	Short: "Update a VPN tunnel route",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]
		routeID := args[1]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		cloudSubnet, _ := cmd.Flags().GetString("cloud-subnet")
		onPremSubnet, _ := cmd.Flags().GetString("onprem-subnet")

		// At least one field must be provided
		if name == "" && !cmd.Flags().Changed("tags") && cloudSubnet == "" && onPremSubnet == "" {
			return fmt.Errorf("at least one field must be provided for update")
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

		// Fetch current VPN route details
		getResp, err := client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			return fmt.Errorf("fetching current VPN route: %w", err)
		}

		current := getResp.Data

		// Block update if VPN route is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPN route while it is in 'InCreation' state. Please wait until the VPN route is fully created")
		}

		// Normalize region code if needed
		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for VPN route")
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
			return fmt.Errorf("updating VPN route: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
			PrintOutput(resp.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgUpdatedAsync("VPN route", routeID))
		}
		return nil
	},
}

var vpnrouteDeleteCmd = &cobra.Command{
	Use:   "delete [vpn-tunnel-id] [route-id]",
	Short: "Delete a VPN tunnel route",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]
		routeID := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			ok, err := confirmDelete("VPN route", routeID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromNetwork().VPNRoutes().Get(ctx, projectID, vpnTunnelID, routeID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: VPN route not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("VPN route", routeID))
			return nil
		}

		resp, err := client.FromNetwork().VPNRoutes().Delete(ctx, projectID, vpnTunnelID, routeID, nil)
		if err != nil {
			return fmt.Errorf("deleting VPN route: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		result := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{routeID, status}
		PrintOutput(result, headers, [][]string{{routeID, status}})
		return nil
	},
}
