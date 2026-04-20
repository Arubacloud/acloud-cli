package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPCPeeringRouteListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{
						StatusCode: 200,
						Data: &types.VPCPeeringRouteList{
							Values: []types.VPCPeeringRouteResponse{{}},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{StatusCode: 200, Data: &types.VPCPeeringRouteList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})),
				[]string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestVPCPeeringRouteGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})),
				[]string{"network", "vpcpeeringroute", "get", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestVPCPeeringRouteCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{},
					}, nil
				}
			},
		},
		{
			name:    "missing required flag --name",
			args:    []string{"network", "vpcpeeringroute", "create", "vpc-001", "peer-001", "--project-id", "proj-123", "--local-network", "10.0.0.0/24", "--remote-network", "10.1.0.0/24"},
			wantErr: true, errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestVPCPeeringRouteDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})),
				[]string{"network", "vpcpeeringroute", "delete", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
