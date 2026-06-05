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
	databaseCmd.AddCommand(dbaasCmd)
	dbaasCmd.AddCommand(dbaasCreateCmd)
	dbaasCmd.AddCommand(dbaasGetCmd)
	dbaasCmd.AddCommand(dbaasUpdateCmd)
	dbaasCmd.AddCommand(dbaasDeleteCmd)
	dbaasCmd.AddCommand(dbaasListCmd)

	dbaasCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasCreateCmd.Flags().String("name", "", "Name for the DBaaS instance (required)")
	dbaasCreateCmd.Flags().String("region", "", "Region code (required)")
	dbaasCreateCmd.Flags().String("zone", "", "Availability zone (required, e.g. ITBG-1)")
	dbaasCreateCmd.Flags().String("engine-id", "", "Database engine ID (required, e.g. mysql-8.0)")
	dbaasCreateCmd.Flags().String("flavor", "", "DBaaS flavor name (required, e.g. DBO4A8)")
	dbaasCreateCmd.Flags().Int("storage-size", 0, "Storage size in GB (required)")
	dbaasCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	dbaasCreateCmd.Flags().String("vpc-id", "", "VPC ID (required when project has a VPC)")
	dbaasCreateCmd.Flags().String("subnet-id", "", "Subnet ID (required when project has a VPC)")
	dbaasCreateCmd.Flags().String("security-group-id", "", "Security group ID (required when project has a VPC)")
	dbaasCreateCmd.Flags().String("elastic-ip-id", "", "Elastic IP ID (optional)")
	dbaasCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
	dbaasCreateCmd.MarkFlagRequired("name")
	dbaasCreateCmd.MarkFlagRequired("region")
	dbaasCreateCmd.MarkFlagRequired("zone")
	dbaasCreateCmd.MarkFlagRequired("engine-id")
	dbaasCreateCmd.MarkFlagRequired("flavor")
	dbaasCreateCmd.MarkFlagRequired("storage-size")

	dbaasGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUpdateCmd.Flags().String("name", "", "New name for the DBaaS instance")
	dbaasUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	dbaasDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasListCmd.Flags().Int("offset", 0, "Number of results to skip")

	dbaasGetCmd.ValidArgsFunction = completeDBaaSID
	dbaasUpdateCmd.ValidArgsFunction = completeDBaaSID
	dbaasDeleteCmd.ValidArgsFunction = completeDBaaSID
}

// dbaasRef builds the URI for a specific DBaaS instance.
func dbaasRef(projectID, dbaasID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Database/dbaas/" + dbaasID)
}

func completeDBaaSID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromDatabase().DBaaS().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, d := range list.Items() {
			id := d.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, d.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var dbaasCmd = &cobra.Command{
	Use:   "dbaas",
	Short: "Manage DBaaS resources",
	Long:  `Perform CRUD operations on DBaaS resources in Aruba Cloud.`,
}

var dbaasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new DBaaS instance",
	Long: `Create a new managed database instance in the specified region.

Use --engine-id to select the database engine (e.g., MySQL, PostgreSQL).
Use --flavor to select the compute profile (CPU/RAM).

After creation, add databases with 'acloud database dbaas database create'
and users with 'acloud database dbaas user create'.`,
	Example: `  acloud database dbaas create \
    --name my-db --region ITBG-Bergamo \
    --engine-id <engine-id> \
    --flavor <flavor-id>`,
	Args: cobra.NoArgs,
	RunE: DatabaseDBaaSCreateRun,
}

var dbaasGetCmd = &cobra.Command{
	Use:   "get [dbaas-id]",
	Short: "Get DBaaS instance details",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSGetRun,
}

var dbaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DBaaS instances",
	Args:  cobra.NoArgs,
	RunE:  DatabaseDBaaSListRun,
}

var dbaasUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id]",
	Short: "Update a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSUpdateRun,
}

var dbaasDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id]",
	Short: "Delete a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// DatabaseDBaaSCreateArgs holds the typed arguments for creating a DBaaS instance.
type DatabaseDBaaSCreateArgs struct {
	ProjectID     string
	Name          string
	Region        aruba.Region
	Zone          string
	Engine        string
	Flavor        string
	SizeGB        int
	BillingPeriod aruba.BillingPeriod
	VPCID         string
	SubnetID      string
	SGID          string
	ElasticIPID   string
	Tags          []string
}

// DatabaseDBaaSGetArgs holds the typed arguments for getting a DBaaS instance.
type DatabaseDBaaSGetArgs struct {
	ProjectID string
	ID        string
}

// DatabaseDBaaSUpdateArgs holds the typed arguments for updating a DBaaS instance.
type DatabaseDBaaSUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
}

// DatabaseDBaaSDeleteArgs holds the typed arguments for deleting a DBaaS instance.
type DatabaseDBaaSDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// DatabaseDBaaSListArgs holds the typed arguments for listing DBaaS instances.
type DatabaseDBaaSListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewDatabaseDBaaSCreateArgsFromCobraCommand parses and validates args for create.
func NewDatabaseDBaaSCreateArgsFromCobraCommand(cmd *cobra.Command) (*DatabaseDBaaSCreateArgs, error) {
	args := &DatabaseDBaaSCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSGetArgsFromCobraCommand parses and validates args for get.
func NewDatabaseDBaaSGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSGetArgs, error) {
	args := &DatabaseDBaaSGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSUpdateArgsFromCobraCommand parses and validates args for update.
func NewDatabaseDBaaSUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSUpdateArgs, error) {
	args := &DatabaseDBaaSUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSDeleteArgsFromCobraCommand parses and validates args for delete.
func NewDatabaseDBaaSDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSDeleteArgs, error) {
	args := &DatabaseDBaaSDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSListArgsFromCobraCommand parses and validates args for list.
func NewDatabaseDBaaSListArgsFromCobraCommand(cmd *cobra.Command) (*DatabaseDBaaSListArgs, error) {
	args := &DatabaseDBaaSListArgs{}
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
func (a *DatabaseDBaaSCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if a.Zone, err = cmd.Flags().GetString("zone"); err != nil {
		errs = append(errs, err)
	}
	if a.Engine, err = cmd.Flags().GetString("engine-id"); err != nil {
		errs = append(errs, err)
	}
	if a.Flavor, err = cmd.Flags().GetString("flavor"); err != nil {
		errs = append(errs, err)
	}
	if a.SizeGB, err = cmd.Flags().GetInt("storage-size"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.VPCID, err = cmd.Flags().GetString("vpc-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SubnetID, err = cmd.Flags().GetString("subnet-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SGID, err = cmd.Flags().GetString("security-group-id"); err != nil {
		errs = append(errs, err)
	}
	if a.ElasticIPID, err = cmd.Flags().GetString("elastic-ip-id"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *DatabaseDBaaSGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *DatabaseDBaaSUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *DatabaseDBaaSDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *DatabaseDBaaSListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
func (a *DatabaseDBaaSCreateArgs) Validate() error {
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
	if a.Engine == "" {
		errs = append(errs, errors.New("--engine-id is required"))
	}
	if a.Flavor == "" {
		errs = append(errs, errors.New("--flavor is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *DatabaseDBaaSGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("DBaaS ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *DatabaseDBaaSUpdateArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *DatabaseDBaaSDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("DBaaS ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *DatabaseDBaaSListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// DatabaseDBaaSCreate creates a new DBaaS instance.
func DatabaseDBaaSCreate(ctx context.Context, client aruba.Client, args DatabaseDBaaSCreateArgs) error {
	dbaas := aruba.NewDBaaS().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		InZone(aruba.Zone(args.Zone)).
		OfEngine(aruba.DatabaseEngine(args.Engine)).
		OfFlavor(aruba.DBaaSFlavor(args.Flavor)).
		SizedGB(args.SizeGB).
		RetaggedAs(args.Tags...)

	if args.VPCID != "" {
		dbaas.WithVPC(aruba.VPCRef(args.ProjectID, args.VPCID))
	}
	if args.SubnetID != "" {
		dbaas.WithSubnet(aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID))
	}
	if args.SGID != "" {
		dbaas.WithSecurityGroup(aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID))
	}
	if args.ElasticIPID != "" {
		dbaas.WithElasticIP(aruba.ElasticIPRef(args.ProjectID, args.ElasticIPID))
	}

	created, err := client.FromDatabase().DBaaS().Create(ctx, dbaas)
	if err != nil {
		return fmt.Errorf("creating DBaaS instance: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "ENGINE", Width: 20},
			{Header: "VERSION", Width: 15},
			{Header: "FLAVOR", Width: 20},
			{Header: "REGION", Width: 20},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		engine := ""
		if raw.Properties.Engine != nil && raw.Properties.Engine.Type != nil {
			engine = *raw.Properties.Engine.Type
		}
		version := ""
		if raw.Properties.Engine != nil && raw.Properties.Engine.Version != nil {
			version = *raw.Properties.Engine.Version
		}
		flavorVal := ""
		if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
			flavorVal = *raw.Properties.Flavor.Name
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, engine, version, flavorVal, regionVal}})
	} else {
		fmt.Println(msgCreatedAsync("DBaaS instance", args.Name))
	}
	return nil
}

// DatabaseDBaaSGet retrieves and displays a DBaaS instance's details.
func DatabaseDBaaSGet(ctx context.Context, client aruba.Client, args DatabaseDBaaSGetArgs) error {
	dbaas, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting DBaaS instance: %w", apiErrFromV2(err))
	}

	if dbaas != nil && dbaas.Raw() != nil {
		raw := dbaas.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(dbaas, nil, nil)
			return nil
		}

		fmt.Println("\nDBaaS Instance Details:")
		fmt.Println("=======================")
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
			fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
		}
		if raw.Properties.Engine != nil {
			if raw.Properties.Engine.Type != nil {
				fmt.Printf("Engine Type:     %s\n", *raw.Properties.Engine.Type)
			}
			if raw.Properties.Engine.Version != nil {
				fmt.Printf("Engine Version:  %s\n", *raw.Properties.Engine.Version)
			}
			if raw.Properties.Engine.Name != nil {
				fmt.Printf("Engine Name:     %s\n", *raw.Properties.Engine.Name)
			}
		}
		if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
			fmt.Printf("Flavor:          %s\n", *raw.Properties.Flavor.Name)
		}
		if raw.Status.State != nil {
			fmt.Printf("Status:          %s\n", string(*raw.Status.State))
		}
		if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
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
		fmt.Println()
	} else {
		fmt.Println("DBaaS instance not found")
	}
	return nil
}

// DatabaseDBaaSUpdate updates a DBaaS instance's name and/or tags.
func DatabaseDBaaSUpdate(ctx context.Context, client aruba.Client, args DatabaseDBaaSUpdateArgs) error {
	dbaas, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting DBaaS instance: %w", apiErrFromV2(err))
	}
	if dbaas == nil || dbaas.Raw() == nil {
		return fmt.Errorf("DBaaS instance not found")
	}

	if args.Name != "" {
		dbaas.Named(args.Name)
	}
	if args.TagsChanged {
		dbaas.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromDatabase().DBaaS().Update(ctx, dbaas)
	if err != nil {
		return fmt.Errorf("updating DBaaS instance: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("DBaaS instance", args.ID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgUpdatedAsync("DBaaS instance", args.ID))
	}
	return nil
}

// DatabaseDBaaSDelete deletes a DBaaS instance.
func DatabaseDBaaSDelete(ctx context.Context, client aruba.Client, args DatabaseDBaaSDeleteArgs) error {
	err := client.FromDatabase().DBaaS().Delete(ctx, dbaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting DBaaS instance: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("DBaaS instance", args.ID))
	return nil
}

// DatabaseDBaaSList lists all DBaaS instances in a project.
func DatabaseDBaaSList(ctx context.Context, client aruba.Client, args DatabaseDBaaSListArgs) error {
	list, err := client.FromDatabase().DBaaS().List(ctx, aruba.URI("/projects/"+args.ProjectID))
	if err != nil {
		return fmt.Errorf("listing DBaaS instances: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 30},
			{Header: "ENGINE", Width: 20},
			{Header: "VERSION", Width: 15},
			{Header: "FLAVOR", Width: 20},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, d := range list.Items() {
			raw := d.Raw()
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
			engine := ""
			if raw.Properties.Engine != nil && raw.Properties.Engine.Type != nil {
				engine = *raw.Properties.Engine.Type
			}
			version := ""
			if raw.Properties.Engine != nil && raw.Properties.Engine.Version != nil {
				version = *raw.Properties.Engine.Version
			}
			flavor := ""
			if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
				flavor = *raw.Properties.Flavor.Name
			}
			region := ""
			if raw.Metadata.LocationResponse != nil {
				region = string(raw.Metadata.LocationResponse.Value)
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, engine, version, flavor, region, status})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No DBaaS instances found")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// DatabaseDBaaSCreateRun is the Cobra RunE handler for DBaaS create.
func DatabaseDBaaSCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewDatabaseDBaaSCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSGetRun is the Cobra RunE handler for DBaaS get.
func DatabaseDBaaSGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSUpdateRun is the Cobra RunE handler for DBaaS update.
func DatabaseDBaaSUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSDeleteRun is the Cobra RunE handler for DBaaS delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func DatabaseDBaaSDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("DBaaS instance", args.ID)
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
		if _, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: DBaaS instance not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("DBaaS instance", args.ID))
		return nil
	}

	if err := DatabaseDBaaSDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSListRun is the Cobra RunE handler for DBaaS list.
func DatabaseDBaaSListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewDatabaseDBaaSListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
