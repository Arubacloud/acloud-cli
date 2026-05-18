package cmd

import (
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// splitRouteString splits a route string in format "destination:gateway"
func splitRouteString(routeStr string) []string {
	return strings.SplitN(routeStr, ":", 2)
}


func subnetListPayload(l *aruba.List[*aruba.Subnet]) any {
	if r, ok := l.Raw().(*types.Response[types.SubnetList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// INIT
func init() {
	// Subnet
	networkCmd.AddCommand(subnetCmd)
	subnetCmd.AddCommand(subnetCreateCmd)
	subnetCmd.AddCommand(subnetGetCmd)
	subnetCmd.AddCommand(subnetUpdateCmd)
	subnetCmd.AddCommand(subnetDeleteCmd)
	subnetCmd.AddCommand(subnetListCmd)

	subnetCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	subnetCreateCmd.Flags().String("name", "", "Subnet name (required)")
	subnetCreateCmd.Flags().String("cidr", "", "Subnet CIDR (optional, if provided subnet type will be Advanced, otherwise Basic)")
	subnetCreateCmd.Flags().String("region", "", "Region for the subnet (required)")
	subnetCreateCmd.MarkFlagRequired("name")
	subnetCreateCmd.MarkFlagRequired("region")
	subnetCreateCmd.Flags().StringSlice("tags", []string{}, "Subnet tags (optional)")
	subnetCreateCmd.Flags().Bool("dhcp-enabled", false, "Enable DHCP for Advanced subnet type (required when CIDR is provided)")
	subnetCreateCmd.Flags().StringSlice("dhcp-routes", []string{}, "DHCP routes for Advanced subnet type (optional, format: destination:gateway, e.g., '0.0.0.0/0:10.0.0.1')")
	subnetCreateCmd.Flags().StringSlice("dhcp-dns", []string{}, "DHCP DNS servers for Advanced subnet type (optional, e.g., '8.8.8.8,8.8.4.4')")
	subnetGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	subnetUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	subnetUpdateCmd.Flags().String("name", "", "Subnet name (optional)")
	subnetUpdateCmd.Flags().String("cidr", "", "Subnet CIDR (optional)")
	subnetUpdateCmd.Flags().StringSlice("tags", []string{}, "Subnet tags (optional)")
	subnetUpdateCmd.Flags().Bool("dhcp-enabled", false, "Enable DHCP for Advanced subnet type")
	subnetUpdateCmd.Flags().StringSlice("dhcp-routes", []string{}, "DHCP routes for Advanced subnet type (optional, format: destination:gateway)")
	subnetUpdateCmd.Flags().StringSlice("dhcp-dns", []string{}, "DHCP DNS servers for Advanced subnet type (optional)")
	subnetDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	subnetDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	subnetDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	subnetListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	subnetListCmd.Flags().String("vpc-id", "", "Parent VPC ID (required)")
	subnetListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	subnetListCmd.Flags().Int32("offset", 0, "Number of results to skip")
}

// Subnet subcommands
var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Manage subnets",
	Long:  `Perform CRUD operations on subnets in Aruba Cloud.`,
}

var subnetCreateCmd = &cobra.Command{
	Use:   "create [vpc-id]",
	Short: "Create a new subnet",
	Long: `Create a new subnet in the specified VPC.

Without --cidr, a Basic subnet is created (the platform assigns the CIDR).
With --cidr, an Advanced subnet is created; --dhcp-enabled is required in that case.

DHCP routes format: "destination:gateway" (e.g., "10.1.0.0/24:10.0.0.1").`,
	Example: `  # Basic subnet (platform assigns CIDR)
  acloud network subnet create <vpc-id> --name my-subnet --region IT-BG

  # Advanced subnet with explicit CIDR and DHCP
  acloud network subnet create <vpc-id> --name my-subnet --region IT-BG \
    --cidr 10.0.1.0/24 --dhcp-enabled \
    --dhcp-dns 8.8.8.8,8.8.4.4`,
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		name, _ := cmd.Flags().GetString("name")
		cidr, _ := cmd.Flags().GetString("cidr")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		dhcpEnabled, _ := cmd.Flags().GetBool("dhcp-enabled")
		dhcpRoutes, _ := cmd.Flags().GetStringSlice("dhcp-routes")
		dhcpDNS, _ := cmd.Flags().GetStringSlice("dhcp-dns")
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

		s := aruba.NewSubnet().
			IntoVPC(aruba.VPCRef(projectID, vpcID)).
			Named(name).
			InRegion(aruba.Region(region))
		if len(tags) > 0 {
			s.ReplaceTags(tags...)
		}

		if cidr != "" {
			if !dhcpEnabled {
				return fmt.Errorf("--dhcp-enabled is required when creating an Advanced subnet (CIDR provided)")
			}
			s.OfType(aruba.SubnetTypeAdvanced).WithCIDR(cidr)

			d := aruba.NewSubnetDHCP().Enabled()
			for _, routeStr := range dhcpRoutes {
				parts := splitRouteString(routeStr)
				if len(parts) == 2 {
					d.AddRoute(parts[0], parts[1])
				} else {
					fmt.Printf("Warning: Invalid route format '%s', expected 'destination:gateway'. Skipping.\n", routeStr)
				}
			}
			for _, ip := range dhcpDNS {
				d.AddDNS(ip)
			}
			s.WithDHCP(d)
		} else {
			s.OfType(aruba.SubnetTypeBasic)
		}

		created, err := client.FromNetwork().Subnets().Create(ctx, s)
		if err != nil {
			return fmt.Errorf("creating subnet: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil && r.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "CIDR", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			displayCIDR := cidr
			if r.Properties.Network != nil && r.Properties.Network.Address != "" {
				displayCIDR = r.Properties.Network.Address
			}
			if displayCIDR == "" {
				displayCIDR = "N/A (Basic)"
			}
			createRegion := ""
			if r.Metadata.LocationResponse != nil {
				createRegion = string(r.Metadata.LocationResponse.Value)
			}
			status := ""
			if r.Status.State != nil {
				status = *r.Status.State
			}
			PrintOutput(r, headers, [][]string{{name, *r.Metadata.ID, createRegion, displayCIDR, status}})
		} else {
			fmt.Println(msgCreatedAsync("Subnet", name))
		}
		return nil
	},
}

var subnetGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [subnet-id]",
	Short: "Get subnet details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		subnetID := args[1]
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
		got, err := client.FromNetwork().Subnets().Get(ctx, aruba.SubnetRef(projectID, vpcID, subnetID))
		if err != nil {
			return fmt.Errorf("getting subnet: %w", apiErrFromV2(err))
		}

		subnet := got.Raw()
		if subnet != nil {
			fmt.Println("\nSubnet Details:")
			fmt.Println("===============")
			if subnet.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *subnet.Metadata.ID)
			}
			if subnet.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *subnet.Metadata.URI)
			}
			if subnet.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *subnet.Metadata.Name)
			}
			if subnet.Metadata.LocationResponse != nil && subnet.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", subnet.Metadata.LocationResponse.Value)
			}
			if subnet.Properties.Type != "" {
				fmt.Printf("Type:            %s\n", subnet.Properties.Type)
			}
			if subnet.Properties.Network != nil {
				fmt.Printf("CIDR:            %s\n", subnet.Properties.Network.Address)
			}
			if subnet.Properties.DHCP != nil {
				fmt.Printf("DHCP Enabled:    %v\n", subnet.Properties.DHCP.Enabled)
				if len(subnet.Properties.DHCP.Routes) > 0 {
					fmt.Printf("DHCP Routes:\n")
					for _, route := range subnet.Properties.DHCP.Routes {
						fmt.Printf("  - %s -> %s\n", route.Address, route.Gateway)
					}
				}
				if len(subnet.Properties.DHCP.DNS) > 0 {
					fmt.Printf("DHCP DNS:        %v\n", subnet.Properties.DHCP.DNS)
				}
			}
			if subnet.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", subnet.Metadata.CreationDate.Format(DateLayout))
			}
			if subnet.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *subnet.Metadata.CreatedBy)
			}
			if len(subnet.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", subnet.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			if subnet.Status.State != nil {
				fmt.Printf("Status:          %s\n", *subnet.Status.State)
			}
		} else {
			fmt.Println("Subnet not found or no data returned.")
		}
		return nil
	},
}

var subnetListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List subnets for a VPC",
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
		list, err := client.FromNetwork().Subnets().List(ctx, aruba.VPCRef(projectID, vpcID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing subnets: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "CIDR", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, subnet := range list.Items() {
				r := subnet.Raw()
				name := subnet.Name()
				id := subnet.SubnetID()
				region := ""
				cidr := ""
				status := ""
				if r != nil {
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					if r.Properties.Network != nil {
						cidr = r.Properties.Network.Address
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}
				rows = append(rows, []string{name, id, region, cidr, status})
			}
			PrintOutput(subnetListPayload(list), headers, rows)
		} else {
			fmt.Println("No subnets found.")
		}
		return nil
	},
}

var subnetUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [subnet-id]",
	Short: "Update a subnet",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		subnetID := args[1]
		name, _ := cmd.Flags().GetString("name")
		cidr, _ := cmd.Flags().GetString("cidr")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		dhcpEnabled, _ := cmd.Flags().GetBool("dhcp-enabled")
		dhcpRoutes, _ := cmd.Flags().GetStringSlice("dhcp-routes")
		dhcpDNS, _ := cmd.Flags().GetStringSlice("dhcp-dns")
		if name == "" && cidr == "" && !cmd.Flags().Changed("tags") && !cmd.Flags().Changed("dhcp-enabled") && !cmd.Flags().Changed("dhcp-routes") && !cmd.Flags().Changed("dhcp-dns") {
			return fmt.Errorf("at least one of --name, --cidr, --tags, --dhcp-enabled, --dhcp-routes, or --dhcp-dns must be provided")
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

		cur, err := client.FromNetwork().Subnets().Get(ctx, aruba.SubnetRef(projectID, vpcID, subnetID))
		if err != nil {
			return fmt.Errorf("fetching current subnet: %w", apiErrFromV2(err))
		}
		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("subnet not found")
		}
		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update subnet while it is in 'InCreation' state. Please wait until the subnet is fully created")
		}
		if cur.Region() == "" {
			return fmt.Errorf("unable to determine region value for subnet")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}
		if cidr != "" {
			cur.WithCIDR(cidr)
		}

		// Rebuild DHCP for Advanced subnets when DHCP flags are touched
		if r.Properties.Type == types.SubnetTypeAdvanced {
			if cmd.Flags().Changed("dhcp-enabled") || len(dhcpRoutes) > 0 || len(dhcpDNS) > 0 {
				d := aruba.NewSubnetDHCP()
				if r.Properties.DHCP != nil && r.Properties.DHCP.Enabled {
					d.Enabled()
				}
				if cmd.Flags().Changed("dhcp-enabled") && dhcpEnabled {
					d.Enabled()
				}
				existingRoutes := []aruba.SubnetDHCPRoute{}
				if r.Properties.DHCP != nil {
					for _, rt := range r.Properties.DHCP.Routes {
						existingRoutes = append(existingRoutes, aruba.SubnetDHCPRoute{Address: rt.Address, Gateway: rt.Gateway})
					}
				}
				if len(dhcpRoutes) > 0 {
					for _, routeStr := range dhcpRoutes {
						parts := splitRouteString(routeStr)
						if len(parts) == 2 {
							d.AddRoute(parts[0], parts[1])
						} else {
							fmt.Printf("Warning: Invalid route format '%s', expected 'destination:gateway'. Skipping.\n", routeStr)
						}
					}
				} else {
					for _, rt := range existingRoutes {
						d.AddRoute(rt.Address, rt.Gateway)
					}
				}
				existingDNS := []string{}
				if r.Properties.DHCP != nil {
					existingDNS = r.Properties.DHCP.DNS
				}
				if len(dhcpDNS) > 0 {
					for _, ip := range dhcpDNS {
						d.AddDNS(ip)
					}
				} else {
					for _, ip := range existingDNS {
						d.AddDNS(ip)
					}
				}
				cur.WithDHCP(d)
			}
		}

		updated, err := client.FromNetwork().Subnets().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating subnet: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "CIDR", Width: 18},
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
			updCIDR := ""
			if ur.Properties.Network != nil {
				updCIDR = ur.Properties.Network.Address
			}
			updStatus := ""
			if ur.Status.State != nil {
				updStatus = *ur.Status.State
			}
			PrintOutput(ur, headers, [][]string{{updName, updID, updCIDR, updStatus}})
		} else {
			fmt.Println(msgUpdatedAsync("Subnet", subnetID))
		}
		return nil
	},
}

var subnetDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [subnet-id]",
	Short: "Delete a subnet",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		subnetID := args[1]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("subnet", subnetID)
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
			_, err = client.FromNetwork().Subnets().Get(ctx, aruba.SubnetRef(projectID, vpcID, subnetID))
			if err != nil {
				return fmt.Errorf("dry-run: subnet not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("subnet", subnetID))
			return nil
		}

		if err := client.FromNetwork().Subnets().Delete(ctx, aruba.SubnetRef(projectID, vpcID, subnetID)); err != nil {
			return fmt.Errorf("deleting subnet: %w", apiErrFromV2(err))
		}

		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		result := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{subnetID, status}
		PrintOutput(result, headers, [][]string{{subnetID, status}})
		return nil
	},
}
