package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	// Job commands
	scheduleCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(jobCreateCmd)
	jobCmd.AddCommand(jobGetCmd)
	jobCmd.AddCommand(jobUpdateCmd)
	jobCmd.AddCommand(jobDeleteCmd)
	jobCmd.AddCommand(jobListCmd)

	// Add flags for job commands
	jobCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	jobCreateCmd.Flags().String("name", "", "Job name (required)")
	jobCreateCmd.Flags().String("region", "", "Region code (required)")
	jobCreateCmd.Flags().String("job-type", "", "Job type: OneShot or Recurring (required)")
	jobCreateCmd.Flags().String("schedule-at", "", "Date and time when the job should run (required for OneShot)")
	jobCreateCmd.Flags().String("cron", "", "CRON expression (required for Recurring)")
	jobCreateCmd.Flags().String("execute-until", "", "End date until which the job can run (required for Recurring)")
	jobCreateCmd.Flags().Bool("enabled", true, "Enable the job (default: true)")
	jobCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	jobCreateCmd.MarkFlagRequired("name")
	jobCreateCmd.MarkFlagRequired("region")
	jobCreateCmd.MarkFlagRequired("job-type")

	jobGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	jobUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	jobUpdateCmd.Flags().String("name", "", "New job name")
	jobUpdateCmd.Flags().Bool("enabled", false, "Enable/disable the job")
	jobUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	jobDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	jobDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	jobListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	jobGetCmd.ValidArgsFunction = completeJobID
	jobUpdateCmd.ValidArgsFunction = completeJobID
	jobDeleteCmd.ValidArgsFunction = completeJobID
}

// Completion functions for schedule resources
func completeJobID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromSchedule().Jobs().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, job := range response.Data.Values {
			if job.Metadata.ID != nil && job.Metadata.Name != nil {
				id := *job.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *job.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Job subcommands
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage scheduled jobs",
	Long:  `Perform CRUD operations on scheduled jobs in Aruba Cloud.`,
}

var jobCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new scheduled job",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		jobType, _ := cmd.Flags().GetString("job-type")
		scheduleAt, _ := cmd.Flags().GetString("schedule-at")
		cron, _ := cmd.Flags().GetString("cron")
		executeUntil, _ := cmd.Flags().GetString("execute-until")
		enabled, _ := cmd.Flags().GetBool("enabled")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" || region == "" || jobType == "" {
			fmt.Println("Error: --name, --region, and --job-type are required")
			return
		}

		// Validate job type
		if jobType != "OneShot" && jobType != "Recurring" {
			fmt.Println("Error: --job-type must be either 'OneShot' or 'Recurring'")
			return
		}

		// Validate required fields based on job type
		if jobType == "OneShot" && scheduleAt == "" {
			fmt.Println("Error: --schedule-at is required for OneShot jobs")
			return
		}

		if jobType == "Recurring" {
			if cron == "" {
				fmt.Println("Error: --cron is required for Recurring jobs")
				return
			}
			if executeUntil == "" {
				fmt.Println("Error: --execute-until is required for Recurring jobs")
				return
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		properties := types.JobPropertiesRequest{
			Enabled: enabled,
			JobType: types.TypeJob(jobType),
		}

		if jobType == "OneShot" {
			properties.ScheduleAt = &scheduleAt
		} else {
			properties.Cron = &cron
			properties.ExecuteUntil = &executeUntil
		}

		createRequest := types.JobRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: properties,
		}

		ctx := context.Background()
		response, err := client.FromSchedule().Jobs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating job: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create job - Status: %d\n", response.StatusCode)
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
				{Header: "TYPE", Width: 15},
				{Header: "ENABLED", Width: 10},
				{Header: "REGION", Width: 20},
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
					if response.Data.Properties.JobType != "" {
						return string(response.Data.Properties.JobType)
					}
					return ""
				}(),
				func() string {
					if response.Data.Properties.Enabled {
						return "Yes"
					}
					return "No"
				}(),
				func() string {
					if response.Data.Metadata.LocationResponse != nil {
						return response.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Job created, but no data returned.")
		}
	},
}

var jobGetCmd = &cobra.Command{
	Use:   "get [job-id]",
	Short: "Get job details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

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
		resp, err := client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
		if err != nil {
			fmt.Printf("Error getting job: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get job - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			job := resp.Data

			fmt.Println("\nJob Details:")
			fmt.Println("============")

			if job.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *job.Metadata.ID)
			}
			if job.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *job.Metadata.URI)
			}
			if job.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *job.Metadata.Name)
			}
			if job.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", job.Metadata.LocationResponse.Value)
			}
			fmt.Printf("Job Type:        %s\n", job.Properties.JobType)
			fmt.Printf("Enabled:         %t\n", job.Properties.Enabled)
			if job.Properties.ScheduleAt != nil {
				fmt.Printf("Schedule At:     %s\n", *job.Properties.ScheduleAt)
			}
			if job.Properties.Cron != nil {
				fmt.Printf("CRON:            %s\n", *job.Properties.Cron)
			}
			if job.Properties.ExecuteUntil != nil {
				fmt.Printf("Execute Until:   %s\n", *job.Properties.ExecuteUntil)
			}
			if job.Status.State != nil {
				fmt.Printf("Status:          %s\n", *job.Status.State)
			}
			if !job.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", job.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if job.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *job.Metadata.CreatedBy)
			}
			if len(job.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", job.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			fmt.Println()
		} else {
			fmt.Println("Job not found")
		}
	},
}

var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scheduled jobs",
	Run: func(cmd *cobra.Command, args []string) {
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
		resp, err := client.FromSchedule().Jobs().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing jobs: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list jobs - Status: %d\n", resp.StatusCode)
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
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "TYPE", Width: 15},
				{Header: "ENABLED", Width: 10},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, job := range resp.Data.Values {
				row := []string{
					func() string {
						if job.Metadata.Name != nil {
							return *job.Metadata.Name
						}
						return ""
					}(),
					func() string {
						if job.Metadata.ID != nil {
							return *job.Metadata.ID
						}
						return ""
					}(),
					func() string {
						if job.Properties.JobType != "" {
							return string(job.Properties.JobType)
						}
						return ""
					}(),
					func() string {
						if job.Properties.Enabled {
							return "Yes"
						}
						return "No"
					}(),
					func() string {
						if job.Metadata.LocationResponse != nil {
							return job.Metadata.LocationResponse.Value
						}
						return ""
					}(),
					func() string {
						if job.Status.State != nil {
							return *job.Status.State
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No jobs found")
		}
	},
}

var jobUpdateCmd = &cobra.Command{
	Use:   "update [job-id]",
	Short: "Update a job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		enabledSet := cmd.Flags().Changed("enabled")
		enabled, _ := cmd.Flags().GetBool("enabled")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !enabledSet && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name, --enabled, or --tags must be provided")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		getResp, err := client.FromSchedule().Jobs().Get(ctx, projectID, jobID, nil)
		if err != nil {
			fmt.Printf("Error getting job: %v\n", err)
			return
		}

		if getResp == nil || getResp.Data == nil {
			fmt.Println("Job not found")
			return
		}

		current := getResp.Data

		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for job")
			return
		}

		updateRequest := types.JobRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *current.Metadata.Name,
					Tags: current.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.JobPropertiesRequest{
				Enabled:      current.Properties.Enabled,
				JobType:      current.Properties.JobType,
				ScheduleAt:   current.Properties.ScheduleAt,
				ExecuteUntil: current.Properties.ExecuteUntil,
				Cron:         current.Properties.Cron,
				// Steps are not included in update as they're read-only in response
			},
		}

		if name != "" {
			updateRequest.Metadata.ResourceMetadataRequest.Name = name
		}

		if enabledSet {
			updateRequest.Properties.Enabled = enabled
		}

		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.ResourceMetadataRequest.Tags = tags
		}

		response, err := client.FromSchedule().Jobs().Update(ctx, projectID, jobID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating job: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update job - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nJob updated successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			fmt.Printf("Enabled:         %t\n", response.Data.Properties.Enabled)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("Warning: Update may have succeeded but response is empty")
		}
	},
}

var jobDeleteCmd = &cobra.Command{
	Use:   "delete [job-id]",
	Short: "Delete a job",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			fmt.Printf("Are you sure you want to delete job %s? (yes/no): ", jobID)
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
		_, err = client.FromSchedule().Jobs().Delete(ctx, projectID, jobID, nil)
		if err != nil {
			fmt.Printf("Error deleting job: %v\n", err)
			return
		}

		fmt.Printf("\nJob %s deleted successfully!\n", jobID)
	},
}
