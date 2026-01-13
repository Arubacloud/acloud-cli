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

	containerregistryListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

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
			fmt.Printf("Error initializing client: %v\n", err)
			return
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

		ctx := context.Background()
		containerClient := client.FromContainer()
		if containerClient == nil {
			fmt.Println("Error: Container client is not available")
			return
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			fmt.Println("Error: Container Registry client returned nil")
			fmt.Println("This may indicate that Container Registry is not available in your SDK version")
			return
		}
		response, err := registryClient.Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating container registry: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create container registry - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nContainer registry created successfully!")
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
			fmt.Println("Container registry creation initiated. Use 'list' or 'get' to check status.")
		}
	},
}

var containerregistryGetCmd = &cobra.Command{
	Use:   "get [containerregistry-id]",
	Short: "Get container registry details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registryID := args[0]

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
		containerClient := client.FromContainer()
		if containerClient == nil {
			fmt.Println("Error: Container client is not available")
			return
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			fmt.Println("Error: Container Registry client returned nil")
			fmt.Println("This may indicate that Container Registry is not available in your SDK version")
			return
		}
		resp, err := registryClient.Get(ctx, projectID, registryID, nil)
		if err != nil {
			fmt.Printf("Error getting container registry: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get container registry - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
				fmt.Printf("Region:          %s\n", registry.Metadata.LocationResponse.Value)
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

			if !registry.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", registry.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}

var containerregistryUpdateCmd = &cobra.Command{
	Use:   "update [containerregistry-id]",
	Short: "Update a container registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registryID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		concurrentUsers, _ := cmd.Flags().GetString("concurrent-users")

		if name == "" && len(tags) == 0 && billingPeriod == "" && concurrentUsers == "" {
			fmt.Println("Error: at least one of --name, --tags, --billing-period, or --concurrent-users must be provided")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()

		containerClient := client.FromContainer()
		if containerClient == nil {
			fmt.Println("Error: Container client is not available")
			return
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			fmt.Println("Error: Container Registry client returned nil")
			fmt.Println("This may indicate that Container Registry is not available in your SDK version")
			return
		}

		// Get current registry to preserve existing values
		getResponse, err := registryClient.Get(ctx, projectID, registryID, nil)
		if err != nil {
			fmt.Printf("Error getting container registry: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Container registry not found")
			return
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

		updateRequest := types.ContainerRegistryRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: updateName,
					Tags: updateTags,
				},
				Location: types.LocationRequest{
					Value: current.Metadata.LocationResponse.Value,
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
			fmt.Printf("Error updating container registry: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update container registry - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nContainer registry updated successfully!")
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
			fmt.Println("Container registry update initiated. Use 'get' to check status.")
		}
	},
}

var containerregistryDeleteCmd = &cobra.Command{
	Use:   "delete [containerregistry-id]",
	Short: "Delete a container registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registryID := args[0]

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
			fmt.Printf("Are you sure you want to delete container registry '%s'? (yes/no): ", registryID)
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "yes" && confirmation != "y" {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		ctx := context.Background()
		containerClient := client.FromContainer()
		if containerClient == nil {
			fmt.Println("Error: Container client is not available")
			return
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			fmt.Println("Error: Container Registry client returned nil")
			fmt.Println("This may indicate that Container Registry is not available in your SDK version")
			return
		}
		response, err := registryClient.Delete(ctx, projectID, registryID, nil)
		if err != nil {
			fmt.Printf("Error deleting container registry: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to delete container registry - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		fmt.Printf("\nContainer registry '%s' deleted successfully!\n", registryID)
	},
}

var containerregistryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all container registries",
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

		containerClient := client.FromContainer()
		if containerClient == nil {
			fmt.Println("Error: Container client is not available")
			return
		}
		registryClient := containerClient.ContainerRegistry()
		if registryClient == nil {
			fmt.Println("Error: Container Registry client returned nil")
			fmt.Println("This may indicate that Container Registry is not available in your SDK version")
			return
		}
		response, err := registryClient.List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing container registries: %v\n", err)
			return
		}

		if response == nil {
			fmt.Println("No response received from server")
			return
		}

		if response.IsError() && response.Error != nil {
			fmt.Printf("Failed to list container registries - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
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

			PrintTable(headers, rows)
		} else {
			fmt.Println("No container registries found")
		}
	},
}
