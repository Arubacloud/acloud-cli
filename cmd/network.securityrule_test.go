package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSecurityRuleListCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleList{
						Values: []types.SecurityRuleResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleList{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleList{
						Values: []types.SecurityRuleResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
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
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
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
		"--target-value", "10.0.0.0/8",
	}
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-new", "my-rule"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
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
			name:        "missing required flag --protocol",
			args:        removeFlag(baseArgs, "--protocol", "TCP"),
			wantErr:     true,
			errContains: "protocol",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API 404 propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSecurityRuleUpdateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"network", "securityrule", "update", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{
							ID:               &id,
							Name:             &name,
							LocationResponse: &types.LocationResponse{Value: region},
						},
						Status: types.ResourceStatus{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				updID, updName := "rule-001", "new-name"
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{ID: &updID, Name: &updName},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no fields provided returns error",
			args:        []string{"network", "securityrule", "update", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"},
			setupSrv:    nil,
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "server error on update propagates",
			args: []string{"network", "securityrule", "update", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{
							ID:               &id,
							Name:             &name,
							LocationResponse: &types.LocationResponse{Value: region},
						},
						Status: types.ResourceStatus{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
		},
		{
			name: "API 404 on get propagates",
			args: []string{"network", "securityrule", "update", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update falls back to VPC region when rule has no region",
			args: []string{"network", "securityrule", "update", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Status:   types.ResourceStatus{State: func() *types.State { s := types.StateActive; return &s }()},
					}))
				vpcID, vpcName := "vpc-001", "my-vpc"
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001",
					jsonResponse(200, types.VPCResponse{
						Metadata: types.ResourceMetadataResponse{
							ID:               &vpcID,
							Name:             &vpcName,
							LocationResponse: &types.LocationResponse{Value: region},
						},
					}))
				updID, updName := "rule-001", "new-name"
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{ID: &updID, Name: &updName},
					}))
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
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
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			args: []string{"network", "securityrule", "delete", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run registers only GET",
			args: []string{"network", "securityrule", "delete", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					jsonResponse(200, types.SecurityRuleResponse{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rule-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "securityrule", "delete", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API 404 propagates",
			args: []string{"network", "securityrule", "delete", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
					errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
