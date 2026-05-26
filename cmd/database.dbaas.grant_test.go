package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

const (
	grantListPath   = "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants"
	grantGetPath    = "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants/grant-001"
	grantCreatePath = "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants"
)

func makeGrantResp() types.GrantResponse {
	return types.GrantResponse{
		User:     types.GrantUser{Username: "myuser"},
		Role:     types.GrantRole{Name: "liteadmin"},
		Database: types.GrantDatabaseResponse{Name: "mydb"},
	}
}

func TestDBaaSGrantListCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "grant", "list", "dbaas-001", "mydb", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantListPath, jsonResponse(200, types.GrantList{
					Values: []types.GrantResponse{makeGrantResp()},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"database", "dbaas", "grant", "list", "dbaas-001", "mydb", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantListPath, jsonResponse(200, types.GrantList{}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "No grants found") {
					t.Errorf("expected 'No grants found', got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "grant", "list", "dbaas-001", "mydb", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantListPath, errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "grant", "list", "dbaas-001", "mydb", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantListPath, errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404)",
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

func TestDBaaSGrantGetCmd(t *testing.T) {
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
				srv.OnGet(grantGetPath, jsonResponse(200, makeGrantResp()))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in output, got: %s", out)
				}
				if !strings.Contains(out, "liteadmin") {
					t.Errorf("expected role in output, got: %s", out)
				}
			},
		},
		{
			name: "success --output json",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantGetPath, jsonResponse(200, makeGrantResp()))
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
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantGetPath, errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting grant",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantGetPath, errorResponse(404, "Not Found", "not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404)",
		},
	}
	getArgs := []string{"database", "dbaas", "grant", "get", "dbaas-001", "mydb", "grant-001", "--project-id", "proj-123"}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			args := getArgs
			if tc.name == "success --output json" {
				args = append(append([]string{}, getArgs...), "--output", "json")
			}
			out, err := runCmdCapture(srv.Client(), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSGrantCreateCmd(t *testing.T) {
	baseArgs := []string{
		"database", "dbaas", "grant", "create", "dbaas-001", "mydb",
		"--project-id", "proj-123",
		"--username", "myuser",
		"--role", "liteadmin",
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
			name: "success with response",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost(grantCreatePath, jsonResponse(200, makeGrantResp()))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name: "success with role and username in output",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				resp := makeGrantResp()
				srv.OnPost(grantCreatePath, jsonResponse(200, resp))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "liteadmin") {
					t.Errorf("expected role in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing --username",
			args:        []string{"database", "dbaas", "grant", "create", "dbaas-001", "mydb", "--project-id", "proj-123", "--role", "liteadmin"},
			wantErr:     true,
			errContains: "username",
		},
		{
			name:        "missing --role",
			args:        []string{"database", "dbaas", "grant", "create", "dbaas-001", "mydb", "--project-id", "proj-123", "--username", "myuser"},
			wantErr:     true,
			errContains: "role",
		},
		{
			name: "server error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost(grantCreatePath, errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "creating grant",
		},
		{
			name: "API error propagates",
			args: baseArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost(grantCreatePath, errorResponse(409, "Conflict", "grant already exists"))
			},
			wantErr:     true,
			errContains: "API error (status 409)",
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

func TestDBaaSGrantDeleteCmd(t *testing.T) {
	deleteArgs := []string{"database", "dbaas", "grant", "delete", "dbaas-001", "mydb", "grant-001", "--project-id", "proj-123", "--yes"}
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
			args: deleteArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete(grantGetPath, jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "grant-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: checks existence, does not delete",
			args: []string{"database", "dbaas", "grant", "delete", "dbaas-001", "mydb", "grant-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantGetPath, jsonResponse(200, makeGrantResp()))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "grant-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: server error propagates",
			args: []string{"database", "dbaas", "grant", "delete", "dbaas-001", "mydb", "grant-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet(grantGetPath, errorResponse(404, "Not Found", "not found"))
			},
			wantErr:     true,
			errContains: "dry-run",
		},
		{
			name: "server error propagates",
			args: deleteArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete(grantGetPath, errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "deleting grant",
		},
		{
			name: "API error propagates",
			args: deleteArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete(grantGetPath, errorResponse(404, "Not Found", "not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404)",
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


func TestGrantRef(t *testing.T) {
	ref := grantRef("proj", "dbaas", "db", "g1")
	_ = ref
}
