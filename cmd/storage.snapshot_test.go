package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSnapshotListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSnapshotsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockSnapshotsClient) {
				id, name := "snap-001", "my-snapshot"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotList], error) {
					return &types.Response[types.SnapshotList]{
						StatusCode: 200,
						Data: &types.SnapshotList{
							Values: []types.SnapshotResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockSnapshotsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotList], error) {
					return &types.Response[types.SnapshotList]{StatusCode: 200, Data: &types.SnapshotList{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSnapshotsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSnapshotsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotList], error) {
					return &types.Response[types.SnapshotList]{
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
			m := &mockSnapshotsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{snapshotsMock: m})),
				[]string{"storage", "snapshot", "list", "--project-id", "proj-123",
					"--volume-uri", "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSnapshotGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSnapshotsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockSnapshotsClient) {
				id, name := "snap-001", "my-snapshot"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
					return &types.Response[types.SnapshotResponse]{
						StatusCode: 200,
						Data:       &types.SnapshotResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSnapshotsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSnapshotsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
					return &types.Response[types.SnapshotResponse]{
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
			m := &mockSnapshotsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{snapshotsMock: m})),
				[]string{"storage", "snapshot", "get", "snap-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSnapshotCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockSnapshotsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{
				"storage", "snapshot", "create",
				"--project-id", "proj-123",
				"--name", "my-snapshot",
				"--region", "IT-BG",
				"--volume-uri", "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001",
			},
			setupMock: func(m *mockSnapshotsClient) {
				id, name := "snap-new", "my-snapshot"
				m.createFn = func(_ context.Context, _ string, _ types.SnapshotRequest, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
					return &types.Response[types.SnapshotResponse]{
						StatusCode: 200,
						Data:       &types.SnapshotResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
		{
			name:    "missing required flag --name",
			args:    []string{"storage", "snapshot", "create", "--project-id", "proj-123", "--region", "IT-BG", "--volume-uri", "/v/vol-001"},
			wantErr: true, errContains: "name",
		},
		{
			name:    "missing required flag --volume-uri",
			args:    []string{"storage", "snapshot", "create", "--project-id", "proj-123", "--name", "my-snapshot", "--region", "IT-BG"},
			wantErr: true, errContains: "volume-uri",
		},
		{
			name: "SDK error propagates",
			args: []string{
				"storage", "snapshot", "create",
				"--project-id", "proj-123",
				"--name", "my-snapshot",
				"--region", "IT-BG",
				"--volume-uri", "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001",
			},
			setupMock: func(m *mockSnapshotsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.SnapshotRequest, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{
				"storage", "snapshot", "create",
				"--project-id", "proj-123",
				"--name", "my-snapshot",
				"--region", "IT-BG",
				"--volume-uri", "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001",
			},
			setupMock: func(m *mockSnapshotsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.SnapshotRequest, _ *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
					return &types.Response[types.SnapshotResponse]{
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
			m := &mockSnapshotsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{snapshotsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSnapshotDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSnapshotsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockSnapshotsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSnapshotsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSnapshotsClient) {
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
			m := &mockSnapshotsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withStorageMock(&mockStorageClient{snapshotsMock: m})),
				[]string{"storage", "snapshot", "delete", "snap-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
