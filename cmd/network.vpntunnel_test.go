package cmd

import (
	"context"
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
			name: "SDK error propagates",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})),
				[]string{"network", "vpntunnel", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})),
				[]string{"network", "vpntunnel", "get", "tun-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestVPNTunnelDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockVPNTunnelsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockVPNTunnelsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
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
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockVPNTunnelsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withNetworkMock(&mockNetworkClient{vpnTunnelsMock: m})),
				[]string{"network", "vpntunnel", "delete", "tun-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
