package cmd

import (
	"context"
	"encoding/json"
	"fmt"
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
	keypairCreateCmd.Flags().String("region", "ITBG-Bergamo", "Region code (required)")
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

// keypairRef builds the combined project+keypair Ref that v0.2.0 Get/Delete need.
func keypairRef(projectID, name string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Compute/keypairs/" + name)
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

		kp := aruba.NewKeyPair().
			InProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			WithPublicKey(publicKey)

		ctx, cancel := newCtx()
		defer cancel()
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
			row := []string{nameVal, publicKeyValue}
			PrintOutput(raw, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Keypair", name))
		}
		return nil
	},
}

var keypairGetCmd = &cobra.Command{
	Use:   "get [keypair-id]",
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
		kp, err := client.FromCompute().KeyPairs().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Compute/keypairs/"+keypairName))
		if err != nil {
			return fmt.Errorf("getting keypair: %w", apiErrFromV2(err))
		}

		if kp != nil && kp.Raw() != nil {
			raw := kp.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
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
			// Show status as 'Active' for consistency
			fmt.Printf("Status:          Active\n")

			if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
				fmt.Printf("Creation Date:   %s\n", raw.Metadata.CreationDate.Format(DateLayout))
			}
			if raw.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
			}

			// Show JSON output if verbose
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				jsonData, _ := json.MarshalIndent(raw, "", "  ")
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
	Use:   "delete [keypair-id]",
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

		keypairRef := aruba.URI("/projects/" + projectID + "/providers/Aruba.Compute/keypairs/" + keypairName)

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromCompute().KeyPairs().Get(ctx, keypairRef)
			if err != nil {
				return fmt.Errorf("dry-run: keypair not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("keypair", keypairName))
			return nil
		}

		err = client.FromCompute().KeyPairs().Delete(ctx, keypairRef)
		if err != nil {
			return fmt.Errorf("deleting keypair: %w", apiErrFromV2(err))
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
		list, err := client.FromCompute().KeyPairs().List(ctx, aruba.URI("/projects/"+projectID))
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
				status := "Active"
				rows = append(rows, []string{name, id, publicKey, status})
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
	},
}
