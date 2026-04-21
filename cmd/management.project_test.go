package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestProjectListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockProjectClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockProjectClient) {
				id, name := "proj-001", "my-project"
				m.listFn = func(_ context.Context, _ *types.RequestParameters) (*types.Response[types.ProjectList], error) {
					return &types.Response[types.ProjectList]{
						StatusCode: 200,
						Data: &types.ProjectList{
							Values: []types.ProjectResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockProjectClient) {
				m.listFn = func(_ context.Context, _ *types.RequestParameters) (*types.Response[types.ProjectList], error) {
					return &types.Response[types.ProjectList]{StatusCode: 200, Data: &types.ProjectList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockProjectClient) {
				m.listFn = func(_ context.Context, _ *types.RequestParameters) (*types.Response[types.ProjectList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockProjectClient) {
				m.listFn = func(_ context.Context, _ *types.RequestParameters) (*types.Response[types.ProjectList], error) {
					return &types.Response[types.ProjectList]{
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
			m := &mockProjectClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withProject(m)), []string{"management", "project", "list"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestProjectGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockProjectClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockProjectClient) {
				id, name := "proj-001", "my-project"
				m.getFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
					return &types.Response[types.ProjectResponse]{
						StatusCode: 200,
						Data:       &types.ProjectResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockProjectClient) {
				m.getFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockProjectClient) {
				m.getFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
					return &types.Response[types.ProjectResponse]{
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
			m := &mockProjectClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withProject(m)), []string{"management", "project", "get", "proj-001"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestProjectCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockProjectClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{"management", "project", "create", "--name", "my-project"},
			setupMock: func(m *mockProjectClient) {
				id, name := "proj-new", "my-project"
				m.createFn = func(_ context.Context, _ types.ProjectRequest, _ *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
					return &types.Response[types.ProjectResponse]{
						StatusCode: 200,
						Data:       &types.ProjectResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"management", "project", "create"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: []string{"management", "project", "create", "--name", "my-project"},
			setupMock: func(m *mockProjectClient) {
				m.createFn = func(_ context.Context, _ types.ProjectRequest, _ *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"management", "project", "create", "--name", "my-project"},
			setupMock: func(m *mockProjectClient) {
				m.createFn = func(_ context.Context, _ types.ProjectRequest, _ *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
					return &types.Response[types.ProjectResponse]{
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
			m := &mockProjectClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withProject(m)), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestProjectDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockProjectClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockProjectClient) {
				m.deleteFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockProjectClient) {
				m.deleteFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockProjectClient) {
				m.deleteFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
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
			m := &mockProjectClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withProject(m)), []string{"management", "project", "delete", "proj-001", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
