package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestKeyPairListCmd(t *testing.T) {
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
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-001", "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
					Values: []types.KeyPairResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-kp") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{}))
			},
		},
		{
			name: "skips entries with empty ID",
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-001", "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
					Values: []types.KeyPairResponse{
						{Metadata: types.ResourceMetadataResponse{Name: &kpName}},
						{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
					},
				}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-001", "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
					Values: []types.KeyPairResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
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
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(404, "Not Found", "resource not found"))
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

func TestKeyPairGetCmd(t *testing.T) {
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
				kpName := "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", jsonResponse(200, types.KeyPairResponse{
					Metadata: types.ResourceMetadataResponse{Name: &kpName},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-kp") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"compute", "keypair", "get", "kp-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKeyPairCreateCmd(t *testing.T) {
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
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-new", "my-kp"
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairResponse{
					Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-kp") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--public-key", "ssh-rsa AAAA"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --public-key",
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp"},
			wantErr:     true,
			errContains: "public-key",
		},
		{
			name: "server error propagates",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(500, "Internal Server Error", "duplicate name"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(404, "Not Found", "resource not found"))
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

func TestKeyPairUpdateCmd(t *testing.T) {
	// keypairUpdateCmd is a pure stub — no API call is made.
	// Registering no routes means any HTTP call would trip the harness t.Errorf.
	srv := newArubaTestServer(t)
	out, err := runCmdCapture(srv.Client(), []string{"compute", "keypair", "update", "kp-001", "--project-id", "proj-123"})
	checkErr(t, err, false, "")
	if !strings.Contains(out, "not supported") {
		t.Errorf("expected 'not supported' in output, got: %s", out)
	}
}

func TestKeyPairDeleteCmd(t *testing.T) {
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
				srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kp-001") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupSrv: func(srv *arubaTestServer) {
				kpName := "my-kp"
				// Only register Get; an erroneous DELETE would trip the harness t.Errorf.
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", jsonResponse(200, types.KeyPairResponse{
					Metadata: types.ResourceMetadataResponse{Name: &kpName},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kp-001") {
					t.Errorf("expected name in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", errorResponse(500, "Internal Server Error", "not found"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Compute/keyPairs/kp-001", errorResponse(404, "Not Found", "resource not found"))
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
			args := []string{"compute", "keypair", "delete", "kp-001", "--project-id", "proj-123", "--yes"}
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
