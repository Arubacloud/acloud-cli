package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	// VPC
	networkCmd.AddCommand(vpcCmd)
	vpcCmd.AddCommand(vpcCreateCmd)
	vpcCmd.AddCommand(vpcGetCmd)
	vpcCmd.AddCommand(vpcUpdateCmd)
	vpcCmd.AddCommand(vpcDeleteCmd)
	vpcCmd.AddCommand(vpcListCmd)

	// Add flags for VPC commands
	vpcCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcCreateCmd.Flags().String("name", "", "Name for the VPC")
	vpcCreateCmd.Flags().String("region", "", "Region code (e.g., IT-BG)")
	vpcCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("name", "", "New name for the VPC")
	vpcUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpcDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	vpcGetCmd.ValidArgsFunction = completeVPCID
	vpcUpdateCmd.ValidArgsFunction = completeVPCID
	vpcDeleteCmd.ValidArgsFunction = completeVPCID
}

// Completion functions for network resources

func completeVPCID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	response, err := client.FromNetwork().VPCs().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, vpc := range response.Data.Values {
			if vpc.Metadata.ID != nil && vpc.Metadata.Name != nil {
				id := *vpc.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *vpc.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// VPC subcommands
var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Manage VPCs",
	Long:  `Perform CRUD operations on VPCs in Aruba Cloud.`,
}

var vpcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VPC",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")

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

		// Build the create request (default and preset are always false)
		setDefault := false
		setPreset := false
		createRequest := types.VPCRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.VPCPropertiesRequest{
				Properties: &types.VPCProperties{
					Default: &setDefault,
					Preset:  &setPreset,
				},
			},
		}

		// Create the VPC using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPCs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating VPC: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create VPC - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nVPC created successfully!")
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *response.Data.Metadata.Name)
			}
			fmt.Printf("Default: %t\n", response.Data.Properties.Default)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("VPC creation initiated. Use 'list' or 'get' to check status.")
		}
	},
}

var vpcGetCmd = &cobra.Command{
	Use:   "get <vpc-id>",
	Short: "Get VPC details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]

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

		// Get VPC details using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if err != nil {
			fmt.Printf("Error getting VPC details: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to get VPC - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			vpc := response.Data

			// Display VPC details
			fmt.Println("\nVPC Details:")
			fmt.Println("============")

			if vpc.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *vpc.Metadata.ID)
			}
			if vpc.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *vpc.Metadata.URI)
			}
			if vpc.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *vpc.Metadata.Name)
			}
			if vpc.Metadata.LocationResponse != nil && vpc.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", vpc.Metadata.LocationResponse.Value)
			}
			fmt.Printf("Default:         %t\n", vpc.Properties.Default)
			fmt.Printf("Linked Resources: %d\n", len(vpc.Properties.LinkedResources))

			if vpc.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", vpc.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if vpc.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *vpc.Metadata.CreatedBy)
			}

			if len(vpc.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", vpc.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			if vpc.Status.State != nil {
				fmt.Printf("Status:          %s\n", *vpc.Status.State)
			}
		} else {
			fmt.Println("VPC not found or no data returned.")
		}
	},
}

var vpcUpdateCmd = &cobra.Command{
	Use:   "update <vpc-id>",
	Short: "Update a VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]

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

		// First, get the current VPC to preserve existing properties
		ctx := context.Background()
		getResponse, err := client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if err != nil {
			fmt.Printf("Error getting VPC details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Error: VPC not found")
			return
		}

		// Check if VPC is in InCreation state
		if getResponse.Data.Status.State != nil && *getResponse.Data.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update VPC while it is in 'InCreation' state. Please wait until the VPC is fully created.")
			return
		}

		// Get region value
		regionValue := ""
		if getResponse.Data.Metadata.LocationResponse != nil {
			regionValue = getResponse.Data.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for VPC")
			return
		}

		// Build the update request, preserving existing values
		updateRequest := types.VPCRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *getResponse.Data.Metadata.Name,
					Tags: getResponse.Data.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.VPCPropertiesRequest{
				Properties: &types.VPCProperties{
					Default: &getResponse.Data.Properties.Default,
				},
			},
		}

		// Apply updates
		if name != "" {
			updateRequest.Metadata.Name = name
		}
		if len(tags) > 0 {
			updateRequest.Metadata.Tags = tags
		}

		// Update the VPC using the SDK
		response, err := client.FromNetwork().VPCs().Update(ctx, projectID, vpcID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating VPC: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nVPC updated successfully!")
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
			fmt.Printf("\nVPC %s update completed.\n", vpcID)
		}
	},
}

var vpcDeleteCmd = &cobra.Command{
	Use:   "delete <vpc-id>",
	Short: "Delete a VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete VPC %s? This action cannot be undone.\n", vpcID)
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

		// Delete the VPC using the SDK
		ctx := context.Background()
		_, err = client.FromNetwork().VPCs().Delete(ctx, projectID, vpcID, nil)
		if err != nil {
			fmt.Printf("Error deleting VPC: %v\n", err)
			return
		}

		fmt.Printf("\nVPC %s deleted successfully!\n", vpcID)
	},
}

var vpcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPCs",
	Run: func(cmd *cobra.Command, args []string) {
		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get projectID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// List VPCs using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPCs().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing VPCs: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 25},
				{Header: "REGION", Width: 18},
				{Header: "SUBNETS", Width: 10},
				{Header: "STATUS", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, vpc := range response.Data.Values {
				name := ""
				if vpc.Metadata.Name != nil && *vpc.Metadata.Name != "" {
					name = *vpc.Metadata.Name
				}

				id := ""
				if vpc.Metadata.ID != nil {
					id = *vpc.Metadata.ID
				}

				region := vpc.Metadata.LocationResponse.Value

				subnets := fmt.Sprintf("%d", len(vpc.Properties.LinkedResources))

				status := ""
				if vpc.Status.State != nil {
					status = *vpc.Status.State
				}

				rows = append(rows, []string{name, id, region, subnets, status})
			}

			// Print the table
			PrintTable(headers, rows)
		} else {
			fmt.Println("No VPCs found")
		}
	},
}
