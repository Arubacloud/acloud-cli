package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPCPeeringListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVPCPeeringsClient) {
				id, name := "peer-001", "my-peering"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringList], error) {
					return &types.Response[types.VPCPeeringList]{
						StatusCode: 200,
						Data: &types.VPCPeeringList{
							Values: []types.VPCPeeringResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringList], error) {
					return &types.Response[types.VPCPeeringList]{StatusCode: 200, Data: &types.VPCPeeringList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockVPCPeeringsClient) {
				id, name := "peer-001", "my-peering"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringList], error) {
					return &types.Response[types.VPCPeeringList]{
						StatusCode: 200,
						Data: &types.VPCPeeringList{
							Values: []types.VPCPeeringResponse{
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
			setupMock: func(m *mockVPCPeeringsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringList], error) {
					return &types.Response[types.VPCPeeringList]{
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
			m := &mockVPCPeeringsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpcpeering", "list", "vpc-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringsMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockVPCPeeringsClient) {
				id, name := "peer-001", "my-peering"
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
					return &types.Response[types.VPCPeeringResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
					return &types.Response[types.VPCPeeringResponse]{
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
			m := &mockVPCPeeringsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringsMock: m})),
				[]string{"network", "vpcpeering", "get", "vpc-001", "peer-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPCPeeringsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "vpcpeering", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-peering", "--peer-vpc-id", "vpc-002", "--region", "IT-BG"},
			setupMock: func(m *mockVPCPeeringsClient) {
				id, name := "peer-new", "my-peering"
				m.createFn = func(_ context.Context, _, _ string, _ types.VPCPeeringRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
					return &types.Response[types.VPCPeeringResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "vpcpeering", "create", "vpc-001", "--project-id", "proj-123", "--peer-vpc-id", "vpc-002", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: []string{"network", "vpcpeering", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-peering", "--peer-vpc-id", "vpc-002", "--region", "IT-BG"},
			setupMock: func(m *mockVPCPeeringsClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.VPCPeeringRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "vpcpeering", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-peering", "--peer-vpc-id", "vpc-002", "--region", "IT-BG"},
			setupMock: func(m *mockVPCPeeringsClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.VPCPeeringRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
					return &types.Response[types.VPCPeeringResponse]{
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
			m := &mockVPCPeeringsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "peer-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringsClient) {
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
			m := &mockVPCPeeringsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpcpeering", "delete", "vpc-001", "peer-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringsMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
