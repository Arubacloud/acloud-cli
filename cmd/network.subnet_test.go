package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSubnetListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSubnetsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockSubnetsClient) {
				id, name := "sub-001", "my-subnet"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetList], error) {
					return &types.Response[types.SubnetList]{
						StatusCode: 200,
						Data: &types.SubnetList{
							Values: []types.SubnetResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockSubnetsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetList], error) {
					return &types.Response[types.SubnetList]{StatusCode: 200, Data: &types.SubnetList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSubnetsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSubnetsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetList], error) {
					return &types.Response[types.SubnetList]{
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
			m := &mockSubnetsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{subnetsMock: m})),
				[]string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSubnetGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSubnetsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockSubnetsClient) {
				id, name := "sub-001", "my-subnet"
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
					return &types.Response[types.SubnetResponse]{
						StatusCode: 200,
						Data:       &types.SubnetResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSubnetsClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSubnetsClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
					return &types.Response[types.SubnetResponse]{
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
			m := &mockSubnetsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{subnetsMock: m})),
				[]string{"network", "subnet", "get", "vpc-001", "sub-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSubnetCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockSubnetsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet", "--region", "IT-BG"},
			setupMock: func(m *mockSubnetsClient) {
				id, name := "sub-new", "my-subnet"
				m.createFn = func(_ context.Context, _, _ string, _ types.SubnetRequest, _ *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
					return &types.Response[types.SubnetResponse]{
						StatusCode: 200,
						Data:       &types.SubnetResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "SDK error propagates",
			args: []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet", "--region", "IT-BG"},
			setupMock: func(m *mockSubnetsClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.SubnetRequest, _ *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet", "--region", "IT-BG"},
			setupMock: func(m *mockSubnetsClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.SubnetRequest, _ *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
					return &types.Response[types.SubnetResponse]{
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
			m := &mockSubnetsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{subnetsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSubnetDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSubnetsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockSubnetsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSubnetsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSubnetsClient) {
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
			m := &mockSubnetsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{subnetsMock: m})),
				[]string{"network", "subnet", "delete", "vpc-001", "sub-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
