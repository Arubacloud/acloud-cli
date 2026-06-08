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
			id := route.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, route.Name()))
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
    --name my-route --region ITBG-Bergamo \
    --cloud-subnet 10.0.0.0/24 \
    --onprem-subnet 192.168.1.0/24`,
	Args: cobra.ExactArgs(1),
	RunE: NetworkVPNRouteCreateRun,
}

var vpnrouteGetCmd = &cobra.Command{
	Use:   "get [vpn-tunnel-id] [route-id]",
	Short: "Get VPN tunnel route details",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPNRouteGetRun,
}

var vpnrouteListCmd = &cobra.Command{
	Use:   "list [vpn-tunnel-id]",
	Short: "List VPN tunnel routes",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPNRouteListRun,
}

var vpnrouteUpdateCmd = &cobra.Command{
	Use:   "update [vpn-tunnel-id] [route-id]",
	Short: "Update a VPN tunnel route",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPNRouteUpdateRun,
}

var vpnrouteDeleteCmd = &cobra.Command{
	Use:   "delete [vpn-tunnel-id] [route-id]",
	Short: "Delete a VPN tunnel route",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPNRouteDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkVPNRouteCreateArgs holds the typed arguments for creating a VPN route.
type NetworkVPNRouteCreateArgs struct {
	ProjectID    string
	TunnelID     string
	Name         string
	Region       aruba.Region
	LocalSubnet  string
	RemoteSubnet string
	Tags         []string
	Verbose      bool
}

// NetworkVPNRouteGetArgs holds the typed arguments for getting a VPN route.
type NetworkVPNRouteGetArgs struct {
	ProjectID string
	TunnelID  string
	RouteID   string
}

// NetworkVPNRouteUpdateArgs holds the typed arguments for updating a VPN route.
type NetworkVPNRouteUpdateArgs struct {
	ProjectID   string
	TunnelID    string
	RouteID     string
	Name        string
	Tags        []string
	TagsChanged bool
}

// NetworkVPNRouteDeleteArgs holds the typed arguments for deleting a VPN route.
type NetworkVPNRouteDeleteArgs struct {
	ProjectID   string
	TunnelID    string
	RouteID     string
	DryRun      bool
	SkipConfirm bool
}

// NetworkVPNRouteListArgs holds the typed arguments for listing VPN routes.
type NetworkVPNRouteListArgs struct {
	ProjectID string
	TunnelID  string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkVPNRouteCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkVPNRouteCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNRouteCreateArgs, error) {
	args := &NetworkVPNRouteCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNRouteGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkVPNRouteGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNRouteGetArgs, error) {
	args := &NetworkVPNRouteGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNRouteUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkVPNRouteUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNRouteUpdateArgs, error) {
	args := &NetworkVPNRouteUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNRouteDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkVPNRouteDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNRouteDeleteArgs, error) {
	args := &NetworkVPNRouteDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNRouteListArgsFromCobraCommand parses and validates args for list.
func NewNetworkVPNRouteListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNRouteListArgs, error) {
	args := &NetworkVPNRouteListArgs{}
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
func (a *NetworkVPNRouteCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.TunnelID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if a.LocalSubnet, err = cmd.Flags().GetString("cloud-subnet"); err != nil {
		errs = append(errs, err)
	}
	if a.RemoteSubnet, err = cmd.Flags().GetString("onprem-subnet"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.Verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkVPNRouteGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.TunnelID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.RouteID = cobraArgs[1]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkVPNRouteUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.TunnelID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.RouteID = cobraArgs[1]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *NetworkVPNRouteDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.TunnelID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.RouteID = cobraArgs[1]
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
func (a *NetworkVPNRouteListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.TunnelID = cobraArgs[0]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *NetworkVPNRouteCreateArgs) Validate() error {
	var errs []error

	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}
	if !slices.Contains(validRegions, a.Region) {
		errs = append(errs, fmt.Errorf("--region %q: must be one of %v", a.Region, validRegions))
	}
	if a.TunnelID == "" {
		errs = append(errs, errors.New("VPN tunnel ID is required"))
	}
	if a.LocalSubnet == "" {
		errs = append(errs, errors.New("--cloud-subnet is required"))
	}
	if a.RemoteSubnet == "" {
		errs = append(errs, errors.New("--onprem-subnet is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkVPNRouteGetArgs) Validate() error {
	var errs []error
	if a.TunnelID == "" {
		errs = append(errs, errors.New("VPN tunnel ID is required"))
	}
	if a.RouteID == "" {
		errs = append(errs, errors.New("route ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *NetworkVPNRouteUpdateArgs) Validate() error {
	var errs []error
	if a.RouteID == "" {
		errs = append(errs, errors.New("route ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one field must be provided for update"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *NetworkVPNRouteDeleteArgs) Validate() error {
	var errs []error
	if a.RouteID == "" {
		errs = append(errs, errors.New("route ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *NetworkVPNRouteListArgs) Validate() error {
	var errs []error
	if a.TunnelID == "" {
		errs = append(errs, errors.New("VPN tunnel ID is required"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkVPNRouteCreate creates a VPN route using the provided args and client.
func NetworkVPNRouteCreate(ctx context.Context, client aruba.Client, args NetworkVPNRouteCreateArgs) error {
	if args.Verbose {
		fmt.Println("Creating VPN route with the following parameters:")
		fmt.Printf("  Name:          %s\n", args.Name)
		fmt.Printf("  Region:        %s\n", args.Region)
		fmt.Printf("  Cloud Subnet:  %s\n", args.LocalSubnet)
		fmt.Printf("  OnPrem Subnet: %s\n", args.RemoteSubnet)
		if len(args.Tags) > 0 {
			fmt.Printf("  Tags:          %v\n", args.Tags)
		}
		fmt.Println()
	}

	route := aruba.NewVPNRoute().
		InVPNTunnel(aruba.VPNTunnelRef(args.ProjectID, args.TunnelID)).
		Named(args.Name).
		InRegion(args.Region).
		WithCloudSubnet(args.LocalSubnet).
		WithOnPremSubnet(args.RemoteSubnet).
		RetaggedAs(args.Tags...)

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
		row := []string{args.Name, id, cloudSubnetVal, onPremSubnetVal, status}
		PrintOutput(resp, headers, [][]string{row})
	} else {
		fmt.Println(msgCreatedAsync("VPN route", args.Name))
	}
	return nil
}

// NetworkVPNRouteGet retrieves and displays a VPN route's details.
func NetworkVPNRouteGet(ctx context.Context, client aruba.Client, args NetworkVPNRouteGetArgs) error {
	route, err := client.FromNetwork().VPNRoutes().Get(ctx, aruba.VPNRouteRef(args.ProjectID, args.TunnelID, args.RouteID))
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
}

// NetworkVPNRouteUpdate updates a VPN route's name and/or tags.
func NetworkVPNRouteUpdate(ctx context.Context, client aruba.Client, args NetworkVPNRouteUpdateArgs) error {
	route, err := client.FromNetwork().VPNRoutes().Get(ctx, aruba.VPNRouteRef(args.ProjectID, args.TunnelID, args.RouteID))
	if err != nil || route == nil || route.Raw() == nil {
		return fmt.Errorf("fetching current VPN route: %w", apiErrFromV2(err))
	}

	if route.Raw().Status.State != nil && *route.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update VPN route while it is in 'InCreation' state. Please wait until the VPN route is fully created")
	}

	if args.Name != "" {
		route.Named(args.Name)
	}
	if args.TagsChanged {
		route.RetaggedAs(args.Tags...)
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
		PrintOutput(updated, headers, [][]string{row})
	} else {
		fmt.Println(msgUpdatedAsync("VPN route", args.RouteID))
	}
	return nil
}

// NetworkVPNRouteDelete deletes a VPN route.
func NetworkVPNRouteDelete(ctx context.Context, client aruba.Client, args NetworkVPNRouteDeleteArgs) error {
	err := client.FromNetwork().VPNRoutes().Delete(ctx, aruba.VPNRouteRef(args.ProjectID, args.TunnelID, args.RouteID))
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
	}{args.RouteID, status}
	PrintOutput(result, headers, [][]string{{args.RouteID, status}})
	return nil
}

// NetworkVPNRouteList lists VPN routes under a tunnel.
func NetworkVPNRouteList(ctx context.Context, client aruba.Client, args NetworkVPNRouteListArgs) error {
	list, err := client.FromNetwork().VPNRoutes().List(ctx, aruba.VPNTunnelRef(args.ProjectID, args.TunnelID), args.CallOpts...)
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
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No VPN routes found.")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkVPNRouteCreateRun is the Cobra RunE handler for VPN route create.
func NetworkVPNRouteCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNRouteCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNRouteCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNRouteGetRun is the Cobra RunE handler for VPN route get.
func NetworkVPNRouteGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNRouteGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNRouteGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNRouteUpdateRun is the Cobra RunE handler for VPN route update.
func NetworkVPNRouteUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNRouteUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNRouteUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNRouteDeleteRun is the Cobra RunE handler for VPN route delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkVPNRouteDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNRouteDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("VPN route", args.RouteID)
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
		if _, err := client.FromNetwork().VPNRoutes().Get(ctx, aruba.VPNRouteRef(args.ProjectID, args.TunnelID, args.RouteID)); err != nil {
			return fmt.Errorf("dry-run: VPN route not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("VPN route", args.RouteID))
		return nil
	}

	if err := NetworkVPNRouteDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNRouteListRun is the Cobra RunE handler for VPN route list.
func NetworkVPNRouteListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNRouteListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNRouteList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
