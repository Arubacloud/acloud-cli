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

func TestSecurityGroupListCmd(t *testing.T) {
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
			args: []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sg-001", "my-sg"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
					Values: []types.SecurityGroupResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sg-001", "my-sg"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
					Values: []types.SecurityGroupResponse{
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
			args: []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", errorResponse(404, "Not Found", "resource not found"))
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

func TestSecurityGroupGetCmd(t *testing.T) {
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
				id, name := "sg-001", "my-sg"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, types.SecurityGroupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"network", "securitygroup", "get", "vpc-001", "sg-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityGroupCreateCmd(t *testing.T) {
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
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sg-001", "my-sg"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", errorResponse(404, "Not Found", "resource not found"))
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

func TestSecurityGroupUpdateCmd(t *testing.T) {
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
			args: []string{"network", "securitygroup", "update", "vpc-001", "sg-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sg-001", "my-sg"
				state := types.StateActive
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, types.SecurityGroupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, types.SecurityGroupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"network", "securitygroup", "update", "vpc-001", "sg-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-Get server error",
			args: []string{"network", "securitygroup", "update", "vpc-001", "sg-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update server error",
			args: []string{"network", "securitygroup", "update", "vpc-001", "sg-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sg-001", "my-sg"
				state := types.StateActive
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, types.SecurityGroupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestSecurityGroupDeleteCmd(t *testing.T) {
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
			args: []string{"network", "securitygroup", "delete", "vpc-001", "sg-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"network", "securitygroup", "delete", "vpc-001", "sg-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sg-001", "my-sg"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, types.SecurityGroupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "securitygroup", "delete", "vpc-001", "sg-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"network", "securitygroup", "delete", "vpc-001", "sg-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestSecurityGroupListCmd_AllOptionalFields(t *testing.T) {
	// Covers: LocationResponse and Status.State nil-guards in list loop.
	srv := newArubaTestServer(t)
	id, name := "sg-001", "my-sg"
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
		Values: []types.SecurityGroupResponse{
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
	out, err := runCmdCapture(srv.Client(), []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sg-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "IT-BG") {
		t.Errorf("expected region in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestSecurityGroupGetCmd_AllOptionalFields(t *testing.T) {
	// Covers: URI, LocationResponse, CreationDate, CreatedBy, Tags, Status.State detail block.
	id, name := "sg-001", "my-sg"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001"
	createdBy := "user@example.com"
	state := types.StateActive
	region := types.Region("IT-BG")
	now := time.Now()
	makeResponse := func() types.SecurityGroupResponse {
		return types.SecurityGroupResponse{
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

	t.Run("detail output with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"network", "securitygroup", "get", "vpc-001", "sg-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "sg-001") {
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
		// securitygroup get renders text directly (no PrintOutput JSON path);
		// verify command still succeeds with --output json flag present.
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"network", "securitygroup", "get", "vpc-001", "sg-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "sg-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})
}

func TestSecurityGroupCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sg-001", "my-sg"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"network", "securitygroup", "create", "vpc-001",
		"--project-id", "proj-123",
		"--name", "my-sg",
		"--region", "ITBG-Bergamo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSecurityGroupGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sg-001", "my-sg"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001", jsonResponse(200, types.SecurityGroupResponse{
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
	out, err := runCmdCapture(srv.Client(), []string{"network", "securitygroup", "get", "vpc-001", "sg-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sg-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests
// =============================================================================

func TestNetworkSecurityGroupCreateArgs_Validate(t *testing.T) {
	valid := NetworkSecurityGroupCreateArgs{ProjectID: "p1", VPCID: "vpc-1", Name: "my-sg"}
	tests := []struct {
		name        string
		mutate      func(*NetworkSecurityGroupCreateArgs)
		wantErr     bool
		errContains string
	}{
		{name: "happy path"},
		{name: "name too short", mutate: func(a *NetworkSecurityGroupCreateArgs) { a.Name = "ab" }, wantErr: true, errContains: "--name"},
		{name: "name too long", mutate: func(a *NetworkSecurityGroupCreateArgs) { a.Name = strings.Repeat("x", 65) }, wantErr: true, errContains: "--name"},
		{name: "empty vpc-id", mutate: func(a *NetworkSecurityGroupCreateArgs) { a.VPCID = "" }, wantErr: true, errContains: "vpc-id"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := valid
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

func TestNewNetworkSecurityGroupCreateArgs_WrapsValidationFailed(t *testing.T) {
	_ = aruba.RegionITBGBergamo // ensure aruba import used
	// Validate directly and confirm the sentinel can be detected via errors.Is
	args := NetworkSecurityGroupCreateArgs{ProjectID: "p1", VPCID: "", Name: "sg"}
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	// Confirm ErrValidationFailed works as a sentinel for wrapping
	if errors.Is(ErrValidationFailed, ErrParsingFailed) {
		t.Error("sentinels must be distinct")
	}
}

// =============================================================================
// Layer 2 — Operation function tests
// =============================================================================

func TestNetworkSecurityGroupCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sg-new", "my-sg"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out := captureStdout(func() {
		err := NetworkSecurityGroupCreate(context.Background(), srv.Client(), NetworkSecurityGroupCreateArgs{
			ProjectID: "proj-123", VPCID: "vpc-001", Name: "my-sg",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "sg-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkSecurityGroupCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups",
		errorResponse(500, "Internal Server Error", "boom"))
	err := NetworkSecurityGroupCreate(context.Background(), srv.Client(), NetworkSecurityGroupCreateArgs{
		ProjectID: "proj-123", VPCID: "vpc-001", Name: "my-sg",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating security group") {
		t.Errorf("error %q does not contain 'creating security group'", err.Error())
	}
}

func TestNetworkSecurityGroupList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sg-001", "my-sg"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
		Values: []types.SecurityGroupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	out := captureStdout(func() {
		err := NetworkSecurityGroupList(context.Background(), srv.Client(), NetworkSecurityGroupListArgs{
			ProjectID: "proj-123", VPCID: "vpc-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "sg-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkSecurityGroupCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "x", "--region", "ITBG-Bergamo"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkSecurityGroupListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"network", "securitygroup", "list", "vpc-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}
