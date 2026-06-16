package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {

	// SecurityRule
	networkCmd.AddCommand(securityruleCmd)
	securityruleCmd.AddCommand(securityruleCreateCmd)
	securityruleCmd.AddCommand(securityruleGetCmd)
	securityruleCmd.AddCommand(securityruleUpdateCmd)
	securityruleCmd.AddCommand(securityruleDeleteCmd)
	securityruleCmd.AddCommand(securityruleListCmd)

	// SecurityRule flags
	securityruleCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleCreateCmd.Flags().String("name", "", "Security Rule Name (required)")
	securityruleCreateCmd.Flags().String("region", "", "Region code (required)")
	securityruleCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	securityruleCreateCmd.Flags().String("direction", "", "Direction: Ingress or Egress (required)")
	securityruleCreateCmd.Flags().String("protocol", "", "Protocol: ANY, TCP, UDP, ICMP (required)")
	securityruleCreateCmd.Flags().String("port", "", "Port: a single numeric port, a port range or * (required for TCP/UDP)")
	securityruleCreateCmd.Flags().String("target-kind", "", "Target Kind: Ip or SecurityGroup (required)")
	securityruleCreateCmd.Flags().String("target-value", "", "Target Value: If kind = Ip, the value must be a valid network address in CIDR notation (included 0.0.0.0/0). If kind = SecurityGroup, the value must be a valid URI of any security group within the same VPC (required)")
	securityruleCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	securityruleCreateCmd.MarkFlagRequired("name")
	securityruleCreateCmd.MarkFlagRequired("region")
	securityruleCreateCmd.MarkFlagRequired("direction")
	securityruleCreateCmd.MarkFlagRequired("protocol")
	securityruleCreateCmd.MarkFlagRequired("target-kind")
	securityruleCreateCmd.MarkFlagRequired("target-value")

	securityruleGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	securityruleUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleUpdateCmd.Flags().String("name", "", "New name for the security rule")
	securityruleUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	securityruleDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	securityruleDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	securityruleListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	securityruleListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	securityruleGetCmd.ValidArgsFunction = completeSecurityRuleID
	securityruleUpdateCmd.ValidArgsFunction = completeSecurityRuleID
	securityruleDeleteCmd.ValidArgsFunction = completeSecurityRuleID
}

func completeSecurityRuleID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	vpcID := args[0]
	securityGroupID := args[1]

	key := cacheKey("security-rule", projectID, vpcID, securityGroupID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().SecurityGroupRules().List(ctx, aruba.SecurityGroupRef(projectID, vpcID, securityGroupID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, rule := range list.Items() {
			id := rule.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, rule.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// SecurityRule subcommands
var securityruleCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Manage security rules",
	Long:  `Perform CRUD operations on security rules in Aruba Cloud.`,
}

var securityruleCreateCmd = &cobra.Command{
	Use:   "create [vpc-id] [securitygroup-id]",
	Short: "Create a new security rule",
	Long: `Create a new security rule in the specified security group.

Valid values:
  --direction   Ingress (inbound) or Egress (outbound)
  --protocol    ANY, TCP, UDP, or ICMP
  --target-kind Ip (CIDR block) or SecurityGroup (another security group ID)
  --port        Port number or range (e.g., "80", "8000-8080"); omit for ANY/ICMP

Example target values:
  --target-kind Ip --target-value 10.0.0.0/8
  --target-kind SecurityGroup --target-value <sg-id>`,
	Example: `  # Allow inbound HTTP from any IP
  acloud network securityrule create <vpc-id> <sg-id> \
    --name allow-http --region IT-BG \
    --direction Ingress --protocol TCP --port 80 \
    --target-kind Ip --target-value 0.0.0.0/0

  # Allow all outbound traffic
  acloud network securityrule create <vpc-id> <sg-id> \
    --name allow-all-out --region IT-BG \
    --direction Egress --protocol ANY \
    --target-kind Ip --target-value 0.0.0.0/0`,
	Args: cobra.ExactArgs(2),
	RunE: NetworkSecurityRuleCreateRun,
}

var securityruleGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Get security rule details",
	Args:  cobra.ExactArgs(3),
	RunE:  NetworkSecurityRuleGetRun,
}

var securityruleListCmd = &cobra.Command{
	Use:   "list [vpc-id] [securitygroup-id]",
	Short: "List security rules for a security group",
	Args:  cobra.ExactArgs(2),
	RunE:  NetworkSecurityRuleListRun,
}

var securityruleUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Update a security rule",
	Args:  cobra.ExactArgs(3),
	RunE:  NetworkSecurityRuleUpdateRun,
}

var securityruleDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Delete a security rule",
	Args:  cobra.ExactArgs(3),
	RunE:  NetworkSecurityRuleDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkSecurityRuleCreateArgs holds the typed arguments for creating a security rule.
type NetworkSecurityRuleCreateArgs struct {
	ProjectID   string
	VPCID       string
	SGID        string
	Name        string
	Region      aruba.Region
	Tags        []string
	Direction   string
	Protocol    string
	Port        string
	TargetKind  string
	TargetValue string
	Verbose     bool
}

// NetworkSecurityRuleGetArgs holds the typed arguments for getting a security rule.
type NetworkSecurityRuleGetArgs struct {
	ProjectID string
	VPCID     string
	SGID      string
	RuleID    string
}

// NetworkSecurityRuleUpdateArgs holds the typed arguments for updating a security rule.
type NetworkSecurityRuleUpdateArgs struct {
	ProjectID   string
	VPCID       string
	SGID        string
	RuleID      string
	Name        string
	Tags        []string
	TagsChanged bool
	Debug       bool
}

// NetworkSecurityRuleDeleteArgs holds the typed arguments for deleting a security rule.
type NetworkSecurityRuleDeleteArgs struct {
	ProjectID   string
	VPCID       string
	SGID        string
	RuleID      string
	DryRun      bool
	SkipConfirm bool
}

// NetworkSecurityRuleListArgs holds the typed arguments for listing security rules.
type NetworkSecurityRuleListArgs struct {
	ProjectID string
	VPCID     string
	SGID      string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkSecurityRuleCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkSecurityRuleCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityRuleCreateArgs, error) {
	args := &NetworkSecurityRuleCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityRuleGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkSecurityRuleGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityRuleGetArgs, error) {
	args := &NetworkSecurityRuleGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityRuleUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkSecurityRuleUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityRuleUpdateArgs, error) {
	args := &NetworkSecurityRuleUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityRuleDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkSecurityRuleDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityRuleDeleteArgs, error) {
	args := &NetworkSecurityRuleDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkSecurityRuleListArgsFromCobraCommand parses and validates args for list.
func NewNetworkSecurityRuleListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkSecurityRuleListArgs, error) {
	args := &NetworkSecurityRuleListArgs{}
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
func (a *NetworkSecurityRuleCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.Direction, err = cmd.Flags().GetString("direction"); err != nil {
		errs = append(errs, err)
	}
	if a.Protocol, err = cmd.Flags().GetString("protocol"); err != nil {
		errs = append(errs, err)
	}
	if a.Port, err = cmd.Flags().GetString("port"); err != nil {
		errs = append(errs, err)
	}
	if a.TargetKind, err = cmd.Flags().GetString("target-kind"); err != nil {
		errs = append(errs, err)
	}
	if a.TargetValue, err = cmd.Flags().GetString("target-value"); err != nil {
		errs = append(errs, err)
	}
	if a.Verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkSecurityRuleGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if len(cobraArgs) > 2 {
		a.RuleID = cobraArgs[2]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkSecurityRuleUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if len(cobraArgs) > 2 {
		a.RuleID = cobraArgs[2]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")
	if debugEnabled, err := rootCmd.PersistentFlags().GetBool("debug"); err == nil {
		a.Debug = debugEnabled
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *NetworkSecurityRuleDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if len(cobraArgs) > 2 {
		a.RuleID = cobraArgs[2]
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
func (a *NetworkSecurityRuleListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks that create args are valid.
func (a *NetworkSecurityRuleCreateArgs) Validate() error {
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
	if a.Direction == "" {
		errs = append(errs, errors.New("--direction is required"))
	}
	if a.Protocol == "" {
		errs = append(errs, errors.New("--protocol is required"))
	}
	if a.TargetKind == "" {
		errs = append(errs, errors.New("--target-kind is required"))
	}
	if a.TargetValue == "" {
		errs = append(errs, errors.New("--target-value is required"))
	}
	return errors.Join(errs...)
}

// Validate checks that get args are valid.
func (a *NetworkSecurityRuleGetArgs) Validate() error {
	return nil
}

// Validate checks that update args are valid.
func (a *NetworkSecurityRuleUpdateArgs) Validate() error {
	var errs []error
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one field (--name or --tags) must be provided for update"))
	}
	return errors.Join(errs...)
}

// Validate checks that delete args are valid.
func (a *NetworkSecurityRuleDeleteArgs) Validate() error {
	return nil
}

// Validate checks that list args are valid.
func (a *NetworkSecurityRuleListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkSecurityRuleCreate creates a new security rule.
func NetworkSecurityRuleCreate(ctx context.Context, client aruba.Client, args NetworkSecurityRuleCreateArgs) error {
	if args.Verbose {
		fmt.Println("Creating security rule with the following parameters:")
		fmt.Printf("  Name:         %s\n", args.Name)
		fmt.Printf("  Region:       %s\n", string(args.Region))
		fmt.Printf("  Direction:    %s\n", args.Direction)
		fmt.Printf("  Protocol:     %s\n", args.Protocol)
		fmt.Printf("  Port:         %s\n", args.Port)
		fmt.Printf("  Target Kind:  %s\n", args.TargetKind)
		fmt.Printf("  Target Value: %s\n", args.TargetValue)
		if len(args.Tags) > 0 {
			fmt.Printf("  Tags:         %v\n", args.Tags)
		}
		fmt.Println()
	}

	rule := aruba.NewSecurityRule().
		InSecurityGroup(aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID)).
		Named(args.Name).
		InRegion(args.Region).
		WithDirection(aruba.RuleDirection(args.Direction)).
		WithProtocol(aruba.RuleProtocol(args.Protocol)).
		RetaggedAs(args.Tags...)

	if args.Port != "" {
		rule = rule.WithPort(args.Port)
	}

	if aruba.EndpointTypeDto(args.TargetKind) == aruba.EndpointTypeSecurityGroup {
		rule = rule.TargetingSecurityGroup(aruba.URI(args.TargetValue))
	} else {
		rule = rule.TargetingCIDR(args.TargetValue)
	}

	resp, err := client.FromNetwork().SecurityGroupRules().Create(ctx, rule)
	if err != nil {
		return fmt.Errorf("creating security rule: %w", apiErrFromV2(err))
	}

	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "DIRECTION", Width: 12},
			{Header: "PROTOCOL", Width: 12},
			{Header: "PORT", Width: 12},
			{Header: "STATUS", Width: 15},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		row := []string{args.Name, id, args.Direction, args.Protocol, args.Port, status}
		PrintOutput(resp, headers, [][]string{row})
	} else {
		fmt.Println(msgCreatedAsync("Security rule", args.Name))
	}
	return nil
}

// NetworkSecurityRuleGet retrieves a security rule by ID.
func NetworkSecurityRuleGet(ctx context.Context, client aruba.Client, args NetworkSecurityRuleGetArgs) error {
	rule, err := client.FromNetwork().SecurityGroupRules().Get(ctx, aruba.SecurityRuleRef(args.ProjectID, args.VPCID, args.SGID, args.RuleID))
	if err != nil {
		return fmt.Errorf("getting security rule: %w", apiErrFromV2(err))
	}

	if rule != nil && rule.Raw() != nil {
		raw := rule.Raw()
		fmt.Println("\nSecurity Rule Details:")
		fmt.Println("=====================")
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
		fmt.Printf("Direction:       %s\n", raw.Properties.Direction)
		fmt.Printf("Protocol:        %s\n", raw.Properties.Protocol)
		fmt.Printf("Port:            %s\n", raw.Properties.Port)
		if raw.Properties.Target != nil {
			fmt.Printf("Target Kind:     %s\n", raw.Properties.Target.Kind)
			fmt.Printf("Target Value:    %s\n", raw.Properties.Target.Value)
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
		fmt.Println("Security rule not found or no data returned.")
	}
	return nil
}

// NetworkSecurityRuleUpdate updates an existing security rule.
func NetworkSecurityRuleUpdate(ctx context.Context, client aruba.Client, args NetworkSecurityRuleUpdateArgs) error {
	rule, err := client.FromNetwork().SecurityGroupRules().Get(ctx, aruba.SecurityRuleRef(args.ProjectID, args.VPCID, args.SGID, args.RuleID))
	if err != nil {
		return fmt.Errorf("fetching current security rule: %w", apiErrFromV2(err))
	}

	if rule == nil || rule.Raw() == nil {
		return fmt.Errorf("security rule not found or no data returned")
	}

	if rule.Raw().Status.State != nil && *rule.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update security rule while it is in 'InCreation' state. Please wait until the security rule is fully created")
	}

	if args.Name != "" {
		rule.Named(args.Name)
	}
	if args.TagsChanged {
		rule.RetaggedAs(args.Tags...)
	}

	if args.Debug {
		fmt.Fprintf(os.Stderr, "\n=== DEBUG: Security Rule Update ===\n")
		fmt.Fprintf(os.Stderr, "VPC ID: %s\n", args.VPCID)
		fmt.Fprintf(os.Stderr, "Security Group ID: %s\n", args.SGID)
		fmt.Fprintf(os.Stderr, "Security Rule ID: %s\n", args.RuleID)
		if reqJSON, err := json.Marshal(rule.RawRequest()); err == nil {
			fmt.Fprintf(os.Stderr, "Request: %s\n", reqJSON)
		}
		fmt.Fprintf(os.Stderr, "===================================\n\n")
	}

	updated, err := client.FromNetwork().SecurityGroupRules().Update(ctx, rule)
	if err != nil {
		return fmt.Errorf("updating security rule: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "DIRECTION", Width: 12},
			{Header: "PROTOCOL", Width: 12},
			{Header: "PORT", Width: 12},
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
		status := ""
		if raw.Status.State != nil {
			status = string(*raw.Status.State)
		}
		row := []string{nameVal, id, string(raw.Properties.Direction), string(raw.Properties.Protocol), raw.Properties.Port, status}
		PrintOutput(updated, headers, [][]string{row})
	} else {
		fmt.Println(msgUpdatedAsync("Security rule", args.RuleID))
	}
	return nil
}

// NetworkSecurityRuleDelete deletes a security rule.
func NetworkSecurityRuleDelete(ctx context.Context, client aruba.Client, args NetworkSecurityRuleDeleteArgs) error {
	err := client.FromNetwork().SecurityGroupRules().Delete(ctx, aruba.SecurityRuleRef(args.ProjectID, args.VPCID, args.SGID, args.RuleID))
	if err != nil {
		return fmt.Errorf("deleting security rule: %w", apiErrFromV2(err))
	}

	headers := []TableColumn{
		{Header: "ID", Width: 26},
		{Header: "STATUS", Width: 15},
	}
	status := "deleted"
	result := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{args.RuleID, status}
	PrintOutput(result, headers, [][]string{{args.RuleID, status}})
	return nil
}

// NetworkSecurityRuleList lists security rules for a security group.
func NetworkSecurityRuleList(ctx context.Context, client aruba.Client, args NetworkSecurityRuleListArgs) error {
	list, err := client.FromNetwork().SecurityGroupRules().List(ctx, aruba.SecurityGroupRef(args.ProjectID, args.VPCID, args.SGID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing security rules: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 26},
			{Header: "DIRECTION", Width: 12},
			{Header: "PROTOCOL", Width: 12},
			{Header: "PORT", Width: 12},
			{Header: "TARGET", Width: 30},
			{Header: "STATUS", Width: 15},
		}
		var rows [][]string
		for _, rule := range list.Items() {
			raw := rule.Raw()
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
			direction := string(raw.Properties.Direction)
			protocol := string(raw.Properties.Protocol)
			port := raw.Properties.Port
			target := ""
			if raw.Properties.Target != nil {
				target = fmt.Sprintf("%s:%s", raw.Properties.Target.Kind, raw.Properties.Target.Value)
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, direction, protocol, port, target, status})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No security rules found.")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkSecurityRuleCreateRun is the RunE wiring for security rule create.
func NetworkSecurityRuleCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityRuleCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityRuleCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityRuleGetRun is the RunE wiring for security rule get.
func NetworkSecurityRuleGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityRuleGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityRuleGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityRuleUpdateRun is the RunE wiring for security rule update.
func NetworkSecurityRuleUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityRuleUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityRuleUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityRuleDeleteRun is the RunE wiring for security rule delete.
func NetworkSecurityRuleDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityRuleDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("security rule", args.RuleID)
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
		_, err = client.FromNetwork().SecurityGroupRules().Get(ctx, aruba.SecurityRuleRef(args.ProjectID, args.VPCID, args.SGID, args.RuleID))
		if err != nil {
			return fmt.Errorf("dry-run: security rule not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("security rule", args.RuleID))
		return nil
	}

	if err := NetworkSecurityRuleDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkSecurityRuleListRun is the RunE wiring for security rule list.
func NetworkSecurityRuleListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkSecurityRuleListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkSecurityRuleList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
