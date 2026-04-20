package cmd

import (
	"context"
	"fmt"
	"strings"

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

// Completion functions for network resources

func completeElasticIPID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Allow completion even if args exist - user might be completing a partial ID

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromNetwork().ElasticIPs().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, eip := range response.Data.Values {
			if eip.Metadata.ID != nil && eip.Metadata.Name != nil {
				id := *eip.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *eip.Metadata.Name))
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
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build the create request
		createRequest := types.ElasticIPRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.ElasticIPPropertiesRequest{
				BillingPlan: types.BillingPeriodResource{
					BillingPeriod: billingPeriod,
				},
			},
		}

		// Create the Elastic IP using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().ElasticIPs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating Elastic IP: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("Elastic IP", name))
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *response.Data.Metadata.Name)
			}
			if response.Data.Properties.Address != nil {
				fmt.Printf("Address: %s\n", *response.Data.Properties.Address)
			}
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", response.Data.Metadata.Tags)
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
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// List Elastic IPs using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().ElasticIPs().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing Elastic IPs: %w", err)
		}
		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "ADDRESS", Width: 16},
				{Header: "STATUS", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, eip := range response.Data.Values {
				name := ""
				if eip.Metadata.Name != nil && *eip.Metadata.Name != "" {
					name = *eip.Metadata.Name
				}

				id := ""
				if eip.Metadata.ID != nil {
					id = *eip.Metadata.ID
				}

				region := ""
				if eip.Metadata.LocationResponse != nil {
					region = eip.Metadata.LocationResponse.Value
				}

				address := ""
				if eip.Properties.Address != nil {
					address = *eip.Properties.Address
				}

				status := ""
				if eip.Status.State != nil {
					status = *eip.Status.State
				}

				rows = append(rows, []string{name, id, region, address, status})
			}

			// Print the table
			PrintTable(headers, rows)
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

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get Elastic IP details using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", err)
		}
		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			eip := response.Data

			// Display Elastic IP details
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

			fmt.Printf("Billing Period:  %s\n", eip.Properties.BillingPlan.BillingPeriod)
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

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one update flag must be provided
		if name == "" && len(tags) == 0 {
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// First, get the current Elastic IP to preserve existing properties
		ctx, cancel := newCtx()
		defer cancel()
		getResponse, err := client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", err)
		}

		if getResponse == nil || getResponse.Data == nil {
			return fmt.Errorf("Elastic IP not found")
		}

		// Check if Elastic IP is in InCreation state
		if getResponse.Data.Status.State != nil && *getResponse.Data.Status.State == StateInCreation {
			return fmt.Errorf("cannot update Elastic IP while it is in 'InCreation' state. Please wait until the Elastic IP is fully created")
		}

		// Get region value
		regionValue := ""
		if getResponse.Data.Metadata.LocationResponse != nil {
			regionValue = getResponse.Data.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for Elastic IP")
		}

		// Build the update request, preserving existing values
		updateRequest := types.ElasticIPRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *getResponse.Data.Metadata.Name,
					Tags: getResponse.Data.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.ElasticIPPropertiesRequest{
				BillingPlan: getResponse.Data.Properties.BillingPlan,
			},
		}

		// Apply updates
		if name != "" {
			updateRequest.Metadata.Name = name
		}
		if len(tags) > 0 {
			updateRequest.Metadata.Tags = tags
		}

		// Update the Elastic IP using the SDK
		response, err := client.FromNetwork().ElasticIPs().Update(ctx, projectID, eipID, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating Elastic IP: %w", err)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("Elastic IP", eipID))
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

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			ok, err := confirmDelete("Elastic IP", eipID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: elastic IP not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("elastic IP", eipID))
			return nil
		}

		// Delete the Elastic IP using the SDK
		deleteResp, err := client.FromNetwork().ElasticIPs().Delete(ctx, projectID, eipID, nil)
		if err != nil {
			return fmt.Errorf("deleting Elastic IP: %w", err)
		}
		if deleteResp != nil && deleteResp.IsError() && deleteResp.Error != nil {
			return fmtAPIError(deleteResp.StatusCode, deleteResp.Error.Title, deleteResp.Error.Detail)
		}

		fmt.Println(msgDeleted("Elastic IP", eipID))
		return nil
	},
}
