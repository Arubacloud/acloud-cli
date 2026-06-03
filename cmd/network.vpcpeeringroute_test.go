package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPCPeeringRouteListCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
					jsonResponse(200, types.VPCPeeringRouteListResponse{
						Values: []types.VPCPeeringRouteResponse{
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
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
					jsonResponse(200, types.VPCPeeringRouteListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
					jsonResponse(200, types.VPCPeeringRouteListResponse{
						Values: []types.VPCPeeringRouteResponse{
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
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
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

func TestVPCPeeringRouteGetCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeeringroute", "get", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					jsonResponse(200, types.VPCPeeringRouteResponse{
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
			args: []string{"network", "vpcpeeringroute", "get", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpcpeeringroute", "get", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
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

func TestVPCPeeringRouteCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
		"--project-id", "proj-123",
		"--name", "my-route",
		"--region", "IT-BG",
		"--local-network", "10.0.0.0/24",
		"--remote-network", "192.168.0.0/24",
		"--billing-period", "monthly",
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
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
					jsonResponse(200, types.VPCPeeringRouteResponse{
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
			name:        "missing required flag --local-network",
			args:        removeFlag(baseArgs, "--local-network", "10.0.0.0/24"),
			wantErr:     true,
			errContains: "local-network",
		},
		{
			name:        "missing required flag --remote-network",
			args:        removeFlag(baseArgs, "--remote-network", "192.168.0.0/24"),
			wantErr:     true,
			errContains: "remote-network",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API 404 propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
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

func TestVPCPeeringRouteUpdateCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeeringroute", "update", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					jsonResponse(200, types.VPCPeeringRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				updID, updName := "route-001", "new-name"
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					jsonResponse(200, types.VPCPeeringRouteResponse{
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
			args:        []string{"network", "vpcpeeringroute", "update", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"},
			setupSrv:    nil,
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "server error on update propagates",
			args: []string{"network", "vpcpeeringroute", "update", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					jsonResponse(200, types.VPCPeeringRouteResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
		},
		{
			name: "API 404 on get propagates",
			args: []string{"network", "vpcpeeringroute", "update", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
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

func TestVPCPeeringRouteDeleteCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeeringroute", "delete", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
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
			args: []string{"network", "vpcpeeringroute", "delete", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "route-001", "my-route"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					jsonResponse(200, types.VPCPeeringRouteResponse{
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
			args: []string{"network", "vpcpeeringroute", "delete", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpcpeeringroute", "delete", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001",
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

func TestVPCPeeringRouteListCmd_WithProperties(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	state := types.StateActive
	period := types.BillingPeriodHour
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes", jsonResponse(200, types.VPCPeeringRouteListResponse{
		Values: []types.VPCPeeringRouteResponse{
			{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				Properties: types.VPCPeeringRoutePropertiesResponse{
					LocalNetworkAddress:  "10.0.0.0/24",
					RemoteNetworkAddress: "10.1.0.0/24",
					BillingPlanCommon:    &types.BillingPlanCommon{BillingPeriod: &period},
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPCPeeringRouteGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001"
	state := types.StateActive
	createdBy := "test-user@example.com"
	period := types.BillingPeriodHour
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes/route-001", jsonResponse(200, types.VPCPeeringRouteResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:           &id,
			Name:         &name,
			URI:          &uri,
			CreationDate: &now,
			CreatedBy:    &createdBy,
			Tags:         []string{"env=test"},
		},
		Properties: types.VPCPeeringRoutePropertiesResponse{
			LocalNetworkAddress:  "10.0.0.0/24",
			RemoteNetworkAddress: "10.1.0.0/24",
			BillingPlanCommon:    &types.BillingPlanCommon{BillingPeriod: &period},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeeringroute", "get", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
