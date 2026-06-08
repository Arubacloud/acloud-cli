package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// securityGroupRef is shared with securityrule.go; uses hyphenated path segment (/security-groups/) to match SDK ID parser.
func securityGroupRef(projectID, vpcID, sgID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/vpcs/" + vpcID + "/security-groups/" + sgID)
}

func init() {
	// SecurityGroup
	networkCmd.AddCommand(securitygroupCmd)
	securitygroupCmd.AddCommand(securitygroupCreateCmd)
	securitygroupCmd.AddCommand(securitygroupGetCmd)
	securitygroupCmd.AddCommand(securitygroupUpdateCmd)
	securitygroupCmd.AddCommand(securitygroupDeleteCmd)
	securitygroupCmd.AddCommand(securitygroupListCmd)

	// SecurityGroup flags
	securitygroupCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupCreateCmd.Flags().String("name", "", "Security group name (required)")
	securitygroupCreateCmd.Flags().String("region", "", "Region code (required)")
	securitygroupCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	securitygroupCreateCmd.MarkFlagRequired("name")
	securitygroupCreateCmd.MarkFlagRequired("region")
	securitygroupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupUpdateCmd.Flags().String("name", "", "New name for the security group")
	securitygroupUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	securitygroupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	securitygroupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	securitygroupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	securitygroupListCmd.Flags().Int32("offset", 0, "Number of results to skip")
}

// SecurityGroup subcommands

var securitygroupCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "Manage security groups",
	Long:  `Perform CRUD operations on security groups in Aruba Cloud.`,
}

var securitygroupCreateCmd = &cobra.Command{
	Use:   "create [vpc-id]",
	Short: "Create a new security group",
	Long: `Create a new security group inside the specified VPC.

Security groups act as virtual firewalls. Add rules with 'acloud network securityrule create'
after the group is created.`,
	Example: `  acloud network securitygroup create <vpc-id> --name web-sg --region IT-BG
  acloud network securitygroup create <vpc-id> --name db-sg --region IT-BG --tags tier=db`,
	Args: cobra.ExactArgs(1),
	RunE: NetworkSecurityGroupCreateRun,
}

var securitygroupGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [securitygroup-id]",
	Short: "Get security group details",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSecurityGroupGetRun,
}

var securitygroupListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List security groups for a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkSecurityGroupListRun,
}

var securitygroupUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [securitygroup-id]",
	Short: "Update a security group",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSecurityGroupUpdateRun,
}

var securitygroupDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [securitygroup-id]",
	Short: "Delete a security group",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSecurityGroupDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkSecurityGroupCreateArgs holds the typed arguments for creating a security group.
type NetworkSecurityGroupCreateArgs struct {
	ProjectID string
	VPCID     string
	Name      string
	Tags      []string
}

// NetworkSecurityGroupGetArgs holds the typed arguments for getting a security group.
type NetworkSecurityGroupGetArgs struct {
	ProjectID string
	VPCID     string
	SGID      string
}

// NetworkSecurityGroupUpdateArgs holds the typed arguments for updating a security group.
type NetworkSecurityGroupUpdateArgs struct {
	ProjectID   string
	VPCID       string
	SGID        string
	Name        string
	Tags        []string
	TagsChanged bool
}

// NetworkSecurityGroupDeleteArgs holds the typed arguments for deleting a security group.
type NetworkSecurityGroupDeleteArgs struct {
	ProjectID   string
	VPCID       string
	SGID        string
	DryRun      bool
	SkipConfirm bool
}

// NetworkSecurityGroupListArgs holds the typed arguments for listing security groups.
type NetworkSecurityGroupListArgs struct {
	ProjectID string
	VPCID     string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkSecurityGroupCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkSecurityGroupCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityGroupCreateArgs, error) {
	args := &NetworkSecurityGroupCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityGroupGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkSecurityGroupGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityGroupGetArgs, error) {
	args := &NetworkSecurityGroupGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityGroupUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkSecurityGroupUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityGroupUpdateArgs, error) {
	args := &NetworkSecurityGroupUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityGroupDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkSecurityGroupDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityGroupDeleteArgs, error) {
	args := &NetworkSecurityGroupDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityGroupListArgsFromCobraCommand parses and validates args for list.
func NewNetworkSecurityGroupListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityGroupListArgs, error) {
	args := &NetworkSecurityGroupListArgs{}
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
func (a *NetworkSecurityGroupCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkSecurityGroupGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.SGID = cobraArgs[1]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkSecurityGroupUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.SGID = cobraArgs[1]
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
func (a *NetworkSecurityGroupDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.SGID = cobraArgs[1]
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
func (a *NetworkSecurityGroupListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.VPCID = cobraArgs[0]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks that create args are valid.
func (a *NetworkSecurityGroupCreateArgs) Validate() error {
	var errs []error
	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}
	if a.VPCID == "" {
		errs = append(errs, errors.New("vpc-id argument is required"))
	}
	return errors.Join(errs...)
}

// Validate checks that get args are valid.
func (a *NetworkSecurityGroupGetArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("vpc-id argument is required"))
	}
	if a.SGID == "" {
		errs = append(errs, errors.New("securitygroup-id argument is required"))
	}
	return errors.Join(errs...)
}

// Validate checks that update args are valid.
func (a *NetworkSecurityGroupUpdateArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("vpc-id argument is required"))
	}
	if a.SGID == "" {
		errs = append(errs, errors.New("securitygroup-id argument is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks that delete args are valid.
func (a *NetworkSecurityGroupDeleteArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("vpc-id argument is required"))
	}
	if a.SGID == "" {
		errs = append(errs, errors.New("securitygroup-id argument is required"))
	}
	return errors.Join(errs...)
}

// Validate checks that list args are valid.
func (a *NetworkSecurityGroupListArgs) Validate() error {
	var errs []error
	if a.VPCID == "" {
		errs = append(errs, errors.New("vpc-id argument is required"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkSecurityGroupCreate creates a new security group.
func NetworkSecurityGroupCreate(ctx context.Context, client aruba.Client, args NetworkSecurityGroupCreateArgs) error {
	sg := aruba.NewSecurityGroup().
		InVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		Named(args.Name).
		RetaggedAs(args.Tags...)

	resp, err := client.FromNetwork().SecurityGroups().Create(ctx, sg)
	if err != nil {
		return fmt.Errorf("creating security group: %w", apiErrFromV2(err))
	}

	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "REGION", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		region := ""
		if raw.Metadata.LocationResponse != nil {
			region = string(raw.Metadata.LocationResponse.Value)
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		PrintOutput(resp, headers, [][]string{{args.Name, id, region, status}})
	} else {
		fmt.Println(msgCreatedAsync("Security group", args.Name))
	}
	return nil
}

// NetworkSecurityGroupGet retrieves a security group by ID.
func NetworkSecurityGroupGet(ctx context.Context, client aruba.Client, args NetworkSecurityGroupGetArgs) error {
	sg, err := client.FromNetwork().SecurityGroups().Get(ctx, aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID))
	if err != nil {
		return fmt.Errorf("getting security group: %w", apiErrFromV2(err))
	}

	if sg != nil && sg.Raw() != nil {
		raw := sg.Raw()
		fmt.Println("\nSecurity Group Details:")
		fmt.Println("======================")
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
		fmt.Println("Security group not found or no data returned.")
	}
	return nil
}

// NetworkSecurityGroupUpdate updates an existing security group.
func NetworkSecurityGroupUpdate(ctx context.Context, client aruba.Client, args NetworkSecurityGroupUpdateArgs) error {
	sg, err := client.FromNetwork().SecurityGroups().Get(ctx, aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID))
	if err != nil || sg == nil || sg.Raw() == nil {
		return fmt.Errorf("fetching current security group: %w", apiErrFromV2(err))
	}

	if sg.Raw().Status.State != nil && *sg.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update security group while it is in 'InCreation' state. Please wait until the security group is fully created")
	}

	if args.Name != "" {
		sg.Named(args.Name)
	}
	if args.TagsChanged {
		sg.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromNetwork().SecurityGroups().Update(ctx, sg)
	if err != nil {
		return fmt.Errorf("updating security group: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "REGION", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		updateRegion := ""
		if raw.Metadata.LocationResponse != nil {
			updateRegion = string(raw.Metadata.LocationResponse.Value)
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		PrintOutput(updated, headers, [][]string{{nameVal, id, updateRegion, status}})
	} else {
		fmt.Println(msgUpdatedAsync("Security group", args.SGID))
	}
	return nil
}

// NetworkSecurityGroupDelete deletes a security group.
func NetworkSecurityGroupDelete(ctx context.Context, client aruba.Client, args NetworkSecurityGroupDeleteArgs) error {
	err := client.FromNetwork().SecurityGroups().Delete(ctx, aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID))
	if err != nil {
		return fmt.Errorf("deleting security group: %w", apiErrFromV2(err))
	}

	headers := []TableColumn{
		{Header: "ID", Width: 26},
		{Header: "STATUS", Width: 15},
	}
	status := "deleted"
	result := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{args.SGID, status}
	PrintOutput(result, headers, [][]string{{args.SGID, status}})
	return nil
}

// NetworkSecurityGroupList lists security groups for a VPC.
func NetworkSecurityGroupList(ctx context.Context, client aruba.Client, args NetworkSecurityGroupListArgs) error {
	list, err := client.FromNetwork().SecurityGroups().List(ctx, aruba.VPCRef(args.ProjectID, args.VPCID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing security groups: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "REGION", Width: 18},
			{Header: "STATUS", Width: 15},
		}
		var rows [][]string
		for _, sg := range list.Items() {
			raw := sg.Raw()
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
		fmt.Println("No security groups found.")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkSecurityGroupCreateRun is the RunE wiring for security group create.
func NetworkSecurityGroupCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityGroupCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityGroupCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityGroupGetRun is the RunE wiring for security group get.
func NetworkSecurityGroupGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityGroupGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityGroupGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityGroupUpdateRun is the RunE wiring for security group update.
func NetworkSecurityGroupUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityGroupUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityGroupUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityGroupDeleteRun is the RunE wiring for security group delete.
func NetworkSecurityGroupDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityGroupDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("security group", args.SGID)
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
		_, err = client.FromNetwork().SecurityGroups().Get(ctx, aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID))
		if err != nil {
			return fmt.Errorf("dry-run: security group not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("security group", args.SGID))
		return nil
	}

	if err := NetworkSecurityGroupDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityGroupListRun is the RunE wiring for security group list.
func NetworkSecurityGroupListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityGroupListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityGroupList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
