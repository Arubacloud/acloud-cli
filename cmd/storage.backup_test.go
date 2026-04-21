package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// storage backup create uses storageBackupCmd directly (no separate "create" subcommand).
// Invocation: storage backup [volume-id] --name <name> ...

// newStorageBackupCreateMock wires a mockStorageClient for the backup create
// flow, which first fetches the volume (to get its URI) then calls
// Backups().Create(). Both sub-clients must be set.
func newStorageBackupCreateMock(backupFn func(context.Context, string, types.StorageBackupRequest, *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)) *mockStorageClient {
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	volumes := &mockVolumesClient{
		getFn: func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
			return &types.Response[types.BlockStorageResponse]{
				StatusCode: 200,
				Data:       &types.BlockStorageResponse{Metadata: types.ResourceMetadataResponse{URI: &volURI}},
			}, nil
		},
	}
	return &mockStorageClient{volumesMock: volumes, backupsMock: &mockStorageBackupsClient{createFn: backupFn}}
}

func TestStorageBackupCreateCmd(t *testing.T) {
	createArgs := []string{"storage", "backup", "vol-001", "--project-id", "proj-123", "--name", "my-backup"}
	tests := []struct {
		name        string
		args        []string
		storageMock *mockStorageClient
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: createArgs,
			storageMock: newStorageBackupCreateMock(func(_ context.Context, _ string, _ types.StorageBackupRequest, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
				id, sname := "bkp-new", "my-backup"
				return &types.Response[types.StorageBackupResponse]{
					StatusCode: 200,
					Data:       &types.StorageBackupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &sname}},
				}, nil
			}),
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "backup", "vol-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: createArgs,
			storageMock: newStorageBackupCreateMock(func(_ context.Context, _ string, _ types.StorageBackupRequest, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
				return nil, fmt.Errorf("quota exceeded")
			}),
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: createArgs,
			storageMock: newStorageBackupCreateMock(func(_ context.Context, _ string, _ types.StorageBackupRequest, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
				return &types.Response[types.StorageBackupResponse]{
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
			err := runCmd(newMockClient(withStorageMock(sm)), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestStorageBackupListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockStorageBackupsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockStorageBackupsClient) {
				id, sname := "bkp-001", "my-backup"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupList], error) {
					return &types.Response[types.StorageBackupList]{
						StatusCode: 200,
						Data: &types.StorageBackupList{
							Values: []types.StorageBackupResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &sname}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockStorageBackupsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupList], error) {
					return &types.Response[types.StorageBackupList]{StatusCode: 200, Data: &types.StorageBackupList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockStorageBackupsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockStorageBackupsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupList], error) {
					return &types.Response[types.StorageBackupList]{
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
			m := &mockStorageBackupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{backupsMock: m})),
				[]string{"storage", "backup", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestStorageBackupGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockStorageBackupsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockStorageBackupsClient) {
				id, sname := "bkp-001", "my-backup"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
					return &types.Response[types.StorageBackupResponse]{
						StatusCode: 200,
						Data:       &types.StorageBackupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &sname}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockStorageBackupsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockStorageBackupsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
					return &types.Response[types.StorageBackupResponse]{
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
			m := &mockStorageBackupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{backupsMock: m})),
				[]string{"storage", "backup", "get", "bkp-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestStorageBackupDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockStorageBackupsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockStorageBackupsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockStorageBackupsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockStorageBackupsClient) {
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
			m := &mockStorageBackupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{backupsMock: m})),
				[]string{"storage", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
