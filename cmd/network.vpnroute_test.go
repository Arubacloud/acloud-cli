package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPNRouteListCmd(t *testing.T) {
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
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteListResponse{
						Values: []types.VPNRouteResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteListResponse{
						Values: []types.VPNRouteResponse{
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
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteListResponse{
						Values: []types.VPNRouteResponse{
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
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteListResponse{
						Values: []types.VPNRouteResponse{
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
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteListResponse{
						Values: []types.VPNRouteResponse{
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
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
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

func TestVPNRouteGetCmd(t *testing.T) {
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
			args: []string{"network", "vpnroute", "get", "vpn-001", "route-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					jsonResponse(200, types.VPNRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpnroute", "get", "vpn-001", "route-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpnroute", "get", "vpn-001", "route-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
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

func TestVPNRouteCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpnroute", "create", "vpn-001",
		"--project-id", "proj-123",
		"--name", "my-route",
		"--region", "ITBG-Bergamo",
		"--cloud-subnet", "10.0.0.0/24",
		"--onprem-subnet", "192.168.1.0/24",
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
				id, name := "route-new", "my-route"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-route"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --cloud-subnet",
			args:        removeFlag(baseArgs, "--cloud-subnet", "10.0.0.0/24"),
			wantErr:     true,
			errContains: "cloud-subnet",
		},
		{
			name:        "missing required flag --onprem-subnet",
			args:        removeFlag(baseArgs, "--onprem-subnet", "192.168.1.0/24"),
			wantErr:     true,
			errContains: "onprem-subnet",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API 404 propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
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

func TestVPNRouteUpdateCmd(t *testing.T) {
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
			args: []string{"network", "vpnroute", "update", "vpn-001", "route-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					jsonResponse(200, types.VPNRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				updID, updName := "route-001", "new-name"
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					jsonResponse(200, types.VPNRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &updID, Name: &updName},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no fields provided returns error",
			args:        []string{"network", "vpnroute", "update", "vpn-001", "route-001", "--project-id", "proj-123"},
			setupSrv:    nil,
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "server error on update propagates",
			args: []string{"network", "vpnroute", "update", "vpn-001", "route-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					jsonResponse(200, types.VPNRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
		},
		{
			name: "API 404 on get propagates",
			args: []string{"network", "vpnroute", "update", "vpn-001", "route-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
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

func TestVPNRouteDeleteCmd(t *testing.T) {
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
			args: []string{"network", "vpnroute", "delete", "vpn-001", "route-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run registers only GET",
			args: []string{"network", "vpnroute", "delete", "vpn-001", "route-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					jsonResponse(200, types.VPNRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpnroute", "delete", "vpn-001", "route-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpnroute", "delete", "vpn-001", "route-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
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

func TestVPNRouteCreateCmd_WithStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes", jsonResponse(200, types.VPNRouteResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"network", "vpnroute", "create", "vpn-001",
		"--project-id", "proj-123",
		"--name", "my-route",
		"--region", "ITBG-Bergamo",
		"--cloud-subnet", "10.0.0.0/24",
		"--onprem-subnet", "192.168.1.0/24",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVPNRouteListCmd_WithStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes", jsonResponse(200, types.VPNRouteListResponse{
		Values: []types.VPNRouteResponse{
			{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				Status:   types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPNRouteGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	uri := "/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001"
	region := types.Region("ITBG-Bergamo")
	state := types.StateActive
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001", jsonResponse(200, types.VPNRouteResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.VPNRoutePropertiesResponse{
			OnPremSubnet: "192.168.1.0/24",
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpnroute", "get", "vpn-001", "route-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validNetworkVPNRouteCreateArgs() NetworkVPNRouteCreateArgs {
	return NetworkVPNRouteCreateArgs{
		ProjectID:    "proj-123",
		TunnelID:     "vpn-001",
		Name:         "my-route",
		Region:       aruba.RegionITBGBergamo,
		LocalSubnet:  "10.0.0.0/24",
		RemoteSubnet: "192.168.1.0/24",
	}
}

func TestNetworkVPNRouteCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*NetworkVPNRouteCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *NetworkVPNRouteCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *NetworkVPNRouteCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *NetworkVPNRouteCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *NetworkVPNRouteCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "empty tunnel ID",
			mutate:      func(a *NetworkVPNRouteCreateArgs) { a.TunnelID = "" },
			wantErr:     true,
			errContains: "VPN tunnel ID",
		},
		{
			name:        "empty local subnet",
			mutate:      func(a *NetworkVPNRouteCreateArgs) { a.LocalSubnet = "" },
			wantErr:     true,
			errContains: "--cloud-subnet",
		},
		{
			name:        "empty remote subnet",
			mutate:      func(a *NetworkVPNRouteCreateArgs) { a.RemoteSubnet = "" },
			wantErr:     true,
			errContains: "--onprem-subnet",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validNetworkVPNRouteCreateArgs()
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

func TestNetworkVPNRouteGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPNRouteGetArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: "route-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty route ID", func(t *testing.T) {
		args := NetworkVPNRouteGetArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty route ID")
		}
	})
	t.Run("empty tunnel ID", func(t *testing.T) {
		args := NetworkVPNRouteGetArgs{ProjectID: "p1", TunnelID: "", RouteID: "route-001"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty tunnel ID")
		}
	})
}

func TestNetworkVPNRouteUpdateArgs_Validate(t *testing.T) {
	t.Run("happy path with name", func(t *testing.T) {
		args := NetworkVPNRouteUpdateArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: "route-001", Name: "new-name"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("happy path with tags changed", func(t *testing.T) {
		args := NetworkVPNRouteUpdateArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: "route-001", TagsChanged: true}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("no name and no tags", func(t *testing.T) {
		args := NetworkVPNRouteUpdateArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: "route-001"}
		err := args.Validate()
		if err == nil {
			t.Fatal("expected error when no flags provided")
		}
		if !strings.Contains(err.Error(), "at least one") {
			t.Errorf("error %q does not contain 'at least one'", err.Error())
		}
	})
	t.Run("empty route ID", func(t *testing.T) {
		args := NetworkVPNRouteUpdateArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: "", Name: "x"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty route ID")
		}
	})
}

func TestNetworkVPNRouteDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPNRouteDeleteArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: "route-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty route ID", func(t *testing.T) {
		args := NetworkVPNRouteDeleteArgs{ProjectID: "p1", TunnelID: "vpn-001", RouteID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty route ID")
		}
	})
}

func TestNetworkVPNRouteListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPNRouteListArgs{ProjectID: "p1", TunnelID: "vpn-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty tunnel ID", func(t *testing.T) {
		args := NetworkVPNRouteListArgs{ProjectID: "p1", TunnelID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty tunnel ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkVPNRouteCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-new", "my-route"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
		jsonResponse(200, types.VPNRouteResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))

	out := captureStdout(func() {
		err := NetworkVPNRouteCreate(context.Background(), srv.Client(), validNetworkVPNRouteCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "route-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPNRouteCreate_CloudSubnetInOutput(t *testing.T) {
	// Verifies that CloudSubnet() from the response is printed correctly — the
	// previous bug (TD-030) caused it to always be blank because the SDK did not
	// parse the {"cidr":"..."} JSON shape returned by the API.
	srv := newArubaTestServer(t)
	id, name := "route-new", "my-route"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
		jsonResponse(200, types.VPNRouteResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
			Properties: types.VPNRoutePropertiesResponse{
				CloudSubnet:  types.SubnetCIDROrRef{CIDR: "10.0.0.0/24"},
				OnPremSubnet: "192.168.1.0/24",
			},
		}))

	out := captureStdout(func() {
		err := NetworkVPNRouteCreate(context.Background(), srv.Client(), validNetworkVPNRouteCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "10.0.0.0/24") {
		t.Errorf("CloudSubnet CIDR missing from create output; got: %s", out)
	}
	if !strings.Contains(out, "192.168.1.0/24") {
		t.Errorf("OnPremSubnet CIDR missing from create output; got: %s", out)
	}
}

func TestNetworkVPNRouteGet_CloudSubnetInOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes/route-001",
		jsonResponse(200, types.VPNRouteResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
			Properties: types.VPNRoutePropertiesResponse{
				CloudSubnet:  types.SubnetCIDROrRef{CIDR: "10.0.0.0/24"},
				OnPremSubnet: "192.168.1.0/24",
			},
		}))

	out := captureStdout(func() {
		err := NetworkVPNRouteGet(context.Background(), srv.Client(), NetworkVPNRouteGetArgs{
			ProjectID: "proj-123",
			TunnelID:  "vpn-001",
			RouteID:   "route-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "10.0.0.0/24") {
		t.Errorf("CloudSubnet CIDR missing from get output; got: %s", out)
	}
}

func TestNetworkVPNRouteList_CloudSubnetInOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
		jsonResponse(200, types.VPNRouteListResponse{
			Values: []types.VPNRouteResponse{
				{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Properties: types.VPNRoutePropertiesResponse{
						CloudSubnet:  types.SubnetCIDROrRef{CIDR: "10.0.0.0/24"},
						OnPremSubnet: "192.168.1.0/24",
					},
				},
			},
		}))

	out := captureStdout(func() {
		err := NetworkVPNRouteList(context.Background(), srv.Client(), NetworkVPNRouteListArgs{
			ProjectID: "proj-123",
			TunnelID:  "vpn-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "10.0.0.0/24") {
		t.Errorf("CloudSubnet CIDR missing from list output; got: %s", out)
	}
}

func TestNetworkVPNRouteCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
		errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := NetworkVPNRouteCreate(context.Background(), srv.Client(), validNetworkVPNRouteCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating VPN route") {
		t.Errorf("error %q does not contain 'creating VPN route'", err.Error())
	}
}

func TestNetworkVPNRouteList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
		jsonResponse(200, types.VPNRouteListResponse{
			Values: []types.VPNRouteResponse{
				{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
			},
		}))

	out := captureStdout(func() {
		err := NetworkVPNRouteList(context.Background(), srv.Client(), NetworkVPNRouteListArgs{
			ProjectID: "proj-123",
			TunnelID:  "vpn-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPNRouteList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
		jsonResponse(200, types.VPNRouteListResponse{}))

	out := captureStdout(func() {
		err := NetworkVPNRouteList(context.Background(), srv.Client(), NetworkVPNRouteListArgs{
			ProjectID: "proj-123",
			TunnelID:  "vpn-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No VPN routes found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestNetworkVPNRouteCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"network", "vpnroute", "create", "tunnel-001",
		"--project-id", "proj-123",
		"--name", "x",
		"--region", "ITBG-Bergamo",
		"--cloud-subnet", "10.0.0.0/24",
		"--onprem-subnet", "10.1.0.0/24",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkVPNRouteListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"network", "vpnroute", "list", "tunnel-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}
