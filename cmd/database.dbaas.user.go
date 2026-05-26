package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	dbaasCmd.AddCommand(dbaasUserCmd)
	dbaasUserCmd.AddCommand(dbaasUserCreateCmd)
	dbaasUserCmd.AddCommand(dbaasUserGetCmd)
	dbaasUserCmd.AddCommand(dbaasUserUpdateCmd)
	dbaasUserCmd.AddCommand(dbaasUserDeleteCmd)
	dbaasUserCmd.AddCommand(dbaasUserListCmd)

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

	dbaasUserGetCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserUpdateCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserDeleteCmd.ValidArgsFunction = completeDBaaSUserID
}

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
	dbaasRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)

	ctx := context.Background()
	list, err := client.FromDatabase().Users().List(ctx, dbaasRef)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, user := range list.Items() {
			raw := user.Raw()
			if raw != nil && raw.Username != "" {
				if toComplete == "" || strings.HasPrefix(raw.Username, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", raw.Username, raw.Username))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

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

		dbaasRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)
		user := aruba.NewUser().InDBaaS(dbaasRef).WithUsername(username).WithPassword(password)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Users().Create(ctx, user)
		if err != nil {
			return fmt.Errorf("creating user: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			fmt.Printf("\n%s\n", msgCreated("User", username))
			fmt.Printf("Username:        %s\n", raw.Username)
			if raw.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
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
		userURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + username
		u, err := client.FromDatabase().Users().Get(ctx, aruba.URI(userURI))
		if err != nil {
			return fmt.Errorf("getting user: %w", apiErrFromV2(err))
		}

		if u != nil && u.Raw() != nil {
			raw := u.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nUser Details:")
			fmt.Println("=============")
			fmt.Printf("Username:        %s\n", raw.Username)
			if raw.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
			}
			if raw.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *raw.CreatedBy)
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

		dbaasRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID)

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromDatabase().Users().List(ctx, dbaasRef)
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
			for _, user := range list.Items() {
				raw := user.Raw()
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
				rows = append(rows, []string{raw.Username, creationDate, createdBy})
			}
			PrintOutput(list, headers, rows)
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

		userURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + username

		ctx, cancel := newCtx()
		defer cancel()
		u, err := client.FromDatabase().Users().Get(ctx, aruba.URI(userURI))
		if err != nil {
			return fmt.Errorf("getting user: %w", apiErrFromV2(err))
		}
		if u == nil || u.Raw() == nil {
			return fmt.Errorf("user not found")
		}

		u.WithPassword(password)

		updated, err := client.FromDatabase().Users().Update(ctx, u)
		if err != nil {
			return fmt.Errorf("updating user: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			fmt.Printf("\n%s\n", msgUpdated("User", username))
			fmt.Printf("Username:        %s\n", updated.Raw().Username)
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

		userURI := "/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/users/" + username

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromDatabase().Users().Get(ctx, aruba.URI(userURI))
			if err != nil {
				return fmt.Errorf("dry-run: database user not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("database user", username))
			return nil
		}

		err = client.FromDatabase().Users().Delete(ctx, aruba.URI(userURI))
		if err != nil {
			return fmt.Errorf("deleting user: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("User", username))
		return nil
	},
}
