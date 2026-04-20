package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSUserListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockUsersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockUsersClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserList], error) {
					return &types.Response[types.UserList]{
						StatusCode: 200,
						Data: &types.UserList{
							Values: []types.UserResponse{
								{Username: "admin"},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockUsersClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserList], error) {
					return &types.Response[types.UserList]{StatusCode: 200, Data: &types.UserList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockUsersClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockUsersClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserList], error) {
					return &types.Response[types.UserList]{
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
			m := &mockUsersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{usersClient: m})),
				[]string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestDBaaSUserGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockUsersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockUsersClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserResponse], error) {
					return &types.Response[types.UserResponse]{
						StatusCode: 200,
						Data:       &types.UserResponse{Username: "admin"},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockUsersClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockUsersClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.UserResponse], error) {
					return &types.Response[types.UserResponse]{
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
			m := &mockUsersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{usersClient: m})),
				[]string{"database", "dbaas", "user", "get", "dbaas-001", "admin", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestDBaaSUserCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockUsersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser", "--password", "Pass1!"},
			setupMock: func(m *mockUsersClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.UserRequest, _ *types.RequestParameters) (*types.Response[types.UserResponse], error) {
					return &types.Response[types.UserResponse]{
						StatusCode: 200,
						Data:       &types.UserResponse{Username: "myuser"},
					}, nil
				}
			},
		},
		{
			name:        "missing required flag --username",
			args:        []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--password", "Pass1!"},
			wantErr:     true,
			errContains: "username",
		},
		{
			name:        "missing required flag --password",
			args:        []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser"},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "SDK error propagates",
			args: []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser", "--password", "Pass1!"},
			setupMock: func(m *mockUsersClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.UserRequest, _ *types.RequestParameters) (*types.Response[types.UserResponse], error) {
					return nil, fmt.Errorf("duplicate user")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser", "--password", "Pass1!"},
			setupMock: func(m *mockUsersClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.UserRequest, _ *types.RequestParameters) (*types.Response[types.UserResponse], error) {
					return &types.Response[types.UserResponse]{
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
			m := &mockUsersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{usersClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestDBaaSUserDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockUsersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockUsersClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockUsersClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockUsersClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
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
			m := &mockUsersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withDatabase(&mockDatabaseClient{usersClient: m})),
				[]string{"database", "dbaas", "user", "delete", "dbaas-001", "myuser", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
