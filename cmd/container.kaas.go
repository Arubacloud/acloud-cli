package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	kaasCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})")
	kaasCreateCmd.Flags().String("subnet-uri", "", "Subnet URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id})")
	kaasCreateCmd.Flags().String("node-cidr-address", "", "Node CIDR address in CIDR notation (required, e.g., 10.0.0.0/16)")
	kaasCreateCmd.Flags().String("node-cidr-name", "", "Node CIDR name (required)")
	kaasCreateCmd.Flags().String("security-group-name", "", "Security group name (required)")
	kaasCreateCmd.Flags().String("kubernetes-version", "", "Kubernetes version (required, e.g., 1.28.0)")
	kaasCreateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year (optional)")

	// Node pool flags (at least one node pool is required)
	kaasCreateCmd.Flags().String("node-pool-name", "", "Node pool name (required)")
	kaasCreateCmd.Flags().Int32("node-pool-nodes", 0, "Number of nodes in the node pool (required)")
	kaasCreateCmd.Flags().String("node-pool-instance", "", "Instance configuration name for nodes (required)")
	kaasCreateCmd.Flags().String("node-pool-zone", "", "Datacenter/zone code for nodes (required)")
	kaasCreateCmd.Flags().Bool("node-pool-autoscaling", false, "Enable autoscaling for node pool")
	kaasCreateCmd.Flags().Int32("node-pool-min-count", 0, "Minimum number of nodes for autoscaling")
	kaasCreateCmd.Flags().Int32("node-pool-max-count", 0, "Maximum number of nodes for autoscaling")

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

	// Properties update flags
	kaasUpdateCmd.Flags().String("kubernetes-version", "", "Kubernetes version to upgrade to")
	kaasUpdateCmd.Flags().String("kubernetes-version-upgrade-date", "", "Upgrade date for Kubernetes version (optional, ISO 8601 format)")
	kaasUpdateCmd.Flags().Bool("ha", false, "Enable/disable high availability")
	kaasUpdateCmd.Flags().Int32("storage-max-cumulative-volume-size", 0, "Maximum cumulative volume size for storage")
	kaasUpdateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")

	// Node pool update flags (for updating existing node pools)
	kaasUpdateCmd.Flags().String("node-pool-name", "", "Node pool name to update")
	kaasUpdateCmd.Flags().Int32("node-pool-nodes", 0, "Number of nodes in the node pool")
	kaasUpdateCmd.Flags().String("node-pool-instance", "", "Instance configuration name for nodes")
	kaasUpdateCmd.Flags().String("node-pool-zone", "", "Datacenter/zone code for nodes")
	kaasUpdateCmd.Flags().Bool("node-pool-autoscaling", false, "Enable autoscaling for node pool")
	kaasUpdateCmd.Flags().Int32("node-pool-min-count", 0, "Minimum number of nodes for autoscaling")
	kaasUpdateCmd.Flags().Int32("node-pool-max-count", 0, "Maximum number of nodes for autoscaling")

	kaasDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kaasDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	kaasListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	kaasConnectCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	kaasGetCmd.ValidArgsFunction = completeKaaSID
	kaasUpdateCmd.ValidArgsFunction = completeKaaSID
	kaasDeleteCmd.ValidArgsFunction = completeKaaSID
	kaasConnectCmd.ValidArgsFunction = completeKaaSID
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
	response, err := client.FromContainer().KaaS().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, kaas := range response.Data.Values {
			if kaas.Metadata.ID != nil && kaas.Metadata.Name != nil {
				id := *kaas.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *kaas.Metadata.Name))
				}
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
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get metadata flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Get required properties flags
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURI, _ := cmd.Flags().GetString("subnet-uri")
		nodeCIDRAddress, _ := cmd.Flags().GetString("node-cidr-address")
		nodeCIDRName, _ := cmd.Flags().GetString("node-cidr-name")
		securityGroupName, _ := cmd.Flags().GetString("security-group-name")
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes-version")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		// Get node pool flags
		nodePoolName, _ := cmd.Flags().GetString("node-pool-name")
		nodePoolNodes, _ := cmd.Flags().GetInt32("node-pool-nodes")
		nodePoolInstance, _ := cmd.Flags().GetString("node-pool-instance")
		nodePoolZone, _ := cmd.Flags().GetString("node-pool-zone")
		nodePoolAutoscaling, _ := cmd.Flags().GetBool("node-pool-autoscaling")
		nodePoolMinCount, _ := cmd.Flags().GetInt32("node-pool-min-count")
		nodePoolMaxCount, _ := cmd.Flags().GetInt32("node-pool-max-count")

		// Get optional flags
		podCIDR, _ := cmd.Flags().GetString("pod-cidr")
		ha, _ := cmd.Flags().GetBool("ha")
		apiServerAuthorizedIPRanges, _ := cmd.Flags().GetStringSlice("api-server-authorized-ip-ranges")
		apiServerEnablePrivateCluster, _ := cmd.Flags().GetBool("api-server-enable-private-cluster")

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build node pool
		nodePool := types.NodePoolProperties{
			Name:        nodePoolName,
			Nodes:       nodePoolNodes,
			Instance:    nodePoolInstance,
			Zone:        nodePoolZone,
			Autoscaling: nodePoolAutoscaling,
		}
		if nodePoolMinCount > 0 {
			nodePool.MinCount = &nodePoolMinCount
		}
		if nodePoolMaxCount > 0 {
			nodePool.MaxCount = &nodePoolMaxCount
		}

		// Build HA flag if set
		var haPtr *bool
		if ha {
			haPtr = &ha
		}

		// Build the create request
		createRequest := types.KaaSRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.KaaSPropertiesRequest{
				VPC: types.ReferenceResource{
					URI: vpcURI,
				},
				Subnet: types.ReferenceResource{
					URI: subnetURI,
				},
				NodeCIDR: types.NodeCIDRProperties{
					Address: nodeCIDRAddress,
					Name:    nodeCIDRName,
				},
				SecurityGroup: types.SecurityGroupProperties{
					Name: securityGroupName,
				},
				KubernetesVersion: types.KubernetesVersionInfo{
					Value: kubernetesVersion,
				},
				NodePools: []types.NodePoolProperties{nodePool},
			},
		}

		// Add optional fields
		if podCIDR != "" {
			createRequest.Properties.PodCIDR = &podCIDR
		}
		if haPtr != nil {
			createRequest.Properties.HA = haPtr
		}
		// BillingPlan is optional - only set if provided
		if billingPeriod != "" {
			createRequest.Properties.BillingPlan = types.BillingPeriodResource{
				BillingPeriod: billingPeriod,
			}
		}
		// APIServerAccessProfile is optional - only set if provided
		if len(apiServerAuthorizedIPRanges) > 0 || apiServerEnablePrivateCluster {
			apiServerAccessProfile := &types.APIServerAccessProfileProperties{
				EnablePrivateCluster: apiServerEnablePrivateCluster,
			}
			if len(apiServerAuthorizedIPRanges) > 0 {
				apiServerAccessProfile.AuthorizedIPRanges = &apiServerAuthorizedIPRanges
			}
			createRequest.Properties.APIServerAccessProfile = apiServerAccessProfile
		}

		ctx := context.Background()
		response, err := client.FromContainer().KaaS().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating KaaS cluster: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create KaaS cluster - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "VERSION", Width: 20},
				{Header: "REGION", Width: 20},
			}
			version := ""
			if response.Data.Properties.KubernetesVersion.Value != nil {
				version = *response.Data.Properties.KubernetesVersion.Value
			}
			row := []string{
				func() string {
					if response.Data.Metadata.ID != nil {
						return *response.Data.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.Name != nil {
						return *response.Data.Metadata.Name
					}
					return ""
				}(),
				version,
				func() string {
					if response.Data.Metadata.LocationResponse != nil {
						return response.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("KaaS cluster created, but no data returned.")
		}
	},
}

var kaasGetCmd = &cobra.Command{
	Use:   "get [kaas-id]",
	Short: "Get KaaS cluster details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
		if err != nil {
			fmt.Printf("Error getting KaaS cluster: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get KaaS cluster - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			kaas := resp.Data

			fmt.Println("\nKaaS Cluster Details:")
			fmt.Println("====================")

			if kaas.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *kaas.Metadata.ID)
			}
			if kaas.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *kaas.Metadata.URI)
			}
			if kaas.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *kaas.Metadata.Name)
			}
			if kaas.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", kaas.Metadata.LocationResponse.Value)
			}
			if kaas.Properties.KubernetesVersion.Value != nil {
				fmt.Printf("Kubernetes Version: %s\n", *kaas.Properties.KubernetesVersion.Value)
			}
			if kaas.Status.State != nil {
				fmt.Printf("Status:          %s\n", *kaas.Status.State)
			}

			if !kaas.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", kaas.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if kaas.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *kaas.Metadata.CreatedBy)
			}

			if len(kaas.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", kaas.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			// Show JSON output if verbose
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				jsonData, _ := json.MarshalIndent(kaas, "", "  ")
				fmt.Println("\nFull JSON Response:")
				fmt.Println("==================")
				fmt.Println(string(jsonData))
			}
		} else {
			fmt.Println("KaaS cluster not found or no data returned.")
		}
	},
}

var kaasUpdateCmd = &cobra.Command{
	Use:   "update [kaas-id]",
	Short: "Update a KaaS cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		getResponse, err := client.FromContainer().KaaS().Get(ctx, projectID, kaasID, nil)
		if err != nil {
			fmt.Printf("Error getting KaaS cluster details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Error: KaaS cluster not found")
			return
		}

		current := getResponse.Data

		// Get metadata update flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		region := ""
		if current.Metadata.LocationResponse != nil {
			region = current.Metadata.LocationResponse.Value
		}

		// Get properties update flags
		kubernetesVersion, _ := cmd.Flags().GetString("kubernetes-version")
		kubernetesVersionUpgradeDate, _ := cmd.Flags().GetString("kubernetes-version-upgrade-date")
		haFlag, _ := cmd.Flags().GetBool("ha")
		storageMaxSize, _ := cmd.Flags().GetInt32("storage-max-cumulative-volume-size")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		// Get node pool update flags
		nodePoolName, _ := cmd.Flags().GetString("node-pool-name")
		nodePoolNodes, _ := cmd.Flags().GetInt32("node-pool-nodes")
		nodePoolInstance, _ := cmd.Flags().GetString("node-pool-instance")
		nodePoolZone, _ := cmd.Flags().GetString("node-pool-zone")
		nodePoolAutoscaling, _ := cmd.Flags().GetBool("node-pool-autoscaling")
		nodePoolMinCount, _ := cmd.Flags().GetInt32("node-pool-min-count")
		nodePoolMaxCount, _ := cmd.Flags().GetInt32("node-pool-max-count")

		// Build metadata update (use current values if not provided)
		updateName := name
		if updateName == "" && current.Metadata.Name != nil {
			updateName = *current.Metadata.Name
		}
		updateTags := tags
		if len(updateTags) == 0 {
			updateTags = current.Metadata.Tags
		}

		// Build properties update
		// KubernetesVersion is required - use provided or current
		kubernetesVersionValue := kubernetesVersion
		if kubernetesVersionValue == "" {
			if current.Properties.KubernetesVersion.Value != nil {
				kubernetesVersionValue = *current.Properties.KubernetesVersion.Value
			} else {
				fmt.Println("Error: Kubernetes version is required")
				return
			}
		}

		kubernetesVersionUpdate := types.KubernetesVersionInfoUpdate{
			Value: kubernetesVersionValue,
		}
		if kubernetesVersionUpgradeDate != "" {
			kubernetesVersionUpdate.UpgradeDate = &kubernetesVersionUpgradeDate
		}

		// NodePools is required - use provided values or current
		var nodePools []types.NodePoolProperties
		if nodePoolName != "" {
			// Update specific node pool or add new one
			nodePool := types.NodePoolProperties{
				Name:        nodePoolName,
				Nodes:       nodePoolNodes,
				Instance:    nodePoolInstance,
				Zone:        nodePoolZone,
				Autoscaling: nodePoolAutoscaling,
			}
			if nodePoolMinCount > 0 {
				nodePool.MinCount = &nodePoolMinCount
			}
			if nodePoolMaxCount > 0 {
				nodePool.MaxCount = &nodePoolMaxCount
			}
			nodePools = []types.NodePoolProperties{nodePool}
		} else {
			// Use current node pools
			if current.Properties.NodePools != nil {
				for _, np := range *current.Properties.NodePools {
					nodePool := types.NodePoolProperties{
						Name: func() string {
							if np.Name != nil {
								return *np.Name
							}
							return ""
						}(),
						Nodes: func() int32 {
							if np.Nodes != nil {
								return *np.Nodes
							}
							return 0
						}(),
						Autoscaling: np.Autoscaling,
						MinCount:    np.MinCount,
						MaxCount:    np.MaxCount,
					}
					if np.Instance != nil && np.Instance.Name != nil {
						nodePool.Instance = *np.Instance.Name
					}
					if np.DataCenter != nil && np.DataCenter.Code != nil {
						nodePool.Zone = *np.DataCenter.Code
					}
					nodePools = append(nodePools, nodePool)
				}
			}
		}

		// Build optional fields
		var haPtr *bool
		if cmd.Flags().Changed("ha") {
			haPtr = &haFlag
		} else if current.Properties.HA != nil {
			haPtr = current.Properties.HA
		}

		var storage *types.StorageKubernetes
		if storageMaxSize > 0 {
			storage = &types.StorageKubernetes{
				MaxCumulativeVolumeSize: &storageMaxSize,
			}
		} else if current.Properties.Storage != nil {
			storage = current.Properties.Storage
		}

		var billingPlan *types.BillingPeriodResource
		if billingPeriod != "" {
			billingPlan = &types.BillingPeriodResource{
				BillingPeriod: billingPeriod,
			}
		} else if current.Properties.BillingPlan != nil && current.Properties.BillingPlan.BillingPeriod != nil {
			billingPlan = &types.BillingPeriodResource{
				BillingPeriod: *current.Properties.BillingPlan.BillingPeriod,
			}
		}

		// Build the update request
		updateRequest := types.KaaSUpdateRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: updateName,
					Tags: updateTags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.KaaSPropertiesUpdateRequest{
				KubernetesVersion: kubernetesVersionUpdate,
				NodePools:         nodePools,
			},
		}

		if haPtr != nil {
			updateRequest.Properties.HA = haPtr
		}
		if storage != nil {
			updateRequest.Properties.Storage = storage
		}
		if billingPlan != nil {
			updateRequest.Properties.BillingPlan = billingPlan
		}

		response, err := client.FromContainer().KaaS().Update(ctx, projectID, kaasID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating KaaS cluster: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update KaaS cluster - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nKaaS cluster updated successfully!")
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *response.Data.Metadata.Name)
			}
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("KaaS cluster update initiated. Use 'get' to check status.")
		}
	},
}

var kaasDeleteCmd = &cobra.Command{
	Use:   "delete [kaas-id]",
	Short: "Delete a KaaS cluster",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Confirmation prompt
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete KaaS cluster '%s'? (yes/no): ", kaasID)
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "yes" && confirmation != "y" {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		ctx := context.Background()
		response, err := client.FromContainer().KaaS().Delete(ctx, projectID, kaasID, nil)
		if err != nil {
			fmt.Printf("Error deleting KaaS cluster: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to delete KaaS cluster - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		fmt.Printf("KaaS cluster '%s' deleted successfully.\n", kaasID)
	},
}

var kaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KaaS clusters",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		response, err := client.FromContainer().KaaS().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing KaaS clusters: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "VERSION", Width: 20},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, kaas := range response.Data.Values {
				id := ""
				if kaas.Metadata.ID != nil {
					id = *kaas.Metadata.ID
				}
				name := ""
				if kaas.Metadata.Name != nil {
					name = *kaas.Metadata.Name
				}
				version := ""
				if kaas.Properties.KubernetesVersion.Value != nil {
					version = *kaas.Properties.KubernetesVersion.Value
				}
				region := ""
				if kaas.Metadata.LocationResponse != nil {
					region = kaas.Metadata.LocationResponse.Value
				}
				status := ""
				if kaas.Status.State != nil {
					status = *kaas.Status.State
				}

				rows = append(rows, []string{
					id,
					name,
					version,
					region,
					status,
				})
			}

			PrintTable(headers, rows)
		} else {
			fmt.Println("No KaaS clusters found")
		}
	},
}

var kaasConnectCmd = &cobra.Command{
	Use:   "connect [kaas-id]",
	Short: "Connect to a KaaS cluster and configure kubectl",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		response, err := client.FromContainer().KaaS().DownloadKubeconfig(ctx, projectID, kaasID, nil)
		if err != nil {
			fmt.Printf("Error downloading kubeconfig: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to connect to KaaS cluster - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response == nil || response.Data == nil {
			fmt.Println("Error: No kubeconfig data returned")
			return
		}

		kubeconfig := response.Data

		// Decode base64 content
		decodedContent, err := base64.StdEncoding.DecodeString(kubeconfig.Content)
		if err != nil {
			fmt.Printf("Error decoding kubeconfig content: %v\n", err)
			return
		}

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting home directory: %v\n", err)
			return
		}

		// Create .kube directory if it doesn't exist
		kubeDir := filepath.Join(homeDir, ".kube")
		err = os.MkdirAll(kubeDir, 0755)
		if err != nil {
			fmt.Printf("Error creating .kube directory: %v\n", err)
			return
		}

		// Write kubeconfig file with name from response
		kubeconfigFile := filepath.Join(kubeDir, kubeconfig.Name)
		err = os.WriteFile(kubeconfigFile, decodedContent, 0600)
		if err != nil {
			fmt.Printf("Error writing kubeconfig file: %v\n", err)
			return
		}

		// Copy to config file (overwrite if exists)
		configFile := filepath.Join(kubeDir, "config")
		err = os.WriteFile(configFile, decodedContent, 0600)
		if err != nil {
			fmt.Printf("Error writing config file: %v\n", err)
			return
		}

		// Run kubectl cluster-info
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

		// Success message
		fmt.Println("KaaS successfully connected")
		fmt.Printf("Kubeconfig saved to: %s\n", kubeconfigFile)
		fmt.Printf("Default config updated: %s\n", configFile)
	},
}
