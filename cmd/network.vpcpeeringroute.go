package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
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
	vpcpeeringrouteCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
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

	vpcID := args[0]
	vpcPeeringID := args[1]

	key := cacheKey("vpc-peering-route", projectID, vpcID, vpcPeeringID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, aruba.VPCPeeringRef(projectID, vpcID, vpcPeeringID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, route := range list.Items() {
			id := route.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, route.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
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
	RunE: NetworkVPCPeeringRouteCreateRun,
}

var vpcpeeringrouteGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [peering-id] [route-id]",
	Short: "Get VPC peering route details",
	Args:  cobra.ExactArgs(3),
	RunE:  NetworkVPCPeeringRouteGetRun,
}

var vpcpeeringrouteListCmd = &cobra.Command{
	Use:   "list [vpc-id] [peering-id]",
	Short: "List VPC peering routes",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPCPeeringRouteListRun,
}

var vpcpeeringrouteUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [peering-id] [route-id]",
	Short: "Update a VPC peering route",
	Args:  cobra.ExactArgs(3),
	RunE:  NetworkVPCPeeringRouteUpdateRun,
}

var vpcpeeringrouteDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [peering-id] [route-id]",
	Short: "Delete a VPC peering route",
	Args:  cobra.ExactArgs(3),
	RunE:  NetworkVPCPeeringRouteDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkVPCPeeringRouteCreateArgs holds the typed arguments for creating a VPC peering route.
type NetworkVPCPeeringRouteCreateArgs struct {
	ProjectID     string
	VPCID         string
	PeeringID     string
	Name          string
	Region        aruba.Region
	LocalNetwork  string
	RemoteNetwork string
	BillingPeriod aruba.BillingPeriod
	Tags          []string
	Verbose       bool
}

// NetworkVPCPeeringRouteGetArgs holds the typed arguments for getting a VPC peering route.
type NetworkVPCPeeringRouteGetArgs struct {
	ProjectID string
	VPCID     string
	PeeringID string
	RouteID   string
}

// NetworkVPCPeeringRouteUpdateArgs holds the typed arguments for updating a VPC peering route.
type NetworkVPCPeeringRouteUpdateArgs struct {
	ProjectID   string
	VPCID       string
	PeeringID   string
	RouteID     string
	Name        string
	Tags        []string
	TagsChanged bool
}

// NetworkVPCPeeringRouteDeleteArgs holds the typed arguments for deleting a VPC peering route.
type NetworkVPCPeeringRouteDeleteArgs struct {
	ProjectID   string
	VPCID       string
	PeeringID   string
	RouteID     string
	DryRun      bool
	SkipConfirm bool
}

// NetworkVPCPeeringRouteListArgs holds the typed arguments for listing VPC peering routes.
type NetworkVPCPeeringRouteListArgs struct {
	ProjectID string
	VPCID     string
	PeeringID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkVPCPeeringRouteCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkVPCPeeringRouteCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringRouteCreateArgs, error) {
	args := &NetworkVPCPeeringRouteCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringRouteGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkVPCPeeringRouteGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringRouteGetArgs, error) {
	args := &NetworkVPCPeeringRouteGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringRouteUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkVPCPeeringRouteUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringRouteUpdateArgs, error) {
	args := &NetworkVPCPeeringRouteUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringRouteDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkVPCPeeringRouteDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringRouteDeleteArgs, error) {
	args := &NetworkVPCPeeringRouteDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringRouteListArgsFromCobraCommand parses and validates args for list.
func NewNetworkVPCPeeringRouteListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringRouteListArgs, error) {
	args := &NetworkVPCPeeringRouteListArgs{}
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
func (a *NetworkVPCPeeringRouteCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if a.LocalNetwork, err = cmd.Flags().GetString("local-network"); err != nil {
		errs = append(errs, err)
	}
	if a.RemoteNetwork, err = cmd.Flags().GetString("remote-network"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
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
func (a *NetworkVPCPeeringRouteGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	if len(cobraArgs) > 2 {
		a.RouteID = cobraArgs[2]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkVPCPeeringRouteUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	if len(cobraArgs) > 2 {
		a.RouteID = cobraArgs[2]
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
func (a *NetworkVPCPeeringRouteDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	if len(cobraArgs) > 2 {
		a.RouteID = cobraArgs[2]
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
func (a *NetworkVPCPeeringRouteListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *NetworkVPCPeeringRouteCreateArgs) Validate() error {
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
	if !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	if a.LocalNetwork == "" {
		errs = append(errs, errors.New("--local-network is required"))
	}
	if a.RemoteNetwork == "" {
		errs = append(errs, errors.New("--remote-network is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkVPCPeeringRouteGetArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.PeeringID == "" {
		errs = append(errs, errors.New("peering ID is required"))
	}
	if a.RouteID == "" {
		errs = append(errs, errors.New("route ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *NetworkVPCPeeringRouteUpdateArgs) Validate() error {
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
func (a *NetworkVPCPeeringRouteDeleteArgs) Validate() error {
	var errs []error
	if a.RouteID == "" {
		errs = append(errs, errors.New("route ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *NetworkVPCPeeringRouteListArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.PeeringID == "" {
		errs = append(errs, errors.New("peering ID is required"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkVPCPeeringRouteCreate creates a VPC peering route using the provided args and client.
func NetworkVPCPeeringRouteCreate(ctx context.Context, client aruba.Client, args NetworkVPCPeeringRouteCreateArgs) error {
	if args.Verbose {
		fmt.Println("Creating VPC peering route with the following parameters:")
		fmt.Printf("  Name:            %s\n", args.Name)
		fmt.Printf("  Local Network:   %s\n", args.LocalNetwork)
		fmt.Printf("  Remote Network:  %s\n", args.RemoteNetwork)
		fmt.Printf("  Billing Period:  %s\n", args.BillingPeriod)
		if len(args.Tags) > 0 {
			fmt.Printf("  Tags:            %v\n", args.Tags)
		}
		fmt.Println()
	}

	route := aruba.NewVPCPeeringRoute().
		InVPCPeering(aruba.VPCPeeringRef(args.ProjectID, args.VPCID, args.PeeringID)).
		Named(args.Name).
		InRegion(args.Region).
		WithLocalCIDR(args.LocalNetwork).
		WithRemoteCIDR(args.RemoteNetwork).
		BilledBy(args.BillingPeriod).
		RetaggedAs(args.Tags...)

	resp, err := client.FromNetwork().VPCPeeringRoutes().Create(ctx, route)
	if err != nil {
		return fmt.Errorf("creating VPC peering route: %w", apiErrFromV2(err))
	}

	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "LOCAL NETWORK", Width: 18},
			{Header: "REMOTE NETWORK", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		nameVal := args.Name
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		localNet := ""
		if raw.Properties.LocalNetworkAddress != "" {
			localNet = raw.Properties.LocalNetworkAddress
		}
		remoteNet := ""
		if raw.Properties.RemoteNetworkAddress != "" {
			remoteNet = raw.Properties.RemoteNetworkAddress
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		row := []string{nameVal, id, localNet, remoteNet, status}
		PrintOutput(resp, headers, [][]string{row})
	} else {
		fmt.Println(msgCreatedAsync("VPC peering route", args.Name))
	}
	return nil
}

// NetworkVPCPeeringRouteGet retrieves and displays a VPC peering route's details.
func NetworkVPCPeeringRouteGet(ctx context.Context, client aruba.Client, args NetworkVPCPeeringRouteGetArgs) error {
	route, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, aruba.VPCPeeringRouteRef(args.ProjectID, args.VPCID, args.PeeringID, args.RouteID))
	if err != nil {
		return fmt.Errorf("getting VPC peering route: %w", apiErrFromV2(err))
	}

	if route != nil && route.Raw() != nil {
		raw := route.Raw()
		fmt.Println("\nVPC Peering Route Details:")
		fmt.Println("==========================")
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		} else {
			fmt.Printf("ID:              %s\n", args.RouteID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		fmt.Printf("Local Network:    %s\n", raw.Properties.LocalNetworkAddress)
		fmt.Printf("Remote Network:   %s\n", raw.Properties.RemoteNetworkAddress)
		if raw.Properties.BillingPlanCommon != nil && raw.Properties.BillingPlanCommon.BillingPeriod != nil {
			fmt.Printf("Billing Period:  %s\n", *raw.Properties.BillingPlanCommon.BillingPeriod)
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
		fmt.Println("VPC peering route not found or no data returned.")
	}
	return nil
}

// NetworkVPCPeeringRouteUpdate updates a VPC peering route's fields.
func NetworkVPCPeeringRouteUpdate(ctx context.Context, client aruba.Client, args NetworkVPCPeeringRouteUpdateArgs) error {
	route, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, aruba.VPCPeeringRouteRef(args.ProjectID, args.VPCID, args.PeeringID, args.RouteID))
	if err != nil || route == nil || route.Raw() == nil {
		return fmt.Errorf("fetching current VPC peering route: %w", apiErrFromV2(err))
	}

	if route.Raw().Status.State != nil && *route.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update VPC peering route while it is in 'InCreation' state. Please wait until the VPC peering route is fully created")
	}

	if args.Name != "" {
		route.Named(args.Name)
	}
	if args.TagsChanged {
		route.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromNetwork().VPCPeeringRoutes().Update(ctx, route)
	if err != nil {
		return fmt.Errorf("updating VPC peering route: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "LOCAL NETWORK", Width: 18},
			{Header: "REMOTE NETWORK", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		id := args.RouteID
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		localNet := raw.Properties.LocalNetworkAddress
		remoteNet := raw.Properties.RemoteNetworkAddress
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		row := []string{nameVal, id, localNet, remoteNet, status}
		PrintOutput(updated, headers, [][]string{row})
	} else {
		fmt.Println(msgUpdatedAsync("VPC peering route", args.RouteID))
	}
	return nil
}

// NetworkVPCPeeringRouteDelete deletes a VPC peering route.
func NetworkVPCPeeringRouteDelete(ctx context.Context, client aruba.Client, args NetworkVPCPeeringRouteDeleteArgs) error {
	err := client.FromNetwork().VPCPeeringRoutes().Delete(ctx, aruba.VPCPeeringRouteRef(args.ProjectID, args.VPCID, args.PeeringID, args.RouteID))
	if err != nil {
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
	}{args.RouteID, status}
	PrintOutput(result, headers, [][]string{{args.RouteID, status}})
	return nil
}

// NetworkVPCPeeringRouteList lists VPC peering routes under a peering connection.
func NetworkVPCPeeringRouteList(ctx context.Context, client aruba.Client, args NetworkVPCPeeringRouteListArgs) error {
	list, err := client.FromNetwork().VPCPeeringRoutes().List(ctx, aruba.VPCPeeringRef(args.ProjectID, args.VPCID, args.PeeringID), args.CallOpts...)
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
			localNetwork := raw.Properties.LocalNetworkAddress
			remoteNetwork := raw.Properties.RemoteNetworkAddress
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, localNetwork, remoteNetwork, status})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No VPC peering routes found.")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkVPCPeeringRouteCreateRun is the Cobra RunE handler for VPC peering route create.
func NetworkVPCPeeringRouteCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringRouteCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringRouteCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringRouteGetRun is the Cobra RunE handler for VPC peering route get.
func NetworkVPCPeeringRouteGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringRouteGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringRouteGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringRouteUpdateRun is the Cobra RunE handler for VPC peering route update.
func NetworkVPCPeeringRouteUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringRouteUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringRouteUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringRouteDeleteRun is the Cobra RunE handler for VPC peering route delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkVPCPeeringRouteDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringRouteDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("VPC peering route", args.RouteID)
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
		if _, err := client.FromNetwork().VPCPeeringRoutes().Get(ctx, aruba.VPCPeeringRouteRef(args.ProjectID, args.VPCID, args.PeeringID, args.RouteID)); err != nil {
			return fmt.Errorf("dry-run: VPC peering route not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("VPC peering route", args.RouteID))
		return nil
	}

	if err := NetworkVPCPeeringRouteDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringRouteListRun is the Cobra RunE handler for VPC peering route list.
func NetworkVPCPeeringRouteListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringRouteListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringRouteList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
