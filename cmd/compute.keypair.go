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
	keypairDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	keypairListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	keypairListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	keypairListCmd.Flags().Int32("offset", 0, "Number of results to skip")

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
	Long: `Create a new SSH keypair by uploading your public key.

The public key must be an OpenSSH-formatted RSA, ECDSA, or Ed25519 public key
(the content of your ~/.ssh/id_rsa.pub or similar file).`,
	Example: `  acloud compute keypair create --name my-key --public-key "$(cat ~/.ssh/id_rsa.pub)"`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		publicKey, _ := cmd.Flags().GetString("public-key")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
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

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromCompute().KeyPairs().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating keypair: %w", err)
		}

		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
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
			PrintOutput(response.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Keypair", name))
		}
		return nil
	},
}

var keypairGetCmd = &cobra.Command{
	Use:   "get [keypair-name]",
	Short: "Get keypair details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keypairName := args[0]

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
		resp, err := client.FromCompute().KeyPairs().Get(ctx, projectID, keypairName, nil)
		if err != nil {
			return fmt.Errorf("getting keypair: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp != nil && resp.Data != nil {
			keypair := resp.Data

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(keypair, nil, nil)
				return nil
			}

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
			// Show status as 'Active' for consistency
			fmt.Printf("Status:          Active\n")

			if keypair.Metadata.CreationDate != nil && !keypair.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", keypair.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var keypairUpdateCmd = &cobra.Command{
	Use:   "update [keypair-name]",
	Short: "Update a keypair (not supported - delete and recreate instead)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Error: Keypair update is not supported by the API.")
		fmt.Println("To change a keypair's public key, delete it and create a new one with the same name.")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Printf("  acloud compute keypair delete %s --yes\n", args[0])
		fmt.Printf("  acloud compute keypair create --name %s --public-key \"<new-key>\"\n", args[0])
		return nil
	},
}

var keypairDeleteCmd = &cobra.Command{
	Use:   "delete [keypair-name]",
	Short: "Delete a keypair",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keypairName := args[0]

		// Confirmation prompt
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("keypair", keypairName)
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
			_, err = client.FromCompute().KeyPairs().Get(ctx, projectID, keypairName, nil)
			if err != nil {
				return fmt.Errorf("dry-run: key pair not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("key pair", keypairName))
			return nil
		}

		response, err := client.FromCompute().KeyPairs().Delete(ctx, projectID, keypairName, nil)
		if err != nil {
			return fmt.Errorf("deleting keypair: %w", err)
		}

		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		fmt.Println(msgDeleted("Keypair", keypairName))
		return nil
	},
}

var keypairListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keypairs",
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
		response, err := client.FromCompute().KeyPairs().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing keypairs: %w", err)
		}
		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 30},
				{Header: "PUBLIC_KEY", Width: 60},
				{Header: "STATUS", Width: 10},
			}

			var rows [][]string
			for _, keypair := range response.Data.Values {
				if keypair.Metadata.ID == nil || *keypair.Metadata.ID == "" {
					continue
				}
				id := *keypair.Metadata.ID
				name := ""
				if keypair.Metadata.Name != nil {
					name = *keypair.Metadata.Name
				}
				publicKey := ""
				if keypair.Properties.Value != "" {
					publicKey = keypair.Properties.Value
					if len(publicKey) > 50 {
						publicKey = publicKey[:50] + "..."
					}
				}
				status := "Active"
				rows = append(rows, []string{name, id, publicKey, status})
			}

			if len(rows) == 0 {
				fmt.Println("No keypairs found")
				return nil
			}
			PrintOutput(response.Data, headers, rows)
		} else {
			fmt.Println("No keypairs found")
		}
		return nil
	},
}
