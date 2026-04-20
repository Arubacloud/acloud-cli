package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestLoadBalancerListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockLoadBalancersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockLoadBalancersClient) {
				id, name := "lb-001", "my-lb"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerList], error) {
					return &types.Response[types.LoadBalancerList]{
						StatusCode: 200,
						Data: &types.LoadBalancerList{
							Values: []types.LoadBalancerResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockLoadBalancersClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerList], error) {
					return &types.Response[types.LoadBalancerList]{StatusCode: 200, Data: &types.LoadBalancerList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockLoadBalancersClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockLoadBalancersClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerList], error) {
					return &types.Response[types.LoadBalancerList]{
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
			m := &mockLoadBalancersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{loadBalancersMock: m})),
				[]string{"network", "loadbalancer", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestLoadBalancerGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockLoadBalancersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockLoadBalancersClient) {
				id, name := "lb-001", "my-lb"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error) {
					return &types.Response[types.LoadBalancerResponse]{
						StatusCode: 200,
						Data:       &types.LoadBalancerResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockLoadBalancersClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockLoadBalancersClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error) {
					return &types.Response[types.LoadBalancerResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error",
		},
		{
			name: "nil data",
			setupMock: func(m *mockLoadBalancersClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error) {
					return &types.Response[types.LoadBalancerResponse]{StatusCode: 200}, nil
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockLoadBalancersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{loadBalancersMock: m})),
				[]string{"network", "loadbalancer", "get", "lb-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
