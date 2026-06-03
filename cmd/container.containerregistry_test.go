package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestContainerRegistryListCmd(t *testing.T) {
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
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
					Values: []types.ContainerRegistryResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cr-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
					Values: []types.ContainerRegistryResponse{
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
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", errorResponse(404, "Not Found", "resource not found"))
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

func TestContainerRegistryGetCmd(t *testing.T) {
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
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", jsonResponse(200, types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cr-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"container", "containerregistry", "get", "cr-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestContainerRegistryCreateCmd(t *testing.T) {
	baseArgs := []string{
		"container", "containerregistry", "create",
		"--project-id", "proj-123",
		"--name", "my-registry",
		"--region", "IT-BG",
		"--public-ip-id", "eip-001",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--security-group-id", "sg-001",
		"--block-storage-id", "vol-001",
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
				id, name := "cr-new", "my-registry"
				srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cr-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-registry"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        removeFlag(baseArgs, "--region", "IT-BG"),
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", errorResponse(404, "Not Found", "resource not found"))
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

func TestContainerRegistryDeleteCmd(t *testing.T) {
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
			args: []string{"container", "containerregistry", "delete", "cr-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Container/registries/cr-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cr-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"container", "containerregistry", "delete", "cr-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", jsonResponse(200, types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cr-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"container", "containerregistry", "delete", "cr-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Container/registries/cr-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"container", "containerregistry", "delete", "cr-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Container/registries/cr-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestContainerRegistryUpdateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success --name",
			args: []string{"container", "containerregistry", "update", "reg-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "reg-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/reg-001", jsonResponse(200, types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Container/registries/reg-001", jsonResponse(200, types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "reg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"container", "containerregistry", "update", "reg-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-GET error",
			args: []string{"container", "containerregistry", "update", "reg-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/reg-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"container", "containerregistry", "update", "reg-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "reg-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/reg-001", jsonResponse(200, types.ContainerRegistryResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Container/registries/reg-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestContainerRegistryCreateCmd_OptionalFlags(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-001", "my-registry"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"container", "containerregistry", "create",
		"--project-id", "proj-123",
		"--name", "my-registry",
		"--region", "IT-BG",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--security-group-id", "sg-001",
		"--public-ip-id", "eip-001",
		"--block-storage-id", "vol-001",
		"--admin-username", "admin",
		"--billing-period", "Month",
		"--concurrent-users", "Small",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContainerRegistryListCmd_WithLocation(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-001", "my-registry"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
		Values: []types.ContainerRegistryResponse{
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
	out, err := runCmdCapture(srv.Client(), []string{"container", "containerregistry", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cr-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestContainerRegistryGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-001", "my-registry"
	uri := "/projects/proj-123/providers/Aruba.Container/registries/cr-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", jsonResponse(200, types.ContainerRegistryResponse{
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
	out, err := runCmdCapture(srv.Client(), []string{"container", "containerregistry", "get", "cr-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cr-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestContainerRegistryGetCmd_JSONOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-001", "my-registry"
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", jsonResponse(200, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"container", "containerregistry", "get", "cr-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestContainerRegistryGetCmd_WithAllOptionalProps(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-001", "my-registry"
	period := types.BillingPeriodHour
	concurrentUsers := "Small"
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-001", jsonResponse(200, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.ContainerRegistryPropertiesResponse{
			PublicIp:        types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"},
			VPC:             types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"},
			Subnet:          types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001"},
			SecurityGroup:   types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001"},
			BlockStorage:    types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Storage/blockStorages/bs-001"},
			BillingPlanCommon:     &types.BillingPlanCommon{BillingPeriod: &period},
			AdminUser:       &types.UserCredentialCommon{Username: "admin"},
			ConcurrentUsers: &concurrentUsers,
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"container", "containerregistry", "get", "cr-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cr-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
