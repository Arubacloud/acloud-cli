package cmd

import (
	"encoding/json"
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerList{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerList{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "loadbalancer", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "lb-001", "my-lb"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerList{
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
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerList{
		Values: []types.LoadBalancerResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.LoadBalancerPropertiesResponse{Address: &addr},
				Status:     types.ResourceStatus{State: &state},
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
		Status:     types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "loadbalancer", "get", "lb-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "lb-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
