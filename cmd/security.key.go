package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

type keyGetView struct {
	ID, Name, Algorithm, Type, Status, CreationSource, PrivateKeyID string
}

// keyRef returns a Ref for a specific key nested inside a KMS instance.
func keyRef(projectID, kmsID, keyID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Security/kms/" + kmsID +
		"/keys/" + keyID)
}

func init() {
	// Key commands (nested under kmsCmd from security.kms.go)
	securityCmd.AddCommand(keyCmd)
	keyCmd.AddCommand(keyCreateCmd)
	keyCmd.AddCommand(keyGetCmd)
	keyCmd.AddCommand(keyDeleteCmd)
	keyCmd.AddCommand(keyListCmd)

	// Add flags for Key commands
	keyCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keyCreateCmd.Flags().String("kms-id", "", "KMS ID (required)")
	keyCreateCmd.Flags().String("name", "", "Key name (required)")
	keyCreateCmd.Flags().String("algorithm", "", "Cryptographic algorithm: Aes, Rsa (required)")
	keyCreateCmd.MarkFlagRequired("kms-id")
	keyCreateCmd.MarkFlagRequired("name")
	keyCreateCmd.MarkFlagRequired("algorithm")

	keyGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keyGetCmd.Flags().String("kms-id", "", "KMS ID (required)")
	keyGetCmd.MarkFlagRequired("kms-id")

	keyDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keyDeleteCmd.Flags().String("kms-id", "", "KMS ID (required)")
	keyDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	keyDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	keyDeleteCmd.MarkFlagRequired("kms-id")

	keyListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keyListCmd.Flags().String("kms-id", "", "KMS ID (required)")
	keyListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	keyListCmd.Flags().Int32("offset", 0, "Number of results to skip")
	keyListCmd.MarkFlagRequired("kms-id")

	// Set up auto-completion for resource IDs
	keyGetCmd.ValidArgsFunction = completeKeyID
	keyDeleteCmd.ValidArgsFunction = completeKeyID
}

func completeKeyID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	kmsID, err := cmd.Flags().GetString("kms-id")
	if err != nil || kmsID == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	key := cacheKey("security-key", projectID, kmsID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromSecurity().Keys().List(ctx, kmsRef(projectID, kmsID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, k := range list.Items() {
			id := k.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, k.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// Key subcommands
var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage KMS cryptographic keys",
	Long:  `Perform CRUD operations on cryptographic keys nested inside a KMS instance.`,
}

var keyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cryptographic key inside a KMS instance",
	Long: `Create a new cryptographic key inside an existing KMS instance.

Supported algorithms: Aes (symmetric), Rsa (asymmetric).`,
	Example: `  acloud security key create --kms-id kms-001 --name my-key --algorithm Aes
  acloud security key create --kms-id kms-001 --name rsa-key --algorithm Rsa --project-id proj-123`,
	Args: cobra.NoArgs,
	RunE: SecurityKeyCreateRun,
}

var keyGetCmd = &cobra.Command{
	Use:   "get [key-id]",
	Short: "Get key details",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKeyGetRun,
}

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys in a KMS instance",
	Args:  cobra.NoArgs,
	RunE:  SecurityKeyListRun,
}

var keyDeleteCmd = &cobra.Command{
	Use:   "delete [key-id]",
	Short: "Delete a cryptographic key",
	Args:  cobra.ExactArgs(1),
	RunE:  SecurityKeyDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// SecurityKeyCreateArgs holds the typed arguments for creating a key.
type SecurityKeyCreateArgs struct {
	ProjectID string
	KMSID     string
	Name      string
	Algorithm aruba.KeyAlgorithm
}

// SecurityKeyGetArgs holds the typed arguments for getting a key.
type SecurityKeyGetArgs struct {
	ProjectID string
	KMSID     string
	ID        string
}

// SecurityKeyDeleteArgs holds the typed arguments for deleting a key.
type SecurityKeyDeleteArgs struct {
	ProjectID   string
	KMSID       string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// SecurityKeyListArgs holds the typed arguments for listing keys.
type SecurityKeyListArgs struct {
	ProjectID string
	KMSID     string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewSecurityKeyCreateArgsFromCobraCommand parses and validates args for create.
func NewSecurityKeyCreateArgsFromCobraCommand(cmd *cobra.Command) (*SecurityKeyCreateArgs, error) {
	args := &SecurityKeyCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKeyGetArgsFromCobraCommand parses and validates args for get.
func NewSecurityKeyGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKeyGetArgs, error) {
	args := &SecurityKeyGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKeyDeleteArgsFromCobraCommand parses and validates args for delete.
func NewSecurityKeyDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*SecurityKeyDeleteArgs, error) {
	args := &SecurityKeyDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewSecurityKeyListArgsFromCobraCommand parses and validates args for list.
func NewSecurityKeyListArgsFromCobraCommand(cmd *cobra.Command) (*SecurityKeyListArgs, error) {
	args := &SecurityKeyListArgs{}
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
func (a *SecurityKeyCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if s, err := cmd.Flags().GetString("algorithm"); err == nil {
		a.Algorithm = aruba.KeyAlgorithm(s)
	} else {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *SecurityKeyGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *SecurityKeyDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *SecurityKeyListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *SecurityKeyCreateArgs) Validate() error {
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
	if !slices.Contains(validKeyAlgorithms, a.Algorithm) {
		errs = append(errs, fmt.Errorf("--algorithm %q: must be one of %v", a.Algorithm, validKeyAlgorithms))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *SecurityKeyGetArgs) Validate() error {
	var errs []error
	if a.KMSID == "" {
		errs = append(errs, errors.New("--kms-id is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("key ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *SecurityKeyDeleteArgs) Validate() error {
	var errs []error
	if a.KMSID == "" {
		errs = append(errs, errors.New("--kms-id is required"))
	}
	if a.ID == "" {
		errs = append(errs, errors.New("key ID is required"))
	}
	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *SecurityKeyListArgs) Validate() error {
	if a.KMSID == "" {
		return errors.New("--kms-id is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// SecurityKeyCreate creates a new cryptographic key inside a KMS instance.
func SecurityKeyCreate(ctx context.Context, client aruba.Client, args SecurityKeyCreateArgs) error {
	k := aruba.NewKey().
		InKMS(kmsRef(args.ProjectID, args.KMSID)).
		Named(args.Name).
		OfAlgorithm(args.Algorithm)

	created, err := client.FromSecurity().Keys().Create(ctx, k)
	if err != nil {
		return fmt.Errorf("creating key: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		headers := []TableColumn{
			{Header: "ID", Width: 36},
			{Header: "NAME", Width: 30},
			{Header: "ALGORITHM", Width: 15},
			{Header: "STATUS", Width: 15},
		}
		id := ""
		if raw.KeyID != nil {
			id = *raw.KeyID
		}
		nameVal := ""
		if raw.Name != nil {
			nameVal = *raw.Name
		}
		algoVal := ""
		if raw.Algorithm != nil {
			algoVal = string(*raw.Algorithm)
		}
		statusVal := ""
		if raw.Status != nil {
			statusVal = string(*raw.Status)
		}
		row := []string{id, nameVal, algoVal, statusVal}
		PrintOutput(created, headers, [][]string{row})
	} else {
		fmt.Println(msgCreatedAsync("key", args.Name))
	}
	return nil
}

// SecurityKeyGet retrieves key details.
func SecurityKeyGet(ctx context.Context, client aruba.Client, args SecurityKeyGetArgs) error {
	k, err := client.FromSecurity().Keys().Get(ctx, keyRef(args.ProjectID, args.KMSID, args.ID))
	if err != nil {
		return fmt.Errorf("getting key: %w", apiErrFromV2(err))
	}

	if k == nil || k.Raw() == nil {
		fmt.Println("Key not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(k, nil, nil)
		return nil
	}
	raw := k.Raw()
	view := keyGetView{}
	if raw.KeyID != nil {
		view.ID = *raw.KeyID
	}
	if raw.Name != nil {
		view.Name = *raw.Name
	}
	if raw.Algorithm != nil {
		view.Algorithm = string(*raw.Algorithm)
	}
	if raw.Type != nil {
		view.Type = string(*raw.Type)
	}
	if raw.Status != nil {
		view.Status = string(*raw.Status)
	}
	if raw.CreationSource != nil {
		view.CreationSource = string(*raw.CreationSource)
	}
	if raw.PrivateKeyID != nil {
		view.PrivateKeyID = *raw.PrivateKeyID
	}
	return renderGet(keyGetTmpl, view)
}

// SecurityKeyDelete deletes a cryptographic key.
func SecurityKeyDelete(ctx context.Context, client aruba.Client, args SecurityKeyDeleteArgs) error {
	if args.DryRun {
		_, err := client.FromSecurity().Keys().Get(ctx, keyRef(args.ProjectID, args.KMSID, args.ID))
		if err != nil {
			return fmt.Errorf("dry-run: key not found or inaccessible: %w", err)
		}
		fmt.Println(msgDryRun("key", args.ID))
		return nil
	}

	if err := client.FromSecurity().Keys().Delete(ctx, keyRef(args.ProjectID, args.KMSID, args.ID)); err != nil {
		return fmt.Errorf("deleting key: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("key", args.ID))
	return nil
}

// SecurityKeyList lists all keys in a KMS instance.
func SecurityKeyList(ctx context.Context, client aruba.Client, args SecurityKeyListArgs) error {
	list, err := client.FromSecurity().Keys().List(ctx, kmsRef(args.ProjectID, args.KMSID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing keys: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No keys found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.Key]{
		{TableColumn: TableColumn{Header: "ID", Width: 36}, Value: func(k *aruba.Key) string { return k.ID() }},
		{TableColumn: TableColumn{Header: "NAME", Width: 30}, Value: func(k *aruba.Key) string { return k.Name() }},
		{TableColumn: TableColumn{Header: "ALGORITHM", Width: 15}, Value: func(k *aruba.Key) string { return string(k.Algorithm()) }},
		{TableColumn: TableColumn{Header: "TYPE", Width: 15}, Value: func(k *aruba.Key) string { return k.Type() }},
		{TableColumn: TableColumn{Header: "STATUS", Width: 15}, Value: func(k *aruba.Key) string { return k.KeyStatus() }},
	}, list.Items(), func(k *aruba.Key) bool { return k.Raw() != nil })
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// SecurityKeyCreateRun is the RunE wiring for key create.
func SecurityKeyCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewSecurityKeyCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKeyCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKeyGetRun is the RunE wiring for key get.
func SecurityKeyGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewSecurityKeyGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKeyGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKeyDeleteRun is the RunE wiring for key delete.
func SecurityKeyDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewSecurityKeyDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("key", a.ID)
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

	if err := SecurityKeyDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// SecurityKeyListRun is the RunE wiring for key list.
func SecurityKeyListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewSecurityKeyListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := SecurityKeyList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
