package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// storage restore create uses storageRestoreCmd directly (no separate "create" subcommand).
// Invocation: storage restore [backup-id] [volume-id] --name <name> ...

// newStorageRestoreCreateMock wires a mockStorageClient for the restore create
// flow, which first fetches the backup and volume (to get their URIs) then calls
// Restores().Create(). All three sub-clients must be set.
func newStorageRestoreCreateMock(restoreFn func(context.Context, string, string, types.RestoreRequest, *types.RequestParameters) (*types.Response[types.RestoreResponse], error)) *mockStorageClient {
	bkpURI := "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001"
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	backups := &mockStorageBackupsClient{
		getFn: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
			return &types.Response[types.StorageBackupResponse]{
				StatusCode: 200,
				Data:       &types.StorageBackupResponse{Metadata: types.ResourceMetadataResponse{URI: &bkpURI}},
			}, nil
		},
	}
	volumes := &mockVolumesClient{
		getFn: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
			return &types.Response[types.BlockStorageResponse]{
				StatusCode: 200,
				Data:       &types.BlockStorageResponse{Metadata: types.ResourceMetadataResponse{URI: &volURI}},
			}, nil
		},
	}
	return &mockStorageClient{volumesMock: volumes, backupsMock: backups, restoresMock: &mockStorageRestoreClient{createFn: restoreFn}}
}

func TestStorageRestoreCreateCmd(t *testing.T) {
	createArgs := []string{"storage", "restore", "bkp-001", "vol-001", "--project-id", "proj-123", "--name", "my-restore"}
	tests := []struct {
		name        string
		args        []string
		storageMock *mockStorageClient
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: createArgs,
			storageMock: newStorageRestoreCreateMock(func(_ context.Context, _, _ string, _ types.RestoreRequest, _ *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
				id, rname := "rst-new", "my-restore"
				return &types.Response[types.RestoreResponse]{
					StatusCode: 200,
					Data:       &types.RestoreResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &rname}},
				}, nil
			}),
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "restore", "bkp-001", "vol-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: createArgs,
			storageMock: newStorageRestoreCreateMock(func(_ context.Context, _, _ string, _ types.RestoreRequest, _ *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
				return nil, fmt.Errorf("quota exceeded")
			}),
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: createArgs,
			storageMock: newStorageRestoreCreateMock(func(_ context.Context, _, _ string, _ types.RestoreRequest, _ *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
				return &types.Response[types.RestoreResponse]{
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
			sm := tc.storageMock
			if sm == nil {
				sm = &mockStorageClient{}
			}
			out, err := runCmdCapture(newMockClient(withStorageMock(sm)), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageRestoreListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockStorageRestoreClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockStorageRestoreClient) {
				id, rname := "rst-001", "my-restore"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreList], error) {
					return &types.Response[types.RestoreList]{
						StatusCode: 200,
						Data: &types.RestoreList{
							Values: []types.RestoreResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &rname}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockStorageRestoreClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreList], error) {
					return &types.Response[types.RestoreList]{StatusCode: 200, Data: &types.RestoreList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockStorageRestoreClient) {
				id, rname := "rst-001", "my-restore"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreList], error) {
					return &types.Response[types.RestoreList]{
						StatusCode: 200,
						Data: &types.RestoreList{
							Values: []types.RestoreResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &rname}},
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
			setupMock: func(m *mockStorageRestoreClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockStorageRestoreClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreList], error) {
					return &types.Response[types.RestoreList]{
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
			m := &mockStorageRestoreClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withStorageMock(&mockStorageClient{restoresMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageRestoreGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockStorageRestoreClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockStorageRestoreClient) {
				id, rname := "rst-001", "my-restore"
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
					return &types.Response[types.RestoreResponse]{
						StatusCode: 200,
						Data:       &types.RestoreResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &rname}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockStorageRestoreClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockStorageRestoreClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
					return &types.Response[types.RestoreResponse]{
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
			m := &mockStorageRestoreClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withStorageMock(&mockStorageClient{restoresMock: m})),
				[]string{"storage", "restore", "get", "bkp-001", "rst-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageRestoreDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockStorageRestoreClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockStorageRestoreClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockStorageRestoreClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockStorageRestoreClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockStorageRestoreClient) {
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
			m := &mockStorageRestoreClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"storage", "restore", "delete", "bkp-001", "rst-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withStorageMock(&mockStorageClient{restoresMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
