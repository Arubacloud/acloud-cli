package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
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
	RunE: NetworkSubnetCreateRun,
}

var subnetGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [subnet-id]",
	Short: "Get subnet details",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSubnetGetRun,
}

var subnetListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List subnets for a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkSubnetListRun,
}

var subnetUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [subnet-id]",
	Short: "Update a subnet",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSubnetUpdateRun,
}

var subnetDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [subnet-id]",
	Short: "Delete a subnet",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSubnetDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkSubnetCreateArgs holds the typed arguments for creating a subnet.
type NetworkSubnetCreateArgs struct {
	ProjectID      string
	VPCID          string
	Name           string
	Region         aruba.Region
	CIDR           string
	Tags           []string
	DHCPEnabled    bool
	DHCPRoutes     []string
	DHCPDNSServers []string
}

// NetworkSubnetGetArgs holds the typed arguments for getting a subnet.
type NetworkSubnetGetArgs struct {
	ProjectID string
	VPCID     string
	SubnetID  string
}

// NetworkSubnetUpdateArgs holds the typed arguments for updating a subnet.
type NetworkSubnetUpdateArgs struct {
	ProjectID          string
	VPCID              string
	SubnetID           string
	Name               string
	CIDR               string
	Tags               []string
	TagsChanged        bool
	DHCPEnabled        bool
	DHCPEnabledChanged bool
	DHCPRoutes         []string
	DHCPRoutesChanged  bool
	DHCPDNSServers     []string
	DHCPDNSChanged     bool
}

// NetworkSubnetDeleteArgs holds the typed arguments for deleting a subnet.
type NetworkSubnetDeleteArgs struct {
	ProjectID   string
	VPCID       string
	SubnetID    string
	DryRun      bool
	SkipConfirm bool
}

// NetworkSubnetListArgs holds the typed arguments for listing subnets.
type NetworkSubnetListArgs struct {
	ProjectID string
	VPCID     string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkSubnetCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkSubnetCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSubnetCreateArgs, error) {
	args := &NetworkSubnetCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSubnetGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkSubnetGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSubnetGetArgs, error) {
	args := &NetworkSubnetGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSubnetUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkSubnetUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSubnetUpdateArgs, error) {
	args := &NetworkSubnetUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSubnetDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkSubnetDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSubnetDeleteArgs, error) {
	args := &NetworkSubnetDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSubnetListArgsFromCobraCommand parses and validates args for list.
func NewNetworkSubnetListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSubnetListArgs, error) {
	args := &NetworkSubnetListArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// =============================================================================
// ParseFromCobraCommand methods
// =============================================================================

// ParseFromCobraCommand reads Cobra flags and positional args into the create args struct.
func (a *NetworkSubnetCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if a.CIDR, err = cmd.Flags().GetString("cidr"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.DHCPEnabled, err = cmd.Flags().GetBool("dhcp-enabled"); err != nil {
		errs = append(errs, err)
	}
	if a.DHCPRoutes, err = cmd.Flags().GetStringSlice("dhcp-routes"); err != nil {
		errs = append(errs, err)
	}
	if a.DHCPDNSServers, err = cmd.Flags().GetStringSlice("dhcp-dns"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkSubnetGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.SubnetID = cobraArgs[1]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkSubnetUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.SubnetID = cobraArgs[1]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.CIDR, err = cmd.Flags().GetString("cidr"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")
	if a.DHCPEnabled, err = cmd.Flags().GetBool("dhcp-enabled"); err != nil {
		errs = append(errs, err)
	}
	a.DHCPEnabledChanged = cmd.Flags().Changed("dhcp-enabled")
	if a.DHCPRoutes, err = cmd.Flags().GetStringSlice("dhcp-routes"); err != nil {
		errs = append(errs, err)
	}
	a.DHCPRoutesChanged = cmd.Flags().Changed("dhcp-routes")
	if a.DHCPDNSServers, err = cmd.Flags().GetStringSlice("dhcp-dns"); err != nil {
		errs = append(errs, err)
	}
	a.DHCPDNSChanged = cmd.Flags().Changed("dhcp-dns")

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *NetworkSubnetDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.SubnetID = cobraArgs[1]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the list args struct.
// VPCID comes from positional cobraArgs[0], not from the --vpc-id flag.
func (a *NetworkSubnetListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *NetworkSubnetCreateArgs) Validate() error {
	var errs []error

	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}
	if !slices.Contains(validRegions, a.Region) {
		errs = append(errs, fmt.Errorf("--region %q: must be one of %v", a.Region, validRegions))
	}
	if a.CIDR != "" && !a.DHCPEnabled {
		errs = append(errs, errors.New("--dhcp-enabled is required when creating an Advanced subnet (CIDR provided)"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkSubnetGetArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.SubnetID == "" {
		errs = append(errs, errors.New("subnet ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *NetworkSubnetUpdateArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.SubnetID == "" {
		errs = append(errs, errors.New("subnet ID is required"))
	}
	if a.Name == "" && a.CIDR == "" && !a.TagsChanged && !a.DHCPEnabledChanged && !a.DHCPRoutesChanged && !a.DHCPDNSChanged {
		errs = append(errs, errors.New("at least one of --name, --cidr, --tags, --dhcp-enabled, --dhcp-routes, or --dhcp-dns must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *NetworkSubnetDeleteArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.SubnetID == "" {
		errs = append(errs, errors.New("subnet ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *NetworkSubnetListArgs) Validate() error {
	var errs []error
	if a.ProjectID == "" {
		errs = append(errs, errors.New("project ID is required"))
	}
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkSubnetCreate creates a subnet using the provided args and client.
func NetworkSubnetCreate(ctx context.Context, client aruba.Client, args NetworkSubnetCreateArgs) error {
	subnetType := aruba.SubnetTypeBasic
	if args.CIDR != "" {
		subnetType = aruba.SubnetTypeAdvanced
	}

	subnet := aruba.NewSubnet().
		InVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		Named(args.Name).
		InRegion(args.Region).
		OfType(subnetType).
		RetaggedAs(args.Tags...)

	if args.CIDR != "" {
		subnet = subnet.WithCIDR(args.CIDR)
	}

	if args.DHCPEnabled {
		dhcp := aruba.NewSubnetDHCP().Enabled()
		for _, routeStr := range args.DHCPRoutes {
			parts := splitRouteString(routeStr)
			if len(parts) == 2 {
				dhcp = dhcp.WithRoutes(aruba.SubnetDHCPRouteCommon{Address: parts[0], Gateway: parts[1]})
			} else {
				fmt.Printf("Warning: Invalid route format '%s', expected 'destination:gateway'. Skipping.\n", routeStr)
			}
		}
		if len(args.DHCPDNSServers) > 0 {
			dhcp = dhcp.WithDNSServers(args.DHCPDNSServers...)
		}
		subnet = subnet.WithDHCP(dhcp)
	}

	resp, err := client.FromNetwork().Subnets().Create(ctx, subnet)
	if err != nil {
		return fmt.Errorf("creating subnet: %w", apiErrFromV2(err))
	}
	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "REGION", Width: 18},
			{Header: "CIDR", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		displayCIDR := args.CIDR
		if raw.Properties.Network != nil && raw.Properties.Network.Address != "" {
			displayCIDR = raw.Properties.Network.Address
		}
		if displayCIDR == "" {
			displayCIDR = "N/A (Basic)"
		}
		createRegion := ""
		if raw.Metadata.LocationResponse != nil {
			createRegion = string(raw.Metadata.LocationResponse.Value)
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		PrintOutput(resp, headers, [][]string{{args.Name, id, createRegion, displayCIDR, status}})
	} else {
		fmt.Println(msgCreatedAsync("Subnet", args.Name))
	}
	return nil
}

// NetworkSubnetGet retrieves and displays a subnet's details.
func NetworkSubnetGet(ctx context.Context, client aruba.Client, args NetworkSubnetGetArgs) error {
	subnet, err := client.FromNetwork().Subnets().Get(ctx, aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID))
	if err != nil {
		return fmt.Errorf("getting subnet: %w", apiErrFromV2(err))
	}
	if subnet != nil && subnet.Raw() != nil {
		raw := subnet.Raw()
		fmt.Println("\nSubnet Details:")
		fmt.Println("===============")
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Metadata.LocationResponse != nil && string(raw.Metadata.LocationResponse.Value) != "" {
			fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
		}
		if raw.Properties.Type != "" {
			fmt.Printf("Type:            %s\n", raw.Properties.Type)
		}
		if raw.Properties.Network != nil {
			fmt.Printf("CIDR:            %s\n", raw.Properties.Network.Address)
		}
		if raw.Properties.DHCP != nil {
			fmt.Printf("DHCP Enabled:    %v\n", raw.Properties.DHCP.Enabled)
			if len(raw.Properties.DHCP.Routes) > 0 {
				fmt.Printf("DHCP Routes:\n")
				for _, route := range raw.Properties.DHCP.Routes {
					fmt.Printf("  - %s -> %s\n", route.Address, route.Gateway)
				}
			}
			if len(raw.Properties.DHCP.DNS) > 0 {
				fmt.Printf("DHCP DNS:        %v\n", raw.Properties.DHCP.DNS)
			}
		}
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
		fmt.Println("Subnet not found or no data returned.")
	}
	return nil
}

// NetworkSubnetList lists subnets in the given VPC.
func NetworkSubnetList(ctx context.Context, client aruba.Client, args NetworkSubnetListArgs) error {
	list, err := client.FromNetwork().Subnets().List(ctx, aruba.VPCRef(args.ProjectID, args.VPCID))
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
			raw := subnet.Raw()
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
			region := ""
			if raw.Metadata.LocationResponse != nil {
				region = string(raw.Metadata.LocationResponse.Value)
			}
			cidr := ""
			if raw.Properties.Network != nil {
				cidr = raw.Properties.Network.Address
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, region, cidr, status})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No subnets found.")
	}
	return nil
}

// NetworkSubnetUpdate updates a subnet using the provided args and client.
func NetworkSubnetUpdate(ctx context.Context, client aruba.Client, args NetworkSubnetUpdateArgs) error {
	subnet, err := client.FromNetwork().Subnets().Get(ctx, aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID))
	if err != nil || subnet == nil || subnet.ID() == "" {
		return fmt.Errorf("fetching current subnet: %w", apiErrFromV2(err))
	}

	if subnet.State() == StateInCreation {
		return fmt.Errorf("cannot update subnet while it is in 'InCreation' state. Please wait until the subnet is fully created")
	}

	if args.Name != "" {
		subnet.Named(args.Name)
	}
	if args.TagsChanged {
		subnet.RetaggedAs(args.Tags...)
	}
	if args.CIDR != "" {
		subnet.WithCIDR(args.CIDR)
	}

	// Update DHCP config for Advanced subnets.
	// sdk-go v1.0.0 exposes Type() and DHCP() accessors (closes #133 / TD-035).
	if subnet.Type() == aruba.SubnetTypeAdvanced {
		currentDHCP := subnet.DHCP()
		if args.DHCPEnabledChanged || args.DHCPRoutesChanged || args.DHCPDNSChanged {
			dhcp := aruba.NewSubnetDHCP()
			// Preserve existing enabled state
			if (currentDHCP != nil && currentDHCP.IsEnabled()) || args.DHCPEnabled {
				dhcp = dhcp.Enabled()
			}
			// Preserve existing routes unless new ones provided
			if len(args.DHCPRoutes) > 0 {
				for _, routeStr := range args.DHCPRoutes {
					parts := splitRouteString(routeStr)
					if len(parts) == 2 {
						dhcp = dhcp.WithRoutes(aruba.SubnetDHCPRouteCommon{Address: parts[0], Gateway: parts[1]})
					} else {
						fmt.Printf("Warning: Invalid route format '%s', expected 'destination:gateway'. Skipping.\n", routeStr)
					}
				}
			} else if currentDHCP != nil {
				for _, r := range currentDHCP.Routes() {
					dhcp = dhcp.WithRoutes(aruba.SubnetDHCPRouteCommon{Address: r.Address, Gateway: r.Gateway})
				}
			}
			// Preserve existing DNS unless new ones provided
			if len(args.DHCPDNSServers) > 0 {
				dhcp = dhcp.WithDNSServers(args.DHCPDNSServers...)
			} else if currentDHCP != nil && len(currentDHCP.DNS()) > 0 {
				dhcp = dhcp.WithDNSServers(currentDHCP.DNS()...)
			}
			subnet.WithDHCP(dhcp)
		}
	}

	updated, err := client.FromNetwork().Subnets().Update(ctx, subnet)
	if err != nil {
		return fmt.Errorf("updating subnet: %w", apiErrFromV2(err))
	}
	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "CIDR", Width: 18},
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
		cidrVal := ""
		if raw.Properties.Network != nil {
			cidrVal = raw.Properties.Network.Address
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		PrintOutput(updated, headers, [][]string{{nameVal, id, cidrVal, status}})
	} else {
		fmt.Println(msgUpdatedAsync("Subnet", args.SubnetID))
	}
	return nil
}

// NetworkSubnetDelete deletes a subnet.
func NetworkSubnetDelete(ctx context.Context, client aruba.Client, args NetworkSubnetDeleteArgs) error {
	err := client.FromNetwork().Subnets().Delete(ctx, aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID))
	if err != nil {
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
	}{args.SubnetID, status}
	PrintOutput(result, headers, [][]string{{args.SubnetID, status}})
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkSubnetCreateRun is the Cobra RunE handler for subnet create.
func NetworkSubnetCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSubnetCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSubnetCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSubnetGetRun is the Cobra RunE handler for subnet get.
func NetworkSubnetGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSubnetGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSubnetGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSubnetListRun is the Cobra RunE handler for subnet list.
func NetworkSubnetListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSubnetListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSubnetList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSubnetUpdateRun is the Cobra RunE handler for subnet update.
func NetworkSubnetUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSubnetUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSubnetUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSubnetDeleteRun is the Cobra RunE handler for subnet delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkSubnetDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSubnetDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("subnet", args.SubnetID)
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

	if args.DryRun {
		_, err = client.FromNetwork().Subnets().Get(ctx, aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID))
		if err != nil {
			return fmt.Errorf("dry-run: subnet not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("subnet", args.SubnetID))
		return nil
	}

	if err := NetworkSubnetDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
