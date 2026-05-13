package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestVPNTunnelListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNTunnelsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockVPNTunnelsClient) {
				id, name := "tun-001", "my-tunnel"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
					return &types.Response[types.VPNTunnelList]{
						StatusCode: 200,
						Data: &types.VPNTunnelList{
							Values: []types.VPNTunnelResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "tun-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
					return &types.Response[types.VPNTunnelList]{StatusCode: 200, Data: &types.VPNTunnelList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockVPNTunnelsClient) {
				id, name := "tun-001", "my-tunnel"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
					return &types.Response[types.VPNTunnelList]{
						StatusCode: 200,
						Data: &types.VPNTunnelList{
							Values: []types.VPNTunnelResponse{
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
			setupMock: func(m *mockVPNTunnelsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
					return &types.Response[types.VPNTunnelList]{
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
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpntunnel", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNTunnelListCmd_RedactsPSKSecret(t *testing.T) {
	const secret = "super-secret-psk-value"
	id, name := "tun-001", "my-tunnel"
	s := secret
	m := &mockVPNTunnelsClient{
		listFn: func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
			return &types.Response[types.VPNTunnelList]{
				StatusCode: 200,
				Data: &types.VPNTunnelList{
					Values: []types.VPNTunnelResponse{{
						Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
						Properties: types.VPNTunnelPropertiesResponse{
							VPNClientSettings: &types.VPNClientSettings{
								PSK: &types.PSKSettings{Secret: &s},
							},
						},
					}},
				},
			}, nil
		},
	}

	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			out := captureStdout(func() {
				rootCmd.PersistentFlags().Set("output", format)
				_ = runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})),
					[]string{"network", "vpntunnel", "list", "--project-id", "proj-123"})
				rootCmd.PersistentFlags().Set("output", "table")
			})
			if strings.Contains(out, secret) {
				t.Fatalf("PSK secret leaked to --output %s:\n%s", format, out)
			}
		})
	}
}

func TestVPNTunnelGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNTunnelsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockVPNTunnelsClient) {
				id, name := "tun-001", "my-tunnel"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return &types.Response[types.VPNTunnelResponse]{
						StatusCode: 200,
						Data:       &types.VPNTunnelResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "tun-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return &types.Response[types.VPNTunnelResponse]{
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
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})),
				[]string{"network", "vpntunnel", "get", "tun-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNTunnelCreateCmd(t *testing.T) {
	baseArgs := []string{
		"network", "vpntunnel", "create",
		"--project-id", "proj-123",
		"--name", "my-tunnel",
		"--region", "IT-BG",
		"--peer-ip", "1.2.3.4",
		"--vpc-uri", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001",
		"--elastic-ip-uri", "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001",
		"--subnet-cidr", "10.0.1.0/24",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockVPNTunnelsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockVPNTunnelsClient) {
				id, name := "tun-new", "my-tunnel"
				m.createFn = func(_ context.Context, _ string, _ types.VPNTunnelRequest, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return &types.Response[types.VPNTunnelResponse]{
						StatusCode: 200,
						Data:       &types.VPNTunnelResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "tun-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-tunnel"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockVPNTunnelsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.VPNTunnelRequest, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockVPNTunnelsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.VPNTunnelRequest, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return &types.Response[types.VPNTunnelResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name:        "rejects invalid --ike-encryption before API call",
			args:        append(baseArgs, "--ike-encryption", "rot13"),
			wantErr:     true,
			errContains: `"rot13" is not a valid value`,
		},
		{
			name:        "rejects invalid --ike-dpd-action before API call",
			args:        append(baseArgs, "--ike-dpd-action", "explode"),
			wantErr:     true,
			errContains: "is not a valid value",
		},
		{
			name:        "rejects invalid --esp-pfs before API call",
			args:        append(baseArgs, "--esp-pfs", "group99"),
			wantErr:     true,
			errContains: "is not a valid value",
		},
		{
			name: "accepts valid --ike-encryption without API error",
			args: append(baseArgs, "--ike-encryption", types.VPNEncryptionAES256),
			setupMock: func(m *mockVPNTunnelsClient) {
				id, name := "tun-new", "my-tunnel"
				m.createFn = func(_ context.Context, _ string, _ types.VPNTunnelRequest, _ *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
					return &types.Response[types.VPNTunnelResponse]{
						StatusCode: 200,
						Data:       &types.VPNTunnelResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestVPNValidateEnum(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		flag    string
		valid   []string
		wantErr bool
		errMsg  string
	}{
		{name: "empty value passes", value: "", flag: "ike-encryption", valid: vpnEncryptionAlgorithms},
		{name: "valid encryption", value: types.VPNEncryptionAES256, flag: "ike-encryption", valid: vpnEncryptionAlgorithms},
		{name: "valid hash", value: types.VPNHashSHA256, flag: "ike-hash", valid: vpnHashAlgorithms},
		{name: "valid dh-group", value: types.VPNDHGroup14, flag: "ike-dh-group", valid: vpnDHGroups},
		{name: "valid dpd-action", value: types.VPNDPDActionTrap, flag: "ike-dpd-action", valid: vpnDPDActions},
		{name: "valid pfs-group", value: types.VPNPFSDHGroup14, flag: "esp-pfs", valid: vpnPFSGroups},
		{
			name: "invalid encryption rejected", value: "rot13", flag: "ike-encryption", valid: vpnEncryptionAlgorithms,
			wantErr: true, errMsg: `--ike-encryption "rot13" is not a valid value`,
		},
		{
			name: "invalid hash rejected", value: "crc32", flag: "ike-hash", valid: vpnHashAlgorithms,
			wantErr: true, errMsg: "--ike-hash",
		},
		{
			name: "invalid dh-group rejected", value: "42", flag: "ike-dh-group", valid: vpnDHGroups,
			wantErr: true, errMsg: "--ike-dh-group",
		},
		{
			name: "invalid dpd-action rejected", value: "explode", flag: "ike-dpd-action", valid: vpnDPDActions,
			wantErr: true, errMsg: "--ike-dpd-action",
		},
		{
			name: "invalid pfs-group rejected", value: "group99", flag: "esp-pfs", valid: vpnPFSGroups,
			wantErr: true, errMsg: "--esp-pfs",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := vpnValidateEnum(tc.value, tc.flag, tc.valid)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestVPNTunnelDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNTunnelsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "tun-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "tun-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockVPNTunnelsClient) {
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
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"network", "vpntunnel", "delete", "tun-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
