package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	// VPNTunnel
	networkCmd.AddCommand(vpntunnelCmd)
	vpntunnelCmd.AddCommand(vpntunnelCreateCmd)
	vpntunnelCmd.AddCommand(vpntunnelGetCmd)
	vpntunnelCmd.AddCommand(vpntunnelUpdateCmd)
	vpntunnelCmd.AddCommand(vpntunnelDeleteCmd)
	vpntunnelCmd.AddCommand(vpntunnelListCmd)

	// Add flags for VPN Tunnel commands
	vpntunnelCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelCreateCmd.Flags().String("name", "", "Name for the VPN tunnel")
	vpntunnelCreateCmd.Flags().String("region", "", "Region code (e.g., ITBG-Bergamo)")
	vpntunnelCreateCmd.Flags().String("peer-ip", "", "Peer client public IP address")
	vpntunnelCreateCmd.Flags().String("vpc-uri", "", "VPC URI (e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})")
	vpntunnelCreateCmd.Flags().String("subnet-cidr", "", "Subnet CIDR (e.g., 10.0.1.0/24)")
	vpntunnelCreateCmd.Flags().String("subnet-name", "", "Subnet name (alternative to CIDR)")
	vpntunnelCreateCmd.Flags().String("elastic-ip-uri", "", "Elastic IP URI (e.g., /projects/{project-id}/providers/Aruba.Network/elasticIps/{ip-id})")
	vpntunnelCreateCmd.Flags().String("vpn-type", "Site-To-Site", "VPN type (default: Site-To-Site)")
	vpntunnelCreateCmd.Flags().String("protocol", "ikev2", "VPN protocol (default: ikev2)")
	vpntunnelCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year")
	vpntunnelCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	// IKE settings
	vpntunnelCreateCmd.Flags().Int32("ike-lifetime", 0, "IKE lifetime (seconds)")
	vpntunnelCreateCmd.Flags().String("ike-encryption", "", "IKE encryption algorithm")
	vpntunnelCreateCmd.Flags().String("ike-hash", "", "IKE hash algorithm")
	vpntunnelCreateCmd.Flags().String("ike-dh-group", "", "IKE DH group")
	vpntunnelCreateCmd.Flags().String("ike-dpd-action", "", "IKE DPD action")
	vpntunnelCreateCmd.Flags().Int32("ike-dpd-interval", 0, "IKE DPD interval (seconds)")
	vpntunnelCreateCmd.Flags().Int32("ike-dpd-timeout", 0, "IKE DPD timeout (seconds)")
	// ESP settings
	vpntunnelCreateCmd.Flags().Int32("esp-lifetime", 0, "ESP lifetime (seconds)")
	vpntunnelCreateCmd.Flags().String("esp-encryption", "", "ESP encryption algorithm")
	vpntunnelCreateCmd.Flags().String("esp-hash", "", "ESP hash algorithm")
	vpntunnelCreateCmd.Flags().String("esp-pfs", "", "ESP PFS group")
	// PSK settings
	vpntunnelCreateCmd.Flags().String("psk-cloud-site", "", "PSK cloud site")
	vpntunnelCreateCmd.Flags().String("psk-onprem-site", "", "PSK on-prem site")
	vpntunnelCreateCmd.Flags().String("psk", "", "Pre-shared key for authentication (PSK secret)")
	vpntunnelGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelUpdateCmd.Flags().String("name", "", "New name for the VPN tunnel")
	vpntunnelUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpntunnelDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpntunnelListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	vpntunnelGetCmd.ValidArgsFunction = completeVPNTunnelID
	vpntunnelUpdateCmd.ValidArgsFunction = completeVPNTunnelID
	vpntunnelDeleteCmd.ValidArgsFunction = completeVPNTunnelID
}

// Completion functions for network resources

func completeVPNTunnelID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Allow completion even if args exist - user might be completing a partial ID

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromNetwork().VPNTunnels().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, vpn := range response.Data.Values {
			if vpn.Metadata.ID != nil && vpn.Metadata.Name != nil {
				id := *vpn.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *vpn.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// VPNTunnel subcommands
var vpntunnelCmd = &cobra.Command{
	Use:   "vpntunnel",
	Short: "Manage VPN tunnels",
	Long:  `Perform CRUD operations on VPN tunnels in Aruba Cloud.`,
}

var vpntunnelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPN tunnels",
	Run: func(cmd *cobra.Command, args []string) {
		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get projectID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// List VPN tunnels using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPNTunnels().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing VPN tunnels: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 25},
				{Header: "REGION", Width: 18},
				{Header: "TYPE", Width: 15},
				{Header: "STATUS", Width: 15},
			}
			// Build rows
			var rows [][]string
			for _, vpn := range response.Data.Values {
				name := ""
				if vpn.Metadata.Name != nil && *vpn.Metadata.Name != "" {
					name = *vpn.Metadata.Name
				}

				id := ""
				if vpn.Metadata.ID != nil {
					id = *vpn.Metadata.ID
				}

				region := vpn.Metadata.LocationResponse.Value

				vpnType := ""
				if vpn.Properties.VPNType != nil {
					vpnType = *vpn.Properties.VPNType
				}

				status := ""
				if vpn.Status.State != nil {
					status = *vpn.Status.State
				}

				rows = append(rows, []string{name, id, region, vpnType, status})
			}

			// Print the table
			PrintTable(headers, rows)
		} else {
			fmt.Println("No VPN tunnels found")
		}
	},
}

var vpntunnelGetCmd = &cobra.Command{
	Use:   "get <vpn-tunnel-id>",
	Short: "Get VPN tunnel details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpnID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get VPN tunnel details using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPNTunnels().Get(ctx, projectID, vpnID, nil)
		if err != nil {
			fmt.Printf("Error getting VPN tunnel details: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			vpn := response.Data

			// Display VPN tunnel details
			fmt.Println("\nVPN Tunnel Details:")
			fmt.Println("===================")

			if vpn.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *vpn.Metadata.ID)
			}
			if vpn.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *vpn.Metadata.URI)
			}
			if vpn.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *vpn.Metadata.Name)
			}
			if vpn.Metadata.LocationResponse != nil && vpn.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", vpn.Metadata.LocationResponse.Value)
			}

			if vpn.Properties.VPNType != nil {
				fmt.Printf("VPN Type:        %s\n", *vpn.Properties.VPNType)
			}
			if vpn.Properties.VPNClientProtocol != nil {
				fmt.Printf("Protocol:        %s\n", *vpn.Properties.VPNClientProtocol)
			}
			if vpn.Properties.VPNClientSettings != nil && vpn.Properties.VPNClientSettings.PeerClientPublicIP != nil {
				fmt.Printf("Peer IP:         %s\n", *vpn.Properties.VPNClientSettings.PeerClientPublicIP)
			}

			if vpn.Properties.IPConfigurations != nil {
				fmt.Println("\nIP Configuration:")
				if vpn.Properties.IPConfigurations.VPC != nil {
					fmt.Printf("  VPC:           %s\n", vpn.Properties.IPConfigurations.VPC.URI)
				}
				if vpn.Properties.IPConfigurations.Subnet != nil {
					fmt.Printf("  Subnet CIDR:   %s\n", vpn.Properties.IPConfigurations.Subnet.CIDR)
					if vpn.Properties.IPConfigurations.Subnet.Name != "" {
						fmt.Printf("  Subnet Name:   %s\n", vpn.Properties.IPConfigurations.Subnet.Name)
					}
				}
				if vpn.Properties.IPConfigurations.PublicIP != nil {
					fmt.Printf("  Public IP:     %s\n", vpn.Properties.IPConfigurations.PublicIP.URI)
				}
			}

			if vpn.Properties.BillingPlan != nil {
				fmt.Printf("\nBilling Period:  %s\n", vpn.Properties.BillingPlan.BillingPeriod)
			}

			if vpn.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", vpn.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if vpn.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *vpn.Metadata.CreatedBy)
			}

			if len(vpn.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", vpn.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			if vpn.Status.State != nil {
				fmt.Printf("Status:          %s\n", *vpn.Status.State)
			}
		}
	},
}

var vpntunnelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VPN tunnel",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		vpnType, _ := cmd.Flags().GetString("vpn-type")
		protocol, _ := cmd.Flags().GetString("protocol")
		peerIP, _ := cmd.Flags().GetString("peer-ip")
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetCIDR, _ := cmd.Flags().GetString("subnet-cidr")
		subnetName, _ := cmd.Flags().GetString("subnet-name")
		publicIPURI, _ := cmd.Flags().GetString("elastic-ip-uri")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		psk, _ := cmd.Flags().GetString("psk")

		// IKE settings
		ikeLifetime, _ := cmd.Flags().GetInt32("ike-lifetime")
		ikeEncryption, _ := cmd.Flags().GetString("ike-encryption")
		ikeHash, _ := cmd.Flags().GetString("ike-hash")
		ikeDHGroup, _ := cmd.Flags().GetString("ike-dh-group")
		ikeDPDAction, _ := cmd.Flags().GetString("ike-dpd-action")
		ikeDPDInterval, _ := cmd.Flags().GetInt32("ike-dpd-interval")
		ikeDPDTimeout, _ := cmd.Flags().GetInt32("ike-dpd-timeout")
		// ESP settings
		espLifetime, _ := cmd.Flags().GetInt32("esp-lifetime")
		espEncryption, _ := cmd.Flags().GetString("esp-encryption")
		if espEncryption == "" {
			espEncryption = "aes256"
		}
		espHash, _ := cmd.Flags().GetString("esp-hash")
		espPFS, _ := cmd.Flags().GetString("esp-pfs")
		// PSK settings
		pskCloudSite, _ := cmd.Flags().GetString("psk-cloud-site")
		pskOnpremSite, _ := cmd.Flags().GetString("psk-onprem-site")

		// Validate required fields
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}
		if region == "" {
			fmt.Println("Error: --region is required")
			return
		}
		if peerIP == "" {
			fmt.Println("Error: --peer-ip is required")
			return
		}
		if vpcURI == "" {
			fmt.Println("Error: --vpc-uri is required (e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})")
			return
		}
		if subnetCIDR == "" && subnetName == "" {
			fmt.Println("Error: --subnet-cidr or --subnet-name is required")
			return
		}
		if publicIPURI == "" {
			fmt.Println("Error: --elastic-ip-uri is required (e.g., /projects/{project-id}/providers/Aruba.Network/elasticIps/{ip-id})")
			return
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build subnet object with both CIDR and Name fields
		subnetRef := &types.SubnetInfo{}
		if subnetCIDR != "" {
			subnetRef.CIDR = subnetCIDR
		}
		if subnetName != "" {
			subnetRef.Name = subnetName
		}

		// Build IP configurations
		ipConfig := types.IPConfigurations{
			VPC: &types.ReferenceResource{
				URI: vpcURI,
			},
			Subnet: subnetRef,
			PublicIP: &types.ReferenceResource{
				URI: publicIPURI,
			},
		}

		// Build IKE settings
		ikeSettings := &types.IKESettings{
			Lifetime:    ikeLifetime,
			Encryption:  nil,
			Hash:        nil,
			DHGroup:     nil,
			DPDAction:   nil,
			DPDInterval: ikeDPDInterval,
			DPDTimeout:  ikeDPDTimeout,
		}
		if ikeEncryption != "" {
			ikeSettings.Encryption = &ikeEncryption
		}
		if ikeHash != "" {
			ikeSettings.Hash = &ikeHash
		}
		if ikeDHGroup != "" {
			ikeSettings.DHGroup = &ikeDHGroup
		}
		if ikeDPDAction != "" {
			ikeSettings.DPDAction = &ikeDPDAction
		}

		// Build ESP settings
		espSettings := &types.ESPSettings{
			Lifetime:   espLifetime,
			Encryption: &espEncryption, // always include, default to aes256
			Hash:       nil,
			PFS:        nil,
		}
		if espHash != "" {
			espSettings.Hash = &espHash
		}
		if espPFS != "" {
			espSettings.PFS = &espPFS
		}

		// Build PSK settings
		pskSettings := &types.PSKSettings{
			Secret:     nil,
			CloudSite:  nil,
			OnPremSite: nil,
		}
		if psk != "" {
			pskSettings.Secret = &psk
		}
		if pskCloudSite != "" {
			pskSettings.CloudSite = &pskCloudSite
		}
		if pskOnpremSite != "" {
			pskSettings.OnPremSite = &pskOnpremSite
		}

		// Build VPN client settings
		vpnClientSettings := &types.VPNClientSettings{
			IKE:                ikeSettings,
			ESP:                espSettings,
			PSK:                pskSettings,
			PeerClientPublicIP: &peerIP,
		}

		// Build the create request using custom types that match the API
		createRequest := types.VPNTunnelRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.VPNTunnelPropertiesRequest{
				VPNType:           &vpnType,
				VPNClientProtocol: &protocol,
				IPConfigurations:  &ipConfig,
				VPNClientSettings: vpnClientSettings,
				BillingPlan: &types.BillingPeriodResource{
					BillingPeriod: billingPeriod,
				},
			},
		}

		// Create the VPN tunnel using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPNTunnels().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			fmt.Printf("Error creating VPN tunnel: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to create VPN tunnel - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		if response != nil && response.Data != nil {
			fmt.Println("\nVPN Tunnel created successfully!")
			if response.Data.Metadata.ID != nil {
				fmt.Printf("ID:       %s\n", *response.Data.Metadata.ID)
			}
			if response.Data.Metadata.Name != nil {
				fmt.Printf("Name:     %s\n", *response.Data.Metadata.Name)
			}
			if response.Data.Metadata.URI != nil {
				fmt.Printf("URI:      %s\n", *response.Data.Metadata.URI)
			}
			if response.Data.Properties.VPNType != nil {
				fmt.Printf("Type:     %s\n", *response.Data.Properties.VPNType)
			}
			if response.Data.Properties.VPNClientProtocol != nil {
				fmt.Printf("Protocol: %s\n", *response.Data.Properties.VPNClientProtocol)
			}
			if len(response.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:     %v\n", response.Data.Metadata.Tags)
			}
		} else {
			fmt.Println("VPN Tunnel creation initiated. Use 'list' or 'get' to check status.")
		}
	},
}

var vpntunnelUpdateCmd = &cobra.Command{
	Use:   "update <vpn-tunnel-id>",
	Short: "Update a VPN tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one update flag must be provided
		if name == "" && len(tags) == 0 {
			fmt.Println("Error: at least one of --name or --tags must be provided")
			return
		}

		// Get project ID
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get Aruba Cloud client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get current VPN tunnel configuration
		ctx := context.Background()
		getResp, err := client.FromNetwork().VPNTunnels().Get(ctx, projectID, vpnTunnelID, nil)
		if err != nil {
			fmt.Printf("Error getting VPN tunnel: %v\n", err)
			return
		}

		if getResp != nil && getResp.IsError() && getResp.Error != nil {
			fmt.Printf("Error getting VPN tunnel - Status: %d\n", getResp.StatusCode)
			if getResp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *getResp.Error.Title)
			}
			if getResp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *getResp.Error.Detail)
			}
			return
		}

		if getResp.Data == nil {
			fmt.Println("Error: VPN tunnel not found")
			return
		}

		// Check if VPN tunnel is in "InCreation" state
		if getResp.Data.Status.State != nil && *getResp.Data.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update VPN tunnel while it is in 'InCreation' state. Please wait until the VPN tunnel is fully created.")
			return
		}

		// Get region value
		regionValue := ""
		if getResp.Data.Metadata.LocationResponse != nil {
			regionValue = getResp.Data.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for VPN tunnel")
			return
		}

		// Build update request, preserving current values
		updateReq := types.VPNTunnelRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: *getResp.Data.Metadata.Name,
					Tags: getResp.Data.Metadata.Tags,
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.VPNTunnelPropertiesRequest{
				VPNType:           getResp.Data.Properties.VPNType,
				VPNClientProtocol: getResp.Data.Properties.VPNClientProtocol,
				IPConfigurations:  getResp.Data.Properties.IPConfigurations,
				VPNClientSettings: getResp.Data.Properties.VPNClientSettings,
				// PeerClientPublicIP is now set via VPNClientSettings only
				BillingPlan: getResp.Data.Properties.BillingPlan,
			},
		}

		// Apply updates
		if name != "" {
			updateReq.Metadata.Name = name
		}
		if len(tags) > 0 {
			updateReq.Metadata.Tags = tags
		}

		// Update VPN tunnel
		resp, err := client.FromNetwork().VPNTunnels().Update(ctx, projectID, vpnTunnelID, updateReq, nil)
		if err != nil {
			fmt.Printf("Error updating VPN tunnel: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to update VPN tunnel - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp.Data != nil {
			fmt.Println("\nVPN Tunnel updated successfully!")
			if resp.Data.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *resp.Data.Metadata.ID)
			}
			if resp.Data.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *resp.Data.Metadata.Name)
			}
			if len(resp.Data.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", resp.Data.Metadata.Tags)
			}
		} else {
			fmt.Printf("\nVPN Tunnel %s update completed.\n", vpnTunnelID)
		}
	},
}

var vpntunnelDeleteCmd = &cobra.Command{
	Use:   "delete <vpn-tunnel-id>",
	Short: "Delete a VPN tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpnTunnelID := args[0]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete VPN tunnel %s? This action cannot be undone.\n", vpnTunnelID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Delete the VPN tunnel using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().VPNTunnels().Delete(ctx, projectID, vpnTunnelID, nil)
		if err != nil {
			fmt.Printf("Error deleting VPN tunnel: %v\n", err)
			return
		}

		if response != nil && response.IsError() && response.Error != nil {
			fmt.Printf("Failed to delete VPN tunnel - Status: %d\n", response.StatusCode)
			if response.Error.Title != nil {
				fmt.Printf("Error: %s\n", *response.Error.Title)
			}
			if response.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *response.Error.Detail)
			}
			return
		}

		fmt.Printf("VPN tunnel %s has been successfully deleted.\n", vpnTunnelID)
	},
}
