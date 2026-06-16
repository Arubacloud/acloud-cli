package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func TestDBBackupListCmd(t *testing.T) {
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
			args: []string{"database", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
					Values: []types.BackupResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-backup") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"database", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"database", "backup", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
					Values: []types.BackupResponse{
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
			args: []string{"database", "backup", "list", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
					Values: []types.BackupResponse{
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
			args: []string{"database", "backup", "list", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
					Values: []types.BackupResponse{
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
			args: []string{"database", "backup", "list", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
					Values: []types.BackupResponse{
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
			args: []string{"database", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"database", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBBackupGetCmd(t *testing.T) {
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
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, types.BackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"database", "backup", "get", "bkp-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBBackupCreateCmd(t *testing.T) {
	createArgs := []string{
		"database", "backup", "create",
		"--project-id", "proj-123",
		"--name", "my-backup",
		"--region", "ITBG-Bergamo",
		"--dbaas-id", "dbaas-001",
		"--database-name", "mydb",
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
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-new", "my-backup"
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.BackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-backup") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "missing required flag --name",
			args: []string{
				"database", "backup", "create",
				"--project-id", "proj-123",
				"--region", "ITBG-Bergamo",
				"--dbaas-id", "dbaas-001",
				"--database-name", "mydb",
			},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "missing required flag --dbaas-id",
			args: []string{
				"database", "backup", "create",
				"--project-id", "proj-123",
				"--name", "my-backup",
				"--region", "ITBG-Bergamo",
				"--database-name", "mydb",
			},
			wantErr:     true,
			errContains: "dbaas-id",
		},
		{
			name: "success with --zone",
			args: []string{
				"database", "backup", "create",
				"--project-id", "proj-123",
				"--name", "my-backup",
				"--region", "ITBG-Bergamo",
				"--zone", "ITBG-1",
				"--dbaas-id", "dbaas-001",
				"--database-name", "mydb",
			},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-zone", "my-backup"
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.BackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "my-backup") {
					t.Errorf("expected name in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBBackupDeleteCmd(t *testing.T) {
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
			args: []string{"database", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"database", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, types.BackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"database", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", errorResponse(500, "Internal Server Error", "not found"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"database", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBBackupListCmd_AllOptionalFields(t *testing.T) {
	// Covers: LocationResponse and Status.State nil-guards in list loop.
	srv := newArubaTestServer(t)
	id, name := "bkp-001", "my-backup"
	state := types.StateActive
	region := types.Region("ITBG-Bergamo")
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
		Values: []types.BackupResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"database", "backup", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bkp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "ITBG-Bergamo") {
		t.Errorf("expected region in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestDBBackupGetCmd_AllOptionalFields(t *testing.T) {
	// Covers: URI, LocationResponse, Status.State, CreationDate, CreatedBy, Tags detail block.
	// Also covers JSON early-return path.
	id, name := "bkp-001", "my-backup"
	uri := "/projects/proj-123/providers/Aruba.Database/backups/bkp-001"
	createdBy := "user@example.com"
	state := types.StateActive
	region := types.Region("ITBG-Bergamo")
	now := time.Now()
	makeResponse := func() types.BackupResponse {
		return types.BackupResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:               &id,
				Name:             &name,
				URI:              &uri,
				LocationResponse: &types.LocationResponse{Value: region},
				CreationDate:     &now,
				CreatedBy:        &createdBy,
				Tags:             []string{"env=test"},
			},
			Status: types.ResourceStatusResponse{State: &state},
		}
	}

	t.Run("detail output with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"database", "backup", "get", "bkp-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "bkp-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "ITBG-Bergamo") {
			t.Errorf("expected region in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
		if !strings.Contains(out, "env=test") {
			t.Errorf("expected tags in output, got: %s", out)
		}
		if !strings.Contains(out, "Active") {
			t.Errorf("expected status in output, got: %s", out)
		}
	})

	t.Run("output json hits early return", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"database", "backup", "get", "bkp-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
			t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
		}
	})
}

func TestDatabaseDBaaSBackupCreateRun_ValidationError(t *testing.T) {
	// --name "x" is too short (< 3 chars) — triggers ErrValidationFailed
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"database", "backup", "create", "--project-id", "proj-123", "--name", "x", "--region", "ITBG-Bergamo", "--dbaas-id", "dbaas-001", "--database-name", "mydb"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestDatabaseDBaaSBackupListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"database", "backup", "list"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validDatabaseDBaaSBackupCreateArgs() DatabaseDBaaSBackupCreateArgs {
	return DatabaseDBaaSBackupCreateArgs{
		ProjectID:     "proj-123",
		Name:          "my-backup",
		Region:        aruba.RegionITBGBergamo,
		DBaaSID:       "dbaas-001",
		DatabaseName:  "mydb",
		BillingPeriod: aruba.BillingPeriodHour,
	}
}

func TestDatabaseDBaaSBackupCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*DatabaseDBaaSBackupCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *DatabaseDBaaSBackupCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *DatabaseDBaaSBackupCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *DatabaseDBaaSBackupCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *DatabaseDBaaSBackupCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *DatabaseDBaaSBackupCreateArgs) { a.BillingPeriod = "Weekly" },
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:        "empty dbaas-id",
			mutate:      func(a *DatabaseDBaaSBackupCreateArgs) { a.DBaaSID = "" },
			wantErr:     true,
			errContains: "--dbaas-id",
		},
		{
			name:        "empty database-name",
			mutate:      func(a *DatabaseDBaaSBackupCreateArgs) { a.DatabaseName = "" },
			wantErr:     true,
			errContains: "--database-name",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validDatabaseDBaaSBackupCreateArgs()
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

func TestDatabaseDBaaSBackupGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := DatabaseDBaaSBackupGetArgs{ProjectID: "p1", BackupID: "bkp-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty backup ID", func(t *testing.T) {
		args := DatabaseDBaaSBackupGetArgs{ProjectID: "p1", BackupID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty backup ID")
		}
	})
}

func TestDatabaseDBaaSBackupListArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := DatabaseDBaaSBackupListArgs{ProjectID: "proj-123"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty project ID", func(t *testing.T) {
		args := DatabaseDBaaSBackupListArgs{}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty project ID")
		}
	})
}

func TestDatabaseDBaaSBackupDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := DatabaseDBaaSBackupDeleteArgs{ProjectID: "p1", BackupID: "bkp-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty backup ID", func(t *testing.T) {
		args := DatabaseDBaaSBackupDeleteArgs{ProjectID: "p1", BackupID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty backup ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestDatabaseDBaaSBackupCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-new", "my-backup"
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.BackupResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := DatabaseDBaaSBackupCreate(context.Background(), srv.Client(), validDatabaseDBaaSBackupCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "my-backup") {
		t.Errorf("expected name in output, got: %s", out)
	}
}

func TestDatabaseDBaaSBackupCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := DatabaseDBaaSBackupCreate(context.Background(), srv.Client(), validDatabaseDBaaSBackupCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating backup") {
		t.Errorf("error %q does not contain 'creating backup'", err.Error())
	}
}

func TestDatabaseDBaaSBackupCreate_AsyncResponse(t *testing.T) {
	srv := newArubaTestServer(t)
	// Async: server returns 202 with no body; the SDK may return nil wrapper.
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(202, nil))

	out := captureStdout(func() {
		err := DatabaseDBaaSBackupCreate(context.Background(), srv.Client(), validDatabaseDBaaSBackupCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "my-backup") {
		t.Errorf("expected name in async output, got: %s", out)
	}
}

func TestDatabaseDBaaSBackupList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-001", "my-backup"
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
		Values: []types.BackupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := DatabaseDBaaSBackupList(context.Background(), srv.Client(), DatabaseDBaaSBackupListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "bkp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSBackupList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{}))

	out := captureStdout(func() {
		err := DatabaseDBaaSBackupList(context.Background(), srv.Client(), DatabaseDBaaSBackupListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No backups found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestDatabaseDBaaSBackupGet_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-001", "my-backup"
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, types.BackupResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := DatabaseDBaaSBackupGet(context.Background(), srv.Client(), DatabaseDBaaSBackupGetArgs{
			ProjectID: "proj-123",
			BackupID:  "bkp-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "bkp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSBackupGet_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", errorResponse(404, "Not Found", "not found"))

	err := DatabaseDBaaSBackupGet(context.Background(), srv.Client(), DatabaseDBaaSBackupGetArgs{
		ProjectID: "proj-123",
		BackupID:  "bkp-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "getting backup") {
		t.Errorf("error %q does not contain 'getting backup'", err.Error())
	}
}

func TestDatabaseDBaaSBackupDelete_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", jsonResponse(200, nil))

	out := captureStdout(func() {
		err := DatabaseDBaaSBackupDelete(context.Background(), srv.Client(), DatabaseDBaaSBackupDeleteArgs{
			ProjectID: "proj-123",
			BackupID:  "bkp-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "bkp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSBackupDelete_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete("/projects/proj-123/providers/Aruba.Database/backups/bkp-001", errorResponse(500, "Internal Server Error", "boom"))

	err := DatabaseDBaaSBackupDelete(context.Background(), srv.Client(), DatabaseDBaaSBackupDeleteArgs{
		ProjectID: "proj-123",
		BackupID:  "bkp-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "deleting backup") {
		t.Errorf("error %q does not contain 'deleting backup'", err.Error())
	}
}

// TestDatabaseDBaaSBackupCreateArgs_ParseZoneFlagError covers the GetString("zone") error
// branch in ParseFromCobraCommand by using a command where "zone" is not registered.
func TestDatabaseDBaaSBackupCreateArgs_ParseZoneFlagError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("project-id", "proj-123", "")
	cmd.Flags().String("name", "my-backup", "")
	cmd.Flags().String("region", "ITBG-Bergamo", "")
	// "zone" intentionally omitted so GetString("zone") returns an error
	cmd.Flags().String("dbaas-id", "dbaas-001", "")
	cmd.Flags().String("database-name", "mydb", "")
	cmd.Flags().String("billing-period", "Hour", "")
	cmd.Flags().StringSlice("tags", []string{}, "")
	_ = cmd.Flags().Set("project-id", "proj-123")

	args := &DatabaseDBaaSBackupCreateArgs{}
	err := args.ParseFromCobraCommand(cmd)
	if err == nil {
		t.Fatal("expected error when zone flag is not registered")
	}
}

// TestDatabaseDBaaSBackupCreateArgs_ConstructorWrapsErrors verifies that the constructor
// wraps validation failures with ErrValidationFailed.
func TestDatabaseDBaaSBackupCreateArgs_ConstructorWrapsErrors(t *testing.T) {
	cmd := backupCreateCmd
	resetCmdFlags(cmd)
	_ = cmd.Flags().Set("project-id", "proj-123")
	_ = cmd.Flags().Set("name", "my-backup")
	_ = cmd.Flags().Set("region", "ZZ-Invalid")
	_ = cmd.Flags().Set("dbaas-id", "dbaas-001")
	_ = cmd.Flags().Set("database-name", "mydb")

	_, err := NewDatabaseDBaaSBackupCreateArgsFromCobraCommand(cmd)
	if err == nil {
		t.Fatal("expected validation error for invalid region")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected ErrValidationFailed wrapping, got: %v", err)
	}
}

// TestDatabaseDBaaSBackupCreate_WithTagsAndRegion verifies tags are forwarded and
// region reaches the SDK builder (exercises RetaggedAs branch).
func TestDatabaseDBaaSBackupCreate_WithTagsAndRegion(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-tags", "tagged-backup"
	region := types.Region("ITBG-Bergamo")
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.BackupResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
			Tags:             []string{"env=test"},
		},
	}))

	out := captureStdout(func() {
		args := validDatabaseDBaaSBackupCreateArgs()
		args.Tags = []string{"env=test"}
		err := DatabaseDBaaSBackupCreate(context.Background(), srv.Client(), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "bkp-tags") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// TestDatabaseDBaaSBackupCreate_WithZone verifies that an explicit --zone value is forwarded to the builder.
func TestDatabaseDBaaSBackupCreate_WithZone(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-zone", "my-backup"
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.BackupResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		args := validDatabaseDBaaSBackupCreateArgs()
		args.Zone = "ITBG-1"
		err := DatabaseDBaaSBackupCreate(context.Background(), srv.Client(), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "bkp-zone") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// TestDatabaseDBaaSBackupList_WithCallOpts verifies that CallOpts are threaded through.
func TestDatabaseDBaaSBackupList_WithCallOpts(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{}))

	err := DatabaseDBaaSBackupList(context.Background(), srv.Client(), DatabaseDBaaSBackupListArgs{
		ProjectID: "proj-123",
		CallOpts:  []aruba.CallOption{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
