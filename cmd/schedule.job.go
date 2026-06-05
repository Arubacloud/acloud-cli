package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
			id := job.JobID()
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
	RunE: ScheduleJobCreateRun,
}

var jobGetCmd = &cobra.Command{
	Use:   "get [job-id]",
	Short: "Get job details",
	Args:  cobra.ExactArgs(1),
	RunE:  ScheduleJobGetRun,
}

var jobListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all scheduled jobs",
	Args:  cobra.NoArgs,
	RunE:  ScheduleJobListRun,
}

var jobUpdateCmd = &cobra.Command{
	Use:   "update [job-id]",
	Short: "Update a job",
	Args:  cobra.ExactArgs(1),
	RunE:  ScheduleJobUpdateRun,
}

var jobDeleteCmd = &cobra.Command{
	Use:   "delete [job-id]",
	Short: "Delete a job",
	Args:  cobra.ExactArgs(1),
	RunE:  ScheduleJobDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// ScheduleJobCreateArgs holds the typed arguments for creating a scheduled job.
type ScheduleJobCreateArgs struct {
	ProjectID       string
	Name            string
	Region          aruba.Region
	JobType         aruba.JobType
	Enabled         bool
	Tags            []string
	CronExpr        string
	EndTime         string
	ShotTime        string
	StepResourceURI string
	StepActionURI   string
	StepHTTPVerb    aruba.HTTPVerb
	StepName        string
}

// ScheduleJobGetArgs holds the typed arguments for getting a scheduled job.
type ScheduleJobGetArgs struct {
	ProjectID string
	ID        string
}

// ScheduleJobListArgs holds the typed arguments for listing scheduled jobs.
type ScheduleJobListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// ScheduleJobUpdateArgs holds the typed arguments for updating a scheduled job.
type ScheduleJobUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Enabled     bool
	EnabledSet  bool
	Tags        []string
	TagsChanged bool
}

// ScheduleJobDeleteArgs holds the typed arguments for deleting a scheduled job.
type ScheduleJobDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// =============================================================================
// Constructors
// =============================================================================

// NewScheduleJobCreateArgsFromCobraCommand parses and validates args for create.
func NewScheduleJobCreateArgsFromCobraCommand(cmd *cobra.Command) (*ScheduleJobCreateArgs, error) {
	args := &ScheduleJobCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewScheduleJobGetArgsFromCobraCommand parses and validates args for get.
func NewScheduleJobGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ScheduleJobGetArgs, error) {
	args := &ScheduleJobGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewScheduleJobListArgsFromCobraCommand parses and validates args for list.
func NewScheduleJobListArgsFromCobraCommand(cmd *cobra.Command) (*ScheduleJobListArgs, error) {
	args := &ScheduleJobListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewScheduleJobUpdateArgsFromCobraCommand parses and validates args for update.
func NewScheduleJobUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ScheduleJobUpdateArgs, error) {
	args := &ScheduleJobUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewScheduleJobDeleteArgsFromCobraCommand parses and validates args for delete.
func NewScheduleJobDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ScheduleJobDeleteArgs, error) {
	args := &ScheduleJobDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// =============================================================================
// ParseFromCobraCommand methods
// =============================================================================

// ParseFromCobraCommand reads Cobra flags into the create args struct.
func (a *ScheduleJobCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("job-type"); err == nil {
		a.JobType = aruba.JobType(s)
	} else {
		errs = append(errs, err)
	}
	if a.Enabled, err = cmd.Flags().GetBool("enabled"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.CronExpr, err = cmd.Flags().GetString("cron"); err != nil {
		errs = append(errs, err)
	}
	if a.EndTime, err = cmd.Flags().GetString("execute-until"); err != nil {
		errs = append(errs, err)
	}
	if a.ShotTime, err = cmd.Flags().GetString("schedule-at"); err != nil {
		errs = append(errs, err)
	}
	if a.StepResourceURI, err = cmd.Flags().GetString("step-resource-uri"); err != nil {
		errs = append(errs, err)
	}
	if a.StepActionURI, err = cmd.Flags().GetString("step-action-uri"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("step-http-verb"); err == nil {
		a.StepHTTPVerb = aruba.HTTPVerb(s)
	} else {
		errs = append(errs, err)
	}
	if a.StepName, err = cmd.Flags().GetString("step-name"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *ScheduleJobGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *ScheduleJobListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *ScheduleJobUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Enabled, err = cmd.Flags().GetBool("enabled"); err != nil {
		errs = append(errs, err)
	}
	a.EnabledSet = cmd.Flags().Changed("enabled")
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *ScheduleJobDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *ScheduleJobCreateArgs) Validate() error {
	var errs []error

	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}
	if !slices.Contains(validRegions, a.Region) {
		errs = append(errs, fmt.Errorf("--region %q: must be one of %v", a.Region, validRegions))
	}
	if !slices.Contains(validJobTypes, a.JobType) {
		errs = append(errs, fmt.Errorf("--job-type %q: must be one of %v", a.JobType, validJobTypes))
	}

	// Mutually exclusive schedule modes.
	if a.JobType == aruba.JobTypeRecurring {
		if a.CronExpr == "" {
			errs = append(errs, errors.New("--cron is required for Recurring jobs"))
		}
		if a.EndTime == "" {
			errs = append(errs, errors.New("--execute-until is required for Recurring jobs"))
		}
	}
	if a.JobType == aruba.JobTypeOneShot {
		if a.ShotTime == "" {
			errs = append(errs, errors.New("--schedule-at is required for OneShot jobs"))
		}
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *ScheduleJobGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("job ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *ScheduleJobListArgs) Validate() error {
	return nil
}

// Validate checks the update args for correctness.
func (a *ScheduleJobUpdateArgs) Validate() error {
	if a.Name == "" && !a.EnabledSet && !a.TagsChanged {
		return errors.New("at least one of --name, --enabled, or --tags must be provided")
	}
	return nil
}

// Validate checks the delete args for correctness.
func (a *ScheduleJobDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("job ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// ScheduleJobCreate creates a new scheduled job.
func ScheduleJobCreate(ctx context.Context, client aruba.Client, args ScheduleJobCreateArgs) error {
	job := aruba.NewJob().
		InProject(projectRef(args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		OfType(args.JobType).
		RetaggedAs(args.Tags...)

	if args.Enabled {
		job.Enabled()
	} else {
		job.Disabled()
	}

	if args.JobType == aruba.JobTypeOneShot {
		t, err := time.Parse(time.RFC3339, args.ShotTime)
		if err != nil {
			return fmt.Errorf("invalid --schedule-at (use RFC3339, e.g. 2026-06-01T10:00:00Z): %w", err)
		}
		job.OneShotAt(t)
	} else {
		t, err := time.Parse(time.RFC3339, args.EndTime)
		if err != nil {
			return fmt.Errorf("invalid --execute-until (use RFC3339, e.g. 2026-12-31T23:59:59Z): %w", err)
		}
		job.WithCron(args.CronExpr)
		job.RecurringUntil(t)
	}

	if args.StepResourceURI != "" {
		step := aruba.NewJobStep().
			Targeting(aruba.URI(args.StepResourceURI)).
			WithVerb(args.StepHTTPVerb)
		if args.StepActionURI != "" {
			step.WithAction(args.StepActionURI)
		}
		if args.StepName != "" {
			step.Named(args.StepName)
		}
		job.WithSteps(step)
	}

	if err := job.Err(); err != nil {
		return fmt.Errorf("invalid job configuration: %w", err)
	}

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
		fmt.Println(msgCreatedAsync("Job", args.Name))
	}
	return nil
}

// ScheduleJobGet retrieves scheduled job details.
func ScheduleJobGet(ctx context.Context, client aruba.Client, args ScheduleJobGetArgs) error {
	job, err := client.FromSchedule().Jobs().Get(ctx, jobRef(args.ProjectID, args.ID))
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

// ScheduleJobList lists all scheduled jobs in a project.
func ScheduleJobList(ctx context.Context, client aruba.Client, args ScheduleJobListArgs) error {
	list, err := client.FromSchedule().Jobs().List(ctx, aruba.URI("/projects/"+args.ProjectID), args.CallOpts...)
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

// ScheduleJobUpdate updates a scheduled job's name, enabled state, and/or tags.
func ScheduleJobUpdate(ctx context.Context, client aruba.Client, args ScheduleJobUpdateArgs) error {
	job, err := client.FromSchedule().Jobs().Get(ctx, jobRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting job: %w", apiErrFromV2(err))
	}
	if job == nil || job.JobID() == "" {
		return fmt.Errorf("job not found")
	}

	if args.Name != "" {
		job.Named(args.Name)
	}
	if args.EnabledSet {
		if args.Enabled {
			job.Enabled()
		} else {
			job.Disabled()
		}
	}
	if args.TagsChanged {
		job.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromSchedule().Jobs().Update(ctx, job)
	if err != nil {
		return fmt.Errorf("updating job: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("Job", args.ID))
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
		fmt.Println(msgUpdatedAsync("Job", args.ID))
	}
	return nil
}

// ScheduleJobDelete deletes a scheduled job.
func ScheduleJobDelete(ctx context.Context, client aruba.Client, args ScheduleJobDeleteArgs) error {
	if args.DryRun {
		_, err := client.FromSchedule().Jobs().Get(ctx, jobRef(args.ProjectID, args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: schedule job not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("schedule job", args.ID))
		return nil
	}

	if err := client.FromSchedule().Jobs().Delete(ctx, jobRef(args.ProjectID, args.ID)); err != nil {
		return fmt.Errorf("deleting job: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Job", args.ID))
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// ScheduleJobCreateRun is the RunE wiring for job create.
func ScheduleJobCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewScheduleJobCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ScheduleJobCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ScheduleJobGetRun is the RunE wiring for job get.
func ScheduleJobGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewScheduleJobGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ScheduleJobGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ScheduleJobListRun is the RunE wiring for job list.
func ScheduleJobListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewScheduleJobListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ScheduleJobList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ScheduleJobUpdateRun is the RunE wiring for job update.
func ScheduleJobUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewScheduleJobUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ScheduleJobUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ScheduleJobDeleteRun is the RunE wiring for job delete.
func ScheduleJobDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewScheduleJobDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("job", a.ID)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ScheduleJobDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
