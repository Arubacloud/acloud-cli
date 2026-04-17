package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

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
	response, err := client.FromSecurity().KMS().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, kms := range response.Data.Values {
			if kms.Metadata.ID != nil && kms.Metadata.Name != nil {
				id := *kms.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *kms.Metadata.Name))
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

		createRequest := types.KmsRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.KmsPropertiesRequest{
				BillingPeriod: billingPeriod,
			},
		}

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromSecurity().KMS().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating KMS: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
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
					if response.Data.Metadata.LocationResponse != nil {
						return response.Data.Metadata.LocationResponse.Value
					}
					return ""
				}(),
				func() string {
					if response.Data.Status.State != nil {
						return *response.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
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
		resp, err := client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
		if err != nil {
			return fmt.Errorf("getting KMS: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil {
			kms := resp.Data

			fmt.Println("\nKMS Details:")
			fmt.Println("============")

			if kms.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *kms.Metadata.ID)
			}
			if kms.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *kms.Metadata.URI)
			}
			if kms.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *kms.Metadata.Name)
			}
			if kms.Metadata.LocationResponse != nil {
				if kms.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", kms.Metadata.LocationResponse.Value)
				}
			}
			if kms.Status.State != nil {
				fmt.Printf("Status:          %s\n", *kms.Status.State)
			}
			if kms.Metadata.CreationDate != nil && !kms.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", kms.Metadata.CreationDate.Format(DateLayout))
			}
			if kms.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *kms.Metadata.CreatedBy)
			}
			if len(kms.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", kms.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			fmt.Println()
		} else {
			fmt.Println("KMS not found")
		}
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
		resp, err := client.FromSecurity().KMS().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing KMS: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, kms := range resp.Data.Values {
				row := []string{
					func() string {
						if kms.Metadata.Name != nil {
							return *kms.Metadata.Name
						}
						return ""
					}(),
					func() string {
						if kms.Metadata.ID != nil {
							return *kms.Metadata.ID
						}
						return ""
					}(),
					func() string {
						if kms.Metadata.LocationResponse != nil {
							return kms.Metadata.LocationResponse.Value
						}
						return ""
					}(),
					func() string {
						if kms.Status.State != nil {
							return *kms.Status.State
						}
						return ""
					}(),
				}
				rows = append(rows, row)
			}
			PrintTable(headers, rows)
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
		getResp, err := client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
		if err != nil {
			return fmt.Errorf("getting KMS: %w", err)
		}

		if getResp == nil || getResp.Data == nil {
			return fmt.Errorf("KMS not found")
		}

		current := getResp.Data

		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for KMS")
		}

		updateRequest := types.KmsRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *current.Metadata.Name,
					Tags: current.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.KmsPropertiesRequest{
				BillingPeriod: current.Properties.BillingPeriod,
			},
		}

		if name != "" {
			updateRequest.Metadata.ResourceMetadataRequest.Name = name
		}

		if cmd.Flags().Changed("tags") {
			updateRequest.Metadata.ResourceMetadataRequest.Tags = tags
		}

		response, err := client.FromSecurity().KMS().Update(ctx, projectID, kmsID, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating KMS: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("KMS", kmsID))
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
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

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			ok, err := confirmDelete("KMS", kmsID)
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
			_, err = client.FromSecurity().KMS().Get(ctx, projectID, kmsID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: KMS not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("KMS", kmsID))
			return nil
		}

		_, err = client.FromSecurity().KMS().Delete(ctx, projectID, kmsID, nil)
		if err != nil {
			return fmt.Errorf("deleting KMS: %w", err)
		}

		fmt.Println(msgDeleted("KMS", kmsID))
		return nil
	},
}
