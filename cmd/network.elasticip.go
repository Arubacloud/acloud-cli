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
	elasticipCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	elasticipGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipUpdateCmd.Flags().String("name", "", "New name for the Elastic IP")
	elasticipUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	elasticipDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	elasticipDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	elasticipListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

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
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")

		// Validate required fields
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}
		if region == "" {
			fmt.Println("Error: --region is required")
			return
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
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
		ctx := context.Background()
		response, err := client.FromNetwork().ElasticIPs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating Elastic IP: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create Elastic IP - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nElastic IP created successfully!")
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
			fmt.Println("Elastic IP creation initiated. Use 'list' or 'get' to check status.")
		}
	},
}

var elasticipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Elastic IPs",
	Run: func(cmd *cobra.Command, args []string) {
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// List Elastic IPs using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().ElasticIPs().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing Elastic IPs: %v\n", err)
			return
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

				region := eip.Metadata.LocationResponse.Value

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
	},
}

var elasticipGetCmd = &cobra.Command{
	Use:   "get <elastic-ip-id>",
	Short: "Get Elastic IP details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eipID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get Elastic IP details using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
		if err != nil {
			fmt.Printf("Error getting Elastic IP details: %v\n", err)
			return
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
				fmt.Printf("Creation Date:   %s\n", eip.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}

var elasticipUpdateCmd = &cobra.Command{
	Use:   "update <elastic-ip-id>",
	Short: "Update an Elastic IP",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eipID := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one update flag must be provided
		if name == "" && len(tags) == 0 {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// First, get the current Elastic IP to preserve existing properties
		ctx := context.Background()
		getResponse, err := client.FromNetwork().ElasticIPs().Get(ctx, projectID, eipID, nil)
		if err != nil {
			fmt.Printf("Error getting Elastic IP details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Error: Elastic IP not found")
			return
		}

		// Check if Elastic IP is in InCreation state
		if getResponse.Data.Status.State != nil && *getResponse.Data.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update Elastic IP while it is in 'InCreation' state. Please wait until the Elastic IP is fully created.")
			return
		}

		// Get region value
		regionValue := ""
		if getResponse.Data.Metadata.LocationResponse != nil {
			regionValue = getResponse.Data.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for Elastic IP")
			return
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
			fmt.Printf("Error updating Elastic IP: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nElastic IP updated successfully!")
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
			fmt.Printf("\nElastic IP %s update completed.\n", eipID)
		}
	},
}

var elasticipDeleteCmd = &cobra.Command{
	Use:   "delete <elastic-ip-id>",
	Short: "Delete an Elastic IP",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		eipID := args[0]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete Elastic IP %s? This action cannot be undone.\n", eipID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Delete the Elastic IP using the SDK
		ctx := context.Background()
		_, err = client.FromNetwork().ElasticIPs().Delete(ctx, projectID, eipID, nil)
		if err != nil {
			fmt.Printf("Error deleting Elastic IP: %v\n", err)
			return
		}

		fmt.Printf("\nElastic IP %s deleted successfully!\n", eipID)
	},
}
