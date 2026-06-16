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

func TestVPCPeeringListCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringListResponse{
						Values: []types.VPCPeeringResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringListResponse{
						Values: []types.VPCPeeringResponse{
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
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringListResponse{
						Values: []types.VPCPeeringResponse{
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
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringListResponse{
						Values: []types.VPCPeeringResponse{
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
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringListResponse{
						Values: []types.VPCPeeringResponse{
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
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
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

func TestVPCPeeringGetCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					jsonResponse(200, types.VPCPeeringResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
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

func TestVPCPeeringCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpcpeering", "create", "vpc-001",
		"--project-id", "proj-123",
		"--name", "my-peering",
		"--region", "ITBG-Bergamo",
		"--peer-vpc-id", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-002",
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
				id, name := "peer-new", "my-peering"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					jsonResponse(200, types.VPCPeeringResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-peering"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --peer-vpc-id",
			args:        removeFlag(baseArgs, "--peer-vpc-id", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-002"),
			wantErr:     true,
			errContains: "peer-vpc-id",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API 404 propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
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

func TestVPCPeeringUpdateCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeering", "update", "vpc-001", "peer-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					jsonResponse(200, types.VPCPeeringResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				updID, updName := "peer-001", "new-name"
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					jsonResponse(200, types.VPCPeeringResponse{
						Metadata: types.ResourceMetadataResponse{ID: &updID, Name: &updName},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no fields provided returns error",
			args:        []string{"network", "vpcpeering", "update", "vpc-001", "peer-001", "--project-id", "proj-123"},
			setupSrv:    nil,
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "server error on update propagates",
			args: []string{"network", "vpcpeering", "update", "vpc-001", "peer-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					jsonResponse(200, types.VPCPeeringResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
		},
		{
			name: "API 404 on get propagates",
			args: []string{"network", "vpcpeering", "update", "vpc-001", "peer-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
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

func TestVPCPeeringDeleteCmd(t *testing.T) {
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
			args: []string{"network", "vpcpeering", "delete", "vpc-001", "peer-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run registers only GET",
			args: []string{"network", "vpcpeering", "delete", "vpc-001", "peer-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "peer-001", "my-peering"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					jsonResponse(200, types.VPCPeeringResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpcpeering", "delete", "vpc-001", "peer-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "vpcpeering", "delete", "vpc-001", "peer-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
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

// TestVPCPeeringGetCmd_FullDetail exercises all nil-guard branches in GET detail output.
func TestVPCPeeringGetCmd_FullDetail(t *testing.T) {
	t.Run("detail with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "peer-001", "my-peering"
		createdBy := "user@example.com"
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		state := types.StateActive
		region := types.Region("IT-BG")
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
			jsonResponse(200, types.VPCPeeringResponse{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
					CreationDate:     &ts,
					CreatedBy:        &createdBy,
					Tags:             []string{"env=prod"},
				},
				Status: types.ResourceStatusResponse{State: &state},
				Properties: types.VPCPeeringPropertiesResponse{
					RemoteVPC: &types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-002"},
				},
			}))
		out, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "peer-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
		if !strings.Contains(out, "IT-BG") {
			t.Errorf("expected region in output, got: %s", out)
		}
		if !strings.Contains(out, "env=prod") {
			t.Errorf("expected tags in output, got: %s", out)
		}
	})

	t.Run("--output json succeeds without error", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "peer-001", "my-peering"
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
			jsonResponse(200, types.VPCPeeringResponse{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
			}))
		_, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestVPCPeeringListCmd_AllOptionalFields exercises nil-guard branches in LIST output.
func TestVPCPeeringListCmd_AllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "peer-001", "my-peering"
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings",
		jsonResponse(200, types.VPCPeeringListResponse{
			Values: []types.VPCPeeringResponse{
				{
					Metadata: types.ResourceMetadataResponse{
						ID:               &id,
						Name:             &name,
						LocationResponse: &types.LocationResponse{Value: region},
					},
					Status: types.ResourceStatusResponse{State: &state},
					Properties: types.VPCPeeringPropertiesResponse{
						RemoteVPC: &types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-002"},
					},
				},
			},
		}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "peer-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "IT-BG") {
		t.Errorf("expected region in output, got: %s", out)
	}
}

// TestVPCPeeringGetCmd_JsonOutput exercises the JSON output path (early return).
func TestVPCPeeringGetCmd_JsonOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "peer-001", "my-peering"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001",
		jsonResponse(200, types.VPCPeeringResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = out // vpcpeering GET uses direct fmt.Printf, not JSON serialization
}

func TestVPCPeeringListCmd_WithAllFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "peer-001", "my-peering"
	region := types.Region("IT-BG")
	state := types.StateActive
	peerVPCURI := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-002"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings", jsonResponse(200, types.VPCPeeringListResponse{
		Values: []types.VPCPeeringResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.VPCPeeringPropertiesResponse{
					RemoteVPC: &types.ReferenceResourceCommon{URI: peerVPCURI},
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "peer-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validNetworkVPCPeeringCreateArgs() NetworkVPCPeeringCreateArgs {
	return NetworkVPCPeeringCreateArgs{
		ProjectID: "proj-123",
		VPCID:     "vpc-001",
		Name:      "my-peering",
		Region:    aruba.RegionITBGBergamo,
		PeerVPCID: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-002",
	}
}

func TestNetworkVPCPeeringCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*NetworkVPCPeeringCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "invalid region",
			mutate:      func(a *NetworkVPCPeeringCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validNetworkVPCPeeringCreateArgs()
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

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkVPCPeeringCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "peer-new", "my-peering"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings", jsonResponse(200, types.VPCPeeringResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := NetworkVPCPeeringCreate(context.Background(), srv.Client(), validNetworkVPCPeeringCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "peer-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPCPeeringCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := NetworkVPCPeeringCreate(context.Background(), srv.Client(), validNetworkVPCPeeringCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating VPC peering") {
		t.Errorf("error %q does not contain 'creating VPC peering'", err.Error())
	}
}

func TestNetworkVPCPeeringCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"network", "vpcpeering", "create", "vpc-001",
		"--project-id", "proj-123",
		"--name", "x",
		"--peer-vpc-id", "peer-vpc-001",
		"--region", "ITBG-Bergamo",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkVPCPeeringListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"network", "vpcpeering", "list", "vpc-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}
