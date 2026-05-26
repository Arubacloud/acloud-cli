package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	dbaasCmd.AddCommand(dbaasDatabaseCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseCreateCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseGetCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseUpdateCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseDeleteCmd)
	dbaasDatabaseCmd.AddCommand(dbaasDatabaseListCmd)

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

	dbaasDatabaseGetCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseUpdateCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseDeleteCmd.ValidArgsFunction = completeDBaaSDatabaseID
}

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
	dbaasRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)

	ctx := context.Background()
	list, err := client.FromDatabase().Databases().List(ctx, dbaasRef)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, db := range list.Items() {
			raw := db.Raw()
			if raw != nil && raw.Name != "" {
				if toComplete == "" || strings.HasPrefix(raw.Name, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", raw.Name, raw.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

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

		dbaasRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)
		db := aruba.NewDatabase().IntoDBaaS(dbaasRef).Named(name)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Databases().Create(ctx, db)
		if err != nil {
			return fmt.Errorf("creating database: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			fmt.Printf("\n%s\n", msgCreated("Database", name))
			fmt.Printf("Name:            %s\n", raw.Name)
			if raw.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
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
		dbURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + databaseName
		db, err := client.FromDatabase().Databases().Get(ctx, aruba.URI(dbURI))
		if err != nil {
			return fmt.Errorf("getting database: %w", apiErrFromV2(err))
		}

		if db != nil && db.Raw() != nil {
			raw := db.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nDatabase Details:")
			fmt.Println("================")
			fmt.Printf("Name:            %s\n", raw.Name)
			if raw.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
			}
			if raw.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *raw.CreatedBy)
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

		dbaasRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromDatabase().Databases().List(ctx, dbaasRef)
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
				raw := db.Raw()
				if raw == nil {
					continue
				}
				creationDate := ""
				if raw.CreationDate != nil {
					creationDate = raw.CreationDate.Format(DateLayout)
				}
				createdBy := ""
				if raw.CreatedBy != nil {
					createdBy = *raw.CreatedBy
				}
				rows = append(rows, []string{raw.Name, creationDate, createdBy})
			}
			PrintOutput(list.Raw(), headers, rows)
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

		dbURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + databaseName

		ctx, cancel := newCtx()
		defer cancel()
		db, err := client.FromDatabase().Databases().Get(ctx, aruba.URI(dbURI))
		if err != nil {
			return fmt.Errorf("getting database: %w", apiErrFromV2(err))
		}
		if db == nil || db.Raw() == nil {
			return fmt.Errorf("database not found")
		}

		db.Named(name)

		updated, err := client.FromDatabase().Databases().Update(ctx, db)
		if err != nil {
			return fmt.Errorf("updating database: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			fmt.Printf("\n%s\n", msgUpdated("Database", databaseName))
			fmt.Printf("Name:            %s\n", updated.Raw().Name)
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

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			ok, err := confirmDelete(fmt.Sprintf("database '%s' in DBaaS instance", databaseName), dbaasID)
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

		dbURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + databaseName

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromDatabase().Databases().Get(ctx, aruba.URI(dbURI))
			if err != nil {
				return fmt.Errorf("dry-run: database not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("database", databaseName))
			return nil
		}

		err = client.FromDatabase().Databases().Delete(ctx, aruba.URI(dbURI))
		if err != nil {
			return fmt.Errorf("deleting database: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Database", databaseName))
		return nil
	},
}
