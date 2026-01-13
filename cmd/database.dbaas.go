package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// DBaaS commands
	databaseCmd.AddCommand(dbaasCmd)
	dbaasCmd.AddCommand(dbaasCreateCmd)
	dbaasCmd.AddCommand(dbaasGetCmd)
	dbaasCmd.AddCommand(dbaasUpdateCmd)
	dbaasCmd.AddCommand(dbaasDeleteCmd)
	dbaasCmd.AddCommand(dbaasListCmd)

	// Add flags for DBaaS commands
	dbaasCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasCreateCmd.Flags().String("name", "", "Name for the DBaaS instance (required)")
	dbaasCreateCmd.Flags().String("region", "", "Region code (required)")
	dbaasCreateCmd.Flags().String("engine-id", "", "Database engine ID (required)")
	dbaasCreateCmd.Flags().String("flavor", "", "DBaaS flavor name (required)")
	dbaasCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	dbaasCreateCmd.MarkFlagRequired("name")
	dbaasCreateCmd.MarkFlagRequired("region")
	dbaasCreateCmd.MarkFlagRequired("engine-id")
	dbaasCreateCmd.MarkFlagRequired("flavor")

	dbaasGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUpdateCmd.Flags().String("name", "", "New name for the DBaaS instance")
	dbaasUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	dbaasDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	dbaasListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	dbaasGetCmd.ValidArgsFunction = completeDBaaSID
	dbaasUpdateCmd.ValidArgsFunction = completeDBaaSID
	dbaasDeleteCmd.ValidArgsFunction = completeDBaaSID
}

// Completion functions for database resources
func completeDBaaSID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	// Note: This assumes FromDatabase().DBaaS().List() exists in the SDK
	// If not available, this will need to be updated when SDK methods are added
	response, err := client.FromDatabase().DBaaS().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, dbaas := range response.Data.Values {
			if dbaas.Metadata.ID != nil && dbaas.Metadata.Name != nil {
				id := *dbaas.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *dbaas.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// DBaaS subcommands
var dbaasCmd = &cobra.Command{
	Use:   "dbaas",
	Short: "Manage DBaaS resources",
	Long:  `Perform CRUD operations on DBaaS resources in Aruba Cloud.`,
}

var dbaasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new DBaaS instance",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		engineID, _ := cmd.Flags().GetString("engine-id")
		flavor, _ := cmd.Flags().GetString("flavor")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" || region == "" || engineID == "" || flavor == "" {
			fmt.Println("Error: --name, --region, --engine-id, and --flavor are required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build the create request
		createRequest := types.DBaaSRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.DBaaSPropertiesRequest{
				Engine: &types.DBaaSEngine{
					ID: &engineID,
				},
				Flavor: &types.DBaaSFlavor{
					Name: &flavor,
				},
			},
		}

		ctx := context.Background()
		response, err := client.FromDatabase().DBaaS().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating DBaaS instance: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create DBaaS instance - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "ENGINE", Width: 20},
				{Header: "VERSION", Width: 15},
				{Header: "FLAVOR", Width: 20},
				{Header: "REGION", Width: 20},
			}
			row := []string{
				func() string {
					if response.Data.Metadata.ID != nil {
						return *response.Data.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.Name != nil {
						return *response.Data.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if response.Data.Properties.Engine != nil && response.Data.Properties.Engine.Type != nil {
						return *response.Data.Properties.Engine.Type
					}
					return ""
				}(),
				func() string {
					if response.Data.Properties.Engine != nil && response.Data.Properties.Engine.Version != nil {
						return *response.Data.Properties.Engine.Version
					}
					return ""
				}(),
				func() string {
					if response.Data.Properties.Flavor != nil && response.Data.Properties.Flavor.Name != nil {
						return *response.Data.Properties.Flavor.Name
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.LocationResponse != nil {
						return response.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("DBaaS instance created, but no data returned.")
		}
	},
}

var dbaasGetCmd = &cobra.Command{
	Use:   "get [dbaas-id]",
	Short: "Get DBaaS instance details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]

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
		resp, err := client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			fmt.Printf("Error getting DBaaS instance: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get DBaaS instance - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			dbaas := resp.Data

			fmt.Println("\nDBaaS Instance Details:")
			fmt.Println("======================")

			if dbaas.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *dbaas.Metadata.ID)
			}
			if dbaas.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *dbaas.Metadata.URI)
			}
			if dbaas.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *dbaas.Metadata.Name)
			}
			if dbaas.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", dbaas.Metadata.LocationResponse.Value)
			}
			if dbaas.Properties.Engine != nil {
				if dbaas.Properties.Engine.Type != nil {
					fmt.Printf("Engine Type:     %s\n", *dbaas.Properties.Engine.Type)
				}
				if dbaas.Properties.Engine.Version != nil {
					fmt.Printf("Engine Version:  %s\n", *dbaas.Properties.Engine.Version)
				}
				if dbaas.Properties.Engine.Name != nil {
					fmt.Printf("Engine Name:    %s\n", *dbaas.Properties.Engine.Name)
				}
			}
			if dbaas.Properties.Flavor != nil && dbaas.Properties.Flavor.Name != nil {
				fmt.Printf("Flavor:         %s\n", *dbaas.Properties.Flavor.Name)
			}
			if dbaas.Status.State != nil {
				fmt.Printf("Status:          %s\n", *dbaas.Status.State)
			}
			if !dbaas.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", dbaas.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if dbaas.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *dbaas.Metadata.CreatedBy)
			}
			if len(dbaas.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", dbaas.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			fmt.Println()
		} else {
			fmt.Println("DBaaS instance not found")
		}
	},
}

var dbaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DBaaS instances",
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
		resp, err := client.FromDatabase().DBaaS().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing DBaaS instances: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list DBaaS instances - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "ENGINE", Width: 20},
				{Header: "VERSION", Width: 15},
				{Header: "FLAVOR", Width: 20},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, dbaas := range resp.Data.Values {
				row := []string{
					func() string {
						if dbaas.Metadata.Name != nil {
							return *dbaas.Metadata.Name
						}
						return ""
					}(),
					func() string {
						if dbaas.Metadata.ID != nil {
							return *dbaas.Metadata.ID
						}
						return ""
					}(),
					func() string {
						if dbaas.Properties.Engine != nil && dbaas.Properties.Engine.Type != nil {
							return *dbaas.Properties.Engine.Type
						}
						return ""
					}(),
					func() string {
						if dbaas.Properties.Engine != nil && dbaas.Properties.Engine.Version != nil {
							return *dbaas.Properties.Engine.Version
						}
						return ""
					}(),
					func() string {
						if dbaas.Properties.Flavor != nil && dbaas.Properties.Flavor.Name != nil {
							return *dbaas.Properties.Flavor.Name
						}
						return ""
					}(),
					func() string {
						if dbaas.Metadata.LocationResponse != nil {
							return dbaas.Metadata.LocationResponse.Value
						}
						return ""
					}(),
					func() string {
						if dbaas.Status.State != nil {
							return *dbaas.Status.State
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No DBaaS instances found")
		}
	},
}

var dbaasUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id]",
	Short: "Update a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		getResp, err := client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			fmt.Printf("Error getting DBaaS instance: %v\n", err)
			return
		}

		if getResp == nil || getResp.Data == nil {
			fmt.Println("DBaaS instance not found")
			return
		}

		current := getResp.Data

		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for DBaaS instance")
			return
		}

		// Build update request preserving current properties
		updateRequest := types.DBaaSRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *current.Metadata.Name,
					Tags: current.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.DBaaSPropertiesRequest{},
		}

		// Preserve current engine if it exists
		if current.Properties.Engine != nil {
			updateRequest.Properties.Engine = &types.DBaaSEngine{
				ID: current.Properties.Engine.ID,
			}
		}

		// Preserve current flavor if it exists
		if current.Properties.Flavor != nil {
			updateRequest.Properties.Flavor = &types.DBaaSFlavor{
				Name: current.Properties.Flavor.Name,
			}
		}

		if name != "" {
			updateRequest.Metadata.ResourceMetadataRequest.Name = name
		}

		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.ResourceMetadataRequest.Tags = tags
		}

		response, err := client.FromDatabase().DBaaS().Update(ctx, projectID, dbaasID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating DBaaS instance: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update DBaaS instance - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nDBaaS instance updated successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("Warning: Update may have succeeded but response is empty")
		}
	},
}

var dbaasDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id]",
	Short: "Delete a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			fmt.Printf("Are you sure you want to delete DBaaS instance %s? (yes/no): ", dbaasID)
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		_, err = client.FromDatabase().DBaaS().Delete(ctx, projectID, dbaasID, nil)
		if err != nil {
			fmt.Printf("Error deleting DBaaS instance: %v\n", err)
			return
		}

		fmt.Printf("\nDBaaS instance %s deleted successfully!\n", dbaasID)
	},
}
