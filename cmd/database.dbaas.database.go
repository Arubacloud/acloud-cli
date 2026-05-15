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
	// DBaaS database commands
	dbaasCmd.AddCommand(dbaasDatabaseCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseCreateCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseGetCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseUpdateCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseDeleteCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseListCmd)

	// Add flags for DBaaS database commands
	dbaasDatabaseCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseCreateCmd.Flags().String("name", "", "Database name (required)")
	dbaasDatabaseCreateCmd.MarkFlagRequired("name")

	dbaasDatabaseGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasDatabaseUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseUpdateCmd.Flags().String("name", "", "New database name")

	dbaasDatabaseDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasDatabaseDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasDatabaseListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDatabaseListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasDatabaseListCmd.Flags().Int("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	dbaasDatabaseGetCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseUpdateCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseDeleteCmd.ValidArgsFunction = completeDBaaSDatabaseID
}

// File-local Ref helpers

func databaseRef(projectID, dbaasID, name string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + name)
}

func databaseFromRaw(d *aruba.Database) *types.DatabaseResponse {
	if d == nil {
		return nil
	}
	return d.Raw()
}

func databaseListPayload(l *aruba.List[*aruba.Database]) any {
	if r, ok := l.Raw().(*types.Response[types.DatabaseList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// Completion functions for database resources
func completeDBaaSDatabaseID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	dbaasID := args[0]

	ctx := context.Background()
	list, err := client.FromDatabase().Databases().List(ctx, dbaasRef(projectID, dbaasID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, db := range list.Items() {
			name := db.DatabaseID()
			if toComplete == "" || strings.HasPrefix(name, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", name, name))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// DBaaS database subcommands
var dbaasDatabaseCmd = &cobra.Command{
	Use:   "database [dbaas-id]",
	Short: "Manage databases in DBaaS",
	Long:  `Perform CRUD operations on databases in DBaaS.`,
}

var dbaasDatabaseCreateCmd = &cobra.Command{
	Use:   "create [dbaas-id]",
	Short: "Create a new database in DBaaS",
	Long: `Create a new database schema inside an existing DBaaS instance.

The DBaaS instance must already exist and be in a ready state.
Use 'acloud database dbaas get <dbaas-id>' to check its status.`,
	Example: `  acloud database dbaas database create <dbaas-id> --name myapp_db`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		db := aruba.NewDatabase().
			IntoDBaaS(dbaasRef(projectID, dbaasID)).
			Named(name)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Databases().Create(ctx, db)
		if err != nil {
			return fmt.Errorf("creating database: %w", apiErrFromV2(err))
		}

		resource := databaseFromRaw(created)
		if resource != nil {
			fmt.Printf("\n%s\n", msgCreated("Database", name))
			fmt.Printf("Name:            %s\n", resource.Name)
			if resource.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", resource.CreationDate.Format(DateLayout))
			}
		} else {
			fmt.Println(msgCreatedAsync("Database", name))
		}
		return nil
	},
}

var dbaasDatabaseGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [database-name]",
	Short: "Get database details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]
		databaseName := args[1]

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
		got, err := client.FromDatabase().Databases().Get(ctx, databaseRef(projectID, dbaasID, databaseName))
		if err != nil {
			return fmt.Errorf("getting database: %w", apiErrFromV2(err))
		}

		resource := databaseFromRaw(got)
		if resource != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(resource, nil, nil)
				return nil
			}

			fmt.Println("\nDatabase Details:")
			fmt.Println("================")

			fmt.Printf("Name:            %s\n", resource.Name)
			if resource.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", resource.CreationDate.Format(DateLayout))
			}
			if resource.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *resource.CreatedBy)
			}
			fmt.Println()
		} else {
			fmt.Println("Database not found")
		}
		return nil
	},
}

var dbaasDatabaseListCmd = &cobra.Command{
	Use:   "list [dbaas-id]",
	Short: "List all databases in DBaaS",
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
		list, err := client.FromDatabase().Databases().List(ctx, dbaasRef(projectID, dbaasID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing databases: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "CREATION DATE", Width: 25},
				{Header: "CREATED BY", Width: 30},
			}

			var rows [][]string
			for _, db := range list.Items() {
				row := []string{
					db.Name(),
					func() string {
						t := db.CreatedAt()
						if t.IsZero() {
							return ""
						}
						return t.Format(DateLayout)
					}(),
					db.CreatedBy(),
				}
				rows = append(rows, row)
			}
			PrintOutput(databaseListPayload(list), headers, rows)
		} else {
			fmt.Println("No databases found")
		}
		return nil
	},
}

var dbaasDatabaseUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id] [database-name]",
	Short: "Update a database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]
		databaseName := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		current, err := client.FromDatabase().Databases().Get(ctx, databaseRef(projectID, dbaasID, databaseName))
		if err != nil {
			return fmt.Errorf("fetching current database: %w", apiErrFromV2(err))
		}

		current.Named(name)

		updated, err := client.FromDatabase().Databases().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating database: %w", apiErrFromV2(err))
		}

		resource := databaseFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("Database", databaseName))
			fmt.Printf("Name:            %s\n", resource.Name)
		} else {
			fmt.Println(msgUpdatedAsync("Database", databaseName))
		}
		return nil
	},
}

var dbaasDatabaseDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [database-name]",
	Short: "Delete a database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]
		databaseName := args[1]

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
			if _, err := client.FromDatabase().Databases().Get(ctx, databaseRef(projectID, dbaasID, databaseName)); err != nil {
				return fmt.Errorf("dry-run: database not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("database", databaseName))
			return nil
		}

		if !skipConfirm {
			ok, err := confirmDelete(fmt.Sprintf("database '%s' in DBaaS instance", databaseName), dbaasID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromDatabase().Databases().Delete(ctx, databaseRef(projectID, dbaasID, databaseName)); err != nil {
			return fmt.Errorf("deleting database: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Database", databaseName))
		return nil
	},
}
