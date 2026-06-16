package cmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// vpnIKEEncryptionAlgorithms is the set of valid IKE encryption algorithm identifiers.
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

// vpnESPEncryptionAlgorithms is the set of valid ESP encryption algorithm identifiers.
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

// vpnIKEHashAlgorithms is the set of valid IKE hash/PRF algorithm identifiers.
var vpnIKEHashAlgorithms = []string{
	string(aruba.IKEHashMD5), string(aruba.IKEHashMD5128),
	string(aruba.IKEHashSHA1), string(aruba.IKEHashSHA1160),
	string(aruba.IKEHashSHA256), string(aruba.IKEHashSHA25696),
	string(aruba.IKEHashSHA384), string(aruba.IKEHashSHA512),
	string(aruba.IKEHashAESXCBC), string(aruba.IKEHashAESCMAC),
	string(aruba.IKEHashAES128GMAC), string(aruba.IKEHashAES192GMAC), string(aruba.IKEHashAES256GMAC),
}

// vpnESPHashAlgorithms is the set of valid ESP integrity/authentication algorithm identifiers.
var vpnESPHashAlgorithms = []string{
	string(aruba.ESPHashMD5), string(aruba.ESPHashMD5128),
	string(aruba.ESPHashSHA1), string(aruba.ESPHashSHA1160),
	string(aruba.ESPHashSHA256), string(aruba.ESPHashSHA25696),
	string(aruba.ESPHashSHA384), string(aruba.ESPHashSHA512),
	string(aruba.ESPHashAESXCBC), string(aruba.ESPHashAESCMAC),
	string(aruba.ESPHashAES128GMAC), string(aruba.ESPHashAES192GMAC), string(aruba.ESPHashAES256GMAC),
}

// vpnIKEDHGroups is the set of valid IKE Diffie-Hellman group identifiers.
var vpnIKEDHGroups = []string{
	string(aruba.IKEDHGroup1), string(aruba.IKEDHGroup2), string(aruba.IKEDHGroup5),
	string(aruba.IKEDHGroup14), string(aruba.IKEDHGroup15), string(aruba.IKEDHGroup16), string(aruba.IKEDHGroup17), string(aruba.IKEDHGroup18),
	string(aruba.IKEDHGroup19), string(aruba.IKEDHGroup20), string(aruba.IKEDHGroup21),
	string(aruba.IKEDHGroup22), string(aruba.IKEDHGroup23), string(aruba.IKEDHGroup24),
	string(aruba.IKEDHGroup25), string(aruba.IKEDHGroup26), string(aruba.IKEDHGroup27), string(aruba.IKEDHGroup28), string(aruba.IKEDHGroup29), string(aruba.IKEDHGroup30),
	string(aruba.IKEDHGroup31), string(aruba.IKEDHGroup32),
}

// vpnIKEDPDActions is the set of valid IKE Dead Peer Detection action identifiers.
var vpnIKEDPDActions = []string{string(aruba.IKEDPDActionTrap), string(aruba.IKEDPDActionClear), string(aruba.IKEDPDActionRestart)}

// vpnESPPFSGroups is the set of valid ESP Perfect Forward Secrecy group identifiers.
var vpnESPPFSGroups = []string{
	string(aruba.ESPPFSGroupEnable),
	string(aruba.ESPPFSGroupDHGroup1), string(aruba.ESPPFSGroupDHGroup2), string(aruba.ESPPFSGroupDHGroup5),
	string(aruba.ESPPFSGroupDHGroup14), string(aruba.ESPPFSGroupDHGroup15), string(aruba.ESPPFSGroupDHGroup16), string(aruba.ESPPFSGroupDHGroup17), string(aruba.ESPPFSGroupDHGroup18),
	string(aruba.ESPPFSGroupDHGroup19), string(aruba.ESPPFSGroupDHGroup20), string(aruba.ESPPFSGroupDHGroup21),
	string(aruba.ESPPFSGroupDHGroup22), string(aruba.ESPPFSGroupDHGroup23), string(aruba.ESPPFSGroupDHGroup24),
	string(aruba.ESPPFSGroupDHGroup25), string(aruba.ESPPFSGroupDHGroup26), string(aruba.ESPPFSGroupDHGroup27), string(aruba.ESPPFSGroupDHGroup28), string(aruba.ESPPFSGroupDHGroup29), string(aruba.ESPPFSGroupDHGroup30),
	string(aruba.ESPPFSGroupDHGroup31), string(aruba.ESPPFSGroupDHGroup32),
	string(aruba.ESPPFSGroupDisable),
}

// vpnTunnelRef builds the combined-URI Ref for a VPN tunnel.
// Used by network.vpnroute.go for VPNRoute refs that encode the tunnel ancestry.
func vpnTunnelRef(projectID, tunnelID string) aruba.Ref {
	return aruba.VPNTunnelRef(projectID, tunnelID)
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
	vpntunnelCreateCmd.Flags().String("vpc-id", "", "VPC ID")
	vpntunnelCreateCmd.Flags().String("subnet-cidr", "", "CIDR of the routing subnet (must already exist in the VPC and be unique per tunnel)")
	vpntunnelCreateCmd.Flags().String("subnet-name", "", "Name of the routing subnet (alternative lookup to --subnet-cidr)")
	vpntunnelCreateCmd.Flags().String("elastic-ip-id", "", "Elastic IP ID")
	vpntunnelCreateCmd.Flags().String("vpn-type", "Site-To-Site", "VPN type (default: Site-To-Site)")
	vpntunnelCreateCmd.Flags().String("protocol", "ikev2", "VPN protocol (default: ikev2)")
	vpntunnelCreateCmd.Flags().String("billing-period", string(aruba.BillingPeriodHour), "Billing period: Hour, Month, Year")
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
	vpntunnelCreateCmd.MarkFlagRequired("vpc-id")
	vpntunnelCreateCmd.MarkFlagRequired("elastic-ip-id")
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
// SDK schema declares the field (PSKSettingsCommon.Secret) — never trust a type that
// can carry credentials.
// Note: RawJSON()/RawYAML() serialize t.response (the raw struct), so we must mutate
// it directly. The TECH_DEBT (#132) for a wrapper-level RedactPSK() mutator remains.
func redactVPNTunnelSecrets(tunnel *aruba.VPNTunnel) {
	if tunnel == nil || tunnel.Raw() == nil {
		return
	}
	s := tunnel.Raw().Properties.VPNClientSettingsCommon
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

	key := cacheKey("vpntunnel", projectID)
	if cached, _ := completionCacheGet(key); cached != nil {
		return filterCompletions(cached, toComplete), cobra.ShellCompDirectiveNoFileComp
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
			id := vpn.ID()
			if id != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, vpn.Name()))
			}
		}
	}

	completionCachePut(key, completions)
	return filterCompletions(completions, toComplete), cobra.ShellCompDirectiveNoFileComp
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
	RunE:  NetworkVPNTunnelListRun,
}

var vpntunnelGetCmd = &cobra.Command{
	Use:   "get <vpn-tunnel-id>",
	Short: "Get VPN tunnel details",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPNTunnelGetRun,
}

var vpntunnelCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VPN tunnel",
	Long: `Create a site-to-site VPN tunnel to an on-premises network.

The VPC and Elastic IP must already exist. Specify the routing subnet via
--subnet-cidr or --subnet-name. The subnet must already exist in the VPC and
its CIDR must be unique across all VPN tunnels in the project — it is a routing
reference, not a provisioning instruction. The 400 "overlaps" error means the
same CIDR is already associated with another tunnel configuration.

VPN type defaults to Site-To-Site. Protocol defaults to ikev2.
Billing period: Hour (default), Month, or Year.

IKE and ESP settings are optional; the platform uses secure defaults when omitted.`,
	Example: `  acloud network vpntunnel create \
    --name my-tunnel --region ITBG-Bergamo \
    --peer-ip 203.0.113.1 \
    --vpc-id <vpc-id> \
    --subnet-cidr 10.0.1.0/24 \
    --elastic-ip-id <eip-id> \
    --psk my-pre-shared-key`,
	Args: cobra.NoArgs,
	RunE: NetworkVPNTunnelCreateRun,
}

var vpntunnelUpdateCmd = &cobra.Command{
	Use:   "update <vpn-tunnel-id>",
	Short: "Update a VPN tunnel",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPNTunnelUpdateRun,
}

var vpntunnelDeleteCmd = &cobra.Command{
	Use:   "delete <vpn-tunnel-id>",
	Short: "Delete a VPN tunnel",
	Args:  cobra.ExactArgs(1),
	RunE:  NetworkVPNTunnelDeleteRun,
}

// =============================================================================
// Args structs
// =============================================================================

// NetworkVPNTunnelCreateArgs holds the typed arguments for creating a VPN tunnel.
type NetworkVPNTunnelCreateArgs struct {
	ProjectID     string
	Name          string
	Region        aruba.Region
	Tags          []string
	VPNType       aruba.VPNType
	Protocol      aruba.VPNClientProtocol
	PeerIP        string
	VPCID         string
	SubnetCIDR    string
	SubnetName    string
	ElasticIPID   string
	BillingPeriod aruba.BillingPeriod
	// IKE settings
	IKELifetime    int32
	IKEEncryption  aruba.IKEEncryption
	IKEHash        aruba.IKEHash
	IKEDHGroup     aruba.IKEDHGroup
	IKEDPDAction   aruba.IKEDPDAction
	IKEDPDInterval int32
	IKEDPDTimeout  int32
	// ESP settings
	ESPLifetime   int32
	ESPEncryption aruba.ESPEncryption
	ESPHash       aruba.ESPHash
	ESPPFS        aruba.ESPPFSGroup
	// PSK settings
	PSK           string
	PSKCloudSite  string
	PSKOnpremSite string
}

// NetworkVPNTunnelGetArgs holds the typed arguments for getting a VPN tunnel.
type NetworkVPNTunnelGetArgs struct {
	ProjectID string
	ID        string
}

// NetworkVPNTunnelUpdateArgs holds the typed arguments for updating a VPN tunnel.
type NetworkVPNTunnelUpdateArgs struct {
	ProjectID   string
	ID          string
	Name        string
	Tags        []string
	TagsChanged bool
}

// NetworkVPNTunnelDeleteArgs holds the typed arguments for deleting a VPN tunnel.
type NetworkVPNTunnelDeleteArgs struct {
	ProjectID   string
	ID          string
	DryRun      bool
	SkipConfirm bool
}

// NetworkVPNTunnelListArgs holds the typed arguments for listing VPN tunnels.
type NetworkVPNTunnelListArgs struct {
	ProjectID string
	CallOpts  []aruba.CallOption
}

// =============================================================================
// Constructors
// =============================================================================

// NewNetworkVPNTunnelCreateArgsFromCobraCommand parses and validates args for create.
func NewNetworkVPNTunnelCreateArgsFromCobraCommand(cmd *cobra.Command) (*NetworkVPNTunnelCreateArgs, error) {
	args := &NetworkVPNTunnelCreateArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNTunnelGetArgsFromCobraCommand parses and validates args for get.
func NewNetworkVPNTunnelGetArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNTunnelGetArgs, error) {
	args := &NetworkVPNTunnelGetArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNTunnelUpdateArgsFromCobraCommand parses and validates args for update.
func NewNetworkVPNTunnelUpdateArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNTunnelUpdateArgs, error) {
	args := &NetworkVPNTunnelUpdateArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNTunnelDeleteArgsFromCobraCommand parses and validates args for delete.
func NewNetworkVPNTunnelDeleteArgsFromCobraCommand(cmd *cobra.Command, cobraArgs []string) (*NetworkVPNTunnelDeleteArgs, error) {
	args := &NetworkVPNTunnelDeleteArgs{}
	if err := args.ParseFromCobraCommand(cmd, cobraArgs); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// NewNetworkVPNTunnelListArgsFromCobraCommand parses and validates args for list.
func NewNetworkVPNTunnelListArgsFromCobraCommand(cmd *cobra.Command) (*NetworkVPNTunnelListArgs, error) {
	args := &NetworkVPNTunnelListArgs{}
	if err := args.ParseFromCobraCommand(cmd); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrParsingFailed, err)
	}
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("%w: [%w]", ErrValidationFailed, err)
	}
	return args, nil
}

// =============================================================================
// ParseFromCobraCommand methods
// =============================================================================

// ParseFromCobraCommand reads Cobra flags into the create args struct.
func (a *NetworkVPNTunnelCreateArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("region"); err == nil {
		a.Region = aruba.Region(s)
	} else {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("vpn-type"); err == nil {
		a.VPNType = aruba.VPNType(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("protocol"); err == nil {
		a.Protocol = aruba.VPNClientProtocol(s)
	} else {
		errs = append(errs, err)
	}
	if a.PeerIP, err = cmd.Flags().GetString("peer-ip"); err != nil {
		errs = append(errs, err)
	}
	if a.VPCID, err = cmd.Flags().GetString("vpc-id"); err != nil {
		errs = append(errs, err)
	}
	if a.SubnetCIDR, err = cmd.Flags().GetString("subnet-cidr"); err != nil {
		errs = append(errs, err)
	}
	if a.SubnetName, err = cmd.Flags().GetString("subnet-name"); err != nil {
		errs = append(errs, err)
	}
	if a.ElasticIPID, err = cmd.Flags().GetString("elastic-ip-id"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("billing-period"); err == nil {
		a.BillingPeriod = aruba.BillingPeriod(s)
	} else {
		errs = append(errs, err)
	}
	// IKE settings
	if a.IKELifetime, err = cmd.Flags().GetInt32("ike-lifetime"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("ike-encryption"); err == nil {
		a.IKEEncryption = aruba.IKEEncryption(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("ike-hash"); err == nil {
		a.IKEHash = aruba.IKEHash(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("ike-dh-group"); err == nil {
		a.IKEDHGroup = aruba.IKEDHGroup(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("ike-dpd-action"); err == nil {
		a.IKEDPDAction = aruba.IKEDPDAction(s)
	} else {
		errs = append(errs, err)
	}
	if a.IKEDPDInterval, err = cmd.Flags().GetInt32("ike-dpd-interval"); err != nil {
		errs = append(errs, err)
	}
	if a.IKEDPDTimeout, err = cmd.Flags().GetInt32("ike-dpd-timeout"); err != nil {
		errs = append(errs, err)
	}
	// ESP settings
	if a.ESPLifetime, err = cmd.Flags().GetInt32("esp-lifetime"); err != nil {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("esp-encryption"); err == nil {
		if s == "" {
			s = "aes256"
		}
		a.ESPEncryption = aruba.ESPEncryption(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("esp-hash"); err == nil {
		a.ESPHash = aruba.ESPHash(s)
	} else {
		errs = append(errs, err)
	}
	if s, err := cmd.Flags().GetString("esp-pfs"); err == nil {
		a.ESPPFS = aruba.ESPPFSGroup(s)
	} else {
		errs = append(errs, err)
	}
	// PSK settings
	if a.PSK, err = cmd.Flags().GetString("psk"); err != nil {
		errs = append(errs, err)
	}
	if a.PSKCloudSite, err = cmd.Flags().GetString("psk-cloud-site"); err != nil {
		errs = append(errs, err)
	}
	if a.PSKOnpremSite, err = cmd.Flags().GetString("psk-onprem-site"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the get args struct.
func (a *NetworkVPNTunnelGetArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the update args struct.
func (a *NetworkVPNTunnelUpdateArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.Name, err = cmd.Flags().GetString("name"); err != nil {
		errs = append(errs, err)
	}
	if a.Tags, err = cmd.Flags().GetStringSlice("tags"); err != nil {
		errs = append(errs, err)
	}
	a.TagsChanged = cmd.Flags().Changed("tags")

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags and positional args into the delete args struct.
func (a *NetworkVPNTunnelDeleteArgs) ParseFromCobraCommand(cmd *cobra.Command, cobraArgs []string) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	if len(cobraArgs) > 0 {
		a.ID = cobraArgs[0]
	}
	if a.DryRun, err = cmd.Flags().GetBool("dry-run"); err != nil {
		errs = append(errs, err)
	}
	if a.SkipConfirm, err = cmd.Flags().GetBool("yes"); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// ParseFromCobraCommand reads Cobra flags into the list args struct.
func (a *NetworkVPNTunnelListArgs) ParseFromCobraCommand(cmd *cobra.Command) error {
	var errs []error
	var err error

	if a.ProjectID, err = GetProjectID(cmd); err != nil {
		errs = append(errs, err)
	}
	a.CallOpts = listOpts(cmd)

	return errors.Join(errs...)
}

// =============================================================================
// Validate methods
// =============================================================================

// Validate checks the create args for correctness.
func (a *NetworkVPNTunnelCreateArgs) Validate() error {
	var errs []error

	if !slices.Contains(validRegions, a.Region) {
		errs = append(errs, fmt.Errorf("--region %q: must be one of %v", a.Region, validRegions))
	}
	if !slices.Contains(validBillingPeriods, a.BillingPeriod) {
		errs = append(errs, fmt.Errorf("--billing-period %q: must be one of %v", a.BillingPeriod, validBillingPeriods))
	}
	if a.SubnetCIDR == "" && a.SubnetName == "" {
		errs = append(errs, errors.New("--subnet-cidr or --subnet-name is required"))
	}
	if string(a.IKEEncryption) != "" && !slices.Contains(vpnIKEEncryptionAlgorithms, string(a.IKEEncryption)) {
		errs = append(errs, fmt.Errorf("--ike-encryption %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.IKEEncryption))
	}
	if string(a.IKEHash) != "" && !slices.Contains(vpnIKEHashAlgorithms, string(a.IKEHash)) {
		errs = append(errs, fmt.Errorf("--ike-hash %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.IKEHash))
	}
	if string(a.IKEDHGroup) != "" && !slices.Contains(vpnIKEDHGroups, string(a.IKEDHGroup)) {
		errs = append(errs, fmt.Errorf("--ike-dh-group %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.IKEDHGroup))
	}
	if string(a.IKEDPDAction) != "" && !slices.Contains(vpnIKEDPDActions, string(a.IKEDPDAction)) {
		errs = append(errs, fmt.Errorf("--ike-dpd-action %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.IKEDPDAction))
	}
	if string(a.ESPEncryption) != "" && !slices.Contains(vpnESPEncryptionAlgorithms, string(a.ESPEncryption)) {
		errs = append(errs, fmt.Errorf("--esp-encryption %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.ESPEncryption))
	}
	if string(a.ESPHash) != "" && !slices.Contains(vpnESPHashAlgorithms, string(a.ESPHash)) {
		errs = append(errs, fmt.Errorf("--esp-hash %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.ESPHash))
	}
	if string(a.ESPPFS) != "" && !slices.Contains(vpnESPPFSGroups, string(a.ESPPFS)) {
		errs = append(errs, fmt.Errorf("--esp-pfs %q is not a valid value; see 'acloud network vpntunnel create --help' or the docs for accepted values", a.ESPPFS))
	}

	return errors.Join(errs...)
}

// Validate checks the get args for correctness.
func (a *NetworkVPNTunnelGetArgs) Validate() error {
	if a.ID == "" {
		return errors.New("VPN tunnel ID is required")
	}
	return nil
}

// Validate checks the update args for correctness.
func (a *NetworkVPNTunnelUpdateArgs) Validate() error {
	var errs []error
	if a.ID == "" {
		errs = append(errs, errors.New("VPN tunnel ID is required"))
	}
	if a.Name == "" && !a.TagsChanged {
		errs = append(errs, errors.New("at least one of --name or --tags must be provided"))
	}
	return errors.Join(errs...)
}

// Validate checks the delete args for correctness.
func (a *NetworkVPNTunnelDeleteArgs) Validate() error {
	if a.ID == "" {
		return errors.New("VPN tunnel ID is required")
	}
	return nil
}

// Validate checks the list args for correctness.
func (a *NetworkVPNTunnelListArgs) Validate() error {
	if a.ProjectID == "" {
		return errors.New("project ID is required")
	}
	return nil
}

// =============================================================================
// Operation functions
// =============================================================================

// NetworkVPNTunnelList lists all VPN tunnels in a project.
func NetworkVPNTunnelList(ctx context.Context, client aruba.Client, args NetworkVPNTunnelListArgs) error {
	list, err := client.FromNetwork().VPNTunnels().List(ctx, aruba.URI("/projects/"+args.ProjectID), args.CallOpts...)
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
			redactVPNTunnelSecrets(vpn)
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
}

// NetworkVPNTunnelGet retrieves and displays a VPN tunnel's details.
func NetworkVPNTunnelGet(ctx context.Context, client aruba.Client, args NetworkVPNTunnelGetArgs) error {
	vpn, err := client.FromNetwork().VPNTunnels().Get(ctx, aruba.VPNTunnelRef(args.ProjectID, args.ID))
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
		if raw.Properties.VPNClientSettingsCommon != nil && raw.Properties.VPNClientSettingsCommon.PeerClientPublicIP != nil {
			fmt.Printf("Peer IP:         %s\n", *raw.Properties.VPNClientSettingsCommon.PeerClientPublicIP)
		}
		if raw.Properties.IPConfigurationsCommon != nil {
			fmt.Println("\nIP Configuration:")
			if raw.Properties.IPConfigurationsCommon.VPC != nil {
				fmt.Printf("  VPC:           %s\n", raw.Properties.IPConfigurationsCommon.VPC.URI)
			}
			if raw.Properties.IPConfigurationsCommon.Subnet != nil {
				fmt.Printf("  Subnet CIDR:   %s\n", raw.Properties.IPConfigurationsCommon.Subnet.CIDR)
				if raw.Properties.IPConfigurationsCommon.Subnet.Name != "" {
					fmt.Printf("  Subnet Name:   %s\n", raw.Properties.IPConfigurationsCommon.Subnet.Name)
				}
			}
			if raw.Properties.IPConfigurationsCommon.PublicIP != nil {
				fmt.Printf("  Public IP:     %s\n", raw.Properties.IPConfigurationsCommon.PublicIP.URI)
			}
		}
		if raw.Properties.BillingPlanCommon != nil && raw.Properties.BillingPlanCommon.BillingPeriod != nil {
			fmt.Printf("\nBilling Period:  %s\n", *raw.Properties.BillingPlanCommon.BillingPeriod)
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
}

// NetworkVPNTunnelCreate creates a VPN tunnel using the provided args and client.
func NetworkVPNTunnelCreate(ctx context.Context, client aruba.Client, args NetworkVPNTunnelCreateArgs) error {
	// Build IP config
	ipConfig := aruba.NewVPNIPConfig().
		WithVPC(aruba.VPCRef(args.ProjectID, args.VPCID)).
		WithElasticIP(aruba.ElasticIPRef(args.ProjectID, args.ElasticIPID)).
		WithSubnet(args.SubnetName, args.SubnetCIDR)

	// Build IKE settings
	ike := aruba.NewVPNIKE().WithLifetimeSeconds(int(args.IKELifetime))
	if string(args.IKEEncryption) != "" {
		ike.WithEncryption(args.IKEEncryption)
	}
	if string(args.IKEHash) != "" {
		ike.WithHash(args.IKEHash)
	}
	if string(args.IKEDHGroup) != "" {
		ike.WithDHGroup(args.IKEDHGroup)
	}
	if string(args.IKEDPDAction) != "" {
		ike.WithDPDAction(args.IKEDPDAction)
	}
	if args.IKEDPDInterval > 0 {
		ike.WithDPDIntervalSeconds(int(args.IKEDPDInterval))
	}
	if args.IKEDPDTimeout > 0 {
		ike.WithDPDTimeoutSeconds(int(args.IKEDPDTimeout))
	}

	// Build ESP settings
	esp := aruba.NewVPNESP().
		WithEncryption(args.ESPEncryption).
		WithLifetimeSeconds(int(args.ESPLifetime))
	if string(args.ESPHash) != "" {
		esp.WithHash(args.ESPHash)
	}
	if string(args.ESPPFS) != "" {
		esp.WithPFS(args.ESPPFS)
	}

	// Build PSK settings
	pskBuilder := aruba.NewVPNPSK()
	if args.PSK != "" {
		pskBuilder.WithKey(args.PSK)
	}
	if args.PSKCloudSite != "" {
		pskBuilder.WithCloudSite(args.PSKCloudSite)
	}
	if args.PSKOnpremSite != "" {
		pskBuilder.WithOnPremSite(args.PSKOnpremSite)
	}

	tunnel := aruba.NewVPNTunnel().
		InProject(aruba.URI("/projects/" + args.ProjectID)).
		Named(args.Name).
		InRegion(args.Region).
		OfType(args.VPNType).
		WithVPNClientProtocol(args.Protocol).
		BilledBy(args.BillingPeriod).
		WithPeerClientPublicIP(args.PeerIP).
		WithIPConfig(ipConfig).
		WithIKESettings(ike).
		WithESPSettings(esp).
		WithPSKSettings(pskBuilder).
		RetaggedAs(args.Tags...)

	resp, err := client.FromNetwork().VPNTunnels().Create(ctx, tunnel)
	if err != nil {
		return fmt.Errorf("creating VPN tunnel: %w", apiErrFromV2(err))
	}

	if resp != nil && resp.Raw() != nil {
		raw := resp.Raw()
		fmt.Printf("\n%s\n", msgCreated("VPN Tunnel", args.Name))
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
		fmt.Println(msgCreatedAsync("VPN Tunnel", args.Name))
	}
	return nil
}

// NetworkVPNTunnelUpdate updates a VPN tunnel's name and/or tags.
func NetworkVPNTunnelUpdate(ctx context.Context, client aruba.Client, args NetworkVPNTunnelUpdateArgs) error {
	vpn, err := client.FromNetwork().VPNTunnels().Get(ctx, aruba.VPNTunnelRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("getting VPN tunnel: %w", apiErrFromV2(err))
	}

	if vpn == nil || vpn.ID() == "" {
		return fmt.Errorf("VPN tunnel not found")
	}

	if vpn.State() == StateInCreation {
		return fmt.Errorf("cannot update VPN tunnel while it is in 'InCreation' state. Please wait until the VPN tunnel is fully created")
	}

	if args.Name != "" {
		vpn.Named(args.Name)
	}
	if args.TagsChanged {
		vpn.RetaggedAs(args.Tags...)
	}
	// sdk-go v1.0.0 fromResponse now rehydrates IKE/ESP/PSK into the wrapper
	// via IKE()/ESP()/PSK() accessors, so Update carries them in the PUT body
	// without manual re-attachment (closes #132 / TD-034).

	updated, err := client.FromNetwork().VPNTunnels().Update(ctx, vpn)
	if err != nil {
		return fmt.Errorf("updating VPN tunnel: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("VPN Tunnel", args.ID))
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
		fmt.Println(msgUpdatedAsync("VPN Tunnel", args.ID))
	}
	return nil
}

// NetworkVPNTunnelDelete deletes a VPN tunnel.
func NetworkVPNTunnelDelete(ctx context.Context, client aruba.Client, args NetworkVPNTunnelDeleteArgs) error {
	err := client.FromNetwork().VPNTunnels().Delete(ctx, aruba.VPNTunnelRef(args.ProjectID, args.ID))
	if err != nil {
		return fmt.Errorf("deleting VPN tunnel: %w", apiErrFromV2(err))
	}
	fmt.Println(msgDeleted("VPN tunnel", args.ID))
	return nil
}

// =============================================================================
// Run wiring functions
// =============================================================================

// NetworkVPNTunnelListRun is the Cobra RunE handler for VPN tunnel list.
func NetworkVPNTunnelListRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkVPNTunnelListArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNTunnelList(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNTunnelGetRun is the Cobra RunE handler for VPN tunnel get.
func NetworkVPNTunnelGetRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNTunnelGetArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNTunnelGet(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNTunnelCreateRun is the Cobra RunE handler for VPN tunnel create.
func NetworkVPNTunnelCreateRun(cmd *cobra.Command, _ []string) error {
	args, err := NewNetworkVPNTunnelCreateArgsFromCobraCommand(cmd)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNTunnelCreate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNTunnelUpdateRun is the Cobra RunE handler for VPN tunnel update.
func NetworkVPNTunnelUpdateRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNTunnelUpdateArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if err := NetworkVPNTunnelUpdate(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}

// NetworkVPNTunnelDeleteRun is the Cobra RunE handler for VPN tunnel delete.
// confirmDelete and --dry-run live here; the operation function is I/O-pure.
func NetworkVPNTunnelDeleteRun(cmd *cobra.Command, cobraArgs []string) error {
	args, err := NewNetworkVPNTunnelDeleteArgsFromCobraCommand(cmd, cobraArgs)
	if err != nil {
		return fmt.Errorf("checking args: %w", err)
	}

	if !args.SkipConfirm {
		ok, err := confirmDelete("VPN tunnel", args.ID)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	ctx, cancel := newCtx()
	defer cancel()

	if args.DryRun {
		if _, err := client.FromNetwork().VPNTunnels().Get(ctx, aruba.VPNTunnelRef(args.ProjectID, args.ID)); err != nil {
			return fmt.Errorf("dry-run: VPN tunnel not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("VPN tunnel", args.ID))
		return nil
	}

	if err := NetworkVPNTunnelDelete(ctx, client, *args); err != nil {
		return fmt.Errorf("running command: %w", err)
	}
	return nil
}
