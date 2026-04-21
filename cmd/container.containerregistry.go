package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// ContainerRegistry commands
	containerCmd.AddCommand(containerregistryCmd)
	containerregistryCmd.AddCommand(containerregistryCreateCmd)
	containerregistryCmd.AddCommand(containerregistryGetCmd)
	containerregistryCmd.AddCommand(containerregistryUpdateCmd)
	containerregistryCmd.AddCommand(containerregistryDeleteCmd)
	containerregistryCmd.AddCommand(containerregistryListCmd)

	// Add flags for ContainerRegistry commands
	containerregistryCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryCreateCmd.Flags().String("name", "", "Name for the container registry (required)")
	containerregistryCreateCmd.Flags().String("region", "", "Region code (required)")
	containerregistryCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")

	// Required properties
	containerregistryCreateCmd.Flags().String("public-ip-uri", "", "Public IP URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/elasticIps/{elasticip-id})")
	containerregistryCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})")
	containerregistryCreateCmd.Flags().String("subnet-uri", "", "Subnet URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/subnets/{subnet-id})")
	containerregistryCreateCmd.Flags().String("security-group-uri", "", "Security group URI (required)")
	containerregistryCreateCmd.Flags().String("block-storage-uri", "", "Block storage URI (required)")

	// Optional properties
	containerregistryCreateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year (optional)")
	containerregistryCreateCmd.Flags().String("admin-username", "", "Administrator username (optional)")
	containerregistryCreateCmd.Flags().String("concurrent-users", "", "Number of concurrent users (optional)")

	containerregistryCreateCmd.MarkFlagRequired("name")
	containerregistryCreateCmd.MarkFlagRequired("region")
	containerregistryCreateCmd.MarkFlagRequired("public-ip-uri")
	containerregistryCreateCmd.MarkFlagRequired("vpc-uri")
	containerregistryCreateCmd.MarkFlagRequired("subnet-uri")
	containerregistryCreateCmd.MarkFlagRequired("security-group-uri")
	containerregistryCreateCmd.MarkFlagRequired("block-storage-uri")

	containerregistryGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	containerregistryUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryUpdateCmd.Flags().String("name", "", "New name for the container registry")
	containerregistryUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	containerregistryUpdateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year")
	containerregistryUpdateCmd.Flags().String("concurrent-users", "", "Number of concurrent users")

	containerregistryDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	containerregistryDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	containerregistryListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	containerregistryListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	containerregistryGetCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryUpdateCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryDeleteCmd.ValidArgsFunction = completeContainerRegistryID
}

// Completion functions for container registry resources
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
	containerClient := client.FromContainer()
	if containerClient == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	registryClient := containerClient.ContainerRegistry()
	if registryClient == nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	response, err := registryClient.List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, registry := range response.Data.Values {
			if registry.Metadata.ID != nil && registry.Metadata.Name != nil {
				id := *registry.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *registry.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// ContainerRegistry subcommands
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
must already exist. Pass their URIs via the corresponding flags.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud container containerregistry create \
    --name my-registry --region IT-BG \
    --vpc-uri /projects/<proj-id>/providers/Aruba.Network/vpcs/<vpc-id> \
    --subnet-uri /projects/<proj-id>/providers/Aruba.Network/subnets/<subnet-id> \
    --security-group-uri /projects/<proj-id>/providers/Aruba.Network/securityGroups/<sg-id> \
    --public-ip-uri /projects/<proj-id>/providers/Aruba.Network/elasticIPs/<eip-id> \
    --block-storage-uri /projects/<proj-id>/providers/Aruba.Storage/blockStorages/<vol-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get metadata flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// Get required properties flags
		publicIPURI, _ := cmd.Flags().GetString("public-ip-uri")
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURI, _ := cmd.Flags().GetString("subnet-uri")
		securityGroupURI, _ := cmd.Flags().GetString("security-group-uri")
		blockStorageURI, _ := cmd.Flags().GetString("block-storage-uri")

		// Get optional properties flags
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		adminUsername, _ := cmd.Flags().GetString("admin-username")
		concurrentUsers, _ := cmd.Flags().GetString("concurrent-users")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build the create request
		createRequest := types.ContainerRegistryRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.ContainerRegistryPropertiesRequest{
				PublicIp: types.ReferenceResource{
					URI: publicIPURI,
				},
				VPC: types.ReferenceResource{
					URI: vpcURI,
				},
				Subnet: types.ReferenceResource{
					URI: subnetURI,
				},
				SecurityGroup: types.ReferenceResource{
					URI: securityGroupURI,
				},
				BlockStorage: types.ReferenceResource{
					URI: blockStorageURI,
				},
			},
		}

		// Add optional fields
		if billingPeriod != "" {
			createRequest.Properties.BillingPlan = &types.BillingPeriodResource{
				BillingPeriod: billingPeriod,
			}
		}
		if adminUsername != "" {
			createRequest.Properties.AdminUser = &types.UserCredential{
				Username: adminUsername,
			}
		}
		if concurrentUsers != "" {
			createRequest.Properties.ConcurrentUsers = &concurrentUsers
		}

		ctx, cancel := newCtx()
		defer cancel()
		containerClient := client.FromContainer()
		if containerClient == nil {
			return fmt.Errorf("container client is not available")
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			return fmt.Errorf("container Registry client returned nil — this may indicate that Container Registry is not available in your SDK version")
		}
		response, err := registryClient.Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating container registry: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("Container registry", name))
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			}
			if response.Data.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", response.Data.Metadata.LocationResponse.Value)
			}
			if response.Data.Status.State != nil {
				fmt.Printf("Status:          %s\n", *response.Data.Status.State)
			}
		} else {
			fmt.Println(msgCreatedAsync("Container registry", name))
		}
		return nil
	},
}

var containerregistryGetCmd = &cobra.Command{
	Use:   "get [containerregistry-id]",
	Short: "Get container registry details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryID := args[0]

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
		containerClient := client.FromContainer()
		if containerClient == nil {
			return fmt.Errorf("container client is not available")
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			return fmt.Errorf("container Registry client returned nil — this may indicate that Container Registry is not available in your SDK version")
		}
		resp, err := registryClient.Get(ctx, projectID, registryID, nil)
		if err != nil {
			return fmt.Errorf("getting container registry: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil {
			registry := resp.Data

			fmt.Println("\nContainer Registry Details:")
			fmt.Println("==========================")

			if registry.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *registry.Metadata.ID)
			}
			if registry.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *registry.Metadata.URI)
			}
			if registry.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *registry.Metadata.Name)
			}
			if registry.Metadata.LocationResponse != nil {
				if registry.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", registry.Metadata.LocationResponse.Value)
				}
			}

			if registry.Properties.PublicIp.URI != "" {
				fmt.Printf("Public IP:       %s\n", registry.Properties.PublicIp.URI)
			}
			if registry.Properties.VPC.URI != "" {
				fmt.Printf("VPC:             %s\n", registry.Properties.VPC.URI)
			}
			if registry.Properties.Subnet.URI != "" {
				fmt.Printf("Subnet:          %s\n", registry.Properties.Subnet.URI)
			}
			if registry.Properties.SecurityGroup.URI != "" {
				fmt.Printf("Security Group:  %s\n", registry.Properties.SecurityGroup.URI)
			}
			if registry.Properties.BlockStorage.URI != "" {
				fmt.Printf("Block Storage:   %s\n", registry.Properties.BlockStorage.URI)
			}

			if registry.Properties.BillingPlan != nil {
				fmt.Printf("Billing Period:  %s\n", registry.Properties.BillingPlan.BillingPeriod)
			}
			if registry.Properties.AdminUser != nil {
				fmt.Printf("Admin User:      %s\n", registry.Properties.AdminUser.Username)
			}
			if registry.Properties.ConcurrentUsers != nil {
				fmt.Printf("Concurrent Users: %s\n", *registry.Properties.ConcurrentUsers)
			}

			if registry.Status.State != nil {
				fmt.Printf("Status:          %s\n", *registry.Status.State)
			}

			if registry.Metadata.CreationDate != nil && !registry.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", registry.Metadata.CreationDate.Format(DateLayout))
			}
			if registry.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *registry.Metadata.CreatedBy)
			}

			if len(registry.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", registry.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
		} else {
			fmt.Println("Container registry not found or no data returned.")
		}
		return nil
	},
}

var containerregistryUpdateCmd = &cobra.Command{
	Use:   "update [containerregistry-id]",
	Short: "Update a container registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		concurrentUsers, _ := cmd.Flags().GetString("concurrent-users")

		if name == "" && len(tags) == 0 && billingPeriod == "" && concurrentUsers == "" {
			return fmt.Errorf("at least one of --name, --tags, --billing-period, or --concurrent-users must be provided")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		containerClient := client.FromContainer()
		if containerClient == nil {
			return fmt.Errorf("container client is not available")
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			return fmt.Errorf("container Registry client returned nil — this may indicate that Container Registry is not available in your SDK version")
		}

		// Get current registry to preserve existing values
		getResponse, err := registryClient.Get(ctx, projectID, registryID, nil)
		if err != nil {
			return fmt.Errorf("getting container registry: %w", err)
		}

		if getResponse == nil || getResponse.Data == nil {
			return fmt.Errorf("container registry not found")
		}

		current := getResponse.Data

		// Build update request - use ContainerRegistryRequest for updates
		// Preserve existing properties that aren't being updated
		updateName := name
		if updateName == "" && current.Metadata.Name != nil {
			updateName = *current.Metadata.Name
		}
		updateTags := tags
		if len(updateTags) == 0 {
			updateTags = current.Metadata.Tags
		}

		registryRegion := ""
		if current.Metadata.LocationResponse != nil {
			registryRegion = current.Metadata.LocationResponse.Value
		}
		updateRequest := types.ContainerRegistryRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: updateName,
					Tags: updateTags,
				},
				Location: types.LocationRequest{
					Value: registryRegion,
				},
			},
			Properties: types.ContainerRegistryPropertiesRequest{
				// Preserve existing required properties
				PublicIp:      current.Properties.PublicIp,
				VPC:           current.Properties.VPC,
				Subnet:        current.Properties.Subnet,
				SecurityGroup: current.Properties.SecurityGroup,
				BlockStorage:  current.Properties.BlockStorage,
			},
		}

		// Update billing period if provided, otherwise preserve current
		if billingPeriod != "" {
			updateRequest.Properties.BillingPlan = &types.BillingPeriodResource{
				BillingPeriod: billingPeriod,
			}
		} else if current.Properties.BillingPlan != nil {
			updateRequest.Properties.BillingPlan = current.Properties.BillingPlan
		}

		// Update concurrent users if provided, otherwise preserve current
		if concurrentUsers != "" {
			updateRequest.Properties.ConcurrentUsers = &concurrentUsers
		} else if current.Properties.ConcurrentUsers != nil {
			updateRequest.Properties.ConcurrentUsers = current.Properties.ConcurrentUsers
		}

		// Preserve admin user if it exists
		if current.Properties.AdminUser != nil {
			updateRequest.Properties.AdminUser = current.Properties.AdminUser
		}

		response, err := registryClient.Update(ctx, projectID, registryID, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating container registry: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("Container registry", registryID))
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			}
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
			if response.Data.Status.State != nil {
				fmt.Printf("Status:          %s\n", *response.Data.Status.State)
			}
		} else {
			fmt.Println(msgUpdatedAsync("Container registry", registryID))
		}
		return nil
	},
}

var containerregistryDeleteCmd = &cobra.Command{
	Use:   "delete [containerregistry-id]",
	Short: "Delete a container registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		registryID := args[0]

		// Confirmation prompt
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("container registry", registryID)
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
		containerClient := client.FromContainer()
		if containerClient == nil {
			return fmt.Errorf("container client is not available")
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			return fmt.Errorf("container Registry client returned nil — this may indicate that Container Registry is not available in your SDK version")
		}
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = registryClient.Get(ctx, projectID, registryID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: container registry not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("container registry", registryID))
			return nil
		}

		response, err := registryClient.Delete(ctx, projectID, registryID, nil)
		if err != nil {
			return fmt.Errorf("deleting container registry: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		fmt.Println(msgDeleted("Container registry", registryID))
		return nil
	},
}

var containerregistryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all container registries",
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

		containerClient := client.FromContainer()
		if containerClient == nil {
			return fmt.Errorf("container client is not available")
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			return fmt.Errorf("container Registry client returned nil — this may indicate that Container Registry is not available in your SDK version")
		}
		response, err := registryClient.List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing container registries: %w", err)
		}

		if response == nil {
			fmt.Println("No response received from server")
			return nil
		}

		if response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 30},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, registry := range response.Data.Values {
				name := ""
				if registry.Metadata.Name != nil {
					name = *registry.Metadata.Name
				}

				id := ""
				if registry.Metadata.ID != nil {
					id = *registry.Metadata.ID
				}

				region := ""
				if registry.Metadata.LocationResponse != nil {
					region = registry.Metadata.LocationResponse.Value
				}

				status := ""
				if registry.Status.State != nil {
					status = *registry.Status.State
				}

				rows = append(rows, []string{
					name,
					id,
					region,
					status,
				})
			}

			PrintOutput(response.Data, headers, rows)
		} else {
			fmt.Println("No container registries found")
		}
		return nil
	},
}
