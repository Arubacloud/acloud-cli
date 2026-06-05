package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

const (
	grantListPath   = "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants"
	grantGetPath    = "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants/grant-001"
	grantCreatePath = "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants"
)

func makeGrantResp() types.GrantResponse {
	return types.GrantResponse{
		User:     types.GrantUserCommon{Username: "myuser"},
		Role:     types.GrantRoleCommon{Name: "liteadmin"},
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
				srv.OnGet(grantListPath, jsonResponse(200, types.GrantListResponse{
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
				srv.OnGet(grantListPath, jsonResponse(200, types.GrantListResponse{}))
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

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validDatabaseDBaaSGrantCreateArgs() DatabaseDBaaSGrantCreateArgs {
	return DatabaseDBaaSGrantCreateArgs{
		ProjectID:    "proj-123",
		DBaaSID:      "dbaas-001",
		DatabaseName: "mydb",
		Username:     "myuser",
		Role:         "liteadmin",
	}
}

func TestDatabaseDBaaSGrantCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*DatabaseDBaaSGrantCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "empty DBaaS ID",
			mutate:      func(a *DatabaseDBaaSGrantCreateArgs) { a.DBaaSID = "" },
			wantErr:     true,
			errContains: "DBaaS ID",
		},
		{
			name:        "empty database name",
			mutate:      func(a *DatabaseDBaaSGrantCreateArgs) { a.DatabaseName = "" },
			wantErr:     true,
			errContains: "database name",
		},
		{
			name:        "empty username",
			mutate:      func(a *DatabaseDBaaSGrantCreateArgs) { a.Username = "" },
			wantErr:     true,
			errContains: "--username",
		},
		{
			name:        "empty role",
			mutate:      func(a *DatabaseDBaaSGrantCreateArgs) { a.Role = "" },
			wantErr:     true,
			errContains: "--role",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validDatabaseDBaaSGrantCreateArgs()
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

func TestDatabaseDBaaSGrantListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := DatabaseDBaaSGrantListArgs{ProjectID: "p1", DBaaSID: "dbaas-001", DatabaseName: "mydb"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty DBaaS ID", func(t *testing.T) {
		args := DatabaseDBaaSGrantListArgs{ProjectID: "p1", DBaaSID: "", DatabaseName: "mydb"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty DBaaS ID")
		}
	})
	t.Run("empty database name", func(t *testing.T) {
		args := DatabaseDBaaSGrantListArgs{ProjectID: "p1", DBaaSID: "dbaas-001", DatabaseName: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty database name")
		}
	})
}

func TestDatabaseDBaaSGrantGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := DatabaseDBaaSGrantGetArgs{ProjectID: "p1", DBaaSID: "dbaas-001", DatabaseName: "mydb", GrantID: "grant-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty grant ID", func(t *testing.T) {
		args := DatabaseDBaaSGrantGetArgs{ProjectID: "p1", DBaaSID: "dbaas-001", DatabaseName: "mydb", GrantID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty grant ID")
		}
	})
	t.Run("empty DBaaS ID", func(t *testing.T) {
		args := DatabaseDBaaSGrantGetArgs{ProjectID: "p1", DBaaSID: "", DatabaseName: "mydb", GrantID: "grant-001"}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty DBaaS ID")
		}
	})
}

func TestDatabaseDBaaSGrantDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := DatabaseDBaaSGrantDeleteArgs{ProjectID: "p1", DBaaSID: "dbaas-001", DatabaseName: "mydb", GrantID: "grant-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty grant ID", func(t *testing.T) {
		args := DatabaseDBaaSGrantDeleteArgs{ProjectID: "p1", DBaaSID: "dbaas-001", DatabaseName: "mydb", GrantID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty grant ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestDatabaseDBaaSGrantCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost(grantCreatePath, jsonResponse(200, makeGrantResp()))

	out := captureStdout(func() {
		err := DatabaseDBaaSGrantCreate(context.Background(), srv.Client(), validDatabaseDBaaSGrantCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "myuser") {
		t.Errorf("expected username in output, got: %s", out)
	}
}

func TestDatabaseDBaaSGrantCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost(grantCreatePath, errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := DatabaseDBaaSGrantCreate(context.Background(), srv.Client(), validDatabaseDBaaSGrantCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating grant") {
		t.Errorf("error %q does not contain 'creating grant'", err.Error())
	}
}

func TestDatabaseDBaaSGrantCreate_AsyncResponse(t *testing.T) {
	srv := newArubaTestServer(t)
	// Async: server returns 202 with a real (non-nil) response body.
	srv.OnPost(grantCreatePath, jsonResponse(202, makeGrantResp()))

	out := captureStdout(func() {
		err := DatabaseDBaaSGrantCreate(context.Background(), srv.Client(), validDatabaseDBaaSGrantCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "myuser") {
		t.Errorf("expected username in async output, got: %s", out)
	}
}

func TestDatabaseDBaaSGrantList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet(grantListPath, jsonResponse(200, types.GrantListResponse{
		Values: []types.GrantResponse{makeGrantResp()},
	}))

	out := captureStdout(func() {
		err := DatabaseDBaaSGrantList(context.Background(), srv.Client(), DatabaseDBaaSGrantListArgs{
			ProjectID:    "proj-123",
			DBaaSID:      "dbaas-001",
			DatabaseName: "mydb",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "myuser") {
		t.Errorf("expected username in output, got: %s", out)
	}
}

func TestDatabaseDBaaSGrantList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet(grantListPath, jsonResponse(200, types.GrantListResponse{}))

	out := captureStdout(func() {
		err := DatabaseDBaaSGrantList(context.Background(), srv.Client(), DatabaseDBaaSGrantListArgs{
			ProjectID:    "proj-123",
			DBaaSID:      "dbaas-001",
			DatabaseName: "mydb",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No grants found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestDatabaseDBaaSGrantGet_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet(grantGetPath, jsonResponse(200, makeGrantResp()))

	out := captureStdout(func() {
		err := DatabaseDBaaSGrantGet(context.Background(), srv.Client(), DatabaseDBaaSGrantGetArgs{
			ProjectID:    "proj-123",
			DBaaSID:      "dbaas-001",
			DatabaseName: "mydb",
			GrantID:      "grant-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "myuser") {
		t.Errorf("expected username in output, got: %s", out)
	}
	if !strings.Contains(out, "liteadmin") {
		t.Errorf("expected role in output, got: %s", out)
	}
}

func TestDatabaseDBaaSGrantGet_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet(grantGetPath, errorResponse(404, "Not Found", "not found"))

	err := DatabaseDBaaSGrantGet(context.Background(), srv.Client(), DatabaseDBaaSGrantGetArgs{
		ProjectID:    "proj-123",
		DBaaSID:      "dbaas-001",
		DatabaseName: "mydb",
		GrantID:      "grant-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "getting grant") {
		t.Errorf("error %q does not contain 'getting grant'", err.Error())
	}
}

func TestDatabaseDBaaSGrantDelete_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete(grantGetPath, jsonResponse(200, nil))

	out := captureStdout(func() {
		err := DatabaseDBaaSGrantDelete(context.Background(), srv.Client(), DatabaseDBaaSGrantDeleteArgs{
			ProjectID:    "proj-123",
			DBaaSID:      "dbaas-001",
			DatabaseName: "mydb",
			GrantID:      "grant-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "grant-001") {
		t.Errorf("expected grant ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSGrantDelete_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete(grantGetPath, errorResponse(500, "Internal Server Error", "boom"))

	err := DatabaseDBaaSGrantDelete(context.Background(), srv.Client(), DatabaseDBaaSGrantDeleteArgs{
		ProjectID:    "proj-123",
		DBaaSID:      "dbaas-001",
		DatabaseName: "mydb",
		GrantID:      "grant-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "deleting grant") {
		t.Errorf("error %q does not contain 'deleting grant'", err.Error())
	}
}

// TestDatabaseDBaaSGrantCreateArgs_ConstructorWrapsErrors verifies that the
// constructor wraps validation failures with ErrValidationFailed.
func TestDatabaseDBaaSGrantCreateArgs_ConstructorWrapsErrors(t *testing.T) {
	cmd := dbaasGrantCreateCmd
	resetCmdFlags(cmd)
	// Provide required flags but leave DBaaSID empty via no positional args
	_ = cmd.Flags().Set("project-id", "proj-123")
	_ = cmd.Flags().Set("username", "myuser")
	_ = cmd.Flags().Set("role", "liteadmin")

	_, err := NewDatabaseDBaaSGrantCreateArgsFromCobraCommand(cmd, []string{})
	if err == nil {
		t.Fatal("expected validation error for empty DBaaS ID")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected ErrValidationFailed wrapping, got: %v", err)
	}
}

// TestDatabaseDBaaSGrantList_WithCallOpts verifies that CallOpts are threaded through.
func TestDatabaseDBaaSGrantList_WithCallOpts(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet(grantListPath, jsonResponse(200, types.GrantListResponse{
		Values: []types.GrantResponse{makeGrantResp()},
	}))

	err := DatabaseDBaaSGrantList(context.Background(), srv.Client(), DatabaseDBaaSGrantListArgs{
		ProjectID:    "proj-123",
		DBaaSID:      "dbaas-001",
		DatabaseName: "mydb",
		CallOpts:     []aruba.CallOption{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
