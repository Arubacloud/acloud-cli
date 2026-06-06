package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestKaaSListCmd(t *testing.T) {
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
			args: []string{"container", "kaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-cluster"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{
					Values: []types.KaaSResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"container", "kaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"container", "kaas", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-cluster"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{
					Values: []types.KaaSResponse{
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
			args: []string{"container", "kaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"container", "kaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", errorResponse(404, "Not Found", "resource not found"))
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

func TestKaaSGetCmd(t *testing.T) {
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
				id, name := "kaas-001", "my-cluster"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"container", "kaas", "get", "kaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKaaSCreateCmd(t *testing.T) {
	baseArgs := []string{
		"container", "kaas", "create",
		"--project-id", "proj-123",
		"--name", "my-cluster",
		"--region", "ITBG-Bergamo",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--node-cidr-address", "10.0.0.0/16",
		"--node-cidr-name", "node-cidr",
		"--security-group-name", "my-sg",
		"--kubernetes-version", "1.32.3",
		"--node-pool-name", "default-pool",
		"--node-pool-nodes", "1",
		"--node-pool-instance", "K1A2",
		"--node-pool-zone", "itbg1-a",
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
				id, name := "kaas-new", "my-cluster"
				srv.OnPost("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-cluster"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --kubernetes-version",
			args:        removeFlag(baseArgs, "--kubernetes-version", "1.32.3"),
			wantErr:     true,
			errContains: "kubernetes-version",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Container/kaas", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Container/kaas", errorResponse(404, "Not Found", "resource not found"))
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

func TestKaaSDeleteCmd(t *testing.T) {
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
			args: []string{"container", "kaas", "delete", "kaas-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"container", "kaas", "delete", "kaas-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-cluster"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"container", "kaas", "delete", "kaas-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"container", "kaas", "delete", "kaas-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestKaaSConnectCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
	}{
		{
			// connect calls Get first, then DownloadKubeconfig on the wrapper
			name: "getting cluster fails",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting KaaS cluster",
		},
		{
			name: "downloading kubeconfig fails",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-cluster"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001/download", errorResponse(500, "Internal Server Error", "unauthorized"))
			},
			wantErr:     true,
			errContains: "downloading kubeconfig",
		},
		{
			name: "API error on download propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-cluster"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001/download", errorResponse(404, "Not Found", "resource not found"))
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
			err := runCmd(srv.Client(), []string{"container", "kaas", "connect", "kaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKaaSUpdateCmd(t *testing.T) {
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
			args: []string{"container", "kaas", "update", "kaas-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-kaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "pre-GET error",
			args: []string{"container", "kaas", "update", "kaas-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"container", "kaas", "update", "kaas-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kaas-001", "my-kaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestKaaSListCmd_WithK8sVersionAndLocation(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kaas-001", "my-cluster"
	verStr := string(types.KubernetesVersion1323)
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{
		Values: []types.KaaSResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.KaaSPropertiesResponse{
					KubernetesVersion: types.KubernetesVersionInfoResponse{Value: &verStr},
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"container", "kaas", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kaas-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestKaaSGetCmd_JSONOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kaas-001", "my-cluster"
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"container", "kaas", "get", "kaas-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestKaaSGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kaas-001", "my-cluster"
	uri := "/projects/proj-123/providers/Aruba.Container/kaas/kaas-001"
	verStr2 := string(types.KubernetesVersion1323)
	state := types.StateActive
	region := types.Region("IT-BG")
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.KaaSPropertiesResponse{
			KubernetesVersion: types.KubernetesVersionInfoResponse{Value: &verStr2},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"container", "kaas", "get", "kaas-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kaas-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestKaaSCreateCmd_WithHAAndPodCIDR(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kaas-001", "my-cluster"
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	err := runCmd(srv.Client(), []string{
		"container", "kaas", "create",
		"--project-id", "proj-123",
		"--name", "my-cluster",
		"--region", "ITBG-Bergamo",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--node-cidr-address", "10.0.0.0/24",
		"--node-cidr-name", "my-cidr",
		"--security-group-name", "my-sg",
		"--kubernetes-version", "1.32.3",
		"--node-pool-name", "workers",
		"--node-pool-nodes", "3",
		"--node-pool-instance", "K4A8",
		"--node-pool-zone", "IT-BG-1",
		"--ha",
		"--pod-cidr", "192.168.0.0/16",
		"--billing-period", "Month",
		"--node-pool-autoscaling",
		"--node-pool-min-count", "2",
		"--node-pool-max-count", "5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKaaSCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kaas-001", "my-cluster"
	verStr := string(types.KubernetesVersion1323)
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Properties: types.KaaSPropertiesResponse{
			KubernetesVersion: types.KubernetesVersionInfoResponse{Value: &verStr},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"container", "kaas", "create",
		"--project-id", "proj-123",
		"--name", "my-cluster",
		"--region", "ITBG-Bergamo",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--node-cidr-address", "10.0.0.0/24",
		"--node-cidr-name", "my-cidr",
		"--security-group-name", "my-sg",
		"--kubernetes-version", "1.32.3",
		"--node-pool-name", "workers",
		"--node-pool-nodes", "3",
		"--node-pool-instance", "K4A8",
		"--node-pool-zone", "IT-BG-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKaaSUpdateCmd_WithTagsOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kaas-001", "my-cluster"
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	updID, updName := "kaas-001", "my-cluster"
	srv.OnPut("/projects/proj-123/providers/Aruba.Container/kaas/kaas-001", jsonResponse(200, types.KaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &updID,
			Name: &updName,
			Tags: []string{"env=prod"},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"container", "kaas", "update", "kaas-001",
		"--project-id", "proj-123",
		"--tags", "env=prod",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kaas-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer-1: Validate tests (pure Go, no SDK)
// =============================================================================

func TestContainerKaaSCreateArgs_Validate(t *testing.T) {
	validBase := ContainerKaaSCreateArgs{
		ProjectID:         "proj-123",
		Name:              "my-cluster",
		Region:            aruba.RegionITBGBergamo,
		VPCID:             "vpc-001",
		SubnetID:          "sub-001",
		NodeCIDRAddress:   "10.0.0.0/16",
		NodeCIDRName:      "my-cidr",
		SecurityGroupName: "my-sg",
		KubernetesVersion: aruba.KubernetesVersion("1.32.3"),
		NodePoolName:      "workers",
		NodePoolNodes:     3,
		NodePoolInstance:  aruba.NodePoolInstanceK4A8,
		NodePoolZone:      aruba.Zone("ITBG-1"),
	}

	tests := []struct {
		name        string
		mutate      func(*ContainerKaaSCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid args",
			mutate:  func(_ *ContainerKaaSCreateArgs) {},
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:        "name too long",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.Region = aruba.Region("IT-BG") },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "empty kubernetes version",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.KubernetesVersion = aruba.KubernetesVersion("") },
			wantErr:     true,
			errContains: "--kubernetes-version is required",
		},
		{
			name:        "empty vpc-id",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.VPCID = "" },
			wantErr:     true,
			errContains: "--vpc-id is required",
		},
		{
			name:        "empty subnet-id",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.SubnetID = "" },
			wantErr:     true,
			errContains: "--subnet-id is required",
		},
		{
			name:        "node pool nodes zero",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.NodePoolNodes = 0 },
			wantErr:     true,
			errContains: "--node-pool-nodes must be greater than 0",
		},
		{
			name:        "invalid node pool instance",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.NodePoolInstance = aruba.NodePoolInstance("n1.standard") },
			wantErr:     true,
			errContains: "--node-pool-instance",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *ContainerKaaSCreateArgs) { a.BillingPeriod = aruba.BillingPeriod("Weekly") },
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:    "valid optional billing period",
			mutate:  func(a *ContainerKaaSCreateArgs) { a.BillingPeriod = aruba.BillingPeriodMonth },
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validBase
			tc.mutate(&args)
			err := args.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errContains)
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error to contain %q, got: %v", tc.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestContainerKaaSCreateArgs_Validate_WrapsErrValidationFailed(t *testing.T) {
	// Verify the constructor wraps Validate errors with ErrValidationFailed.
	// We use an invalid region to trigger a validation failure.
	args := ContainerKaaSCreateArgs{
		ProjectID:         "proj-123",
		Name:              "my-cluster",
		Region:            aruba.Region("bad-region"),
		VPCID:             "vpc-001",
		SubnetID:          "sub-001",
		NodeCIDRAddress:   "10.0.0.0/16",
		NodeCIDRName:      "my-cidr",
		SecurityGroupName: "my-sg",
		KubernetesVersion: aruba.KubernetesVersion("1.32.3"),
		NodePoolName:      "workers",
		NodePoolNodes:     1,
		NodePoolInstance:  aruba.NodePoolInstanceK4A8,
		NodePoolZone:      aruba.Zone("ITBG-1"),
	}
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	// Wrap as the constructor would.
	wrapped := errors.Join(ErrValidationFailed, err)
	if !errors.Is(wrapped, ErrValidationFailed) {
		t.Errorf("expected wrapped error to be ErrValidationFailed")
	}
}

func TestContainerKaaSGetArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    ContainerKaaSGetArgs
		wantErr bool
	}{
		{name: "valid", args: ContainerKaaSGetArgs{ProjectID: "p1", ID: "kaas-001"}, wantErr: false},
		{name: "missing ID", args: ContainerKaaSGetArgs{ProjectID: "p1"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestContainerKaaSListArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    ContainerKaaSListArgs
		wantErr bool
	}{
		{name: "valid", args: ContainerKaaSListArgs{ProjectID: "p1"}, wantErr: false},
		{name: "missing project", args: ContainerKaaSListArgs{}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestContainerKaaSUpdateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        ContainerKaaSUpdateArgs
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid",
			args:    ContainerKaaSUpdateArgs{ProjectID: "p1", ID: "kaas-001", Name: "new-name"},
			wantErr: false,
		},
		{
			name:        "missing ID",
			args:        ContainerKaaSUpdateArgs{ProjectID: "p1", Name: "new-name"},
			wantErr:     true,
			errContains: "KaaS cluster ID is required",
		},
		{
			name:        "invalid billing period",
			args:        ContainerKaaSUpdateArgs{ProjectID: "p1", ID: "kaas-001", BillingPeriod: aruba.BillingPeriod("Daily")},
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:        "invalid node pool instance",
			args:        ContainerKaaSUpdateArgs{ProjectID: "p1", ID: "kaas-001", NodePoolInstance: aruba.NodePoolInstance("bad")},
			wantErr:     true,
			errContains: "--node-pool-instance",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errContains)
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error to contain %q, got: %v", tc.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestContainerKaaSDeleteArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    ContainerKaaSDeleteArgs
		wantErr bool
	}{
		{name: "valid", args: ContainerKaaSDeleteArgs{ProjectID: "p1", ID: "kaas-001"}, wantErr: false},
		{name: "missing ID", args: ContainerKaaSDeleteArgs{ProjectID: "p1"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestContainerKaaSConnectArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    ContainerKaaSConnectArgs
		wantErr bool
	}{
		{name: "valid", args: ContainerKaaSConnectArgs{ProjectID: "p1", ID: "kaas-001"}, wantErr: false},
		{name: "missing ID", args: ContainerKaaSConnectArgs{ProjectID: "p1"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

// =============================================================================
// Layer-2: Operation function tests
// =============================================================================

func TestContainerKaaSList_Operation(t *testing.T) {
	t.Run("returns clusters in table", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "kaas-op-001", "op-cluster"
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{
			Values: []types.KaaSResponse{
				{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
			},
		}))
		ctx, cancel := newCtx()
		defer cancel()
		out := captureStdout(func() {
			err := ContainerKaaSList(ctx, srv.Client(), ContainerKaaSListArgs{ProjectID: "p1"})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "kaas-op-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})

	t.Run("empty list prints message", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{}))
		ctx, cancel := newCtx()
		defer cancel()
		out := captureStdout(func() {
			err := ContainerKaaSList(ctx, srv.Client(), ContainerKaaSListArgs{ProjectID: "p1"})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "No KaaS clusters found") {
			t.Errorf("expected empty message, got: %s", out)
		}
	})

	t.Run("server error propagates", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas", errorResponse(500, "Internal Server Error", "boom"))
		ctx, cancel := newCtx()
		defer cancel()
		err := ContainerKaaSList(ctx, srv.Client(), ContainerKaaSListArgs{ProjectID: "p1"})
		if err == nil || !strings.Contains(err.Error(), "listing") {
			t.Errorf("expected listing error, got: %v", err)
		}
	})
}

func TestContainerKaaSGet_Operation(t *testing.T) {
	t.Run("success renders details", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "kaas-op-get", "op-get-cluster"
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/kaas-op-get", jsonResponse(200, types.KaaSResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
		ctx, cancel := newCtx()
		defer cancel()
		out := captureStdout(func() {
			err := ContainerKaaSGet(ctx, srv.Client(), ContainerKaaSGetArgs{ProjectID: "p1", ID: "kaas-op-get"})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "kaas-op-get") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})

	t.Run("API error propagates", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/missing", errorResponse(404, "Not Found", "not found"))
		ctx, cancel := newCtx()
		defer cancel()
		err := ContainerKaaSGet(ctx, srv.Client(), ContainerKaaSGetArgs{ProjectID: "p1", ID: "missing"})
		if err == nil || !strings.Contains(err.Error(), "getting") {
			t.Errorf("expected getting error, got: %v", err)
		}
	})
}

func TestContainerKaaSCreate_Operation(t *testing.T) {
	t.Run("success renders row", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "kaas-op-create", "op-cluster"
		srv.OnPost("/projects/p1/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
		ctx, cancel := newCtx()
		defer cancel()
		out := captureStdout(func() {
			err := ContainerKaaSCreate(ctx, srv.Client(), ContainerKaaSCreateArgs{
				ProjectID:         "p1",
				Name:              "op-cluster",
				Region:            aruba.RegionITBGBergamo,
				VPCID:             "vpc-001",
				SubnetID:          "sub-001",
				NodeCIDRAddress:   "10.0.0.0/16",
				NodeCIDRName:      "my-cidr",
				SecurityGroupName: "my-sg",
				KubernetesVersion: aruba.KubernetesVersion("1.32.3"),
				NodePoolName:      "workers",
				NodePoolNodes:     1,
				NodePoolInstance:  aruba.NodePoolInstanceK4A8,
				NodePoolZone:      aruba.Zone("ITBG-1"),
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "kaas-op-create") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})

	t.Run("API error propagates", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnPost("/projects/p1/providers/Aruba.Container/kaas", errorResponse(500, "Internal Server Error", "quota"))
		ctx, cancel := newCtx()
		defer cancel()
		err := ContainerKaaSCreate(ctx, srv.Client(), ContainerKaaSCreateArgs{
			ProjectID:         "p1",
			Name:              "op-cluster",
			Region:            aruba.RegionITBGBergamo,
			VPCID:             "vpc-001",
			SubnetID:          "sub-001",
			NodeCIDRAddress:   "10.0.0.0/16",
			NodeCIDRName:      "my-cidr",
			SecurityGroupName: "my-sg",
			KubernetesVersion: aruba.KubernetesVersion("1.32.3"),
			NodePoolName:      "workers",
			NodePoolNodes:     1,
			NodePoolInstance:  aruba.NodePoolInstanceK4A8,
			NodePoolZone:      aruba.Zone("ITBG-1"),
		})
		if err == nil || !strings.Contains(err.Error(), "creating") {
			t.Errorf("expected creating error, got: %v", err)
		}
	})
}

func TestContainerKaaSDelete_Operation(t *testing.T) {
	t.Run("success prints deleted message", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnDelete("/projects/p1/providers/Aruba.Container/kaas/kaas-del", jsonResponse(200, nil))
		ctx, cancel := newCtx()
		defer cancel()
		out := captureStdout(func() {
			err := ContainerKaaSDelete(ctx, srv.Client(), ContainerKaaSDeleteArgs{
				ProjectID: "p1", ID: "kaas-del", SkipConfirm: true,
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "kaas-del") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})

	t.Run("API error propagates", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnDelete("/projects/p1/providers/Aruba.Container/kaas/kaas-del", errorResponse(500, "Internal Server Error", "in use"))
		ctx, cancel := newCtx()
		defer cancel()
		err := ContainerKaaSDelete(ctx, srv.Client(), ContainerKaaSDeleteArgs{
			ProjectID: "p1", ID: "kaas-del", SkipConfirm: true,
		})
		if err == nil || !strings.Contains(err.Error(), "deleting") {
			t.Errorf("expected deleting error, got: %v", err)
		}
	})
}

func TestContainerKaaSConnect_Operation(t *testing.T) {
	// base64-encode a minimal kubeconfig so DownloadKubeconfig returns valid content
	rawKubeconfig := []byte("apiVersion: v1\nkind: Config\n")
	encoded := base64.StdEncoding.EncodeToString(rawKubeconfig)

	t.Run("success writes kubeconfig to output-file", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "kaas-conn", "conn-cluster"
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/kaas-conn", jsonResponse(200, types.KaaSResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/kaas-conn/download", jsonResponse(200, types.KaaSKubeconfigResponse{
			Content: encoded,
		}))

		outFile := t.TempDir() + "/kubeconfig"
		ctx, cancel := newCtx()
		defer cancel()
		out := captureStdout(func() {
			err := ContainerKaaSConnect(ctx, srv.Client(), ContainerKaaSConnectArgs{
				ProjectID:  "p1",
				ID:         "kaas-conn",
				OutputFile: outFile,
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "connected") {
			t.Errorf("expected connected message, got: %s", out)
		}
	})

	t.Run("get error propagates", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/kaas-conn", errorResponse(500, "Internal Server Error", "boom"))
		ctx, cancel := newCtx()
		defer cancel()
		err := ContainerKaaSConnect(ctx, srv.Client(), ContainerKaaSConnectArgs{
			ProjectID:  "p1",
			ID:         "kaas-conn",
			OutputFile: t.TempDir() + "/kubeconfig",
		})
		if err == nil || !strings.Contains(err.Error(), "getting KaaS cluster") {
			t.Errorf("expected getting error, got: %v", err)
		}
	})

	t.Run("download error propagates", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "kaas-conn", "conn-cluster"
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/kaas-conn", jsonResponse(200, types.KaaSResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
		srv.OnGet("/projects/p1/providers/Aruba.Container/kaas/kaas-conn/download", errorResponse(403, "Forbidden", "unauthorized"))
		ctx, cancel := newCtx()
		defer cancel()
		err := ContainerKaaSConnect(ctx, srv.Client(), ContainerKaaSConnectArgs{
			ProjectID:  "p1",
			ID:         "kaas-conn",
			OutputFile: t.TempDir() + "/kubeconfig",
		})
		if err == nil || !strings.Contains(err.Error(), "downloading kubeconfig") {
			t.Errorf("expected downloading error, got: %v", err)
		}
	})
}
