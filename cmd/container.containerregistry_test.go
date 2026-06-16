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
			name: "--output yaml emits valid YAML",
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
					Values: []types.ContainerRegistryResponse{
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
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
					Values: []types.ContainerRegistryResponse{
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
			args: []string{"container", "containerregistry", "list", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cr-001", "my-registry"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
					Values: []types.ContainerRegistryResponse{
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
		"--region", "ITBG-Bergamo",
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
			args:        removeFlag(baseArgs, "--region", "ITBG-Bergamo"),
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
	region := types.Region("ITBG-Bergamo")
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
		"--region", "ITBG-Bergamo",
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
	region := types.Region("ITBG-Bergamo")
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
	region := types.Region("ITBG-Bergamo")
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
			PublicIp:          types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"},
			VPC:               types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"},
			Subnet:            types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001"},
			SecurityGroup:     types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001"},
			BlockStorage:      types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Storage/blockStorages/bs-001"},
			BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &period},
			AdminUser:         &types.UserCredentialCommon{Username: "admin"},
			ConcurrentUsers:   &concurrentUsers,
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

// =============================================================================
// Layer 1 — Validate tests (pure Go, no SDK, no server)
// =============================================================================

func validContainerRegistryCreateArgs() ContainerContainerRegistryCreateArgs {
	return ContainerContainerRegistryCreateArgs{
		ProjectID:      "proj-123",
		Name:           "my-registry",
		Region:         aruba.RegionITBGBergamo,
		VPCID:          "vpc-001",
		SubnetID:       "sub-001",
		SGID:           "sg-001",
		PublicIPID:     "eip-001",
		BlockStorageID: "vol-001",
		Tags:           []string{},
	}
}

func TestContainerContainerRegistryCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*ContainerContainerRegistryCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid args",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *ContainerContainerRegistryCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *ContainerContainerRegistryCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		{
			name:        "invalid region",
			mutate:      func(a *ContainerContainerRegistryCreateArgs) { a.Region = aruba.Region("ZZ-Invalid") },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *ContainerContainerRegistryCreateArgs) { a.BillingPeriod = aruba.BillingPeriod("Quarterly") },
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:    "valid billing period Hour",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.BillingPeriod = aruba.BillingPeriodHour },
			wantErr: false,
		},
		{
			name:    "billing period empty (optional) is valid",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.BillingPeriod = "" },
			wantErr: false,
		},
		{
			name: "invalid size flavor",
			mutate: func(a *ContainerContainerRegistryCreateArgs) {
				a.SizeFlavor = aruba.ContainerRegistrySizeFlavor("Huge")
			},
			wantErr:     true,
			errContains: "--concurrent-users",
		},
		{
			name:    "valid size flavor Small",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.SizeFlavor = aruba.ContainerRegistrySizeFlavorSmall },
			wantErr: false,
		},
		{
			name:    "valid size flavor Medium",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.SizeFlavor = aruba.ContainerRegistrySizeFlavorMedium },
			wantErr: false,
		},
		{
			name: "valid size flavor HighPerf",
			mutate: func(a *ContainerContainerRegistryCreateArgs) {
				a.SizeFlavor = aruba.ContainerRegistrySizeFlavorHighPerf
			},
			wantErr: false,
		},
		{
			name:    "size flavor empty (optional) is valid",
			mutate:  func(a *ContainerContainerRegistryCreateArgs) { a.SizeFlavor = "" },
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validContainerRegistryCreateArgs()
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
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestContainerContainerRegistryCreateArgs_Validate_WrappedByConstructor(t *testing.T) {
	args := validContainerRegistryCreateArgs()
	args.Name = "x" // too short
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	wrapped := errors.Join(ErrValidationFailed, err)
	if !errors.Is(wrapped, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed in chain, got: %v", wrapped)
	}
}

func TestContainerContainerRegistryUpdateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        ContainerContainerRegistryUpdateArgs
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid with name",
			args:    ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001", Name: "new-name"},
			wantErr: false,
		},
		{
			name:    "valid with tags changed",
			args:    ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001", TagsChanged: true},
			wantErr: false,
		},
		{
			name:    "valid with billing period",
			args:    ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001", BillingPeriod: aruba.BillingPeriodMonth},
			wantErr: false,
		},
		{
			name:    "valid with size flavor",
			args:    ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001", SizeFlavor: aruba.ContainerRegistrySizeFlavorSmall},
			wantErr: false,
		},
		{
			name:        "no fields changed",
			args:        ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name:        "invalid billing period",
			args:        ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001", BillingPeriod: "Quarterly"},
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:        "invalid size flavor",
			args:        ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", ID: "cr-001", SizeFlavor: "Huge"},
			wantErr:     true,
			errContains: "--concurrent-users",
		},
		{
			name:        "missing ID",
			args:        ContainerContainerRegistryUpdateArgs{ProjectID: "proj-123", Name: "new-name"},
			wantErr:     true,
			errContains: "container registry ID is required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
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

func TestContainerContainerRegistryCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-new", "my-registry"
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ContainerContainerRegistryCreate(context.Background(), srv.Client(), validContainerRegistryCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cr-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestContainerContainerRegistryCreate_AsyncResponse(t *testing.T) {
	srv := newArubaTestServer(t)
	// SDK validates metadata; use a valid response with ID and Name set.
	id, name := "reg-new", "my-registry"
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(202, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ContainerContainerRegistryCreate(context.Background(), srv.Client(), validContainerRegistryCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "my-registry") && !strings.Contains(out, "reg-new") {
		t.Errorf("expected name or ID in output, got: %s", out)
	}
}

func TestContainerContainerRegistryCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := ContainerContainerRegistryCreate(context.Background(), srv.Client(), validContainerRegistryCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating container registry") {
		t.Errorf("error %q does not contain 'creating container registry'", err.Error())
	}
}

func TestContainerContainerRegistryList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-001", "my-registry"
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
		Values: []types.ContainerRegistryResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := ContainerContainerRegistryList(context.Background(), srv.Client(), ContainerContainerRegistryListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cr-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestContainerContainerRegistryList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{}))

	out := captureStdout(func() {
		err := ContainerContainerRegistryList(context.Background(), srv.Client(), ContainerContainerRegistryListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No container registries found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestContainerContainerRegistryList_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", errorResponse(500, "Internal Server Error", "boom"))

	err := ContainerContainerRegistryList(context.Background(), srv.Client(), ContainerContainerRegistryListArgs{
		ProjectID: "proj-123",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "listing container registries") {
		t.Errorf("error %q does not contain 'listing container registries'", err.Error())
	}
}

// =============================================================================
// Coverage-gap tests: ErrValidationFailed and ErrParsingFailed branches
// =============================================================================

func TestContainerContainerRegistryCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"container", "containerregistry", "create",
		"--project-id", "proj-123", "--name", "x", // "x" is too short → Validate fails
		"--region", "ITBG-Bergamo",
		"--public-ip-id", "eip-001", "--vpc-id", "vpc-001",
		"--subnet-id", "sub-001", "--security-group-id", "sg-001",
		"--block-storage-id", "vol-001",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args' in error, got: %v", err)
	}
}

func TestContainerContainerRegistryListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"container", "containerregistry", "list"})
	if err == nil {
		t.Fatal("expected error (no project-id, no context)")
	}
}
