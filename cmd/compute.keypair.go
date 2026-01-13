package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
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
	keypairCreateCmd.MarkFlagRequired("name")
	keypairCreateCmd.MarkFlagRequired("public-key")

	keypairGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	keypairUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairUpdateCmd.Flags().String("public-key", "", "New public key value")

	keypairDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	keypairListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	keypairGetCmd.ValidArgsFunction = completeKeyPairID
	keypairUpdateCmd.ValidArgsFunction = completeKeyPairID
	keypairDeleteCmd.ValidArgsFunction = completeKeyPairID
}

// Completion functions for keypair resources
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
	response, err := client.FromCompute().KeyPairs().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, keypair := range response.Data.Values {
			if keypair.Metadata.Name != nil {
				name := *keypair.Metadata.Name
				if toComplete == "" || strings.HasPrefix(name, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tKeypair", name))
				}
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
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		publicKey, _ := cmd.Flags().GetString("public-key")

		if name == "" || publicKey == "" {
			fmt.Println("Error: --name and --public-key are required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build the create request
		// KeyPairRequest may use RegionalResourceMetadataRequest even though keypairs don't have regions
		createRequest := types.KeyPairRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
				},
				// Location may be optional for keypairs
			},
			Properties: types.KeyPairPropertiesRequest{
				Value: publicKey,
			},
		}

		ctx := context.Background()
		response, err := client.FromCompute().KeyPairs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating keypair: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create keypair - Status: %d\n", response.StatusCode)
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
				{Header: "NAME", Width: 40},
				{Header: "PUBLIC_KEY", Width: 60},
			}
			publicKeyValue := response.Data.Properties.Value
			if len(publicKeyValue) > 50 {
				publicKeyValue = publicKeyValue[:50] + "..."
			}
			row := []string{
				func() string {
					if response.Data.Metadata.Name != nil {
						return *response.Data.Metadata.Name
					}
					return ""
				}(),
				publicKeyValue,
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Keypair created, but no data returned.")
		}
	},
}

var keypairGetCmd = &cobra.Command{
	Use:   "get [keypair-name]",
	Short: "Get keypair details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keypairName := args[0]

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
		resp, err := client.FromCompute().KeyPairs().Get(ctx, projectID, keypairName, nil)
		if err != nil {
			fmt.Printf("Error getting keypair: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get keypair - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			keypair := resp.Data

			fmt.Println("\nKeypair Details:")
			fmt.Println("===============")

			if keypair.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *keypair.Metadata.Name)
			}
			if keypair.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *keypair.Metadata.URI)
			}
			if keypair.Properties.Value != "" {
				fmt.Printf("Public Key:      %s\n", keypair.Properties.Value)
			}

			if !keypair.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", keypair.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if keypair.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *keypair.Metadata.CreatedBy)
			}

			// Show JSON output if verbose
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				jsonData, _ := json.MarshalIndent(keypair, "", "  ")
				fmt.Println("\nFull JSON Response:")
				fmt.Println("==================")
				fmt.Println(string(jsonData))
			}
		} else {
			fmt.Println("Keypair not found or no data returned.")
		}
	},
}

var keypairUpdateCmd = &cobra.Command{
	Use:   "update [keypair-name]",
	Short: "Update a keypair (not supported - delete and recreate instead)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Error: Keypair update is not supported by the API.")
		fmt.Println("To change a keypair's public key, delete it and create a new one with the same name.")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Printf("  acloud compute keypair delete %s --yes\n", args[0])
		fmt.Printf("  acloud compute keypair create --name %s --public-key \"<new-key>\"\n", args[0])
	},
}

var keypairDeleteCmd = &cobra.Command{
	Use:   "delete [keypair-name]",
	Short: "Delete a keypair",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keypairName := args[0]

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

		// Confirmation prompt
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete keypair '%s'? (yes/no): ", keypairName)
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "yes" && confirmation != "y" {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		ctx := context.Background()
		response, err := client.FromCompute().KeyPairs().Delete(ctx, projectID, keypairName, nil)
		if err != nil {
			fmt.Printf("Error deleting keypair: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to delete keypair - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		fmt.Printf("Keypair '%s' deleted successfully.\n", keypairName)
	},
}

var keypairListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keypairs",
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
		response, err := client.FromCompute().KeyPairs().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing keypairs: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 30},
				{Header: "PUBLIC_KEY", Width: 60},
			}

			// Extract IDs from raw JSON response if available
			// The SDK type definition uses Request types but actual response has ID fields
			idMap := make(map[int]string) // Map keypair index to ID
			if response.RawBody != nil {
				var rawResponse map[string]interface{}
				if err := json.Unmarshal(response.RawBody, &rawResponse); err == nil {
					if values, ok := rawResponse["values"].([]interface{}); ok {
						for i, val := range values {
							if keypairMap, ok := val.(map[string]interface{}); ok {
								if metadata, ok := keypairMap["metadata"].(map[string]interface{}); ok {
									if idVal, ok := metadata["id"].(string); ok && idVal != "" {
										idMap[i] = idVal
									}
								}
							}
						}
					}
				}
			}

			var rows [][]string
			for idx, keypair := range response.Data.Values {
				name := ""
				if keypair.Metadata.Name != nil {
					name = *keypair.Metadata.Name
				}

				// Get ID from raw JSON map, fallback to name
				id := idMap[idx]
				if id == "" {
					id = name
				}

				publicKey := ""
				if keypair.Properties.Value != "" {
					publicKey = keypair.Properties.Value
					if len(publicKey) > 50 {
						publicKey = publicKey[:50] + "..."
					}
				}

				rows = append(rows, []string{name, id, publicKey})
			}

			PrintTable(headers, rows)
		} else {
			fmt.Println("No keypairs found")
		}
	},
}
