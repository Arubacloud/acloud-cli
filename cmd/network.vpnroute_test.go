package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
					jsonResponse(200, types.VPNRouteList{
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
					jsonResponse(200, types.VPNRouteList{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "vpnroute", "list", "vpn-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes",
					jsonResponse(200, types.VPNRouteList{
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
		"--region", "IT-BG",
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
						Status:   types.ResourceStatus{State: func() *types.State { s := types.StateActive; return &s }()},
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
						Status:   types.ResourceStatus{State: func() *types.State { s := types.StateActive; return &s }()},
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
		Status:   types.ResourceStatus{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"network", "vpnroute", "create", "vpn-001",
		"--project-id", "proj-123",
		"--name", "my-route",
		"--region", "IT-BG",
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
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes", jsonResponse(200, types.VPNRouteList{
		Values: []types.VPNRouteResponse{
			{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				Status:   types.ResourceStatus{State: &state},
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
	region := types.Region("IT-BG")
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
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpnroute", "get", "vpn-001", "route-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
