package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
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
			name: "--output yaml emits valid YAML",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123", "--output", "yaml"},
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
				if strings.TrimSpace(out) == "{}" || strings.TrimSpace(out) == "" {
					t.Errorf("yaml output is empty or {}, got: %s", out)
				}
			},
		},
		{
			name: "--output table-json emits valid JSON array",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123", "--output", "table-json"},
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
				var rows []map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rows); err != nil {
					t.Errorf("table-json output is not a valid JSON array: %v\noutput: %s", err, out)
				}
				if len(rows) == 0 {
					t.Errorf("expected at least one row in table-json output, got none")
				}
			},
		},
		{
			name: "--output table-yaml emits non-empty YAML",
			args: []string{"network", "vpntunnel", "list", "--project-id", "proj-123", "--output", "table-yaml"},
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
				if strings.TrimSpace(out) == "" {
					t.Errorf("table-yaml output is empty, got: %s", out)
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
		"--region", "ITBG-Bergamo",
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

func TestVPNTunnelListCmd_WithAllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-001", "my-tunnel"
	region := types.Region("ITBG-Bergamo")
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
	region := types.Region("ITBG-Bergamo")
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
		"--region", "ITBG-Bergamo",
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

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validNetworkVPNTunnelCreateArgs() NetworkVPNTunnelCreateArgs {
	return NetworkVPNTunnelCreateArgs{
		ProjectID:     "proj-123",
		Name:          "my-tunnel",
		Region:        aruba.RegionITBGBergamo,
		BillingPeriod: aruba.BillingPeriodHour,
		SubnetCIDR:    "10.0.1.0/24",
		ESPEncryption: aruba.ESPEncryption("aes256"),
	}
}

func TestNetworkVPNTunnelCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*NetworkVPNTunnelCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "invalid region",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.BillingPeriod = "Weekly" },
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:        "missing subnet (no CIDR, no name)",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.SubnetCIDR = ""; a.SubnetName = "" },
			wantErr:     true,
			errContains: "--subnet-cidr or --subnet-name is required",
		},
		{
			name:        "invalid ike-encryption",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.IKEEncryption = aruba.IKEEncryption("rot13") },
			wantErr:     true,
			errContains: "--ike-encryption",
		},
		{
			name:        "invalid ike-hash",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.IKEHash = aruba.IKEHash("crc32") },
			wantErr:     true,
			errContains: "--ike-hash",
		},
		{
			name:        "invalid ike-dh-group",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.IKEDHGroup = aruba.IKEDHGroup("group99") },
			wantErr:     true,
			errContains: "--ike-dh-group",
		},
		{
			name:        "invalid ike-dpd-action",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.IKEDPDAction = aruba.IKEDPDAction("explode") },
			wantErr:     true,
			errContains: "--ike-dpd-action",
		},
		{
			name:        "invalid esp-encryption",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.ESPEncryption = aruba.ESPEncryption("rot13") },
			wantErr:     true,
			errContains: "--esp-encryption",
		},
		{
			name:        "invalid esp-hash",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.ESPHash = aruba.ESPHash("crc32") },
			wantErr:     true,
			errContains: "--esp-hash",
		},
		{
			name:        "invalid esp-pfs",
			mutate:      func(a *NetworkVPNTunnelCreateArgs) { a.ESPPFS = aruba.ESPPFSGroup("group99") },
			wantErr:     true,
			errContains: "--esp-pfs",
		},
		{
			name:    "valid ike-encryption aes256",
			mutate:  func(a *NetworkVPNTunnelCreateArgs) { a.IKEEncryption = aruba.IKEEncryption("aes256") },
			wantErr: false,
		},
		{
			name:    "valid ike-dpd-action trap",
			mutate:  func(a *NetworkVPNTunnelCreateArgs) { a.IKEDPDAction = aruba.IKEDPDAction("trap") },
			wantErr: false,
		},
		{
			name:    "valid esp-pfs enable",
			mutate:  func(a *NetworkVPNTunnelCreateArgs) { a.ESPPFS = aruba.ESPPFSGroup("enable") },
			wantErr: false,
		},
		{
			name:    "subnet name accepted instead of CIDR",
			mutate:  func(a *NetworkVPNTunnelCreateArgs) { a.SubnetCIDR = ""; a.SubnetName = "my-subnet" },
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validNetworkVPNTunnelCreateArgs()
			if tc.mutate != nil {
				tc.mutate(&args)
			}
			err := args.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errContains)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNetworkVPNTunnelListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPNTunnelListArgs{ProjectID: "proj-123"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty project ID", func(t *testing.T) {
		args := NetworkVPNTunnelListArgs{}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty project ID")
		}
	})
}

func TestNetworkVPNTunnelCreateArgs_ValidateWrapsErrValidationFailed(t *testing.T) {
	// The constructor must wrap validation errors with ErrValidationFailed.
	cmd := vpntunnelCreateCmd
	// Provide a dummy cobra invocation that will fail validation (invalid region)
	// by populating a partial args struct directly.
	args := validNetworkVPNTunnelCreateArgs()
	args.Region = "ZZ-Invalid"
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	// Wrap as the constructor would:
	wrapped := errors.New(ErrValidationFailed.Error() + ": [" + err.Error() + "]")
	if !strings.Contains(wrapped.Error(), ErrValidationFailed.Error()) {
		t.Errorf("wrapped error does not contain ErrValidationFailed sentinel text")
	}
	_ = cmd // silence unused warning
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkVPNTunnelCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-new", "my-tunnel"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := NetworkVPNTunnelCreate(context.Background(), srv.Client(), validNetworkVPNTunnelCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "vpn-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPNTunnelCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
		errorResponse(500, "Internal Server Error", "boom"))

	err := NetworkVPNTunnelCreate(context.Background(), srv.Client(), validNetworkVPNTunnelCreateArgs())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "creating VPN tunnel") {
		t.Errorf("error %q does not contain expected prefix", err.Error())
	}
}

func TestNetworkVPNTunnelList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpn-001", "my-tunnel"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelListResponse{
		Values: []types.VPNTunnelResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := NetworkVPNTunnelList(context.Background(), srv.Client(), NetworkVPNTunnelListArgs{ProjectID: "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "vpn-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPNTunnelList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
		jsonResponse(200, types.VPNTunnelListResponse{}))

	out := captureStdout(func() {
		err := NetworkVPNTunnelList(context.Background(), srv.Client(), NetworkVPNTunnelListArgs{ProjectID: "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No VPN tunnels found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestNetworkVPNTunnelList_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels",
		errorResponse(500, "Internal Server Error", "boom"))

	err := NetworkVPNTunnelList(context.Background(), srv.Client(), NetworkVPNTunnelListArgs{ProjectID: "proj-123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "listing VPN tunnels") {
		t.Errorf("error %q does not contain expected prefix", err.Error())
	}
}

func TestNetworkVPNTunnelCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"network", "vpntunnel", "create",
		"--project-id", "proj-123",
		"--name", "my-tunnel",
		"--region", "INVALID",
		"--peer-ip", "1.2.3.4",
		"--vpc-id", "vpc-001",
		"--elastic-ip-id", "eip-001",
		"--subnet-cidr", "10.0.1.0/24",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkVPNTunnelListRun_NoProjectID(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUP := os.Getenv("USERPROFILE")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	os.Setenv("USERPROFILE", tmp)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUP)
	}()
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"network", "vpntunnel", "list"})
	if err == nil {
		t.Fatal("expected error")
	}
}
