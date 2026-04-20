package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestElasticIPListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockElasticIPsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockElasticIPsClient) {
				id, name := "eip-001", "my-eip"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticList], error) {
					return &types.Response[types.ElasticList]{
						StatusCode: 200,
						Data: &types.ElasticList{
							Values: []types.ElasticIPResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockElasticIPsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticList], error) {
					return &types.Response[types.ElasticList]{StatusCode: 200, Data: &types.ElasticList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockElasticIPsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockElasticIPsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticList], error) {
					return &types.Response[types.ElasticList]{
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
			m := &mockElasticIPsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{elasticIPsMock: m})),
				[]string{"network", "elasticip", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestElasticIPGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockElasticIPsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockElasticIPsClient) {
				id, name := "eip-001", "my-eip"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
					return &types.Response[types.ElasticIPResponse]{
						StatusCode: 200,
						Data:       &types.ElasticIPResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockElasticIPsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockElasticIPsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
					return &types.Response[types.ElasticIPResponse]{
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
			m := &mockElasticIPsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{elasticIPsMock: m})),
				[]string{"network", "elasticip", "get", "eip-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestElasticIPCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockElasticIPsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "IT-BG"},
			setupMock: func(m *mockElasticIPsClient) {
				id, name := "eip-new", "my-eip"
				m.createFn = func(_ context.Context, _ string, _ types.ElasticIPRequest, _ *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
					return &types.Response[types.ElasticIPResponse]{
						StatusCode: 200,
						Data:       &types.ElasticIPResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "elasticip", "create", "--project-id", "proj-123", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "SDK error propagates",
			args: []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "IT-BG"},
			setupMock: func(m *mockElasticIPsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.ElasticIPRequest, _ *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "IT-BG"},
			setupMock: func(m *mockElasticIPsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.ElasticIPRequest, _ *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
					return &types.Response[types.ElasticIPResponse]{
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
			m := &mockElasticIPsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{elasticIPsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestElasticIPDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockElasticIPsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockElasticIPsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockElasticIPsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockElasticIPsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
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
			m := &mockElasticIPsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{elasticIPsMock: m})),
				[]string{"network", "elasticip", "delete", "eip-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
