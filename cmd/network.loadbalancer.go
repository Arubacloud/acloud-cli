package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {

	// LoadBalancer (read-only)
	networkCmd.AddCommand(loadbalancerCmd)
	loadbalancerCmd.AddCommand(loadbalancerListCmd)
	loadbalancerCmd.AddCommand(loadbalancerGetCmd)

	// Add flags for Load Balancer commands
	loadbalancerGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	loadbalancerListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	loadbalancerListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	loadbalancerListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	loadbalancerGetCmd.ValidArgsFunction = completeLoadBalancerID

}

type loadBalancerGetView struct {
	ID, URI, Name, Address, VPC, LinkedResources, CreatedAt, CreatedBy, Tags, Status string
}

func loadBalancerRef(projectID, lbID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/loadBalancers/" + lbID)
}

// Completion functions for network resources
func completeLoadBalancerID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("loadbalancer", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().LoadBalancers().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, lb := range list.Items() {
			id := lb.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, lb.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// LoadBalancer subcommands
var loadbalancerCmd = &cobra.Command{
	Use:   "loadbalancer",
	Short: "Manage Load Balancers",
	Long:  `View Load Balancers in Aruba Cloud. Load Balancers are read-only resources.`,
}

var loadbalancerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Load Balancers",
	Args:  cobra.NoArgs,
	RunE:  NetworkLoadBalancerListRun,
}

var loadbalancerGetCmd = &cobra.Command{
	Use:   "get <loadbalancer-id>",
	Short: "Get Load Balancer details",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkLoadBalancerGetRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkLoadBalancerListArgs holds the typed arguments for listing Load Balancers.
type NetworkLoadBalancerListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// NetworkLoadBalancerGetArgs holds the typed arguments for getting a Load Balancer.
type NetworkLoadBalancerGetArgs struct {
	ProjectID string
	ID        string
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkLoadBalancerListArgsFromCobraCommand parses and validates args for list.
func NewNetworkLoadBalancerListArgsFromCobraCommand(cmd *cobra.Command) (*NetworkLoadBalancerListArgs, error) {
	args := &NetworkLoadBalancerListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkLoadBalancerGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkLoadBalancerGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkLoadBalancerGetArgs, error) {
	args := &NetworkLoadBalancerGetArgs{}
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

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *NetworkLoadBalancerListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkLoadBalancerGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the list args for correctness.
func (a *NetworkLoadBalancerListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the get args for correctness.
func (a *NetworkLoadBalancerGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("Load Balancer ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkLoadBalancerList lists all Load Balancers in a project.
func NetworkLoadBalancerList(ctx context.Context, client aruba.Client, args NetworkLoadBalancerListArgs) error {
	list, err := client.FromNetwork().LoadBalancers().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing Load Balancers: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No Load Balancers found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.LoadBalancer]{
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(lb *aruba.LoadBalancer) string {
			if r := lb.Raw(); r != nil && r.Metadata.Name != nil {
				return *r.Metadata.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ID", Width: 26}, Value: func(lb *aruba.LoadBalancer) string {
			if r := lb.Raw(); r != nil && r.Metadata.ID != nil {
				return *r.Metadata.ID
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "REGION", Width: 18}, Value: func(lb *aruba.LoadBalancer) string {
			if r := lb.Raw(); r != nil && r.Metadata.LocationResponse != nil {
				return string(r.Metadata.LocationResponse.Value)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ADDRESS", Width: 16}, Value: func(lb *aruba.LoadBalancer) string {
			if r := lb.Raw(); r != nil && r.Properties.Address != nil {
				return *r.Properties.Address
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(lb *aruba.LoadBalancer) string {
			if r := lb.Raw(); r != nil && r.Status.State != nil {
				return string(*r.Status.State)
			}
			return ""
		}},
	}, list.Items(), func(lb *aruba.LoadBalancer) bool { return lb.Raw() != nil })
	return nil
}

// NetworkLoadBalancerGet retrieves and displays a Load Balancer's details.
func NetworkLoadBalancerGet(ctx context.Context, client aruba.Client, args NetworkLoadBalancerGetArgs) error {
	lb, err := client.FromNetwork().LoadBalancers().Get(ctx, aruba.LoadBalancerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting Load Balancer details: %w", apiErrFromV2(err))
	}

	if lb == nil || lb.Raw() == nil {
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(lb, nil, nil)
		return nil
	}
	raw := lb.Raw()
	view := loadBalancerGetView{
		LinkedResources: fmt.Sprintf("%d", len(raw.Properties.LinkedResources)),
		Tags:            "[]",
	}
	if raw.Metadata.ID != nil {
		view.ID = *raw.Metadata.ID
	}
	if raw.Metadata.URI != nil {
		view.URI = *raw.Metadata.URI
	}
	if raw.Metadata.Name != nil {
		view.Name = *raw.Metadata.Name
	}
	if raw.Properties.Address != nil {
		view.Address = *raw.Properties.Address
	}
	if raw.Properties.VPC != nil {
		view.VPC = raw.Properties.VPC.URI
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
	return renderGet(loadBalancerGetTmpl, view)
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkLoadBalancerListRun is the Cobra RunE handler for Load Balancer list.
func NetworkLoadBalancerListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkLoadBalancerListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkLoadBalancerList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkLoadBalancerGetRun is the Cobra RunE handler for Load Balancer get.
func NetworkLoadBalancerGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkLoadBalancerGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkLoadBalancerGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
