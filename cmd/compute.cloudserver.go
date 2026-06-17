package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	cloudserverCreateCmd.Flags().String("boot-disk-id", "", "Bootable block storage ID (required)")
	cloudserverCreateCmd.MarkFlagRequired("boot-disk-id")
	cloudserverCreateCmd.MarkFlagRequired("vpc-id")
	cloudserverCreateCmd.MarkFlagRequired("subnet-id")
	cloudserverCreateCmd.MarkFlagRequired("security-group-id")
	// CloudServer commands
	computeCmd.AddCommand(cloudserverCmd)
	cloudserverCmd.AddCommand(cloudserverCreateCmd)
	cloudserverCmd.AddCommand(cloudserverGetCmd)
	cloudserverCmd.AddCommand(cloudserverUpdateCmd)
	cloudserverCmd.AddCommand(cloudserverDeleteCmd)
	cloudserverCmd.AddCommand(cloudserverListCmd)
	cloudserverCmd.AddCommand(cloudserverPowerOnCmd)
	cloudserverCmd.AddCommand(cloudserverPowerOffCmd)
	cloudserverCmd.AddCommand(cloudserverSetPasswordCmd)
	cloudserverCmd.AddCommand(cloudserverConnectCmd)

	// Add flags for cloudserver commands
	cloudserverCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverCreateCmd.Flags().String("name", "", "Name for the cloud server (required)")
	cloudserverCreateCmd.Flags().String("region", "", "Region code (required)")
	cloudserverCreateCmd.Flags().String("zone", "", "Zone code (required, e.g., ITBG-1)")
	cloudserverCreateCmd.Flags().String("flavor", "", "Flavor name (required)")
	cloudserverCreateCmd.Flags().String("keypair-id", "", "Keypair ID (optional)")
	cloudserverCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	cloudserverCreateCmd.Flags().String("user-data-file", "", "Path to cloud-init YAML file (will be base64 encoded)")
	cloudserverCreateCmd.Flags().String("vpc-id", "", "VPC ID (required)")
	cloudserverCreateCmd.Flags().StringSlice("subnet-id", []string{}, "Subnet ID(s) (required, comma-separated)")
	cloudserverCreateCmd.Flags().StringSlice("security-group-id", []string{}, "Security Group ID(s) (required, comma-separated)")
	cloudserverCreateCmd.Flags().String("elasticip-id", "", "Elastic IP ID (optional)")
	cloudserverCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year (optional, default: Hour)")
	cloudserverCreateCmd.MarkFlagRequired("name")
	cloudserverCreateCmd.MarkFlagRequired("region")
	cloudserverCreateCmd.MarkFlagRequired("flavor")
	cloudserverCreateCmd.MarkFlagRequired("zone")

	cloudserverGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverGetCmd.Flags().Bool("verbose", false, "Print full JSON response")

	cloudserverCreateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")
	cloudserverUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverUpdateCmd.Flags().String("name", "", "New name for the cloud server")
	cloudserverUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	cloudserverUpdateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")

	cloudserverDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cloudserverDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	cloudserverListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	cloudserverListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	cloudserverPowerOnCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverPowerOffCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	cloudserverSetPasswordCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverSetPasswordCmd.Flags().String("password", "", "New password for the cloud server (required)")
	cloudserverSetPasswordCmd.MarkFlagRequired("password")

	cloudserverConnectCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverConnectCmd.Flags().String("user", "<user>", "SSH username (required - see documentation for image-specific users)")

	// Set up auto-completion for resource IDs
	cloudserverGetCmd.ValidArgsFunction = completeCloudServerID
	cloudserverUpdateCmd.ValidArgsFunction = completeCloudServerID
	cloudserverDeleteCmd.ValidArgsFunction = completeCloudServerID
	cloudserverPowerOnCmd.ValidArgsFunction = completeCloudServerID
	cloudserverPowerOffCmd.ValidArgsFunction = completeCloudServerID
	cloudserverSetPasswordCmd.ValidArgsFunction = completeCloudServerID
	cloudserverConnectCmd.ValidArgsFunction = completeCloudServerID
}

// cloudServerRef builds the combined project+server Ref that Get/Delete need.
type cloudServerGetView struct {
	ID, Name, Region, Flavor, CPU, RAM, HD, BootVolumeURI, KeypairURI, Status, Tags string
}

func cloudServerRef(projectID, serverID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Compute/cloudServers/" + serverID)
}

// Completion functions for compute resources
func completeCloudServerID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("cloudserver", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromCompute().CloudServers().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, server := range list.Items() {
			raw := server.Raw()
			if raw == nil || raw.Metadata.ID == nil || *raw.Metadata.ID == "" {
				continue
			}
			id := *raw.Metadata.ID
			var name string
			if raw.Metadata.Name != nil {
				name = *raw.Metadata.Name
			}
			completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// CloudServer subcommands
var cloudserverCmd = &cobra.Command{
	Use:   "cloudserver",
	Short: "Manage cloud servers",
	Long:  `Perform CRUD operations on cloud servers in Aruba Cloud.`,
}

var cloudserverCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloud server",
	Long: `Create a new cloud server in the specified region and VPC.

The boot disk is specified as a block storage volume ID. Network resources (VPC,
subnet, security group) must already exist; pass their IDs with --vpc-id,
--subnet-id, and --security-group-id.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud compute cloudserver create \
    --name my-server --region ITBG-Bergamo --zone ITBG-1 \
    --flavor CSO4A8 \
    --boot-disk-id <volume-id> \
    --vpc-id <vpc-id> \
    --subnet-id <subnet-id> \
    --security-group-id <sg-id>`,
	Args: cobra.NoArgs,
	RunE: ComputeCloudServerCreateRun,
}

var cloudserverGetCmd = &cobra.Command{
	Use:   "get [cloudserver-id]",
	Short: "Get cloud server details",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerGetRun,
}

var cloudserverUpdateCmd = &cobra.Command{
	Use:   "update [cloudserver-id]",
	Short: "Update a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerUpdateRun,
}

var cloudserverDeleteCmd = &cobra.Command{
	Use:   "delete [cloudserver-id]",
	Short: "Delete a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerDeleteRun,
}

var cloudserverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cloud servers",
	Args:  cobra.NoArgs,
	RunE:  ComputeCloudServerListRun,
}

var cloudserverPowerOnCmd = &cobra.Command{
	Use:   "power-on [cloudserver-id]",
	Short: "Power on a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerPowerOnRun,
}

var cloudserverPowerOffCmd = &cobra.Command{
	Use:   "power-off [cloudserver-id]",
	Short: "Power off a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerPowerOffRun,
}

var cloudserverSetPasswordCmd = &cobra.Command{
	Use:   "set-password [cloudserver-id]",
	Short: "Set password for a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerSetPasswordRun,
}

var cloudserverConnectCmd = &cobra.Command{
	Use:   "connect [cloudserver-id]",
	Short: "Get SSH connection information for a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeCloudServerConnectRun,
}

// =============================================================================
// Args structs
// =============================================================================

// ComputeCloudServerCreateArgs holds the typed arguments for creating a cloud server.
type ComputeCloudServerCreateArgs struct {
	ProjectID        string
	Name             string
	Region           aruba.Region
	Zone             aruba.Zone
	Flavor           aruba.CloudServerFlavor
	BootDiskID       string
	VPCID            string
	SubnetIDs        []string
	SecurityGroupIDs []string
	ElasticIPID      string
	KeypairID        string
	Tags             []string
	BillingPeriod    aruba.BillingPeriod
	UserDataFile     string
	Wait             bool
}

// ComputeCloudServerGetArgs holds the typed arguments for getting a cloud server.
type ComputeCloudServerGetArgs struct {
	ProjectID string
	ID        string
	Verbose   bool
}

// ComputeCloudServerUpdateArgs holds the typed arguments for updating a cloud server.
type ComputeCloudServerUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
	Wait        bool
}

// ComputeCloudServerDeleteArgs holds the typed arguments for deleting a cloud server.
type ComputeCloudServerDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// ComputeCloudServerListArgs holds the typed arguments for listing cloud servers.
type ComputeCloudServerListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// ComputeCloudServerPowerOnArgs holds the typed arguments for powering on a cloud server.
type ComputeCloudServerPowerOnArgs struct {
	ProjectID string
	ID        string
}

// ComputeCloudServerPowerOffArgs holds the typed arguments for powering off a cloud server.
type ComputeCloudServerPowerOffArgs struct {
	ProjectID string
	ID        string
}

// ComputeCloudServerSetPasswordArgs holds the typed arguments for setting a cloud server password.
type ComputeCloudServerSetPasswordArgs struct {
	ProjectID string
	ID        string
	Password  string
}

// ComputeCloudServerConnectArgs holds the typed arguments for connecting to a cloud server.
type ComputeCloudServerConnectArgs struct {
	ProjectID string
	ID        string
	User      string
}

// =============================================================================
// Constructors
// =============================================================================

// NewComputeCloudServerCreateArgsFromCobraCommand parses and validates args for create.
func NewComputeCloudServerCreateArgsFromCobraCommand(cmd *cobra.Command) (*ComputeCloudServerCreateArgs, error) {
	args := &ComputeCloudServerCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerGetArgsFromCobraCommand parses and validates args for get.
func NewComputeCloudServerGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerGetArgs, error) {
	args := &ComputeCloudServerGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerUpdateArgsFromCobraCommand parses and validates args for update.
func NewComputeCloudServerUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerUpdateArgs, error) {
	args := &ComputeCloudServerUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerDeleteArgsFromCobraCommand parses and validates args for delete.
func NewComputeCloudServerDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerDeleteArgs, error) {
	args := &ComputeCloudServerDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerListArgsFromCobraCommand parses and validates args for list.
func NewComputeCloudServerListArgsFromCobraCommand(cmd *cobra.Command) (*ComputeCloudServerListArgs, error) {
	args := &ComputeCloudServerListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerPowerOnArgsFromCobraCommand parses and validates args for power-on.
func NewComputeCloudServerPowerOnArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerPowerOnArgs, error) {
	args := &ComputeCloudServerPowerOnArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerPowerOffArgsFromCobraCommand parses and validates args for power-off.
func NewComputeCloudServerPowerOffArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerPowerOffArgs, error) {
	args := &ComputeCloudServerPowerOffArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerSetPasswordArgsFromCobraCommand parses and validates args for set-password.
func NewComputeCloudServerSetPasswordArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerSetPasswordArgs, error) {
	args := &ComputeCloudServerSetPasswordArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeCloudServerConnectArgsFromCobraCommand parses and validates args for connect.
func NewComputeCloudServerConnectArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeCloudServerConnectArgs, error) {
	args := &ComputeCloudServerConnectArgs{}
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
func (a *ComputeCloudServerCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if s, err := cmd.Flags().GetString("zone"); err == nil {
		a.Zone = aruba.Zone(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("flavor"); err == nil {
		a.Flavor = aruba.CloudServerFlavor(s)
	} else {
		errs = append(errs, err)
	}
	if a.BootDiskID, err = cmd.Flags().GetString("boot-disk-id"); err != nil {
		errs = append(errs, err)
	}
	if a.VPCID, err = cmd.Flags().GetString("vpc-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SubnetIDs, err = cmd.Flags().GetStringSlice("subnet-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SecurityGroupIDs, err = cmd.Flags().GetStringSlice("security-group-id"); err != nil {
		errs = append(errs, err)
	}
	if a.ElasticIPID, err = cmd.Flags().GetString("elasticip-id"); err != nil {
		errs = append(errs, err)
	}
	if a.KeypairID, err = cmd.Flags().GetString("keypair-id"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.UserDataFile, err = cmd.Flags().GetString("user-data-file"); err != nil {
		errs = append(errs, err)
	}
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *ComputeCloudServerGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.Verbose, err = cmd.Flags().GetBool("verbose"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *ComputeCloudServerUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *ComputeCloudServerDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *ComputeCloudServerListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the power-on args struct.
func (a *ComputeCloudServerPowerOnArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the power-off args struct.
func (a *ComputeCloudServerPowerOffArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the set-password args struct.
func (a *ComputeCloudServerSetPasswordArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.Password, err = cmd.Flags().GetString("password"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the connect args struct.
func (a *ComputeCloudServerConnectArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.User, err = cmd.Flags().GetString("user"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *ComputeCloudServerCreateArgs) Validate() error {
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
	if !slices.Contains(validZones, a.Zone) {
		errs = append(errs, fmt.Errorf("--zone %q: must be one of %v", a.Zone, validZones))
	}
	if !slices.Contains(validCloudServerFlavors, a.Flavor) {
		errs = append(errs, fmt.Errorf("--flavor %q: must be one of %v", a.Flavor, validCloudServerFlavors))
	}
	if !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	if a.VPCID == "" {
		errs = append(errs, errors.New("--vpc-id is required"))
	}
	if a.BootDiskID == "" {
		errs = append(errs, errors.New("--boot-disk-id is required"))
	}
	if len(a.SubnetIDs) == 0 {
		errs = append(errs, errors.New("--subnet-id requires at least one value"))
	}
	if len(a.SecurityGroupIDs) == 0 {
		errs = append(errs, errors.New("--security-group-id requires at least one value"))
	}
	if a.UserDataFile != "" {
		if _, err := os.Stat(a.UserDataFile); err != nil {
			errs = append(errs, fmt.Errorf("--user-data-file %q: %w", a.UserDataFile, err))
		}
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *ComputeCloudServerGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("cloud server ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *ComputeCloudServerUpdateArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("cloud server ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *ComputeCloudServerDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("cloud server ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *ComputeCloudServerListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the power-on args for correctness.
func (a *ComputeCloudServerPowerOnArgs) Validate() error {
	if a.ID == "" {
		return errors.New("cloud server ID is required")
	}
	return nil
}

// Validate checks the power-off args for correctness.
func (a *ComputeCloudServerPowerOffArgs) Validate() error {
	if a.ID == "" {
		return errors.New("cloud server ID is required")
	}
	return nil
}

// Validate checks the set-password args for correctness.
func (a *ComputeCloudServerSetPasswordArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("cloud server ID is required"))
	}
	if a.Password == "" {
		errs = append(errs, errors.New("--password is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the connect args for correctness.
func (a *ComputeCloudServerConnectArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("cloud server ID is required"))
	}
	if a.User == "" || a.User == "<user>" {
		errs = append(errs, errors.New("--user is required (e.g. ubuntu, centos, root)"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// ComputeCloudServerCreate creates a cloud server using the provided args and client.
func ComputeCloudServerCreate(ctx context.Context, client aruba.Client, args ComputeCloudServerCreateArgs) error {
	subnetRefs := make([]aruba.Ref, len(args.SubnetIDs))
	for i, s := range args.SubnetIDs {
		subnetRefs[i] = aruba.SubnetRef(args.ProjectID, args.VPCID, s)
	}
	sgRefs := make([]aruba.Ref, len(args.SecurityGroupIDs))
	for i, sg := range args.SecurityGroupIDs {
		sgRefs[i] = aruba.SecurityGroupRef(args.ProjectID, args.VPCID, sg)
	}

	server := aruba.NewCloudServer().
		InProject(projectRef(args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		InZone(args.Zone).
		OfFlavor(args.Flavor).
		BootingFrom(volumeRef(args.ProjectID, args.BootDiskID)).
		WithVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		OnSubnets(subnetRefs...).
		WithSecurityGroups(sgRefs...).
		RetaggedAs(args.Tags...).
		BilledBy(args.BillingPeriod)

	if args.ElasticIPID != "" {
		server.WithElasticIP(aruba.ElasticIPRef(args.ProjectID, args.ElasticIPID))
	}
	if args.KeypairID != "" {
		server.UsingKeyPair(keypairRef(args.ProjectID, args.KeypairID))
	}
	if args.UserDataFile != "" {
		fileContent, err := os.ReadFile(args.UserDataFile)
		if err != nil {
			return fmt.Errorf("reading user-data file: %w", err)
		}
		userDataBase64 := base64.StdEncoding.EncodeToString(fileContent)
		server.WithUserData(userDataBase64)
	}

	resp, err := client.FromCompute().CloudServers().Create(ctx, server)
	if err != nil {
		return fmt.Errorf("creating cloud server: %w", apiErrFromV2(err))
	}

	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "FLAVOR", Width: 20},
			{Header: "CPU", Width: 10},
			{Header: "RAM(GB)", Width: 15},
			{Header: "HD(GB)", Width: 15},
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
		flavorName := raw.Properties.Flavor.Name
		cpu := raw.Properties.Flavor.CPU
		ram := raw.Properties.Flavor.RAM
		hd := raw.Properties.Flavor.HD
		regionValue := ""
		if raw.Metadata.LocationResponse != nil {
			regionValue = string(raw.Metadata.LocationResponse.Value)
		}
		row := []string{
			id,
			nameVal,
			string(flavorName),
			fmt.Sprintf("%d", cpu),
			fmt.Sprintf("%d", ram),
			fmt.Sprintf("%d", hd),
			regionValue,
		}
		PrintOutput(resp, headers, [][]string{row})
		if args.Wait && id != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, id))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "Cloud server", args.Name); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgCreatedAsync("Cloud server", args.Name))
	}
	return nil
}

// ComputeCloudServerGet retrieves and displays a cloud server's details.
func ComputeCloudServerGet(ctx context.Context, client aruba.Client, args ComputeCloudServerGetArgs) error {
	server, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
	}

	if server == nil || server.Raw() == nil {
		fmt.Println("Cloud server not found or no data returned.")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(server, nil, nil)
		return nil
	}
	raw := server.Raw()
	view := cloudServerGetView{
		CPU:  fmt.Sprintf("%d", raw.Properties.Flavor.CPU),
		RAM:  fmt.Sprintf("%d GB", raw.Properties.Flavor.RAM),
		HD:   fmt.Sprintf("%d GB", raw.Properties.Flavor.HD),
		Tags: "[]",
	}
	if raw.Metadata.ID != nil {
		view.ID = *raw.Metadata.ID
	}
	if raw.Metadata.Name != nil {
		view.Name = *raw.Metadata.Name
	}
	if raw.Metadata.LocationResponse != nil {
		view.Region = string(raw.Metadata.LocationResponse.Value)
	}
	view.Flavor = string(raw.Properties.Flavor.Name)
	view.BootVolumeURI = raw.Properties.BootVolume.URI
	view.KeypairURI = raw.Properties.KeyPair.URI
	if raw.Status.State != nil {
		view.Status = string(*raw.Status.State)
	}
	if len(raw.Metadata.Tags) > 0 {
		view.Tags = fmt.Sprintf("%v", raw.Metadata.Tags)
	}
	if err := renderGet(cloudServerGetTmpl, view); err != nil {
		return err
	}
	if args.Verbose {
		jsonData, _ := json.MarshalIndent(raw, "", "  ")
		fmt.Println("\nFull JSON Response:")
		fmt.Println("==================")
		fmt.Println(string(jsonData))
	}
	return nil
}

// ComputeCloudServerUpdate updates a cloud server's name and/or tags.
func ComputeCloudServerUpdate(ctx context.Context, client aruba.Client, args ComputeCloudServerUpdateArgs) error {
	server, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("fetching current cloud server: %w", apiErrFromV2(err))
	}

	if server == nil || server.Raw() == nil {
		return fmt.Errorf("cloud server not found")
	}

	// The SDK's fromResponse hydrates subnetRefs from NetworkInterfaces[].Subnet but
	// cannot extract securityGroupRefs because the API surfaces them only inside
	// linkedResources[] without a distinguishable type field. Re-inject them here so
	// the PUT body keeps them; without them the API strips all SGs and returns 400.
	for _, lr := range server.Raw().Properties.LinkedResources {
		if strings.Contains(lr.URI, "/securityGroups/") {
			server.WithSecurityGroups(aruba.URI(lr.URI))
		}
	}

	if args.Name != "" {
		server.Named(args.Name)
	}
	if args.TagsChanged {
		server.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromCompute().CloudServers().Update(ctx, server)
	if err != nil {
		return fmt.Errorf("updating cloud server: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Cloud server", args.ID))
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
		if args.Wait && args.ID != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "Cloud server", args.ID); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgUpdatedAsync("Cloud server", args.ID))
	}
	return nil
}

// ComputeCloudServerDelete deletes a cloud server.
func ComputeCloudServerDelete(ctx context.Context, client aruba.Client, args ComputeCloudServerDeleteArgs) error {
	err := client.FromCompute().CloudServers().Delete(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting cloud server: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Cloud server", args.ID))
	return nil
}

// ComputeCloudServerList lists all cloud servers in a project.
func ComputeCloudServerList(ctx context.Context, client aruba.Client, args ComputeCloudServerListArgs) error {
	list, err := client.FromCompute().CloudServers().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing cloud servers: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No cloud servers found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.CloudServer]{
		{TableColumn: TableColumn{Header: "NAME", Width: 25}, Value: func(s *aruba.CloudServer) string {
			if r := s.Raw(); r != nil && r.Metadata.Name != nil {
				return *r.Metadata.Name
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "ID", Width: 30}, Value: func(s *aruba.CloudServer) string {
			if r := s.Raw(); r != nil && r.Metadata.ID != nil {
				return *r.Metadata.ID
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "LOCATION", Width: 15}, Value: func(s *aruba.CloudServer) string {
			if r := s.Raw(); r != nil && r.Metadata.LocationResponse != nil {
				return string(r.Metadata.LocationResponse.Value)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "FLAVOR", Width: 15}, Value: func(s *aruba.CloudServer) string {
			if r := s.Raw(); r != nil {
				return string(r.Properties.Flavor.Name)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(s *aruba.CloudServer) string {
			if r := s.Raw(); r != nil && r.Status.State != nil {
				return string(*r.Status.State)
			}
			return ""
		}},
	}, list.Items(), func(s *aruba.CloudServer) bool {
		r := s.Raw()
		return r != nil && r.Metadata.ID != nil && *r.Metadata.ID != ""
	})
	return nil
}

// ComputeCloudServerPowerOn powers on a cloud server.
func ComputeCloudServerPowerOn(ctx context.Context, client aruba.Client, args ComputeCloudServerPowerOnArgs) error {
	server, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
	}

	if err := server.PowerOn(ctx); err != nil {
		return fmt.Errorf("powering on cloud server: %w", apiErrFromV2(err))
	}

	fmt.Println(msgAction("Cloud server", args.ID, "powered on"))
	if server.Raw() != nil {
		if server.Raw().Metadata.Name != nil {
			fmt.Printf("Server: %s\n", *server.Raw().Metadata.Name)
		}
		if server.Raw().Status.State != nil {
			fmt.Printf("Status: %s\n", *server.Raw().Status.State)
		}
	}
	return nil
}

// ComputeCloudServerPowerOff powers off a cloud server.
func ComputeCloudServerPowerOff(ctx context.Context, client aruba.Client, args ComputeCloudServerPowerOffArgs) error {
	server, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
	}

	if err := server.PowerOff(ctx); err != nil {
		return fmt.Errorf("powering off cloud server: %w", apiErrFromV2(err))
	}

	fmt.Println(msgAction("Cloud server", args.ID, "powered off"))
	if server.Raw() != nil {
		if server.Raw().Metadata.Name != nil {
			fmt.Printf("Server: %s\n", *server.Raw().Metadata.Name)
		}
		if server.Raw().Status.State != nil {
			fmt.Printf("Status: %s\n", *server.Raw().Status.State)
		}
	}
	return nil
}

// ComputeCloudServerSetPassword sets the password for a cloud server.
func ComputeCloudServerSetPassword(ctx context.Context, client aruba.Client, args ComputeCloudServerSetPasswordArgs) error {
	server, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
	}

	if err := server.SetPassword(ctx, args.Password); err != nil {
		return fmt.Errorf("setting cloud server password: %w", apiErrFromV2(err))
	}

	fmt.Println(msgAction("Cloud server", args.ID, "password set"))
	return nil
}

// ComputeCloudServerConnect prints the SSH connection command for a cloud server.
func ComputeCloudServerConnect(ctx context.Context, client aruba.Client, args ComputeCloudServerConnectArgs) error {
	server, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
	}

	if server == nil || server.Raw() == nil {
		fmt.Println("Cloud server not found or no data returned.")
		return nil
	}

	raw := server.Raw()

	// Check for ElasticIP in linked resources
	var elasticIPURI string
	for _, linkedResource := range raw.Properties.LinkedResources {
		if strings.Contains(linkedResource.URI, "providers/Aruba.Network/elasticIps") {
			elasticIPURI = linkedResource.URI
			break
		}
	}

	if elasticIPURI == "" {
		fmt.Println("No Elastic IP found for this cloud server.")
		fmt.Println("The server must have an Elastic IP linked to use the connect command.")
		return nil
	}

	// Extract ElasticIP ID from URI
	elasticIPID := extractIDFromURI(elasticIPURI)
	if elasticIPID == "" {
		return fmt.Errorf("could not extract Elastic IP ID from URI: %s", elasticIPURI)
	}

	// Get ElasticIP details
	eip, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(args.ProjectID, elasticIPID))
	if err != nil {
		return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
	}

	if eip == nil || eip.Raw() == nil {
		fmt.Println("Elastic IP not found or no data returned.")
		return nil
	}

	eipRaw := eip.Raw()
	if eipRaw.Properties.Address == nil || *eipRaw.Properties.Address == "" {
		fmt.Println("Elastic IP address not available.")
		return nil
	}

	// Print SSH connection command
	fmt.Printf("Connect by running: ssh %s@%s\n", args.User, *eipRaw.Properties.Address)
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// ComputeCloudServerCreateRun is the Cobra RunE handler for cloud server create.
func ComputeCloudServerCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewComputeCloudServerCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerGetRun is the Cobra RunE handler for cloud server get.
func ComputeCloudServerGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerUpdateRun is the Cobra RunE handler for cloud server update.
func ComputeCloudServerUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerDeleteRun is the Cobra RunE handler for cloud server delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func ComputeCloudServerDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("cloud server", args.ID)
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
		if _, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: cloud server not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("cloud server", args.ID))
		return nil
	}

	if err := ComputeCloudServerDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerListRun is the Cobra RunE handler for cloud server list.
func ComputeCloudServerListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewComputeCloudServerListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerPowerOnRun is the Cobra RunE handler for cloud server power-on.
func ComputeCloudServerPowerOnRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerPowerOnArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerPowerOn(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerPowerOffRun is the Cobra RunE handler for cloud server power-off.
func ComputeCloudServerPowerOffRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerPowerOffArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerPowerOff(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerSetPasswordRun is the Cobra RunE handler for cloud server set-password.
func ComputeCloudServerSetPasswordRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerSetPasswordArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerSetPassword(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeCloudServerConnectRun is the Cobra RunE handler for cloud server connect.
func ComputeCloudServerConnectRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeCloudServerConnectArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		// Print user-friendly guidance before returning the error
		if errors.Is(err, ErrValidationFailed) && strings.Contains(err.Error(), "--user is required") {
			fmt.Println("Common SSH users by image type:")
			fmt.Println("  - Ubuntu/Debian: ubuntu")
			fmt.Println("  - CentOS/RHEL: centos or root")
			fmt.Println("  - Other Linux: root or check image documentation")
			fmt.Println("\nFor more information, see: https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx")
		}
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeCloudServerConnect(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
