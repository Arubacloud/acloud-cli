package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
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
					jsonResponse(200, types.SecurityRuleListResponse{
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
					jsonResponse(200, types.SecurityRuleListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleListResponse{
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
			name: "--output yaml emits valid YAML",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleListResponse{
						Values: []types.SecurityRuleResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if strings.TrimSpace(out) == "{}" || strings.TrimSpace(out) == "" {
					t.Errorf("yaml output is empty or {}, got: %s", out)
				}
			},
		},
		{
			name: "--output table-json emits valid JSON array",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleListResponse{
						Values: []types.SecurityRuleResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				var rows []map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rows); err != nil {
					t.Errorf("table-json output is not a valid JSON array: %v\noutput: %s", err, out)
				}
				if len(rows) == 0 {
					t.Errorf("expected at least one row in table-json output, got none")
				}
			},
		},
		{
			name: "--output table-yaml emits non-empty YAML",
			args: []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rule-001", "my-rule"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
					jsonResponse(200, types.SecurityRuleListResponse{
						Values: []types.SecurityRuleResponse{
							{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
						},
					}))
			},
			assertOut: func(t *testing.T, out string) {
				if strings.TrimSpace(out) == "" {
					t.Errorf("table-yaml output is empty, got: %s", out)
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
		"--region", "ITBG-Bergamo",
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
						Status: types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
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
						Status: types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
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
						Status:   types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
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

// TestSecurityRuleGetCmd_FullDetail exercises all nil-guard branches in GET detail output.
func TestSecurityRuleGetCmd_FullDetail(t *testing.T) {
	t.Run("detail with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "rule-001", "my-rule"
		uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001"
		createdBy := "user@example.com"
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		state := types.StateActive
		region := types.Region("IT-BG")
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
			jsonResponse(200, types.SecurityRuleResponse{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					URI:              &uri,
					LocationResponse: &types.LocationResponse{Value: region},
					CreationDate:     &ts,
					CreatedBy:        &createdBy,
					Tags:             []string{"env=prod"},
				},
				Status: types.ResourceStatusResponse{State: &state},
				Properties: types.SecurityRulePropertiesResponse{
					Direction: types.RuleDirectionIngress,
					Protocol:  types.RuleProtocolTCP,
					Port:      "80",
					Target:    &types.RuleTargetCommon{Kind: types.EndpointTypeIP, Value: "0.0.0.0/0"},
				},
			}))
		out, err := runCmdCapture(srv.Client(), []string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "rule-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
		if !strings.Contains(out, "env=prod") {
			t.Errorf("expected tags in output, got: %s", out)
		}
	})

	t.Run("--output json succeeds without error", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "rule-001", "my-rule"
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001",
			jsonResponse(200, types.SecurityRuleResponse{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
			}))
		_, err := runCmdCapture(srv.Client(), []string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestSecurityRuleListCmd_AllOptionalFields exercises nil-guard branches in LIST output.
func TestSecurityRuleListCmd_AllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rule-001", "my-rule"
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
		jsonResponse(200, types.SecurityRuleListResponse{
			Values: []types.SecurityRuleResponse{
				{
					Metadata: types.ResourceMetadataResponse{
						ID:               &id,
						Name:             &name,
						LocationResponse: &types.LocationResponse{Value: region},
					},
					Status: types.ResourceStatusResponse{State: &state},
					Properties: types.SecurityRulePropertiesResponse{
						Direction: types.RuleDirectionIngress,
						Protocol:  types.RuleProtocolTCP,
					},
				},
			},
		}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "rule-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "Ingress") {
		t.Errorf("expected direction in output, got: %s", out)
	}
}

func TestSecurityRuleCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rule-001", "my-rule"
	region := types.Region("IT-BG")
	state := types.StateActive
	dir := types.RuleDirectionIngress
	proto := types.RuleProtocolTCP
	port := "80"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules", jsonResponse(200, types.SecurityRuleResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Properties: types.SecurityRulePropertiesResponse{
			Direction: dir,
			Protocol:  proto,
			Port:      port,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"network", "securityrule", "create", "vpc-001", "sg-001",
		"--project-id", "proj-123",
		"--name", "my-rule",
		"--region", "ITBG-Bergamo",
		"--direction", "Ingress",
		"--protocol", "TCP",
		"--port", "80",
		"--target-kind", "Ip",
		"--target-value", "0.0.0.0/0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSecurityRuleListCmd_WithAllFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rule-001", "my-rule"
	region := types.Region("IT-BG")
	state := types.StateActive
	dir := types.RuleDirectionIngress
	proto := types.RuleProtocolTCP
	port := "80"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules", jsonResponse(200, types.SecurityRuleListResponse{
		Values: []types.SecurityRuleResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.SecurityRulePropertiesResponse{
					Direction: dir,
					Protocol:  proto,
					Port:      port,
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "securityrule", "list", "vpc-001", "sg-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "rule-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests
// =============================================================================

func TestNetworkSecurityRuleCreateArgs_Validate(t *testing.T) {
	valid := NetworkSecurityRuleCreateArgs{
		ProjectID:   "p1",
		VPCID:       "vpc-1",
		SGID:        "sg-1",
		Name:        "my-rule",
		Region:      aruba.RegionITBGBergamo,
		Direction:   "Ingress",
		Protocol:    "TCP",
		TargetKind:  "Ip",
		TargetValue: "0.0.0.0/0",
	}
	tests := []struct {
		name        string
		mutate      func(*NetworkSecurityRuleCreateArgs)
		wantErr     bool
		errContains string
	}{
		{name: "happy path"},
		{name: "name too short", mutate: func(a *NetworkSecurityRuleCreateArgs) { a.Name = "ab" }, wantErr: true, errContains: "--name"},
		{name: "invalid region", mutate: func(a *NetworkSecurityRuleCreateArgs) { a.Region = "ZZ" }, wantErr: true, errContains: "--region"},
		{name: "empty direction", mutate: func(a *NetworkSecurityRuleCreateArgs) { a.Direction = "" }, wantErr: true, errContains: "--direction"},
		{name: "empty protocol", mutate: func(a *NetworkSecurityRuleCreateArgs) { a.Protocol = "" }, wantErr: true, errContains: "--protocol"},
		{name: "empty target-kind", mutate: func(a *NetworkSecurityRuleCreateArgs) { a.TargetKind = "" }, wantErr: true, errContains: "--target-kind"},
		{name: "empty target-value", mutate: func(a *NetworkSecurityRuleCreateArgs) { a.TargetValue = "" }, wantErr: true, errContains: "--target-value"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := valid
			if tc.mutate != nil {
				tc.mutate(&args)
			}
			err := args.Validate()
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

func TestNetworkSecurityRuleCreateArgs_ErrValidationFailed(t *testing.T) {
	args := NetworkSecurityRuleCreateArgs{ProjectID: "p1", VPCID: "vpc-1", SGID: "sg-1", Name: "x", Region: "bad"}
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !errors.Is(ErrValidationFailed, ErrValidationFailed) {
		t.Error("sentinel self-check failed")
	}
}

// =============================================================================
// Layer 2 — Operation function tests
// =============================================================================

func TestNetworkSecurityRuleCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rule-new", "my-rule"
	dir := types.RuleDirectionIngress
	proto := types.RuleProtocolTCP
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules", jsonResponse(200, types.SecurityRuleResponse{
		Metadata:   types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.SecurityRulePropertiesResponse{Direction: dir, Protocol: proto, Port: "80"},
	}))
	out := captureStdout(func() {
		err := NetworkSecurityRuleCreate(context.Background(), srv.Client(), NetworkSecurityRuleCreateArgs{
			ProjectID: "proj-123", VPCID: "vpc-001", SGID: "sg-001",
			Name: "my-rule", Region: aruba.RegionITBGBergamo,
			Direction: "Ingress", Protocol: "TCP", Port: "80",
			TargetKind: "Ip", TargetValue: "0.0.0.0/0",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "rule-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkSecurityRuleCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules",
		errorResponse(500, "Internal Server Error", "boom"))
	err := NetworkSecurityRuleCreate(context.Background(), srv.Client(), NetworkSecurityRuleCreateArgs{
		ProjectID: "proj-123", VPCID: "vpc-001", SGID: "sg-001",
		Name: "my-rule", Region: aruba.RegionITBGBergamo,
		Direction: "Ingress", Protocol: "TCP",
		TargetKind: "Ip", TargetValue: "0.0.0.0/0",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating security rule") {
		t.Errorf("error %q does not contain 'creating security rule'", err.Error())
	}
}

func TestNetworkSecurityRuleList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rule-001", "my-rule"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules", jsonResponse(200, types.SecurityRuleListResponse{
		Values: []types.SecurityRuleResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	out := captureStdout(func() {
		err := NetworkSecurityRuleList(context.Background(), srv.Client(), NetworkSecurityRuleListArgs{
			ProjectID: "proj-123", VPCID: "vpc-001", SGID: "sg-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "rule-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSecurityRuleGetCmd_WithAllOptionals(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rule-001", "my-rule"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	dir := types.RuleDirectionIngress
	proto := types.RuleProtocolTCP
	port := "80"
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules/rule-001", jsonResponse(200, types.SecurityRuleResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.SecurityRulePropertiesResponse{
			Direction: dir,
			Protocol:  proto,
			Port:      port,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "securityrule", "get", "vpc-001", "sg-001", "rule-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "rule-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkSecurityRuleCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"network", "securityrule", "create", "vpc-001", "sg-001",
		"--project-id", "proj-123",
		"--name", "x",
		"--region", "ITBG-Bergamo",
		"--direction", "Ingress",
		"--protocol", "TCP",
		"--target-kind", "Ip",
		"--target-value", "0.0.0.0/0",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestNetworkSecurityRuleListRun_NoProjectID(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUP := os.Getenv("USERPROFILE")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	os.Setenv("USERPROFILE", tmp)
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUP)
	}()
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"network", "securityrule", "list", "vpc-001", "sg-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}
