package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	containerCmd.AddCommand(kaasCmd)
	kaasCmd.AddCommand(kaasCreateCmd)
	kaasCmd.AddCommand(kaasGetCmd)
	kaasCmd.AddCommand(kaasUpdateCmd)
	kaasCmd.AddCommand(kaasDeleteCmd)
	kaasCmd.AddCommand(kaasListCmd)
	kaasCmd.AddCommand(kaasConnectCmd)

	kaasCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasCreateCmd.Flags().String("name", "", "Name for the KaaS cluster (required)")
	kaasCreateCmd.Flags().String("region", "", "Region code (required)")
	kaasCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	kaasCreateCmd.Flags().String("vpc-id", "", "VPC ID (required)")
	kaasCreateCmd.Flags().String("subnet-id", "", "Subnet ID (required)")
	kaasCreateCmd.Flags().String("node-cidr-address", "", "Node CIDR address in CIDR notation (required, e.g., 10.0.0.0/16)")
	kaasCreateCmd.Flags().String("node-cidr-name", "", "Node CIDR name (required)")
	kaasCreateCmd.Flags().String("security-group-name", "", "Security group name (required)")
	kaasCreateCmd.Flags().String("kubernetes-version", "", "Kubernetes version (required, e.g., 1.29)")
	kaasCreateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year (optional)")
	kaasCreateCmd.Flags().String("node-pool-name", "", "Node pool name (required)")
	kaasCreateCmd.Flags().Int("node-pool-nodes", 0, "Number of nodes in the node pool (required)")
	kaasCreateCmd.Flags().String("node-pool-instance", "", "Instance configuration name for nodes (required)")
	kaasCreateCmd.Flags().String("node-pool-zone", "", "Datacenter/zone code for nodes (required)")
	kaasCreateCmd.Flags().Bool("node-pool-autoscaling", false, "Enable autoscaling for node pool")
	kaasCreateCmd.Flags().Int32("node-pool-min-count", 0, "Minimum number of nodes for autoscaling")
	kaasCreateCmd.Flags().Int32("node-pool-max-count", 0, "Maximum number of nodes for autoscaling")
	kaasCreateCmd.Flags().String("pod-cidr", "", "Pod CIDR (optional)")
	kaasCreateCmd.Flags().Bool("ha", false, "Enable high availability")
	kaasCreateCmd.Flags().StringSlice("api-server-authorized-ip-ranges", []string{}, "Authorized IP ranges for API server access (optional)")
	kaasCreateCmd.Flags().Bool("api-server-enable-private-cluster", false, "Enable private cluster for API server (optional)")
	kaasCreateCmd.MarkFlagRequired("name")
	kaasCreateCmd.MarkFlagRequired("region")
	kaasCreateCmd.MarkFlagRequired("vpc-id")
	kaasCreateCmd.MarkFlagRequired("subnet-id")
	kaasCreateCmd.MarkFlagRequired("node-cidr-address")
	kaasCreateCmd.MarkFlagRequired("node-cidr-name")
	kaasCreateCmd.MarkFlagRequired("security-group-name")
	kaasCreateCmd.MarkFlagRequired("kubernetes-version")
	kaasCreateCmd.MarkFlagRequired("node-pool-name")
	kaasCreateCmd.MarkFlagRequired("node-pool-nodes")
	kaasCreateCmd.MarkFlagRequired("node-pool-instance")
	kaasCreateCmd.MarkFlagRequired("node-pool-zone")

	kaasGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	kaasUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasUpdateCmd.Flags().String("name", "", "New name for the KaaS cluster")
	kaasUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	kaasUpdateCmd.Flags().String("kubernetes-version", "", "Kubernetes version to upgrade to")
	kaasUpdateCmd.Flags().String("kubernetes-version-upgrade-date", "", "Upgrade date for Kubernetes version (optional, ISO 8601 format)")
	kaasUpdateCmd.Flags().Bool("ha", false, "Enable/disable high availability")
	kaasUpdateCmd.Flags().Int("storage-max-cumulative-volume-size", 0, "Maximum cumulative volume size for storage")
	kaasUpdateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")
	kaasUpdateCmd.Flags().String("node-pool-name", "", "Node pool name to update")
	kaasUpdateCmd.Flags().Int("node-pool-nodes", 0, "Number of nodes in the node pool")
	kaasUpdateCmd.Flags().String("node-pool-instance", "", "Instance configuration name for nodes")
	kaasUpdateCmd.Flags().String("node-pool-zone", "", "Datacenter/zone code for nodes")
	kaasUpdateCmd.Flags().Bool("node-pool-autoscaling", false, "Enable autoscaling for node pool")
	kaasUpdateCmd.Flags().Int("node-pool-min-count", 0, "Minimum number of nodes for autoscaling")
	kaasUpdateCmd.Flags().Int("node-pool-max-count", 0, "Maximum number of nodes for autoscaling")

	kaasDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	kaasDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	kaasListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	kaasListCmd.Flags().Int("offset", 0, "Number of results to skip")

	kaasConnectCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasConnectCmd.Flags().String("output-file", "", "Write kubeconfig to this file path (default: ~/.kube/<cluster-name>)")

	kaasGetCmd.ValidArgsFunction = completeKaaSID
	kaasUpdateCmd.ValidArgsFunction = completeKaaSID
	kaasDeleteCmd.ValidArgsFunction = completeKaaSID
	kaasConnectCmd.ValidArgsFunction = completeKaaSID
}

// kaasRef builds the URI for a specific KaaS cluster.
func kaasRef(projectID, kaasID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Container/kaas/" + kaasID)
}

func completeKaaSID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromContainer().KaaS().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, k := range list.Items() {
			id := k.KaaSID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, k.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var kaasCmd = &cobra.Command{
	Use:   "kaas",
	Short: "Manage Kubernetes as a Service (KaaS)",
	Long:  `Perform CRUD operations on KaaS resources in Aruba Cloud.`,
}

var kaasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new KaaS cluster",
	Long: `Create a managed Kubernetes cluster in the specified region.

The VPC and subnet must already exist. A node pool is required at creation time;
specify its name, instance type, node count, and availability zone.

Pass --ha to enable a highly available control plane.
Use --node-pool-autoscaling to enable cluster autoscaler (requires --node-pool-min-count
and --node-pool-max-count).

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud container kaas create \
    --name my-cluster --region ITBG-Bergamo \
    --vpc-id <vpc-id> \
    --subnet-id <subnet-id> \
    --node-cidr-address 10.0.0.0/24 --node-cidr-name my-cidr \
    --security-group-name my-sg \
    --kubernetes-version 1.32.3 \
    --node-pool-name workers --node-pool-nodes 3 \
    --node-pool-instance <flavor-id> --node-pool-zone IT-BG-1`,
	Args: cobra.NoArgs,
	RunE: ContainerKaaSCreateRun,
}

var kaasGetCmd = &cobra.Command{
	Use:   "get [kaas-id]",
	Short: "Get KaaS cluster details",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerKaaSGetRun,
}

var kaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KaaS clusters",
	Args:  cobra.NoArgs,
	RunE:  ContainerKaaSListRun,
}

var kaasUpdateCmd = &cobra.Command{
	Use:   "update [kaas-id]",
	Short: "Update a KaaS cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerKaaSUpdateRun,
}

var kaasDeleteCmd = &cobra.Command{
	Use:   "delete [kaas-id]",
	Short: "Delete a KaaS cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerKaaSDeleteRun,
}

var kaasConnectCmd = &cobra.Command{
	Use:   "connect [kaas-id]",
	Short: "Connect to a KaaS cluster and configure kubectl",
	Args:  cobra.ExactArgs(1),
	RunE:  ContainerKaaSConnectRun,
}

// =============================================================================
// Args structs
// =============================================================================

// ContainerKaaSCreateArgs holds the typed arguments for creating a KaaS cluster.
type ContainerKaaSCreateArgs struct {
	ProjectID                     string
	Name                          string
	Region                        aruba.Region
	Tags                          []string
	VPCID                         string
	SubnetID                      string
	NodeCIDRAddress               string
	NodeCIDRName                  string
	SecurityGroupName             string
	KubernetesVersion             aruba.KubernetesVersion
	BillingPeriod                 aruba.BillingPeriod
	NodePoolName                  string
	NodePoolNodes                 int
	NodePoolInstance              aruba.NodePoolInstance
	NodePoolZone                  aruba.Zone
	NodePoolAutoscaling           bool
	NodePoolMinCount              int
	NodePoolMaxCount              int
	PodCIDR                       string
	HA                            bool
	APIServerAuthorizedIPRanges   []string
	APIServerEnablePrivateCluster bool
}

// ContainerKaaSGetArgs holds the typed arguments for getting a KaaS cluster.
type ContainerKaaSGetArgs struct {
	ProjectID string
	ID        string
}

// ContainerKaaSListArgs holds the typed arguments for listing KaaS clusters.
type ContainerKaaSListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// ContainerKaaSUpdateArgs holds the typed arguments for updating a KaaS cluster.
type ContainerKaaSUpdateArgs struct {
	ProjectID           string
	ID                  string
	Name                string
	Tags                []string
	TagsChanged         bool
	KubernetesVersion   aruba.KubernetesVersion
	HA                  bool
	HAChanged           bool
	StorageMaxSize      int
	BillingPeriod       aruba.BillingPeriod
	NodePoolName        string
	NodePoolNodes       int
	NodePoolInstance    aruba.NodePoolInstance
	NodePoolZone        aruba.Zone
	NodePoolAutoscaling bool
	NodePoolMinCount    int
	NodePoolMaxCount    int
}

// ContainerKaaSDeleteArgs holds the typed arguments for deleting a KaaS cluster.
type ContainerKaaSDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// ContainerKaaSConnectArgs holds the typed arguments for connecting to a KaaS cluster.
type ContainerKaaSConnectArgs struct {
	ProjectID  string
	ID         string
	OutputFile string
}

// =============================================================================
// Constructors
// =============================================================================

// NewContainerKaaSCreateArgsFromCobraCommand parses and validates args for create.
func NewContainerKaaSCreateArgsFromCobraCommand(cmd *cobra.Command) (*ContainerKaaSCreateArgs, error) {
	args := &ContainerKaaSCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerKaaSGetArgsFromCobraCommand parses and validates args for get.
func NewContainerKaaSGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerKaaSGetArgs, error) {
	args := &ContainerKaaSGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerKaaSListArgsFromCobraCommand parses and validates args for list.
func NewContainerKaaSListArgsFromCobraCommand(cmd *cobra.Command) (*ContainerKaaSListArgs, error) {
	args := &ContainerKaaSListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerKaaSUpdateArgsFromCobraCommand parses and validates args for update.
func NewContainerKaaSUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerKaaSUpdateArgs, error) {
	args := &ContainerKaaSUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerKaaSDeleteArgsFromCobraCommand parses and validates args for delete.
func NewContainerKaaSDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerKaaSDeleteArgs, error) {
	args := &ContainerKaaSDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewContainerKaaSConnectArgsFromCobraCommand parses and validates args for connect.
func NewContainerKaaSConnectArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ContainerKaaSConnectArgs, error) {
	args := &ContainerKaaSConnectArgs{}
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
func (a *ContainerKaaSCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if a.VPCID, err = cmd.Flags().GetString("vpc-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SubnetID, err = cmd.Flags().GetString("subnet-id"); err != nil {
		errs = append(errs, err)
	}
	if a.NodeCIDRAddress, err = cmd.Flags().GetString("node-cidr-address"); err != nil {
		errs = append(errs, err)
	}
	if a.NodeCIDRName, err = cmd.Flags().GetString("node-cidr-name"); err != nil {
		errs = append(errs, err)
	}
	if a.SecurityGroupName, err = cmd.Flags().GetString("security-group-name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("kubernetes-version"); err == nil {
		a.KubernetesVersion = aruba.KubernetesVersion(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.NodePoolName, err = cmd.Flags().GetString("node-pool-name"); err != nil {
		errs = append(errs, err)
	}
	if a.NodePoolNodes, err = cmd.Flags().GetInt("node-pool-nodes"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("node-pool-instance"); err == nil {
		a.NodePoolInstance = aruba.NodePoolInstance(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("node-pool-zone"); err == nil {
		a.NodePoolZone = aruba.Zone(s)
	} else {
		errs = append(errs, err)
	}
	if a.NodePoolAutoscaling, err = cmd.Flags().GetBool("node-pool-autoscaling"); err != nil {
		errs = append(errs, err)
	}
	if n, err := cmd.Flags().GetInt32("node-pool-min-count"); err == nil {
		a.NodePoolMinCount = int(n)
	} else {
		errs = append(errs, err)
	}
	if n, err := cmd.Flags().GetInt32("node-pool-max-count"); err == nil {
		a.NodePoolMaxCount = int(n)
	} else {
		errs = append(errs, err)
	}
	if a.PodCIDR, err = cmd.Flags().GetString("pod-cidr"); err != nil {
		errs = append(errs, err)
	}
	if a.HA, err = cmd.Flags().GetBool("ha"); err != nil {
		errs = append(errs, err)
	}
	if a.APIServerAuthorizedIPRanges, err = cmd.Flags().GetStringSlice("api-server-authorized-ip-ranges"); err != nil {
		errs = append(errs, err)
	}
	if a.APIServerEnablePrivateCluster, err = cmd.Flags().GetBool("api-server-enable-private-cluster"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *ContainerKaaSGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *ContainerKaaSListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *ContainerKaaSUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if s, err := cmd.Flags().GetString("kubernetes-version"); err == nil {
		a.KubernetesVersion = aruba.KubernetesVersion(s)
	} else {
		errs = append(errs, err)
	}
	if a.HA, err = cmd.Flags().GetBool("ha"); err != nil {
		errs = append(errs, err)
	}
	a.HAChanged = cmd.Flags().Changed("ha")
	if a.StorageMaxSize, err = cmd.Flags().GetInt("storage-max-cumulative-volume-size"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.NodePoolName, err = cmd.Flags().GetString("node-pool-name"); err != nil {
		errs = append(errs, err)
	}
	if a.NodePoolNodes, err = cmd.Flags().GetInt("node-pool-nodes"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("node-pool-instance"); err == nil {
		a.NodePoolInstance = aruba.NodePoolInstance(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("node-pool-zone"); err == nil {
		a.NodePoolZone = aruba.Zone(s)
	} else {
		errs = append(errs, err)
	}
	if a.NodePoolAutoscaling, err = cmd.Flags().GetBool("node-pool-autoscaling"); err != nil {
		errs = append(errs, err)
	}
	if a.NodePoolMinCount, err = cmd.Flags().GetInt("node-pool-min-count"); err != nil {
		errs = append(errs, err)
	}
	if a.NodePoolMaxCount, err = cmd.Flags().GetInt("node-pool-max-count"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *ContainerKaaSDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the connect args struct.
func (a *ContainerKaaSConnectArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.OutputFile, err = cmd.Flags().GetString("output-file"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *ContainerKaaSCreateArgs) Validate() error {
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
	// KubernetesVersion: no fixed SDK set — check non-empty only.
	if a.KubernetesVersion == "" {
		errs = append(errs, errors.New("--kubernetes-version is required"))
	}
	if a.VPCID == "" {
		errs = append(errs, errors.New("--vpc-id is required"))
	}
	if a.SubnetID == "" {
		errs = append(errs, errors.New("--subnet-id is required"))
	}
	if a.NodeCIDRAddress == "" {
		errs = append(errs, errors.New("--node-cidr-address is required"))
	}
	if a.NodeCIDRName == "" {
		errs = append(errs, errors.New("--node-cidr-name is required"))
	}
	if a.SecurityGroupName == "" {
		errs = append(errs, errors.New("--security-group-name is required"))
	}
	if a.NodePoolName == "" {
		errs = append(errs, errors.New("--node-pool-name is required"))
	}
	if a.NodePoolNodes <= 0 {
		errs = append(errs, errors.New("--node-pool-nodes must be greater than 0"))
	}
	if !slices.Contains(validNodePoolInstances, a.NodePoolInstance) {
		errs = append(errs, fmt.Errorf("--node-pool-instance %q: must be one of %v", a.NodePoolInstance, validNodePoolInstances))
	}
	// BillingPeriod is optional; validate only when provided.
	if a.BillingPeriod != "" && !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *ContainerKaaSGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("KaaS cluster ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *ContainerKaaSListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *ContainerKaaSUpdateArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("KaaS cluster ID is required"))
	}
	// KubernetesVersion: no fixed SDK set — check non-empty only when provided.
	// BillingPeriod is optional; validate only when provided.
	if a.BillingPeriod != "" && !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	// NodePoolInstance is optional; validate only when provided.
	if a.NodePoolInstance != "" && !slices.Contains(validNodePoolInstances, a.NodePoolInstance) {
		errs = append(errs, fmt.Errorf("--node-pool-instance %q: must be one of %v", a.NodePoolInstance, validNodePoolInstances))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *ContainerKaaSDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("KaaS cluster ID is required")
	}
	return nil
}

// Validate checks the connect args for correctness.
func (a *ContainerKaaSConnectArgs) Validate() error {
	if a.ID == "" {
		return errors.New("KaaS cluster ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// ContainerKaaSCreate creates a new KaaS cluster.
func ContainerKaaSCreate(ctx context.Context, client aruba.Client, args ContainerKaaSCreateArgs) error {
	nodePool := aruba.NewNodePool().
		Named(args.NodePoolName).
		OfInstance(args.NodePoolInstance).
		InZone(args.NodePoolZone).
		WithCount(args.NodePoolNodes)
	if args.NodePoolAutoscaling {
		nodePool.WithAutoscaling(args.NodePoolMinCount, args.NodePoolMaxCount)
	}

	sg := aruba.NewSecurityGroup().Named(args.SecurityGroupName)

	kaas := aruba.NewKaaS().
		InProject(projectRef(args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		WithVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		WithSubnet(aruba.SubnetRef(args.ProjectID, args.VPCID, args.SubnetID)).
		WithNodeCIDR(args.NodeCIDRAddress, args.NodeCIDRName).
		WithSecurityGroup(sg).
		WithKubernetesVersion(args.KubernetesVersion).
		WithNodePools(nodePool).
		RetaggedAs(args.Tags...)

	if args.HA {
		kaas.HighlyAvailable()
	}
	if args.PodCIDR != "" {
		kaas.WithPodCIDR(args.PodCIDR)
	}
	if args.BillingPeriod != "" {
		kaas.BilledBy(args.BillingPeriod)
	}
	if len(args.APIServerAuthorizedIPRanges) > 0 || args.APIServerEnablePrivateCluster {
		// TECH_DEBT: TD-033 (#131) — types.KaaSAPIServerAccessProfilePropertiesRequest must be
		// referenced directly because the SDK provides no aruba-level constructor for
		// this type yet. Remove the types import from this file once sdk-go exposes
		// aruba.NewAPIServerAccessProfile() or an equivalent fluent setter.
		profile := &types.KaaSAPIServerAccessProfilePropertiesRequest{
			EnablePrivateCluster: args.APIServerEnablePrivateCluster,
		}
		if len(args.APIServerAuthorizedIPRanges) > 0 {
			profile.AuthorizedIPRanges = &args.APIServerAuthorizedIPRanges
		}
		kaas.WithAPIServerAccessProfile(profile)
	}

	created, err := client.FromContainer().KaaS().Create(ctx, kaas)
	if err != nil {
		return fmt.Errorf("creating KaaS cluster: %w", apiErrFromV2(err))
	}

	if created != nil && created.KaaSID() != "" {
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "VERSION", Width: 20},
			{Header: "REGION", Width: 20},
		}
		PrintOutput(created, headers, [][]string{{
			created.KaaSID(),
			created.Name(),
			string(created.KubernetesVersion()),
			string(created.Region()),
		}})
	} else {
		fmt.Println(msgCreatedAsync("KaaS cluster", args.Name))
	}
	return nil
}

// ContainerKaaSGet retrieves and displays a KaaS cluster's details.
func ContainerKaaSGet(ctx context.Context, client aruba.Client, args ContainerKaaSGetArgs) error {
	kaas, err := client.FromContainer().KaaS().Get(ctx, kaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
	}

	if kaas != nil && kaas.KaaSID() != "" {
		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(kaas, nil, nil)
			return nil
		}

		fmt.Println("\nKaaS Cluster Details:")
		fmt.Println("====================")
		fmt.Printf("ID:              %s\n", kaas.KaaSID())
		if kaas.URI() != "" {
			fmt.Printf("URI:             %s\n", kaas.URI())
		}
		fmt.Printf("Name:            %s\n", kaas.Name())
		if r := kaas.Region(); r != "" {
			fmt.Printf("Region:          %s\n", string(r))
		}
		if v := kaas.KubernetesVersion(); v != "" {
			fmt.Printf("Kubernetes Version: %s\n", string(v))
		}
		if s := kaas.State(); s != "" {
			fmt.Printf("Status:          %s\n", string(s))
		}
		if !kaas.CreatedAt().IsZero() {
			fmt.Printf("Creation Date:   %s\n", kaas.CreatedAt().Format(DateLayout))
		}
		// CreatedBy has no wrapper accessor in sdk-go v1.0.0 — TECH_DEBT: TD-033
		if raw := kaas.Raw(); raw != nil && raw.Metadata.CreatedBy != nil {
			fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
		}
		if tags := kaas.Tags(); len(tags) > 0 {
			fmt.Printf("Tags:            %v\n", tags)
		} else {
			fmt.Printf("Tags:            []\n")
		}
		fmt.Println()
	} else {
		fmt.Println("KaaS cluster not found")
	}
	return nil
}

// ContainerKaaSList lists all KaaS clusters in a project.
func ContainerKaaSList(ctx context.Context, client aruba.Client, args ContainerKaaSListArgs) error {
	list, err := client.FromContainer().KaaS().List(ctx, aruba.URI("/projects/"+args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing KaaS clusters: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "VERSION", Width: 20},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, k := range list.Items() {
			if k.KaaSID() == "" {
				continue
			}
			rows = append(rows, []string{
				k.KaaSID(),
				k.Name(),
				string(k.KubernetesVersion()),
				string(k.Region()),
				string(k.State()),
			})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No KaaS clusters found")
	}
	return nil
}

// ContainerKaaSUpdate updates a KaaS cluster's mutable fields.
func ContainerKaaSUpdate(ctx context.Context, client aruba.Client, args ContainerKaaSUpdateArgs) error {
	kaas, err := client.FromContainer().KaaS().Get(ctx, kaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
	}
	if kaas == nil || kaas.KaaSID() == "" {
		return fmt.Errorf("KaaS cluster not found")
	}

	if args.Name != "" {
		kaas.Named(args.Name)
	}
	if args.TagsChanged {
		kaas.RetaggedAs(args.Tags...)
	}
	if args.KubernetesVersion != "" {
		kaas.WithKubernetesVersion(args.KubernetesVersion)
	}
	if args.HAChanged && args.HA {
		kaas.HighlyAvailable()
	}
	if args.StorageMaxSize > 0 {
		kaas.WithMaxStorageQuotaGB(args.StorageMaxSize)
	}
	if args.BillingPeriod != "" {
		kaas.BilledBy(args.BillingPeriod)
	}
	if args.NodePoolName != "" {
		np := aruba.NewNodePool().
			Named(args.NodePoolName).
			OfInstance(args.NodePoolInstance).
			InZone(args.NodePoolZone).
			WithCount(args.NodePoolNodes)
		if args.NodePoolAutoscaling {
			np.WithAutoscaling(args.NodePoolMinCount, args.NodePoolMaxCount)
		}
		kaas.ReplaceNodePools(np)
	}

	updated, err := client.FromContainer().KaaS().Update(ctx, kaas)
	if err != nil {
		return fmt.Errorf("updating KaaS cluster: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.KaaSID() != "" {
		fmt.Printf("\n%s\n", msgUpdated("KaaS cluster", args.ID))
		fmt.Printf("ID:      %s\n", updated.KaaSID())
		fmt.Printf("Name:    %s\n", updated.Name())
		if tags := updated.Tags(); len(tags) > 0 {
			fmt.Printf("Tags:    %v\n", tags)
		}
	} else {
		fmt.Println(msgUpdatedAsync("KaaS cluster", args.ID))
	}
	return nil
}

// ContainerKaaSDelete deletes a KaaS cluster.
func ContainerKaaSDelete(ctx context.Context, client aruba.Client, args ContainerKaaSDeleteArgs) error {
	err := client.FromContainer().KaaS().Delete(ctx, kaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting KaaS cluster: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("KaaS cluster", args.ID))
	return nil
}

// ContainerKaaSConnect downloads the kubeconfig for a KaaS cluster and configures kubectl.
func ContainerKaaSConnect(ctx context.Context, client aruba.Client, args ContainerKaaSConnectArgs) error {
	got, err := client.FromContainer().KaaS().Get(ctx, kaasRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
	}
	if got == nil || got.KaaSID() == "" {
		return fmt.Errorf("KaaS cluster not found")
	}

	kubeconfigBytes, err := got.DownloadKubeconfig(ctx)
	if err != nil {
		return fmt.Errorf("downloading kubeconfig: %w", apiErrFromV2(err))
	}
	if kubeconfigBytes == nil {
		return fmt.Errorf("no kubeconfig data returned")
	}

	// kubeconfigBytes is []byte(resp.Data.Content) — base64-encoded YAML; decode before writing.
	decodedContent, err := base64.StdEncoding.DecodeString(string(kubeconfigBytes))
	if err != nil {
		// Already raw if decode fails — use as-is.
		decodedContent = kubeconfigBytes
	}

	// If an explicit output file was requested, write there and return.
	if args.OutputFile != "" {
		if err := os.WriteFile(args.OutputFile, decodedContent, 0600); err != nil {
			return fmt.Errorf("writing kubeconfig file: %w", err)
		}
		fmt.Println(msgAction("KaaS cluster", args.ID, "connected"))
		fmt.Printf("Kubeconfig saved to: %s\n", args.OutputFile)
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	kubeDir := filepath.Join(homeDir, ".kube")
	if err = os.MkdirAll(kubeDir, 0755); err != nil {
		return fmt.Errorf("creating .kube directory: %w", err)
	}

	clusterName := args.ID
	if n := got.Name(); n != "" {
		clusterName = n
	}

	kubeconfigFile := filepath.Join(kubeDir, clusterName)
	if err = os.WriteFile(kubeconfigFile, decodedContent, 0600); err != nil {
		return fmt.Errorf("writing kubeconfig file: %w", err)
	}

	configFile := filepath.Join(kubeDir, "config")
	if err = os.WriteFile(configFile, decodedContent, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	kubectlCmd := exec.Command("kubectl", "cluster-info")
	output, err := kubectlCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl cluster-info failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Println(msgAction("KaaS cluster", args.ID, "connected"))
	fmt.Printf("Kubeconfig saved to: %s\n", kubeconfigFile)
	fmt.Printf("Default config updated: %s\n", configFile)
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// ContainerKaaSCreateRun is the Cobra RunE handler for KaaS create.
func ContainerKaaSCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewContainerKaaSCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerKaaSCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerKaaSGetRun is the Cobra RunE handler for KaaS get.
func ContainerKaaSGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerKaaSGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerKaaSGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerKaaSListRun is the Cobra RunE handler for KaaS list.
func ContainerKaaSListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewContainerKaaSListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerKaaSList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerKaaSUpdateRun is the Cobra RunE handler for KaaS update.
func ContainerKaaSUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerKaaSUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerKaaSUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerKaaSDeleteRun is the Cobra RunE handler for KaaS delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func ContainerKaaSDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerKaaSDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("KaaS cluster", args.ID)
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
		if _, err := client.FromContainer().KaaS().Get(ctx, kaasRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: KaaS cluster not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("KaaS cluster", args.ID))
		return nil
	}

	if err := ContainerKaaSDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ContainerKaaSConnectRun is the Cobra RunE handler for KaaS connect.
func ContainerKaaSConnectRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewContainerKaaSConnectArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ContainerKaaSConnect(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
