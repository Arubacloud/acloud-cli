package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestKaaSListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-001", "my-cluster"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{
						StatusCode: 200,
						Data: &types.KaaSList{
							Values: []types.KaaSResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockKaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{StatusCode: 200, Data: &types.KaaSList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{kaasClient: m})),
				[]string{"container", "kaas", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKaaSGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-001", "my-cluster"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
						StatusCode: 200,
						Data:       &types.KaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{kaasClient: m})),
				[]string{"container", "kaas", "get", "kaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKaaSCreateCmd(t *testing.T) {
	baseArgs := []string{
		"container", "kaas", "create",
		"--project-id", "proj-123",
		"--name", "my-cluster",
		"--region", "IT-BG",
		"--vpc-uri", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001",
		"--subnet-uri", "/projects/proj-123/providers/Aruba.Network/subnets/sub-001",
		"--node-cidr-address", "10.0.0.0/16",
		"--node-cidr-name", "node-cidr",
		"--security-group-name", "my-sg",
		"--kubernetes-version", "1.28.0",
		"--node-pool-name", "default-pool",
		"--node-pool-nodes", "1",
		"--node-pool-instance", "n1.standard",
		"--node-pool-zone", "itbg1-a",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-new", "my-cluster"
				m.createFn = func(_ context.Context, _ string, _ types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
						StatusCode: 200,
						Data:       &types.KaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
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
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockKaaSClient) {
				m.createFn = func(_ context.Context, _ string, _ types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockKaaSClient) {
				m.createFn = func(_ context.Context, _ string, _ types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{kaasClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKaaSDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockKaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{kaasClient: m})),
				[]string{"container", "kaas", "delete", "kaas-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKaaSConnectCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			// connect calls DownloadKubeconfig; SDK error must propagate before file I/O or kubectl
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.downloadKubeconfigFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
					return nil, fmt.Errorf("unauthorized")
				}
			},
			wantErr:     true,
			errContains: "downloading kubeconfig",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.downloadKubeconfigFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
					return &types.Response[types.KaaSKubeconfigResponse]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{kaasClient: m})),
				[]string{"container", "kaas", "connect", "kaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
