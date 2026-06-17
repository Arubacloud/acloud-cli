package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ─── Command-level tests ─────────────────────────────────────────────────────

func TestKmipListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		args        []string
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kmip-001", "my-kmip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipListResponse{
					Values: []types.KmipResponse{
						{ID: &id, Name: &name},
					},
				}))
			},
			args: []string{"security", "kmip", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kmip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipListResponse{}))
			},
			args: []string{"security", "kmip", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
		},
		{
			name: "--output json emits valid JSON",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kmip-001", "my-kmip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipListResponse{
					Values: []types.KmipResponse{
						{ID: &id, Name: &name},
					},
				}))
			},
			args: []string{"security", "kmip", "list", "--project-id", "proj-123", "--kms-id", "kms-001", "--output", "json"},
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
				}
			},
		},
		{
			name: "--output table-json emits valid JSON array",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kmip-001", "my-kmip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipListResponse{
					Values: []types.KmipResponse{
						{ID: &id, Name: &name},
					},
				}))
			},
			args: []string{"security", "kmip", "list", "--project-id", "proj-123", "--kms-id", "kms-001", "--output", "table-json"},
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
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", errorResponse(500, "Internal Server Error", "boom"))
			},
			args:        []string{"security", "kmip", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", errorResponse(404, "Not Found", "resource not found"))
			},
			args:        []string{"security", "kmip", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name:        "missing required --kms-id",
			args:        []string{"security", "kmip", "list", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "kms-id",
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

func TestKmipGetCmd(t *testing.T) {
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
				id, name := "kmip-001", "my-kmip"
				status := types.ServiceStatusCertificateAvailable
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", jsonResponse(200, types.KmipResponse{
					ID: &id, Name: &name, Status: &status,
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kmip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
				if !strings.Contains(out, "CertificateAvailable") {
					t.Errorf("expected status in output, got: %s", out)
				}
			},
		},
		{
			name: "success --output json",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kmip-001", "my-kmip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", jsonResponse(200, types.KmipResponse{
					ID: &id, Name: &name,
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
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", errorResponse(404, "Not Found", "resource not found"))
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
			args := []string{"security", "kmip", "get", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001"}
			if tc.name == "success --output json" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(srv.Client(), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestKmipCreateCmd(t *testing.T) {
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
			args: []string{"security", "kmip", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-kmip"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kmip-new", "my-kmip"
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipResponse{
					ID: &id, Name: &name,
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kmip-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"security", "kmip", "create", "--project-id", "proj-123", "--kms-id", "kms-001"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --kms-id",
			args:        []string{"security", "kmip", "create", "--project-id", "proj-123", "--name", "my-kmip"},
			wantErr:     true,
			errContains: "kms-id",
		},
		{
			name: "server error propagates",
			args: []string{"security", "kmip", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-kmip"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"security", "kmip", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-kmip"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", errorResponse(404, "Not Found", "kms not found"))
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

func TestKmipDeleteCmd(t *testing.T) {
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
			args: []string{"security", "kmip", "delete", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kmip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"security", "kmip", "delete", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kmip-001", "my-kmip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", jsonResponse(200, types.KmipResponse{
					ID: &id, Name: &name,
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kmip-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"security", "kmip", "delete", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", errorResponse(500, "Internal Server Error", "kmip in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"security", "kmip", "delete", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestKmipDownloadCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		args        []string
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001/download", jsonResponse(200, types.KmipCertificateResponse{
					Cert: "-----BEGIN CERTIFICATE-----\nMIICxx\n-----END CERTIFICATE-----",
					Key:  "-----BEGIN PRIVATE KEY-----\nMIIExx\n-----END PRIVATE KEY-----",
				}))
			},
			args: []string{"security", "kmip", "download", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "BEGIN CERTIFICATE") {
					t.Errorf("expected certificate in output, got: %s", out)
				}
				if !strings.Contains(out, "BEGIN PRIVATE KEY") {
					t.Errorf("expected private key in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001/download", errorResponse(404, "Not Found", "certificate not ready"))
			},
			args:        []string{"security", "kmip", "download", "kmip-001", "--project-id", "proj-123", "--kms-id", "kms-001"},
			wantErr:     true,
			errContains: "downloading",
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

// ─── Operation-function tests ─────────────────────────────────────────────────

func validKmipCreateArgs() SecurityKmipCreateArgs {
	return SecurityKmipCreateArgs{
		ProjectID: "proj-123",
		KMSID:     "kms-001",
		Name:      "my-kmip",
	}
}

func TestSecurityKmipCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*SecurityKmipCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid args",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *SecurityKmipCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *SecurityKmipCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *SecurityKmipCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *SecurityKmipCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		{
			name:        "missing kms-id",
			mutate:      func(a *SecurityKmipCreateArgs) { a.KMSID = "" },
			wantErr:     true,
			errContains: "--kms-id is required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validKmipCreateArgs()
			if tc.mutate != nil {
				tc.mutate(&args)
			}
			err := args.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
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

func TestSecurityKmipCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kmip-new", "my-kmip"
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipResponse{
		ID: &id, Name: &name,
	}))

	out := captureStdout(func() {
		err := SecurityKmipCreate(context.Background(), srv.Client(), validKmipCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "kmip-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSecurityKmipCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", errorResponse(422, "Unprocessable Entity", "kms capacity exceeded"))

	err := SecurityKmipCreate(context.Background(), srv.Client(), validKmipCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating KMIP") {
		t.Errorf("expected 'creating KMIP' in error, got: %v", err)
	}
}

func TestSecurityKmipList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip", jsonResponse(200, types.KmipListResponse{}))

	out := captureStdout(func() {
		args := SecurityKmipListArgs{ProjectID: "proj-123", KMSID: "kms-001"}
		err := SecurityKmipList(context.Background(), srv.Client(), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No KMIP services found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestSecurityKmipDelete_DryRun(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kmip-001", "my-kmip"
	// Only register Get; unregistered DELETE would cause the harness to fail the test.
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001", jsonResponse(200, types.KmipResponse{
		ID: &id, Name: &name,
	}))

	args := SecurityKmipDeleteArgs{
		ProjectID:   "proj-123",
		KMSID:       "kms-001",
		ID:          "kmip-001",
		DryRun:      true,
		SkipConfirm: true,
	}
	out := captureStdout(func() {
		if err := SecurityKmipDelete(context.Background(), srv.Client(), args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "kmip-001") {
		t.Errorf("expected ID in dry-run output, got: %s", out)
	}
}

func TestSecurityKmipDownload_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/kmip/kmip-001/download", jsonResponse(200, types.KmipCertificateResponse{
		Cert: "-----BEGIN CERTIFICATE-----\nMIICxx\n-----END CERTIFICATE-----",
		Key:  "-----BEGIN PRIVATE KEY-----\nMIIExx\n-----END PRIVATE KEY-----",
	}))

	out := captureStdout(func() {
		args := SecurityKmipDownloadArgs{ProjectID: "proj-123", KMSID: "kms-001", ID: "kmip-001"}
		if err := SecurityKmipDownload(context.Background(), srv.Client(), args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "BEGIN CERTIFICATE") {
		t.Errorf("expected certificate in output, got: %s", out)
	}
	if !strings.Contains(out, "BEGIN PRIVATE KEY") {
		t.Errorf("expected private key in output, got: %s", out)
	}
}

func TestSecurityKmipCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	// --name "ab" is too short (< 3 chars) — triggers ErrValidationFailed
	err := runCmd(srv.Client(), []string{"security", "kmip", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "ab"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}
