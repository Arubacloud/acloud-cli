package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestContainerRegistryListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockContainerRegistryClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockContainerRegistryClient) {
				id, name := "cr-001", "my-registry"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryList], error) {
					return &types.Response[types.ContainerRegistryList]{
						StatusCode: 200,
						Data: &types.ContainerRegistryList{
							Values: []types.ContainerRegistryResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockContainerRegistryClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryList], error) {
					return &types.Response[types.ContainerRegistryList]{StatusCode: 200, Data: &types.ContainerRegistryList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockContainerRegistryClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockContainerRegistryClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryList], error) {
					return &types.Response[types.ContainerRegistryList]{
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
			m := &mockContainerRegistryClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{containerRegistryClient: m})),
				[]string{"container", "containerregistry", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestContainerRegistryGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockContainerRegistryClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockContainerRegistryClient) {
				id, name := "cr-001", "my-registry"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
					return &types.Response[types.ContainerRegistryResponse]{
						StatusCode: 200,
						Data:       &types.ContainerRegistryResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockContainerRegistryClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockContainerRegistryClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
					return &types.Response[types.ContainerRegistryResponse]{
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
			m := &mockContainerRegistryClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{containerRegistryClient: m})),
				[]string{"container", "containerregistry", "get", "cr-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestContainerRegistryCreateCmd(t *testing.T) {
	baseArgs := []string{
		"container", "containerregistry", "create",
		"--project-id", "proj-123",
		"--name", "my-registry",
		"--region", "IT-BG",
		"--public-ip-uri", "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001",
		"--vpc-uri", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001",
		"--subnet-uri", "/projects/proj-123/providers/Aruba.Network/subnets/sub-001",
		"--security-group-uri", "/projects/proj-123/providers/Aruba.Network/securityGroups/sg-001",
		"--block-storage-uri", "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockContainerRegistryClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockContainerRegistryClient) {
				id, name := "cr-new", "my-registry"
				m.createFn = func(_ context.Context, _ string, _ types.ContainerRegistryRequest, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
					return &types.Response[types.ContainerRegistryResponse]{
						StatusCode: 200,
						Data:       &types.ContainerRegistryResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
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
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockContainerRegistryClient) {
				m.createFn = func(_ context.Context, _ string, _ types.ContainerRegistryRequest, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockContainerRegistryClient) {
				m.createFn = func(_ context.Context, _ string, _ types.ContainerRegistryRequest, _ *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
					return &types.Response[types.ContainerRegistryResponse]{
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
			m := &mockContainerRegistryClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{containerRegistryClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestContainerRegistryDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockContainerRegistryClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockContainerRegistryClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockContainerRegistryClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockContainerRegistryClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
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
			m := &mockContainerRegistryClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{containerRegistryClient: m})),
				[]string{"container", "containerregistry", "delete", "cr-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
