package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func TestCloudServerListCmd(t *testing.T) {
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
			args: []string{"compute", "cloudserver", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{
					Values: []types.CloudServerResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"compute", "cloudserver", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{}))
			},
		},
		{
			name: "skips entries with nil ID",
			args: []string{"compute", "cloudserver", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{
					Values: []types.CloudServerResponse{
						{Metadata: types.ResourceMetadataResponse{Name: &name}},
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"compute", "cloudserver", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{
					Values: []types.CloudServerResponse{
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
			args: []string{"compute", "cloudserver", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"compute", "cloudserver", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", errorResponse(404, "Not Found", "resource not found"))
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

func TestCloudServerGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"compute", "cloudserver", "get", "cs-001", "--project-id", "proj-123"})
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
		"--region", "ITBG-Bergamo",
		"--zone", "ITBG-1",
		"--flavor", "CSO4A8",
		"--boot-disk-id", "vol-001",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--security-group-id", "sg-001",
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
				id, name := "cs-new", "my-cs"
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
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
			args:        removeFlag(baseArgs, "--region", "ITBG-Bergamo"),
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", errorResponse(404, "Not Found", "resource not found"))
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

func TestCloudServerUpdateCmd(t *testing.T) {
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
			args: []string{"compute", "cloudserver", "update", "cs-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing flags",
			args:        []string{"compute", "cloudserver", "update", "cs-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-Get API error",
			args: []string{"compute", "cloudserver", "update", "cs-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "Update server error",
			args: []string{"compute", "cloudserver", "update", "cs-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "updating",
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

func TestCloudServerDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				// Only register Get; an erroneous DELETE would trip the harness t.Errorf.
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "cs-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", errorResponse(404, "Not Found", "resource not found"))
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
			args := []string{"compute", "cloudserver", "delete", "cs-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(srv.Client(), args)
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
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweron", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweron", errorResponse(500, "Internal Server Error", "server busy"))
			},
			wantErr:     true,
			errContains: "power",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweron", errorResponse(404, "Not Found", "resource not found"))
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
			err := runCmd(srv.Client(), []string{"compute", "cloudserver", "power-on", "cs-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestCloudServerPowerOffCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweroff", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweroff", errorResponse(500, "Internal Server Error", "server busy"))
			},
			wantErr:     true,
			errContains: "power",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweroff", errorResponse(404, "Not Found", "resource not found"))
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
			err := runCmd(srv.Client(), []string{"compute", "cloudserver", "power-off", "cs-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestCloudServerSetPasswordCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/password", jsonResponse(200, nil))
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/password", errorResponse(500, "Internal Server Error", "invalid password"))
			},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "cs-001", "my-server"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/password", errorResponse(404, "Not Found", "resource not found"))
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
			err := runCmd(srv.Client(), []string{"compute", "cloudserver", "set-password", "cs-001", "--project-id", "proj-123", "--password", "Pass1!"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestCompleteCloudServerID(t *testing.T) {
	id, name := "cs-001", "test-server"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{
		Values: []types.CloudServerResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	cmd := cloudserverGetCmd
	if err := cmd.Flags().Set("project-id", "proj-123"); err != nil {
		cmd.Flags().String("project-id", "proj-123", "")
	}

	completions, directive := completeCloudServerID(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
	if len(completions) != 1 {
		t.Errorf("expected 1 completion, got %d: %v", len(completions), completions)
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

func TestCloudServerCreateCmd_WithElasticIPAndKeypair(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "server-001", "my-server"
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	err := runCmd(srv.Client(), []string{
		"compute", "cloudserver", "create",
		"--project-id", "proj-123",
		"--name", "my-server",
		"--region", "ITBG-Bergamo",
		"--zone", "ITBG-1",
		"--flavor", "CSO4A8",
		"--boot-disk-id", "vol-001",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--security-group-id", "sg-001",
		"--elasticip-id", "eip-001",
		"--keypair-id", "kp-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloudServerCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "server-001", "my-server"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Properties: types.CloudServerPropertiesResponse{
			Flavor: types.CloudServerFlavorResponse{
				Name: types.CloudServerFlavor("my-flavor"),
				CPU:  2,
				RAM:  4096,
				HD:   50,
			},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"compute", "cloudserver", "create",
		"--project-id", "proj-123",
		"--name", "my-server",
		"--region", "ITBG-Bergamo",
		"--zone", "ITBG-1",
		"--flavor", "CSO4A8",
		"--boot-disk-id", "vol-001",
		"--vpc-id", "vpc-001",
		"--subnet-id", "sub-001",
		"--security-group-id", "sg-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloudServerGetCmd_JSONOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "server-001", "my-server"
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/server-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"compute", "cloudserver", "get", "server-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestCloudServerGetCmd_WithFullDetailFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
			Tags:             []string{"env=test"},
		},
		Properties: types.CloudServerPropertiesResponse{
			Flavor: types.CloudServerFlavorResponse{
				Name: types.CloudServerFlavor("my-flavor"),
				CPU:  2,
				RAM:  4,
				HD:   50,
			},
			BootVolume: types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Storage/blockStorages/bs-001"},
			KeyPair:    types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Compute/keypairs/kp-001"},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"compute", "cloudserver", "get", "cs-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() exhaustive tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validCreateArgs() ComputeCloudServerCreateArgs {
	return ComputeCloudServerCreateArgs{
		ProjectID:        "proj-123",
		Name:             "my-server",
		Region:           aruba.RegionITBGBergamo,
		Zone:             aruba.ZoneITBG1,
		Flavor:           aruba.CloudServerFlavorCSO4A8,
		BootDiskID:       "vol-001",
		VPCID:            "vpc-001",
		SubnetIDs:        []string{"sub-001"},
		SecurityGroupIDs: []string{"sg-001"},
		BillingPeriod:    aruba.BillingPeriodHour,
	}
}

func TestComputeCloudServerCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*ComputeCloudServerCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		// Required field checks
		{
			name:        "name too short",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *ComputeCloudServerCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *ComputeCloudServerCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		// Enum checks
		{
			name:        "invalid region",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.Region = "ZZ" },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid zone",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.Zone = "nowhere-1" },
			wantErr:     true,
			errContains: "--zone",
		},
		{
			name:        "invalid flavor",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.Flavor = "m1.small" },
			wantErr:     true,
			errContains: "--flavor",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.BillingPeriod = "Weekly" },
			wantErr:     true,
			errContains: "--billing-period",
		},
		// Required ID fields
		{
			name:        "empty vpc-id",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.VPCID = "" },
			wantErr:     true,
			errContains: "--vpc-id is required",
		},
		{
			name:        "empty boot-disk-id",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.BootDiskID = "" },
			wantErr:     true,
			errContains: "--boot-disk-id is required",
		},
		// Slice cardinality
		{
			name:        "empty subnet-ids",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.SubnetIDs = nil },
			wantErr:     true,
			errContains: "--subnet-id requires at least one value",
		},
		{
			name:        "empty security-group-ids",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.SecurityGroupIDs = nil },
			wantErr:     true,
			errContains: "--security-group-id requires at least one value",
		},
		// UserDataFile: non-empty non-existent path
		{
			name:        "user-data-file not found",
			mutate:      func(a *ComputeCloudServerCreateArgs) { a.UserDataFile = "/nonexistent/path/userdata.yaml" },
			wantErr:     true,
			errContains: "--user-data-file",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validCreateArgs()
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

func TestComputeCloudServerGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := ComputeCloudServerGetArgs{ProjectID: "p1", ID: "cs-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := ComputeCloudServerGetArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

func TestComputeCloudServerUpdateArgs_Validate(t *testing.T) {
	t.Run("happy path with name", func(t *testing.T) {
		args := ComputeCloudServerUpdateArgs{ProjectID: "p1", ID: "cs-001", Name: "new-name"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("happy path with tags changed", func(t *testing.T) {
		args := ComputeCloudServerUpdateArgs{ProjectID: "p1", ID: "cs-001", TagsChanged: true}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("no name and no tags", func(t *testing.T) {
		args := ComputeCloudServerUpdateArgs{ProjectID: "p1", ID: "cs-001"}
		err := args.Validate()
		if err == nil {
			t.Fatal("expected error when no flags provided")
		}
		if !strings.Contains(err.Error(), "at least one") {
			t.Errorf("error %q does not contain 'at least one'", err.Error())
		}
	})
}

func TestComputeCloudServerDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := ComputeCloudServerDeleteArgs{ProjectID: "p1", ID: "cs-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := ComputeCloudServerDeleteArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

func TestComputeCloudServerSetPasswordArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := ComputeCloudServerSetPasswordArgs{ProjectID: "p1", ID: "cs-001", Password: "s3cr3t"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty password", func(t *testing.T) {
		args := ComputeCloudServerSetPasswordArgs{ProjectID: "p1", ID: "cs-001", Password: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty password")
		}
	})
}

func TestComputeCloudServerConnectArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := ComputeCloudServerConnectArgs{ProjectID: "p1", ID: "cs-001", User: "ubuntu"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty user", func(t *testing.T) {
		args := ComputeCloudServerConnectArgs{ProjectID: "p1", ID: "cs-001", User: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty user")
		}
	})
	t.Run("placeholder user", func(t *testing.T) {
		args := ComputeCloudServerConnectArgs{ProjectID: "p1", ID: "cs-001", User: "<user>"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for placeholder user")
		}
	})
}

// constructor wraps Validate() errors with ErrValidationFailed
func TestNewComputeCloudServerCreateArgsFromCobraCommand_ValidationError(t *testing.T) {
	cmd := cloudserverCreateCmd
	// Provide all required flags but an invalid region
	_ = cmd.Flags().Set("project-id", "proj-123")
	_ = cmd.Flags().Set("name", "my-server")
	_ = cmd.Flags().Set("region", "INVALID")
	_ = cmd.Flags().Set("zone", "ITBG-1")
	_ = cmd.Flags().Set("flavor", "CSO4A8")
	_ = cmd.Flags().Set("boot-disk-id", "vol-001")
	_ = cmd.Flags().Set("vpc-id", "vpc-001")
	_ = cmd.Flags().Set("subnet-id", "sub-001")
	_ = cmd.Flags().Set("security-group-id", "sg-001")

	_, err := NewComputeCloudServerCreateArgsFromCobraCommand(cmd)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed, got: %v", err)
	}
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestComputeCloudServerCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-new", "my-server"
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerCreate(context.Background(), srv.Client(), validCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeCloudServerCreate_AsyncResponse(t *testing.T) {
	srv := newArubaTestServer(t)
	// Simulate async creation: 202 with a response that has metadata ID and name set
	// but no flavor/region so it falls through to the msgCreatedAsync branch.
	id, name := "cs-new", "my-server"
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(202, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerCreate(context.Background(), srv.Client(), validCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Either sync output (with ID) or async message (with name) should appear
	if !strings.Contains(out, "cs-new") && !strings.Contains(out, "my-server") {
		t.Errorf("expected ID or name in output, got: %s", out)
	}
}

func TestComputeCloudServerCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := ComputeCloudServerCreate(context.Background(), srv.Client(), validCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating cloud server") {
		t.Errorf("error %q does not contain 'creating cloud server'", err.Error())
	}
}

func TestComputeCloudServerCreate_WithUserDataFile(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-new", "my-server"
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	// Write a temp user-data file
	udPath := t.TempDir() + "/userdata.yaml"
	if err := os.WriteFile(udPath, []byte("#cloud-config\n"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	args := validCreateArgs()
	args.UserDataFile = udPath
	err := ComputeCloudServerCreate(context.Background(), srv.Client(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestComputeCloudServerGet_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerGet(context.Background(), srv.Client(), ComputeCloudServerGetArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeCloudServerGet_NotFound(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-999",
		errorResponse(404, "Not Found", "not found"))

	err := ComputeCloudServerGet(context.Background(), srv.Client(), ComputeCloudServerGetArgs{
		ProjectID: "proj-123",
		ID:        "cs-999",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestComputeCloudServerList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{
		Values: []types.CloudServerResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerList(context.Background(), srv.Client(), ComputeCloudServerListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeCloudServerList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{}))

	out := captureStdout(func() {
		err := ComputeCloudServerList(context.Background(), srv.Client(), ComputeCloudServerListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No cloud servers found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestComputeCloudServerDelete_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, nil))

	out := captureStdout(func() {
		err := ComputeCloudServerDelete(context.Background(), srv.Client(), ComputeCloudServerDeleteArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in delete message, got: %s", out)
	}
}

func TestComputeCloudServerDelete_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
		errorResponse(500, "Internal Server Error", "boom"))

	err := ComputeCloudServerDelete(context.Background(), srv.Client(), ComputeCloudServerDeleteArgs{
		ProjectID: "proj-123",
		ID:        "cs-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestComputeCloudServerUpdate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	newName := "updated-server"
	srv.OnPut("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &newName},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerUpdate(context.Background(), srv.Client(), ComputeCloudServerUpdateArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
			Name:      "updated-server",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeCloudServerUpdate_PreGetError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
		errorResponse(404, "Not Found", "not found"))

	err := ComputeCloudServerUpdate(context.Background(), srv.Client(), ComputeCloudServerUpdateArgs{
		ProjectID: "proj-123",
		ID:        "cs-001",
		Name:      "new-name",
	})
	if err == nil {
		t.Fatal("expected error on pre-GET failure")
	}
}

func TestComputeCloudServerConnect_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	eipID := "eip-001"
	eipURI := "/projects/proj-123/providers/Aruba.Network/elasticIps/" + eipID
	addr := "203.0.113.1"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.CloudServerPropertiesResponse{
			LinkedResources: []types.LinkedResourceCommon{{URI: eipURI}},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
		Metadata:   types.ResourceMetadataResponse{ID: &eipID},
		Properties: types.ElasticIPPropertiesResponse{Address: &addr},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerConnect(context.Background(), srv.Client(), ComputeCloudServerConnectArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
			User:      "ubuntu",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "203.0.113.1") {
		t.Errorf("expected IP in output, got: %s", out)
	}
}

func TestComputeCloudServerConnect_NoElasticIP(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerConnect(context.Background(), srv.Client(), ComputeCloudServerConnectArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
			User:      "ubuntu",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No Elastic IP") {
		t.Errorf("expected 'No Elastic IP' message, got: %s", out)
	}
}

func TestComputeCloudServerPowerOn_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweron", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerPowerOn(context.Background(), srv.Client(), ComputeCloudServerPowerOnArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeCloudServerPowerOff_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweroff", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	out := captureStdout(func() {
		err := ComputeCloudServerPowerOff(context.Background(), srv.Client(), ComputeCloudServerPowerOffArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeCloudServerSetPassword_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/password", jsonResponse(200, nil))

	out := captureStdout(func() {
		err := ComputeCloudServerSetPassword(context.Background(), srv.Client(), ComputeCloudServerSetPasswordArgs{
			ProjectID: "proj-123",
			ID:        "cs-001",
			Password:  "s3cr3t",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestCloudServerPowerOnCmd_Success(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweron", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"compute", "cloudserver", "power-on", "cs-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestCloudServerPowerOffCmd_Success(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001/poweroff", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"compute", "cloudserver", "power-off", "cs-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestCloudServerConnectCmd_Success(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	eipID := "eip-001"
	eipURI := "/projects/proj-123/providers/Aruba.Network/elasticIps/" + eipID
	addr := "203.0.113.1"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.CloudServerPropertiesResponse{
			LinkedResources: []types.LinkedResourceCommon{
				{URI: eipURI},
			},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
		Metadata:   types.ResourceMetadataResponse{ID: &eipID},
		Properties: types.ElasticIPPropertiesResponse{Address: &addr},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"compute", "cloudserver", "connect", "cs-001",
		"--project-id", "proj-123",
		"--user", "ubuntu",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "203.0.113.1") {
		t.Errorf("expected IP in output, got: %s", out)
	}
}

func TestCloudServerUpdateCmd_Success(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cs-001", "my-server"
	newName := "updated-server"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPut("/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001", jsonResponse(200, types.CloudServerResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &newName},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"compute", "cloudserver", "update", "cs-001",
		"--project-id", "proj-123",
		"--name", "updated-server",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "cs-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
