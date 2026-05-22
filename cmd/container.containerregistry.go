package cmd

import (
	"context"
	"fmt"
	"strings"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
	containerregistryCreateCmd.Flags().String("concurrent-users", "", "Concurrent-users tier: Small, Medium, HighPerf (optional)")

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
	containerregistryUpdateCmd.Flags().String("concurrent-users", "", "Concurrent-users tier: Small, Medium, HighPerf")

	containerregistryDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	containerregistryDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	containerregistryListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	containerregistryListCmd.Flags().Int("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	containerregistryGetCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryUpdateCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryDeleteCmd.ValidArgsFunction = completeContainerRegistryID
}

// File-local Ref helpers

func containerRegistryRef(projectID, registryID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Container/registries/" + registryID)
}

func containerRegistryFromRaw(r *aruba.ContainerRegistry) *types.ContainerRegistryResponse {
	if r == nil {
		return nil
	}
	return r.Raw()
}

func containerRegistryListPayload(l *aruba.List[*aruba.ContainerRegistry]) any {
	if r, ok := l.Raw().(*types.Response[types.ContainerRegistryList]); ok && r != nil {
		return r.Data
	}
	return nil
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
	list, err := client.FromContainer().ContainerRegistry().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, r := range list.Items() {
			id := r.ContainerRegistryID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, r.Name()))
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

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		publicIPURI, _ := cmd.Flags().GetString("public-ip-uri")
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURI, _ := cmd.Flags().GetString("subnet-uri")
		securityGroupURI, _ := cmd.Flags().GetString("security-group-uri")
		blockStorageURI, _ := cmd.Flags().GetString("block-storage-uri")

		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		adminUsername, _ := cmd.Flags().GetString("admin-username")
		concurrentUsers, _ := cmd.Flags().GetString("concurrent-users")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		r := aruba.NewContainerRegistry().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithElasticIP(aruba.URI(publicIPURI)).
			WithVPC(aruba.URI(vpcURI)).
			WithSubnet(aruba.URI(subnetURI)).
			WithSecurityGroup(aruba.URI(securityGroupURI)).
			WithBlockStorage(aruba.URI(blockStorageURI))

		if len(tags) > 0 {
			r.ReplaceTags(tags...)
		}
		if billingPeriod != "" {
			r.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if adminUsername != "" {
			r.WithAdminUsername(adminUsername)
		}
		if concurrentUsers != "" {
			r.OfSize(aruba.ContainerRegistrySizeFlavor(concurrentUsers))
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromContainer().ContainerRegistry().Create(ctx, r)
		if err != nil {
			return fmt.Errorf("creating container registry: %w", apiErrFromV2(err))
		}

		resource := containerRegistryFromRaw(created)
		if resource != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}
			id := ""
			if resource.Metadata.ID != nil {
				id = *resource.Metadata.ID
			}
			resName := ""
			if resource.Metadata.Name != nil {
				resName = *resource.Metadata.Name
			}
			resRegion := ""
			if resource.Metadata.LocationResponse != nil {
				resRegion = string(resource.Metadata.LocationResponse.Value)
			}
			status := ""
			if resource.Status.State != nil {
				status = *resource.Status.State
			}
			PrintOutput(resource, headers, [][]string{{id, resName, resRegion, status}})
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
		got, err := client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(projectID, registryID))
		if err != nil {
			return fmt.Errorf("getting container registry: %w", apiErrFromV2(err))
		}

		resource := containerRegistryFromRaw(got)
		if resource != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(resource, nil, nil)
				return nil
			}

			fmt.Println("\nContainer Registry Details:")
			fmt.Println("==========================")

			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *resource.Metadata.URI)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if resource.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", resource.Metadata.LocationResponse.Value)
			}

			if resource.Properties.PublicIp.URI != "" {
				fmt.Printf("Public IP:       %s\n", resource.Properties.PublicIp.URI)
			}
			if resource.Properties.VPC.URI != "" {
				fmt.Printf("VPC:             %s\n", resource.Properties.VPC.URI)
			}
			if resource.Properties.Subnet.URI != "" {
				fmt.Printf("Subnet:          %s\n", resource.Properties.Subnet.URI)
			}
			if resource.Properties.SecurityGroup.URI != "" {
				fmt.Printf("Security Group:  %s\n", resource.Properties.SecurityGroup.URI)
			}
			if resource.Properties.BlockStorage.URI != "" {
				fmt.Printf("Block Storage:   %s\n", resource.Properties.BlockStorage.URI)
			}

			if resource.Properties.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", *resource.Properties.BillingPeriod)
			}
			if resource.Properties.AdminUser != nil {
				fmt.Printf("Admin User:      %s\n", resource.Properties.AdminUser.Username)
			}
			if resource.Properties.ConcurrentUsers != nil {
				fmt.Printf("Size:            %s\n", *resource.Properties.ConcurrentUsers)
			}

			if resource.Status.State != nil {
				fmt.Printf("Status:          %s\n", *resource.Status.State)
			}

			if resource.Metadata.CreationDate != nil && !resource.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", resource.Metadata.CreationDate.Format(DateLayout))
			}
			if resource.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *resource.Metadata.CreatedBy)
			}

			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", resource.Metadata.Tags)
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

		if name == "" && !cmd.Flags().Changed("tags") && billingPeriod == "" && concurrentUsers == "" {
			return fmt.Errorf("at least one of --name, --tags, --billing-period, or --concurrent-users must be provided")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		current, err := client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(projectID, registryID))
		if err != nil {
			return fmt.Errorf("fetching current container registry: %w", apiErrFromV2(err))
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}
		if billingPeriod != "" {
			current.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if concurrentUsers != "" {
			current.OfSize(aruba.ContainerRegistrySizeFlavor(concurrentUsers))
		}

		updated, err := client.FromContainer().ContainerRegistry().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating container registry: %w", apiErrFromV2(err))
		}

		resource := containerRegistryFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("Container registry", registryID))
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *resource.Metadata.Name)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", resource.Metadata.Tags)
			}
			if resource.Status.State != nil {
				fmt.Printf("Status:  %s\n", *resource.Status.State)
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
			if _, err := client.FromContainer().ContainerRegistry().Get(ctx, containerRegistryRef(projectID, registryID)); err != nil {
				return fmt.Errorf("dry-run: container registry not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("container registry", registryID))
			return nil
		}

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

		if err := client.FromContainer().ContainerRegistry().Delete(ctx, containerRegistryRef(projectID, registryID)); err != nil {
			return fmt.Errorf("deleting container registry: %w", apiErrFromV2(err))
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

		list, err := client.FromContainer().ContainerRegistry().List(ctx, projectRef(projectID), listOpts(cmd)...)
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
			for _, r := range list.Items() {
				raw := containerRegistryFromRaw(r)
				name, id, region, status := "", "", "", ""
				if raw != nil {
					if raw.Metadata.Name != nil {
						name = *raw.Metadata.Name
					}
					if raw.Metadata.ID != nil {
						id = *raw.Metadata.ID
					}
					if raw.Metadata.LocationResponse != nil {
						region = string(raw.Metadata.LocationResponse.Value)
					}
					if raw.Status.State != nil {
						status = *raw.Status.State
					}
				}
				rows = append(rows, []string{name, id, region, status})
			}

			PrintOutput(containerRegistryListPayload(list), headers, rows)
		} else {
			fmt.Println("No container registries found")
		}
		return nil
	},
}
