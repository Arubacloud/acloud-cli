package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPNRouteListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVPNRoutesClient) {
				id, name := "vpnr-001", "my-vpnroute"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteList], error) {
					return &types.Response[types.VPNRouteList]{
						StatusCode: 200,
						Data: &types.VPNRouteList{
							Values: []types.VPNRouteResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpnr-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockVPNRoutesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteList], error) {
					return &types.Response[types.VPNRouteList]{StatusCode: 200, Data: &types.VPNRouteList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockVPNRoutesClient) {
				id, name := "vpnr-001", "my-vpnroute"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteList], error) {
					return &types.Response[types.VPNRouteList]{
						StatusCode: 200,
						Data: &types.VPNRouteList{
							Values: []types.VPNRouteResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPNRoutesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPNRoutesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteList], error) {
					return &types.Response[types.VPNRouteList]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpnroute", "list", "tun-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNRouteGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockVPNRoutesClient) {
				id, name := "vpnr-001", "my-vpnroute"
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
					return &types.Response[types.VPNRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPNRouteResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpnr-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPNRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPNRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
					return &types.Response[types.VPNRouteResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})),
				[]string{"network", "vpnroute", "get", "tun-001", "vpnr-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNRouteCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpnroute", "create", "tun-001",
		"--project-id", "proj-123",
		"--name", "my-route",
		"--region", "IT-BG",
		"--cloud-subnet", "10.0.0.0/24",
		"--onprem-subnet", "192.168.1.0/24",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockVPNRoutesClient) {
				id, name := "vpnr-new", "my-route"
				m.createFn = func(_ context.Context, _, _ string, _ types.VPNRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
					return &types.Response[types.VPNRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPNRouteResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpnr-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-route"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockVPNRoutesClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.VPNRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockVPNRoutesClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.VPNRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
					return &types.Response[types.VPNRouteResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNRouteDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPNRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpnr-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockVPNRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpnr-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPNRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPNRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpnroute", "delete", "tun-001", "vpnr-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
