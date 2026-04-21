package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	cloudserverCreateCmd.Flags().String("boot-disk-uri", "", "Bootable block storage URI (required, e.g., /projects/{project-id}/providers/Aruba.Storage/blockStorages/{volume-id})")
	cloudserverCreateCmd.MarkFlagRequired("boot-disk-uri")
	cloudserverCreateCmd.MarkFlagRequired("vpc-uri")
	cloudserverCreateCmd.MarkFlagRequired("subnet-uri")
	cloudserverCreateCmd.MarkFlagRequired("security-group-uri")
	// CloudServer commands
	computeCmd.AddCommand(cloudserverCmd)
	cloudserverCmd.AddCommand(cloudserverCreateCmd)
	cloudserverCmd.AddCommand(cloudserverGetCmd)
	cloudserverCmd.AddCommand(cloudserverUpdateCmd)
	cloudserverCmd.AddCommand(cloudserverDeleteCmd)
	cloudserverCmd.AddCommand(cloudserverListCmd)
	cloudserverCmd.AddCommand(cloudserverPowerOnCmd)
	cloudserverCmd.AddCommand(cloudserverPowerOffCmd)
	cloudserverCmd.AddCommand(cloudserverSetPasswordCmd)
	cloudserverCmd.AddCommand(cloudserverConnectCmd)

	// Add flags for cloudserver commands
	cloudserverCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverCreateCmd.Flags().String("name", "", "Name for the cloud server (required)")
	cloudserverCreateCmd.Flags().String("region", "", "Region code (required)")
	cloudserverCreateCmd.Flags().String("zone", "", "Zone code (required, e.g., itbg1-a)")
	cloudserverCreateCmd.Flags().String("flavor", "", "Flavor name (required)")
	cloudserverCreateCmd.Flags().String("image", "", "Image ID or name (required)")
	cloudserverCreateCmd.Flags().String("keypair-uri", "", "Keypair URI (e.g., /projects/{project-id}/providers/Aruba.Compute/keyPairs/{keypair-name})")
	cloudserverCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	cloudserverCreateCmd.Flags().String("user-data-file", "", "Path to cloud-init YAML file (will be base64 encoded)")
	cloudserverCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})")
	cloudserverCreateCmd.Flags().StringSlice("subnet-uri", []string{}, "Subnet URI(s) (required, comma-separated)")
	cloudserverCreateCmd.Flags().StringSlice("security-group-uri", []string{}, "Security Group URI(s) (required, comma-separated)")
	cloudserverCreateCmd.Flags().String("elasticip-uri", "", "Elastic IP URI (optional)")
	cloudserverCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year (optional, default: Hour)")
	cloudserverCreateCmd.MarkFlagRequired("name")
	cloudserverCreateCmd.MarkFlagRequired("region")
	cloudserverCreateCmd.MarkFlagRequired("flavor")
	cloudserverCreateCmd.MarkFlagRequired("image")
	cloudserverCreateCmd.MarkFlagRequired("zone")

	cloudserverGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	cloudserverUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverUpdateCmd.Flags().String("name", "", "New name for the cloud server")
	cloudserverUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	cloudserverDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cloudserverDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	cloudserverListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	cloudserverListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	cloudserverPowerOnCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverPowerOffCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	cloudserverSetPasswordCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverSetPasswordCmd.Flags().String("password", "", "New password for the cloud server (required)")
	cloudserverSetPasswordCmd.MarkFlagRequired("password")

	cloudserverConnectCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverConnectCmd.Flags().String("user", "<user>", "SSH username (required - see documentation for image-specific users)")

	// Set up auto-completion for resource IDs
	cloudserverGetCmd.ValidArgsFunction = completeCloudServerID
	cloudserverUpdateCmd.ValidArgsFunction = completeCloudServerID
	cloudserverDeleteCmd.ValidArgsFunction = completeCloudServerID
	cloudserverPowerOnCmd.ValidArgsFunction = completeCloudServerID
	cloudserverPowerOffCmd.ValidArgsFunction = completeCloudServerID
	cloudserverSetPasswordCmd.ValidArgsFunction = completeCloudServerID
	cloudserverConnectCmd.ValidArgsFunction = completeCloudServerID
}

// Helper function to extract ID from URI
func extractIDFromURI(uri string) string {
	if uri == "" {
		return ""
	}
	// URI format: /projects/{projectId}/providers/Aruba.Compute/cloudServers/{serverId}
	parts := strings.Split(uri, "/")
	if len(parts) > 0 {
		// Get the last part which should be the ID
		return parts[len(parts)-1]
	}
	return ""
}

// Completion functions for compute resources
func completeCloudServerID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromCompute().CloudServers().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, server := range response.Data.Values {
			if server.Metadata.ID == nil || *server.Metadata.ID == "" {
				continue
			}
			id := *server.Metadata.ID
			var name string
			if server.Metadata.Name != nil {
				name = *server.Metadata.Name
			}
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CloudServer subcommands
var cloudserverCmd = &cobra.Command{
	Use:   "cloudserver",
	Short: "Manage cloud servers",
	Long:  `Perform CRUD operations on cloud servers in Aruba Cloud.`,
}

var cloudserverCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloud server",
	Long: `Create a new cloud server in the specified region and VPC.

The boot disk is specified as a URI referencing a compute template (image). Network
resources (VPC, subnet, security group) must already exist; pass their URIs with
--vpc-uri, --subnet-uri, and --security-group-uri.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud compute cloudserver create \
    --name my-server --region IT-BG --zone IT-BG-1 \
    --flavor <flavor-id> \
    --boot-disk-uri /projects/<proj-id>/providers/Aruba.Compute/templates/<template-id> \
    --vpc-uri /projects/<proj-id>/providers/Aruba.Network/vpcs/<vpc-id> \
    --subnet-uri /projects/<proj-id>/providers/Aruba.Network/subnets/<subnet-id> \
    --security-group-uri /projects/<proj-id>/providers/Aruba.Network/securityGroups/<sg-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get new network flags
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURIs, _ := cmd.Flags().GetStringSlice("subnet-uri")
		securityGroupURIs, _ := cmd.Flags().GetStringSlice("security-group-uri")
		elasticIPURI, _ := cmd.Flags().GetString("elasticip-uri")
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		zone, _ := cmd.Flags().GetString("zone")
		flavor, _ := cmd.Flags().GetString("flavor")
		bootDiskURI, _ := cmd.Flags().GetString("boot-disk-uri")
		keypairURI, _ := cmd.Flags().GetString("keypair-uri")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		userDataFile, _ := cmd.Flags().GetString("user-data-file")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build the create request
		// Note: Template (image) should be provided as a ReferenceResource URI
		// Format: /projects/{projectId}/providers/Aruba.Compute/templates/{templateId}
		// Use bootDiskURI for BootVolume
		bootVolumeURI := bootDiskURI

		createRequest := types.CloudServerRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.CloudServerPropertiesRequest{
				Zone:       zone,
				FlavorName: &flavor,
				BootVolume: types.ReferenceResource{
					URI: bootVolumeURI,
				},
				VPC: types.ReferenceResource{URI: vpcURI},
				Subnets: func() []types.ReferenceResource {
					var refs []types.ReferenceResource
					for _, s := range subnetURIs {
						refs = append(refs, types.ReferenceResource{URI: s})
					}
					return refs
				}(),
				SecurityGroups: func() []types.ReferenceResource {
					var refs []types.ReferenceResource
					for _, sg := range securityGroupURIs {
						refs = append(refs, types.ReferenceResource{URI: sg})
					}
					return refs
				}(),
			},
		}

		// Optionally set Elastic IP
		if elasticIPURI != "" {
			createRequest.Properties.ElasticIP = types.ReferenceResource{URI: elasticIPURI}
		}

		if keypairURI != "" {
			createRequest.Properties.KeyPair = types.ReferenceResource{
				URI: keypairURI,
			}
		}

		// Handle userData file if provided
		if userDataFile != "" {
			fileContent, err := os.ReadFile(userDataFile)
			if err != nil {
				return fmt.Errorf("reading user-data file: %w", err)
			}
			// Encode file content to base64
			userDataBase64 := base64.StdEncoding.EncodeToString(fileContent)
			createRequest.Properties.UserData = &userDataBase64
		}

		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromCompute().CloudServers().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating cloud server: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "FLAVOR", Width: 20},
				{Header: "CPU", Width: 10},
				{Header: "RAM(GB)", Width: 15},
				{Header: "HD(GB)", Width: 15},
				{Header: "REGION", Width: 20},
			}
			var id, name string
			if response.Data.Metadata.ID != nil {
				id = *response.Data.Metadata.ID
			}
			if response.Data.Metadata.Name != nil {
				name = *response.Data.Metadata.Name
			}
			flavorName := response.Data.Properties.Flavor.Name
			cpu := response.Data.Properties.Flavor.CPU
			ram := response.Data.Properties.Flavor.RAM
			hd := response.Data.Properties.Flavor.HD
			regionValue := ""
			if response.Data.Metadata.LocationResponse != nil {
				regionValue = response.Data.Metadata.LocationResponse.Value
			}
			row := []string{
				id,
				name,
				flavorName,
				fmt.Sprintf("%d", cpu),
				fmt.Sprintf("%d", ram),
				fmt.Sprintf("%d", hd),
				regionValue,
			}
			PrintOutput(response.Data, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Cloud server", name))
		}
		return nil
	},
}

var cloudserverGetCmd = &cobra.Command{
	Use:   "get [cloudserver-id]",
	Short: "Get cloud server details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		resp, err := client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			return fmt.Errorf("getting cloud server: %w", err)
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			return fmtAPIError(resp.StatusCode, resp.Error.Title, resp.Error.Detail)
		}

		if resp != nil && resp.Data != nil {
			server := resp.Data

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(server, nil, nil)
				return nil
			}

			fmt.Println("\nCloud Server Details:")
			fmt.Println("====================")

			if server.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *server.Metadata.ID)
			}
			if server.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *server.Metadata.Name)
			}
			if server.Metadata.LocationResponse != nil && server.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", server.Metadata.LocationResponse.Value)
			}

			// Flavor is a direct struct, not a pointer
			if server.Properties.Flavor.Name != "" {
				fmt.Printf("Flavor:          %s\n", server.Properties.Flavor.Name)
			}
			fmt.Printf("CPU:             %d\n", server.Properties.Flavor.CPU)
			fmt.Printf("RAM:             %d GB\n", server.Properties.Flavor.RAM)
			fmt.Printf("HD:              %d GB\n", server.Properties.Flavor.HD)

			if server.Properties.BootVolume.URI != "" {
				fmt.Printf("Boot Volume URI: %s\n", server.Properties.BootVolume.URI)
			}

			if server.Properties.KeyPair.URI != "" {
				fmt.Printf("Keypair URI:     %s\n", server.Properties.KeyPair.URI)
			}

			if server.Status.State != nil {
				fmt.Printf("Status:          %s\n", *server.Status.State)
			}

			if len(server.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", server.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			// Show JSON output if verbose
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				jsonData, _ := json.MarshalIndent(server, "", "  ")
				fmt.Println("\nFull JSON Response:")
				fmt.Println("==================")
				fmt.Println(string(jsonData))
			}
		} else {
			fmt.Println("Cloud server not found or no data returned.")
		}
		return nil
	},
}

var cloudserverUpdateCmd = &cobra.Command{
	Use:   "update [cloudserver-id]",
	Short: "Update a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && len(tags) == 0 {
			return fmt.Errorf("at least one of --name or --tags must be provided")
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
		getResponse, err := client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			return fmt.Errorf("getting cloud server details: %w", err)
		}

		if getResponse == nil || getResponse.Data == nil {
			return fmt.Errorf("cloud server not found")
		}

		current := getResponse.Data

		// Get region value
		var regionValue string
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for cloud server")
		}

		// Build the update request, preserving existing values
		flavorName := current.Properties.Flavor.Name

		var currentName string
		if current.Metadata.Name != nil {
			currentName = *current.Metadata.Name
		}

		updateRequest := types.CloudServerRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: currentName,
					Tags: current.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.CloudServerPropertiesRequest{
				FlavorName: &flavorName,
				BootVolume: current.Properties.BootVolume,
			},
		}

		if current.Properties.KeyPair.URI != "" {
			updateRequest.Properties.KeyPair = current.Properties.KeyPair
		}

		// Apply updates
		if name != "" {
			updateRequest.Metadata.Name = name
		}
		if len(tags) > 0 {
			updateRequest.Metadata.Tags = tags
		}

		response, err := client.FromCompute().CloudServers().Update(ctx, projectID, serverID, updateRequest, nil)
		if err != nil {
			return fmt.Errorf("updating cloud server: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("Cloud server", serverID))
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *response.Data.Metadata.Name)
			}
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("Cloud server", serverID))
		}
		return nil
	},
}

var cloudserverDeleteCmd = &cobra.Command{
	Use:   "delete [cloudserver-id]",
	Short: "Delete a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		// Confirmation prompt
		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("cloud server", serverID)
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
			_, err = client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: cloud server not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("cloud server", serverID))
			return nil
		}

		response, err := client.FromCompute().CloudServers().Delete(ctx, projectID, serverID, nil)
		if err != nil {
			return fmt.Errorf("deleting cloud server: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		fmt.Println(msgDeleted("Cloud server", serverID))
		return nil
	},
}

var cloudserverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cloud servers",
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
		response, err := client.FromCompute().CloudServers().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing cloud servers: %w", err)
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 25},
				{Header: "ID", Width: 30},
				{Header: "LOCATION", Width: 15},
				{Header: "FLAVOR", Width: 15},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, server := range response.Data.Values {
				if server.Metadata.ID == nil || *server.Metadata.ID == "" {
					continue
				}
				id := *server.Metadata.ID
				var name string
				if server.Metadata.Name != nil {
					name = *server.Metadata.Name
				}
				var location string
				if server.Metadata.LocationResponse != nil {
					location = server.Metadata.LocationResponse.Value
				}
				flavor := server.Properties.Flavor.Name
				status := ""
				if server.Status.State != nil {
					status = *server.Status.State
				}
				rows = append(rows, []string{
					name,
					id,
					location,
					flavor,
					status,
				})
			}

			if len(rows) == 0 {
				fmt.Println("No cloud servers found")
				return nil
			}
			PrintOutput(response.Data, headers, rows)
		} else {
			fmt.Println("No cloud servers found")
		}
		return nil
	},
}

var cloudserverPowerOnCmd = &cobra.Command{
	Use:   "power-on [cloudserver-id]",
	Short: "Power on a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		response, err := client.FromCompute().CloudServers().PowerOn(ctx, projectID, serverID, nil)
		if err != nil {
			return fmt.Errorf("powering on cloud server: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Println(msgAction("Cloud server", serverID, "powered on"))
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Server: %s\n", *response.Data.Metadata.Name)
			}
			if response.Data.Status.State != nil {
				fmt.Printf("Status: %s\n", *response.Data.Status.State)
			}
		} else {
			fmt.Println(msgAction("Cloud server", serverID, "power-on initiated"))
		}
		return nil
	},
}

var cloudserverPowerOffCmd = &cobra.Command{
	Use:   "power-off [cloudserver-id]",
	Short: "Power off a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		response, err := client.FromCompute().CloudServers().PowerOff(ctx, projectID, serverID, nil)
		if err != nil {
			return fmt.Errorf("powering off cloud server: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			fmt.Println(msgAction("Cloud server", serverID, "powered off"))
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Server: %s\n", *response.Data.Metadata.Name)
			}
			if response.Data.Status.State != nil {
				fmt.Printf("Status: %s\n", *response.Data.Status.State)
			}
		} else {
			fmt.Println(msgAction("Cloud server", serverID, "power-off initiated"))
		}
		return nil
	},
}

var cloudserverSetPasswordCmd = &cobra.Command{
	Use:   "set-password [cloudserver-id]",
	Short: "Set password for a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			return fmt.Errorf("--password is required")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		passwordRequest := types.CloudServerPasswordRequest{
			Password: password,
		}

		response, err := client.FromCompute().CloudServers().SetPassword(ctx, projectID, serverID, passwordRequest, nil)
		if err != nil {
			return fmt.Errorf("setting password for cloud server: %w", err)
		}

		if response != nil && response.IsError() && response.Error != nil {
			return fmtAPIError(response.StatusCode, response.Error.Title, response.Error.Detail)
		}

		if response != nil && response.Data != nil {
			// Try to cast to CloudServerResponse to get detailed info
			// response.Data is *any, so we need to dereference and assert
			if data, ok := (*response.Data).(*types.CloudServerResponse); ok && data != nil {
				fmt.Println(msgAction("Cloud server", serverID, "password set"))
				if data.Metadata.Name != nil {
					fmt.Printf("Server: %s\n", *data.Metadata.Name)
				}
				if data.Status.State != nil {
					fmt.Printf("Status: %s\n", *data.Status.State)
				}
			} else {
				fmt.Println(msgAction("Cloud server", serverID, "password set"))
			}
		} else {
			fmt.Println(msgAction("Cloud server", serverID, "password set"))
		}
		return nil
	},
}

var cloudserverConnectCmd = &cobra.Command{
	Use:   "connect [cloudserver-id]",
	Short: "Get SSH connection information for a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		user, _ := cmd.Flags().GetString("user")
		if user == "" || user == "<user>" {
			fmt.Println("Error: --user is required")
			fmt.Println("\nCommon SSH users by image type:")
			fmt.Println("  - Ubuntu/Debian: ubuntu")
			fmt.Println("  - CentOS/RHEL: centos or root")
			fmt.Println("  - Other Linux: root or check image documentation")
			fmt.Println("\nFor more information, see: https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx")
			return fmt.Errorf("--user is required")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		// First, get the cloud server details
		serverResp, err := client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			return fmt.Errorf("getting cloud server: %w", err)
		}

		if serverResp != nil && serverResp.IsError() && serverResp.Error != nil {
			return fmtAPIError(serverResp.StatusCode, serverResp.Error.Title, serverResp.Error.Detail)
		}

		if serverResp == nil || serverResp.Data == nil {
			fmt.Println("Cloud server not found or no data returned.")
			return nil
		}

		server := serverResp.Data

		// Check for ElasticIP in linked resources
		var elasticIPURI string
		for _, linkedResource := range server.Properties.LinkedResources {
			if strings.Contains(linkedResource.URI, "providers/Aruba.Network/elasticIps") {
				elasticIPURI = linkedResource.URI
				break
			}
		}

		if elasticIPURI == "" {
			fmt.Println("No Elastic IP found for this cloud server.")
			fmt.Println("The server must have an Elastic IP linked to use the connect command.")
			return nil
		}

		// Extract ElasticIP ID from URI
		elasticIPID := extractIDFromURI(elasticIPURI)
		if elasticIPID == "" {
			return fmt.Errorf("could not extract Elastic IP ID from URI: %s", elasticIPURI)
		}

		// Get ElasticIP details
		eipResp, err := client.FromNetwork().ElasticIPs().Get(ctx, projectID, elasticIPID, nil)
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", err)
		}

		if eipResp != nil && eipResp.IsError() && eipResp.Error != nil {
			return fmtAPIError(eipResp.StatusCode, eipResp.Error.Title, eipResp.Error.Detail)
		}

		if eipResp == nil || eipResp.Data == nil {
			fmt.Println("Elastic IP not found or no data returned.")
			return nil
		}

		eip := eipResp.Data
		if eip.Properties.Address == nil || *eip.Properties.Address == "" {
			fmt.Println("Elastic IP address not available.")
			return nil
		}

		// Print SSH connection command
		fmt.Printf("Connect by running: ssh %s@%s\n", user, *eip.Properties.Address)
		return nil
	},
}
