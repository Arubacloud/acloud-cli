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
	dbaasUserDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasUserListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasUserListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a new database user inside an existing DBaaS instance.

The user is granted access to the instance. Assign database-level privileges
separately through your database client after the user is created.`,
	Example: `  acloud database dbaas user create <dbaas-id> --username myuser --password mypassword`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		createRequest := types.UserRequest{
			Username: username,
			Password: password,
		}

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromDatabase().Users().Create(ctx, projectID, dbaasID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("User", username))
			fmt.Printf("Username:        %s\n", response.Data.Username)
			if response.Data.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", response.Data.CreationDate.Format(DateLayout))
			}
		} else {
			fmt.Println(msgCreatedAsync("User", username))
		}
		return nil
	},
}

var dbaasUserGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [username]",
	Short: "Get user details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]
		username := args[1]

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
		resp, err := client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil {
			user := resp.Data

			fmt.Println("\nUser Details:")
			fmt.Println("=============")

			fmt.Printf("Username:        %s\n", user.Username)
			if user.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", user.CreationDate.Format(DateLayout))
			}
			if user.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *user.CreatedBy)
			}
			fmt.Println()
		} else {
			fmt.Println("User not found")
		}
		return nil
	},
}

var dbaasUserListCmd = &cobra.Command{
	Use:   "list [dbaas-id]",
	Short: "List all users in DBaaS",
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
		resp, err := client.FromDatabase().Users().List(ctx, projectID, dbaasID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing users: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
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
							return user.CreationDate.Format(DateLayout)
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
		return nil
	},
}

var dbaasUserUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id] [username]",
	Short: "Update a user (change password)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]
		username := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		password, _ := cmd.Flags().GetString("password")

		if password == "" {
			return fmt.Errorf("--password is required")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		updateRequest := types.UserRequest{
			Username: username,
			Password: password,
		}

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromDatabase().Users().Update(ctx, projectID, dbaasID, username, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating user: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("User", username))
			fmt.Printf("Username:        %s\n", response.Data.Username)
		} else {
			fmt.Println(msgUpdatedAsync("User", username))
		}
		return nil
	},
}

var dbaasUserDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [username]",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]
		username := args[1]

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			ok, err := confirmDelete(fmt.Sprintf("user '%s' in DBaaS instance", username), dbaasID)
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
			_, err = client.FromDatabase().Users().Get(ctx, projectID, dbaasID, username, nil)
			if err != nil {
				return fmt.Errorf("dry-run: database user not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("database user", username))
			return nil
		}

		_, err = client.FromDatabase().Users().Delete(ctx, projectID, dbaasID, username, nil)
		if err != nil {
			return fmt.Errorf("deleting user: %w", err)
		}

		fmt.Println(msgDeleted("User", username))
		return nil
	},
}
