package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestKaaSListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-001", "my-cluster"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{
						StatusCode: 200,
						Data: &types.KaaSList{
							Values: []types.KaaSResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockKaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{StatusCode: 200, Data: &types.KaaSList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-001", "my-cluster"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{
						StatusCode: 200,
						Data: &types.KaaSList{
							Values: []types.KaaSResponse{
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
			setupMock: func(m *mockKaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSList], error) {
					return &types.Response[types.KaaSList]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"container", "kaas", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withContainer(&mockContainerClient{kaasClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKaaSGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-001", "my-cluster"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
						StatusCode: 200,
						Data:       &types.KaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withContainer(&mockContainerClient{kaasClient: m})),
				[]string{"container", "kaas", "get", "kaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKaaSCreateCmd(t *testing.T) {
	baseArgs := []string{
		"container", "kaas", "create",
		"--project-id", "proj-123",
		"--name", "my-cluster",
		"--region", "IT-BG",
		"--vpc-uri", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001",
		"--subnet-uri", "/projects/proj-123/providers/Aruba.Network/subnets/sub-001",
		"--node-cidr-address", "10.0.0.0/16",
		"--node-cidr-name", "node-cidr",
		"--security-group-name", "my-sg",
		"--kubernetes-version", "1.28.0",
		"--node-pool-name", "default-pool",
		"--node-pool-nodes", "1",
		"--node-pool-instance", "n1.standard",
		"--node-pool-zone", "itbg1-a",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-new", "my-cluster"
				m.createFn = func(_ context.Context, _ string, _ types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
						StatusCode: 200,
						Data:       &types.KaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-cluster"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockKaaSClient) {
				m.createFn = func(_ context.Context, _ string, _ types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockKaaSClient) {
				m.createFn = func(_ context.Context, _ string, _ types.KaaSRequest, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withContainer(&mockContainerClient{kaasClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKaaSDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockKaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockKaaSClient) {
				id, name := "kaas-001", "my-cluster"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
					return &types.Response[types.KaaSResponse]{StatusCode: 200, Data: &types.KaaSResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}}}, nil
				}
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kaas-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"container", "kaas", "delete", "kaas-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withContainer(&mockContainerClient{kaasClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKaaSConnectCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKaaSClient)
		wantErr     bool
		errContains string
	}{
		{
			// connect calls DownloadKubeconfig; SDK error must propagate before file I/O or kubectl
			name: "SDK error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.downloadKubeconfigFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
					return nil, fmt.Errorf("unauthorized")
				}
			},
			wantErr:     true,
			errContains: "downloading kubeconfig",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKaaSClient) {
				m.downloadKubeconfigFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
					return &types.Response[types.KaaSKubeconfigResponse]{
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
			m := &mockKaaSClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withContainer(&mockContainerClient{kaasClient: m})),
				[]string{"container", "kaas", "connect", "kaas-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
