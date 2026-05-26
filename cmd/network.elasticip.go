package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	// ElasticIP
	networkCmd.AddCommand(elasticipCmd)

	elasticipCmd.AddCommand(elasticipCreateCmd)
	elasticipCmd.AddCommand(elasticipGetCmd)
	elasticipCmd.AddCommand(elasticipUpdateCmd)
	elasticipCmd.AddCommand(elasticipDeleteCmd)
	elasticipCmd.AddCommand(elasticipListCmd)

	// Add flags for Elastic IP commands
	elasticipCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipCreateCmd.Flags().String("name", "", "Name for the Elastic IP")
	elasticipCreateCmd.Flags().String("region", "", "Region code (e.g., IT-BG)")
	elasticipCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	elasticipCreateCmd.MarkFlagRequired("name")
	elasticipCreateCmd.MarkFlagRequired("region")
	elasticipCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	elasticipGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipUpdateCmd.Flags().String("name", "", "New name for the Elastic IP")
	elasticipUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	elasticipDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	elasticipDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	elasticipListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	elasticipListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	elasticipGetCmd.ValidArgsFunction = completeElasticIPID
	elasticipUpdateCmd.ValidArgsFunction = completeElasticIPID
	elasticipDeleteCmd.ValidArgsFunction = completeElasticIPID
}

func completeElasticIPID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().ElasticIPs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, eip := range list.Items() {
			raw := eip.Raw()
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

// ElasticIP subcommands
var elasticipCmd = &cobra.Command{
	Use:   "elasticip",
	Short: "Manage Elastic IPs",
	Long:  `Perform CRUD operations on Elastic IPs in Aruba Cloud.`,
}

var elasticipCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Elastic IP",
	Long: `Create a new static public IP address (Elastic IP) in the specified region.

The IP can be attached to a cloud server or VPN tunnel after creation.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud network elasticip create --name my-eip --region IT-BG
  acloud network elasticip create --name prod-eip --region IT-BG --billing-period Month`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		eip := aruba.NewElasticIP().
			IntoProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod)).
			ReplaceTags(tags...)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromNetwork().ElasticIPs().Create(ctx, eip)
		if err != nil {
			return fmt.Errorf("creating Elastic IP: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			fmt.Printf("\n%s\n", msgCreated("Elastic IP", name))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
			}
			if raw.Properties.Address != nil {
				fmt.Printf("Address: %s\n", *raw.Properties.Address)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
			}
		} else {
			fmt.Println(msgCreatedAsync("Elastic IP", name))
		}
		return nil
	},
}

var elasticipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Elastic IPs",
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
		list, err := client.FromNetwork().ElasticIPs().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing Elastic IPs: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "ADDRESS", Width: 16},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, eip := range list.Items() {
				raw := eip.Raw()
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
				address := ""
				if raw.Properties.Address != nil {
					address = *raw.Properties.Address
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, region, address, status})
			}

			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No Elastic IPs found")
		}
		return nil
	},
}

var elasticipGetCmd = &cobra.Command{
	Use:   "get <elastic-ip-id>",
	Short: "Get Elastic IP details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eipID := args[0]

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
		eip, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(projectID, eipID))
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
		}

		if eip != nil && eip.Raw() != nil {
			raw := eip.Raw()

			fmt.Println("\nElastic IP Details:")
			fmt.Println("===================")

			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Metadata.LocationResponse != nil && raw.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", raw.Metadata.LocationResponse.Value)
			}
			if raw.Properties.Address != nil {
				fmt.Printf("Address:         %s\n", *raw.Properties.Address)
			}
			if raw.Properties.BillingPlan != nil && raw.Properties.BillingPlan.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", *raw.Properties.BillingPlan.BillingPeriod)
			}
			fmt.Printf("Linked Resources: %d\n", len(raw.Properties.LinkedResources))
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
		}
		return nil
	},
}

var elasticipUpdateCmd = &cobra.Command{
	Use:   "update <elastic-ip-id>",
	Short: "Update an Elastic IP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eipID := args[0]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
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
		eip, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(projectID, eipID))
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
		}

		if eip == nil || eip.Raw() == nil {
			return fmt.Errorf("Elastic IP not found")
		}

		if eip.Raw().Status.State != nil && *eip.Raw().Status.State == StateInCreation {
			return fmt.Errorf("cannot update Elastic IP while it is in 'InCreation' state. Please wait until the Elastic IP is fully created")
		}

		if name != "" {
			eip.Named(name)
		}
		if len(tags) > 0 {
			eip.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().ElasticIPs().Update(ctx, eip)
		if err != nil {
			return fmt.Errorf("updating Elastic IP: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("Elastic IP", eipID))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("Elastic IP", eipID))
		}
		return nil
	},
}

var elasticipDeleteCmd = &cobra.Command{
	Use:   "delete <elastic-ip-id>",
	Short: "Delete an Elastic IP",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eipID := args[0]

		skipConfirm, _ := cmd.Flags().GetBool("yes")

		if !skipConfirm {
			ok, err := confirmDelete("Elastic IP", eipID)
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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(projectID, eipID))
			if err != nil {
				return fmt.Errorf("dry-run: elastic IP not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("elastic IP", eipID))
			return nil
		}

		err = client.FromNetwork().ElasticIPs().Delete(ctx, aruba.ElasticIPRef(projectID, eipID))
		if err != nil {
			return fmt.Errorf("deleting Elastic IP: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Elastic IP", eipID))
		return nil
	},
}
