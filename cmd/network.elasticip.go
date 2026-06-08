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
	// ElasticIP
	networkCmd.AddCommand(elasticipCmd)

	elasticipCmd.AddCommand(elasticipCreateCmd)
	elasticipCmd.AddCommand(elasticipGetCmd)
	elasticipCmd.AddCommand(elasticipUpdateCmd)
	elasticipCmd.AddCommand(elasticipDeleteCmd)
	elasticipCmd.AddCommand(elasticipListCmd)

	// Add flags for Elastic IP commands
	elasticipCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipCreateCmd.Flags().String("name", "", "Name for the Elastic IP")
	elasticipCreateCmd.Flags().String("region", "", "Region code (e.g., IT-BG)")
	elasticipCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
	elasticipCreateCmd.MarkFlagRequired("name")
	elasticipCreateCmd.MarkFlagRequired("region")
	elasticipCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	elasticipGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipUpdateCmd.Flags().String("name", "", "New name for the Elastic IP")
	elasticipUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	elasticipDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	elasticipDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	elasticipListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	elasticipListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	elasticipGetCmd.ValidArgsFunction = completeElasticIPID
	elasticipUpdateCmd.ValidArgsFunction = completeElasticIPID
	elasticipDeleteCmd.ValidArgsFunction = completeElasticIPID
}

func completeElasticIPID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().ElasticIPs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, eip := range list.Items() {
			id := eip.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, eip.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ElasticIP subcommands
var elasticipCmd = &cobra.Command{
	Use:   "elasticip",
	Short: "Manage Elastic IPs",
	Long:  `Perform CRUD operations on Elastic IPs in Aruba Cloud.`,
}

var elasticipCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Elastic IP",
	Long: `Create a new static public IP address (Elastic IP) in the specified region.

The IP can be attached to a cloud server or VPN tunnel after creation.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud network elasticip create --name my-eip --region IT-BG
  acloud network elasticip create --name prod-eip --region IT-BG --billing-period Month`,
	Args: cobra.NoArgs,
	RunE: NetworkElasticIPCreateRun,
}

var elasticipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Elastic IPs",
	Args:  cobra.NoArgs,
	RunE:  NetworkElasticIPListRun,
}

var elasticipGetCmd = &cobra.Command{
	Use:   "get <elastic-ip-id>",
	Short: "Get Elastic IP details",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkElasticIPGetRun,
}

var elasticipUpdateCmd = &cobra.Command{
	Use:   "update <elastic-ip-id>",
	Short: "Update an Elastic IP",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkElasticIPUpdateRun,
}

var elasticipDeleteCmd = &cobra.Command{
	Use:   "delete <elastic-ip-id>",
	Short: "Delete an Elastic IP",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkElasticIPDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkElasticIPCreateArgs holds the typed arguments for creating an Elastic IP.
type NetworkElasticIPCreateArgs struct {
	ProjectID     string
	Name          string
	Region        aruba.Region
	BillingPeriod aruba.BillingPeriod
	Tags          []string
}

// NetworkElasticIPGetArgs holds the typed arguments for getting an Elastic IP.
type NetworkElasticIPGetArgs struct {
	ProjectID string
	ID        string
}

// NetworkElasticIPUpdateArgs holds the typed arguments for updating an Elastic IP.
type NetworkElasticIPUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
}

// NetworkElasticIPDeleteArgs holds the typed arguments for deleting an Elastic IP.
type NetworkElasticIPDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// NetworkElasticIPListArgs holds the typed arguments for listing Elastic IPs.
type NetworkElasticIPListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkElasticIPCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkElasticIPCreateArgsFromCobraCommand(cmd *cobra.Command) (*NetworkElasticIPCreateArgs, error) {
	args := &NetworkElasticIPCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkElasticIPGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkElasticIPGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkElasticIPGetArgs, error) {
	args := &NetworkElasticIPGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkElasticIPUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkElasticIPUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkElasticIPUpdateArgs, error) {
	args := &NetworkElasticIPUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkElasticIPDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkElasticIPDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkElasticIPDeleteArgs, error) {
	args := &NetworkElasticIPDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkElasticIPListArgsFromCobraCommand parses and validates args for list.
func NewNetworkElasticIPListArgsFromCobraCommand(cmd *cobra.Command) (*NetworkElasticIPListArgs, error) {
	args := &NetworkElasticIPListArgs{}
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
func (a *NetworkElasticIPCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkElasticIPGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *NetworkElasticIPUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *NetworkElasticIPDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *NetworkElasticIPListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
func (a *NetworkElasticIPCreateArgs) Validate() error {
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

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkElasticIPGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("Elastic IP ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *NetworkElasticIPUpdateArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("Elastic IP ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *NetworkElasticIPDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("Elastic IP ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *NetworkElasticIPListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkElasticIPCreate creates an Elastic IP using the provided args and client.
func NetworkElasticIPCreate(ctx context.Context, client aruba.Client, args NetworkElasticIPCreateArgs) error {
	eip := aruba.NewElasticIP().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		BilledBy(args.BillingPeriod).
		RetaggedAs(args.Tags...)

	created, err := client.FromNetwork().ElasticIPs().Create(ctx, eip)
	if err != nil {
		return fmt.Errorf("creating Elastic IP: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		fmt.Printf("\n%s\n", msgCreated("Elastic IP", args.Name))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		if raw.Properties.Address != nil {
			fmt.Printf("Address: %s\n", *raw.Properties.Address)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgCreatedAsync("Elastic IP", args.Name))
	}
	return nil
}

// NetworkElasticIPGet retrieves and displays an Elastic IP's details.
func NetworkElasticIPGet(ctx context.Context, client aruba.Client, args NetworkElasticIPGetArgs) error {
	eip, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
	}

	if eip != nil && eip.Raw() != nil {
		raw := eip.Raw()

		fmt.Println("\nElastic IP Details:")
		fmt.Println("===================")

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
		if raw.Properties.Address != nil {
			fmt.Printf("Address:         %s\n", *raw.Properties.Address)
		}
		if raw.Properties.BillingPlanCommon != nil && raw.Properties.BillingPlanCommon.BillingPeriod != nil {
			fmt.Printf("Billing Period:  %s\n", *raw.Properties.BillingPlanCommon.BillingPeriod)
		}
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
	}
	return nil
}

// NetworkElasticIPUpdate updates an Elastic IP's name and/or tags.
func NetworkElasticIPUpdate(ctx context.Context, client aruba.Client, args NetworkElasticIPUpdateArgs) error {
	eip, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
	}

	if eip == nil || eip.Raw() == nil {
		return fmt.Errorf("Elastic IP not found")
	}

	if eip.Raw().Status.State != nil && *eip.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update Elastic IP while it is in 'InCreation' state. Please wait until the Elastic IP is fully created")
	}

	if args.Name != "" {
		eip.Named(args.Name)
	}
	if len(args.Tags) > 0 {
		eip.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromNetwork().ElasticIPs().Update(ctx, eip)
	if err != nil {
		return fmt.Errorf("updating Elastic IP: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Elastic IP", args.ID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgUpdatedAsync("Elastic IP", args.ID))
	}
	return nil
}

// NetworkElasticIPDelete deletes an Elastic IP.
func NetworkElasticIPDelete(ctx context.Context, client aruba.Client, args NetworkElasticIPDeleteArgs) error {
	err := client.FromNetwork().ElasticIPs().Delete(ctx, aruba.ElasticIPRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting Elastic IP: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Elastic IP", args.ID))
	return nil
}

// NetworkElasticIPList lists all Elastic IPs in a project.
func NetworkElasticIPList(ctx context.Context, client aruba.Client, args NetworkElasticIPListArgs) error {
	list, err := client.FromNetwork().ElasticIPs().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing Elastic IPs: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "REGION", Width: 18},
			{Header: "ADDRESS", Width: 16},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, eip := range list.Items() {
			raw := eip.Raw()
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
			address := ""
			if raw.Properties.Address != nil {
				address = *raw.Properties.Address
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, region, address, status})
		}

		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No Elastic IPs found")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkElasticIPCreateRun is the Cobra RunE handler for Elastic IP create.
func NetworkElasticIPCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkElasticIPCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkElasticIPCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkElasticIPGetRun is the Cobra RunE handler for Elastic IP get.
func NetworkElasticIPGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkElasticIPGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkElasticIPGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkElasticIPUpdateRun is the Cobra RunE handler for Elastic IP update.
func NetworkElasticIPUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkElasticIPUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkElasticIPUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkElasticIPDeleteRun is the Cobra RunE handler for Elastic IP delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkElasticIPDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkElasticIPDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("Elastic IP", args.ID)
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
		if _, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: elastic IP not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("elastic IP", args.ID))
		return nil
	}

	if err := NetworkElasticIPDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkElasticIPListRun is the Cobra RunE handler for Elastic IP list.
func NetworkElasticIPListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkElasticIPListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkElasticIPList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
