package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// vpnTunnelRef is shared with vpnroute.go.
func vpnTunnelRef(projectID, tunnelID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/vpnTunnels/" + tunnelID)
}

func vpnTunnelListPayload(l *aruba.List[*aruba.VPNTunnel]) any {
	if r, ok := l.Raw().(*types.Response[types.VPNTunnelList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// vpnTunnelReattachSettings reconstructs IKE/ESP/PSK/IPConfig sub-builders from
// the raw response and re-attaches them to the wrapper before Update. The wrapper's
// fromResponse does not rehydrate sub-builders, so a naive Update would drop them.
func vpnTunnelReattachSettings(cur *aruba.VPNTunnel) {
	r := cur.Raw()
	if r == nil {
		return
	}
	if cfg := r.Properties.IPConfigurations; cfg != nil {
		ipc := aruba.NewVPNIPConfig()
		if cfg.VPC != nil {
			ipc.WithVPC(aruba.URI(cfg.VPC.URI))
		}
		if cfg.PublicIP != nil {
			ipc.WithElasticIP(aruba.URI(cfg.PublicIP.URI))
		}
		if cfg.Subnet != nil {
			ipc.WithSubnet(cfg.Subnet.Name, cfg.Subnet.CIDR)
		}
		cur.WithIPConfig(ipc)
	}
	if cs := r.Properties.VPNClientSettings; cs != nil {
		if ike := cs.IKE; ike != nil {
			k := aruba.NewVPNIKE().
				WithLifetimeSeconds(int(ike.Lifetime)).
				WithDPDIntervalSeconds(int(ike.DPDInterval)).
				WithDPDTimeoutSeconds(int(ike.DPDTimeout))
			if ike.Encryption != nil {
				k.WithEncryption(*ike.Encryption)
			}
			if ike.Hash != nil {
				k.WithHash(*ike.Hash)
			}
			if ike.DHGroup != nil {
				k.WithDHGroup(*ike.DHGroup)
			}
			if ike.DPDAction != nil {
				k.WithDPDAction(*ike.DPDAction)
			}
			cur.WithIKESettings(k)
		}
		if esp := cs.ESP; esp != nil {
			e := aruba.NewVPNESP().WithLifetimeSeconds(int(esp.Lifetime))
			if esp.Encryption != nil {
				e.WithEncryption(*esp.Encryption)
			}
			if esp.Hash != nil {
				e.WithHash(*esp.Hash)
			}
			if esp.PFS != nil {
				e.WithPFS(*esp.PFS)
			}
			cur.WithESPSettings(e)
		}
		if psk := cs.PSK; psk != nil {
			p := aruba.NewVPNPSK()
			if psk.Secret != nil {
				p.WithKey(*psk.Secret)
			}
			if psk.CloudSite != nil {
				p.WithCloudSite(*psk.CloudSite)
			}
			if psk.OnPremSite != nil {
				p.WithOnPremSite(*psk.OnPremSite)
			}
			cur.WithPSKSettings(p)
		}
	}
}

// IKE-specific enum slices
var vpnIKEEncryptionAlgorithms = []string{
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

// ESP-specific encryption slice (same algorithms, separate type)
var vpnESPEncryptionAlgorithms = []string{
	string(aruba.ESPEncryptionAES128), string(aruba.ESPEncryptionAES192), string(aruba.ESPEncryptionAES256),
	string(aruba.ESPEncryptionAES128CTR), string(aruba.ESPEncryptionAES192CTR), string(aruba.ESPEncryptionAES256CTR),
	string(aruba.ESPEncryptionAES128CCM64), string(aruba.ESPEncryptionAES128CCM96), string(aruba.ESPEncryptionAES128CCM128),
	string(aruba.ESPEncryptionAES192CCM64), string(aruba.ESPEncryptionAES192CCM96), string(aruba.ESPEncryptionAES192CCM128),
	string(aruba.ESPEncryptionAES256CCM64), string(aruba.ESPEncryptionAES256CCM96), string(aruba.ESPEncryptionAES256CCM128),
	string(aruba.ESPEncryptionAES128GCM64), string(aruba.ESPEncryptionAES128GCM96), string(aruba.ESPEncryptionAES128GCM128),
	string(aruba.ESPEncryptionAES192GCM64), string(aruba.ESPEncryptionAES192GCM96), string(aruba.ESPEncryptionAES192GCM128),
	string(aruba.ESPEncryptionAES256GCM64), string(aruba.ESPEncryptionAES256GCM96), string(aruba.ESPEncryptionAES256GCM128),
	string(aruba.ESPEncryptionAES128GMAC), string(aruba.ESPEncryptionAES192GMAC), string(aruba.ESPEncryptionAES256GMAC),
	string(aruba.ESPEncryption3DES),
	string(aruba.ESPEncryptionBlowfish128), string(aruba.ESPEncryptionBlowfish192), string(aruba.ESPEncryptionBlowfish256),
	string(aruba.ESPEncryptionCamellia128), string(aruba.ESPEncryptionCamellia192), string(aruba.ESPEncryptionCamellia256),
	string(aruba.ESPEncryptionCamellia128CTR), string(aruba.ESPEncryptionCamellia192CTR), string(aruba.ESPEncryptionCamellia256CTR),
	string(aruba.ESPEncryptionCamellia128CCM64), string(aruba.ESPEncryptionCamellia128CCM96), string(aruba.ESPEncryptionCamellia128CCM128),
	string(aruba.ESPEncryptionCamellia192CCM64), string(aruba.ESPEncryptionCamellia192CCM96), string(aruba.ESPEncryptionCamellia192CCM128),
	string(aruba.ESPEncryptionCamellia256CCM64), string(aruba.ESPEncryptionCamellia256CCM96), string(aruba.ESPEncryptionCamellia256CCM128),
	string(aruba.ESPEncryptionSerpent128), string(aruba.ESPEncryptionSerpent192), string(aruba.ESPEncryptionSerpent256),
	string(aruba.ESPEncryptionTwofish128), string(aruba.ESPEncryptionTwofish192), string(aruba.ESPEncryptionTwofish256),
	string(aruba.ESPEncryptionCAST128), string(aruba.ESPEncryptionChaCha20Poly1305),
}

var vpnIKEHashAlgorithms = []string{
	string(aruba.IKEHashMD5), string(aruba.IKEHashMD5128),
	string(aruba.IKEHashSHA1), string(aruba.IKEHashSHA1160),
	string(aruba.IKEHashSHA256), string(aruba.IKEHashSHA25696),
	string(aruba.IKEHashSHA384), string(aruba.IKEHashSHA512),
	string(aruba.IKEHashAESXCBC), string(aruba.IKEHashAESCMAC),
	string(aruba.IKEHashAES128GMAC), string(aruba.IKEHashAES192GMAC), string(aruba.IKEHashAES256GMAC),
}

var vpnESPHashAlgorithms = []string{
	string(aruba.ESPHashMD5), string(aruba.ESPHashMD5128),
	string(aruba.ESPHashSHA1), string(aruba.ESPHashSHA1160),
	string(aruba.ESPHashSHA256), string(aruba.ESPHashSHA25696),
	string(aruba.ESPHashSHA384), string(aruba.ESPHashSHA512),
	string(aruba.ESPHashAESXCBC), string(aruba.ESPHashAESCMAC),
	string(aruba.ESPHashAES128GMAC), string(aruba.ESPHashAES192GMAC), string(aruba.ESPHashAES256GMAC),
}

var vpnIKEDHGroups = []string{
	string(aruba.IKEDHGroup1), string(aruba.IKEDHGroup2), string(aruba.IKEDHGroup5),
	string(aruba.IKEDHGroup14), string(aruba.IKEDHGroup15), string(aruba.IKEDHGroup16),
	string(aruba.IKEDHGroup17), string(aruba.IKEDHGroup18),
	string(aruba.IKEDHGroup19), string(aruba.IKEDHGroup20), string(aruba.IKEDHGroup21),
	string(aruba.IKEDHGroup22), string(aruba.IKEDHGroup23), string(aruba.IKEDHGroup24),
	string(aruba.IKEDHGroup25), string(aruba.IKEDHGroup26), string(aruba.IKEDHGroup27),
	string(aruba.IKEDHGroup28), string(aruba.IKEDHGroup29), string(aruba.IKEDHGroup30),
	string(aruba.IKEDHGroup31), string(aruba.IKEDHGroup32),
}

var vpnIKEDPDActions = []string{
	string(aruba.IKEDPDActionTrap), string(aruba.IKEDPDActionClear), string(aruba.IKEDPDActionRestart),
}

var vpnESPPFSGroups = []string{
	string(aruba.ESPPFSGroupEnable),
	string(aruba.ESPPFSGroupDHGroup1), string(aruba.ESPPFSGroupDHGroup2), string(aruba.ESPPFSGroupDHGroup5),
	string(aruba.ESPPFSGroupDHGroup14), string(aruba.ESPPFSGroupDHGroup15), string(aruba.ESPPFSGroupDHGroup16),
	string(aruba.ESPPFSGroupDHGroup17), string(aruba.ESPPFSGroupDHGroup18),
	string(aruba.ESPPFSGroupDHGroup19), string(aruba.ESPPFSGroupDHGroup20), string(aruba.ESPPFSGroupDHGroup21),
	string(aruba.ESPPFSGroupDHGroup22), string(aruba.ESPPFSGroupDHGroup23), string(aruba.ESPPFSGroupDHGroup24),
	string(aruba.ESPPFSGroupDHGroup25), string(aruba.ESPPFSGroupDHGroup26), string(aruba.ESPPFSGroupDHGroup27),
	string(aruba.ESPPFSGroupDHGroup28), string(aruba.ESPPFSGroupDHGroup29), string(aruba.ESPPFSGroupDHGroup30),
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
func redactVPNTunnelSecrets(tunnels []types.VPNTunnelResponse) {
	for i := range tunnels {
		s := tunnels[i].Properties.VPNClientSettings
		if s != nil && s.PSK != nil {
			s.PSK.Secret = nil
		}
	}
}

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
	list, err := client.FromNetwork().VPNTunnels().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vpn := range list.Items() {
			id := vpn.VPNTunnelID()
			name := vpn.Name()
			if id == "" {
				continue
			}
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
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
		list, err := client.FromNetwork().VPNTunnels().List(ctx, projectRef(projectID), listOpts(cmd)...)
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
				r := vpn.Raw()
				name := vpn.Name()
				id := vpn.VPNTunnelID()
				region := ""
				vpnType := ""
				status := ""
				if r != nil {
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					if r.Properties.VPNType != nil {
						vpnType = string(*r.Properties.VPNType)
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}
				rows = append(rows, []string{name, id, region, vpnType, status})
			}

			payload := vpnTunnelListPayload(list)
			if pl, ok := payload.(types.VPNTunnelList); ok {
				redactVPNTunnelSecrets(pl.Values)
			}
			PrintOutput(payload, headers, rows)
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
		got, err := client.FromNetwork().VPNTunnels().Get(ctx, vpnTunnelRef(projectID, vpnID))
		if err != nil {
			return fmt.Errorf("getting VPN tunnel details: %w", apiErrFromV2(err))
		}

		vpn := got.Raw()
		if vpn != nil {
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
				fmt.Printf("VPN Type:        %s\n", string(*vpn.Properties.VPNType))
			}
			if vpn.Properties.VPNClientProtocol != nil {
				fmt.Printf("Protocol:        %s\n", string(*vpn.Properties.VPNClientProtocol))
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

			if vpn.Properties.BillingPeriod != nil {
				fmt.Printf("\nBilling Period:  %s\n", string(*vpn.Properties.BillingPeriod))
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
		pskSecret, _ := cmd.Flags().GetString("psk")

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
			{ikeEncryption, "ike-encryption", vpnIKEEncryptionAlgorithms},
			{ikeHash, "ike-hash", vpnIKEHashAlgorithms},
			{ikeDHGroup, "ike-dh-group", vpnIKEDHGroups},
			{ikeDPDAction, "ike-dpd-action", vpnIKEDPDActions},
			{espEncryption, "esp-encryption", vpnESPEncryptionAlgorithms},
			{espHash, "esp-hash", vpnESPHashAlgorithms},
			{espPFS, "esp-pfs", vpnESPPFSGroups},
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

		t := aruba.NewVPNTunnel().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfType(aruba.VPNType(vpnType)).
			WithVPNClientProtocol(aruba.VPNClientProtocol(protocol)).
			WithBillingPeriod(aruba.BillingPeriod(billingPeriod)).
			WithPeerClientPublicIP(peerIP)
		if len(tags) > 0 {
			t.ReplaceTags(tags...)
		}

		ipc := aruba.NewVPNIPConfig().
			WithVPC(aruba.URI(vpcURI)).
			WithElasticIP(aruba.URI(publicIPURI))
		if subnetCIDR != "" || subnetName != "" {
			ipc.WithSubnet(subnetName, subnetCIDR)
		}
		t.WithIPConfig(ipc)

		ike := aruba.NewVPNIKE().
			WithLifetimeSeconds(int(ikeLifetime)).
			WithDPDIntervalSeconds(int(ikeDPDInterval)).
			WithDPDTimeoutSeconds(int(ikeDPDTimeout))
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
		t.WithIKESettings(ike)

		esp := aruba.NewVPNESP().
			WithLifetimeSeconds(int(espLifetime)).
			WithEncryption(aruba.ESPEncryption(espEncryption))
		if espHash != "" {
			esp.WithHash(aruba.ESPHash(espHash))
		}
		if espPFS != "" {
			esp.WithPFS(aruba.ESPPFSGroup(espPFS))
		}
		t.WithESPSettings(esp)

		pskB := aruba.NewVPNPSK()
		if pskSecret != "" {
			pskB.WithKey(pskSecret)
		}
		if pskCloudSite != "" {
			pskB.WithCloudSite(pskCloudSite)
		}
		if pskOnpremSite != "" {
			pskB.WithOnPremSite(pskOnpremSite)
		}
		t.WithPSKSettings(pskB)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromNetwork().VPNTunnels().Create(ctx, t)
		if err != nil {
			return fmt.Errorf("creating VPN tunnel: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil {
			fmt.Printf("\n%s\n", msgCreated("VPN Tunnel", name))
			if r.Metadata.ID != nil {
				fmt.Printf("ID:       %s\n", *r.Metadata.ID)
			}
			if r.Metadata.Name != nil {
				fmt.Printf("Name:     %s\n", *r.Metadata.Name)
			}
			if r.Metadata.URI != nil {
				fmt.Printf("URI:      %s\n", *r.Metadata.URI)
			}
			if r.Properties.VPNType != nil {
				fmt.Printf("Type:     %s\n", string(*r.Properties.VPNType))
			}
			if r.Properties.VPNClientProtocol != nil {
				fmt.Printf("Protocol: %s\n", string(*r.Properties.VPNClientProtocol))
			}
			if len(r.Metadata.Tags) > 0 {
				fmt.Printf("Tags:     %v\n", r.Metadata.Tags)
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
		cur, err := client.FromNetwork().VPNTunnels().Get(ctx, vpnTunnelRef(projectID, vpnTunnelID))
		if err != nil {
			return fmt.Errorf("getting VPN tunnel: %w", apiErrFromV2(err))
		}

		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("VPN tunnel not found")
		}
		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPN tunnel while it is in 'InCreation' state. Please wait until the VPN tunnel is fully created")
		}

		// Reattach IKE/ESP/PSK/IPConfig sub-builders before Update — fromResponse does
		// not rehydrate sub-builders, so omitting this would drop them from the PUT body.
		vpnTunnelReattachSettings(cur)

		if name != "" {
			cur.Named(name)
		}
		if len(tags) > 0 {
			cur.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().VPNTunnels().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating VPN tunnel: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			fmt.Printf("\n%s\n", msgUpdated("VPN Tunnel", vpnTunnelID))
			if ur.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *ur.Metadata.ID)
			}
			if ur.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *ur.Metadata.Name)
			}
			if len(ur.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", ur.Metadata.Tags)
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
			_, err = client.FromNetwork().VPNTunnels().Get(ctx, vpnTunnelRef(projectID, vpnTunnelID))
			if err != nil {
				return fmt.Errorf("dry-run: VPN tunnel not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPN tunnel", vpnTunnelID))
			return nil
		}

		if err := client.FromNetwork().VPNTunnels().Delete(ctx, vpnTunnelRef(projectID, vpnTunnelID)); err != nil {
			return fmt.Errorf("deleting VPN tunnel: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("VPN tunnel", vpnTunnelID))
		return nil
	},
}
