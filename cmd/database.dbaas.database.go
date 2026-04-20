package cmd

import (
	"context"
	"fmt"
	"strings"

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
	dbaasDatabaseListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasDatabaseListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	dbaasDatabaseGetCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseUpdateCmd.ValidArgsFunction = completeDBaaSDatabaseID
	dbaasDatabaseDeleteCmd.ValidArgsFunction = completeDBaaSDatabaseID
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
	response, err := client.FromDatabase().Databases().List(ctx, projectID, dbaasID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, db := range response.Data.Values {
			if db.Name != "" {
				if toComplete == "" || strings.HasPrefix(db.Name, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", db.Name, db.Name))
				}
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

		createRequest := types.DatabaseRequest{
			Name: name,
		}

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromDatabase().Databases().Create(ctx, projectID, dbaasID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating database: %w", err)
		}

		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("Database", name))
			fmt.Printf("Name:            %s\n", response.Data.Name)
			if response.Data.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", response.Data.CreationDate.Format(DateLayout))
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
		resp, err := client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			return fmt.Errorf("getting database: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil {
			db := resp.Data

			fmt.Println("\nDatabase Details:")
			fmt.Println("================")

			fmt.Printf("Name:            %s\n", db.Name)
			if db.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", db.CreationDate.Format(DateLayout))
			}
			if db.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *db.CreatedBy)
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
		resp, err := client.FromDatabase().Databases().List(ctx, projectID, dbaasID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing databases: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "CREATION DATE", Width: 25},
				{Header: "CREATED BY", Width: 30},
			}

			var rows [][]string
			for _, db := range resp.Data.Values {
				row := []string{
					db.Name,
					func() string {
						if db.CreationDate != nil {
							return db.CreationDate.Format(DateLayout)
						}
						return ""
					}(),
					func() string {
						if db.CreatedBy != nil {
							return *db.CreatedBy
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintTable(headers, rows)
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

		updateRequest := types.DatabaseRequest{
			Name: name,
		}

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromDatabase().Databases().Update(ctx, projectID, dbaasID, databaseName, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating database: %w", err)
		}

		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("Database", databaseName))
			fmt.Printf("Name:            %s\n", response.Data.Name)
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

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
			if err != nil {
				return fmt.Errorf("dry-run: database not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("database", databaseName))
			return nil
		}

		deleteResp, err := client.FromDatabase().Databases().Delete(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			return fmt.Errorf("deleting database: %w", err)
		}
		if deleteResp != nil && deleteResp.IsError() {
			return apiErrFromResp(deleteResp.StatusCode, deleteResp.Error)
		}

		fmt.Println(msgDeleted("Database", databaseName))
		return nil
	},
}
