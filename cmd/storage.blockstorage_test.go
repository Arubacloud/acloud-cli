package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestBlockStorageListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVolumesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVolumesClient) {
				id, name := "vol-001", "my-volume"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageList], error) {
					return &types.Response[types.BlockStorageList]{
						StatusCode: 200,
						Data: &types.BlockStorageList{
							Values: []types.BlockStorageResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockVolumesClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageList], error) {
					return &types.Response[types.BlockStorageList]{StatusCode: 200, Data: &types.BlockStorageList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockVolumesClient) {
				id, name := "vol-001", "my-volume"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageList], error) {
					return &types.Response[types.BlockStorageList]{
						StatusCode: 200,
						Data: &types.BlockStorageList{
							Values: []types.BlockStorageResponse{
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
			setupMock: func(m *mockVolumesClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVolumesClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageList], error) {
					return &types.Response[types.BlockStorageList]{
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
			m := &mockVolumesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"storage", "blockstorage", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withStorage(m)), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestBlockStorageGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVolumesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockVolumesClient) {
				id, name := "vol-001", "my-volume"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
					return &types.Response[types.BlockStorageResponse]{
						StatusCode: 200,
						Data:       &types.BlockStorageResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVolumesClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVolumesClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
					return &types.Response[types.BlockStorageResponse]{
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
			m := &mockVolumesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withStorage(m)),
				[]string{"storage", "blockstorage", "get", "vol-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestBlockStorageCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVolumesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupMock: func(m *mockVolumesClient) {
				id, name := "vol-new", "my-vol"
				m.createFn = func(_ context.Context, _ string, _ types.BlockStorageRequest, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
					return &types.Response[types.BlockStorageResponse]{
						StatusCode: 200,
						Data:       &types.BlockStorageResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--region", "ITBG-Bergamo", "--size", "10"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --size",
			args:        []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "size",
		},
		{
			name: "SDK error propagates",
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupMock: func(m *mockVolumesClient) {
				m.createFn = func(_ context.Context, _ string, _ types.BlockStorageRequest, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupMock: func(m *mockVolumesClient) {
				m.createFn = func(_ context.Context, _ string, _ types.BlockStorageRequest, _ *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
					return &types.Response[types.BlockStorageResponse]{
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
			m := &mockVolumesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withStorage(m)), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestBlockStorageDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVolumesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVolumesClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockVolumesClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVolumesClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVolumesClient) {
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
			m := &mockVolumesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"storage", "blockstorage", "delete", "vol-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withStorage(m)), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
