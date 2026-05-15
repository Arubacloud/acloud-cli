package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func kmsRef(projectID, kmsID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Security/kms/" + kmsID)
}

func kmsFromRaw(k *aruba.KMS) *types.KmsResponse { return k.Raw() }

func kmsListPayload(l *aruba.List[*aruba.KMS]) any {
	if r, ok := l.Raw().(*types.Response[types.KmsList]); ok && r != nil {
		return r.Data
	}
	return nil
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
	kmsCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	kmsCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	kmsCreateCmd.MarkFlagRequired("name")
	kmsCreateCmd.MarkFlagRequired("region")

	kmsGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	kmsUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	kmsUpdateCmd.Flags().String("name", "", "New KMS name")
	kmsUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

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

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromSecurity().KMS().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, k := range list.Items() {
			id := k.KMSID()
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, k.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
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
	Example: `  acloud security kms create --name my-kms --region IT-BG
  acloud security kms create --name prod-kms --region IT-BG --billing-period Month`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		k := aruba.NewKMS().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		if len(tags) > 0 {
			k.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromSecurity().KMS().Create(ctx, k)
		if err != nil {
			return fmt.Errorf("creating KMS: %w", apiErrFromV2(err))
		}

		if created.Raw() != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}
			raw := created.Raw()
			row := []string{
				created.KMSID(),
				created.Name(),
				string(created.Region()),
				created.State(),
			}
			PrintOutput(raw, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("KMS", name))
		}
		return nil
	},
}

var kmsGetCmd = &cobra.Command{
	Use:   "get [kms-id]",
	Short: "Get KMS resource details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kmsID := args[0]

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
		got, err := client.FromSecurity().KMS().Get(ctx, kmsRef(projectID, kmsID))
		if err != nil {
			return fmt.Errorf("getting KMS: %w", apiErrFromV2(err))
		}

		format := resolveOutputFormat()
		if format == OutputFormatJSON || format == OutputFormatYAML {
			PrintOutput(kmsFromRaw(got), nil, nil)
			return nil
		}

		raw := got.Raw()
		if raw == nil {
			fmt.Println("KMS not found")
			return nil
		}

		fmt.Println("\nKMS Details:")
		fmt.Println("============")
		fmt.Printf("ID:              %s\n", got.KMSID())
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		fmt.Printf("Name:            %s\n", got.Name())
		if raw.Metadata.LocationResponse != nil {
			fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
		}
		if state := got.State(); state != "" {
			fmt.Printf("Status:          %s\n", state)
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
		return nil
	},
}

var kmsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KMS resources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		list, err := client.FromSecurity().KMS().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing KMS: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, k := range list.Items() {
				rows = append(rows, []string{
					k.Name(),
					k.KMSID(),
					string(k.Region()),
					k.State(),
				})
			}
			PrintOutput(kmsListPayload(list), headers, rows)
		} else {
			fmt.Println("No KMS resources found")
		}
		return nil
	},
}

var kmsUpdateCmd = &cobra.Command{
	Use:   "update [kms-id]",
	Short: "Update a KMS resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kmsID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		current, err := client.FromSecurity().KMS().Get(ctx, kmsRef(projectID, kmsID))
		if err != nil {
			return fmt.Errorf("getting KMS: %w", apiErrFromV2(err))
		}

		if name != "" {
			current.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			current.ReplaceTags(tags...)
		}

		updated, err := client.FromSecurity().KMS().Update(ctx, current)
		if err != nil {
			return fmt.Errorf("updating KMS: %w", apiErrFromV2(err))
		}

		fmt.Printf("\n%s\n", msgUpdated("KMS", kmsID))
		fmt.Printf("ID:              %s\n", updated.KMSID())
		fmt.Printf("Name:            %s\n", updated.Name())
		if raw := updated.Raw(); raw != nil && len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		}
		return nil
	},
}

var kmsDeleteCmd = &cobra.Command{
	Use:   "delete [kms-id]",
	Short: "Delete a KMS resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kmsID := args[0]

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
			if _, err := client.FromSecurity().KMS().Get(ctx, kmsRef(projectID, kmsID)); err != nil {
				return fmt.Errorf("dry-run: KMS not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("KMS", kmsID))
			return nil
		}

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("KMS", kmsID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := client.FromSecurity().KMS().Delete(ctx, kmsRef(projectID, kmsID)); err != nil {
			return fmt.Errorf("deleting KMS: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDeleted("KMS", kmsID))
		return nil
	},
}
