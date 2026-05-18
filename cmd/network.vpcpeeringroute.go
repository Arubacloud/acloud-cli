package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)


func vpcPeeringRouteListPayload(l *aruba.List[*aruba.VPCPeeringRoute]) any {
	if r, ok := l.Raw().(*types.Response[types.VPCPeeringRouteList]); ok && r != nil {
		return r.Data
	}
	return nil
}

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
	list, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, aruba.VPCPeeringRef(projectID, vpcID, vpcPeeringID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, route := range list.Items() {
			id := route.VPCPeeringRouteID()
			name := route.Name()
			if id == "" {
				continue
			}
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
			}
		}
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

		route := aruba.NewVPCPeeringRoute().
			IntoVPCPeering(aruba.VPCPeeringRef(projectID, vpcID, peeringID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithLocalCIDR(localNetwork).
			WithRemoteCIDR(remoteNetwork).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		if len(tags) > 0 {
			route.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromNetwork().VPCPeeringRoutes().Create(ctx, route)
		if err != nil {
			return fmt.Errorf("creating VPC peering route: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			id := ""
			if r.Metadata.ID != nil {
				id = *r.Metadata.ID
			}
			createdName := name
			if r.Metadata.Name != nil {
				createdName = *r.Metadata.Name
			}
			status := ""
			if r.Status.State != nil {
				status = *r.Status.State
			}
			PrintOutput(r, headers, [][]string{{createdName, id, localNetwork, remoteNetwork, status}})
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
		got, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, aruba.VPCPeeringRouteRef(projectID, vpcID, peeringID, routeID))
		if err != nil {
			return fmt.Errorf("getting VPC peering route: %w", apiErrFromV2(err))
		}

		route := got.Raw()
		if route != nil {
			fmt.Println("\nVPC Peering Route Details:")
			fmt.Println("==========================")
			fmt.Printf("ID:              %s\n", routeID)
			if route.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *route.Metadata.Name)
			}
			fmt.Printf("Local Network:    %s\n", route.Properties.LocalNetworkAddress)
			fmt.Printf("Remote Network:   %s\n", route.Properties.RemoteNetworkAddress)
			if route.Properties.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", string(*route.Properties.BillingPeriod))
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
		list, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, aruba.VPCPeeringRef(projectID, vpcID, peeringID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing VPC peering routes: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, route := range list.Items() {
				r := route.Raw()
				name := route.Name()
				id := route.VPCPeeringRouteID()
				localNetwork := ""
				remoteNetwork := ""
				status := ""
				if r != nil {
					localNetwork = r.Properties.LocalNetworkAddress
					remoteNetwork = r.Properties.RemoteNetworkAddress
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}
				rows = append(rows, []string{name, id, localNetwork, remoteNetwork, status})
			}
			PrintOutput(vpcPeeringRouteListPayload(list), headers, rows)
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

		cur, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, aruba.VPCPeeringRouteRef(projectID, vpcID, peeringID, routeID))
		if err != nil {
			return fmt.Errorf("fetching current VPC peering route: %w", apiErrFromV2(err))
		}
		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("VPC peering route not found")
		}
		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC peering route while it is in 'InCreation' state. Please wait until the VPC peering route is fully created")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}
		if localNetwork != "" {
			cur.WithLocalCIDR(localNetwork)
		}
		if remoteNetwork != "" {
			cur.WithRemoteCIDR(remoteNetwork)
		}
		if billingPeriod != "" {
			cur.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}

		updated, err := client.FromNetwork().VPCPeeringRoutes().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating VPC peering route: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "LOCAL NETWORK", Width: 18},
				{Header: "REMOTE NETWORK", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			updName := ""
			if ur.Metadata.Name != nil {
				updName = *ur.Metadata.Name
			}
			updID := routeID
			if ur.Metadata.ID != nil {
				updID = *ur.Metadata.ID
			}
			updStatus := ""
			if ur.Status.State != nil {
				updStatus = *ur.Status.State
			}
			PrintOutput(ur, headers, [][]string{{updName, updID, ur.Properties.LocalNetworkAddress, ur.Properties.RemoteNetworkAddress, updStatus}})
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("VPC peering route", routeID)
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
			_, err = client.FromNetwork().VPCPeeringRoutes().Get(ctx, aruba.VPCPeeringRouteRef(projectID, vpcID, peeringID, routeID))
			if err != nil {
				return fmt.Errorf("dry-run: VPC peering route not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPC peering route", routeID))
			return nil
		}

		if err := client.FromNetwork().VPCPeeringRoutes().Delete(ctx, aruba.VPCPeeringRouteRef(projectID, vpcID, peeringID, routeID)); err != nil {
			return fmt.Errorf("deleting VPC peering route: %w", apiErrFromV2(err))
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
