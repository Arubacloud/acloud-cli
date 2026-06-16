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
			name: "--output yaml emits valid YAML",
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123", "--output", "yaml"},
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
				if strings.TrimSpace(out) == "{}" || strings.TrimSpace(out) == "" {
					t.Errorf("yaml output is empty or {}, got: %s", out)
				}
			},
		},
		{
			name: "--output table-json emits valid JSON array",
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123", "--output", "table-json"},
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
			args: []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123", "--output", "table-yaml"},
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
				if strings.TrimSpace(out) == "" {
					t.Errorf("table-yaml output is empty, got: %s", out)
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
		"--region", "ITBG-Bergamo",
		"--local-network", "10.0.0.0/24",
		"--remote-network", "192.168.0.0/24",
		"--billing-period", "Hour",
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

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validNetworkVPCPeeringRouteCreateArgs() NetworkVPCPeeringRouteCreateArgs {
	return NetworkVPCPeeringRouteCreateArgs{
		ProjectID:     "proj-123",
		VPCID:         "vpc-001",
		PeeringID:     "peer-001",
		Name:          "my-route",
		Region:        aruba.RegionITBGBergamo,
		LocalNetwork:  "10.0.0.0/24",
		RemoteNetwork: "10.1.0.0/24",
		BillingPeriod: aruba.BillingPeriodHour,
	}
}

func TestNetworkVPCPeeringRouteCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*NetworkVPCPeeringRouteCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *NetworkVPCPeeringRouteCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *NetworkVPCPeeringRouteCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *NetworkVPCPeeringRouteCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *NetworkVPCPeeringRouteCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *NetworkVPCPeeringRouteCreateArgs) { a.BillingPeriod = "Weekly" },
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:        "empty local network",
			mutate:      func(a *NetworkVPCPeeringRouteCreateArgs) { a.LocalNetwork = "" },
			wantErr:     true,
			errContains: "--local-network",
		},
		{
			name:        "empty remote network",
			mutate:      func(a *NetworkVPCPeeringRouteCreateArgs) { a.RemoteNetwork = "" },
			wantErr:     true,
			errContains: "--remote-network",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validNetworkVPCPeeringRouteCreateArgs()
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

func TestNetworkVPCPeeringRouteGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPCPeeringRouteGetArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: "route-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty route ID", func(t *testing.T) {
		args := NetworkVPCPeeringRouteGetArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty route ID")
		}
	})
	t.Run("empty VPC ID", func(t *testing.T) {
		args := NetworkVPCPeeringRouteGetArgs{ProjectID: "p1", VPCID: "", PeeringID: "peer-001", RouteID: "route-001"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty VPC ID")
		}
	})
}

func TestNetworkVPCPeeringRouteUpdateArgs_Validate(t *testing.T) {
	t.Run("happy path with name", func(t *testing.T) {
		args := NetworkVPCPeeringRouteUpdateArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: "route-001", Name: "new-name"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("happy path with tags changed", func(t *testing.T) {
		args := NetworkVPCPeeringRouteUpdateArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: "route-001", TagsChanged: true}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("no name and no tags", func(t *testing.T) {
		args := NetworkVPCPeeringRouteUpdateArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: "route-001"}
		err := args.Validate()
		if err == nil {
			t.Fatal("expected error when no flags provided")
		}
		if !strings.Contains(err.Error(), "at least one") {
			t.Errorf("error %q does not contain 'at least one'", err.Error())
		}
	})
	t.Run("empty route ID", func(t *testing.T) {
		args := NetworkVPCPeeringRouteUpdateArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: "", Name: "x"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty route ID")
		}
	})
}

func TestNetworkVPCPeeringRouteDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPCPeeringRouteDeleteArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: "route-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty route ID", func(t *testing.T) {
		args := NetworkVPCPeeringRouteDeleteArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001", RouteID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty route ID")
		}
	})
}

func TestNetworkVPCPeeringRouteListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPCPeeringRouteListArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: "peer-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty VPC ID", func(t *testing.T) {
		args := NetworkVPCPeeringRouteListArgs{ProjectID: "p1", VPCID: "", PeeringID: "peer-001"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty VPC ID")
		}
	})
	t.Run("empty peering ID", func(t *testing.T) {
		args := NetworkVPCPeeringRouteListArgs{ProjectID: "p1", VPCID: "vpc-001", PeeringID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty peering ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkVPCPeeringRouteCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-new", "my-route"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
		jsonResponse(200, types.VPCPeeringRouteResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))

	out := captureStdout(func() {
		err := NetworkVPCPeeringRouteCreate(context.Background(), srv.Client(), validNetworkVPCPeeringRouteCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "route-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPCPeeringRouteCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
		errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := NetworkVPCPeeringRouteCreate(context.Background(), srv.Client(), validNetworkVPCPeeringRouteCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating VPC peering route") {
		t.Errorf("error %q does not contain 'creating VPC peering route'", err.Error())
	}
}

func TestNetworkVPCPeeringRouteList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "route-001", "my-route"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
		jsonResponse(200, types.VPCPeeringRouteListResponse{
			Values: []types.VPCPeeringRouteResponse{
				{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
			},
		}))

	out := captureStdout(func() {
		err := NetworkVPCPeeringRouteList(context.Background(), srv.Client(), NetworkVPCPeeringRouteListArgs{
			ProjectID: "proj-123",
			VPCID:     "vpc-001",
			PeeringID: "peer-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "route-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPCPeeringRouteList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes",
		jsonResponse(200, types.VPCPeeringRouteListResponse{}))

	out := captureStdout(func() {
		err := NetworkVPCPeeringRouteList(context.Background(), srv.Client(), NetworkVPCPeeringRouteListArgs{
			ProjectID: "proj-123",
			VPCID:     "vpc-001",
			PeeringID: "peer-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No VPC peering routes found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestNetworkVPCPeeringRouteCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"network", "vpcpeeringroute", "create", "vpc-001", "peering-001",
		"--project-id", "proj-123",
		"--name", "x",
		"--local-network", "10.0.0.0/24",
		"--remote-network", "10.1.0.0/24",
		"--region", "ITBG-Bergamo",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkVPCPeeringRouteListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"network", "vpcpeeringroute", "list", "vpc-001", "peering-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}
