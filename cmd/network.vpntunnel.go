package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

var vpnEncryptionAlgorithms = []string{
	string(aruba.IKEEncryptionAES128), string(aruba.IKEEncryptionAES192), string(aruba.IKEEncryptionAES256),
	string(aruba.IKEEncryptionAES128CTR), string(aruba.IKEEncryptionAES192CTR), string(aruba.IKEEncryptionAES256CTR),
	string(aruba.IKEEncryptionAES128CCM64), string(aruba.IKEEncryptionAES128CCM96), string(aruba.IKEEncryptionAES128CCM128),
	string(aruba.IKEEncryptionAES192CCM64), string(aruba.IKEEncryptionAES192CCM96), string(aruba.IKEEncryptionAES192CCM128),
	string(aruba.IKEEncryptionAES256CCM64), string(aruba.IKEEncryptionAES256CCM96), string(aruba.IKEEncryptionAES256CCM128),
	string(aruba.IKEEncryptionAES128GCM64), string(aruba.IKEEncryptionAES128GCM96), string(aruba.IKEEncryptionAES128GCM128),
	string(aruba.IKEEncryptionAES192GCM64), string(aruba.IKEEncryptionAES192GCM96), string(aruba.IKEEncryptionAES192GCM128),
	string(aruba.IKEEncryptionAES256GCM64), string(aruba.IKEEncryptionAES256GCM96), string(aruba.IKEEncryptionAES256GCM128),
	string(aruba.IKEEncryptionAES128GMAC), string(aruba.IKEEncryptionAES192GMAC), string(aruba.IKEEncryptionAES256GMAC),
	string(aruba.IKEEncryption3DES),
	string(aruba.IKEEncryptionBlowfish128), string(aruba.IKEEncryptionBlowfish192), string(aruba.IKEEncryptionBlowfish256),
	string(aruba.IKEEncryptionCamellia128), string(aruba.IKEEncryptionCamellia192), string(aruba.IKEEncryptionCamellia256),
	string(aruba.IKEEncryptionCamellia128CTR), string(aruba.IKEEncryptionCamellia192CTR), string(aruba.IKEEncryptionCamellia256CTR),
	string(aruba.IKEEncryptionCamellia128CCM64), string(aruba.IKEEncryptionCamellia128CCM96), string(aruba.IKEEncryptionCamellia128CCM128),
	string(aruba.IKEEncryptionCamellia192CCM64), string(aruba.IKEEncryptionCamellia192CCM96), string(aruba.IKEEncryptionCamellia192CCM128),
	string(aruba.IKEEncryptionCamellia256CCM64), string(aruba.IKEEncryptionCamellia256CCM96), string(aruba.IKEEncryptionCamellia256CCM128),
	string(aruba.IKEEncryptionSerpent128), string(aruba.IKEEncryptionSerpent192), string(aruba.IKEEncryptionSerpent256),
	string(aruba.IKEEncryptionTwofish128), string(aruba.IKEEncryptionTwofish192), string(aruba.IKEEncryptionTwofish256),
	string(aruba.IKEEncryptionCAST128), string(aruba.IKEEncryptionChaCha20Poly1305),
}

var vpnHashAlgorithms = []string{
	string(aruba.IKEHashMD5), string(aruba.IKEHashMD5128),
	string(aruba.IKEHashSHA1), string(aruba.IKEHashSHA1160),
	string(aruba.IKEHashSHA256), string(aruba.IKEHashSHA25696),
	string(aruba.IKEHashSHA384), string(aruba.IKEHashSHA512),
	string(aruba.IKEHashAESXCBC), string(aruba.IKEHashAESCMAC),
	string(aruba.IKEHashAES128GMAC), string(aruba.IKEHashAES192GMAC), string(aruba.IKEHashAES256GMAC),
}

var vpnDHGroups = []string{
	string(aruba.IKEDHGroup1), string(aruba.IKEDHGroup2), string(aruba.IKEDHGroup5),
	string(aruba.IKEDHGroup14), string(aruba.IKEDHGroup15), string(aruba.IKEDHGroup16), string(aruba.IKEDHGroup17), string(aruba.IKEDHGroup18),
	string(aruba.IKEDHGroup19), string(aruba.IKEDHGroup20), string(aruba.IKEDHGroup21),
	string(aruba.IKEDHGroup22), string(aruba.IKEDHGroup23), string(aruba.IKEDHGroup24),
	string(aruba.IKEDHGroup25), string(aruba.IKEDHGroup26), string(aruba.IKEDHGroup27), string(aruba.IKEDHGroup28), string(aruba.IKEDHGroup29), string(aruba.IKEDHGroup30),
	string(aruba.IKEDHGroup31), string(aruba.IKEDHGroup32),
}

var vpnDPDActions = []string{string(aruba.IKEDPDActionTrap), string(aruba.IKEDPDActionClear), string(aruba.IKEDPDActionRestart)}

var vpnPFSGroups = []string{
	string(aruba.ESPPFSGroupEnable),
	string(aruba.ESPPFSGroupDHGroup1), string(aruba.ESPPFSGroupDHGroup2), string(aruba.ESPPFSGroupDHGroup5),
	string(aruba.ESPPFSGroupDHGroup14), string(aruba.ESPPFSGroupDHGroup15), string(aruba.ESPPFSGroupDHGroup16), string(aruba.ESPPFSGroupDHGroup17), string(aruba.ESPPFSGroupDHGroup18),
	string(aruba.ESPPFSGroupDHGroup19), string(aruba.ESPPFSGroupDHGroup20), string(aruba.ESPPFSGroupDHGroup21),
	string(aruba.ESPPFSGroupDHGroup22), string(aruba.ESPPFSGroupDHGroup23), string(aruba.ESPPFSGroupDHGroup24),
	string(aruba.ESPPFSGroupDHGroup25), string(aruba.ESPPFSGroupDHGroup26), string(aruba.ESPPFSGroupDHGroup27), string(aruba.ESPPFSGroupDHGroup28), string(aruba.ESPPFSGroupDHGroup29), string(aruba.ESPPFSGroupDHGroup30),
	string(aruba.ESPPFSGroupDHGroup31), string(aruba.ESPPFSGroupDHGroup32),
	string(aruba.ESPPFSGroupDisable),
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
func redactVPNTunnelSecrets(tunnel *aruba.VPNTunnel) {
	if tunnel == nil || tunnel.Raw() == nil {
		return
	}
	s := tunnel.Raw().Properties.VPNClientSettings
	if s != nil && s.PSK != nil {
		s.PSK.Secret = nil
	}
}

// Completion functions for network resources
func completeVPNTunnelID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().VPNTunnels().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vpn := range list.Items() {
			raw := vpn.Raw()
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
		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromNetwork().VPNTunnels().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing VPN tunnels: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 40},
				{Header: "ID", Width: 25},
				{Header: "REGION", Width: 18},
				{Header: "TYPE", Width: 15},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, vpn := range list.Items() {
				raw := vpn.Raw()
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
				vpnType := ""
				if raw.Properties.VPNType != nil {
					vpnType = string(*raw.Properties.VPNType)
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, region, vpnType, status})
			}
			PrintOutput(list, headers, rows)
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
		vpn, err := client.FromNetwork().VPNTunnels().Get(ctx, aruba.VPNTunnelRef(projectID, vpnID))
		if err != nil {
			return fmt.Errorf("getting VPN tunnel details: %w", apiErrFromV2(err))
		}

		if vpn != nil && vpn.Raw() != nil {
			raw := vpn.Raw()

			fmt.Println("\nVPN Tunnel Details:")
			fmt.Println("===================")

			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if raw.Metadata.LocationResponse != nil && raw.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", raw.Metadata.LocationResponse.Value)
			}
			if raw.Properties.VPNType != nil {
				fmt.Printf("VPN Type:        %s\n", *raw.Properties.VPNType)
			}
			if raw.Properties.VPNClientProtocol != nil {
				fmt.Printf("Protocol:        %s\n", *raw.Properties.VPNClientProtocol)
			}
			if raw.Properties.VPNClientSettings != nil && raw.Properties.VPNClientSettings.PeerClientPublicIP != nil {
				fmt.Printf("Peer IP:         %s\n", *raw.Properties.VPNClientSettings.PeerClientPublicIP)
			}
			if raw.Properties.IPConfigurations != nil {
				fmt.Println("\nIP Configuration:")
				if raw.Properties.IPConfigurations.VPC != nil {
					fmt.Printf("  VPC:           %s\n", raw.Properties.IPConfigurations.VPC.URI)
				}
				if raw.Properties.IPConfigurations.Subnet != nil {
					fmt.Printf("  Subnet CIDR:   %s\n", raw.Properties.IPConfigurations.Subnet.CIDR)
					if raw.Properties.IPConfigurations.Subnet.Name != "" {
						fmt.Printf("  Subnet Name:   %s\n", raw.Properties.IPConfigurations.Subnet.Name)
					}
				}
				if raw.Properties.IPConfigurations.PublicIP != nil {
					fmt.Printf("  Public IP:     %s\n", raw.Properties.IPConfigurations.PublicIP.URI)
				}
			}
			if raw.Properties.BillingPlan != nil && raw.Properties.BillingPlan.BillingPeriod != nil {
				fmt.Printf("\nBilling Period:  %s\n", *raw.Properties.BillingPlan.BillingPeriod)
			}
			if raw.Metadata.CreationDate != nil {
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
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", *raw.Status.State)
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

		if subnetCIDR == "" && subnetName == "" {
			return fmt.Errorf("--subnet-cidr or --subnet-name is required")
		}

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

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		// Build IP config
		ipConfig := aruba.NewVPNIPConfig().
			WithVPC(aruba.URI(vpcURI)).
			WithElasticIP(aruba.URI(publicIPURI)).
			WithSubnet(subnetName, subnetCIDR)

		// Build IKE settings
		ike := aruba.NewVPNIKE().WithLifetimeSeconds(int(ikeLifetime))
		if ikeEncryption != "" {
			ike.WithEncryption(aruba.IKEEncryption(ikeEncryption))
		}
		if ikeHash != "" {
			ike.WithHash(aruba.IKEHash(ikeHash))
		}
		if ikeDHGroup != "" {
			ike.WithDHGroup(aruba.IKEDHGroup(ikeDHGroup))
		}
		if ikeDPDAction != "" {
			ike.WithDPDAction(aruba.IKEDPDAction(ikeDPDAction))
		}
		if ikeDPDInterval > 0 {
			ike.WithDPDIntervalSeconds(int(ikeDPDInterval))
		}
		if ikeDPDTimeout > 0 {
			ike.WithDPDTimeoutSeconds(int(ikeDPDTimeout))
		}

		// Build ESP settings
		esp := aruba.NewVPNESP().
			WithEncryption(aruba.ESPEncryption(espEncryption)).
			WithLifetimeSeconds(int(espLifetime))
		if espHash != "" {
			esp.WithHash(aruba.ESPHash(espHash))
		}
		if espPFS != "" {
			esp.WithPFS(aruba.ESPPFSGroup(espPFS))
		}

		// Build PSK settings
		pskBuilder := aruba.NewVPNPSK()
		if psk != "" {
			pskBuilder.WithKey(psk)
		}
		if pskCloudSite != "" {
			pskBuilder.WithCloudSite(pskCloudSite)
		}
		if pskOnpremSite != "" {
			pskBuilder.WithOnPremSite(pskOnpremSite)
		}

		tunnel := aruba.NewVPNTunnel().
			InProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfType(aruba.VPNType(vpnType)).
			WithVPNClientProtocol(aruba.VPNClientProtocol(protocol)).
			BilledBy(aruba.BillingPeriod(billingPeriod)).
			WithPeerClientPublicIP(peerIP).
			WithIPConfig(ipConfig).
			WithIKESettings(ike).
			WithESPSettings(esp).
			WithPSKSettings(pskBuilder).
			RetaggedAs(tags...)

		ctx, cancel := newCtx()
		defer cancel()
		resp, err := client.FromNetwork().VPNTunnels().Create(ctx, tunnel)
		if err != nil {
			return fmt.Errorf("creating VPN tunnel: %w", apiErrFromV2(err))
		}

		if resp != nil && resp.Raw() != nil {
			raw := resp.Raw()
			fmt.Printf("\n%s\n", msgCreated("VPN Tunnel", name))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:       %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:     %s\n", *raw.Metadata.Name)
			}
			if raw.Metadata.URI != nil {
				fmt.Printf("URI:      %s\n", *raw.Metadata.URI)
			}
			if raw.Properties.VPNType != nil {
				fmt.Printf("Type:     %s\n", *raw.Properties.VPNType)
			}
			if raw.Properties.VPNClientProtocol != nil {
				fmt.Printf("Protocol: %s\n", *raw.Properties.VPNClientProtocol)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:     %v\n", raw.Metadata.Tags)
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

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
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
		vpn, err := client.FromNetwork().VPNTunnels().Get(ctx, aruba.VPNTunnelRef(projectID, vpnTunnelID))
		if err != nil {
			return fmt.Errorf("getting VPN tunnel: %w", apiErrFromV2(err))
		}

		if vpn == nil || vpn.Raw() == nil {
			return fmt.Errorf("VPN tunnel not found")
		}

		// Check if VPN tunnel is in "InCreation" state
		if vpn.Raw().Status.State != nil && *vpn.Raw().Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPN tunnel while it is in 'InCreation' state. Please wait until the VPN tunnel is fully created")
		}

		if name != "" {
			vpn.Named(name)
		}
		if len(tags) > 0 {
			vpn.RetaggedAs(tags...)
		}

		updated, err := client.FromNetwork().VPNTunnels().Update(ctx, vpn)
		if err != nil {
			return fmt.Errorf("updating VPN tunnel: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("VPN Tunnel", vpnTunnelID))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
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

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("VPN tunnel", vpnTunnelID)
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
			_, err = client.FromNetwork().VPNTunnels().Get(ctx, aruba.VPNTunnelRef(projectID, vpnTunnelID))
			if err != nil {
				return fmt.Errorf("dry-run: VPN tunnel not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPN tunnel", vpnTunnelID))
			return nil
		}

		err = client.FromNetwork().VPNTunnels().Delete(ctx, aruba.VPNTunnelRef(projectID, vpnTunnelID))
		if err != nil {
			return fmt.Errorf("deleting VPN tunnel: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("VPN tunnel", vpnTunnelID))
		return nil
	},
}
