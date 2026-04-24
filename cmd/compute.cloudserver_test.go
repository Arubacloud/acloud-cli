package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func TestCloudServerListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockCloudServersClient) {
				id, name := "cs-001", "my-server"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
					return &types.Response[types.CloudServerList]{
						StatusCode: 200,
						Data: &types.CloudServerList{
							Values: []types.CloudServerResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockCloudServersClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
					return &types.Response[types.CloudServerList]{StatusCode: 200, Data: &types.CloudServerList{}}, nil
				}
			},
		},
		{
			name: "skips entries with nil ID",
			setupMock: func(m *mockCloudServersClient) {
				id, name := "cs-001", "my-server"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
					return &types.Response[types.CloudServerList]{
						StatusCode: 200,
						Data: &types.CloudServerList{
							Values: []types.CloudServerResponse{
								{Metadata: types.ResourceMetadataResponse{Name: &name}},
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockCloudServersClient) {
				id, name := "cs-001", "my-server"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
					return &types.Response[types.CloudServerList]{
						StatusCode: 200,
						Data: &types.CloudServerList{
							Values: []types.CloudServerResponse{
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
			setupMock: func(m *mockCloudServersClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
					return &types.Response[types.CloudServerList]{
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"compute", "cloudserver", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withCompute(m)), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestCloudServerGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockCloudServersClient) {
				id, name := "cs-001", "my-server"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{
						StatusCode: 200,
						Data:       &types.CloudServerResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withCompute(m)), []string{"compute", "cloudserver", "get", "cs-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestCloudServerCreateCmd(t *testing.T) {
	baseArgs := []string{
		"compute", "cloudserver", "create",
		"--project-id", "proj-123",
		"--name", "my-cs",
		"--region", "IT-BG",
		"--zone", "itbg1-a",
		"--flavor", "m1.small",
		"--image", "img-001",
		"--boot-disk-uri", "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001",
		"--vpc-uri", "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001",
		"--subnet-uri", "/projects/proj-123/providers/Aruba.Network/subnets/sub-001",
		"--security-group-uri", "/projects/proj-123/providers/Aruba.Network/securityGroups/sg-001",
	}
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: baseArgs,
			setupMock: func(m *mockCloudServersClient) {
				id, name := "cs-new", "my-cs"
				m.createFn = func(_ context.Context, _ string, _ types.CloudServerRequest, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{
						StatusCode: 200,
						Data:       &types.CloudServerResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseArgs, "--name", "my-cs"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        removeFlag(baseArgs, "--region", "IT-BG"),
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "SDK error propagates",
			args: baseArgs,
			setupMock: func(m *mockCloudServersClient) {
				m.createFn = func(_ context.Context, _ string, _ types.CloudServerRequest, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupMock: func(m *mockCloudServersClient) {
				m.createFn = func(_ context.Context, _ string, _ types.CloudServerRequest, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withCompute(m)), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestCloudServerDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockCloudServersClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockCloudServersClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockCloudServersClient) {
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"compute", "cloudserver", "delete", "cs-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withCompute(m)), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestCloudServerPowerOnCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockCloudServersClient) {
				m.powerOnFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.powerOnFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return nil, fmt.Errorf("server busy")
				}
			},
			wantErr:     true,
			errContains: "power",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.powerOnFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withCompute(m)), []string{"compute", "cloudserver", "power-on", "cs-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestCloudServerPowerOffCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockCloudServersClient) {
				m.powerOffFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.powerOffFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return nil, fmt.Errorf("server busy")
				}
			},
			wantErr:     true,
			errContains: "power",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.powerOffFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
					return &types.Response[types.CloudServerResponse]{
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withCompute(m)), []string{"compute", "cloudserver", "power-off", "cs-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestCloudServerSetPasswordCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockCloudServersClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockCloudServersClient) {
				m.setPasswordFn = func(_ context.Context, _, _ string, _ types.CloudServerPasswordRequest, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.setPasswordFn = func(_ context.Context, _, _ string, _ types.CloudServerPasswordRequest, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("invalid password")
				}
			},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockCloudServersClient) {
				m.setPasswordFn = func(_ context.Context, _, _ string, _ types.CloudServerPasswordRequest, _ *types.RequestParameters) (*types.Response[any], error) {
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
			m := &mockCloudServersClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withCompute(m)), []string{"compute", "cloudserver", "set-password", "cs-001", "--project-id", "proj-123", "--password", "Pass1!"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

// checkErr is a helper shared across test files in this package.
func checkErr(t *testing.T, err error, wantErr bool, errContains string) {
	t.Helper()
	if wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if errContains != "" && !strings.Contains(err.Error(), errContains) {
			t.Errorf("error %q does not contain %q", err.Error(), errContains)
		}
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExtractIDFromURI(t *testing.T) {
	cases := []struct {
		uri  string
		want string
	}{
		{"/projects/p1/providers/Aruba.Compute/cloudServers/cs-123", "cs-123"},
		{"cs-abc", "cs-abc"},
		{"", ""},
	}
	for _, c := range cases {
		got := extractIDFromURI(c.uri)
		if got != c.want {
			t.Errorf("extractIDFromURI(%q) = %q, want %q", c.uri, got, c.want)
		}
	}
}

func TestCompleteCloudServerID(t *testing.T) {
	id := "cs-001"
	name := "test-server"
	m := &mockCloudServersClient{
		listFn: func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
			return &types.Response[types.CloudServerList]{
				StatusCode: 200,
				Data: &types.CloudServerList{
					Values: []types.CloudServerResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				},
			}, nil
		},
	}
	setClientForTesting(newMockClient(withCompute(m)))
	defer resetClientState()

	cmd := cloudserverCmd
	_ = cmd.Flags().Lookup("project-id") // ensure flag exists
	if cmd.Flags().Lookup("project-id") == nil {
		cmd.Flags().String("project-id", "proj-123", "")
	} else {
		_ = cmd.Flags().Set("project-id", "proj-123")
	}

	completions, directive := completeCloudServerID(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d: %v", len(completions), completions)
	}
}

// removeFlag removes a flag and its value from an args slice.
func removeFlag(args []string, flag, value string) []string {
	out := make([]string, 0, len(args))
	skip := false
	for _, a := range args {
		if skip {
			skip = false
			continue
		}
		if a == flag {
			skip = true
			continue
		}
		out = append(out, a)
	}
	_ = value
	return out
}
