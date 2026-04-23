package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSDatabaseListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDatabasesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockDatabasesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseList], error) {
					return &types.Response[types.DatabaseList]{
						StatusCode: 200,
						Data: &types.DatabaseList{
							Values: []types.DatabaseResponse{
								{Name: "my-db"},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockDatabasesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseList], error) {
					return &types.Response[types.DatabaseList]{StatusCode: 200, Data: &types.DatabaseList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockDatabasesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseList], error) {
					return &types.Response[types.DatabaseList]{
						StatusCode: 200,
						Data: &types.DatabaseList{
							Values: []types.DatabaseResponse{
								{Name: "my-db"},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDatabasesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDatabasesClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseList], error) {
					return &types.Response[types.DatabaseList]{
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
			m := &mockDatabasesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{databasesClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSDatabaseGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDatabasesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockDatabasesClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return &types.Response[types.DatabaseResponse]{
						StatusCode: 200,
						Data:       &types.DatabaseResponse{Name: "my-db"},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDatabasesClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDatabasesClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return &types.Response[types.DatabaseResponse]{
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
			m := &mockDatabasesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{databasesClient: m})),
				[]string{"database", "dbaas", "database", "get", "dbaas-001", "my-db", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSDatabaseCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockDatabasesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "my-db"},
			setupMock: func(m *mockDatabasesClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.DatabaseRequest, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return &types.Response[types.DatabaseResponse]{
						StatusCode: 200,
						Data:       &types.DatabaseResponse{Name: "my-db"},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "my-db"},
			setupMock: func(m *mockDatabasesClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.DatabaseRequest, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "my-db"},
			setupMock: func(m *mockDatabasesClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.DatabaseRequest, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return &types.Response[types.DatabaseResponse]{
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
			m := &mockDatabasesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{databasesClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSDatabaseDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDatabasesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockDatabasesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockDatabasesClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
					return &types.Response[types.DatabaseResponse]{StatusCode: 200, Data: &types.DatabaseResponse{Name: "my-db"}}, nil
				}
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDatabasesClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDatabasesClient) {
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
			m := &mockDatabasesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"database", "dbaas", "database", "delete", "dbaas-001", "my-db", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{databasesClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
