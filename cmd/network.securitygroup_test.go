package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "IT-BG"},
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
			args:        []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--region", "IT-BG"},
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
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "IT-BG"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "IT-BG"},
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
		"--region", "IT-BG",
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
