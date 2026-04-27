package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPCPeeringRouteListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{
						StatusCode: 200,
						Data: &types.VPCPeeringRouteList{
							Values: []types.VPCPeeringRouteResponse{{}},
						},
					}, nil
				}
			},
		},
		{
			name: "success with non-nil name and id",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					name := "my-route"
					id := "route-abc"
					return &types.Response[types.VPCPeeringRouteList]{
						StatusCode: 200,
						Data: &types.VPCPeeringRouteList{
							Values: []types.VPCPeeringRouteResponse{{
								Metadata: types.ResourceMetadataResponse{Name: &name, ID: &id},
							}},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{StatusCode: 200, Data: &types.VPCPeeringRouteList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{
						StatusCode: 200,
						Data: &types.VPCPeeringRouteList{
							Values: []types.VPCPeeringRouteResponse{{}},
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
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
					return &types.Response[types.VPCPeeringRouteList]{
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
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpcpeeringroute", "list", "vpc-001", "peer-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringRouteGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{},
					}, nil
				}
			},
		},
		{
			name: "success with non-nil name",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					name := "my-route"
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{Metadata: types.ResourceMetadataResponse{Name: &name}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
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
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})),
				[]string{"network", "vpcpeeringroute", "get", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringRouteCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{},
					}, nil
				}
			},
		},
		{
			name: "success with non-nil metadata in response",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					name := "my-route"
					id := "route-abc"
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{Metadata: types.ResourceMetadataResponse{Name: &name, ID: &id}},
					}, nil
				}
			},
		},
		{
			name:    "missing required flag --name",
			args:    []string{"network", "vpcpeeringroute", "create", "vpc-001", "peer-001", "--project-id", "proj-123", "--local-network", "10.0.0.0/24", "--remote-network", "10.1.0.0/24"},
			wantErr: true, errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{
				"network", "vpcpeeringroute", "create", "vpc-001", "peer-001",
				"--project-id", "proj-123",
				"--name", "my-route",
				"--local-network", "10.0.0.0/24",
				"--remote-network", "10.1.0.0/24",
			},
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
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
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringRouteUpdateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpcpeeringroute", "update", "vpc-001", "peer-001", "route-001",
		"--project-id", "proj-123",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: append(baseArgs, "--name", "updated-route"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.updateFn = func(_ context.Context, _, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{},
					}, nil
				}
			},
		},
		{
			name: "success with non-nil metadata in response",
			args: append(baseArgs, "--name", "updated-route"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					currentName := "old-route"
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{Metadata: types.ResourceMetadataResponse{Name: &currentName}},
					}, nil
				}
				m.updateFn = func(_ context.Context, _, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					name := "updated-route"
					id := "route-001"
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{Metadata: types.ResourceMetadataResponse{Name: &name, ID: &id}},
					}, nil
				}
			},
		},
		{
			name: "name falls back to current when not provided",
			args: append(baseArgs, "--tags", "t1"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					currentName := "existing-name"
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{Metadata: types.ResourceMetadataResponse{Name: &currentName}},
					}, nil
				}
			},
		},
		{
			name:        "no fields provided",
			args:        baseArgs,
			wantErr:     true,
			errContains: "at least one field",
		},
		{
			name: "InCreation blocks update",
			args: append(baseArgs, "--name", "new-name"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					state := StateInCreation
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 200,
						Data:       &types.VPCPeeringRouteResponse{Status: types.ResourceStatus{State: &state}},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "InCreation",
		},
		{
			name: "SDK error on get",
			args: append(baseArgs, "--name", "new-name"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "fetching current",
		},
		{
			name: "SDK error on update",
			args: append(baseArgs, "--name", "new-name"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.updateFn = func(_ context.Context, _, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return nil, fmt.Errorf("timeout")
				}
			},
			wantErr:     true,
			errContains: "updating",
		},
		{
			name: "API error on update",
			args: append(baseArgs, "--name", "new-name"),
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.updateFn = func(_ context.Context, _, _, _, _ string, _ types.VPCPeeringRouteRequest, _ *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
					return &types.Response[types.VPCPeeringRouteResponse]{
						StatusCode: 422,
						Error:      &types.ErrorResponse{Title: strPtr("Unprocessable"), Detail: strPtr("invalid CIDR")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 422): Unprocessable",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPCPeeringRouteDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPCPeeringRoutesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "route-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPCPeeringRoutesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
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
			m := &mockVPCPeeringRoutesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpcpeeringroute", "delete", "vpc-001", "peer-001", "route-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpcPeeringRoutesMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
