package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// DBaaS user commands
	dbaasCmd.AddCommand(dbaasUserCmd)
	dbaasUserCmd.AddCommand(dbaasUserCreateCmd)
	dbaasUserCmd.AddCommand(dbaasUserGetCmd)
	dbaasUserCmd.AddCommand(dbaasUserUpdateCmd)
	dbaasUserCmd.AddCommand(dbaasUserDeleteCmd)
	dbaasUserCmd.AddCommand(dbaasUserListCmd)

	// Add flags for DBaaS user commands
	dbaasUserCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserCreateCmd.Flags().String("username", "", "Username (required)")
	dbaasUserCreateCmd.Flags().String("password", "", "Password (required)")
	dbaasUserCreateCmd.MarkFlagRequired("username")
	dbaasUserCreateCmd.MarkFlagRequired("password")

	dbaasUserGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasUserUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserUpdateCmd.Flags().String("password", "", "New password (required)")

	dbaasUserDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	dbaasUserListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	dbaasUserGetCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserUpdateCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserDeleteCmd.ValidArgsFunction = completeDBaaSUserID
}

// Completion functions for database resources
func completeDBaaSUserID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	response, err := client.FromDatabase().Users().List(ctx, projectID, dbaasID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, user := range response.Data.Values {
			if user.Username != "" {
				if toComplete == "" || strings.HasPrefix(user.Username, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", user.Username, user.Username))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// DBaaS user subcommands
var dbaasUserCmd = &cobra.Command{
	Use:   "user [dbaas-id]",
	Short: "Manage users in DBaaS",
	Long:  `Perform CRUD operations on users in DBaaS.`,
}

var dbaasUserCreateCmd = &cobra.Command{
	Use:   "create [dbaas-id]",
	Short: "Create a new user in DBaaS",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if username == "" || password == "" {
			fmt.Println("Error: --username and --password are required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		createRequest := types.UserRequest{
			Username: username,
			Password: password,
		}

		ctx := context.Background()
		response, err := client.FromDatabase().Users().Create(ctx, projectID, dbaasID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating user: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create user - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nUser created successfully!")
			fmt.Printf("Username:        %s\n", response.Data.Username)
			if response.Data.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", response.Data.CreationDate.Format("02-01-2006 15:04:05"))
			}
		} else {
			fmt.Println("User created, but no data returned.")
		}
	},
}

var dbaasUserGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [username]",
	Short: "Get user details",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]
		username := args[1]

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
		resp, err := client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
		if err != nil {
			fmt.Printf("Error getting user: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get user - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			user := resp.Data

			fmt.Println("\nUser Details:")
			fmt.Println("=============")

			fmt.Printf("Username:        %s\n", user.Username)
			if user.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", user.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if user.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *user.CreatedBy)
			}
			fmt.Println()
		} else {
			fmt.Println("User not found")
		}
	},
}

var dbaasUserListCmd = &cobra.Command{
	Use:   "list [dbaas-id]",
	Short: "List all users in DBaaS",
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
		resp, err := client.FromDatabase().Users().List(ctx, projectID, dbaasID, nil)
		if err != nil {
			fmt.Printf("Error listing users: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list users - Status: %d\n", resp.StatusCode)
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
				{Header: "USERNAME", Width: 40},
				{Header: "CREATION DATE", Width: 25},
				{Header: "CREATED BY", Width: 30},
			}

			var rows [][]string
			for _, user := range resp.Data.Values {
				row := []string{
					user.Username,
					func() string {
						if user.CreationDate != nil {
							return user.CreationDate.Format("02-01-2006 15:04:05")
						}
						return ""
					}(),
					func() string {
						if user.CreatedBy != nil {
							return *user.CreatedBy
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No users found")
		}
	},
}

var dbaasUserUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id] [username]",
	Short: "Update a user (change password)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]
		username := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		password, _ := cmd.Flags().GetString("password")

		if password == "" {
			fmt.Println("Error: --password is required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		updateRequest := types.UserRequest{
			Username: username,
			Password: password,
		}

		ctx := context.Background()
		response, err := client.FromDatabase().Users().Update(ctx, projectID, dbaasID, username, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating user: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update user - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nUser updated successfully!")
			fmt.Printf("Username:        %s\n", response.Data.Username)
		} else {
			fmt.Println("Warning: Update may have succeeded but response is empty")
		}
	},
}

var dbaasUserDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [username]",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbaasID := args[0]
		username := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			fmt.Printf("Are you sure you want to delete user '%s' in DBaaS instance %s? (yes/no): ", username, dbaasID)
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
		_, err = client.FromDatabase().Users().Delete(ctx, projectID, dbaasID, username, nil)
		if err != nil {
			fmt.Printf("Error deleting user: %v\n", err)
			return
		}

		fmt.Printf("\nUser '%s' deleted successfully!\n", username)
	},
}
