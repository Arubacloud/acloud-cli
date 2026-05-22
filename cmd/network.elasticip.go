package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
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

func elasticIPListPayload(l *aruba.List[*aruba.ElasticIP]) any {
	if r, ok := l.Raw().(*types.Response[types.ElasticList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// Completion functions for network resources

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
	list, err := client.FromNetwork().ElasticIPs().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, eip := range list.Items() {
			id := eip.ElasticIPID()
			name := eip.Name()
			if id == "" {
				continue
			}
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
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
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		if len(tags) > 0 {
			eip.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromNetwork().ElasticIPs().Create(ctx, eip)
		if err != nil {
			return fmt.Errorf("creating Elastic IP: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil {
			fmt.Printf("\n%s\n", msgCreated("Elastic IP", name))
			if r.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *r.Metadata.ID)
			}
			if r.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *r.Metadata.Name)
			}
			if r.Properties.Address != nil {
				fmt.Printf("Address: %s\n", *r.Properties.Address)
			}
			if len(r.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", r.Metadata.Tags)
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
		list, err := client.FromNetwork().ElasticIPs().List(ctx, projectRef(projectID), listOpts(cmd)...)
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
				r := eip.Raw()
				name := eip.Name()
				id := eip.ElasticIPID()

				region := ""
				address := ""
				status := ""
				if r != nil {
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					if r.Properties.Address != nil {
						address = *r.Properties.Address
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}

				rows = append(rows, []string{name, id, region, address, status})
			}

			PrintOutput(elasticIPListPayload(list), headers, rows)
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
		got, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(projectID, eipID))
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
		}

		eip := got.Raw()
		if eip != nil {
			fmt.Println("\nElastic IP Details:")
			fmt.Println("===================")

			if eip.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *eip.Metadata.ID)
			}
			if eip.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *eip.Metadata.URI)
			}
			if eip.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *eip.Metadata.Name)
			}
			if eip.Metadata.LocationResponse != nil && eip.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", eip.Metadata.LocationResponse.Value)
			}
			if eip.Properties.Address != nil {
				fmt.Printf("Address:         %s\n", *eip.Properties.Address)
			}

			if eip.Properties.BillingPeriod != nil {
				fmt.Printf("Billing Period:  %s\n", *eip.Properties.BillingPeriod)
			}
			fmt.Printf("Linked Resources: %d\n", len(eip.Properties.LinkedResources))

			if eip.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", eip.Metadata.CreationDate.Format(DateLayout))
			}
			if eip.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *eip.Metadata.CreatedBy)
			}

			if len(eip.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", eip.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			if eip.Status.State != nil {
				fmt.Printf("Status:          %s\n", *eip.Status.State)
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
		cur, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.ElasticIPRef(projectID, eipID))
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
		}

		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("Elastic IP not found")
		}

		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update Elastic IP while it is in 'InCreation' state. Please wait until the Elastic IP is fully created")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().ElasticIPs().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating Elastic IP: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			fmt.Printf("\n%s\n", msgUpdated("Elastic IP", eipID))
			if ur.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *ur.Metadata.ID)
			}
			if ur.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *ur.Metadata.Name)
			}
			if len(ur.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", ur.Metadata.Tags)
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

		if err := client.FromNetwork().ElasticIPs().Delete(ctx, aruba.ElasticIPRef(projectID, eipID)); err != nil {
			return fmt.Errorf("deleting Elastic IP: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Elastic IP", eipID))
		return nil
	},
}
