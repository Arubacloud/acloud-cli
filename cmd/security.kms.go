package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

type kmsGetView struct {
	ID, URI, Name, Region, Status, CreatedAt, CreatedBy, Tags string
}

func kmsRef(projectID, kmsID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Security/kms/" + kmsID)
}

func init() {
	// KMS commands
	securityCmd.AddCommand(kmsCmd)
	kmsCmd.AddCommand(kmsCreateCmd)
	kmsCmd.AddCommand(kmsGetCmd)
	kmsCmd.AddCommand(kmsUpdateCmd)
	kmsCmd.AddCommand(kmsDeleteCmd)
	kmsCmd.AddCommand(kmsListCmd)

	// Add flags for KMS commands
	kmsCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmsCreateCmd.Flags().String("name", "", "KMS name (required)")
	kmsCreateCmd.Flags().String("region", "", "Region code (required)")
	kmsCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
	kmsCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	kmsCreateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")
	kmsCreateCmd.MarkFlagRequired("name")
	kmsCreateCmd.MarkFlagRequired("region")

	kmsGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	kmsUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmsUpdateCmd.Flags().String("name", "", "New KMS name")
	kmsUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	kmsUpdateCmd.Flags().Bool("wait", false, "Wait until the resource becomes Active (use --timeout to control the deadline)")

	kmsDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmsDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	kmsDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	kmsListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmsListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	kmsListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	kmsGetCmd.ValidArgsFunction = completeKMSID
	kmsUpdateCmd.ValidArgsFunction = completeKMSID
	kmsDeleteCmd.ValidArgsFunction = completeKMSID
}

// Completion functions for security resources
func completeKMSID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("kms", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromSecurity().KMS().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, kms := range list.Items() {
			id := kms.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, kms.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// KMS subcommands
var kmsCmd = &cobra.Command{
	Use:   "kms",
	Short: "Manage Key Management System (KMS)",
	Long:  `Perform CRUD operations on KMS resources in Aruba Cloud.`,
}

var kmsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new KMS resource",
	Long: `Create a new Key Management System instance for storing and managing secrets.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud security kms create --name my-kms --region ITBG-Bergamo
  acloud security kms create --name prod-kms --region ITBG-Bergamo --billing-period Month`,
	Args: cobra.NoArgs,
	RunE: SecurityKMSCreateRun,
}

var kmsGetCmd = &cobra.Command{
	Use:   "get [kms-id]",
	Short: "Get KMS resource details",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKMSGetRun,
}

var kmsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KMS resources",
	Args:  cobra.NoArgs,
	RunE:  SecurityKMSListRun,
}

var kmsUpdateCmd = &cobra.Command{
	Use:   "update [kms-id]",
	Short: "Update a KMS resource",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKMSUpdateRun,
}

var kmsDeleteCmd = &cobra.Command{
	Use:   "delete [kms-id]",
	Short: "Delete a KMS resource",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKMSDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// SecurityKMSCreateArgs holds the typed arguments for creating a KMS resource.
type SecurityKMSCreateArgs struct {
	ProjectID     string
	Name          string
	Region        aruba.Region
	BillingPeriod aruba.BillingPeriod
	Tags          []string
	Wait          bool
}

// SecurityKMSGetArgs holds the typed arguments for getting a KMS resource.
type SecurityKMSGetArgs struct {
	ProjectID string
	ID        string
}

// SecurityKMSUpdateArgs holds the typed arguments for updating a KMS resource.
type SecurityKMSUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
	Wait        bool
}

// SecurityKMSDeleteArgs holds the typed arguments for deleting a KMS resource.
type SecurityKMSDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// SecurityKMSListArgs holds the typed arguments for listing KMS resources.
type SecurityKMSListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewSecurityKMSCreateArgsFromCobraCommand parses and validates args for create.
func NewSecurityKMSCreateArgsFromCobraCommand(cmd *cobra.Command) (*SecurityKMSCreateArgs, error) {
	args := &SecurityKMSCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKMSGetArgsFromCobraCommand parses and validates args for get.
func NewSecurityKMSGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKMSGetArgs, error) {
	args := &SecurityKMSGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKMSUpdateArgsFromCobraCommand parses and validates args for update.
func NewSecurityKMSUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKMSUpdateArgs, error) {
	args := &SecurityKMSUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKMSDeleteArgsFromCobraCommand parses and validates args for delete.
func NewSecurityKMSDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKMSDeleteArgs, error) {
	args := &SecurityKMSDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKMSListArgsFromCobraCommand parses and validates args for list.
func NewSecurityKMSListArgsFromCobraCommand(cmd *cobra.Command) (*SecurityKMSListArgs, error) {
	args := &SecurityKMSListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
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
func (a *SecurityKMSCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *SecurityKMSGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *SecurityKMSUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *SecurityKMSDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *SecurityKMSListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *SecurityKMSCreateArgs) Validate() error {
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
	if !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *SecurityKMSGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("KMS ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *SecurityKMSUpdateArgs) Validate() error {
	if a.Name == "" && !a.TagsChanged {
		return errors.New("at least one of --name or --tags must be provided")
	}
	return nil
}

// Validate checks the delete args for correctness.
func (a *SecurityKMSDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("KMS ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *SecurityKMSListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// SecurityKMSCreate creates a new KMS resource.
func SecurityKMSCreate(ctx context.Context, client aruba.Client, args SecurityKMSCreateArgs) error {
	kms := aruba.NewKMS().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		BilledBy(args.BillingPeriod).
		RetaggedAs(args.Tags...)

	created, err := client.FromSecurity().KMS().Create(ctx, kms)
	if err != nil {
		return fmt.Errorf("creating KMS: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 30},
			{Header: "NAME", Width: 40},
			{Header: "REGION", Width: 20},
			{Header: "STATUS", Width: 15},
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		regionVal := ""
		if raw.Metadata.LocationResponse != nil {
			regionVal = string(raw.Metadata.LocationResponse.Value)
		}
		statusVal := ""
		if raw.Status.State != nil {
			statusVal = string(*raw.Status.State)
		}
		row := []string{id, nameVal, regionVal, statusVal}
		PrintOutput(created, headers, [][]string{row})
		if args.Wait && id != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromSecurity().KMS().Get(ctx, kmsRef(args.ProjectID, id))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "KMS", args.Name); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgCreatedAsync("KMS", args.Name))
	}
	return nil
}

// SecurityKMSGet retrieves KMS resource details.
func SecurityKMSGet(ctx context.Context, client aruba.Client, args SecurityKMSGetArgs) error {
	kms, err := client.FromSecurity().KMS().Get(ctx, kmsRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting KMS: %w", apiErrFromV2(err))
	}

	if kms == nil || kms.Raw() == nil {
		fmt.Println("KMS not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(kms, nil, nil)
		return nil
	}
	raw := kms.Raw()
	view := kmsGetView{Tags: "[]"}
	if raw.Metadata.ID != nil {
		view.ID = *raw.Metadata.ID
	}
	if raw.Metadata.URI != nil {
		view.URI = *raw.Metadata.URI
	}
	if raw.Metadata.Name != nil {
		view.Name = *raw.Metadata.Name
	}
	if raw.Metadata.LocationResponse != nil {
		view.Region = string(raw.Metadata.LocationResponse.Value)
	}
	if raw.Status.State != nil {
		view.Status = string(*raw.Status.State)
	}
	if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
		view.CreatedAt = raw.Metadata.CreationDate.Format(DateLayout)
	}
	if raw.Metadata.CreatedBy != nil {
		view.CreatedBy = *raw.Metadata.CreatedBy
	}
	if len(raw.Metadata.Tags) > 0 {
		view.Tags = fmt.Sprintf("%v", raw.Metadata.Tags)
	}
	return renderGet(kmsGetTmpl, view)
}

// SecurityKMSUpdate updates a KMS resource's name and/or tags.
func SecurityKMSUpdate(ctx context.Context, client aruba.Client, args SecurityKMSUpdateArgs) error {
	kms, err := client.FromSecurity().KMS().Get(ctx, kmsRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting KMS: %w", apiErrFromV2(err))
	}

	if kms == nil || kms.Raw() == nil {
		return fmt.Errorf("KMS not found")
	}

	if args.Name != "" {
		kms.Named(args.Name)
	}
	if args.TagsChanged {
		kms.RetaggedAs(args.Tags...)
	}

	updated, err := client.FromSecurity().KMS().Update(ctx, kms)
	if err != nil {
		return fmt.Errorf("updating KMS: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("KMS", args.ID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
		if args.Wait && args.ID != "" {
			getter := func(ctx context.Context) (string, error) {
				res, err := client.FromSecurity().KMS().Get(ctx, kmsRef(args.ProjectID, args.ID))
				if err != nil {
					return "", apiErrFromV2(err)
				}
				if res == nil || res.Raw() == nil || res.Raw().Status.State == nil {
					return "", nil
				}
				return string(*res.Raw().Status.State), nil
			}
			if err := WaitUntilActive(ctx, getter, "KMS", args.ID); err != nil {
				return err
			}
		}
	} else {
		fmt.Println(msgUpdatedAsync("KMS", args.ID))
	}
	return nil
}

// SecurityKMSDelete deletes a KMS resource.
func SecurityKMSDelete(ctx context.Context, client aruba.Client, args SecurityKMSDeleteArgs) error {
	if args.DryRun {
		_, err := client.FromSecurity().KMS().Get(ctx, kmsRef(args.ProjectID, args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: KMS not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("KMS", args.ID))
		return nil
	}

	if err := client.FromSecurity().KMS().Delete(ctx, kmsRef(args.ProjectID, args.ID)); err != nil {
		return fmt.Errorf("deleting KMS: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("KMS", args.ID))
	return nil
}

// SecurityKMSList lists all KMS resources in a project.
func SecurityKMSList(ctx context.Context, client aruba.Client, args SecurityKMSListArgs) error {
	list, err := client.FromSecurity().KMS().List(ctx, aruba.URI("/projects/"+args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing KMS: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No KMS resources found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.KMS]{
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(k *aruba.KMS) string { return k.Name() }},
		{TableColumn: TableColumn{Header: "ID", Width: 30}, Value: func(k *aruba.KMS) string { return k.KMSID() }},
		{TableColumn: TableColumn{Header: "REGION", Width: 20}, Value: func(k *aruba.KMS) string { return string(k.Region()) }},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(k *aruba.KMS) string { return string(k.State()) }},
	}, list.Items(), func(k *aruba.KMS) bool { return k.Raw() != nil })
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// SecurityKMSCreateRun is the RunE wiring for KMS create.
func SecurityKMSCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewSecurityKMSCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKMSCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKMSGetRun is the RunE wiring for KMS get.
func SecurityKMSGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewSecurityKMSGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKMSGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKMSUpdateRun is the RunE wiring for KMS update.
func SecurityKMSUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewSecurityKMSUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKMSUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKMSDeleteRun is the RunE wiring for KMS delete.
func SecurityKMSDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewSecurityKMSDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("KMS", a.ID)
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

	if err := SecurityKMSDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKMSListRun is the RunE wiring for KMS list.
func SecurityKMSListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewSecurityKMSListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKMSList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
