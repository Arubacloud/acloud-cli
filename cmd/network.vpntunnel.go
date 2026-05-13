package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

var vpnEncryptionAlgorithms = []string{
	types.VPNEncryptionAES128, types.VPNEncryptionAES192, types.VPNEncryptionAES256,
	types.VPNEncryptionAES128CTR, types.VPNEncryptionAES192CTR, types.VPNEncryptionAES256CTR,
	types.VPNEncryptionAES128CCM64, types.VPNEncryptionAES128CCM96, types.VPNEncryptionAES128CCM128,
	types.VPNEncryptionAES192CCM64, types.VPNEncryptionAES192CCM96, types.VPNEncryptionAES192CCM128,
	types.VPNEncryptionAES256CCM64, types.VPNEncryptionAES256CCM96, types.VPNEncryptionAES256CCM128,
	types.VPNEncryptionAES128GCM64, types.VPNEncryptionAES128GCM96, types.VPNEncryptionAES128GCM128,
	types.VPNEncryptionAES192GCM64, types.VPNEncryptionAES192GCM96, types.VPNEncryptionAES192GCM128,
	types.VPNEncryptionAES256GCM64, types.VPNEncryptionAES256GCM96, types.VPNEncryptionAES256GCM128,
	types.VPNEncryptionAES128GMAC, types.VPNEncryptionAES192GMAC, types.VPNEncryptionAES256GMAC,
	types.VPNEncryption3DES,
	types.VPNEncryptionBlowfish128, types.VPNEncryptionBlowfish192, types.VPNEncryptionBlowfish256,
	types.VPNEncryptionCamellia128, types.VPNEncryptionCamellia192, types.VPNEncryptionCamellia256,
	types.VPNEncryptionCamellia128CTR, types.VPNEncryptionCamellia192CTR, types.VPNEncryptionCamellia256CTR,
	types.VPNEncryptionCamellia128CCM64, types.VPNEncryptionCamellia128CCM96, types.VPNEncryptionCamellia128CCM128,
	types.VPNEncryptionCamellia192CCM64, types.VPNEncryptionCamellia192CCM96, types.VPNEncryptionCamellia192CCM128,
	types.VPNEncryptionCamellia256CCM64, types.VPNEncryptionCamellia256CCM96, types.VPNEncryptionCamellia256CCM128,
	types.VPNEncryptionSerpent128, types.VPNEncryptionSerpent192, types.VPNEncryptionSerpent256,
	types.VPNEncryptionTwofish128, types.VPNEncryptionTwofish192, types.VPNEncryptionTwofish256,
	types.VPNEncryptionCAST128, types.VPNEncryptionChaCha20Poly1305,
}

var vpnHashAlgorithms = []string{
	types.VPNHashMD5, types.VPNHashMD5128,
	types.VPNHashSHA1, types.VPNHashSHA1160,
	types.VPNHashSHA256, types.VPNHashSHA25696,
	types.VPNHashSHA384, types.VPNHashSHA512,
	types.VPNHashAESXCBC, types.VPNHashAESCMAC,
	types.VPNHashAES128GMAC, types.VPNHashAES192GMAC, types.VPNHashAES256GMAC,
}

var vpnDHGroups = []string{
	types.VPNDHGroup1, types.VPNDHGroup2, types.VPNDHGroup5,
	types.VPNDHGroup14, types.VPNDHGroup15, types.VPNDHGroup16, types.VPNDHGroup17, types.VPNDHGroup18,
	types.VPNDHGroup19, types.VPNDHGroup20, types.VPNDHGroup21,
	types.VPNDHGroup22, types.VPNDHGroup23, types.VPNDHGroup24,
	types.VPNDHGroup25, types.VPNDHGroup26, types.VPNDHGroup27, types.VPNDHGroup28, types.VPNDHGroup29, types.VPNDHGroup30,
	types.VPNDHGroup31, types.VPNDHGroup32,
}

var vpnDPDActions = []string{types.VPNDPDActionTrap, types.VPNDPDActionClear, types.VPNDPDActionRestart}

var vpnPFSGroups = []string{
	types.VPNPFSEnable,
	types.VPNPFSDHGroup1, types.VPNPFSDHGroup2, types.VPNPFSDHGroup5,
	types.VPNPFSDHGroup14, types.VPNPFSDHGroup15, types.VPNPFSDHGroup16, types.VPNPFSDHGroup17, types.VPNPFSDHGroup18,
	types.VPNPFSDHGroup19, types.VPNPFSDHGroup20, types.VPNPFSDHGroup21,
	types.VPNPFSDHGroup22, types.VPNPFSDHGroup23, types.VPNPFSDHGroup24,
	types.VPNPFSDHGroup25, types.VPNPFSDHGroup26, types.VPNPFSDHGroup27, types.VPNPFSDHGroup28, types.VPNPFSDHGroup29, types.VPNPFSDHGroup30,
	types.VPNPFSDHGroup31, types.VPNPFSDHGroup32,
	types.VPNPFSDisable,
}

func vpnValidateEnum(value, flag string, valid []string) error {
	if value == "" {
		return nil
	}
	for _, v := range valid {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("--%s %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", flag, value)
}

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
	vpntunnelCreateCmd.Flags().Int32("ike-lifetime", 0, "IKE lifetime in seconds (0-86400)")
	vpntunnelCreateCmd.Flags().String("ike-encryption", "", "IKE encryption algorithm (e.g. aes256; see docs for full list)")
	vpntunnelCreateCmd.Flags().String("ike-hash", "", "IKE hash algorithm (e.g. sha256; see docs for full list)")
	vpntunnelCreateCmd.Flags().String("ike-dh-group", "", "IKE DH group number (1, 2, 5, or 14-32)")
	vpntunnelCreateCmd.Flags().String("ike-dpd-action", "", "IKE DPD action (trap, clear, restart)")
	vpntunnelCreateCmd.Flags().Int32("ike-dpd-interval", 0, "IKE DPD interval in seconds (2-86400)")
	vpntunnelCreateCmd.Flags().Int32("ike-dpd-timeout", 0, "IKE DPD timeout in seconds (2-86400)")
	// ESP settings
	vpntunnelCreateCmd.Flags().Int32("esp-lifetime", 0, "ESP lifetime in seconds (30-86400)")
	vpntunnelCreateCmd.Flags().String("esp-encryption", "", "ESP encryption algorithm (default: aes256; see docs for full list)")
	vpntunnelCreateCmd.Flags().String("esp-hash", "", "ESP hash algorithm (e.g. sha256; see docs for full list)")
	vpntunnelCreateCmd.Flags().String("esp-pfs", "", "ESP PFS group (enable, dh-group1..dh-group32, disable)")
	// PSK settings
	vpntunnelCreateCmd.Flags().String("psk-cloud-site", "", "PSK ID for the Aruba (cloud) side — alphanumeric, '-' and '.', 3-100 chars")
	vpntunnelCreateCmd.Flags().String("psk-onprem-site", "", "PSK ID for the customer (on-prem) side — alphanumeric, '-' and '.', 3-100 chars")
	vpntunnelCreateCmd.Flags().String("psk", "", "Pre-shared key for authentication (PSK secret)")
	vpntunnelCreateCmd.MarkFlagRequired("name")
	vpntunnelCreateCmd.MarkFlagRequired("region")
	vpntunnelCreateCmd.MarkFlagRequired("peer-ip")
	vpntunnelCreateCmd.MarkFlagRequired("vpc-uri")
	vpntunnelCreateCmd.MarkFlagRequired("elastic-ip-uri")
	vpntunnelGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelUpdateCmd.Flags().String("name", "", "New name for the VPN tunnel")
	vpntunnelUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpntunnelDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpntunnelDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	vpntunnelListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpntunnelListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpntunnelListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	vpntunnelGetCmd.ValidArgsFunction = completeVPNTunnelID
	vpntunnelUpdateCmd.ValidArgsFunction = completeVPNTunnelID
	vpntunnelDeleteCmd.ValidArgsFunction = completeVPNTunnelID
}

// redactVPNTunnelSecrets strips PSK.Secret from each tunnel so it cannot leak
// via --output json|yaml. The API should not return the secret on read, but the
// SDK schema declares the field (PSKSettings.Secret) — never trust a type that
// can carry credentials.
func redactVPNTunnelSecrets(tunnels []types.VPNTunnelResponse) {
	for i := range tunnels {
		s := tunnels[i].Properties.VPNClientSettings
		if s != nil && s.PSK != nil {
			s.PSK.Secret = nil
		}
	}
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
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get projectID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// List VPN tunnels using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().VPNTunnels().List(ctx, projectID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing VPN tunnels: %w", err)
		}
		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
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

				region := ""
				if vpn.Metadata.LocationResponse != nil {
					region = vpn.Metadata.LocationResponse.Value
				}

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

			redactVPNTunnelSecrets(response.Data.Values)
			PrintOutput(response.Data, headers, rows)
		} else {
			fmt.Println("No VPN tunnels found")
		}
		return nil
	},
}

var vpntunnelGetCmd = &cobra.Command{
	Use:   "get <vpn-tunnel-id>",
	Short: "Get VPN tunnel details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get VPN tunnel details using the SDK
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().VPNTunnels().Get(ctx, projectID, vpnID, nil)
		if err != nil {
			return fmt.Errorf("getting VPN tunnel details: %w", err)
		}
		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
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
				fmt.Printf("Creation Date:   %s\n", vpn.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var vpntunnelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VPN tunnel",
	Long: `Create a site-to-site VPN tunnel to an on-premises network.

The VPC and Elastic IP must already exist. Specify the subnet the tunnel will
use via --subnet-cidr (CIDR of existing subnet) or --subnet-name.

VPN type defaults to Site-To-Site. Protocol defaults to ikev2.
Billing period: Hour (default), Month, or Year.

IKE and ESP settings are optional; the platform uses secure defaults when omitted.`,
	Example: `  acloud network vpntunnel create \
    --name my-tunnel --region IT-BG \
    --peer-ip 203.0.113.1 \
    --vpc-uri /projects/<proj-id>/providers/Aruba.Network/vpcs/<vpc-id> \
    --subnet-cidr 10.0.1.0/24 \
    --elastic-ip-uri /projects/<proj-id>/providers/Aruba.Network/elasticIPs/<eip-id> \
    --psk my-pre-shared-key`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Validate mutual-exclusive subnet flags
		if subnetCIDR == "" && subnetName == "" {
			return fmt.Errorf("--subnet-cidr or --subnet-name is required")
		}

		// Validate enum fields client-side to surface typos before the API call.
		for _, check := range []struct {
			val   string
			flag  string
			valid []string
		}{
			{ikeEncryption, "ike-encryption", vpnEncryptionAlgorithms},
			{ikeHash, "ike-hash", vpnHashAlgorithms},
			{ikeDHGroup, "ike-dh-group", vpnDHGroups},
			{ikeDPDAction, "ike-dpd-action", vpnDPDActions},
			{espEncryption, "esp-encryption", vpnEncryptionAlgorithms},
			{espHash, "esp-hash", vpnHashAlgorithms},
			{espPFS, "esp-pfs", vpnPFSGroups},
		} {
			if err := vpnValidateEnum(check.val, check.flag, check.valid); err != nil {
				return err
			}
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
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
		ctx, cancel := newCtx()
		defer cancel()
		response, err := client.FromNetwork().VPNTunnels().Create(ctx, projectID, createRequest, nil)
		if err != nil {
			return fmt.Errorf("creating VPN tunnel: %w", err)
		}

		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		if response != nil && response.Data != nil {
			fmt.Printf("\n%s\n", msgCreated("VPN Tunnel", name))
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
			fmt.Println(msgCreatedAsync("VPN Tunnel", name))
		}
		return nil
	},
}

var vpntunnelUpdateCmd = &cobra.Command{
	Use:   "update <vpn-tunnel-id>",
	Short: "Update a VPN tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one update flag must be provided
		if name == "" && len(tags) == 0 {
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		// Get project ID
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get Aruba Cloud client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Get current VPN tunnel configuration
		ctx, cancel := newCtx()
		defer cancel()
		getResp, err := client.FromNetwork().VPNTunnels().Get(ctx, projectID, vpnTunnelID, nil)
		if err != nil {
			return fmt.Errorf("getting VPN tunnel: %w", err)
		}

		if getResp != nil && getResp.IsError() {
			return apiErrFromResp(getResp.StatusCode, getResp.Error)
		}

		if getResp.Data == nil {
			return fmt.Errorf("VPN tunnel not found")
		}

		// Check if VPN tunnel is in "InCreation" state
		if getResp.Data.Status.State != nil && *getResp.Data.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPN tunnel while it is in 'InCreation' state. Please wait until the VPN tunnel is fully created")
		}

		// Get region value
		regionValue := ""
		if getResp.Data.Metadata.LocationResponse != nil {
			regionValue = getResp.Data.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for VPN tunnel")
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
			return fmt.Errorf("updating VPN tunnel: %w", err)
		}

		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}

		if resp.Data != nil {
			fmt.Printf("\n%s\n", msgUpdated("VPN Tunnel", vpnTunnelID))
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
			fmt.Println(msgUpdatedAsync("VPN Tunnel", vpnTunnelID))
		}
		return nil
	},
}

var vpntunnelDeleteCmd = &cobra.Command{
	Use:   "delete <vpn-tunnel-id>",
	Short: "Delete a VPN tunnel",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpnTunnelID := args[0]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			ok, err := confirmDelete("VPN tunnel", vpnTunnelID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromNetwork().VPNTunnels().Get(ctx, projectID, vpnTunnelID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: VPN tunnel not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("VPN tunnel", vpnTunnelID))
			return nil
		}

		// Delete the VPN tunnel using the SDK
		response, err := client.FromNetwork().VPNTunnels().Delete(ctx, projectID, vpnTunnelID, nil)
		if err != nil {
			return fmt.Errorf("deleting VPN tunnel: %w", err)
		}

		if response != nil && response.IsError() {
			return apiErrFromResp(response.StatusCode, response.Error)
		}

		fmt.Println(msgDeleted("VPN tunnel", vpnTunnelID))
		return nil
	},
}
