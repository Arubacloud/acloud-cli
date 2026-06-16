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

func TestVPCListCmd(t *testing.T) {
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
			args: []string{"network", "vpc", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
					Values: []types.VPCResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "vpc", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "vpc", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
					Values: []types.VPCResponse{
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
			args: []string{"network", "vpc", "list", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
					Values: []types.VPCResponse{
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
			args: []string{"network", "vpc", "list", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
					Values: []types.VPCResponse{
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
			args: []string{"network", "vpc", "list", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
					Values: []types.VPCResponse{
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
			args: []string{"network", "vpc", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"network", "vpc", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", errorResponse(404, "Not Found", "resource not found"))
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

func TestVPCGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, types.VPCResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"network", "vpc", "get", "vpc-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCCreateCmd(t *testing.T) {
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
			args: []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "new-vpc"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "vpc", "create", "--project-id", "proj-123", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", errorResponse(404, "Not Found", "resource not found"))
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

func TestVPCUpdateCmd(t *testing.T) {
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
			args: []string{"network", "vpc", "update", "vpc-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				state := types.StateActive
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, types.VPCResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, LocationResponse: &types.LocationResponse{Value: region}},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, types.VPCResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"network", "vpc", "update", "vpc-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-Get server error",
			args: []string{"network", "vpc", "update", "vpc-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update server error",
			args: []string{"network", "vpc", "update", "vpc-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				state := types.StateActive
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, types.VPCResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, LocationResponse: &types.LocationResponse{Value: region}},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
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

func TestVPCDeleteCmd(t *testing.T) {
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
			args: []string{"network", "vpc", "delete", "vpc-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"network", "vpc", "delete", "vpc-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vpc-001", "my-vpc"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, types.VPCResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "vpc", "delete", "vpc-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"network", "vpc", "delete", "vpc-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestVPCListCmd_AllOptionalFields(t *testing.T) {
	// Covers: LocationResponse and Status.State nil-guards in list loop.
	srv := newArubaTestServer(t)
	id, name := "vpc-001", "my-vpc"
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpc", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpc-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "IT-BG") {
		t.Errorf("expected region in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestVPCGetCmd_AllOptionalFields(t *testing.T) {
	// Covers: URI, LocationResponse, CreationDate, CreatedBy, Tags, Status.State detail block.
	id, name := "vpc-001", "my-vpc"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"
	createdBy := "user@example.com"
	state := types.StateActive
	region := types.Region("IT-BG")
	now := time.Now()
	makeResponse := func() types.VPCResponse {
		return types.VPCResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:               &id,
				Name:             &name,
				URI:              &uri,
				LocationResponse: &types.LocationResponse{Value: region},
				CreationDate:     &now,
				CreatedBy:        &createdBy,
				Tags:             []string{"env=test"},
			},
			Status: types.ResourceStatusResponse{State: &state},
		}
	}

	t.Run("detail with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"network", "vpc", "get", "vpc-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "vpc-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "IT-BG") {
			t.Errorf("expected region in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
		if !strings.Contains(out, "env=test") {
			t.Errorf("expected tags in output, got: %s", out)
		}
		if !strings.Contains(out, "Active") {
			t.Errorf("expected status in output, got: %s", out)
		}
	})

	t.Run("output json flag succeeds", func(t *testing.T) {
		// vpc get renders text directly (no PrintOutput JSON path);
		// verify command still succeeds with --output json flag present.
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"network", "vpc", "get", "vpc-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "vpc-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})
}

func TestVPCListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-001", "my-vpc"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpc", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpc-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestVPCGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-001", "my-vpc"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "vpc", "get", "vpc-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vpc-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validNetworkVPCCreateArgs() NetworkVPCCreateArgs {
	return NetworkVPCCreateArgs{
		ProjectID: "proj-123",
		Name:      "my-vpc",
		Region:    aruba.RegionITBGBergamo,
	}
}

func TestNetworkVPCCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*NetworkVPCCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *NetworkVPCCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *NetworkVPCCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *NetworkVPCCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *NetworkVPCCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validNetworkVPCCreateArgs()
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

func TestNetworkVPCGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPCGetArgs{ProjectID: "p1", ID: "vpc-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := NetworkVPCGetArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

func TestNetworkVPCUpdateArgs_Validate(t *testing.T) {
	t.Run("happy path with name", func(t *testing.T) {
		args := NetworkVPCUpdateArgs{ProjectID: "p1", ID: "vpc-001", Name: "new-name"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("happy path with tags changed", func(t *testing.T) {
		args := NetworkVPCUpdateArgs{ProjectID: "p1", ID: "vpc-001", TagsChanged: true}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("no name and no tags", func(t *testing.T) {
		args := NetworkVPCUpdateArgs{ProjectID: "p1", ID: "vpc-001"}
		err := args.Validate()
		if err == nil {
			t.Fatal("expected error when no flags provided")
		}
		if !strings.Contains(err.Error(), "at least one") {
			t.Errorf("error %q does not contain 'at least one'", err.Error())
		}
	})
}

func TestNetworkVPCDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPCDeleteArgs{ProjectID: "p1", ID: "vpc-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := NetworkVPCDeleteArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

func TestNetworkVPCListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkVPCListArgs{ProjectID: "proj-123"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty project ID", func(t *testing.T) {
		args := NetworkVPCListArgs{}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty project ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkVPCCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-new", "my-vpc"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := NetworkVPCCreate(context.Background(), srv.Client(), validNetworkVPCCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "vpc-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPCCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := NetworkVPCCreate(context.Background(), srv.Client(), validNetworkVPCCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating VPC") {
		t.Errorf("error %q does not contain 'creating VPC'", err.Error())
	}
}

func TestNetworkVPCCreate_AsyncResponse(t *testing.T) {
	srv := newArubaTestServer(t)
	// Async creation: server returns 202 with metadata but the SDK may not populate Raw().
	id, name := "vpc-new", "my-vpc"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(202, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := NetworkVPCCreate(context.Background(), srv.Client(), validNetworkVPCCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "vpc-new") && !strings.Contains(out, "my-vpc") {
		t.Errorf("expected ID or name in output, got: %s", out)
	}
}

func TestNetworkVPCList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-001", "my-vpc"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := NetworkVPCList(context.Background(), srv.Client(), NetworkVPCListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "vpc-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkVPCList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{}))

	out := captureStdout(func() {
		err := NetworkVPCList(context.Background(), srv.Client(), NetworkVPCListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No VPCs found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestNetworkVPCCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "x", "--region", "ITBG-Bergamo"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkVPCListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"network", "vpc", "list"})
	if err == nil {
		t.Fatal("expected error")
	}
}
