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
	dbaasDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a new managed database instance in the specified region.

Use --engine-id to select the database engine (e.g., MySQL, PostgreSQL).
Use --flavor to select the compute profile (CPU/RAM).

After creation, add databases with 'acloud database dbaas database create'
and users with 'acloud database dbaas user create'.`,
	Example: `  acloud database dbaas create \
    --name my-db --region IT-BG \
    --engine-id <engine-id> \
    --flavor <flavor-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		engineID, _ := cmd.Flags().GetString("engine-id")
		flavor, _ := cmd.Flags().GetString("flavor")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
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

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromDatabase().DBaaS().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating DBaaS instance: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
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
			PrintOutput(response.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("DBaaS instance", name))
		}
		return nil
	},
}

var dbaasGetCmd = &cobra.Command{
	Use:   "get [dbaas-id]",
	Short: "Get DBaaS instance details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

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
		resp, err := client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			return fmt.Errorf("getting DBaaS instance: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil {
			dbaas := resp.Data

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(dbaas, nil, nil)
				return nil
			}

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
				if dbaas.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", dbaas.Metadata.LocationResponse.Value)
				}
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
			if dbaas.Metadata.CreationDate != nil && !dbaas.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", dbaas.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var dbaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DBaaS instances",
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
		resp, err := client.FromDatabase().DBaaS().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing DBaaS instances: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
			PrintOutput(resp.Data, headers, rows)
		} else {
			fmt.Println("No DBaaS instances found")
		}
		return nil
	},
}

var dbaasUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id]",
	Short: "Update a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		getResp, err := client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
		if err != nil {
			return fmt.Errorf("getting DBaaS instance: %w", err)
		}

		if getResp == nil || getResp.Data == nil {
			return fmt.Errorf("DBaaS instance not found")
		}

		current := getResp.Data

		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for DBaaS instance")
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
			return fmt.Errorf("updating DBaaS instance: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("DBaaS instance", dbaasID))
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("DBaaS instance", dbaasID))
		}
		return nil
	},
}

var dbaasDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id]",
	Short: "Delete a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			ok, err := confirmDelete("DBaaS instance", dbaasID)
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
			_, err = client.FromDatabase().DBaaS().Get(ctx, projectID, dbaasID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: DBaaS instance not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("DBaaS instance", dbaasID))
			return nil
		}

		_, err = client.FromDatabase().DBaaS().Delete(ctx, projectID, dbaasID, nil)
		if err != nil {
			return fmt.Errorf("deleting DBaaS instance: %w", err)
		}

		fmt.Println(msgDeleted("DBaaS instance", dbaasID))
		return nil
	},
}
