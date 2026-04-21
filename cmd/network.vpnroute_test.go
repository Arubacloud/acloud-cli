package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPNRouteListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
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
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})),
				[]string{"network", "vpnroute", "list", "tun-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestVPNRouteGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
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
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})),
				[]string{"network", "vpnroute", "get", "tun-001", "vpnr-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
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
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestVPNRouteDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNRoutesClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPNRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
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
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnRoutesMock: m})),
				[]string{"network", "vpnroute", "delete", "tun-001", "vpnr-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
