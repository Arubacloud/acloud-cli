package cmd

import (
	"context"
	"fmt"
	"strings"

	aruba "github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// databaseRef returns a Ref for a specific database inside a DBaaS instance.
func databaseRef(projectID, dbaasID, dbName string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Database/dbaas/" + dbaasID +
		"/databases/" + dbName)
}

func init() {
	dbaasCmd.AddCommand(dbaasGrantCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantCreateCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantListCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantGetCmd)
	dbaasGrantCmd.AddCommand(dbaasGrantDeleteCmd)

	dbaasGrantCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasGrantCreateCmd.Flags().String("username", "", "Username to grant access to (required)")
	dbaasGrantCreateCmd.Flags().String("role", "", "Role to grant (e.g. liteadmin, readonly, readwrite) (required)")
	dbaasGrantCreateCmd.MarkFlagRequired("username")
	dbaasGrantCreateCmd.MarkFlagRequired("role")

	dbaasGrantListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasGrantListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasGrantListCmd.Flags().Int("offset", 0, "Number of results to skip")

	dbaasGrantGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasGrantDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasGrantDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasGrantDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
}

func grantRef(projectID, dbaasID, dbName, grantID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Database/dbaas/" + dbaasID + "/databases/" + dbName + "/grants/" + grantID)
}


func completeGrantID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
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
	dbaasID, dbName := args[0], args[1]
	ctx := context.Background()
	list, err := client.FromDatabase().Grants().List(ctx, databaseRef(projectID, dbaasID, dbName))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	var completions []string
	if list != nil {
		for _, g := range list.Items() {
			id := g.ID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s→%s", id, g.Username(), g.RoleName()))
			}
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}

var dbaasGrantCmd = &cobra.Command{
	Use:   "grant [dbaas-id] [database-name]",
	Short: "Manage grants on a database",
	Long:  `Assign or revoke user access roles on a database inside a DBaaS instance.`,
}

var dbaasGrantCreateCmd = &cobra.Command{
	Use:   "create [dbaas-id] [database-name]",
	Short: "Grant a user access to a database",
	Long: `Grant a user a role on a specific database inside a DBaaS instance.

Common roles: liteadmin, readonly, readwrite.`,
	Example: `  acloud database dbaas grant create <dbaas-id> mydb --username myuser --role liteadmin`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID, dbName := args[0], args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		username, _ := cmd.Flags().GetString("username")
		role, _ := cmd.Flags().GetString("role")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		g := aruba.NewGrant().
			InDatabase(databaseRef(projectID, dbaasID, dbName)).
			ForUser(username).
			OfRole(role)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().Grants().Create(ctx, g)
		if err != nil {
			return fmt.Errorf("creating grant: %w", apiErrFromV2(err))
		}

		if created != nil {
			fmt.Printf("\n%s\n", msgCreated("Grant", username))
			fmt.Printf("Username:        %s\n", created.Username())
			fmt.Printf("Role:            %s\n", created.RoleName())
			fmt.Printf("Database:        %s\n", created.DatabaseName())
		} else {
			fmt.Printf("\n%s\n", msgCreatedAsync("Grant", username))
		}
		return nil
	},
}

var dbaasGrantListCmd = &cobra.Command{
	Use:   "list [dbaas-id] [database-name]",
	Short: "List all grants on a database",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID, dbName := args[0], args[1]

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
		list, err := client.FromDatabase().Grants().List(ctx, databaseRef(projectID, dbaasID, dbName), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing grants: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "GRANT ID", Width: 30},
				{Header: "USERNAME", Width: 30},
				{Header: "ROLE", Width: 20},
				{Header: "DATABASE", Width: 25},
			}
			var rows [][]string
			for _, g := range list.Items() {
				row := []string{
					g.ID(),
					g.Username(),
					g.RoleName(),
					g.DatabaseName(),
				}
				rows = append(rows, row)
			}
			PrintOutput(nil, headers, rows)
		} else {
			fmt.Println("No grants found")
		}
		return nil
	},
}

var dbaasGrantGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [database-name] [grant-id]",
	Short: "Get grant details",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID, dbName, grantID := args[0], args[1], args[2]

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
		got, err := client.FromDatabase().Grants().Get(ctx, grantRef(projectID, dbaasID, dbName, grantID))
		if err != nil {
			return fmt.Errorf("getting grant: %w", apiErrFromV2(err))
		}

		if got != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(got, nil, nil)
				return nil
			}
			fmt.Println("\nGrant Details:")
			fmt.Println("==============")
			fmt.Printf("Username:        %s\n", got.Username())
			fmt.Printf("Role:            %s\n", got.RoleName())
			fmt.Printf("Database:        %s\n", got.DatabaseName())
			if t := got.CreatedAt(); !t.IsZero() {
				fmt.Printf("Creation Date:   %s\n", t.Format(DateLayout))
			}
			if s := got.CreatedBy(); s != "" {
				fmt.Printf("Created By:      %s\n", s)
			}
			fmt.Println()
		} else {
			fmt.Println("Grant not found")
		}
		return nil
	},
}

var dbaasGrantDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [database-name] [grant-id]",
	Short: "Delete a grant",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID, dbName, grantID := args[0], args[1], args[2]

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
			if _, err := client.FromDatabase().Grants().Get(ctx, grantRef(projectID, dbaasID, dbName, grantID)); err != nil {
				return fmt.Errorf("dry-run: grant not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("grant", grantID))
			return nil
		}

		if !skipConfirm {
			ok, err := confirmDelete(fmt.Sprintf("grant '%s' on database '%s'", grantID, dbName), dbaasID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromDatabase().Grants().Delete(ctx, grantRef(projectID, dbaasID, dbName, grantID)); err != nil {
			return fmt.Errorf("deleting grant: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Grant", grantID))
		return nil
	},
}
