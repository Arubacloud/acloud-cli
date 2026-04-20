package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockDBaaSClient) {
				id, name := "dbaas-001", "my-dbaas"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSList], error) {
					return &types.Response[types.DBaaSList]{
						StatusCode: 200,
						Data: &types.DBaaSList{
							Values: []types.DBaaSResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockDBaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSList], error) {
					return &types.Response[types.DBaaSList]{StatusCode: 200, Data: &types.DBaaSList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDBaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDBaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSList], error) {
					return &types.Response[types.DBaaSList]{
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
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})),
				[]string{"database", "dbaas", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestDBaaSGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockDBaaSClient) {
				id, name := "dbaas-001", "my-dbaas"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return &types.Response[types.DBaaSResponse]{
						StatusCode: 200,
						Data:       &types.DBaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDBaaSClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDBaaSClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return &types.Response[types.DBaaSResponse]{
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
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})),
				[]string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestDBaaSCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--engine-id", "postgres14", "--flavor", "db.small"},
			setupMock: func(m *mockDBaaSClient) {
				id, name := "dbaas-new", "my-dbaas"
				m.createFn = func(_ context.Context, _ string, _ types.DBaaSRequest, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return &types.Response[types.DBaaSResponse]{
						StatusCode: 200,
						Data:       &types.DBaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"database", "dbaas", "create", "--project-id", "proj-123", "--region", "IT-BG", "--engine-id", "postgres14", "--flavor", "db.small"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --engine-id",
			args:        []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--flavor", "db.small"},
			wantErr:     true,
			errContains: "engine-id",
		},
		{
			name: "SDK error propagates",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--engine-id", "postgres14", "--flavor", "db.small"},
			setupMock: func(m *mockDBaaSClient) {
				m.createFn = func(_ context.Context, _ string, _ types.DBaaSRequest, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--engine-id", "postgres14", "--flavor", "db.small"},
			setupMock: func(m *mockDBaaSClient) {
				m.createFn = func(_ context.Context, _ string, _ types.DBaaSRequest, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return &types.Response[types.DBaaSResponse]{
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
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestDBaaSDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockDBaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDBaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDBaaSClient) {
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
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})),
				[]string{"database", "dbaas", "delete", "dbaas-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
