package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

type kmipGetView struct {
	ID, Name, Type, Status, CreationDate, DeletionDate string
}

// kmipRef returns a Ref for a specific KMIP service nested inside a KMS instance.
// Note: the URI uses "kmips" (plural) so the SDK's ID extractor can parse the
// kmip ID from the path segment, even though the actual HTTP path uses "kmip"
// (singular, per the API URL scheme in sdk-go/internal/clients/security/path.go).
func kmipRef(projectID, kmsID, kmipID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Security/kms/" + kmsID +
		"/kmips/" + kmipID)
}

func init() {
	// KMIP commands (nested under securityCmd)
	securityCmd.AddCommand(kmipCmd)
	kmipCmd.AddCommand(kmipCreateCmd)
	kmipCmd.AddCommand(kmipGetCmd)
	kmipCmd.AddCommand(kmipDeleteCmd)
	kmipCmd.AddCommand(kmipListCmd)
	kmipCmd.AddCommand(kmipDownloadCmd)

	// Add flags for KMIP commands
	kmipCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmipCreateCmd.Flags().String("kms-id", "", "KMS ID (required)")
	kmipCreateCmd.Flags().String("name", "", "KMIP service name (required)")
	kmipCreateCmd.Flags().Bool("wait", false, "Wait until the certificate is available (use --timeout to control the deadline)")
	kmipCreateCmd.MarkFlagRequired("kms-id")
	kmipCreateCmd.MarkFlagRequired("name")

	kmipGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmipGetCmd.Flags().String("kms-id", "", "KMS ID (required)")
	kmipGetCmd.MarkFlagRequired("kms-id")

	kmipDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmipDeleteCmd.Flags().String("kms-id", "", "KMS ID (required)")
	kmipDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	kmipDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	kmipDeleteCmd.MarkFlagRequired("kms-id")

	kmipListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmipListCmd.Flags().String("kms-id", "", "KMS ID (required)")
	kmipListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	kmipListCmd.Flags().Int32("offset", 0, "Number of results to skip")
	kmipListCmd.MarkFlagRequired("kms-id")

	kmipDownloadCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmipDownloadCmd.Flags().String("kms-id", "", "KMS ID (required)")
	kmipDownloadCmd.MarkFlagRequired("kms-id")

	// Set up auto-completion for resource IDs
	kmipGetCmd.ValidArgsFunction = completeKmipID
	kmipDeleteCmd.ValidArgsFunction = completeKmipID
	kmipDownloadCmd.ValidArgsFunction = completeKmipID
}

func completeKmipID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	kmsID, err := cmd.Flags().GetString("kms-id")
	if err != nil || kmsID == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("security-kmip", projectID, kmsID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromSecurity().Kmips().List(ctx, kmsRef(projectID, kmsID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, km := range list.Items() {
			id := km.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, km.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// KMIP subcommands
var kmipCmd = &cobra.Command{
	Use:   "kmip",
	Short: "Manage KMIP services inside a KMS instance",
	Long:  `Perform CRUD operations on KMIP services nested inside a KMS instance.`,
}

var kmipCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new KMIP service inside a KMS instance",
	Long: `Create a new KMIP service inside an existing KMS instance.

Use --wait to block until the certificate becomes available for download.`,
	Example: `  acloud security kmip create --kms-id kms-001 --name my-kmip
  acloud security kmip create --kms-id kms-001 --name my-kmip --wait --project-id proj-123`,
	Args: cobra.NoArgs,
	RunE: SecurityKmipCreateRun,
}

var kmipGetCmd = &cobra.Command{
	Use:   "get [kmip-id]",
	Short: "Get KMIP service details",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKmipGetRun,
}

var kmipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KMIP services in a KMS instance",
	Args:  cobra.NoArgs,
	RunE:  SecurityKmipListRun,
}

var kmipDeleteCmd = &cobra.Command{
	Use:   "delete [kmip-id]",
	Short: "Delete a KMIP service",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKmipDeleteRun,
}

var kmipDownloadCmd = &cobra.Command{
	Use:   "download [kmip-id]",
	Short: "Download the KMIP certificate (key and cert PEM pair)",
	Long: `Download the PEM-encoded certificate and private key for a KMIP service.

The certificate is only available after the service reaches the CertificateAvailable status.`,
	Args: cobra.ExactArgs(1),
	RunE: SecurityKmipDownloadRun,
}

// =============================================================================
// Args structs
// =============================================================================

// SecurityKmipCreateArgs holds the typed arguments for creating a KMIP service.
type SecurityKmipCreateArgs struct {
	ProjectID string
	KMSID     string
	Name      string
	Wait      bool
}

// SecurityKmipGetArgs holds the typed arguments for getting a KMIP service.
type SecurityKmipGetArgs struct {
	ProjectID string
	KMSID     string
	ID        string
}

// SecurityKmipDeleteArgs holds the typed arguments for deleting a KMIP service.
type SecurityKmipDeleteArgs struct {
	ProjectID   string
	KMSID       string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// SecurityKmipListArgs holds the typed arguments for listing KMIP services.
type SecurityKmipListArgs struct {
	ProjectID string
	KMSID     string
	CallOpts  []aruba.CallOption
}

// SecurityKmipDownloadArgs holds the typed arguments for downloading a KMIP certificate.
type SecurityKmipDownloadArgs struct {
	ProjectID string
	KMSID     string
	ID        string
}

// =============================================================================
// Constructors
// =============================================================================

// NewSecurityKmipCreateArgsFromCobraCommand parses and validates args for create.
func NewSecurityKmipCreateArgsFromCobraCommand(cmd *cobra.Command) (*SecurityKmipCreateArgs, error) {
	args := &SecurityKmipCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKmipGetArgsFromCobraCommand parses and validates args for get.
func NewSecurityKmipGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKmipGetArgs, error) {
	args := &SecurityKmipGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKmipDeleteArgsFromCobraCommand parses and validates args for delete.
func NewSecurityKmipDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKmipDeleteArgs, error) {
	args := &SecurityKmipDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKmipListArgsFromCobraCommand parses and validates args for list.
func NewSecurityKmipListArgsFromCobraCommand(cmd *cobra.Command) (*SecurityKmipListArgs, error) {
	args := &SecurityKmipListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKmipDownloadArgsFromCobraCommand parses and validates args for download.
func NewSecurityKmipDownloadArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKmipDownloadArgs, error) {
	args := &SecurityKmipDownloadArgs{}
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
func (a *SecurityKmipCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.KMSID, err = cmd.Flags().GetString("kms-id"); err != nil {
		errs = append(errs, err)
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Wait, err = cmd.Flags().GetBool("wait"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *SecurityKmipGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.KMSID, err = cmd.Flags().GetString("kms-id"); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *SecurityKmipDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.KMSID, err = cmd.Flags().GetString("kms-id"); err != nil {
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
func (a *SecurityKmipListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.KMSID, err = cmd.Flags().GetString("kms-id"); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the download args struct.
func (a *SecurityKmipDownloadArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.KMSID, err = cmd.Flags().GetString("kms-id"); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *SecurityKmipCreateArgs) Validate() error {
	var errs []error

	if len(a.Name) < 3 {
		errs = append(errs, errors.New("--name must be at least 3 characters"))
	}
	if len(a.Name) > 64 {
		errs = append(errs, errors.New("--name must be at most 64 characters"))
	}
	if a.KMSID == "" {
		errs = append(errs, errors.New("--kms-id is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *SecurityKmipGetArgs) Validate() error {
	var errs []error
	if a.KMSID == "" {
		errs = append(errs, errors.New("--kms-id is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("KMIP ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *SecurityKmipDeleteArgs) Validate() error {
	var errs []error
	if a.KMSID == "" {
		errs = append(errs, errors.New("--kms-id is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("KMIP ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *SecurityKmipListArgs) Validate() error {
	if a.KMSID == "" {
		return errors.New("--kms-id is required")
	}
	return nil
}

// Validate checks the download args for correctness.
func (a *SecurityKmipDownloadArgs) Validate() error {
	var errs []error
	if a.KMSID == "" {
		errs = append(errs, errors.New("--kms-id is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("KMIP ID is required"))
	}
	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// SecurityKmipCreate creates a new KMIP service inside a KMS instance.
func SecurityKmipCreate(ctx context.Context, client aruba.Client, args SecurityKmipCreateArgs) error {
	km := aruba.NewKmip().
		InKMS(kmsRef(args.ProjectID, args.KMSID)).
		Named(args.Name)

	created, err := client.FromSecurity().Kmips().Create(ctx, km)
	if err != nil {
		return fmt.Errorf("creating KMIP: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 36},
			{Header: "NAME", Width: 30},
			{Header: "TYPE", Width: 15},
			{Header: "STATUS", Width: 25},
		}
		id := ""
		if raw.ID != nil {
			id = *raw.ID
		}
		nameVal := ""
		if raw.Name != nil {
			nameVal = *raw.Name
		}
		typeVal := ""
		if raw.Type != nil {
			typeVal = *raw.Type
		}
		statusVal := ""
		if raw.Status != nil {
			statusVal = string(*raw.Status)
		}
		row := []string{id, nameVal, typeVal, statusVal}
		PrintOutput(created, headers, [][]string{row})
		if args.Wait && id != "" {
			if err := created.WaitUntilCertificateAvailable(ctx); err != nil {
				return fmt.Errorf("waiting for KMIP certificate: %w", err)
			}
		}
	} else {
		fmt.Println(msgCreatedAsync("KMIP", args.Name))
	}
	return nil
}

// SecurityKmipGet retrieves KMIP service details.
func SecurityKmipGet(ctx context.Context, client aruba.Client, args SecurityKmipGetArgs) error {
	km, err := client.FromSecurity().Kmips().Get(ctx, kmipRef(args.ProjectID, args.KMSID, args.ID))
	if err != nil {
		return fmt.Errorf("getting KMIP: %w", apiErrFromV2(err))
	}

	if km == nil || km.Raw() == nil {
		fmt.Println("KMIP not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(km, nil, nil)
		return nil
	}
	raw := km.Raw()
	view := kmipGetView{}
	if raw.ID != nil {
		view.ID = *raw.ID
	}
	if raw.Name != nil {
		view.Name = *raw.Name
	}
	if raw.Type != nil {
		view.Type = *raw.Type
	}
	if raw.Status != nil {
		view.Status = string(*raw.Status)
	}
	if raw.CreationDate != nil {
		view.CreationDate = *raw.CreationDate
	}
	if raw.DeletionDate != nil {
		view.DeletionDate = *raw.DeletionDate
	}
	return renderGet(kmipGetTmpl, view)
}

// SecurityKmipDelete deletes a KMIP service.
func SecurityKmipDelete(ctx context.Context, client aruba.Client, args SecurityKmipDeleteArgs) error {
	if args.DryRun {
		_, err := client.FromSecurity().Kmips().Get(ctx, kmipRef(args.ProjectID, args.KMSID, args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: KMIP not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("KMIP", args.ID))
		return nil
	}

	if err := client.FromSecurity().Kmips().Delete(ctx, kmipRef(args.ProjectID, args.KMSID, args.ID)); err != nil {
		return fmt.Errorf("deleting KMIP: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("KMIP", args.ID))
	return nil
}

// SecurityKmipList lists all KMIP services in a KMS instance.
func SecurityKmipList(ctx context.Context, client aruba.Client, args SecurityKmipListArgs) error {
	list, err := client.FromSecurity().Kmips().List(ctx, kmsRef(args.ProjectID, args.KMSID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing KMIP services: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No KMIP services found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.Kmip]{
		{TableColumn: TableColumn{Header: "ID", Width: 36}, Value: func(km *aruba.Kmip) string { return km.ID() }},
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(km *aruba.Kmip) string { return km.Name() }},
		{TableColumn: TableColumn{Header: "TYPE", Width: 15}, Value: func(km *aruba.Kmip) string { return km.Type() }},
		{TableColumn: TableColumn{Header: "STATUS", Width: 25}, Value: func(km *aruba.Kmip) string { return km.KmipStatus() }},
	}, list.Items(), func(km *aruba.Kmip) bool { return km.Raw() != nil })
	return nil
}

// SecurityKmipDownload downloads the KMIP certificate key+cert pair.
func SecurityKmipDownload(ctx context.Context, client aruba.Client, args SecurityKmipDownloadArgs) error {
	cert, err := client.FromSecurity().Kmips().Download(ctx, kmipRef(args.ProjectID, args.KMSID, args.ID))
	if err != nil {
		return fmt.Errorf("downloading KMIP certificate: %w", apiErrFromV2(err))
	}

	if cert == nil || cert.Raw() == nil {
		fmt.Println("No certificate available for KMIP", args.ID)
		return nil
	}

	fmt.Printf("=== Certificate ===\n%s\n", cert.Cert())
	fmt.Printf("=== Private Key ===\n%s\n", cert.Key())
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// SecurityKmipCreateRun is the RunE wiring for KMIP create.
func SecurityKmipCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewSecurityKmipCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKmipCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKmipGetRun is the RunE wiring for KMIP get.
func SecurityKmipGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewSecurityKmipGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKmipGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKmipDeleteRun is the RunE wiring for KMIP delete.
func SecurityKmipDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewSecurityKmipDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("KMIP", a.ID)
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

	if err := SecurityKmipDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKmipListRun is the RunE wiring for KMIP list.
func SecurityKmipListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewSecurityKmipListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKmipList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKmipDownloadRun is the RunE wiring for KMIP download.
func SecurityKmipDownloadRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewSecurityKmipDownloadArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKmipDownload(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
