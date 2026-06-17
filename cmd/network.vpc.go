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

	// VPC
	networkCmd.AddCommand(vpcCmd)
	vpcCmd.AddCommand(vpcCreateCmd)
	vpcCmd.AddCommand(vpcGetCmd)
	vpcCmd.AddCommand(vpcUpdateCmd)
	vpcCmd.AddCommand(vpcDeleteCmd)
	vpcCmd.AddCommand(vpcListCmd)

	// Add flags for VPC commands
	vpcCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcCreateCmd.Flags().String("name", "", "Name for the VPC")
	vpcCreateCmd.Flags().String("region", "", "Region code (e.g., IT-BG)")
	vpcCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcCreateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")
	vpcCreateCmd.MarkFlagRequired("name")
	vpcCreateCmd.MarkFlagRequired("region")
	vpcGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("name", "", "New name for the VPC")
	vpcUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpcUpdateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")
	vpcDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	vpcListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpcListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	vpcGetCmd.ValidArgsFunction = completeVPCID
	vpcUpdateCmd.ValidArgsFunction = completeVPCID
	vpcDeleteCmd.ValidArgsFunction = completeVPCID
}

// Completion functions for network resources

func completeVPCID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("vpc", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().VPCs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vpc := range list.Items() {
			id := vpc.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, vpc.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// VPC subcommands
var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Manage VPCs",
	Long:  `Perform CRUD operations on VPCs in Aruba Cloud.`,
}

var vpcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VPC",
	Long: `Create a new Virtual Private Cloud (VPC) in the specified region.

A VPC is the top-level network boundary. Subnets, security groups, and other
network resources are created within a VPC.`,
	Example: `  acloud network vpc create --name my-vpc --region IT-BG
  acloud network vpc create --name prod-vpc --region IT-BG --tags env=prod,team=infra`,
	Args: cobra.NoArgs,
	RunE: NetworkVPCCreateRun,
}

var vpcGetCmd = &cobra.Command{
	Use:   "get <vpc-id>",
	Short: "Get VPC details",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPCGetRun,
}

var vpcUpdateCmd = &cobra.Command{
	Use:   "update <vpc-id>",
	Short: "Update a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPCUpdateRun,
}

var vpcDeleteCmd = &cobra.Command{
	Use:   "delete <vpc-id>",
	Short: "Delete a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPCDeleteRun,
}

var vpcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPCs",
	Args:  cobra.NoArgs,
	RunE:  NetworkVPCListRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkVPCCreateArgs holds the typed arguments for creating a VPC.
type NetworkVPCCreateArgs struct {
	ProjectID string
	Name      string
	Region    aruba.Region
	Tags      []string
	Wait      bool
}

// NetworkVPCGetArgs holds the typed arguments for getting a VPC.
type NetworkVPCGetArgs struct {
	ProjectID string
	ID        string
}

// NetworkVPCUpdateArgs holds the typed arguments for updating a VPC.
type NetworkVPCUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
	Wait        bool
}

// NetworkVPCDeleteArgs holds the typed arguments for deleting a VPC.
type NetworkVPCDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// NetworkVPCListArgs holds the typed arguments for listing VPCs.
type NetworkVPCListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkVPCCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkVPCCreateArgsFromCobraCommand(cmd *cobra.Command) (*NetworkVPCCreateArgs, error) {
	args := &NetworkVPCCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkVPCGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCGetArgs, error) {
	args := &NetworkVPCGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkVPCUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCUpdateArgs, error) {
	args := &NetworkVPCUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkVPCDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPCDeleteArgs, error) {
	args := &NetworkVPCDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPCListArgsFromCobraCommand parses and validates args for list.
func NewNetworkVPCListArgsFromCobraCommand(cmd *cobra.Command) (*NetworkVPCListArgs, error) {
	args := &NetworkVPCListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
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

// ParseFromCobraCommand reads Cobra flags into the create args struct.
func (a *NetworkVPCCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

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
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkVPCGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkVPCUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *NetworkVPCDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *NetworkVPCListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

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
func (a *NetworkVPCCreateArgs) Validate() error {
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

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkVPCGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("VPC ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *NetworkVPCUpdateArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("VPC ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *NetworkVPCDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("VPC ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *NetworkVPCListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkVPCCreate creates a VPC using the provided args and client.
func NetworkVPCCreate(ctx context.Context, client aruba.Client, args NetworkVPCCreateArgs) error {
	vpc := aruba.NewVPC().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		RetaggedAs(args.Tags...)

	created, err := client.FromNetwork().VPCs().Create(ctx, vpc)
	if err != nil {
		return fmt.Errorf("creating VPC: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		fmt.Printf("\n%s\n", msgCreated("VPC", args.Name))
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
			fmt.Printf("ID:      %s\n", id)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		fmt.Printf("Default: %t\n", raw.Properties.Default)
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
		if args.Wait && id != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(args.ProjectID, id))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "VPC", args.Name); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgCreatedAsync("VPC", args.Name))
	}
	return nil
}

// NetworkVPCGet retrieves and displays a VPC's details.
func NetworkVPCGet(ctx context.Context, client aruba.Client, args NetworkVPCGetArgs) error {
	vpc, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting VPC details: %w", apiErrFromV2(err))
	}

	if vpc != nil && vpc.Raw() != nil {
		raw := vpc.Raw()

		fmt.Println("\nVPC Details:")
		fmt.Println("============")

		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Metadata.LocationResponse != nil && raw.Metadata.LocationResponse.Value != "" {
			fmt.Printf("Region:          %s\n", raw.Metadata.LocationResponse.Value)
		}
		fmt.Printf("Default:         %t\n", raw.Properties.Default)
		fmt.Printf("Linked Resources: %d\n", len(raw.Properties.LinkedResources))

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
		fmt.Println("VPC not found or no data returned.")
	}
	return nil
}

// NetworkVPCUpdate updates a VPC's name and/or tags.
func NetworkVPCUpdate(ctx context.Context, client aruba.Client, args NetworkVPCUpdateArgs) error {
	vpc, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting VPC details: %w", apiErrFromV2(err))
	}

	if vpc == nil || vpc.Raw() == nil {
		return fmt.Errorf("VPC not found")
	}

	if vpc.Raw().Status.State != nil && *vpc.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update VPC while it is in 'InCreation' state. Please wait until the VPC is fully created")
	}

	if args.Name != "" {
		vpc.Named(args.Name)
	}
	if len(args.Tags) > 0 {
		vpc.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromNetwork().VPCs().Update(ctx, vpc)
	if err != nil {
		return fmt.Errorf("updating VPC: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("VPC", args.ID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
		if args.Wait && args.ID != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(args.ProjectID, args.ID))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "VPC", args.ID); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgUpdatedAsync("VPC", args.ID))
	}
	return nil
}

// NetworkVPCDelete deletes a VPC.
func NetworkVPCDelete(ctx context.Context, client aruba.Client, args NetworkVPCDeleteArgs) error {
	err := client.FromNetwork().VPCs().Delete(ctx, aruba.VPCRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting VPC: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("VPC", args.ID))
	return nil
}

// NetworkVPCList lists all VPCs in a project.
func NetworkVPCList(ctx context.Context, client aruba.Client, args NetworkVPCListArgs) error {
	list, err := client.FromNetwork().VPCs().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing VPCs: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No VPCs found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.VPC]{
		{TableColumn: TableColumn{Header: "NAME", Width: 40}, Value: func(v *aruba.VPC) string {
			if r := v.Raw(); r != nil && r.Metadata.Name != nil {
				return *r.Metadata.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ID", Width: 25}, Value: func(v *aruba.VPC) string {
			if r := v.Raw(); r != nil && r.Metadata.ID != nil {
				return *r.Metadata.ID
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "REGION", Width: 18}, Value: func(v *aruba.VPC) string {
			if r := v.Raw(); r != nil && r.Metadata.LocationResponse != nil {
				return string(r.Metadata.LocationResponse.Value)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "SUBNETS", Width: 10}, Value: func(v *aruba.VPC) string {
			if r := v.Raw(); r != nil {
				return fmt.Sprintf("%d", len(r.Properties.LinkedResources))
			}
			return "0"
		}},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(v *aruba.VPC) string {
			if r := v.Raw(); r != nil && r.Status.State != nil {
				return string(*r.Status.State)
			}
			return ""
		}},
	}, list.Items(), func(v *aruba.VPC) bool { return v.Raw() != nil })
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkVPCCreateRun is the Cobra RunE handler for VPC create.
func NetworkVPCCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkVPCCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCGetRun is the Cobra RunE handler for VPC get.
func NetworkVPCGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCUpdateRun is the Cobra RunE handler for VPC update.
func NetworkVPCUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCDeleteRun is the Cobra RunE handler for VPC delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkVPCDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPCDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("VPC", args.ID)
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
		if _, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: VPC not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("VPC", args.ID))
		return nil
	}

	if err := NetworkVPCDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPCListRun is the Cobra RunE handler for VPC list.
func NetworkVPCListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkVPCListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPCList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
