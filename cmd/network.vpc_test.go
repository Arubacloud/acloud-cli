package cmd

import (
	"context"
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
			errContains: "API error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetwork(m)), []string{"network", "vpc", "list", "--project-id", "proj-123"})
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
			errContains: "API error",
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
			err := runCmd(newMockClient(withNetwork(m)), []string{"network", "vpc", "get", "vpc-001", "--project-id", "proj-123"})
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
			errContains: "API error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetwork(m)), tc.args)
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
		})
	}
}

func TestVPCDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes flag",
			setupMock: func(m *mockVPCsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
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
			errContains: "API error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			// --yes skips the interactive confirmation prompt
			err := runCmd(newMockClient(withNetwork(m)), []string{"network", "vpc", "delete", "vpc-001", "--project-id", "proj-123", "--yes"})
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
		})
	}
}
