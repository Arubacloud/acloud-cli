package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	containerCmd.AddCommand(containerregistryCmd)
	containerregistryCmd.AddCommand(containerregistryCreateCmd)
	containerregistryCmd.AddCommand(containerregistryGetCmd)
	containerregistryCmd.AddCommand(containerregistryUpdateCmd)
	containerregistryCmd.AddCommand(containerregistryDeleteCmd)
	containerregistryCmd.AddCommand(containerregistryListCmd)

	containerregistryCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryCreateCmd.Flags().String("name", "", "Name for the container registry (required)")
	containerregistryCreateCmd.Flags().String("region", "", "Region code (required)")
	containerregistryCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	containerregistryCreateCmd.Flags().String("public-ip-uri", "", "Public IP URI (required)")
	containerregistryCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required)")
	containerregistryCreateCmd.Flags().String("subnet-uri", "", "Subnet URI (required)")
	containerregistryCreateCmd.Flags().String("security-group-uri", "", "Security group URI (required)")
	containerregistryCreateCmd.Flags().String("block-storage-uri", "", "Block storage URI (required)")
	containerregistryCreateCmd.Flags().String("billing-period", "", "Billing period: Hour, Month, Year (optional)")
	containerregistryCreateCmd.Flags().String("admin-username", "", "Administrator username (optional)")
	containerregistryCreateCmd.Flags().String("concurrent-users", "", "Concurrent users tier: Small, Medium, HighPerf (optional)")
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
	containerregistryUpdateCmd.Flags().String("concurrent-users", "", "Concurrent users tier: Small, Medium, HighPerf")

	containerregistryDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	containerregistryDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	containerregistryListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	containerregistryListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	containerregistryListCmd.Flags().Int("offset", 0, "Number of results to skip")

	containerregistryGetCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryUpdateCmd.ValidArgsFunction = completeContainerRegistryID
	containerregistryDeleteCmd.ValidArgsFunction = completeContainerRegistryID
}

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
	list, err := client.FromContainer().ContainerRegistry().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, cr := range list.Items() {
			raw := cr.Raw()
			if raw != nil && raw.Metadata.ID != nil && raw.Metadata.Name != nil {
				id := *raw.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *raw.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

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

		registry := aruba.NewContainerRegistry().
			IntoProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithElasticIP(aruba.URI(publicIPURI)).
			WithVPC(aruba.URI(vpcURI)).
			WithSubnet(aruba.URI(subnetURI)).
			WithSecurityGroup(aruba.URI(securityGroupURI)).
			WithBlockStorage(aruba.URI(blockStorageURI)).
			ReplaceTags(tags...)

		if adminUsername != "" {
			registry.WithAdminUsername(adminUsername)
		}
		if billingPeriod != "" {
			registry.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if concurrentUsers != "" {
			registry.OfSize(aruba.ContainerRegistrySizeFlavor(concurrentUsers))
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromContainer().ContainerRegistry().Create(ctx, registry)
		if err != nil {
			return fmt.Errorf("creating container registry: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			nameVal := ""
			if raw.Metadata.Name != nil {
				nameVal = *raw.Metadata.Name
			}
			regionVal := ""
			if raw.Metadata.LocationResponse != nil {
				regionVal = string(raw.Metadata.LocationResponse.Value)
			}
			statusVal := ""
			if raw.Status.State != nil {
				statusVal = string(*raw.Status.State)
			}
			PrintOutput(raw, headers, [][]string{{id, nameVal, regionVal, statusVal}})
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
		registryURI := "/projects/" + projectID + "/providers/Aruba.Container/registries/" + registryID
		registry, err := client.FromContainer().ContainerRegistry().Get(ctx, aruba.URI(registryURI))
		if err != nil {
			return fmt.Errorf("getting container registry: %w", apiErrFromV2(err))
		}

		if registry != nil && registry.Raw() != nil {
			raw := registry.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nContainer Registry Details:")
			fmt.Println("==========================")
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
			if raw.Properties.PublicIp.URI != "" {
				fmt.Printf("Public IP:       %s\n", raw.Properties.PublicIp.URI)
			}
			if raw.Properties.VPC.URI != "" {
				fmt.Printf("VPC:             %s\n", raw.Properties.VPC.URI)
			}
			if raw.Properties.Subnet.URI != "" {
				fmt.Printf("Subnet:          %s\n", raw.Properties.Subnet.URI)
			}
			if raw.Properties.SecurityGroup.URI != "" {
				fmt.Printf("Security Group:  %s\n", raw.Properties.SecurityGroup.URI)
			}
			if raw.Properties.BlockStorage.URI != "" {
				fmt.Printf("Block Storage:   %s\n", raw.Properties.BlockStorage.URI)
			}
			if raw.Properties.BillingPlan != nil && raw.Properties.BillingPlan.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", string(*raw.Properties.BillingPlan.BillingPeriod))
			}
			if raw.Properties.AdminUser != nil {
				fmt.Printf("Admin User:      %s\n", raw.Properties.AdminUser.Username)
			}
			if raw.Properties.ConcurrentUsers != nil {
				fmt.Printf("Concurrent Users: %s\n", *raw.Properties.ConcurrentUsers)
			}
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", string(*raw.Status.State))
			}
			if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
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
			fmt.Println()
		} else {
			fmt.Println("Container registry not found")
		}
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
		list, err := client.FromContainer().ContainerRegistry().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing container registries: %w", err)
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 30},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, cr := range list.Items() {
				raw := cr.Raw()
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
			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No container registries found")
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
		registryURI := "/projects/" + projectID + "/providers/Aruba.Container/registries/" + registryID
		registry, err := client.FromContainer().ContainerRegistry().Get(ctx, aruba.URI(registryURI))
		if err != nil {
			return fmt.Errorf("fetching current container registry: %w", apiErrFromV2(err))
		}
		if registry == nil || registry.Raw() == nil {
			return fmt.Errorf("container registry not found")
		}

		if name != "" {
			registry.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			registry.ReplaceTags(tags...)
		}
		if billingPeriod != "" {
			registry.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if concurrentUsers != "" {
			registry.OfSize(aruba.ContainerRegistrySizeFlavor(concurrentUsers))
		}

		updated, err := client.FromContainer().ContainerRegistry().Update(ctx, registry)
		if err != nil {
			return fmt.Errorf("updating container registry: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("Container registry", registryID))
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
			}
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", string(*raw.Status.State))
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

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
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
		registryURI := "/projects/" + projectID + "/providers/Aruba.Container/registries/" + registryID

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromContainer().ContainerRegistry().Get(ctx, aruba.URI(registryURI))
			if err != nil {
				return fmt.Errorf("dry-run: container registry not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("container registry", registryID))
			return nil
		}

		err = client.FromContainer().ContainerRegistry().Delete(ctx, aruba.URI(registryURI))
		if err != nil {
			return fmt.Errorf("deleting container registry: %w", err)
		}

		fmt.Println(msgDeleted("Container registry", registryID))
		return nil
	},
}
