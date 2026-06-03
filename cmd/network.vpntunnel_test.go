package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPNTunnelListCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-001", "my-tunnel"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					jsonResponse(200, types.VPNTunnelListResponse{
						Values: []types.VPNTunnelResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpn-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					jsonResponse(200, types.VPNTunnelListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-001", "my-tunnel"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					jsonResponse(200, types.VPNTunnelListResponse{
						Values: []types.VPNTunnelResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNTunnelListCmd_RedactsPSKSecret(t *testing.T) {
	const secret = "super-secret-psk-value"
	id, name := "vpn-001", "my-tunnel"
	s := secret

	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			srv := newArubaTestServer(t)
			srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
				jsonResponse(200, types.VPNTunnelListResponse{
					Values: []types.VPNTunnelResponse{{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Properties: types.VPNTunnelPropertiesResponse{
							VPNClientSettingsCommon: &types.VPNClientSettingsCommon{
								PSK: &types.PSKSettingsCommon{Secret: &s},
							},
						},
					}},
				}))
			out := captureStdout(func() {
				rootCmd.PersistentFlags().Set("output", format)
				_ = runCmd(srv.Client(), []string{"network", "vpntunnel", "list", "--project-id", "proj-123"})
				rootCmd.PersistentFlags().Set("output", "table")
			})
			if strings.Contains(out, secret) {
				t.Fatalf("PSK secret leaked to --output %s:\n%s", format, out)
			}
		})
	}
}

func TestVPNTunnelGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "vpntunnel", "get", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-001", "my-tunnel"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpn-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpntunnel", "get", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpntunnel", "get", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNTunnelCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpntunnel", "create",
		"--project-id", "proj-123",
		"--name", "my-tunnel",
		"--region", "IT-BG",
		"--peer-ip", "1.2.3.4",
		"--vpc-id", "vpc-001",
		"--elastic-ip-id", "eip-001",
		"--subnet-cidr", "10.0.1.0/24",
	}
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-new", "my-tunnel"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpn-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-tunnel"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --peer-ip",
			args:        removeFlag(baseArgs, "--peer-ip", "1.2.3.4"),
			wantErr:     true,
			errContains: "peer-ip",
		},
		{
			name:        "missing required flag --vpc-id",
			args:        removeFlag(baseArgs, "--vpc-id", "vpc-001"),
			wantErr:     true,
			errContains: "vpc-id",
		},
		{
			name:        "rejects invalid --ike-encryption before API call",
			args:        append(append([]string(nil), baseArgs...), "--ike-encryption", "rot13"),
			wantErr:     true,
			errContains: `"rot13" is not a valid value`,
		},
		{
			name:        "rejects invalid --ike-dpd-action before API call",
			args:        append(append([]string(nil), baseArgs...), "--ike-dpd-action", "explode"),
			wantErr:     true,
			errContains: "is not a valid value",
		},
		{
			name:        "rejects invalid --esp-pfs before API call",
			args:        append(append([]string(nil), baseArgs...), "--esp-pfs", "group99"),
			wantErr:     true,
			errContains: "is not a valid value",
		},
		{
			name: "accepts valid --ike-encryption without API error",
			args: append(append([]string(nil), baseArgs...), "--ike-encryption", "aes256"),
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-new", "my-tunnel"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API 404 propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNTunnelUpdateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "vpntunnel", "update", "vpn-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-001", "my-tunnel"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				updID, updName := "vpn-001", "new-name"
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &updID, Name: &updName},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpn-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no fields provided returns error",
			args:        []string{"network", "vpntunnel", "update", "vpn-001", "--project-id", "proj-123"},
			setupSrv:    nil,
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "server error on update propagates",
			args: []string{"network", "vpntunnel", "update", "vpn-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-001", "my-tunnel"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
		},
		{
			name: "API 404 on get propagates",
			args: []string{"network", "vpntunnel", "update", "vpn-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNTunnelDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			args: []string{"network", "vpntunnel", "delete", "vpn-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpn-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run registers only GET",
			args: []string{"network", "vpntunnel", "delete", "vpn-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpn-001", "my-tunnel"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					jsonResponse(200, types.VPNTunnelResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpn-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpntunnel", "delete", "vpn-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpntunnel", "delete", "vpn-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNValidateEnum(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		flag    string
		valid   []string
		wantErr bool
		errMsg  string
	}{
		{name: "empty value passes", value: "", flag: "ike-encryption", valid: vpnEncryptionAlgorithms},
		{name: "valid IKE encryption", value: "aes256", flag: "ike-encryption", valid: vpnEncryptionAlgorithms},
		{name: "valid IKE hash", value: "sha256", flag: "ike-hash", valid: vpnHashAlgorithms},
		{name: "valid IKE dh-group", value: "14", flag: "ike-dh-group", valid: vpnDHGroups},
		{name: "valid IKE dpd-action", value: "trap", flag: "ike-dpd-action", valid: vpnDPDActions},
		{name: "valid ESP pfs", value: "enable", flag: "esp-pfs", valid: vpnPFSGroups},
		{name: "valid ESP encryption", value: "aes256", flag: "esp-encryption", valid: vpnEncryptionAlgorithms},
		{name: "valid ESP hash", value: "sha256", flag: "esp-hash", valid: vpnHashAlgorithms},
		{
			name: "invalid IKE encryption rejected", value: "rot13", flag: "ike-encryption", valid: vpnEncryptionAlgorithms,
			wantErr: true, errMsg: `--ike-encryption "rot13" is not a valid value`,
		},
		{
			name: "invalid IKE hash rejected", value: "crc32", flag: "ike-hash", valid: vpnHashAlgorithms,
			wantErr: true, errMsg: "--ike-hash",
		},
		{
			name: "invalid dpd-action rejected", value: "explode", flag: "ike-dpd-action", valid: vpnDPDActions,
			wantErr: true, errMsg: "--ike-dpd-action",
		},
		{
			name: "invalid esp-pfs rejected", value: "group99", flag: "esp-pfs", valid: vpnPFSGroups,
			wantErr: true, errMsg: "--esp-pfs",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := vpnValidateEnum(tc.value, tc.flag, tc.valid)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestVPNTunnelListCmd_WithAllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-001", "my-tunnel"
	region := types.Region("IT-BG")
	vpnType := types.VPNTypeSiteToSite
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelListResponse{
		Values: []types.VPNTunnelResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.VPNTunnelPropertiesResponse{
					VPNType: &vpnType,
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpntunnel", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpn-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPNTunnelGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-001", "my-tunnel"
	uri := "/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001"
	region := types.Region("IT-BG")
	vpnType := types.VPNTypeSiteToSite
	state := types.StateActive
	peerIP := "203.0.113.1"
	protocol := types.VPNClientProtocol("ikev2")
	period := types.BillingPeriodHour
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.VPNTunnelPropertiesResponse{
			VPNType:           &vpnType,
			VPNClientProtocol: &protocol,
			VPNClientSettingsCommon: &types.VPNClientSettingsCommon{
				PeerClientPublicIP: &peerIP,
			},
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &period},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpntunnel", "get", "vpn-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpn-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPNTunnelGetCmd_WithIPConfig(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-001", "my-tunnel"
	vpcURI := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"
	pubIPURI := "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"
	vpnType := types.VPNTypeSiteToSite
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.VPNTunnelPropertiesResponse{
			VPNType: &vpnType,
			IPConfigurationsCommon: &types.IPConfigurationsCommon{
				VPC:      &types.ReferenceResourceCommon{URI: vpcURI},
				Subnet:   &types.SubnetInfoCommon{CIDR: "10.0.0.0/24", Name: "my-subnet"},
				PublicIP: &types.ReferenceResourceCommon{URI: pubIPURI},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpntunnel", "get", "vpn-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpn-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPNTunnelCreateCmd_WithTypeAndProtocol(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-new", "my-tunnel"
	uri := "/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-new"
	vpnType := types.VPNTypeSiteToSite
	protocol := types.VPNClientProtocolIKEv2
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			Name: &name,
			URI:  &uri,
			Tags: []string{"env=test"},
		},
		Properties: types.VPNTunnelPropertiesResponse{
			VPNType:           &vpnType,
			VPNClientProtocol: &protocol,
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"network", "vpntunnel", "create",
		"--project-id", "proj-123",
		"--name", "my-tunnel",
		"--region", "IT-BG",
		"--vpn-type", "Site-To-Site",
		"--protocol", "ikev2",
		"--billing-period", "Hour",
		"--peer-ip", "203.0.113.1",
		"--vpc-id", "vpc-001",
		"--elastic-ip-id", "eip-001",
		"--subnet-cidr", "10.0.1.0/24",
		"--ike-encryption", "aes256",
		"--ike-lifetime", "28800",
		"--esp-encryption", "aes256",
		"--esp-lifetime", "3600",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpn-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPNTunnelUpdateCmd_WithTagsOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-001", "my-tunnel"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	updID, updName := "vpn-001", "my-tunnel"
	srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &updID,
			Name: &updName,
			Tags: []string{"env=prod"},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"network", "vpntunnel", "update", "vpn-001",
		"--project-id", "proj-123",
		"--tags", "env=prod",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpn-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
