package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSecurityGroupListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockSecurityGroupsClient) {
				id, name := "sg-001", "my-sg"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupList], error) {
					return &types.Response[types.SecurityGroupList]{
						StatusCode: 200,
						Data: &types.SecurityGroupList{
							Values: []types.SecurityGroupResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupList], error) {
					return &types.Response[types.SecurityGroupList]{StatusCode: 200, Data: &types.SecurityGroupList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockSecurityGroupsClient) {
				id, name := "sg-001", "my-sg"
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupList], error) {
					return &types.Response[types.SecurityGroupList]{
						StatusCode: 200,
						Data: &types.SecurityGroupList{
							Values: []types.SecurityGroupResponse{
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
			setupMock: func(m *mockSecurityGroupsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.listFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupList], error) {
					return &types.Response[types.SecurityGroupList]{
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
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityGroupGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockSecurityGroupsClient) {
				id, name := "sg-001", "my-sg"
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
					return &types.Response[types.SecurityGroupResponse]{
						StatusCode: 200,
						Data:       &types.SecurityGroupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.getFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
					return &types.Response[types.SecurityGroupResponse]{
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
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})),
				[]string{"network", "securitygroup", "get", "vpc-001", "sg-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityGroupCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockSecurityGroupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "IT-BG"},
			setupMock: func(m *mockSecurityGroupsClient) {
				id, name := "sg-new", "my-sg"
				m.createFn = func(_ context.Context, _, _ string, _ types.SecurityGroupRequest, _ *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
					return &types.Response[types.SecurityGroupResponse]{
						StatusCode: 200,
						Data:       &types.SecurityGroupResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "SDK error propagates",
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "IT-BG"},
			setupMock: func(m *mockSecurityGroupsClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.SecurityGroupRequest, _ *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "securitygroup", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-sg", "--region", "IT-BG"},
			setupMock: func(m *mockSecurityGroupsClient) {
				m.createFn = func(_ context.Context, _, _ string, _ types.SecurityGroupRequest, _ *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
					return &types.Response[types.SecurityGroupResponse]{
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
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSecurityGroupDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sg-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSecurityGroupsClient) {
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
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "securitygroup", "delete", "vpc-001", "sg-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
