package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSecurityRuleListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupRulesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				id, name := "rule-001", "my-rule"
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleList], error) {
					return &types.Response[types.SecurityRuleList]{
						StatusCode: 200,
						Data: &types.SecurityRuleList{
							Values: []types.SecurityRuleResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleList], error) {
					return &types.Response[types.SecurityRuleList]{StatusCode: 200, Data: &types.SecurityRuleList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				id, name := "rule-001", "my-rule"
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleList], error) {
					return &types.Response[types.SecurityRuleList]{
						StatusCode: 200,
						Data: &types.SecurityRuleList{
							Values: []types.SecurityRuleResponse{
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
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.listFn = func(_ context.Context, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleList], error) {
					return &types.Response[types.SecurityRuleList]{
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
			m := &mockSecurityGroupRulesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupRules: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityRuleGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupRulesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				id, name := "rule-001", "my-rule"
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
					return &types.Response[types.SecurityRuleResponse]{
						StatusCode: 200,
						Data:       &types.SecurityRuleResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.getFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
					return &types.Response[types.SecurityRuleResponse]{
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
			m := &mockSecurityGroupRulesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupRules: m})),
				[]string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityRuleCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "securityrule", "create", "vpc-001", "sg-001",
		"--project-id", "proj-123",
		"--name", "my-rule",
		"--region", "IT-BG",
		"--direction", "Ingress",
		"--protocol", "TCP",
		"--target-kind", "Ip",
		"--target-value", "0.0.0.0/0",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockSecurityGroupRulesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockSecurityGroupRulesClient) {
				id, name := "rule-new", "my-rule"
				m.createFn = func(_ context.Context, _, _, _ string, _ types.SecurityRuleRequest, _ *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
					return &types.Response[types.SecurityRuleResponse]{
						StatusCode: 200,
						Data:       &types.SecurityRuleResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-rule"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --direction",
			args:        removeFlag(baseArgs, "--direction", "Ingress"),
			wantErr:     true,
			errContains: "direction",
		},
		{
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.SecurityRuleRequest, _ *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
					return nil, fmt.Errorf("validation error")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.createFn = func(_ context.Context, _, _, _ string, _ types.SecurityRuleRequest, _ *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
					return &types.Response[types.SecurityRuleResponse]{
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
			m := &mockSecurityGroupRulesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupRules: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityRuleDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockSecurityGroupRulesClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockSecurityGroupRulesClient) {
				m.deleteFn = func(_ context.Context, _, _, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockSecurityGroupRulesClient) {
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
			m := &mockSecurityGroupRulesClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "securityrule", "delete", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{securityGroupRules: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
