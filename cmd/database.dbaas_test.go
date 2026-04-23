package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
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
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
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
			name: "--output=json emits valid JSON",
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
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
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
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"database", "dbaas", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
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
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
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
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})),
				[]string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
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
		assertOut   func(*testing.T, string)
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
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-new") {
					t.Errorf("expected ID in output, got: %s", out)
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
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockDBaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockDBaaSClient) {
				id, name := "dbaas-001", "my-dbaas"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
					return &types.Response[types.DBaaSResponse]{StatusCode: 200, Data: &types.DBaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}}}, nil
				}
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
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
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDBaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"database", "dbaas", "delete", "dbaas-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{dbaasClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
