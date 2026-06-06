package cmd

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSUserListCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.DatabaseUserListResponse{
					Values: []types.UserResponse{
						{Username: "admin"},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "admin") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.DatabaseUserListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.DatabaseUserListResponse{
					Values: []types.UserResponse{
						{Username: "admin"},
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
			args: []string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSUserGetCmd(t *testing.T) {
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/admin", jsonResponse(200, types.UserResponse{
					Username: "admin",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "admin") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/admin", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/admin", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "user", "get", "dbaas-001", "admin", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestDBaaSUserCreateCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser", "--password", "Pass1!"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.UserResponse{
					Username: "myuser",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --username",
			args:        []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--password", "Pass1!"},
			wantErr:     true,
			errContains: "username",
		},
		{
			name:        "missing required flag --password",
			args:        []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser"},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser", "--password", "Pass1!"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", errorResponse(500, "Internal Server Error", "duplicate user"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "user", "create", "dbaas-001", "--project-id", "proj-123", "--username", "myuser", "--password", "Pass1!"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSUserDeleteCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "user", "delete", "dbaas-001", "myuser", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"database", "dbaas", "user", "delete", "dbaas-001", "myuser", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", jsonResponse(200, types.UserResponse{
					Username: "myuser",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "user", "delete", "dbaas-001", "myuser", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", errorResponse(500, "Internal Server Error", "not found"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "user", "delete", "dbaas-001", "myuser", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSUserUpdateCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "user", "update", "dbaas-001", "myuser", "--project-id", "proj-123", "--password", "newpass"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", jsonResponse(200, types.UserResponse{
					Username: "myuser",
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", jsonResponse(200, types.UserResponse{
					Username: "myuser",
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "myuser") {
					t.Errorf("expected username in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing --password",
			args:        []string{"database", "dbaas", "user", "update", "dbaas-001", "myuser", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "password",
		},
		{
			name: "pre-GET error",
			args: []string{"database", "dbaas", "user", "update", "dbaas-001", "myuser", "--project-id", "proj-123", "--password", "p"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", errorResponse(404, "Not Found", "not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"database", "dbaas", "user", "update", "dbaas-001", "myuser", "--project-id", "proj-123", "--password", "p"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", jsonResponse(200, types.UserResponse{
					Username: "myuser",
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/myuser", errorResponse(500, "Internal Server Error", "boom"))
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

// TestDBaaSUserGetCmd_FullDetail exercises all nil-guard branches in GET detail output.
func TestDBaaSUserGetCmd_FullDetail(t *testing.T) {
	t.Run("detail with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		createdBy := "user@example.com"
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/admin", jsonResponse(200, types.UserResponse{
			Username:     "admin",
			CreatedBy:    &createdBy,
			CreationDate: &ts,
		}))
		out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "user", "get", "dbaas-001", "admin", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "admin") {
			t.Errorf("expected username in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
	})

	t.Run("--output json emits valid JSON", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users/admin", jsonResponse(200, types.UserResponse{Username: "admin"}))
		out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "user", "get", "dbaas-001", "admin", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
			t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
		}
	})
}

// TestDBaaSUserListCmd_AllOptionalFields exercises nil-guard branches in LIST output.
func TestDBaaSUserListCmd_AllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	createdBy := "user@example.com"
	ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.DatabaseUserListResponse{
		Values: []types.UserResponse{
			{Username: "admin", CreatedBy: &createdBy, CreationDate: &ts},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "user", "list", "dbaas-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "admin") {
		t.Errorf("expected username in output, got: %s", out)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("expected createdBy in output, got: %s", out)
	}
}

func TestDatabaseDBaaSUserListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"database", "dbaas", "user", "list", "dbaas-001"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDBaaSUserCreateCmd_WithCreationDate(t *testing.T) {
	srv := newArubaTestServer(t)
	now := time.Now()
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.UserResponse{
		Username:     "myuser",
		CreationDate: &now,
	}))
	err := runCmd(srv.Client(), []string{
		"database", "dbaas", "user", "create", "dbaas-001",
		"--project-id", "proj-123",
		"--username", "myuser",
		"--password", "mypassword",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
