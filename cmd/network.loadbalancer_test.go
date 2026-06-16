package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestLoadBalancerListCmd(t *testing.T) {
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
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
					Values: []types.LoadBalancerResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "lb-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
					Values: []types.LoadBalancerResponse{
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
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
					Values: []types.LoadBalancerResponse{
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
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
					Values: []types.LoadBalancerResponse{
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
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
					Values: []types.LoadBalancerResponse{
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
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", errorResponse(404, "Not Found", "resource not found"))
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

func TestLoadBalancerGetCmd(t *testing.T) {
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
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-001", jsonResponse(200, types.LoadBalancerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "lb-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"network", "loadbalancer", "get", "lb-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestLoadBalancerListCmd_WithAllFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "lb-001", "my-lb"
	region := types.Region("IT-BG")
	state := types.StateActive
	addr := "192.0.2.1"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
		Values: []types.LoadBalancerResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.LoadBalancerPropertiesResponse{Address: &addr},
				Status:     types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "loadbalancer", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "lb-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestLoadBalancerGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "lb-001", "my-lb"
	uri := "/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	addr := "192.0.2.1"
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-001", jsonResponse(200, types.LoadBalancerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.LoadBalancerPropertiesResponse{Address: &addr},
		Status:     types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "loadbalancer", "get", "lb-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "lb-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func TestNetworkLoadBalancerListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkLoadBalancerListArgs{ProjectID: "proj-123"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty project ID", func(t *testing.T) {
		args := NetworkLoadBalancerListArgs{}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty project ID")
		}
	})
}

func TestNetworkLoadBalancerGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkLoadBalancerGetArgs{ProjectID: "p1", ID: "lb-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := NetworkLoadBalancerGetArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkLoadBalancerList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "lb-001", "my-lb"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
		Values: []types.LoadBalancerResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := NetworkLoadBalancerList(context.Background(), srv.Client(), NetworkLoadBalancerListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "lb-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkLoadBalancerList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{}))

	out := captureStdout(func() {
		err := NetworkLoadBalancerList(context.Background(), srv.Client(), NetworkLoadBalancerListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No Load Balancers found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestNetworkLoadBalancerGet_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "lb-001", "my-lb"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-001", jsonResponse(200, types.LoadBalancerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	out := captureStdout(func() {
		err := NetworkLoadBalancerGet(context.Background(), srv.Client(), NetworkLoadBalancerGetArgs{
			ProjectID: "proj-123",
			ID:        "lb-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "lb-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkLoadBalancerGet_NotFound(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers/lb-999",
		errorResponse(404, "Not Found", "not found"))

	err := NetworkLoadBalancerGet(context.Background(), srv.Client(), NetworkLoadBalancerGetArgs{
		ProjectID: "proj-123",
		ID:        "lb-999",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNetworkLoadBalancerListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"network", "loadbalancer", "list"})
	if err == nil {
		t.Fatal("expected error")
	}
}
