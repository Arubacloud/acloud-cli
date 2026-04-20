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
	vpcCreateCmd.MarkFlagRequired("name")
	vpcCreateCmd.MarkFlagRequired("region")
	vpcGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("name", "", "New name for the VPC")
	vpcUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpcDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	vpcListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpcListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a new Virtual Private Cloud (VPC) in the specified region.

A VPC is the top-level network boundary. Subnets, security groups, and other
network resources are created within a VPC.`,
	Example: `  acloud network vpc create --name my-vpc --region IT-BG
  acloud network vpc create --name prod-vpc --region IT-BG --tags env=prod,team=infra`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")

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
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().VPCs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating VPC: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("VPC", name))
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
			fmt.Println(msgCreatedAsync("VPC", name))
		}
		return nil
	},
}

var vpcGetCmd = &cobra.Command{
	Use:   "get <vpc-id>",
	Short: "Get VPC details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]

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

		// Get VPC details using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if err != nil {
			return fmt.Errorf("getting VPC details: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
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
				fmt.Printf("Creation Date:   %s\n", vpc.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var vpcUpdateCmd = &cobra.Command{
	Use:   "update <vpc-id>",
	Short: "Update a VPC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]

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

		// First, get the current VPC to preserve existing properties
		ctx, cancel := newCtx()
		defer cancel()
		getResponse, err := client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
		if err != nil {
			return fmt.Errorf("getting VPC details: %w", err)
		}

		if getResponse == nil || getResponse.Data == nil {
			return fmt.Errorf("VPC not found")
		}

		// Check if VPC is in InCreation state
		if getResponse.Data.Status.State != nil && *getResponse.Data.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC while it is in 'InCreation' state. Please wait until the VPC is fully created")
		}

		// Get region value
		regionValue := ""
		if getResponse.Data.Metadata.LocationResponse != nil {
			regionValue = getResponse.Data.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for VPC")
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
			return fmt.Errorf("updating VPC: %w", err)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("VPC", vpcID))
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
			fmt.Println(msgUpdatedAsync("VPC", vpcID))
		}
		return nil
	},
}

var vpcDeleteCmd = &cobra.Command{
	Use:   "delete <vpc-id>",
	Short: "Delete a VPC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			ok, err := confirmDelete("VPC", vpcID)
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
			_, err = client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: VPC not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("VPC", vpcID))
			return nil
		}

		// Delete the VPC using the SDK
		deleteResp, err := client.FromNetwork().VPCs().Delete(ctx, projectID, vpcID, nil)
		if err != nil {
			return fmt.Errorf("deleting VPC: %w", err)
		}
		if deleteResp != nil && deleteResp.IsError() && deleteResp.Error != nil {
			return fmtAPIError(deleteResp.StatusCode, deleteResp.Error.Title, deleteResp.Error.Detail)
		}

		fmt.Println(msgDeleted("VPC", vpcID))
		return nil
	},
}

var vpcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPCs",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get projectID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// List VPCs using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().VPCs().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing VPCs: %w", err)
		}
		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
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

				region := ""
				if vpc.Metadata.LocationResponse != nil {
					region = vpc.Metadata.LocationResponse.Value
				}

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
		return nil
	},
}
