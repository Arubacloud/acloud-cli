package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	managementCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectCreateCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectUpdateCmd)
	projectCmd.AddCommand(projectDeleteCmd)
	projectCmd.AddCommand(projectListCmd)

	// Add completion for project IDs
	projectGetCmd.ValidArgsFunction = completeProjectID
	projectUpdateCmd.ValidArgsFunction = completeProjectID
	projectDeleteCmd.ValidArgsFunction = completeProjectID

	// Add flags for project create command
	projectCreateCmd.Flags().String("name", "", "Name for the project (required)")
	projectCreateCmd.Flags().String("description", "", "Description for the project")
	projectCreateCmd.Flags().StringSlice("tags", []string{}, "Tags for the project (comma-separated)")
	projectCreateCmd.Flags().Bool("default", false, "Set as default project")
	projectCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")
	projectCreateCmd.MarkFlagRequired("name")

	// Add flags for project delete command
	projectDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	projectDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	// Add flags for project update command
	projectUpdateCmd.Flags().String("description", "", "New description for the project")
	projectUpdateCmd.Flags().StringSlice("tags", []string{}, "Tags for the project (comma-separated)")

	// Add flags for project list command
	projectListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	projectListCmd.Flags().Int32("offset", 0, "Number of results to skip")
}

// projectFromRaw re-parses the full types.ProjectResponse from a Project
// wrapper's raw HTTP body. The v0.2.0 *aruba.Project has only unexported fields
// (not JSON-marshalable) and no accessor for Properties.ResourcesNumber, so
// commands that render the RESOURCES column or emit -o json/yaml re-parse the
// wire shape here. Returns nil on empty/malformed body — mirrors the v0.1.x
// `response.Data == nil` async branch.
func projectFromRaw(p *aruba.Project) *types.ProjectResponse {
	if p == nil {
		return nil
	}
	_, body := p.RawHTTP()
	if len(body) == 0 {
		return nil
	}
	var pr types.ProjectResponse
	if err := json.Unmarshal(body, &pr); err != nil {
		return nil
	}
	return &pr
}

// projectListPayload extracts the typed *types.ProjectList from a List wrapper
// for -o json/yaml rendering; the List[*Project] wrapper is not JSON-marshalable.
func projectListPayload(l *aruba.List[*aruba.Project]) any {
	if r, ok := l.Raw().(*types.Response[types.ProjectList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// completeProjectID provides completion for project IDs
func completeProjectID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Allow completion even if args exist - user might be completing a partial ID

	// Get SDK client
	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// List projects
	ctx := context.Background()
	list, err := client.FromProject().List(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, p := range list.Items() {
			id := p.ID()
			name := p.Name()
			// Filter by partial input - use HasPrefix for more reliable matching
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				// Format: "id\tname" - the tab separates the completion from the description
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long:  `Perform CRUD operations on projects in Aruba Cloud.`,
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long: `Create a new project in your Aruba Cloud account.

Pass --default to make this project the active default for subsequent commands.
Use 'acloud context set-project' to switch the active project at any time.`,
	Example: `  acloud management project create --name my-project --description "Production workloads"
  acloud management project create --name dev-project --default`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		setDefault, _ := cmd.Flags().GetBool("default")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build the create request
		proj := aruba.NewProject().Named(name).ReplaceTags(tags...)
		if description != "" {
			proj.WithDescription(description)
		}
		if setDefault {
			proj.AsDefault()
		}

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating project with the following parameters:")
			fmt.Printf("  Name:        %s\n", name)
			if description != "" {
				fmt.Printf("  Description: %s\n", description)
			}
			fmt.Printf("  Default:     %t\n", setDefault)
			if len(tags) > 0 {
				fmt.Printf("  Tags:        %v\n", tags)
			}
			fmt.Println()
		}

		// Create the project using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromProject().Create(ctx, proj)
		if err != nil {
			return fmt.Errorf("creating project: %w", apiErrFromV2(err))
		}

		if created != nil && created.ID() != "" {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "DEFAULT", Width: 10},
			}
			defaultVal := "No"
			if created.IsDefault() {
				defaultVal = "Yes"
			}
			row := []string{created.ID(), created.Name(), defaultVal}
			PrintOutput(created.Raw(), headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Project", name))
		}
		return nil
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get [project-id]",
	Short: "Get project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get project details using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		p, err := client.FromProject().Get(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("getting project: %w", apiErrFromV2(err))
		}

		if p != nil && p.ID() != "" {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(p.Raw(), nil, nil)
				return nil
			}

			// Display project details
			fmt.Println("\nProject Details:")
			fmt.Println("================")

			fmt.Printf("ID:              %s\n", p.ID())
			fmt.Printf("Name:            %s\n", p.Name())

			if p.Description() != "" {
				fmt.Printf("Description:     %s\n", p.Description())
			}

			fmt.Printf("Default:         %t\n", p.IsDefault())

			raw := p.Raw()
			if raw != nil {
				if raw.CreationDate != nil && !raw.CreationDate.IsZero() {
					fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
				}
				if raw.CreatedBy != nil {
					fmt.Printf("Created By:      %s\n", *raw.CreatedBy)
				}
				if raw.UpdateDate != nil && !raw.UpdateDate.IsZero() {
					fmt.Printf("Update Date:     %s\n", raw.UpdateDate.Format(DateLayout))
				}
				if raw.UpdatedBy != nil {
					fmt.Printf("Updated By:      %s\n", *raw.UpdatedBy)
				}
			}

			tags := p.Tags()
			if len(tags) > 0 {
				fmt.Printf("Tags:            %v\n", tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			fmt.Println()
		} else {
			fmt.Println("Project not found")
		}
		return nil
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project-id]",
	Short: "Update a project (description and/or tags only)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Get flags
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one field must be provided
		if description == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --description or --tags must be provided")
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// First, get the current project details to preserve existing values.
		// projectWrapper calls Get and normalises any API error via apiErrFromV2.
		ctx, cancel := newCtx()
		defer cancel()
		p, err := client.FromProject().Get(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("fetching current project: %w", apiErrFromV2(err))
		}

		if p == nil || p.ID() == "" {
			return fmt.Errorf("project not found or no data returned")
		}

		// Apply updates
		if description != "" {
			p.WithDescription(description)
		}
		if cmd.Flags().Changed("tags") {
			p.ReplaceTags(tags...)
		}

		// Update the project using the SDK
		updated, err := client.FromProject().Update(ctx, p)
		if err != nil {
			return fmt.Errorf("updating project: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.ID() != "" {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "DEFAULT", Width: 10},
			}
			defaultVal := "No"
			if updated.IsDefault() {
				defaultVal = "Yes"
			}
			row := []string{updated.ID(), updated.Name(), defaultVal}
			PrintOutput(updated.Raw(), headers, [][]string{row})
		} else {
			fmt.Println(msgUpdatedAsync("Project", projectID))
		}
		return nil
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project-id]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID := args[0]

		// Get confirmation flag
		confirm, _ := cmd.Flags().GetBool("yes")

		// If not confirmed, ask for confirmation
		if !confirm {
			ok, err := confirmDelete("project", projectID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		projectRef := aruba.URI("/projects/" + projectID)

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromProject().Get(ctx, projectRef)
			if err != nil {
				return fmt.Errorf("dry-run: project not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("project", projectID))
			return nil
		}

		// Delete the project using the SDK
		err = client.FromProject().Delete(ctx, projectRef)
		if err != nil {
			return fmt.Errorf("deleting project: %w", apiErrFromV2(err))
		}

		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		result := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{projectID, status}
		PrintOutput(result, headers, [][]string{{projectID, status}})
		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// List projects using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromProject().List(ctx)
		if err != nil {
			return fmt.Errorf("listing projects: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 30},
				{Header: "CREATION DATE", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, p := range list.Items() {
				name := p.Name()
				id := p.ID()

				// Format creation date as dd-mm-yyyy
				creationDate := "N/A"
				if !p.CreatedAt().IsZero() {
					creationDate = p.CreatedAt().Format("02-01-2006")
				}

				rows = append(rows, []string{name, id, creationDate})
			}

			// Print the table
			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No projects found")
		}
		return nil
	},
}
