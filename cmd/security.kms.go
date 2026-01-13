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

	kmsListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

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
	response, err := client.FromSecurity().KMSKeys().List(ctx, projectID, nil)
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
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" || region == "" {
			fmt.Println("Error: --name and --region are required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
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
				BillingPeriod: types.BillingPeriodResource{
					BillingPeriod: billingPeriod,
				},
			},
		}

		ctx := context.Background()
		response, err := client.FromSecurity().KMSKeys().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating KMS: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create KMS - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
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
			fmt.Println("KMS created, but no data returned.")
		}
	},
}

var kmsGetCmd = &cobra.Command{
	Use:   "get [kms-id]",
	Short: "Get KMS resource details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kmsID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromSecurity().KMSKeys().Get(ctx, projectID, kmsID, nil)
		if err != nil {
			fmt.Printf("Error getting KMS: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get KMS - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
				fmt.Printf("Region:          %s\n", kms.Metadata.LocationResponse.Value)
			}
			if kms.Status.State != nil {
				fmt.Printf("Status:          %s\n", *kms.Status.State)
			}
			if !kms.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", kms.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}

var kmsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all KMS resources",
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromSecurity().KMSKeys().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing KMS: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list KMS - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
	},
}

var kmsUpdateCmd = &cobra.Command{
	Use:   "update [kms-id]",
	Short: "Update a KMS resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kmsID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		getResp, err := client.FromSecurity().KMSKeys().Get(ctx, projectID, kmsID, nil)
		if err != nil {
			fmt.Printf("Error getting KMS: %v\n", err)
			return
		}

		if getResp == nil || getResp.Data == nil {
			fmt.Println("KMS not found")
			return
		}

		current := getResp.Data

		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for KMS")
			return
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

		response, err := client.FromSecurity().KMSKeys().Update(ctx, projectID, kmsID, updateRequest, nil)
		if err != nil {
			fmt.Printf("Error updating KMS: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update KMS - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nKMS updated successfully!")
			fmt.Printf("ID:              %s\n", *response.Data.Metadata.ID)
			fmt.Printf("Name:            %s\n", *response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("Warning: Update may have succeeded but response is empty")
		}
	},
}

var kmsDeleteCmd = &cobra.Command{
	Use:   "delete [kms-id]",
	Short: "Delete a KMS resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		kmsID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		confirm, _ := cmd.Flags().GetBool("yes")

		if !confirm {
			fmt.Printf("Are you sure you want to delete KMS %s? (yes/no): ", kmsID)
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		_, err = client.FromSecurity().KMSKeys().Delete(ctx, projectID, kmsID, nil)
		if err != nil {
			fmt.Printf("Error deleting KMS: %v\n", err)
			return
		}

		fmt.Printf("\nKMS %s deleted successfully!\n", kmsID)
	},
}
