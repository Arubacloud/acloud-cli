package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSDatabaseListCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
					Values: []types.DatabaseResponse{
						{Name: "my-db"},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
					Values: []types.DatabaseResponse{
						{Name: "my-db"},
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
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
					Values: []types.DatabaseResponse{
						{Name: "my-db"},
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
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
					Values: []types.DatabaseResponse{
						{Name: "my-db"},
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
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
					Values: []types.DatabaseResponse{
						{Name: "my-db"},
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
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSDatabaseGetCmd(t *testing.T) {
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", jsonResponse(200, types.DatabaseResponse{
					Name: "my-db",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "database", "get", "dbaas-001", "my-db", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSDatabaseCreateCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "my-db"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseResponse{
					Name: "my-db",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "my-db"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "my-db"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSDatabaseDeleteCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "database", "delete", "dbaas-001", "my-db", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"database", "dbaas", "database", "delete", "dbaas-001", "my-db", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", jsonResponse(200, types.DatabaseResponse{
					Name: "my-db",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-db") {
					t.Errorf("expected name in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "database", "delete", "dbaas-001", "my-db", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "database", "delete", "dbaas-001", "my-db", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", errorResponse(404, "Not Found", "resource not found"))
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

// TestDBaaSDatabaseGetCmd_FullDetail exercises all nil-guard branches in GET detail output.
func TestDBaaSDatabaseGetCmd_FullDetail(t *testing.T) {
	t.Run("detail with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		createdBy := "user@example.com"
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", jsonResponse(200, types.DatabaseResponse{
			Name:         "my-db",
			CreatedBy:    &createdBy,
			CreationDate: &ts,
		}))
		out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "database", "get", "dbaas-001", "my-db", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "my-db") {
			t.Errorf("expected name in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
	})
	t.Run("--output json emits valid JSON", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/my-db", jsonResponse(200, types.DatabaseResponse{
			Name: "my-db",
		}))
		out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "database", "get", "dbaas-001", "my-db", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
			t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
		}
	})
}

// TestDBaaSDatabaseListCmd_AllOptionalFields exercises all nil-guard branches in LIST output.
func TestDBaaSDatabaseListCmd_AllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	createdBy := "user@example.com"
	ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
		Values: []types.DatabaseResponse{
			{Name: "my-db", CreatedBy: &createdBy, CreationDate: &ts},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "database", "list", "dbaas-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-db") {
		t.Errorf("expected name in output, got: %s", out)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("expected createdBy in output, got: %s", out)
	}
}

func TestDBaaSDatabaseUpdateCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "database", "update", "dbaas-001", "mydb", "--project-id", "proj-123", "--name", "newdb"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb", jsonResponse(200, types.DatabaseResponse{
					Name: "mydb",
				}))
				// SDK uses the NEW name in the PUT path for databases
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/newdb", jsonResponse(200, types.DatabaseResponse{
					Name: "newdb",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "mydb") {
					t.Errorf("expected db name in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing --name",
			args:        []string{"database", "dbaas", "database", "update", "dbaas-001", "mydb", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "pre-GET error",
			args: []string{"database", "dbaas", "database", "update", "dbaas-001", "mydb", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb", errorResponse(404, "Not Found", "not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"database", "dbaas", "database", "update", "dbaas-001", "mydb", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb", jsonResponse(200, types.DatabaseResponse{
					Name: "mydb",
				}))
				// SDK uses the NEW name in the PUT path for databases
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/x", errorResponse(500, "Internal Server Error", "boom"))
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

func TestDatabaseDBaaSDatabaseCreateRun_ValidationError(t *testing.T) {
	// --name "x" is too short (< 3 chars) — triggers ErrValidationFailed
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"database", "dbaas", "database", "create", "dbaas-001", "--project-id", "proj-123", "--name", "x"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestDatabaseDBaaSDatabaseListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"database", "dbaas", "database", "list", "dbaas-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDBaaSDatabaseCreateCmd_WithCreationDate(t *testing.T) {
	srv := newArubaTestServer(t)
	now := time.Now()
	dbName := "myapp_db"
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseResponse{
		Name:         dbName,
		CreationDate: &now,
	}))
	err := runCmd(srv.Client(), []string{
		"database", "dbaas", "database", "create", "dbaas-001",
		"--project-id", "proj-123",
		"--name", "myapp_db",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ─── Layer 1: Validate ───────────────────────────────────────────────────────

func TestDatabaseDBaaSDatabaseCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        DatabaseDBaaSDatabaseCreateArgs
		wantErr     bool
		errContains string
	}{
		{name: "valid", args: DatabaseDBaaSDatabaseCreateArgs{DBaaSID: "d-001", Name: "mydb"}, wantErr: false},
		{name: "missing dbaas-id", args: DatabaseDBaaSDatabaseCreateArgs{Name: "mydb"}, wantErr: true, errContains: "DBaaS ID"},
		{name: "name too short", args: DatabaseDBaaSDatabaseCreateArgs{DBaaSID: "d-001", Name: "ab"}, wantErr: true, errContains: "--name must be at least 3"},
		{name: "name too long", args: DatabaseDBaaSDatabaseCreateArgs{DBaaSID: "d-001", Name: string(make([]byte, 65))}, wantErr: true, errContains: "--name must be at most 64"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
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

func TestDatabaseDBaaSDatabaseGetArgs_Validate(t *testing.T) {
	valid := DatabaseDBaaSDatabaseGetArgs{DBaaSID: "d-001", DatabaseName: "mydb"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	noDB := DatabaseDBaaSDatabaseGetArgs{DatabaseName: "mydb"}
	if err := noDB.Validate(); err == nil {
		t.Fatal("expected error for missing DBaaSID")
	}
	noName := DatabaseDBaaSDatabaseGetArgs{DBaaSID: "d-001"}
	if err := noName.Validate(); err == nil {
		t.Fatal("expected error for missing DatabaseName")
	}
}

func TestDatabaseDBaaSDatabaseUpdateArgs_Validate(t *testing.T) {
	valid := DatabaseDBaaSDatabaseUpdateArgs{DBaaSID: "d-001", DatabaseName: "mydb", Name: "new"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	noDB := DatabaseDBaaSDatabaseUpdateArgs{DatabaseName: "mydb", Name: "new"}
	if err := noDB.Validate(); err == nil {
		t.Fatal("expected error for missing DBaaSID")
	}
	noDBName := DatabaseDBaaSDatabaseUpdateArgs{DBaaSID: "d-001", Name: "new"}
	if err := noDBName.Validate(); err == nil {
		t.Fatal("expected error for missing DatabaseName")
	}
	noNewName := DatabaseDBaaSDatabaseUpdateArgs{DBaaSID: "d-001", DatabaseName: "mydb"}
	if err := noNewName.Validate(); err == nil {
		t.Fatal("expected error for missing Name")
	}
}

func TestDatabaseDBaaSDatabaseDeleteArgs_Validate(t *testing.T) {
	valid := DatabaseDBaaSDatabaseDeleteArgs{DBaaSID: "d-001", DatabaseName: "mydb"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	noDB := DatabaseDBaaSDatabaseDeleteArgs{DatabaseName: "mydb"}
	if err := noDB.Validate(); err == nil {
		t.Fatal("expected error for missing DBaaSID")
	}
	noName := DatabaseDBaaSDatabaseDeleteArgs{DBaaSID: "d-001"}
	if err := noName.Validate(); err == nil {
		t.Fatal("expected error for missing DatabaseName")
	}
}

func TestDatabaseDBaaSDatabaseListArgs_Validate(t *testing.T) {
	valid := DatabaseDBaaSDatabaseListArgs{DBaaSID: "d-001"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	empty := DatabaseDBaaSDatabaseListArgs{}
	if err := empty.Validate(); err == nil {
		t.Fatal("expected error for missing DBaaSID")
	}
}
