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

	dbaasDatabaseListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		createRequest := types.DatabaseRequest{
			Name: name,
		}

		ctx := context.Background()
		response, err := client.FromDatabase().Databases().Create(ctx, projectID, dbaasID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating database: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create database - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nDatabase created successfully!")
			fmt.Printf("Name:            %s\n", response.Data.Name)
			if response.Data.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", response.Data.CreationDate.Format("02-01-2006 15:04:05"))
			}
		} else {
			fmt.Println("Database created, but no data returned.")
		}
	},
}

var dbaasDatabaseGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [database-name]",
	Short: "Get database details",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]
		databaseName := args[1]

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
		resp, err := client.FromDatabase().Databases().Get(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			fmt.Printf("Error getting database: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get database - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			db := resp.Data

			fmt.Println("\nDatabase Details:")
			fmt.Println("================")

			fmt.Printf("Name:            %s\n", db.Name)
			if db.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", db.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if db.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *db.CreatedBy)
			}
			fmt.Println()
		} else {
			fmt.Println("Database not found")
		}
	},
}

var dbaasDatabaseListCmd = &cobra.Command{
	Use:   "list [dbaas-id]",
	Short: "List all databases in DBaaS",
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
		resp, err := client.FromDatabase().Databases().List(ctx, projectID, dbaasID, nil)
		if err != nil {
			fmt.Printf("Error listing databases: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list databases - Status: %d\n", resp.StatusCode)
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
							return db.CreationDate.Format("02-01-2006 15:04:05")
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
	},
}

var dbaasDatabaseUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id] [database-name]",
	Short: "Update a database",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]
		databaseName := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		updateRequest := types.DatabaseRequest{
			Name: name,
		}

		ctx := context.Background()
		response, err := client.FromDatabase().Databases().Update(ctx, projectID, dbaasID, databaseName, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating database: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update database - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nDatabase updated successfully!")
			fmt.Printf("Name:            %s\n", response.Data.Name)
		} else {
			fmt.Println("Warning: Update may have succeeded but response is empty")
		}
	},
}

var dbaasDatabaseDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [database-name]",
	Short: "Delete a database",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]
		databaseName := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			fmt.Printf("Are you sure you want to delete database '%s' in DBaaS instance %s? (yes/no): ", databaseName, dbaasID)
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
		_, err = client.FromDatabase().Databases().Delete(ctx, projectID, dbaasID, databaseName, nil)
		if err != nil {
			fmt.Printf("Error deleting database: %v\n", err)
			return
		}

		fmt.Printf("\nDatabase '%s' deleted successfully!\n", databaseName)
	},
}
