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
	vpcpeeringrouteCreateCmd.Flags().String("region", "", "Region code (required)")
	vpcpeeringrouteCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	vpcpeeringrouteCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcpeeringrouteCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	vpcpeeringrouteCreateCmd.MarkFlagRequired("name")
	vpcpeeringrouteCreateCmd.MarkFlagRequired("local-network")
	vpcpeeringrouteCreateCmd.MarkFlagRequired("remote-network")
	vpcpeeringrouteCreateCmd.MarkFlagRequired("region")

	vpcpeeringrouteGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpcpeeringrouteUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringrouteUpdateCmd.Flags().String("name", "", "New name for the VPC peering route")
	vpcpeeringrouteUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpcpeeringrouteUpdateCmd.Flags().String("local-network", "", "Local network address in CIDR notation")
	vpcpeeringrouteUpdateCmd.Flags().String("remote-network", "", "Remote network address in CIDR notation")
	vpcpeeringrouteUpdateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")

	vpcpeeringrouteDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringrouteDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcpeeringrouteDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	vpcpeeringrouteListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringrouteListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpcpeeringrouteListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a route that directs traffic through a VPC peering connection.

Specify the local subnet CIDR with --local-network and the remote subnet CIDR
with --remote-network. Both values should be valid CIDR blocks.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud network vpcpeeringroute create <vpc-id> <peering-id> \
    --name my-route \
    --region ITBG-Bergamo \
    --local-network 10.0.0.0/24 \
    --remote-network 10.1.0.0/24`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		peeringID := args[1]

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		localNetwork, _ := cmd.Flags().GetString("local-network")
		remoteNetwork, _ := cmd.Flags().GetString("remote-network")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
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
		req := types.VPCPeeringRouteRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{Value: region},
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

		ctx, cancel := newCtx()
		defer cancel()
		resp, err := client.FromNetwork().VPCPeeringRoutes().Create(ctx, projectID, vpcID, peeringID, req, nil)
		if err != nil {
			return fmt.Errorf("creating VPC peering route: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				func() string {
					if resp.Data.Metadata.Name != nil {
						return *resp.Data.Metadata.Name
					}
					return name
				}(),
				func() string {
					if resp.Data.Metadata.ID != nil {
						return *resp.Data.Metadata.ID
					}
					return ""
				}(),
				localNetwork,
				remoteNetwork,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintOutput(resp.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("VPC peering route", name))
		}
		return nil
	},
}

var vpcpeeringrouteGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [peering-id] [route-id]",
	Short: "Get VPC peering route details",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		peeringID := args[1]
		routeID := args[2]

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
		resp, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil {
			return fmt.Errorf("getting VPC peering route: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil {
			route := resp.Data
			fmt.Println("\nVPC Peering Route Details:")
			fmt.Println("==========================")
			fmt.Printf("ID:              %s\n", routeID)
			if route.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *route.Metadata.Name)
			}
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
		return nil
	},
}

var vpcpeeringrouteListCmd = &cobra.Command{
	Use:   "list [vpc-id] [peering-id]",
	Short: "List VPC peering routes",
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
		resp, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, projectID, vpcID, peeringID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing VPC peering routes: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
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
				localNetwork := route.Properties.LocalNetworkAddress
				remoteNetwork := route.Properties.RemoteNetworkAddress
				status := ""
				if route.Status.State != nil {
					status = *route.Status.State
				}
				rows = append(rows, []string{name, id, localNetwork, remoteNetwork, status})
			}
			PrintOutput(resp.Data, headers, rows)
		} else {
			fmt.Println("No VPC peering routes found.")
		}
		return nil
	},
}

var vpcpeeringrouteUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [peering-id] [route-id]",
	Short: "Update a VPC peering route",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Fetch current VPC peering route details
		getResp, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			return fmt.Errorf("fetching current VPC peering route: %w", err)
		}

		current := getResp.Data

		// Block update if VPC peering route is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC peering route while it is in 'InCreation' state. Please wait until the VPC peering route is fully created")
		}

		// Preserve the region from the current resource for the update request.
		var regionValue string
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}

		// Build update request by merging user input with current values
		req := types.VPCPeeringRouteRequest{
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
				Location: types.LocationRequest{Value: regionValue},
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
			return fmt.Errorf("updating VPC peering route: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
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
					return routeID
				}(),
				resp.Data.Properties.LocalNetworkAddress,
				resp.Data.Properties.RemoteNetworkAddress,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintOutput(resp.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgUpdatedAsync("VPC peering route", routeID))
		}
		return nil
	},
}

var vpcpeeringrouteDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [peering-id] [route-id]",
	Short: "Delete a VPC peering route",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		peeringID := args[1]
		routeID := args[2]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			ok, err := confirmDelete("VPC peering route", routeID)
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
			_, err = client.FromNetwork().VPCPeeringRoutes().Get(ctx, projectID, vpcID, peeringID, routeID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: VPC peering route not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("VPC peering route", routeID))
			return nil
		}

		resp, err := client.FromNetwork().VPCPeeringRoutes().Delete(ctx, projectID, vpcID, peeringID, routeID, nil)
		if err != nil {
			return fmt.Errorf("deleting VPC peering route: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
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
