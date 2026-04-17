package cmd

import (
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// splitRouteString splits a route string in format "destination:gateway"
func splitRouteString(routeStr string) []string {
	return strings.SplitN(routeStr, ":", 2)
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

		// Determine SubnetType: Advanced if CIDR is provided, Basic otherwise
		var subnetType types.SubnetType = types.SubnetTypeBasic
		if cidr != "" {
			subnetType = types.SubnetTypeAdvanced
			// For Advanced subnet type, DHCP enabled is required
			if !dhcpEnabled {
				return fmt.Errorf("--dhcp-enabled is required when creating an Advanced subnet (CIDR provided)")
			}
		}

		// Build DHCP configuration for Advanced subnet type
		var dhcpConfig *types.SubnetDHCP
		if subnetType == types.SubnetTypeAdvanced && dhcpEnabled {
			dhcpConfig = &types.SubnetDHCP{
				Enabled: dhcpEnabled,
			}

			// Parse DHCP routes if provided
			if len(dhcpRoutes) > 0 {
				var routes []types.SubnetDHCPRoute
				for _, routeStr := range dhcpRoutes {
					// Parse route in format "destination:gateway" (e.g., "0.0.0.0/0:10.0.0.1")
					parts := splitRouteString(routeStr)
					if len(parts) == 2 {
						routes = append(routes, types.SubnetDHCPRoute{
							Address: parts[0],
							Gateway: parts[1],
						})
					} else {
						fmt.Printf("Warning: Invalid route format '%s', expected 'destination:gateway'. Skipping.\n", routeStr)
					}
				}
				if len(routes) > 0 {
					dhcpConfig.Routes = routes
				}
			}

			// Set DNS servers if provided
			if len(dhcpDNS) > 0 {
				dhcpConfig.DNS = dhcpDNS
			}
		}

		req := types.SubnetRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.SubnetPropertiesRequest{
				Type: subnetType,
				Network: func() *types.SubnetNetwork {
					if cidr != "" {
						return &types.SubnetNetwork{
							Address: cidr,
						}
					}
					return nil
				}(),
				DHCP: dhcpConfig,
			},
		}
		resp, err := client.FromNetwork().Subnets().Create(ctx, projectID, vpcID, req, nil)
		if err != nil {
			return fmt.Errorf("creating subnet: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil && resp.Data.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "CIDR", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			// Get CIDR from response or use provided value
			displayCIDR := cidr
			if resp.Data.Properties.Network != nil && resp.Data.Properties.Network.Address != "" {
				displayCIDR = resp.Data.Properties.Network.Address
			}
			if displayCIDR == "" {
				displayCIDR = "N/A (Basic)"
			}

			createRegion := ""
			if resp.Data.Metadata.LocationResponse != nil {
				createRegion = resp.Data.Metadata.LocationResponse.Value
			}
			row := []string{
				name,
				*resp.Data.Metadata.ID,
				createRegion,
				displayCIDR,
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
		resp, err := client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
		if err != nil {
			return fmt.Errorf("getting subnet: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil {
			subnet := resp.Data
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
			// Display DHCP information for Advanced subnets
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
		resp, err := client.FromNetwork().Subnets().List(ctx, projectID, vpcID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing subnets: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "CIDR", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, subnet := range resp.Data.Values {
				name := ""
				if subnet.Metadata.Name != nil {
					name = *subnet.Metadata.Name
				}
				id := ""
				if subnet.Metadata.ID != nil {
					id = *subnet.Metadata.ID
				}
				region := ""
				if subnet.Metadata.LocationResponse != nil {
					region = subnet.Metadata.LocationResponse.Value
				}
				cidr := ""
				if subnet.Properties.Network != nil {
					cidr = subnet.Properties.Network.Address
				}
				status := ""
				if subnet.Status.State != nil {
					status = *subnet.Status.State
				}
				rows = append(rows, []string{name, id, region, cidr, status})
			}
			PrintTable(headers, rows)
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
		if name == "" && cidr == "" && !cmd.Flags().Changed("tags") && !cmd.Flags().Changed("dhcp-enabled") && len(dhcpRoutes) == 0 && len(dhcpDNS) == 0 {
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
		// Enable debug logging if supported
		if logger, ok := interface{}(client).(interface{ SetDebug(bool) }); ok {
			logger.SetDebug(true)
		}
		ctx, cancel := newCtx()
		defer cancel()
		// Fetch current subnet details
		getResp, err := client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			return fmt.Errorf("fetching current subnet: %w", err)
		}
		current := getResp.Data
		// Block update if subnet is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == StateInCreation {
			return fmt.Errorf("cannot update subnet while it is in 'InCreation' state. Please wait until the subnet is fully created")
		}

		// Get region value
		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for subnet")
		}

		// Determine if this is an Advanced subnet
		isAdvanced := current.Properties.Type == types.SubnetTypeAdvanced

		// Build DHCP configuration for Advanced subnet type
		var dhcpConfig *types.SubnetDHCP
		if isAdvanced {
			// Start with current DHCP config if it exists
			if current.Properties.DHCP != nil {
				dhcpConfig = &types.SubnetDHCP{
					Enabled: current.Properties.DHCP.Enabled,
					Routes:  current.Properties.DHCP.Routes,
					DNS:     current.Properties.DHCP.DNS,
				}
			} else {
				dhcpConfig = &types.SubnetDHCP{}
			}

			// Update DHCP enabled if flag is provided
			if cmd.Flags().Changed("dhcp-enabled") {
				dhcpConfig.Enabled = dhcpEnabled
			}

			// Update DHCP routes if provided
			if len(dhcpRoutes) > 0 {
				var routes []types.SubnetDHCPRoute
				for _, routeStr := range dhcpRoutes {
					parts := splitRouteString(routeStr)
					if len(parts) == 2 {
						routes = append(routes, types.SubnetDHCPRoute{
							Address: parts[0],
							Gateway: parts[1],
						})
					} else {
						fmt.Printf("Warning: Invalid route format '%s', expected 'destination:gateway'. Skipping.\n", routeStr)
					}
				}
				if len(routes) > 0 {
					dhcpConfig.Routes = routes
				}
			}

			// Update DNS servers if provided
			if len(dhcpDNS) > 0 {
				dhcpConfig.DNS = dhcpDNS
			}
		}

		// Build update request by merging user input with all current valid fields
		req := types.SubnetRequest{
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
			Properties: types.SubnetPropertiesRequest{
				Type: current.Properties.Type,
				Network: &types.SubnetNetwork{
					Address: func() string {
						if cidr != "" {
							return cidr
						}
						if current.Properties.Network != nil {
							return current.Properties.Network.Address
						}
						return ""
					}(),
				},
				DHCP: dhcpConfig,
			},
		}

		resp, err := client.FromNetwork().Subnets().Update(ctx, projectID, vpcID, subnetID, req, nil)
		if err != nil {
			return fmt.Errorf("updating subnet: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "CIDR", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			name := ""
			if resp.Data.Metadata.Name != nil {
				name = *resp.Data.Metadata.Name
			}
			id := ""
			if resp.Data.Metadata.ID != nil {
				id = *resp.Data.Metadata.ID
			}
			cidr := ""
			if resp.Data.Properties.Network != nil {
				cidr = resp.Data.Properties.Network.Address
			}
			status := ""
			if resp.Data.Status.State != nil {
				status = *resp.Data.Status.State
			}
			PrintTable(headers, [][]string{{name, id, cidr, status}})
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

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
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
			_, err = client.FromNetwork().Subnets().Get(ctx, projectID, vpcID, subnetID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: subnet not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("subnet", subnetID))
			return nil
		}

		resp, err := client.FromNetwork().Subnets().Delete(ctx, projectID, vpcID, subnetID, nil)
		if err != nil {
			return fmt.Errorf("deleting subnet: %w", err)
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}
		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		PrintTable(headers, [][]string{{subnetID, status}})
		return nil
	},
}
