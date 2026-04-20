package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSecurityGroupListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupsClient)
		wantErr     bool
		errContains string
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
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})),
				[]string{"network", "securitygroup", "list", "vpc-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestSecurityGroupGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupsClient)
		wantErr     bool
		errContains string
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
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})),
				[]string{"network", "securitygroup", "get", "vpc-001", "sg-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
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
			errContains: "API error",
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
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockSecurityGroupsClient) {
				m.deleteFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
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
			errContains: "API error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockSecurityGroupsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupsMock: m})),
				[]string{"network", "securitygroup", "delete", "vpc-001", "sg-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
