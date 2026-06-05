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
	containerCmd.AddCommand(containerregistryCmd)
	containerregistryCmd.AddCommand(containerregistryCreateCmd)
	containerregistryCmd.AddCommand(containerregistryGetCmd)
	containerregistryCmd.AddCommand(containerregistryUpdateCmd)
	containerregistryCmd.AddCommand(containerregistryDeleteCmd)
	containerregistryCmd.AddCommand(containerregistryListCmd)

	containerregistryCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryCreateCmd.Flags().String("name", "", "Name for the container registry (required)")
	containerregistryCreateCmd.Flags().String("region", "", "Region code (required)")
	containerregistryCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	containerregistryCreateCmd.Flags().String("public-ip-id", "", "Public IP (Elastic IP) ID (required)")
	containerregistryCreateCmd.Flags().String("vpc-id", "", "VPC ID (required)")
	containerregistryCreateCmd.Flags().String("subnet-id", "", "Subnet ID (required)")
	containerregistryCreateCmd.Flags().String("security-group-id", "", "Security group ID (required)")
	containerregistryCreateCmd.Flags().String("block-storage-id", "", "Block storage ID (required)")
	containerregistryCreateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year (optional)")
	containerregistryCreateCmd.Flags().String("admin-username", "", "Administrator username (optional)")
	containerregistryCreateCmd.Flags().String("concurrent-users", "", "Concurrent users tier: Small, Medium, HighPerf (optional)")
	containerregistryCreateCmd.MarkFlagRequired("name")
	containerregistryCreateCmd.MarkFlagRequired("region")
	containerregistryCreateCmd.MarkFlagRequired("public-ip-id")
	containerregistryCreateCmd.MarkFlagRequired("vpc-id")
	containerregistryCreateCmd.MarkFlagRequired("subnet-id")
	containerregistryCreateCmd.MarkFlagRequired("security-group-id")
	containerregistryCreateCmd.MarkFlagRequired("block-storage-id")

	containerregistryGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	containerregistryUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryUpdateCmd.Flags().String("name", "", "New name for the container registry")
	containerregistryUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	containerregistryUpdateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")
	containerregistryUpdateCmd.Flags().String("concurrent-users", "", "Concurrent users tier: Small, Medium, HighPerf")

	containerregistryDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	containerregistryDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	containerregistryListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	containerregistryListCmd.Flags().Int("offset", 0, "Number of results to skip")

	containerregistryGetCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryUpdateCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryDeleteCmd.ValidArgsFunction = completeContainerRegistryID
}

// containerRegistryRef builds the URI for a specific container registry.
func containerRegistryRef(projectID, registryID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Container/registries/" + registryID)
}

func completeContainerRegistryID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromContainer().ContainerRegistry().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, cr := range list.Items() {
			id := cr.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, cr.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var containerregistryCmd = &cobra.Command{
	Use:   "containerregistry",
	Short: "Manage Container Registry",
	Long:  `Perform CRUD operations on Container Registry resources in Aruba Cloud.`,
}

var containerregistryCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new container registry",
	Long: `Create a new private container registry in the specified region.

All network resources (VPC, subnet, security group, public IP, block storage)
must already exist. Pass their IDs via the corresponding flags.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud container containerregistry create \
    --name my-registry --region IT-BG \
    --vpc-id <vpc-id> \
    --subnet-id <subnet-id> \
    --security-group-id <sg-id> \
    --public-ip-id <eip-id> \
    --block-storage-id <vol-id>`,
	Args: cobra.NoArgs,
	RunE: ContainerContainerRegistryCreateRun,
}

var containerregistryGetCmd = &cobra.Command{
	Use:   "get [containerregistry-id]",
	Short: "Get container registry details",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerContainerRegistryGetRun,
}

var containerregistryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all container registries",
	Args:  cobra.NoArgs,
	RunE:  ContainerContainerRegistryListRun,
}

var containerregistryUpdateCmd = &cobra.Command{
	Use:   "update [containerregistry-id]",
	Short: "Update a container registry",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerContainerRegistryUpdateRun,
}

var containerregistryDeleteCmd = &cobra.Command{
	Use:   "delete [containerregistry-id]",
	Short: "Delete a container registry",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerContainerRegistryDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// ContainerContainerRegistryCreateArgs holds the typed arguments for creating a container registry.
type ContainerContainerRegistryCreateArgs struct {
	ProjectID      string
	Name           string
	Region         aruba.Region
	VPCID          string
	SubnetID       string
	SGID           string
	PublicIPID     string
	BlockStorageID string
	BillingPeriod  aruba.BillingPeriod
	AdminUsername  string
	SizeFlavor     aruba.ContainerRegistrySizeFlavor
	Tags           []string
}

// ContainerContainerRegistryGetArgs holds the typed arguments for getting a container registry.
type ContainerContainerRegistryGetArgs struct {
	ProjectID string
	ID        string
}

// ContainerContainerRegistryListArgs holds the typed arguments for listing container registries.
type ContainerContainerRegistryListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// ContainerContainerRegistryUpdateArgs holds the typed arguments for updating a container registry.
type ContainerContainerRegistryUpdateArgs struct {
	ProjectID     string
	ID            string
	Name          string
	Tags          []string
	TagsChanged   bool
	BillingPeriod aruba.BillingPeriod
	SizeFlavor    aruba.ContainerRegistrySizeFlavor
}

// ContainerContainerRegistryDeleteArgs holds the typed arguments for deleting a container registry.
type ContainerContainerRegistryDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// =============================================================================
// Constructors
// =============================================================================

// NewContainerContainerRegistryCreateArgsFromCobraCommand parses and validates args for create.
func NewContainerContainerRegistryCreateArgsFromCobraCommand(cmd *cobra.Command) (*ContainerContainerRegistryCreateArgs, error) {
	args := &ContainerContainerRegistryCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerContainerRegistryGetArgsFromCobraCommand parses and validates args for get.
func NewContainerContainerRegistryGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerContainerRegistryGetArgs, error) {
	args := &ContainerContainerRegistryGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerContainerRegistryListArgsFromCobraCommand parses and validates args for list.
func NewContainerContainerRegistryListArgsFromCobraCommand(cmd *cobra.Command) (*ContainerContainerRegistryListArgs, error) {
	args := &ContainerContainerRegistryListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerContainerRegistryUpdateArgsFromCobraCommand parses and validates args for update.
func NewContainerContainerRegistryUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerContainerRegistryUpdateArgs, error) {
	args := &ContainerContainerRegistryUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerContainerRegistryDeleteArgsFromCobraCommand parses and validates args for delete.
func NewContainerContainerRegistryDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerContainerRegistryDeleteArgs, error) {
	args := &ContainerContainerRegistryDeleteArgs{}
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

// ParseFromCobraCommand reads Cobra flags into the create args struct.
func (a *ContainerContainerRegistryCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if a.VPCID, err = cmd.Flags().GetString("vpc-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SubnetID, err = cmd.Flags().GetString("subnet-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SGID, err = cmd.Flags().GetString("security-group-id"); err != nil {
		errs = append(errs, err)
	}
	if a.PublicIPID, err = cmd.Flags().GetString("public-ip-id"); err != nil {
		errs = append(errs, err)
	}
	if a.BlockStorageID, err = cmd.Flags().GetString("block-storage-id"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.AdminUsername, err = cmd.Flags().GetString("admin-username"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("concurrent-users"); err == nil {
		a.SizeFlavor = aruba.ContainerRegistrySizeFlavor(s)
	} else {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *ContainerContainerRegistryGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *ContainerContainerRegistryListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *ContainerContainerRegistryUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("concurrent-users"); err == nil {
		a.SizeFlavor = aruba.ContainerRegistrySizeFlavor(s)
	} else {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *ContainerContainerRegistryDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *ContainerContainerRegistryCreateArgs) Validate() error {
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
	// BillingPeriod is optional; validate only when provided.
	if a.BillingPeriod != "" && !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	// SizeFlavor is optional; validate only when provided.
	if a.SizeFlavor != "" && !slices.Contains(validContainerRegistrySizeFlavors, a.SizeFlavor) {
		errs = append(errs, fmt.Errorf("--concurrent-users %q: must be one of %v", a.SizeFlavor, validContainerRegistrySizeFlavors))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *ContainerContainerRegistryGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("container registry ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *ContainerContainerRegistryListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *ContainerContainerRegistryUpdateArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("container registry ID is required"))
	}
	if a.Name == "" && !a.TagsChanged && a.BillingPeriod == "" && a.SizeFlavor == "" {
		errs = append(errs, errors.New("at least one of --name, --tags, --billing-period, or --concurrent-users must be provided"))
	}
	// BillingPeriod is optional; validate only when provided.
	if a.BillingPeriod != "" && !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	// SizeFlavor is optional; validate only when provided.
	if a.SizeFlavor != "" && !slices.Contains(validContainerRegistrySizeFlavors, a.SizeFlavor) {
		errs = append(errs, fmt.Errorf("--concurrent-users %q: must be one of %v", a.SizeFlavor, validContainerRegistrySizeFlavors))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *ContainerContainerRegistryDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("container registry ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// ContainerContainerRegistryCreate creates a new container registry.
func ContainerContainerRegistryCreate(ctx context.Context, client aruba.Client, args ContainerContainerRegistryCreateArgs) error {
	registry := aruba.NewContainerRegistry().
		InProject(projectRef(args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		WithElasticIP(aruba.ElasticIPRef(args.ProjectID, args.PublicIPID)).
		WithVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		WithSubnet(aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID)).
		WithSecurityGroup(aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID)).
		WithBlockStorage(volumeRef(args.ProjectID, args.BlockStorageID)).
		RetaggedAs(args.Tags...)

	if args.AdminUsername != "" {
		registry.WithAdminUsername(args.AdminUsername)
	}
	if args.BillingPeriod != "" {
		registry.BilledBy(args.BillingPeriod)
	}
	if args.SizeFlavor != "" {
		registry.OfSize(args.SizeFlavor)
	}

	created, err := client.FromContainer().ContainerRegistry().Create(ctx, registry)
	if err != nil {
		return fmt.Errorf("creating container registry: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		statusVal := ""
		if raw.Status.State != nil {
			statusVal = string(*raw.Status.State)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, regionVal, statusVal}})
	} else {
		fmt.Println(msgCreatedAsync("Container registry", args.Name))
	}
	return nil
}

// ContainerContainerRegistryGet retrieves and displays a container registry's details.
func ContainerContainerRegistryGet(ctx context.Context, client aruba.Client, args ContainerContainerRegistryGetArgs) error {
	registry, err := client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting container registry: %w", apiErrFromV2(err))
	}

	if registry != nil && registry.Raw() != nil {
		raw := registry.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(registry, nil, nil)
			return nil
		}

		fmt.Println("\nContainer Registry Details:")
		fmt.Println("==========================")
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
		if raw.Properties.PublicIp.URI != "" {
			fmt.Printf("Public IP:       %s\n", raw.Properties.PublicIp.URI)
		}
		if raw.Properties.VPC.URI != "" {
			fmt.Printf("VPC:             %s\n", raw.Properties.VPC.URI)
		}
		if raw.Properties.Subnet.URI != "" {
			fmt.Printf("Subnet:          %s\n", raw.Properties.Subnet.URI)
		}
		if raw.Properties.SecurityGroup.URI != "" {
			fmt.Printf("Security Group:  %s\n", raw.Properties.SecurityGroup.URI)
		}
		if raw.Properties.BlockStorage.URI != "" {
			fmt.Printf("Block Storage:   %s\n", raw.Properties.BlockStorage.URI)
		}
		if raw.Properties.BillingPlanCommon != nil && raw.Properties.BillingPlanCommon.BillingPeriod != nil {
			fmt.Printf("Billing Period:  %s\n", string(*raw.Properties.BillingPlanCommon.BillingPeriod))
		}
		if raw.Properties.AdminUser != nil {
			fmt.Printf("Admin User:      %s\n", raw.Properties.AdminUser.Username)
		}
		if raw.Properties.ConcurrentUsers != nil {
			fmt.Printf("Concurrent Users: %s\n", *raw.Properties.ConcurrentUsers)
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
		fmt.Println("Container registry not found")
	}
	return nil
}

// ContainerContainerRegistryList lists all container registries in a project.
func ContainerContainerRegistryList(ctx context.Context, client aruba.Client, args ContainerContainerRegistryListArgs) error {
	list, err := client.FromContainer().ContainerRegistry().List(ctx, aruba.URI("/projects/"+args.ProjectID))
	if err != nil {
		return fmt.Errorf("listing container registries: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 40},
			{Header: "ID", Width: 30},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, cr := range list.Items() {
			raw := cr.Raw()
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
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, region, status})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No container registries found")
	}
	return nil
}

// ContainerContainerRegistryUpdate updates a container registry's mutable fields.
func ContainerContainerRegistryUpdate(ctx context.Context, client aruba.Client, args ContainerContainerRegistryUpdateArgs) error {
	registry, err := client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("fetching current container registry: %w", apiErrFromV2(err))
	}
	if registry == nil || registry.Raw() == nil {
		return fmt.Errorf("container registry not found")
	}

	if args.Name != "" {
		registry.Named(args.Name)
	}
	if args.TagsChanged {
		registry.RetaggedAs(args.Tags...)
	}
	if args.BillingPeriod != "" {
		registry.BilledBy(args.BillingPeriod)
	}
	if args.SizeFlavor != "" {
		registry.OfSize(args.SizeFlavor)
	}

	updated, err := client.FromContainer().ContainerRegistry().Update(ctx, registry)
	if err != nil {
		return fmt.Errorf("updating container registry: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Container registry", args.ID))
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
		if raw.Status.State != nil {
			fmt.Printf("Status:          %s\n", string(*raw.Status.State))
		}
	} else {
		fmt.Println(msgUpdatedAsync("Container registry", args.ID))
	}
	return nil
}

// ContainerContainerRegistryDelete deletes a container registry.
func ContainerContainerRegistryDelete(ctx context.Context, client aruba.Client, args ContainerContainerRegistryDeleteArgs) error {
	err := client.FromContainer().ContainerRegistry().Delete(ctx, containerRegistryRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting container registry: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Container registry", args.ID))
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// ContainerContainerRegistryCreateRun is the Cobra RunE handler for container registry create.
func ContainerContainerRegistryCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewContainerContainerRegistryCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerContainerRegistryCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerContainerRegistryGetRun is the Cobra RunE handler for container registry get.
func ContainerContainerRegistryGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerContainerRegistryGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerContainerRegistryGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerContainerRegistryListRun is the Cobra RunE handler for container registry list.
func ContainerContainerRegistryListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewContainerContainerRegistryListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerContainerRegistryList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerContainerRegistryUpdateRun is the Cobra RunE handler for container registry update.
func ContainerContainerRegistryUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerContainerRegistryUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerContainerRegistryUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerContainerRegistryDeleteRun is the Cobra RunE handler for container registry delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func ContainerContainerRegistryDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerContainerRegistryDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("container registry", args.ID)
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
		if _, err := client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: container registry not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("container registry", args.ID))
		return nil
	}

	if err := ContainerContainerRegistryDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
