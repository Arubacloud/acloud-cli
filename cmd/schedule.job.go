package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func jobRef(projectID, jobID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Schedule/jobs/" + jobID)
}

func init() {
	scheduleCmd.AddCommand(jobCmd)
	jobCmd.AddCommand(jobCreateCmd)
	jobCmd.AddCommand(jobGetCmd)
	jobCmd.AddCommand(jobUpdateCmd)
	jobCmd.AddCommand(jobDeleteCmd)
	jobCmd.AddCommand(jobListCmd)

	jobCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	jobCreateCmd.Flags().String("name", "", "Job name (required)")
	jobCreateCmd.Flags().String("region", "", "Region code (required)")
	jobCreateCmd.Flags().String("job-type", "", "Job type: OneShot or Recurring (required)")
	jobCreateCmd.Flags().String("schedule-at", "", "Date and time when the job should run (RFC3339, required for OneShot)")
	jobCreateCmd.Flags().String("cron", "", "CRON expression (required for Recurring)")
	jobCreateCmd.Flags().String("execute-until", "", "End date until which the job can run (RFC3339, required for Recurring)")
	jobCreateCmd.Flags().Bool("enabled", true, "Enable the job (default: true)")
	jobCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	jobCreateCmd.Flags().String("step-resource-uri", "", "Resource URI targeted by the step (required by the API)")
	jobCreateCmd.Flags().String("step-action-uri", "", "Action URI to invoke on the resource (e.g. poweroff, start)")
	jobCreateCmd.Flags().String("step-http-verb", string(aruba.HTTPVerbPOST), "HTTP verb for the step action")
	jobCreateCmd.Flags().String("step-name", "", "Optional display name for the step")
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
	jobDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	jobListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	jobListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	jobListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	jobGetCmd.ValidArgsFunction = completeJobID
	jobUpdateCmd.ValidArgsFunction = completeJobID
	jobDeleteCmd.ValidArgsFunction = completeJobID
}

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
	list, err := client.FromSchedule().Jobs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, job := range list.Items() {
			id := job.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, job.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Manage scheduled jobs",
	Long:  `Perform CRUD operations on scheduled jobs in Aruba Cloud.`,
}

var jobCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new scheduled job",
	Long: `Create a new scheduled job to automate tasks.

Job types:
  OneShot   — runs once at the time specified by --schedule-at (RFC3339 format)
  Recurring — runs on a cron schedule specified by --cron

For Recurring jobs, use --execute-until (RFC3339) to set an end date.
The job is enabled by default; pass --enabled=false to create it disabled.`,
	Example: `  # One-time job (power off a cloud server)
  acloud schedule job create --name my-job --region ITBG-Bergamo \
    --job-type OneShot --schedule-at 2026-06-01T10:00:00Z \
    --step-resource-uri /projects/<pid>/providers/Aruba.Compute/cloudServers/<id> \
    --step-action-uri poweroff

  # Recurring job (every day at midnight)
  acloud schedule job create --name daily-job --region ITBG-Bergamo \
    --job-type Recurring --cron "0 0 * * *" --execute-until 2027-01-01T00:00:00Z \
    --step-resource-uri /projects/<pid>/providers/Aruba.Compute/cloudServers/<id> \
    --step-action-uri poweroff`,
	Args: cobra.NoArgs,
	RunE: runJobCreate,
}

var jobGetCmd = &cobra.Command{
	Use:   "get [job-id]",
	Short: "Get job details",
	Args:  cobra.ExactArgs(1),
	RunE:  runJobGet,
}

var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scheduled jobs",
	Args:  cobra.NoArgs,
	RunE:  runJobList,
}

var jobUpdateCmd = &cobra.Command{
	Use:   "update [job-id]",
	Short: "Update a job",
	Args:  cobra.ExactArgs(1),
	RunE:  runJobUpdate,
}

var jobDeleteCmd = &cobra.Command{
	Use:   "delete [job-id]",
	Short: "Delete a job",
	Args:  cobra.ExactArgs(1),
	RunE:  runJobDelete,
}

func runJobCreate(cmd *cobra.Command, args []string) error {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	region, _ := cmd.Flags().GetString("region")
	jobType, _ := cmd.Flags().GetString("job-type")
	scheduleAt, _ := cmd.Flags().GetString("schedule-at")
	cron, _ := cmd.Flags().GetString("cron")
	executeUntil, _ := cmd.Flags().GetString("execute-until")
	enabled, _ := cmd.Flags().GetBool("enabled")
	tags, _ := cmd.Flags().GetStringSlice("tags")
	stepResourceURI, _ := cmd.Flags().GetString("step-resource-uri")
	stepActionURI, _ := cmd.Flags().GetString("step-action-uri")
	stepHTTPVerb, _ := cmd.Flags().GetString("step-http-verb")
	stepName, _ := cmd.Flags().GetString("step-name")

	if jobType != "OneShot" && jobType != "Recurring" {
		return fmt.Errorf("--job-type must be either 'OneShot' or 'Recurring'")
	}
	if jobType == "OneShot" && scheduleAt == "" {
		return fmt.Errorf("--schedule-at is required for OneShot jobs")
	}
	if jobType == "Recurring" {
		if cron == "" {
			return fmt.Errorf("--cron is required for Recurring jobs")
		}
		if executeUntil == "" {
			return fmt.Errorf("--execute-until is required for Recurring jobs")
		}
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	job := aruba.NewJob().
		InProject(aruba.URI("/projects/" + projectID)).
		Named(name).
		InRegion(aruba.Region(region)).
		RetaggedAs(tags...)

	if enabled {
		job.Enabled()
	} else {
		job.Disabled()
	}

	if jobType == "OneShot" {
		t, err := time.Parse(time.RFC3339, scheduleAt)
		if err != nil {
			return fmt.Errorf("invalid --schedule-at (use RFC3339, e.g. 2026-06-01T10:00:00Z): %w", err)
		}
		job.OneShotAt(t)
	} else {
		t, err := time.Parse(time.RFC3339, executeUntil)
		if err != nil {
			return fmt.Errorf("invalid --execute-until (use RFC3339, e.g. 2026-12-31T23:59:59Z): %w", err)
		}
		job.WithCron(cron)
		job.RecurringUntil(t)
	}

	if stepResourceURI != "" {
		step := aruba.NewJobStep().
			Targeting(aruba.URI(stepResourceURI)).
			WithVerb(aruba.HTTPVerb(stepHTTPVerb))
		if stepActionURI != "" {
			step.WithAction(stepActionURI)
		}
		if stepName != "" {
			step.Named(stepName)
		}
		job.WithSteps(step)
	}

	ctx, cancel := newCtx()
	defer cancel()
	created, err := client.FromSchedule().Jobs().Create(ctx, job)
	if err != nil {
		return fmt.Errorf("creating job: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "TYPE", Width: 15},
			{Header: "ENABLED", Width: 10},
			{Header: "REGION", Width: 20},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		jobTypeVal := string(raw.Properties.JobType)
		enabledVal := "No"
		if raw.Properties.Enabled {
			enabledVal = "Yes"
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		PrintOutput(created, headers, [][]string{{id, nameVal, jobTypeVal, enabledVal, regionVal}})
	} else {
		fmt.Println(msgCreatedAsync("Job", name))
	}
	return nil
}

func runJobGet(cmd *cobra.Command, args []string) error {
	jobID := args[0]

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
	job, err := client.FromSchedule().Jobs().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Schedule/jobs/"+jobID))
	if err != nil {
		return fmt.Errorf("getting job: %w", apiErrFromV2(err))
	}

	if job != nil && job.Raw() != nil {
		raw := job.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(job, nil, nil)
			return nil
		}

		fmt.Println("\nJob Details:")
		fmt.Println("============")
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Metadata.LocationResponse != nil {
			fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
		}
		fmt.Printf("Job Type:        %s\n", string(raw.Properties.JobType))
		fmt.Printf("Enabled:         %t\n", raw.Properties.Enabled)
		if raw.Properties.ScheduleAt != nil {
			fmt.Printf("Schedule At:     %s\n", *raw.Properties.ScheduleAt)
		}
		if raw.Properties.Cron != nil {
			fmt.Printf("CRON:            %s\n", *raw.Properties.Cron)
		}
		if raw.Properties.ExecuteUntil != nil {
			fmt.Printf("Execute Until:   %s\n", *raw.Properties.ExecuteUntil)
		}
		if raw.Status.State != nil {
			fmt.Printf("Status:          %s\n", string(*raw.Status.State))
		}
		if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
			fmt.Printf("Creation Date:   %s\n", raw.Metadata.CreationDate.Format(DateLayout))
		}
		if raw.Metadata.CreatedBy != nil {
			fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		} else {
			fmt.Printf("Tags:            []\n")
		}
		fmt.Println()
	} else {
		fmt.Println("Job not found")
	}
	return nil
}

func runJobList(cmd *cobra.Command, args []string) error {
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
	list, err := client.FromSchedule().Jobs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return fmt.Errorf("listing jobs: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 30},
			{Header: "ID", Width: 30},
			{Header: "TYPE", Width: 15},
			{Header: "ENABLED", Width: 10},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, job := range list.Items() {
			raw := job.Raw()
			if raw == nil {
				continue
			}
			name := ""
			if raw.Metadata.Name != nil {
				name = *raw.Metadata.Name
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			jobType := string(raw.Properties.JobType)
			enabledVal := "No"
			if raw.Properties.Enabled {
				enabledVal = "Yes"
			}
			region := ""
			if raw.Metadata.LocationResponse != nil {
				region = string(raw.Metadata.LocationResponse.Value)
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, jobType, enabledVal, region, status})
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No jobs found")
	}
	return nil
}

func runJobUpdate(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	enabledSet := cmd.Flags().Changed("enabled")
	enabled, _ := cmd.Flags().GetBool("enabled")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	if name == "" && !enabledSet && !cmd.Flags().Changed("tags") {
		return fmt.Errorf("at least one of --name, --enabled, or --tags must be provided")
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()
	job, err := client.FromSchedule().Jobs().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Schedule/jobs/"+jobID))
	if err != nil {
		return fmt.Errorf("getting job: %w", apiErrFromV2(err))
	}
	if job == nil || job.ID() == "" {
		return fmt.Errorf("job not found")
	}

	if name != "" {
		job.Named(name)
	}
	if enabledSet {
		if enabled {
			job.Enabled()
		} else {
			job.Disabled()
		}
	}
	if cmd.Flags().Changed("tags") {
		job.RetaggedAs(tags...)
	}

	updated, err := client.FromSchedule().Jobs().Update(ctx, job)
	if err != nil {
		return fmt.Errorf("updating job: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Job", jobID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		fmt.Printf("Enabled:         %t\n", raw.Properties.Enabled)
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgUpdatedAsync("Job", jobID))
	}
	return nil
}

func runJobDelete(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	confirm, _ := cmd.Flags().GetBool("yes")
	if !confirm {
		ok, err := confirmDelete("job", jobID)
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
		_, err = client.FromSchedule().Jobs().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Schedule/jobs/"+jobID))
		if err != nil {
			return fmt.Errorf("dry-run: schedule job not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("schedule job", jobID))
		return nil
	}

	if err = client.FromSchedule().Jobs().Delete(ctx, jobRef(projectID, jobID)); err != nil {
		return fmt.Errorf("deleting job: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Job", jobID))
	return nil
}
