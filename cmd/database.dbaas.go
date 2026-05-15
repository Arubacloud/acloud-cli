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
	dbaasListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasListCmd.Flags().Int("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	dbaasGetCmd.ValidArgsFunction = completeDBaaSID
	dbaasUpdateCmd.ValidArgsFunction = completeDBaaSID
	dbaasDeleteCmd.ValidArgsFunction = completeDBaaSID
}

// File-local Ref helpers

func dbaasRef(projectID, dbaasID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)
}

func dbaasFromRaw(d *aruba.DBaaS) *types.DBaaSResponse {
	if d == nil {
		return nil
	}
	return d.Raw()
}

func dbaasListPayload(l *aruba.List[*aruba.DBaaS]) any {
	if r, ok := l.Raw().(*types.Response[types.DBaaSList]); ok && r != nil {
		return r.Data
	}
	return nil
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
	list, err := client.FromDatabase().DBaaS().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, r := range list.Items() {
			id := r.DBaaSID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, r.Name()))
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

		d := aruba.NewDBaaS().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfEngine(aruba.DatabaseEngine(engineID)).
			OfFlavor(aruba.DBaaSFlavor(flavor))
		if len(tags) > 0 {
			d.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().DBaaS().Create(ctx, d)
		if err != nil {
			return fmt.Errorf("creating DBaaS instance: %w", apiErrFromV2(err))
		}

		resource := dbaasFromRaw(created)
		if resource != nil {
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
					if resource.Metadata.ID != nil {
						return *resource.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if resource.Metadata.Name != nil {
						return *resource.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if resource.Properties.Engine != nil && resource.Properties.Engine.Type != nil {
						return *resource.Properties.Engine.Type
					}
					return ""
				}(),
				func() string {
					if resource.Properties.Engine != nil && resource.Properties.Engine.Version != nil {
						return *resource.Properties.Engine.Version
					}
					return ""
				}(),
				func() string {
					if resource.Properties.Flavor != nil && resource.Properties.Flavor.Name != nil {
						return *resource.Properties.Flavor.Name
					}
					return ""
				}(),
				func() string {
					if resource.Metadata.LocationResponse != nil {
						return string(resource.Metadata.LocationResponse.Value)
					}
					return ""
				}(),
			}
			PrintOutput(resource, headers, [][]string{row})
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
		got, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(projectID, dbaasID))
		if err != nil {
			return fmt.Errorf("getting DBaaS instance: %w", apiErrFromV2(err))
		}

		resource := dbaasFromRaw(got)
		if resource != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(resource, nil, nil)
				return nil
			}

			fmt.Println("\nDBaaS Instance Details:")
			fmt.Println("======================")

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
			if resource.Properties.Engine != nil {
				if resource.Properties.Engine.Type != nil {
					fmt.Printf("Engine Type:     %s\n", *resource.Properties.Engine.Type)
				}
				if resource.Properties.Engine.Version != nil {
					fmt.Printf("Engine Version:  %s\n", *resource.Properties.Engine.Version)
				}
				if resource.Properties.Engine.Name != nil {
					fmt.Printf("Engine Name:    %s\n", *resource.Properties.Engine.Name)
				}
			}
			if resource.Properties.Flavor != nil && resource.Properties.Flavor.Name != nil {
				fmt.Printf("Flavor:         %s\n", *resource.Properties.Flavor.Name)
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
		list, err := client.FromDatabase().DBaaS().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing DBaaS instances: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
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
			for _, d := range list.Items() {
				raw := dbaasFromRaw(d)
				if raw == nil {
					continue
				}
				row := []string{
					func() string {
						if raw.Metadata.Name != nil {
							return *raw.Metadata.Name
						}
						return ""
					}(),
					func() string {
						if raw.Metadata.ID != nil {
							return *raw.Metadata.ID
						}
						return ""
					}(),
					func() string {
						if raw.Properties.Engine != nil && raw.Properties.Engine.Type != nil {
							return *raw.Properties.Engine.Type
						}
						return ""
					}(),
					func() string {
						if raw.Properties.Engine != nil && raw.Properties.Engine.Version != nil {
							return *raw.Properties.Engine.Version
						}
						return ""
					}(),
					func() string {
						if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
							return *raw.Properties.Flavor.Name
						}
						return ""
					}(),
					func() string {
						if raw.Metadata.LocationResponse != nil {
							return string(raw.Metadata.LocationResponse.Value)
						}
						return ""
					}(),
					func() string {
						if raw.Status.State != nil {
							return *raw.Status.State
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintOutput(dbaasListPayload(list), headers, rows)
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
		current, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(projectID, dbaasID))
		if err != nil {
			return fmt.Errorf("fetching current DBaaS instance: %w", apiErrFromV2(err))
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}

		updated, err := client.FromDatabase().DBaaS().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating DBaaS instance: %w", apiErrFromV2(err))
		}

		resource := dbaasFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("DBaaS instance", dbaasID))
			if resource.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *resource.Metadata.ID)
			}
			if resource.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *resource.Metadata.Name)
			}
			if len(resource.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", resource.Metadata.Tags)
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")

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
			if _, err := client.FromDatabase().DBaaS().Get(ctx, dbaasRef(projectID, dbaasID)); err != nil {
				return fmt.Errorf("dry-run: DBaaS instance not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("DBaaS instance", dbaasID))
			return nil
		}

		if !skipConfirm {
			ok, err := confirmDelete("DBaaS instance", dbaasID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromDatabase().DBaaS().Delete(ctx, dbaasRef(projectID, dbaasID)); err != nil {
			return fmt.Errorf("deleting DBaaS instance: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("DBaaS instance", dbaasID))
		return nil
	},
}
