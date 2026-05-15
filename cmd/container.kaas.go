package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// KaaS commands
	containerCmd.AddCommand(kaasCmd)
	kaasCmd.AddCommand(kaasCreateCmd)
	kaasCmd.AddCommand(kaasGetCmd)
	kaasCmd.AddCommand(kaasUpdateCmd)
	kaasCmd.AddCommand(kaasDeleteCmd)
	kaasCmd.AddCommand(kaasListCmd)
	kaasCmd.AddCommand(kaasConnectCmd)

	// Add flags for KaaS commands
	kaasCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasCreateCmd.Flags().String("name", "", "Name for the KaaS cluster (required)")
	kaasCreateCmd.Flags().String("region", "", "Region code (required)")
	kaasCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")

	// Required properties
	kaasCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required)")
	kaasCreateCmd.Flags().String("subnet-uri", "", "Subnet URI (required)")
	kaasCreateCmd.Flags().String("node-cidr-address", "", "Node CIDR address in CIDR notation (required)")
	kaasCreateCmd.Flags().String("node-cidr-name", "", "Node CIDR name (required)")
	kaasCreateCmd.Flags().String("security-group-name", "", "Security group name (required)")
	kaasCreateCmd.Flags().String("kubernetes-version", "", "Kubernetes version (required)")
	kaasCreateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year (optional)")

	// Node pool flags
	kaasCreateCmd.Flags().String("node-pool-name", "", "Node pool name (required)")
	kaasCreateCmd.Flags().Int("node-pool-nodes", 0, "Number of nodes in the node pool (required)")
	kaasCreateCmd.Flags().String("node-pool-instance", "", "Instance configuration name for nodes (required)")
	kaasCreateCmd.Flags().String("node-pool-zone", "", "Datacenter/zone code for nodes (required)")
	kaasCreateCmd.Flags().Bool("node-pool-autoscaling", false, "Enable autoscaling for node pool")
	kaasCreateCmd.Flags().Int("node-pool-min-count", 0, "Minimum number of nodes for autoscaling")
	kaasCreateCmd.Flags().Int("node-pool-max-count", 0, "Maximum number of nodes for autoscaling")

	// Optional properties
	kaasCreateCmd.Flags().String("pod-cidr", "", "Pod CIDR (optional)")
	kaasCreateCmd.Flags().Bool("ha", false, "Enable high availability")
	kaasCreateCmd.Flags().StringSlice("api-server-authorized-ip-ranges", []string{}, "Authorized IP ranges for API server access (optional)")
	kaasCreateCmd.Flags().Bool("api-server-enable-private-cluster", false, "Enable private cluster for API server (optional)")

	kaasCreateCmd.MarkFlagRequired("name")
	kaasCreateCmd.MarkFlagRequired("region")
	kaasCreateCmd.MarkFlagRequired("vpc-uri")
	kaasCreateCmd.MarkFlagRequired("subnet-uri")
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

	// Set up auto-completion for resource IDs
	kaasGetCmd.ValidArgsFunction = completeKaaSID
	kaasUpdateCmd.ValidArgsFunction = completeKaaSID
	kaasDeleteCmd.ValidArgsFunction = completeKaaSID
	kaasConnectCmd.ValidArgsFunction = completeKaaSID
}

// File-local Ref helpers

func kaasRef(projectID, kaasID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID)
}

func kaasFromRaw(k *aruba.KaaS) *types.KaaSResponse {
	if k == nil {
		return nil
	}
	return k.Raw()
}

func kaasListPayload(l *aruba.List[*aruba.KaaS]) any {
	if r, ok := l.Raw().(*types.Response[types.KaaSList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// Completion functions for container resources
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
	list, err := client.FromContainer().KaaS().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, k := range list.Items() {
			id := k.KaaSID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, k.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// KaaS subcommands
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
    --name my-cluster --region IT-BG \
    --vpc-uri /projects/<proj-id>/providers/Aruba.Network/vpcs/<vpc-id> \
    --subnet-uri /projects/<proj-id>/providers/Aruba.Network/subnets/<subnet-id> \
    --node-cidr-address 10.0.0.0/24 --node-cidr-name my-cidr \
    --security-group-name my-sg \
    --kubernetes-version 1.32.3 \
    --node-pool-name workers --node-pool-nodes 3 \
    --node-pool-instance <flavor-id> --node-pool-zone IT-BG-1`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURI, _ := cmd.Flags().GetString("subnet-uri")
		nodeCIDRAddress, _ := cmd.Flags().GetString("node-cidr-address")
		nodeCIDRName, _ := cmd.Flags().GetString("node-cidr-name")
		securityGroupName, _ := cmd.Flags().GetString("security-group-name")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes-version")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		nodePoolName, _ := cmd.Flags().GetString("node-pool-name")
		nodePoolNodes, _ := cmd.Flags().GetInt("node-pool-nodes")
		nodePoolInstance, _ := cmd.Flags().GetString("node-pool-instance")
		nodePoolZone, _ := cmd.Flags().GetString("node-pool-zone")
		nodePoolAutoscaling, _ := cmd.Flags().GetBool("node-pool-autoscaling")
		nodePoolMinCount, _ := cmd.Flags().GetInt("node-pool-min-count")
		nodePoolMaxCount, _ := cmd.Flags().GetInt("node-pool-max-count")

		podCIDR, _ := cmd.Flags().GetString("pod-cidr")
		ha, _ := cmd.Flags().GetBool("ha")
		apiServerAuthorizedIPRanges, _ := cmd.Flags().GetStringSlice("api-server-authorized-ip-ranges")
		apiServerEnablePrivateCluster, _ := cmd.Flags().GetBool("api-server-enable-private-cluster")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		nodePool := aruba.NewNodePool().
			Named(nodePoolName).
			WithCount(nodePoolNodes).
			OfInstance(aruba.NodePoolInstance(nodePoolInstance)).
			InZone(aruba.Zone(nodePoolZone))
		if nodePoolAutoscaling {
			nodePool.WithAutoscaling(nodePoolMinCount, nodePoolMaxCount)
		}

		// WithSecurityGroup requires *aruba.SecurityGroup (not a URI Ref) — the KaaS API stores only the SG name
		sg := aruba.NewSecurityGroup().Named(securityGroupName)

		k := aruba.NewKaaS().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithKubernetesVersion(aruba.KubernetesVersion(kubernetesVersion)).
			WithNodeCIDR(nodeCIDRAddress, nodeCIDRName).
			WithSecurityGroup(sg).
			WithVPC(aruba.URI(vpcURI)).
			WithSubnet(aruba.URI(subnetURI)).
			AddNodePool(nodePool)

		if len(tags) > 0 {
			k.ReplaceTags(tags...)
		}
		if podCIDR != "" {
			k.WithPodCIDR(podCIDR)
		}
		if ha {
			k.WithHA(true)
		}
		if billingPeriod != "" {
			k.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if apiServerEnablePrivateCluster || len(apiServerAuthorizedIPRanges) > 0 {
			profile := &types.APIServerAccessProfileProperties{
				EnablePrivateCluster: apiServerEnablePrivateCluster,
			}
			if len(apiServerAuthorizedIPRanges) > 0 {
				profile.AuthorizedIPRanges = &apiServerAuthorizedIPRanges
			}
			k.WithAPIServerAccessProfile(profile)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromContainer().KaaS().Create(ctx, k)
		if err != nil {
			return fmt.Errorf("creating KaaS cluster: %w", apiErrFromV2(err))
		}

		resource := kaasFromRaw(created)
		if resource != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "VERSION", Width: 20},
				{Header: "REGION", Width: 20},
			}
			id := ""
			if resource.Metadata.ID != nil {
				id = *resource.Metadata.ID
			}
			resName := ""
			if resource.Metadata.Name != nil {
				resName = *resource.Metadata.Name
			}
			version := ""
			if resource.Properties.KubernetesVersion.Value != nil {
				version = string(*resource.Properties.KubernetesVersion.Value)
			}
			resRegion := ""
			if resource.Metadata.LocationResponse != nil {
				resRegion = string(resource.Metadata.LocationResponse.Value)
			}
			PrintOutput(resource, headers, [][]string{{id, resName, version, resRegion}})
		} else {
			fmt.Println(msgCreatedAsync("KaaS cluster", name))
		}
		return nil
	},
}

var kaasGetCmd = &cobra.Command{
	Use:   "get [kaas-id]",
	Short: "Get KaaS cluster details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		got, err := client.FromContainer().KaaS().Get(ctx, kaasRef(projectID, kaasID))
		if err != nil {
			return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
		}

		resource := kaasFromRaw(got)
		if resource != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(resource, nil, nil)
				return nil
			}

			fmt.Println("\nKaaS Cluster Details:")
			fmt.Println("====================")

			if resource.Metadata.ID != nil {
				fmt.Printf("ID:                 %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.URI != nil {
				fmt.Printf("URI:                %s\n", *resource.Metadata.URI)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:               %s\n", *resource.Metadata.Name)
			}
			if resource.Metadata.LocationResponse != nil {
				fmt.Printf("Region:             %s\n", resource.Metadata.LocationResponse.Value)
			}
			if resource.Properties.KubernetesVersion.Value != nil {
				fmt.Printf("Kubernetes Version: %s\n", string(*resource.Properties.KubernetesVersion.Value))
			}
			if resource.Status.State != nil {
				fmt.Printf("Status:             %s\n", *resource.Status.State)
			}
			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:      %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
			if resource.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:         %s\n", *resource.Metadata.CreatedBy)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:               %v\n", resource.Metadata.Tags)
			} else {
				fmt.Printf("Tags:               []\n")
			}
			fmt.Println()
		} else {
			fmt.Println("KaaS cluster not found or no data returned.")
		}
		return nil
	},
}

var kaasUpdateCmd = &cobra.Command{
	Use:   "update [kaas-id]",
	Short: "Update a KaaS cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		// Get current state — fromResponse auto-populates all mutable fields for round-trip
		current, err := client.FromContainer().KaaS().Get(ctx, kaasRef(projectID, kaasID))
		if err != nil {
			return fmt.Errorf("fetching current KaaS cluster: %w", apiErrFromV2(err))
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes-version")
		haFlag, _ := cmd.Flags().GetBool("ha")
		storageMaxSize, _ := cmd.Flags().GetInt("storage-max-cumulative-volume-size")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		nodePoolName, _ := cmd.Flags().GetString("node-pool-name")
		nodePoolNodes, _ := cmd.Flags().GetInt("node-pool-nodes")
		nodePoolInstance, _ := cmd.Flags().GetString("node-pool-instance")
		nodePoolZone, _ := cmd.Flags().GetString("node-pool-zone")
		nodePoolAutoscaling, _ := cmd.Flags().GetBool("node-pool-autoscaling")
		nodePoolMinCount, _ := cmd.Flags().GetInt("node-pool-min-count")
		nodePoolMaxCount, _ := cmd.Flags().GetInt("node-pool-max-count")

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}
		if kubernetesVersion != "" {
			current.WithKubernetesVersion(aruba.KubernetesVersion(kubernetesVersion))
		}
		if cmd.Flags().Changed("ha") {
			current.WithHA(haFlag)
		}
		if storageMaxSize > 0 {
			current.WithMaxStorageQuotaGB(storageMaxSize)
		}
		if billingPeriod != "" {
			current.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if nodePoolName != "" {
			np := aruba.NewNodePool().
				Named(nodePoolName).
				WithCount(nodePoolNodes).
				OfInstance(aruba.NodePoolInstance(nodePoolInstance)).
				InZone(aruba.Zone(nodePoolZone))
			if nodePoolAutoscaling {
				np.WithAutoscaling(nodePoolMinCount, nodePoolMaxCount)
			}
			current.AddNodePool(np)
		}

		updated, err := client.FromContainer().KaaS().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating KaaS cluster: %w", apiErrFromV2(err))
		}

		resource := kaasFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("KaaS cluster", kaasID))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *resource.Metadata.Name)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", resource.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("KaaS cluster", kaasID))
		}
		return nil
	},
}

var kaasDeleteCmd = &cobra.Command{
	Use:   "delete [kaas-id]",
	Short: "Delete a KaaS cluster",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			if _, err := client.FromContainer().KaaS().Get(ctx, kaasRef(projectID, kaasID)); err != nil {
				return fmt.Errorf("dry-run: KaaS cluster not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("KaaS cluster", kaasID))
			return nil
		}

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("KaaS cluster", kaasID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromContainer().KaaS().Delete(ctx, kaasRef(projectID, kaasID)); err != nil {
			return fmt.Errorf("deleting KaaS cluster: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("KaaS cluster", kaasID))
		return nil
	},
}

var kaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KaaS clusters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromContainer().KaaS().List(ctx, projectRef(projectID), listOpts(cmd)...)
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
				rows = append(rows, []string{
					k.KaaSID(),
					k.Name(),
					string(k.KubernetesVersion()),
					string(k.Region()),
					k.State(),
				})
			}
			PrintOutput(kaasListPayload(list), headers, rows)
		} else {
			fmt.Println("No KaaS clusters found")
		}
		return nil
	},
}

var kaasConnectCmd = &cobra.Command{
	Use:   "connect [kaas-id]",
	Short: "Connect to a KaaS cluster and configure kubectl",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		// DownloadKubeconfig is a method on *KaaS (not on KaaSClient) — Get the wrapper first
		got, err := client.FromContainer().KaaS().Get(ctx, kaasRef(projectID, kaasID))
		if err != nil {
			return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
		}

		kubeconfigBytes, err := got.DownloadKubeconfig(ctx)
		if err != nil {
			return fmt.Errorf("downloading kubeconfig: %w", apiErrFromV2(err))
		}
		if kubeconfigBytes == nil {
			return fmt.Errorf("no kubeconfig data returned")
		}

		// Wire response carries base64-encoded YAML; decode before writing.
		// If decoding fails, assume content is already raw YAML and write as-is.
		decodedContent, err := base64.StdEncoding.DecodeString(string(kubeconfigBytes))
		if err != nil {
			decodedContent = kubeconfigBytes
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", err)
		}

		kubeDir := filepath.Join(homeDir, ".kube")
		if err = os.MkdirAll(kubeDir, 0755); err != nil {
			return fmt.Errorf("creating .kube directory: %w", err)
		}

		kubeconfigFile := filepath.Join(kubeDir, kaasID+".yaml")
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
			fmt.Printf("Error: kubectl cluster-info failed\n")
			fmt.Printf("Error details: %v\n", err)
			if len(output) > 0 {
				fmt.Printf("kubectl output: %s\n", string(output))
			}
			os.Exit(1)
		}

		fmt.Println(msgAction("KaaS cluster", kaasID, "connected"))
		fmt.Printf("Kubeconfig saved to: %s\n", kubeconfigFile)
		fmt.Printf("Default config updated: %s\n", configFile)
		return nil
	},
}
