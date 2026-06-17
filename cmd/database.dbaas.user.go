package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// userRef returns a Ref for a specific user inside a DBaaS instance.
type userGetView struct {
	Username, CreatedAt, CreatedBy string
}

func userRef(projectID, dbaasID, username string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Database/dbaas/" + dbaasID +
		"/users/" + username)
}

func init() {
	dbaasCmd.AddCommand(dbaasUserCmd)
	dbaasUserCmd.AddCommand(dbaasUserCreateCmd)
	dbaasUserCmd.AddCommand(dbaasUserGetCmd)
	dbaasUserCmd.AddCommand(dbaasUserUpdateCmd)
	dbaasUserCmd.AddCommand(dbaasUserDeleteCmd)
	dbaasUserCmd.AddCommand(dbaasUserListCmd)

	dbaasUserCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserCreateCmd.Flags().String("username", "", "Username (required)")
	dbaasUserCreateCmd.Flags().String("password", "", "Password (required)")
	dbaasUserCreateCmd.MarkFlagRequired("username")
	dbaasUserCreateCmd.MarkFlagRequired("password")

	dbaasUserGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasUserUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserUpdateCmd.Flags().String("password", "", "New password (required)")

	dbaasUserDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasUserDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasUserListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUserListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasUserListCmd.Flags().Int("offset", 0, "Number of results to skip")

	dbaasUserGetCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserUpdateCmd.ValidArgsFunction = completeDBaaSUserID
	dbaasUserDeleteCmd.ValidArgsFunction = completeDBaaSUserID
}

func completeDBaaSUserID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	dbaasID := args[0]

	key := cacheKey("dbaas-user", projectID, dbaasID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromDatabase().Users().List(ctx, dbaasRef(projectID, dbaasID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, user := range list.Items() {
			raw := user.Raw()
			if raw != nil && raw.Username != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", raw.Username, raw.Username))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
}

var dbaasUserCmd = &cobra.Command{
	Use:   "user [dbaas-id]",
	Short: "Manage users in DBaaS",
	Long:  `Perform CRUD operations on users in DBaaS.`,
}

var dbaasUserCreateCmd = &cobra.Command{
	Use:   "create [dbaas-id]",
	Short: "Create a new user in DBaaS",
	Long: `Create a new database user inside an existing DBaaS instance.

The user is granted access to the instance. Assign database-level privileges
separately through your database client after the user is created.`,
	Example: `  acloud database dbaas user create <dbaas-id> --username myuser --password mypassword`,
	Args:    cobra.ExactArgs(1),
	RunE:    DatabaseDBaaSUserCreateRun,
}

var dbaasUserGetCmd = &cobra.Command{
	Use:   "get [dbaas-id] [username]",
	Short: "Get user details",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSUserGetRun,
}

var dbaasUserListCmd = &cobra.Command{
	Use:   "list [dbaas-id]",
	Short: "List all users in DBaaS",
	Args:  cobra.ExactArgs(1),
	RunE:  DatabaseDBaaSUserListRun,
}

var dbaasUserUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id] [username]",
	Short: "Update a user (change password)",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSUserUpdateRun,
}

var dbaasUserDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id] [username]",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(2),
	RunE:  DatabaseDBaaSUserDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// DatabaseDBaaSUserCreateArgs holds the typed arguments for creating a user.
type DatabaseDBaaSUserCreateArgs struct {
	ProjectID string
	DBaaSID   string
	Username  string
	Password  string
}

// DatabaseDBaaSUserGetArgs holds the typed arguments for getting a user.
type DatabaseDBaaSUserGetArgs struct {
	ProjectID string
	DBaaSID   string
	Username  string
}

// DatabaseDBaaSUserUpdateArgs holds the typed arguments for updating a user.
type DatabaseDBaaSUserUpdateArgs struct {
	ProjectID string
	DBaaSID   string
	Username  string
	Password  string
}

// DatabaseDBaaSUserDeleteArgs holds the typed arguments for deleting a user.
type DatabaseDBaaSUserDeleteArgs struct {
	ProjectID   string
	DBaaSID     string
	Username    string
	DryRun      bool
	SkipConfirm bool
}

// DatabaseDBaaSUserListArgs holds the typed arguments for listing users.
type DatabaseDBaaSUserListArgs struct {
	ProjectID string
	DBaaSID   string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewDatabaseDBaaSUserCreateArgsFromCobraCommand parses and validates args for create.
func NewDatabaseDBaaSUserCreateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSUserCreateArgs, error) {
	args := &DatabaseDBaaSUserCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSUserGetArgsFromCobraCommand parses and validates args for get.
func NewDatabaseDBaaSUserGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSUserGetArgs, error) {
	args := &DatabaseDBaaSUserGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSUserUpdateArgsFromCobraCommand parses and validates args for update.
func NewDatabaseDBaaSUserUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSUserUpdateArgs, error) {
	args := &DatabaseDBaaSUserUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSUserDeleteArgsFromCobraCommand parses and validates args for delete.
func NewDatabaseDBaaSUserDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSUserDeleteArgs, error) {
	args := &DatabaseDBaaSUserDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewDatabaseDBaaSUserListArgsFromCobraCommand parses and validates args for list.
func NewDatabaseDBaaSUserListArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*DatabaseDBaaSUserListArgs, error) {
	args := &DatabaseDBaaSUserListArgs{}
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

// ParseFromCobraCommand reads Cobra flags and positional args into the create args struct.
func (a *DatabaseDBaaSUserCreateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if a.Username, err = cmd.Flags().GetString("username"); err != nil {
		errs = append(errs, err)
	}
	if a.Password, err = cmd.Flags().GetString("password"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *DatabaseDBaaSUserGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.Username = cobraArgs[1]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *DatabaseDBaaSUserUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.Username = cobraArgs[1]
	}
	if a.Password, err = cmd.Flags().GetString("password"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *DatabaseDBaaSUserDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	if len(cobraArgs) > 1 {
		a.Username = cobraArgs[1]
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the list args struct.
func (a *DatabaseDBaaSUserListArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.DBaaSID = cobraArgs[0]
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *DatabaseDBaaSUserCreateArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.Username == "" {
		errs = append(errs, errors.New("--username is required"))
	}
	if a.Password == "" {
		errs = append(errs, errors.New("--password is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *DatabaseDBaaSUserGetArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.Username == "" {
		errs = append(errs, errors.New("username is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the update args for correctness.
func (a *DatabaseDBaaSUserUpdateArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.Username == "" {
		errs = append(errs, errors.New("username is required"))
	}
	if a.Password == "" {
		errs = append(errs, errors.New("--password is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *DatabaseDBaaSUserDeleteArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}
	if a.Username == "" {
		errs = append(errs, errors.New("username is required"))
	}

	return errors.Join(errs...)
}

// Validate checks the list args for correctness.
func (a *DatabaseDBaaSUserListArgs) Validate() error {
	var errs []error

	if a.DBaaSID == "" {
		errs = append(errs, errors.New("DBaaS ID is required"))
	}

	return errors.Join(errs...)
}

// =============================================================================
// Operation functions
// =============================================================================

// DatabaseDBaaSUserCreate creates a new user inside a DBaaS instance.
func DatabaseDBaaSUserCreate(ctx context.Context, client aruba.Client, args DatabaseDBaaSUserCreateArgs) error {
	user := aruba.NewUser().
		InDBaaS(dbaasRef(args.ProjectID, args.DBaaSID)).
		WithUsername(args.Username).
		WithPassword(args.Password)

	created, err := client.FromDatabase().Users().Create(ctx, user)
	if err != nil {
		return fmt.Errorf("creating user: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		fmt.Printf("\n%s\n", msgCreated("User", args.Username))
		fmt.Printf("Username:        %s\n", raw.Username)
		if raw.CreationDate != nil {
			fmt.Printf("Creation Date:   %s\n", raw.CreationDate.Format(DateLayout))
		}
	} else {
		fmt.Println(msgCreatedAsync("User", args.Username))
	}
	return nil
}

// DatabaseDBaaSUserGet retrieves and displays a user's details.
func DatabaseDBaaSUserGet(ctx context.Context, client aruba.Client, args DatabaseDBaaSUserGetArgs) error {
	u, err := client.FromDatabase().Users().Get(ctx, userRef(args.ProjectID, args.DBaaSID, args.Username))
	if err != nil {
		return fmt.Errorf("getting user: %w", apiErrFromV2(err))
	}

	if u == nil || u.Raw() == nil {
		fmt.Println("User not found")
		return nil
	}
	format := resolveOutputFormat()
	if format == OutputFormatJSON || format == OutputFormatYAML {
		PrintOutput(u, nil, nil)
		return nil
	}
	raw := u.Raw()
	view := userGetView{Username: raw.Username}
	if raw.CreationDate != nil {
		view.CreatedAt = raw.CreationDate.Format(DateLayout)
	}
	if raw.CreatedBy != nil {
		view.CreatedBy = *raw.CreatedBy
	}
	return renderGet(userGetTmpl, view)
}

// DatabaseDBaaSUserUpdate updates a user's password.
func DatabaseDBaaSUserUpdate(ctx context.Context, client aruba.Client, args DatabaseDBaaSUserUpdateArgs) error {
	u, err := client.FromDatabase().Users().Get(ctx, userRef(args.ProjectID, args.DBaaSID, args.Username))
	if err != nil {
		return fmt.Errorf("getting user: %w", apiErrFromV2(err))
	}
	if u == nil || u.Raw() == nil {
		return fmt.Errorf("user not found")
	}

	u.WithPassword(args.Password)

	updated, err := client.FromDatabase().Users().Update(ctx, u)
	if err != nil {
		return fmt.Errorf("updating user: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		fmt.Printf("\n%s\n", msgUpdated("User", args.Username))
		fmt.Printf("Username:        %s\n", updated.Raw().Username)
	} else {
		fmt.Println(msgUpdatedAsync("User", args.Username))
	}
	return nil
}

// DatabaseDBaaSUserDelete deletes a user from a DBaaS instance.
func DatabaseDBaaSUserDelete(ctx context.Context, client aruba.Client, args DatabaseDBaaSUserDeleteArgs) error {
	err := client.FromDatabase().Users().Delete(ctx, userRef(args.ProjectID, args.DBaaSID, args.Username))
	if err != nil {
		return fmt.Errorf("deleting user: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("User", args.Username))
	return nil
}

// DatabaseDBaaSUserList lists all users in a DBaaS instance.
func DatabaseDBaaSUserList(ctx context.Context, client aruba.Client, args DatabaseDBaaSUserListArgs) error {
	list, err := client.FromDatabase().Users().List(ctx, dbaasRef(args.ProjectID, args.DBaaSID), args.CallOpts...)
	if err != nil {
		return fmt.Errorf("listing users: %w", apiErrFromV2(err))
	}

	if list == nil || len(list.Items()) == 0 {
		fmt.Println("No users found")
		return nil
	}
	renderList(list, []ListColumn[*aruba.User]{
		{TableColumn: TableColumn{Header: "USERNAME", Width: 40}, Value: func(u *aruba.User) string {
			if r := u.Raw(); r != nil {
				return r.Username
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "CREATION DATE", Width: 25}, Value: func(u *aruba.User) string {
			if r := u.Raw(); r != nil && r.CreationDate != nil {
				return r.CreationDate.Format(DateLayout)
			}
			return ""
		}},
		{TableColumn: TableColumn{Header: "CREATED BY", Width: 30}, Value: func(u *aruba.User) string {
			if r := u.Raw(); r != nil && r.CreatedBy != nil {
				return *r.CreatedBy
			}
			return ""
		}},
	}, list.Items(), func(u *aruba.User) bool { return u.Raw() != nil })
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// DatabaseDBaaSUserCreateRun is the Cobra RunE handler for user create.
func DatabaseDBaaSUserCreateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSUserCreateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSUserCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSUserGetRun is the Cobra RunE handler for user get.
func DatabaseDBaaSUserGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSUserGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSUserGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSUserUpdateRun is the Cobra RunE handler for user update.
func DatabaseDBaaSUserUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSUserUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSUserUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSUserDeleteRun is the Cobra RunE handler for user delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func DatabaseDBaaSUserDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSUserDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete(fmt.Sprintf("user '%s' in DBaaS instance", args.Username), args.DBaaSID)
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

	if args.DryRun {
		if _, err := client.FromDatabase().Users().Get(ctx, userRef(args.ProjectID, args.DBaaSID, args.Username)); err != nil {
			return fmt.Errorf("dry-run: database user not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("database user", args.Username))
		return nil
	}

	if err := DatabaseDBaaSUserDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// DatabaseDBaaSUserListRun is the Cobra RunE handler for user list.
func DatabaseDBaaSUserListRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewDatabaseDBaaSUserListArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := DatabaseDBaaSUserList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
