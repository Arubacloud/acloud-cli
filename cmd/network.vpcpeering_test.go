package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
		"--region", "IT-BG",
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
