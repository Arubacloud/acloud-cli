package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	// KeyPair commands
	computeCmd.AddCommand(keypairCmd)
	keypairCmd.AddCommand(keypairCreateCmd)
	keypairCmd.AddCommand(keypairGetCmd)
	// Note: Update is not supported by the API, but we keep the command for user guidance
	keypairCmd.AddCommand(keypairUpdateCmd)
	keypairCmd.AddCommand(keypairDeleteCmd)
	keypairCmd.AddCommand(keypairListCmd)

	// Add flags for keypair commands
	keypairCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairCreateCmd.Flags().String("name", "", "Name for the keypair (required)")
	keypairCreateCmd.Flags().String("public-key", "", "Public key value (required)")
	keypairCreateCmd.Flags().String("region", "", "Region code (required, e.g. IT-BG)")
	keypairCreateCmd.MarkFlagRequired("name")
	keypairCreateCmd.MarkFlagRequired("public-key")
	keypairCreateCmd.MarkFlagRequired("region")

	keypairGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	keypairUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairUpdateCmd.Flags().String("public-key", "", "New public key value")

	keypairDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	keypairDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	keypairListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	keypairListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	keypairGetCmd.ValidArgsFunction = completeKeyPairID
	keypairUpdateCmd.ValidArgsFunction = completeKeyPairID
	keypairDeleteCmd.ValidArgsFunction = completeKeyPairID
}

// keypairRef builds the combined project+keypair Ref that v0.2.0 Get/Delete need.
func keypairRef(projectID, name string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Compute/keyPairs/" + name)
}

// completeKeyPairID provides shell completion for keypair names.
func completeKeyPairID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromCompute().KeyPairs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, kp := range list.Items() {
			name := kp.Name()
			if toComplete == "" || strings.HasPrefix(name, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\tKeypair", name))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// KeyPair subcommands
var keypairCmd = &cobra.Command{
	Use:   "keypair",
	Short: "Manage keypairs",
	Long:  `Perform CRUD operations on keypairs in Aruba Cloud.`,
}

var keypairCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new keypair",
	Long: `Create a new SSH keypair by uploading your public key.

The public key must be an OpenSSH-formatted RSA, ECDSA, or Ed25519 public key
(the content of your ~/.ssh/id_rsa.pub or similar file).`,
	Example: `  acloud compute keypair create --name my-key --public-key "$(cat ~/.ssh/id_rsa.pub)"`,
	Args:    cobra.NoArgs,
	RunE:    ComputeKeyPairCreateRun,
}

var keypairGetCmd = &cobra.Command{
	Use:   "get [keypair-id]",
	Short: "Get keypair details",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeKeyPairGetRun,
}

var keypairUpdateCmd = &cobra.Command{
	Use:   "update [keypair-name]",
	Short: "Update a keypair (not supported - delete and recreate instead)",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeKeyPairUpdateRun,
}

var keypairDeleteCmd = &cobra.Command{
	Use:   "delete [keypair-id]",
	Short: "Delete a keypair",
	Args:  cobra.ExactArgs(1),
	RunE:  ComputeKeyPairDeleteRun,
}

var keypairListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keypairs",
	Args:  cobra.NoArgs,
	RunE:  ComputeKeyPairListRun,
}

// =============================================================================
// Args structs
// =============================================================================

// ComputeKeyPairCreateArgs holds the typed arguments for creating a keypair.
type ComputeKeyPairCreateArgs struct {
	ProjectID string
	Name      string
	Region    aruba.Region
	PublicKey string
}

// ComputeKeyPairGetArgs holds the typed arguments for getting a keypair.
type ComputeKeyPairGetArgs struct {
	ProjectID string
	ID        string
}

// ComputeKeyPairUpdateArgs holds the typed arguments for the update stub.
type ComputeKeyPairUpdateArgs struct {
	ProjectID string
	ID        string
}

// ComputeKeyPairDeleteArgs holds the typed arguments for deleting a keypair.
type ComputeKeyPairDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// ComputeKeyPairListArgs holds the typed arguments for listing keypairs.
type ComputeKeyPairListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewComputeKeyPairCreateArgsFromCobraCommand parses and validates args for create.
func NewComputeKeyPairCreateArgsFromCobraCommand(cmd *cobra.Command) (*ComputeKeyPairCreateArgs, error) {
	args := &ComputeKeyPairCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeKeyPairGetArgsFromCobraCommand parses and validates args for get.
func NewComputeKeyPairGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeKeyPairGetArgs, error) {
	args := &ComputeKeyPairGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeKeyPairUpdateArgsFromCobraCommand parses and validates args for the update stub.
func NewComputeKeyPairUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeKeyPairUpdateArgs, error) {
	args := &ComputeKeyPairUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeKeyPairDeleteArgsFromCobraCommand parses and validates args for delete.
func NewComputeKeyPairDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*ComputeKeyPairDeleteArgs, error) {
	args := &ComputeKeyPairDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewComputeKeyPairListArgsFromCobraCommand parses and validates args for list.
func NewComputeKeyPairListArgsFromCobraCommand(cmd *cobra.Command) (*ComputeKeyPairListArgs, error) {
	args := &ComputeKeyPairListArgs{}
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
func (a *ComputeKeyPairCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
	if a.PublicKey, err = cmd.Flags().GetString("public-key"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *ComputeKeyPairGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the update stub args struct.
func (a *ComputeKeyPairUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *ComputeKeyPairDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
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
func (a *ComputeKeyPairListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
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
func (a *ComputeKeyPairCreateArgs) Validate() error {
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
	if a.PublicKey == "" {
		errs = append(errs, errors.New("--public-key is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *ComputeKeyPairGetArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("keypair ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the update stub args for correctness.
func (a *ComputeKeyPairUpdateArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("keypair ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *ComputeKeyPairDeleteArgs) Validate() error {
	var errs []error

	if a.ID == "" {
		errs = append(errs, errors.New("keypair ID is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *ComputeKeyPairListArgs) Validate() error {
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// ComputeKeyPairCreate creates a new keypair.
func ComputeKeyPairCreate(ctx context.Context, client aruba.Client, args ComputeKeyPairCreateArgs) error {
	kp := aruba.NewKeyPair().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		WithPublicKey(args.PublicKey)

	resp, err := client.FromCompute().KeyPairs().Create(ctx, kp)
	if err != nil {
		return fmt.Errorf("creating keypair: %w", apiErrFromV2(err))
	}

	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		headers := []TableColumn{
			{Header: "NAME", Width: 40},
			{Header: "ID", Width: 26},
			{Header: "PUBLIC_KEY", Width: 60},
			{Header: "STATUS", Width: 10},
		}
		publicKeyValue := ""
		if raw.Properties.Value != "" {
			publicKeyValue = raw.Properties.Value
			if len(publicKeyValue) > 50 {
				publicKeyValue = publicKeyValue[:50] + "..."
			}
		}
		nameVal := ""
		if raw.Metadata.Name != nil {
			nameVal = *raw.Metadata.Name
		}
		id := ""
		if raw.Metadata.ID != nil {
			id = *raw.Metadata.ID
		}
		PrintOutput(resp, headers, [][]string{{nameVal, id, publicKeyValue, "Active"}})
	} else {
		fmt.Println(msgCreatedAsync("Keypair", args.Name))
	}
	return nil
}

// ComputeKeyPairGet retrieves keypair details.
func ComputeKeyPairGet(ctx context.Context, client aruba.Client, args ComputeKeyPairGetArgs) error {
	kp, err := client.FromCompute().KeyPairs().Get(ctx, keypairRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting keypair: %w", apiErrFromV2(err))
	}

	if kp != nil && kp.Raw() != nil {
		raw := kp.Raw()

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(kp, nil, nil)
			return nil
		}

		fmt.Println("\nKeypair Details:")
		fmt.Println("===============")

		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Properties.Value != "" {
			fmt.Printf("Public Key:      %s\n", raw.Properties.Value)
		}
		fmt.Printf("Status:          Active\n")
		if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
			fmt.Printf("Creation Date:   %s\n", raw.Metadata.CreationDate.Format(DateLayout))
		}
		if raw.Metadata.CreatedBy != nil {
			fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
		}
	} else {
		fmt.Println("Keypair not found or no data returned.")
	}
	return nil
}

// ComputeKeyPairUpdate prints a "not supported" message (the API does not support keypair updates).
func ComputeKeyPairUpdate(_ context.Context, _ aruba.Client, args ComputeKeyPairUpdateArgs) error {
	fmt.Println("Error: Keypair update is not supported by the API.")
	fmt.Println("To change a keypair's public key, delete it and create a new one with the same name.")
	fmt.Println("")
	fmt.Println("Example:")
	fmt.Printf("  acloud compute keypair delete %s --yes\n", args.ID)
	fmt.Printf("  acloud compute keypair create --name %s --public-key \"<new-key>\"\n", args.ID)
	return nil
}

// ComputeKeyPairDelete deletes a keypair.
func ComputeKeyPairDelete(ctx context.Context, client aruba.Client, args ComputeKeyPairDeleteArgs) error {
	err := client.FromCompute().KeyPairs().Delete(ctx, keypairRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting keypair: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("Keypair", args.ID))
	return nil
}

// ComputeKeyPairList lists all keypairs in a project.
func ComputeKeyPairList(ctx context.Context, client aruba.Client, args ComputeKeyPairListArgs) error {
	list, err := client.FromCompute().KeyPairs().List(ctx, projectRef(args.ProjectID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing keypairs: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 40},
			{Header: "ID", Width: 30},
			{Header: "PUBLIC_KEY", Width: 60},
			{Header: "STATUS", Width: 10},
		}

		var rows [][]string
		for _, kp := range list.Items() {
			raw := kp.Raw()
			if raw == nil {
				continue
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			if id == "" {
				continue
			}
			name := ""
			if raw.Metadata.Name != nil {
				name = *raw.Metadata.Name
			}
			publicKey := ""
			if raw.Properties.Value != "" {
				publicKey = raw.Properties.Value
				if len(publicKey) > 50 {
					publicKey = publicKey[:50] + "..."
				}
			}
			rows = append(rows, []string{name, id, publicKey, "Active"})
		}

		if len(rows) == 0 {
			fmt.Println("No keypairs found")
			return nil
		}
		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No keypairs found")
	}
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// ComputeKeyPairCreateRun is the RunE wiring for keypair create.
func ComputeKeyPairCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewComputeKeyPairCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeKeyPairCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeKeyPairGetRun is the RunE wiring for keypair get.
func ComputeKeyPairGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeKeyPairGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeKeyPairGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeKeyPairUpdateRun is the RunE wiring for the keypair update stub.
func ComputeKeyPairUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewComputeKeyPairUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	// Update is a stub — no client needed.
	return ComputeKeyPairUpdate(context.Background(), nil, *args)
}

// ComputeKeyPairDeleteRun is the RunE wiring for keypair delete.
func ComputeKeyPairDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	a, err := NewComputeKeyPairDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !a.SkipConfirm {
		ok, err := confirmDelete("keypair", a.ID)
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

	if a.DryRun {
		_, err = client.FromCompute().KeyPairs().Get(ctx, keypairRef(a.ProjectID, a.ID))
		if err != nil {
			return fmt.Errorf("dry-run: keypair not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("keypair", a.ID))
		return nil
	}

	if err := ComputeKeyPairDelete(ctx, client, *a); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// ComputeKeyPairListRun is the RunE wiring for keypair list.
func ComputeKeyPairListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewComputeKeyPairListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := ComputeKeyPairList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
