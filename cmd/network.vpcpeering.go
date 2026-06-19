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
	// Peering
	networkCmd.AddCommand(vpcpeeringCmd)

	vpcpeeringCmd.AddCommand(vpcpeeringCreateCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringGetCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringUpdateCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringDeleteCmd)
	vpcpeeringCmd.AddCommand(vpcpeeringListCmd)

	// VPC Peering flags
	vpcpeeringCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringCreateCmd.Flags().String("name", "", "VPC Peering name (required)")
	vpcpeeringCreateCmd.Flags().String("peer-vpc-id", "", "Peer VPC ID or URI (required)")
	vpcpeeringCreateCmd.Flags().String("region", "", "Region code (e.g., ITBG-Bergamo) (required)")
	vpcpeeringCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcpeeringCreateCmd.MarkFlagRequired("name")
	vpcpeeringCreateCmd.MarkFlagRequired("peer-vpc-id")
	vpcpeeringCreateCmd.MarkFlagRequired("region")

	vpcpeeringGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	vpcpeeringUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringUpdateCmd.Flags().String("name", "", "New name for the VPC peering")
	vpcpeeringUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	vpcpeeringDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcpeeringDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	vpcpeeringListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcpeeringListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpcpeeringListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	vpcpeeringCreateCmd.ValidArgsFunction = completeVPCPeeringID
	vpcpeeringListCmd.ValidArgsFunction = completeVPCPeeringID
	vpcpeeringGetCmd.ValidArgsFunction = completeVPCPeeringID
	vpcpeeringUpdateCmd.ValidArgsFunction = completeVPCPeeringID
	vpcpeeringDeleteCmd.ValidArgsFunction = completeVPCPeeringID
}

func completeVPCPeeringID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return completeVPCID(cmd, args, toComplete)
	}
	if len(args) > 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	vpcID := args[0]
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("vpc-peering", projectID, vpcID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().VPCPeerings().List(ctx, aruba.VPCRef(projectID, vpcID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, p := range list.Items() {
			id := p.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, p.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// Peering subcommands
type vpcPeeringGetView struct {
	ID, Name, PeerVPC, Region, CreatedAt, CreatedBy, Tags, Status string
}

var vpcpeeringCmd = &cobra.Command{
	Use:   "vpcpeering",
	Short: "Manage VPC peering",
	Long:  `Perform CRUD operations on VPC peering in Aruba Cloud.`,
}

var vpcpeeringCreateCmd = &cobra.Command{
	Use:   "create [vpc-id]",
	Short: "Create a new VPC peering",
	Long: `Create a VPC peering connection between two VPCs.

Provide the ID of the peer VPC with --peer-vpc-id. Both VPCs must be in the
same region. Add routes via 'acloud network vpcpeeringroute create' after the
peering is established.`,
	Example: `  acloud network vpcpeering create <vpc-id> \
    --name my-peering --region IT-BG --peer-vpc-id <peer-vpc-id>`,
	Args: cobra.ExactArgs(1),
	RunE: NetworkVPCPeeringCreateRun,
}

var vpcpeeringGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [peering-id]",
	Short: "Get VPC peering details",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPCPeeringGetRun,
}

var vpcpeeringListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List VPC peerings for a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPCPeeringListRun,
}

var vpcpeeringUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [peering-id]",
	Short: "Update a VPC peering",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPCPeeringUpdateRun,
}

var vpcpeeringDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [peering-id]",
	Short: "Delete a VPC peering",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkVPCPeeringDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkVPCPeeringCreateArgs holds the typed arguments for creating a VPC peering.
type NetworkVPCPeeringCreateArgs struct {
	ProjectID string
	VPCID     string
	Name      string
	Region    aruba.Region
	PeerVPCID string
	Tags      []string
}

// NetworkVPCPeeringGetArgs holds the typed arguments for getting a VPC peering.
type NetworkVPCPeeringGetArgs struct {
	ProjectID string
	VPCID     string
	PeeringID string
}

// NetworkVPCPeeringUpdateArgs holds the typed arguments for updating a VPC peering.
type NetworkVPCPeeringUpdateArgs struct {
	ProjectID   string
	VPCID       string
	PeeringID   string
	Name        string
	Tags        []string
	TagsChanged bool
}

// NetworkVPCPeeringDeleteArgs holds the typed arguments for deleting a VPC peering.
type NetworkVPCPeeringDeleteArgs struct {
	ProjectID   string
	VPCID       string
	PeeringID   string
	DryRun      bool
	SkipConfirm bool
}

// NetworkVPCPeeringListArgs holds the typed arguments for listing VPC peerings.
type NetworkVPCPeeringListArgs struct {
	ProjectID string
	VPCID     string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkVPCPeeringCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkVPCPeeringCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringCreateArgs, error) {
	args := &NetworkVPCPeeringCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkVPCPeeringGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringGetArgs, error) {
	args := &NetworkVPCPeeringGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkVPCPeeringUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringUpdateArgs, error) {
	args := &NetworkVPCPeeringUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkVPCPeeringDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringDeleteArgs, error) {
	args := &NetworkVPCPeeringDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCPeeringListArgsFromCobraCommand parses and validates args for list.
func NewNetworkVPCPeeringListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCPeeringListArgs, error) {
	args := &NetworkVPCPeeringListArgs{}
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
func (a *NetworkVPCPeeringCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if a.PeerVPCID, err = cmd.Flags().GetString("peer-vpc-id"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkVPCPeeringGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkVPCPeeringUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
	}
	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
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
func (a *NetworkVPCPeeringDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.PeeringID = cobraArgs[1]
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
func (a *NetworkVPCPeeringListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *NetworkVPCPeeringCreateArgs) Validate() error {
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
	if a.PeerVPCID == "" {
		errs = append(errs, errors.New("--peer-vpc-id is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkVPCPeeringGetArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.PeeringID == "" {
		errs = append(errs, errors.New("peering ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *NetworkVPCPeeringUpdateArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.PeeringID == "" {
		errs = append(errs, errors.New("peering ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *NetworkVPCPeeringDeleteArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.PeeringID == "" {
		errs = append(errs, errors.New("peering ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *NetworkVPCPeeringListArgs) Validate() error {
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

// NetworkVPCPeeringCreate creates a VPC peering using the provided args and client.
func NetworkVPCPeeringCreate(ctx context.Context, client aruba.Client, args NetworkVPCPeeringCreateArgs) error {
	peering := aruba.NewVPCPeering().
		InVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		Named(args.Name).
		InRegion(args.Region).
		PeeredWith(aruba.URI(args.PeerVPCID)).
		RetaggedAs(args.Tags...)

	resp, err := client.FromNetwork().VPCPeerings().Create(ctx, peering)
	if err != nil {
		return fmt.Errorf("creating VPC peering: %w", apiErrFromV2(err))
	}
	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "PEER VPC", Width: 26},
			{Header: "REGION", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		peerVPC := ""
		if raw.Properties.RemoteVPC != nil {
			peerVPC = raw.Properties.RemoteVPC.URI
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		PrintOutput(resp, headers, [][]string{{args.Name, id, peerVPC, regionVal, status}})
	} else {
		fmt.Println(msgCreatedAsync("VPC peering", args.Name))
	}
	return nil
}

// NetworkVPCPeeringGet retrieves and displays a VPC peering's details.
func NetworkVPCPeeringGet(ctx context.Context, client aruba.Client, args NetworkVPCPeeringGetArgs) error {
	peering, err := client.FromNetwork().VPCPeerings().Get(ctx, aruba.VPCPeeringRef(args.ProjectID, args.VPCID, args.PeeringID))
	if err != nil {
		return fmt.Errorf("getting VPC peering: %w", apiErrFromV2(err))
	}
	if peering == nil || peering.Raw() == nil {
		fmt.Println("VPC peering not found or no data returned.")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(peering, nil, nil)
		return nil
	}
	raw := peering.Raw()
	view := vpcPeeringGetView{Tags: "[]"}
	if raw.Metadata.ID != nil {
		view.ID = *raw.Metadata.ID
	}
	if raw.Metadata.Name != nil {
		view.Name = *raw.Metadata.Name
	}
	if raw.Properties.RemoteVPC != nil {
		view.PeerVPC = raw.Properties.RemoteVPC.URI
	}
	if raw.Metadata.LocationResponse != nil {
		view.Region = string(raw.Metadata.LocationResponse.Value)
	}
	if raw.Metadata.CreationDate != nil {
		view.CreatedAt = raw.Metadata.CreationDate.Format(DateLayout)
	}
	if raw.Metadata.CreatedBy != nil {
		view.CreatedBy = *raw.Metadata.CreatedBy
	}
	if len(raw.Metadata.Tags) > 0 {
		view.Tags = fmt.Sprintf("%v", raw.Metadata.Tags)
	}
	if raw.Status.State != nil {
		view.Status = string(*raw.Status.State)
	}
	return renderGet(vpcPeeringGetTmpl, view)
}

// NetworkVPCPeeringList lists VPC peerings in the given VPC.
func NetworkVPCPeeringList(ctx context.Context, client aruba.Client, args NetworkVPCPeeringListArgs) error {
	list, err := client.FromNetwork().VPCPeerings().List(ctx, aruba.VPCRef(args.ProjectID, args.VPCID))
	if err != nil {
		return fmt.Errorf("listing VPC peerings: %w", apiErrFromV2(err))
	}
	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No VPC peerings found.")
		return nil
	}
	renderList(list, []ListColumn[*aruba.VPCPeering]{
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(p *aruba.VPCPeering) string {
			if r := p.Raw(); r != nil && r.Metadata.Name != nil {
				return *r.Metadata.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ID", Width: 26}, Value: func(p *aruba.VPCPeering) string {
			if r := p.Raw(); r != nil && r.Metadata.ID != nil {
				return *r.Metadata.ID
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "PEER VPC", Width: 26}, Value: func(p *aruba.VPCPeering) string {
			if r := p.Raw(); r != nil && r.Properties.RemoteVPC != nil {
				return r.Properties.RemoteVPC.URI
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "REGION", Width: 18}, Value: func(p *aruba.VPCPeering) string {
			if r := p.Raw(); r != nil && r.Metadata.LocationResponse != nil {
				return string(r.Metadata.LocationResponse.Value)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(p *aruba.VPCPeering) string {
			if r := p.Raw(); r != nil && r.Status.State != nil {
				return string(*r.Status.State)
			}
			return ""
		}},
	}, list.Items(), func(p *aruba.VPCPeering) bool { return p.Raw() != nil })
	return nil
}

// NetworkVPCPeeringUpdate updates a VPC peering's name and/or tags.
func NetworkVPCPeeringUpdate(ctx context.Context, client aruba.Client, args NetworkVPCPeeringUpdateArgs) error {
	peering, err := client.FromNetwork().VPCPeerings().Get(ctx, aruba.VPCPeeringRef(args.ProjectID, args.VPCID, args.PeeringID))
	if err != nil || peering == nil || peering.Raw() == nil {
		return fmt.Errorf("fetching current VPC peering: %w", apiErrFromV2(err))
	}

	if peering.Raw().Status.State != nil && *peering.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update VPC peering while it is in 'InCreation' state. Please wait until the VPC peering is fully created")
	}

	if args.Name != "" {
		peering.Named(args.Name)
	}
	if args.TagsChanged {
		peering.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromNetwork().VPCPeerings().Update(ctx, peering)
	if err != nil {
		return fmt.Errorf("updating VPC peering: %w", apiErrFromV2(err))
	}
	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "PEER VPC", Width: 26},
			{Header: "REGION", Width: 18},
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
		peerVPC := ""
		if raw.Properties.RemoteVPC != nil {
			peerVPC = raw.Properties.RemoteVPC.URI
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		PrintOutput(updated, headers, [][]string{{nameVal, id, peerVPC, regionVal, status}})
	} else {
		fmt.Println(msgUpdatedAsync("VPC peering", args.PeeringID))
	}
	return nil
}

// NetworkVPCPeeringDelete deletes a VPC peering.
func NetworkVPCPeeringDelete(ctx context.Context, client aruba.Client, args NetworkVPCPeeringDeleteArgs) error {
	err := client.FromNetwork().VPCPeerings().Delete(ctx, aruba.VPCPeeringRef(args.ProjectID, args.VPCID, args.PeeringID))
	if err != nil {
		return fmt.Errorf("deleting VPC peering: %w", apiErrFromV2(err))
	}
	headers := []TableColumn{
		{Header: "ID", Width: 26},
		{Header: "STATUS", Width: 15},
	}
	status := "deleted"
	result := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{args.PeeringID, status}
	PrintOutput(result, headers, [][]string{{args.PeeringID, status}})
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkVPCPeeringCreateRun is the Cobra RunE handler for VPC peering create.
func NetworkVPCPeeringCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringGetRun is the Cobra RunE handler for VPC peering get.
func NetworkVPCPeeringGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringListRun is the Cobra RunE handler for VPC peering list.
func NetworkVPCPeeringListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringUpdateRun is the Cobra RunE handler for VPC peering update.
func NetworkVPCPeeringUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCPeeringUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCPeeringDeleteRun is the Cobra RunE handler for VPC peering delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkVPCPeeringDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCPeeringDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("VPC peering", args.PeeringID)
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
		_, err = client.FromNetwork().VPCPeerings().Get(ctx, aruba.VPCPeeringRef(args.ProjectID, args.VPCID, args.PeeringID))
		if err != nil {
			return fmt.Errorf("dry-run: VPC peering not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("VPC peering", args.PeeringID))
		return nil
	}

	if err := NetworkVPCPeeringDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
