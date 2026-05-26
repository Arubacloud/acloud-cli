package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
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
	list, err := client.FromSecurity().KMS().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, kms := range list.Items() {
			raw := kms.Raw()
			if raw != nil && raw.Metadata.ID != nil && raw.Metadata.Name != nil {
				id := *raw.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *raw.Metadata.Name))
				}
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

		kms := aruba.NewKMS().
			IntoProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod)).
			ReplaceTags(tags...)

		ctx, cancel := newCtx()
		defer cancel()
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
		kms, err := client.FromSecurity().KMS().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Security/kms/"+kmsID))
		if err != nil {
			return fmt.Errorf("getting KMS: %w", apiErrFromV2(err))
		}

		if kms != nil && kms.Raw() != nil {
			raw := kms.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nKMS Details:")
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
				fmt.Printf("Region:          %s\n", raw.Metadata.LocationResponse.Value)
			}
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", *raw.Status.State)
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
		list, err := client.FromSecurity().KMS().List(ctx, aruba.URI("/projects/"+projectID))
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
			for _, kms := range list.Items() {
				raw := kms.Raw()
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
				region := ""
				if raw.Metadata.LocationResponse != nil {
					region = string(raw.Metadata.LocationResponse.Value)
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, region, status})
			}
			PrintOutput(list.Raw(), headers, rows)
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
		kms, err := client.FromSecurity().KMS().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Security/kms/"+kmsID))
		if err != nil {
			return fmt.Errorf("getting KMS: %w", err)
		}

		if kms == nil || kms.Raw() == nil {
			return fmt.Errorf("KMS not found")
		}

		if name != "" {
			kms.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			kms.ReplaceTags(tags...)
		}

		updated, err := client.FromSecurity().KMS().Update(ctx, kms)
		if err != nil {
			return fmt.Errorf("updating KMS: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("KMS", kmsID))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("KMS", kmsID))
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
			_, err = client.FromSecurity().KMS().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Security/kms/"+kmsID))
			if err != nil {
				return fmt.Errorf("dry-run: KMS not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("KMS", kmsID))
			return nil
		}

		err = client.FromSecurity().KMS().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Security/kms/"+kmsID))
		if err != nil {
			return fmt.Errorf("deleting KMS: %w", err)
		}

		if err := client.FromSecurity().KMS().Delete(ctx, kmsRef(projectID, kmsID)); err != nil {
			return fmt.Errorf("deleting KMS: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDeleted("KMS", kmsID))
		return nil
	},
}
