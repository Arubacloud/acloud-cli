package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func vpnRouteListPayload(l *aruba.List[*aruba.VPNRoute]) any {
	if r, ok := l.Raw().(*types.Response[types.VPNRouteList]); ok && r != nil {
		return r.Data
	}
	return nil
}

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

	vpnrouteGetCmd.ValidArgsFunction = completeVPNRouteID
	vpnrouteUpdateCmd.ValidArgsFunction = completeVPNRouteID
	vpnrouteDeleteCmd.ValidArgsFunction = completeVPNRouteID
}

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
	list, err := client.FromNetwork().VPNRoutes().List(ctx, aruba.VPNTunnelRef(projectID, vpnTunnelID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	if list != nil {
		for _, route := range list.Items() {
			raw := route.Raw()
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

		route := aruba.NewVPNRoute().
			IntoVPNTunnel(aruba.VPNTunnelRef(projectID, vpnTunnelID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithCloudSubnet(cloudSubnet).
			WithOnPremSubnet(onPremSubnet).
			ReplaceTags(tags...)

		ctx, cancel := newCtx()
		defer cancel()
		resp, err := client.FromNetwork().VPNRoutes().Create(ctx, route)
		if err != nil {
			return fmt.Errorf("creating VPN route: %w", apiErrFromV2(err))
		}

		if resp != nil && resp.Raw() != nil {
			raw := resp.Raw()
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "CLOUD SUBNET", Width: 18},
				{Header: "ONPREM SUBNET", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			cloudSubnetVal := raw.Properties.CloudSubnet.CIDR
			onPremSubnetVal := raw.Properties.OnPremSubnet
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			row := []string{name, id, cloudSubnetVal, onPremSubnetVal, status}
			PrintOutput(raw, headers, [][]string{row})
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
		route, err := client.FromNetwork().VPNRoutes().Get(ctx, aruba.VPNRouteRef(projectID, vpnTunnelID, routeID))
		if err != nil {
			return fmt.Errorf("getting VPN route: %w", apiErrFromV2(err))
		}

		if route != nil && route.Raw() != nil {
			raw := route.Raw()
			fmt.Println("\nVPN Route Details:")
			fmt.Println("==================")
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", raw.Metadata.LocationResponse.Value)
			}
			fmt.Printf("Cloud Subnet:    %s\n", raw.Properties.CloudSubnet.CIDR)
			fmt.Printf("OnPrem Subnet:   %s\n", raw.Properties.OnPremSubnet)
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
		list, err := client.FromNetwork().VPNRoutes().List(ctx, aruba.VPNTunnelRef(projectID, vpnTunnelID))
		if err != nil {
			return fmt.Errorf("listing VPN routes: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "CLOUD SUBNET", Width: 18},
				{Header: "ONPREM SUBNET", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, route := range list.Items() {
				raw := route.Raw()
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
				cloudSubnet := raw.Properties.CloudSubnet.CIDR
				onPremSubnet := raw.Properties.OnPremSubnet
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, cloudSubnet, onPremSubnet, status})
			}
			PrintOutput(list.Raw(), headers, rows)
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
		route, err := client.FromNetwork().VPNRoutes().Get(ctx, aruba.VPNRouteRef(projectID, vpnTunnelID, routeID))
		if err != nil || route == nil || route.Raw() == nil {
			return fmt.Errorf("fetching current VPN route: %w", apiErrFromV2(err))
		}

		// Block update if VPN route is in 'InCreation' state
		if route.Raw().Status.State != nil && *route.Raw().Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPN route while it is in 'InCreation' state. Please wait until the VPN route is fully created")
		}

		if name != "" {
			route.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			route.ReplaceTags(tags...)
		}
		if cloudSubnet != "" {
			route.WithCloudSubnet(cloudSubnet)
		}
		if onPremSubnet != "" {
			route.WithOnPremSubnet(onPremSubnet)
		}

		updated, err := client.FromNetwork().VPNRoutes().Update(ctx, route)
		if err != nil {
			return fmt.Errorf("updating VPN route: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "CLOUD SUBNET", Width: 18},
				{Header: "ONPREM SUBNET", Width: 18},
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
			cloudSubnetVal := raw.Properties.CloudSubnet.CIDR
			onPremSubnetVal := raw.Properties.OnPremSubnet
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			row := []string{nameVal, id, cloudSubnetVal, onPremSubnetVal, status}
			PrintOutput(raw, headers, [][]string{row})
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")
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
			_, err = client.FromNetwork().VPNRoutes().Get(ctx, aruba.VPNRouteRef(projectID, vpnTunnelID, routeID))
			if err != nil {
				return fmt.Errorf("dry-run: VPN route not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPN route", routeID))
			return nil
		}

		err = client.FromNetwork().VPNRoutes().Delete(ctx, aruba.VPNRouteRef(projectID, vpnTunnelID, routeID))
		if err != nil {
			return fmt.Errorf("deleting VPN route: %w", apiErrFromV2(err))
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
