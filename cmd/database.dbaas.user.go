package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
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
	dbaasUserListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasUserListCmd.Flags().Int("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	dbaasUserGetCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserUpdateCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserDeleteCmd.ValidArgsFunction = completeDBaaSUserID
}

// File-local Ref helpers

func userRef(projectID, dbaasID, username string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + username)
}

func userFromRaw(u *aruba.User) *types.UserResponse {
	if u == nil {
		return nil
	}
	return u.Raw()
}

func userListPayload(l *aruba.List[*aruba.User]) any {
	if r, ok := l.Raw().(*types.Response[types.UserList]); ok && r != nil {
		return r.Data
	}
	return nil
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
	list, err := client.FromDatabase().Users().List(ctx, dbaasRef(projectID, dbaasID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, u := range list.Items() {
			username := u.Username()
			if toComplete == "" || strings.HasPrefix(username, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", username, username))
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
	Args:    cobra.ExactArgs(1),
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

		// API expects the password base64-encoded (matches Terraform provider behaviour).
		encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

		u := aruba.NewUser().
			IntoDBaaS(dbaasRef(projectID, dbaasID)).
			WithUsername(username).
			WithPassword(encodedPassword)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Users().Create(ctx, u)
		if err != nil {
			return fmt.Errorf("creating user: %w", apiErrFromV2(err))
		}

		resource := userFromRaw(created)
		if resource != nil {
			fmt.Printf("\n%s\n", msgCreated("User", username))
			fmt.Printf("Username:        %s\n", resource.Username)
			if resource.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", resource.CreationDate.Format(DateLayout))
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
		got, err := client.FromDatabase().Users().Get(ctx, userRef(projectID, dbaasID, username))
		if err != nil {
			return fmt.Errorf("getting user: %w", apiErrFromV2(err))
		}

		resource := userFromRaw(got)
		if resource != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(resource, nil, nil)
				return nil
			}

			fmt.Println("\nUser Details:")
			fmt.Println("=============")

			fmt.Printf("Username:        %s\n", resource.Username)
			if resource.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", resource.CreationDate.Format(DateLayout))
			}
			if resource.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *resource.CreatedBy)
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
		list, err := client.FromDatabase().Users().List(ctx, dbaasRef(projectID, dbaasID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing users: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "USERNAME", Width: 40},
				{Header: "CREATION DATE", Width: 25},
				{Header: "CREATED BY", Width: 30},
			}

			var rows [][]string
			for _, u := range list.Items() {
				row := []string{
					u.Username(),
					func() string {
						t := u.CreatedAt()
						if t.IsZero() {
							return ""
						}
						return t.Format(DateLayout)
					}(),
					u.CreatedBy(),
				}
				rows = append(rows, row)
			}
			PrintOutput(userListPayload(list), headers, rows)
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

		ctx, cancel := newCtx()
		defer cancel()
		current, err := client.FromDatabase().Users().Get(ctx, userRef(projectID, dbaasID, username))
		if err != nil {
			return fmt.Errorf("fetching current user: %w", apiErrFromV2(err))
		}

		// API expects the password base64-encoded (matches Terraform provider behaviour).
		current.WithPassword(base64.StdEncoding.EncodeToString([]byte(password)))

		updated, err := client.FromDatabase().Users().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating user: %w", apiErrFromV2(err))
		}

		resource := userFromRaw(updated)
		if resource != nil {
			fmt.Printf("\n%s\n", msgUpdated("User", username))
			fmt.Printf("Username:        %s\n", resource.Username)
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
			if _, err := client.FromDatabase().Users().Get(ctx, userRef(projectID, dbaasID, username)); err != nil {
				return fmt.Errorf("dry-run: database user not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("database user", username))
			return nil
		}

		if !skipConfirm {
			ok, err := confirmDelete(fmt.Sprintf("user '%s' in DBaaS instance", username), dbaasID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromDatabase().Users().Delete(ctx, userRef(projectID, dbaasID, username)); err != nil {
			return fmt.Errorf("deleting user: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("User", username))
		return nil
	},
}
