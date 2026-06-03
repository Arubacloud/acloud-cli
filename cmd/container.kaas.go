package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	kaasGetCmd.ValidArgsFunction = completeKaaSID
	kaasUpdateCmd.ValidArgsFunction = completeKaaSID
	kaasDeleteCmd.ValidArgsFunction = completeKaaSID
	kaasConnectCmd.ValidArgsFunction = completeKaaSID
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
			id := k.ID()
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
    --name my-cluster --region IT-BG \
    --vpc-id <vpc-id> \
    --subnet-id <subnet-id> \
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
		vpcID, _ := cmd.Flags().GetString("vpc-id")
		subnetID, _ := cmd.Flags().GetString("subnet-id")
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
		nodePoolMinCount, _ := cmd.Flags().GetInt32("node-pool-min-count")
		nodePoolMaxCount, _ := cmd.Flags().GetInt32("node-pool-max-count")
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
			OfInstance(aruba.NodePoolInstance(nodePoolInstance)).
			InZone(aruba.Zone(nodePoolZone)).
			WithCount(int(nodePoolNodes))
		if nodePoolAutoscaling {
			nodePool.WithAutoscaling(int(nodePoolMinCount), int(nodePoolMaxCount))
		}

		kaas := aruba.NewKaaS().
			InProject(aruba.URI("/projects/"+projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithVPC(aruba.VPCRef(projectID, vpcID)).
			WithSubnet(aruba.SubnetRef(projectID, vpcID, subnetID)).
			WithNodeCIDR(nodeCIDRAddress, nodeCIDRName).
			WithSecurityGroupName(securityGroupName).
			WithKubernetesVersion(aruba.KubernetesVersion(kubernetesVersion)).
			WithNodePools(nodePool).
			RetaggedAs(tags...)
		if ha {
			kaas.HighlyAvailable()
		}

		if podCIDR != "" {
			kaas.WithPodCIDR(podCIDR)
		}
		if billingPeriod != "" {
			kaas.BilledBy(aruba.BillingPeriod(billingPeriod))
		}
		if len(apiServerAuthorizedIPRanges) > 0 || apiServerEnablePrivateCluster {
			// TECH_DEBT: TD-033 (#131) — types.KaaSAPIServerAccessProfilePropertiesRequest must be
			// referenced directly because the SDK provides no aruba-level constructor for
			// this type yet. Remove the types import from this file once sdk-go exposes
			// aruba.NewAPIServerAccessProfile() or an equivalent fluent setter.
			profile := &types.KaaSAPIServerAccessProfilePropertiesRequest{
				EnablePrivateCluster: apiServerEnablePrivateCluster,
			}
			if len(apiServerAuthorizedIPRanges) > 0 {
				profile.AuthorizedIPRanges = &apiServerAuthorizedIPRanges
			}
			kaas.WithAPIServerAccessProfile(profile)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromContainer().KaaS().Create(ctx, kaas)
		if err != nil {
			return fmt.Errorf("creating KaaS cluster: %w", apiErrFromV2(err))
		}

		if created != nil && created.ID() != "" {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "VERSION", Width: 20},
				{Header: "REGION", Width: 20},
			}
			PrintOutput(created, headers, [][]string{{
				created.ID(),
				created.Name(),
				string(created.KubernetesVersion()),
				string(created.Region()),
			}})
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
		kaasURI := "/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID
		kaas, err := client.FromContainer().KaaS().Get(ctx, aruba.URI(kaasURI))
		if err != nil {
			return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
		}

		if kaas != nil && kaas.ID() != "" {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(kaas, nil, nil)
				return nil
			}

			fmt.Println("\nKaaS Cluster Details:")
			fmt.Println("====================")
			fmt.Printf("ID:              %s\n", kaas.ID())
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
		list, err := client.FromContainer().KaaS().List(ctx, aruba.URI("/projects/"+projectID))
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
				if k.ID() == "" {
					continue
				}
				rows = append(rows, []string{
					k.ID(),
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
		kaasURI := "/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID
		kaas, err := client.FromContainer().KaaS().Get(ctx, aruba.URI(kaasURI))
		if err != nil {
			return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
		}
		if kaas == nil || kaas.ID() == "" {
			return fmt.Errorf("KaaS cluster not found")
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
			kaas.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			kaas.RetaggedAs(tags...)
		}
		if kubernetesVersion != "" {
			kaas.WithKubernetesVersion(aruba.KubernetesVersion(kubernetesVersion))
		}
		if cmd.Flags().Changed("ha") && haFlag {
			kaas.HighlyAvailable()
		}
		if storageMaxSize > 0 {
			kaas.WithMaxStorageQuotaGB(int(storageMaxSize))
		}
		if billingPeriod != "" {
			kaas.BilledBy(aruba.BillingPeriod(billingPeriod))
		}
		if nodePoolName != "" {
			np := aruba.NewNodePool().
				Named(nodePoolName).
				OfInstance(aruba.NodePoolInstance(nodePoolInstance)).
				InZone(aruba.Zone(nodePoolZone)).
				WithCount(int(nodePoolNodes))
			if nodePoolAutoscaling {
				np.WithAutoscaling(int(nodePoolMinCount), int(nodePoolMaxCount))
			}
			kaas.ReplaceNodePools(np)
		}

		updated, err := client.FromContainer().KaaS().Update(ctx, kaas)
		if err != nil {
			return fmt.Errorf("updating KaaS cluster: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.ID() != "" {
			fmt.Printf("\n%s\n", msgUpdated("KaaS cluster", kaasID))
			fmt.Printf("ID:      %s\n", updated.ID())
			fmt.Printf("Name:    %s\n", updated.Name())
			if tags := updated.Tags(); len(tags) > 0 {
				fmt.Printf("Tags:    %v\n", tags)
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

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			ok, err := confirmDelete("KaaS cluster", kaasID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

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
		kaasURI := "/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromContainer().KaaS().Get(ctx, aruba.URI(kaasURI))
			if err != nil {
				return fmt.Errorf("dry-run: KaaS cluster not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("KaaS cluster", kaasID))
			return nil
		}

		err = client.FromContainer().KaaS().Delete(ctx, aruba.URI(kaasURI))
		if err != nil {
			return fmt.Errorf("deleting KaaS cluster: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("KaaS cluster", kaasID))
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
		kaasURI := "/projects/" + projectID + "/providers/Aruba.Container/kaas/" + kaasID
		kaas, err := client.FromContainer().KaaS().Get(ctx, aruba.URI(kaasURI))
		if err != nil {
			return fmt.Errorf("getting KaaS cluster: %w", apiErrFromV2(err))
		}
		if kaas == nil || kaas.ID() == "" {
			return fmt.Errorf("KaaS cluster not found")
		}

		kubeconfigBytes, err := kaas.DownloadKubeconfig(ctx)
		if err != nil {
			return fmt.Errorf("downloading kubeconfig: %w", apiErrFromV2(err))
		}
		if kubeconfigBytes == nil {
			return fmt.Errorf("no kubeconfig data returned")
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("getting home directory: %w", apiErrFromV2(err))
		}

		kubeDir := filepath.Join(homeDir, ".kube")
		if err = os.MkdirAll(kubeDir, 0755); err != nil {
			return fmt.Errorf("creating .kube directory: %w", apiErrFromV2(err))
		}

		clusterName := kaasID
		if n := kaas.Name(); n != "" {
			clusterName = n
		}

		kubeconfigFile := filepath.Join(kubeDir, clusterName)
		err = os.WriteFile(kubeconfigFile, kubeconfigBytes, 0600)
		if err != nil {
			return fmt.Errorf("writing kubeconfig file: %w", err)
		}

		configFile := filepath.Join(kubeDir, "config")
		err = os.WriteFile(configFile, kubeconfigBytes, 0600)
		if err != nil {
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
