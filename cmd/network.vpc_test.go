package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPCListCmd(t *testing.T) {
	vpcID := "vpc-001"
	vpcName := "my-vpc"

	tests := []struct {
		name        string
		setupMock   func(*mockVPCsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVPCsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPCList], error) {
					return &types.Response[types.VPCList]{
						StatusCode: 200,
						Data: &types.VPCList{
							Values: []types.VPCResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &vpcID, Name: &vpcName}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success with no results",
			setupMock: func(m *mockVPCsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPCList], error) {
					return &types.Response[types.VPCList]{StatusCode: 200, Data: &types.VPCList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockVPCsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPCList], error) {
					return &types.Response[types.VPCList]{
						StatusCode: 200,
						Data: &types.VPCList{
							Values: []types.VPCResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &vpcID, Name: &vpcName}},
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
			setupMock: func(m *mockVPCsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPCList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing VPCs",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPCList], error) {
					return &types.Response[types.VPCList]{
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
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpc", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetwork(m)), args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errContains)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCGetCmd(t *testing.T) {
	vpcID := "vpc-001"
	vpcName := "my-vpc"

	tests := []struct {
		name        string
		setupMock   func(*mockVPCsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockVPCsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return &types.Response[types.VPCResponse]{
						StatusCode: 200,
						Data:       &types.VPCResponse{Metadata: types.ResourceMetadataResponse{ID: &vpcID, Name: &vpcName}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting VPC details",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return &types.Response[types.VPCResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "nil data — not found message",
			setupMock: func(m *mockVPCsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return &types.Response[types.VPCResponse]{StatusCode: 200}, nil
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetwork(m)), []string{"network", "vpc", "get", "vpc-001", "--project-id", "proj-123"})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errContains)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCCreateCmd(t *testing.T) {
	vpcID := "vpc-new"
	vpcName := "new-vpc"

	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPCsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc", "--region", "IT-BG"},
			setupMock: func(m *mockVPCsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.VPCRequest, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return &types.Response[types.VPCResponse]{
						StatusCode: 200,
						Data:       &types.VPCResponse{Metadata: types.ResourceMetadataResponse{ID: &vpcID, Name: &vpcName}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "vpc", "create", "--project-id", "proj-123", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "SDK error propagates",
			args: []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc", "--region", "IT-BG"},
			setupMock: func(m *mockVPCsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.VPCRequest, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating VPC",
		},
		{
			name: "API error propagates",
			args: []string{"network", "vpc", "create", "--project-id", "proj-123", "--name", "new-vpc", "--region", "IT-BG"},
			setupMock: func(m *mockVPCsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.VPCRequest, _ *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
					return &types.Response[types.VPCResponse]{
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
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetwork(m)), tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errContains)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes flag",
			setupMock: func(m *mockVPCsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockVPCsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vpc-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting VPC",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCsClient) {
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
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			// --yes skips the interactive confirmation prompt
			args := []string{"network", "vpc", "delete", "vpc-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetwork(m)), args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errContains)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
