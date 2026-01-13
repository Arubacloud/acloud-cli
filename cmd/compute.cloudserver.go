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
	cloudserverCreateCmd.Flags().String("flavor", "", "Flavor name (required)")
	cloudserverCreateCmd.Flags().String("image", "", "Image ID or name (required)")
	cloudserverCreateCmd.Flags().String("keypair", "", "Keypair name")
	cloudserverCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	cloudserverCreateCmd.MarkFlagRequired("name")
	cloudserverCreateCmd.MarkFlagRequired("region")
	cloudserverCreateCmd.MarkFlagRequired("flavor")
	cloudserverCreateCmd.MarkFlagRequired("image")

	cloudserverGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	cloudserverUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverUpdateCmd.Flags().String("name", "", "New name for the cloud server")
	cloudserverUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	cloudserverDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	cloudserverListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

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
			name := server.Metadata.Name
			// Try to extract ID from BootVolume URI or other URI fields
			id := name // Default to name
			if server.Properties.BootVolume.URI != "" {
				// Try to extract from a URI pattern if available
				// For now, use name as identifier
				id = name
			}
			if name != "" {
				// For completion, we can use name or try to extract ID from URI if available
				// The response structure may vary, so we'll use name as the primary identifier
				if toComplete == "" || strings.HasPrefix(name, toComplete) || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\tCloud Server", id))
				}
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
	Run: func(cmd *cobra.Command, args []string) {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		flavor, _ := cmd.Flags().GetString("flavor")
		image, _ := cmd.Flags().GetString("image")
		keypair, _ := cmd.Flags().GetString("keypair")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" || region == "" || flavor == "" || image == "" {
			fmt.Println("Error: --name, --region, --flavor, and --image are required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build the create request
		// Note: Template (image) should be provided as a ReferenceResource URI
		// Format: /projects/{projectId}/providers/Aruba.Compute/templates/{templateId}
		templateURI := image
		if !strings.HasPrefix(templateURI, "/") {
			// If image is just an ID, construct the URI
			templateURI = fmt.Sprintf("/projects/%s/providers/Aruba.Compute/templates/%s", projectID, image)
		}

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
				FlavorName: &flavor,
				BootVolume: types.ReferenceResource{
					URI: templateURI,
				},
			},
		}

		if keypair != "" {
			// KeyPair should be a ReferenceResource URI
			// Format: /projects/{projectId}/providers/Aruba.Compute/keyPairs/{keypairName}
			keypairURI := keypair
			if !strings.HasPrefix(keypairURI, "/") {
				keypairURI = fmt.Sprintf("/projects/%s/providers/Aruba.Compute/keyPairs/%s", projectID, keypair)
			}
			createRequest.Properties.KeyPair = types.ReferenceResource{
				URI: keypairURI,
			}
		}

		ctx := context.Background()
		response, err := client.FromCompute().CloudServers().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating cloud server: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create cloud server - Status: %d\n", response.StatusCode)
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
				{Header: "FLAVOR", Width: 20},
				{Header: "CPU", Width: 10},
				{Header: "RAM(GB)", Width: 15},
				{Header: "HD(GB)", Width: 15},
				{Header: "REGION", Width: 20},
			}
			// CloudServer response may not expose ID in metadata
			// Use name or extract from URI if needed
			id := response.Data.Metadata.Name // Fallback to name for display
			flavorName := response.Data.Properties.Flavor.Name
			cpu := response.Data.Properties.Flavor.CPU
			ram := response.Data.Properties.Flavor.RAM
			hd := response.Data.Properties.Flavor.HD
			regionValue := ""
			if response.Data.Metadata.Location.Value != "" {
				regionValue = response.Data.Metadata.Location.Value
			}
			row := []string{
				id,
				response.Data.Metadata.Name,
				flavorName,
				fmt.Sprintf("%d", cpu),
				fmt.Sprintf("%d", ram),
				fmt.Sprintf("%d", hd),
				regionValue,
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Cloud server created, but no data returned.")
		}
	},
}

var cloudserverGetCmd = &cobra.Command{
	Use:   "get [cloudserver-id]",
	Short: "Get cloud server details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

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
		resp, err := client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			fmt.Printf("Error getting cloud server: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get cloud server - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			server := resp.Data

			fmt.Println("\nCloud Server Details:")
			fmt.Println("====================")

			// CloudServer response metadata may not expose ID directly
			// ID can be extracted from URI if needed
			fmt.Printf("Name:            %s\n", server.Metadata.Name)
			if server.Metadata.Location.Value != "" {
				fmt.Printf("Region:          %s\n", server.Metadata.Location.Value)
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
	},
}

var cloudserverUpdateCmd = &cobra.Command{
	Use:   "update [cloudserver-id]",
	Short: "Update a cloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && len(tags) == 0 {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}

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
		getResponse, err := client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			fmt.Printf("Error getting cloud server details: %v\n", err)
			return
		}

		if getResponse == nil || getResponse.Data == nil {
			fmt.Println("Error: Cloud server not found")
			return
		}

		current := getResponse.Data

		// Get region value
		regionValue := current.Metadata.Location.Value
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for cloud server")
			return
		}

		// Build the update request, preserving existing values
		flavorName := current.Properties.Flavor.Name

		updateRequest := types.CloudServerRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: current.Metadata.Name,
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
			fmt.Printf("Error updating cloud server: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to update cloud server - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nCloud server updated successfully!")
			fmt.Printf("Name:    %s\n", response.Data.Metadata.Name)
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("Cloud server update initiated. Use 'get' to check status.")
		}
	},
}

var cloudserverDeleteCmd = &cobra.Command{
	Use:   "delete [cloudserver-id]",
	Short: "Delete a cloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

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
			fmt.Printf("Are you sure you want to delete cloud server '%s'? (yes/no): ", serverID)
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "yes" && confirmation != "y" {
				fmt.Println("Deletion cancelled.")
				return
			}
		}

		ctx := context.Background()
		response, err := client.FromCompute().CloudServers().Delete(ctx, projectID, serverID, nil)
		if err != nil {
			fmt.Printf("Error deleting cloud server: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to delete cloud server - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		fmt.Printf("Cloud server '%s' deleted successfully.\n", serverID)
	},
}

var cloudserverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cloud servers",
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
		response, err := client.FromCompute().CloudServers().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing cloud servers: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 25},
				{Header: "ID", Width: 30},
				{Header: "LOCATION", Width: 15},
				{Header: "FLAVOR", Width: 15},
				{Header: "STATUS", Width: 15},
			}

			// Extract IDs from raw JSON response if available
			// The SDK type definition uses Request types but actual response has ID fields
			idMap := make(map[int]string) // Map server index to ID
			if response.RawBody != nil {
				var rawResponse map[string]interface{}
				if err := json.Unmarshal(response.RawBody, &rawResponse); err == nil {
					if values, ok := rawResponse["values"].([]interface{}); ok {
						for i, val := range values {
							if serverMap, ok := val.(map[string]interface{}); ok {
								if metadata, ok := serverMap["metadata"].(map[string]interface{}); ok {
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
			for idx, server := range response.Data.Values {
				name := server.Metadata.Name
				location := server.Metadata.Location.Value
				flavor := server.Properties.Flavor.Name
				status := ""
				if server.Status.State != nil {
					status = *server.Status.State
				}

				// Get ID from raw JSON map, fallback to name
				id := idMap[idx]
				if id == "" {
					id = name
				}

				rows = append(rows, []string{
					name,
					id,
					location,
					flavor,
					status,
				})
			}

			PrintTable(headers, rows)
		} else {
			fmt.Println("No cloud servers found")
		}
	},
}

var cloudserverPowerOnCmd = &cobra.Command{
	Use:   "power-on [cloudserver-id]",
	Short: "Power on a cloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

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
		response, err := client.FromCompute().CloudServers().PowerOn(ctx, projectID, serverID, nil)
		if err != nil {
			fmt.Printf("Error powering on cloud server: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to power on cloud server - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("Cloud server powered on successfully!")
			fmt.Printf("Server: %s\n", response.Data.Metadata.Name)
			if response.Data.Status.State != nil {
				fmt.Printf("Status: %s\n", *response.Data.Status.State)
			}
		} else {
			fmt.Println("Cloud server power-on initiated. Use 'get' to check status.")
		}
	},
}

var cloudserverPowerOffCmd = &cobra.Command{
	Use:   "power-off [cloudserver-id]",
	Short: "Power off a cloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

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
		response, err := client.FromCompute().CloudServers().PowerOff(ctx, projectID, serverID, nil)
		if err != nil {
			fmt.Printf("Error powering off cloud server: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to power off cloud server - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("Cloud server powered off successfully!")
			fmt.Printf("Server: %s\n", response.Data.Metadata.Name)
			if response.Data.Status.State != nil {
				fmt.Printf("Status: %s\n", *response.Data.Status.State)
			}
		} else {
			fmt.Println("Cloud server power-off initiated. Use 'get' to check status.")
		}
	},
}

var cloudserverSetPasswordCmd = &cobra.Command{
	Use:   "set-password [cloudserver-id]",
	Short: "Set password for a cloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			fmt.Println("Error: --password is required")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		passwordRequest := types.CloudServerPasswordRequest{
			Password: password,
		}

		response, err := client.FromCompute().CloudServers().SetPassword(ctx, projectID, serverID, passwordRequest, nil)
		if err != nil {
			fmt.Printf("Error setting password for cloud server: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to set password for cloud server - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			// Try to cast to CloudServerResponse to get detailed info
			// response.Data is *any, so we need to dereference and assert
			if data, ok := (*response.Data).(*types.CloudServerResponse); ok && data != nil {
				fmt.Println("Cloud server password set successfully!")
				fmt.Printf("Server: %s\n", data.Metadata.Name)
				if data.Status.State != nil {
					fmt.Printf("Status: %s\n", *data.Status.State)
				}
			} else {
				// If response doesn't have CloudServerResponse structure, show simple success
				fmt.Println("Cloud server password set successfully!")
				fmt.Printf("Server ID: %s\n", serverID)
			}
		} else {
			fmt.Println("Cloud server password set successfully!")
			fmt.Printf("Server ID: %s\n", serverID)
		}
	},
}

var cloudserverConnectCmd = &cobra.Command{
	Use:   "connect [cloudserver-id]",
	Short: "Get SSH connection information for a cloud server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		user, _ := cmd.Flags().GetString("user")
		if user == "" || user == "<user>" {
			fmt.Println("Error: --user is required")
			fmt.Println("\nCommon SSH users by image type:")
			fmt.Println("  - Ubuntu/Debian: ubuntu")
			fmt.Println("  - CentOS/RHEL: centos or root")
			fmt.Println("  - Other Linux: root or check image documentation")
			fmt.Println("\nFor more information, see: https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx")
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()

		// First, get the cloud server details
		serverResp, err := client.FromCompute().CloudServers().Get(ctx, projectID, serverID, nil)
		if err != nil {
			fmt.Printf("Error getting cloud server: %v\n", err)
			return
		}

		if serverResp != nil && serverResp.IsError() && serverResp.Error != nil {
			fmt.Printf("Failed to get cloud server - Status: %d\n", serverResp.StatusCode)
			if serverResp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *serverResp.Error.Title)
			}
			if serverResp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *serverResp.Error.Detail)
			}
			return
		}

		if serverResp == nil || serverResp.Data == nil {
			fmt.Println("Cloud server not found or no data returned.")
			return
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
			return
		}

		// Extract ElasticIP ID from URI
		elasticIPID := extractIDFromURI(elasticIPURI)
		if elasticIPID == "" {
			fmt.Printf("Error: Could not extract Elastic IP ID from URI: %s\n", elasticIPURI)
			return
		}

		// Get ElasticIP details
		eipResp, err := client.FromNetwork().ElasticIPs().Get(ctx, projectID, elasticIPID, nil)
		if err != nil {
			fmt.Printf("Error getting Elastic IP details: %v\n", err)
			return
		}

		if eipResp != nil && eipResp.IsError() && eipResp.Error != nil {
			fmt.Printf("Failed to get Elastic IP - Status: %d\n", eipResp.StatusCode)
			if eipResp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *eipResp.Error.Title)
			}
			if eipResp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *eipResp.Error.Detail)
			}
			return
		}

		if eipResp == nil || eipResp.Data == nil {
			fmt.Println("Elastic IP not found or no data returned.")
			return
		}

		eip := eipResp.Data
		if eip.Properties.Address == nil || *eip.Properties.Address == "" {
			fmt.Println("Elastic IP address not available.")
			return
		}

		// Print SSH connection command
		fmt.Printf("Connect by running: ssh %s@%s\n", user, *eip.Properties.Address)
	},
}
