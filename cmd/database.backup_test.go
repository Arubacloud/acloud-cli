package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBBackupListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBBackupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockDBBackupsClient) {
				id, name := "bkp-001", "my-backup"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BackupList], error) {
					return &types.Response[types.BackupList]{
						StatusCode: 200,
						Data: &types.BackupList{
							Values: []types.BackupResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockDBBackupsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BackupList], error) {
					return &types.Response[types.BackupList]{StatusCode: 200, Data: &types.BackupList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockDBBackupsClient) {
				id, name := "bkp-001", "my-backup"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BackupList], error) {
					return &types.Response[types.BackupList]{
						StatusCode: 200,
						Data: &types.BackupList{
							Values: []types.BackupResponse{
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
			setupMock: func(m *mockDBBackupsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BackupList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDBBackupsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BackupList], error) {
					return &types.Response[types.BackupList]{
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
			m := &mockDBBackupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"database", "backup", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{backupsClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBBackupGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBBackupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockDBBackupsClient) {
				id, name := "bkp-001", "my-backup"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
					return &types.Response[types.BackupResponse]{
						StatusCode: 200,
						Data:       &types.BackupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDBBackupsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDBBackupsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
					return &types.Response[types.BackupResponse]{
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
			m := &mockDBBackupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{backupsClient: m})),
				[]string{"database", "backup", "get", "bkp-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

// newDBBackupCreateMock returns a mockDatabaseClient wired for backup create.
// The backup create command first fetches the DBaaS instance and database, then
// calls Backups().Create(). All three sub-clients must be set.
func newDBBackupCreateMock(backupsFn func(context.Context, string, types.BackupRequest, *types.RequestParameters) (*types.Response[types.BackupResponse], error)) *mockDatabaseClient {
	dbaasURI := "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001"
	dbaas := &mockDBaaSClient{
		getFn: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
			return &types.Response[types.DBaaSResponse]{
				StatusCode: 200,
				Data:       &types.DBaaSResponse{Metadata: types.ResourceMetadataResponse{URI: &dbaasURI}},
			}, nil
		},
	}
	databases := &mockDatabasesClient{
		getFn: func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
			return &types.Response[types.DatabaseResponse]{StatusCode: 200, Data: &types.DatabaseResponse{Name: "mydb"}}, nil
		},
	}
	backups := &mockDBBackupsClient{createFn: backupsFn}
	return &mockDatabaseClient{dbaasClient: dbaas, databasesClient: databases, backupsClient: backups}
}

func TestDBBackupCreateCmd(t *testing.T) {
	createArgs := []string{"database", "backup", "create", "--project-id", "proj-123", "--name", "my-backup", "--region", "IT-BG", "--dbaas-id", "dbaas-001", "--database-name", "mydb"}
	tests := []struct {
		name        string
		args        []string
		dbMock      *mockDatabaseClient
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: createArgs,
			dbMock: newDBBackupCreateMock(func(_ context.Context, _ string, _ types.BackupRequest, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
				id, bname := "bkp-new", "my-backup"
				return &types.Response[types.BackupResponse]{
					StatusCode: 200,
					Data:       &types.BackupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &bname}},
				}, nil
			}),
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"database", "backup", "create", "--project-id", "proj-123", "--region", "IT-BG", "--dbaas-id", "dbaas-001", "--database-name", "mydb"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --dbaas-id",
			args:        []string{"database", "backup", "create", "--project-id", "proj-123", "--name", "my-backup", "--region", "IT-BG", "--database-name", "mydb"},
			wantErr:     true,
			errContains: "dbaas-id",
		},
		{
			name: "SDK error propagates",
			args: createArgs,
			dbMock: newDBBackupCreateMock(func(_ context.Context, _ string, _ types.BackupRequest, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
				return nil, fmt.Errorf("quota exceeded")
			}),
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: createArgs,
			dbMock: newDBBackupCreateMock(func(_ context.Context, _ string, _ types.BackupRequest, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
				return &types.Response[types.BackupResponse]{
					StatusCode: 404,
					Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
				}, nil
			}),
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dbMock := tc.dbMock
			if dbMock == nil {
				dbMock = &mockDatabaseClient{}
			}
			out, err := runCmdCapture(newMockClient(withDatabase(dbMock)), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBBackupDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockDBBackupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockDBBackupsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockDBBackupsClient) {
				id, name := "bkp-001", "my-backup"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
					return &types.Response[types.BackupResponse]{StatusCode: 200, Data: &types.BackupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}}}, nil
				}
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockDBBackupsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockDBBackupsClient) {
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
			m := &mockDBBackupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"database", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withDatabase(&mockDatabaseClient{backupsClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
