package cmd

import (
	"context"
	"fmt"
	"strings"

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

	// Add flags for project update command
	projectUpdateCmd.Flags().String("description", "", "New description for the project")
	projectUpdateCmd.Flags().StringSlice("tags", []string{}, "Tags for the project (comma-separated)")
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
	response, err := client.FromProject().List(ctx, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, project := range response.Data.Values {
			if project.Metadata.ID != nil && project.Metadata.Name != nil {
				id := *project.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					// Format: "id\tname" - the tab separates the completion from the description
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *project.Metadata.Name))
				}
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
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		setDefault, _ := cmd.Flags().GetBool("default")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Name is required
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build the create request
		createRequest := types.ProjectRequest{
			Metadata: types.ResourceMetadataRequest{
				Name: name,
				Tags: tags,
			},
			Properties: types.ProjectPropertiesRequest{
				Default: setDefault,
			},
		}

		// Add description if provided
		if description != "" {
			createRequest.Properties.Description = &description
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
		ctx := context.Background()
		response, err := client.FromProject().Create(ctx, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating project: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create project - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "DEFAULT", Width: 10},
				{Header: "RESOURCES", Width: 12},
			}
			row := []string{
				func() string {
					if response.Data.Metadata.ID != nil {
						return *response.Data.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.Name != nil {
						return *response.Data.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if response.Data.Properties.Default {
						return "Yes"
					}
					return "No"
				}(),
				fmt.Sprintf("%d", response.Data.Properties.ResourcesNumber),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Project created, but no data returned.")
		}
	},
}

var projectGetCmd = &cobra.Command{
	Use:   "get [project-id]",
	Short: "Get project details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get project details using the SDK
		ctx := context.Background()
		response, err := client.FromProject().Get(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error getting project: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to get project - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			project := response.Data

			// Display project details
			fmt.Println("\nProject Details:")
			fmt.Println("================")

			if project.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *project.Metadata.ID)
			}

			if project.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *project.Metadata.Name)
			}

			if project.Properties.Description != nil {
				fmt.Printf("Description:     %s\n", *project.Properties.Description)
			}

			fmt.Printf("Default:         %t\n", project.Properties.Default)
			fmt.Printf("Resources:       %d\n", project.Properties.ResourcesNumber)

			if !project.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", project.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}

			if project.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *project.Metadata.CreatedBy)
			}

			if project.Metadata.UpdateDate != nil && !project.Metadata.UpdateDate.IsZero() {
				fmt.Printf("Update Date:     %s\n", project.Metadata.UpdateDate.Format("02-01-2006 15:04:05"))
			}

			if project.Metadata.UpdatedBy != nil {
				fmt.Printf("Updated By:      %s\n", *project.Metadata.UpdatedBy)
			}

			if len(project.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", project.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			fmt.Println()
		} else {
			fmt.Println("Project not found")
		}
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update [project-id]",
	Short: "Update a project (description and/or tags only)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]

		// Get flags
		description, _ := cmd.Flags().GetString("description")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one field must be provided
		if description == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --description or --tags must be provided")
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// First, get the current project details to preserve existing values
		ctx := context.Background()
		getResponse, err := client.FromProject().Get(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error fetching current project: %v\n", err)
			return
		}

		if getResponse != nil && getResponse.IsError() && getResponse.Error != nil {
			fmt.Printf("Failed to get project - Status: %d\n", getResponse.StatusCode)
			if getResponse.Error.Title != nil {
				fmt.Printf("Error: %s\n", *getResponse.Error.Title)
			}
			if getResponse.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *getResponse.Error.Detail)
			}
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Project not found or no data returned.")
			return
		}

		currentProject := getResponse.Data

		// Build the update request with current values as defaults
		updateRequest := types.ProjectRequest{
			Metadata: types.ResourceMetadataRequest{
				Name: *currentProject.Metadata.Name,
				Tags: currentProject.Metadata.Tags,
			},
			Properties: types.ProjectPropertiesRequest{
				Default: currentProject.Properties.Default,
			},
		}

		// Update description if provided
		if description != "" {
			updateRequest.Properties.Description = &description
		} else if currentProject.Properties.Description != nil {
			updateRequest.Properties.Description = currentProject.Properties.Description
		}

		// Update tags if provided
		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.Tags = tags
		}

		// Update the project using the SDK
		response, err := client.FromProject().Update(ctx, projectID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating project: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update project - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "DEFAULT", Width: 10},
				{Header: "RESOURCES", Width: 12},
			}
			row := []string{
				func() string {
					if response.Data.Metadata.ID != nil {
						return *response.Data.Metadata.ID
					}
					return ""
				}(),
				func() string {
					if response.Data.Metadata.Name != nil {
						return *response.Data.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if response.Data.Properties.Default {
						return "Yes"
					}
					return "No"
				}(),
				fmt.Sprintf("%d", response.Data.Properties.ResourcesNumber),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Printf("Project '%s' updated.\n", projectID)
		}
	},
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [project-id]",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]

		// Get confirmation flag
		confirm, _ := cmd.Flags().GetBool("yes")

		// If not confirmed, ask for confirmation
		if !confirm {
			fmt.Printf("Are you sure you want to delete project %s? This action cannot be undone.\n", projectID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Delete the project using the SDK
		ctx := context.Background()
		resp, err := client.FromProject().Delete(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error deleting project: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to delete project - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		PrintTable(headers, [][]string{{projectID, status}})
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Run: func(cmd *cobra.Command, args []string) {
		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// List projects using the SDK
		ctx := context.Background()
		response, err := client.FromProject().List(ctx, nil)
		if err != nil {
			fmt.Printf("Error listing projects: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to list projects - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 30},
				{Header: "CREATION DATE", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, project := range response.Data.Values {
				name := ""
				if project.Metadata.Name != nil && *project.Metadata.Name != "" {
					name = *project.Metadata.Name
				}

				id := ""
				if project.Metadata.ID != nil && *project.Metadata.ID != "" {
					id = *project.Metadata.ID
				}

				// Format creation date as dd-mm-yyyy
				creationDate := "N/A"
				if !project.Metadata.CreationDate.IsZero() {
					creationDate = project.Metadata.CreationDate.Format("02-01-2006")
				}

				rows = append(rows, []string{name, id, creationDate})
			}

			// Print the table
			PrintTable(headers, rows)
		} else {
			fmt.Println("No projects found")
		}
	},
}
