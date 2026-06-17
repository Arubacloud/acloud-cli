package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// ─── Command-level tests ─────────────────────────────────────────────────────

func TestKeyListCmd(t *testing.T) {
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
				id, name := "key-001", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyListResponse{
					Values: []types.KeyResponse{
						{KeyID: &id, Name: &name, Algorithm: &algo},
					},
				}))
			},
			args: []string{"security", "key", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "key-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyListResponse{}))
			},
			args: []string{"security", "key", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
		},
		{
			name: "--output json emits valid JSON",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "key-001", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyListResponse{
					Values: []types.KeyResponse{
						{KeyID: &id, Name: &name, Algorithm: &algo},
					},
				}))
			},
			args: []string{"security", "key", "list", "--project-id", "proj-123", "--kms-id", "kms-001", "--output", "json"},
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
				id, name := "key-001", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyListResponse{
					Values: []types.KeyResponse{
						{KeyID: &id, Name: &name, Algorithm: &algo},
					},
				}))
			},
			args: []string{"security", "key", "list", "--project-id", "proj-123", "--kms-id", "kms-001", "--output", "table-json"},
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", errorResponse(500, "Internal Server Error", "boom"))
			},
			args:        []string{"security", "key", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", errorResponse(404, "Not Found", "resource not found"))
			},
			args:        []string{"security", "key", "list", "--project-id", "proj-123", "--kms-id", "kms-001"},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name:        "missing required --kms-id",
			args:        []string{"security", "key", "list", "--project-id", "proj-123"},
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

func TestKeyGetCmd(t *testing.T) {
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
				id, name := "key-001", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", jsonResponse(200, types.KeyResponse{
					KeyID: &id, Name: &name, Algorithm: &algo,
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "key-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
				if !strings.Contains(out, "Aes") {
					t.Errorf("expected algorithm in output, got: %s", out)
				}
			},
		},
		{
			name: "success --output json",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "key-001", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", jsonResponse(200, types.KeyResponse{
					KeyID: &id, Name: &name, Algorithm: &algo,
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", errorResponse(404, "Not Found", "resource not found"))
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
			args := []string{"security", "key", "get", "key-001", "--project-id", "proj-123", "--kms-id", "kms-001"}
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

func TestKeyCreateCmd(t *testing.T) {
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
			args: []string{"security", "key", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-key", "--algorithm", "Aes"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "key-new", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyResponse{
					KeyID: &id, Name: &name, Algorithm: &algo,
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "key-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"security", "key", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--algorithm", "Aes"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --algorithm",
			args:        []string{"security", "key", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-key"},
			wantErr:     true,
			errContains: "algorithm",
		},
		{
			name:        "missing required flag --kms-id",
			args:        []string{"security", "key", "create", "--project-id", "proj-123", "--name", "my-key", "--algorithm", "Aes"},
			wantErr:     true,
			errContains: "kms-id",
		},
		{
			name: "server error propagates",
			args: []string{"security", "key", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-key", "--algorithm", "Aes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"security", "key", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "my-key", "--algorithm", "Aes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", errorResponse(404, "Not Found", "kms not found"))
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

func TestKeyDeleteCmd(t *testing.T) {
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
			args: []string{"security", "key", "delete", "key-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "key-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"security", "key", "delete", "key-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "key-001", "my-key"
				algo := types.KeyAlgorithmAes
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", jsonResponse(200, types.KeyResponse{
					KeyID: &id, Name: &name, Algorithm: &algo,
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "key-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"security", "key", "delete", "key-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", errorResponse(500, "Internal Server Error", "key in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"security", "key", "delete", "key-001", "--project-id", "proj-123", "--kms-id", "kms-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", errorResponse(404, "Not Found", "resource not found"))
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

// ─── Operation-function tests ─────────────────────────────────────────────────

func validKeyCreateArgs() SecurityKeyCreateArgs {
	return SecurityKeyCreateArgs{
		ProjectID: "proj-123",
		KMSID:     "kms-001",
		Name:      "my-key",
		Algorithm: aruba.KeyAlgorithmAes,
	}
}

func TestSecurityKeyCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*SecurityKeyCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid args",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *SecurityKeyCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *SecurityKeyCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *SecurityKeyCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *SecurityKeyCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		{
			name:        "invalid algorithm",
			mutate:      func(a *SecurityKeyCreateArgs) { a.Algorithm = "Invalid" },
			wantErr:     true,
			errContains: "--algorithm",
		},
		{
			name:        "missing kms-id",
			mutate:      func(a *SecurityKeyCreateArgs) { a.KMSID = "" },
			wantErr:     true,
			errContains: "--kms-id is required",
		},
		{
			name:    "Rsa algorithm valid",
			mutate:  func(a *SecurityKeyCreateArgs) { a.Algorithm = aruba.KeyAlgorithmRsa },
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validKeyCreateArgs()
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

func TestSecurityKeyCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "key-new", "my-key"
	algo := types.KeyAlgorithmAes
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyResponse{
		KeyID: &id, Name: &name, Algorithm: &algo,
	}))

	out := captureStdout(func() {
		err := SecurityKeyCreate(context.Background(), srv.Client(), validKeyCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "key-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSecurityKeyCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", errorResponse(422, "Unprocessable Entity", "invalid algorithm"))

	err := SecurityKeyCreate(context.Background(), srv.Client(), validKeyCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating key") {
		t.Errorf("expected 'creating key' in error, got: %v", err)
	}
}

func TestSecurityKeyList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyListResponse{}))

	out := captureStdout(func() {
		args := SecurityKeyListArgs{ProjectID: "proj-123", KMSID: "kms-001"}
		err := SecurityKeyList(context.Background(), srv.Client(), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No keys found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestSecurityKeyDelete_DryRun(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "key-001", "my-key"
	algo := types.KeyAlgorithmAes
	// Only register Get; unregistered DELETE would cause the harness to fail the test.
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", jsonResponse(200, types.KeyResponse{
		KeyID: &id, Name: &name, Algorithm: &algo,
	}))

	args := SecurityKeyDeleteArgs{
		ProjectID:   "proj-123",
		KMSID:       "kms-001",
		ID:          "key-001",
		DryRun:      true,
		SkipConfirm: true,
	}
	out := captureStdout(func() {
		if err := SecurityKeyDelete(context.Background(), srv.Client(), args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "key-001") {
		t.Errorf("expected ID in dry-run output, got: %s", out)
	}
}

func TestSecurityKeyCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	// --name "ab" is too short (< 3 chars) — triggers ErrValidationFailed
	err := runCmd(srv.Client(), []string{"security", "key", "create", "--project-id", "proj-123", "--kms-id", "kms-001", "--name", "ab", "--algorithm", "Aes"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestSecurityKeyCreate_WithStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "key-new", "my-key"
	algo := types.KeyAlgorithmAes
	keyType := types.KeyTypeSymmetric
	status := types.KeyStatusActive
	src := types.KeyCreationSourceCmp
	privID := "priv-key-001"
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyResponse{
		KeyID: &id, Name: &name, Algorithm: &algo,
		Type: &keyType, Status: &status, CreationSource: &src, PrivateKeyID: &privID,
	}))

	out := captureStdout(func() {
		err := SecurityKeyCreate(context.Background(), srv.Client(), validKeyCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "key-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSecurityKeyGet_WithAllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "key-001", "my-key"
	algo := types.KeyAlgorithmRsa
	keyType := types.KeyTypeAsymmetric
	status := types.KeyStatusActive
	src := types.KeyCreationSourceCmp
	privID := "priv-key-001"
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys/key-001", jsonResponse(200, types.KeyResponse{
		KeyID: &id, Name: &name, Algorithm: &algo,
		Type: &keyType, Status: &status, CreationSource: &src, PrivateKeyID: &privID,
	}))

	out := captureStdout(func() {
		args := SecurityKeyGetArgs{ProjectID: "proj-123", KMSID: "kms-001", ID: "key-001"}
		if err := SecurityKeyGet(context.Background(), srv.Client(), args); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Rsa") {
		t.Errorf("expected algorithm in output, got: %s", out)
	}
	if !strings.Contains(out, "Asymmetric") {
		t.Errorf("expected type in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
	if !strings.Contains(out, "priv-key-001") {
		t.Errorf("expected private key ID in output, got: %s", out)
	}
}

func TestSecurityKeyGetArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        SecurityKeyGetArgs
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid",
			args:    SecurityKeyGetArgs{ProjectID: "p", KMSID: "k", ID: "key-001"},
			wantErr: false,
		},
		{
			name:        "empty kms-id",
			args:        SecurityKeyGetArgs{ProjectID: "p", KMSID: "", ID: "key-001"},
			wantErr:     true,
			errContains: "--kms-id is required",
		},
		{
			name:        "empty id",
			args:        SecurityKeyGetArgs{ProjectID: "p", KMSID: "k", ID: ""},
			wantErr:     true,
			errContains: "key ID is required",
		},
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

func TestSecurityKeyDeleteArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        SecurityKeyDeleteArgs
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid",
			args:    SecurityKeyDeleteArgs{ProjectID: "p", KMSID: "k", ID: "key-001"},
			wantErr: false,
		},
		{
			name:        "empty kms-id",
			args:        SecurityKeyDeleteArgs{ProjectID: "p", KMSID: "", ID: "key-001"},
			wantErr:     true,
			errContains: "--kms-id is required",
		},
		{
			name:        "empty id",
			args:        SecurityKeyDeleteArgs{ProjectID: "p", KMSID: "k", ID: ""},
			wantErr:     true,
			errContains: "key ID is required",
		},
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

func TestSecurityKeyListArgs_Validate(t *testing.T) {
	a1 := SecurityKeyListArgs{ProjectID: "p", KMSID: ""}
	if err := a1.Validate(); err == nil || !strings.Contains(err.Error(), "--kms-id is required") {
		t.Errorf("expected kms-id error, got: %v", err)
	}
	a2 := SecurityKeyListArgs{ProjectID: "p", KMSID: "k"}
	if err := a2.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompleteKeyID_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "key-001", "my-key"
	algo := types.KeyAlgorithmAes
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", jsonResponse(200, types.KeyListResponse{
		Values: []types.KeyResponse{{KeyID: &id, Name: &name, Algorithm: &algo}},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	resetCmdFlags(keyGetCmd)
	if err := keyGetCmd.Flags().Set("kms-id", "kms-001"); err != nil {
		t.Fatalf("set kms-id: %v", err)
	}
	if err := keyGetCmd.Flags().Set("project-id", "proj-123"); err != nil {
		t.Fatalf("set project-id: %v", err)
	}

	completions, directive := completeKeyID(keyGetCmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteKeyID_NoKMSID(t *testing.T) {
	srv := newArubaTestServer(t)
	setClientForTesting(srv.Client())
	defer resetClientState()

	resetCmdFlags(keyGetCmd)
	if err := keyGetCmd.Flags().Set("project-id", "proj-123"); err != nil {
		t.Fatalf("set project-id: %v", err)
	}
	// kms-id not set — should return early with NoFileComp
	_, directive := completeKeyID(keyGetCmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
}

func TestCompleteKeyID_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001/keys", errorResponse(500, "Internal Server Error", "boom"))
	setClientForTesting(srv.Client())
	defer resetClientState()

	resetCmdFlags(keyGetCmd)
	if err := keyGetCmd.Flags().Set("kms-id", "kms-001"); err != nil {
		t.Fatalf("set kms-id: %v", err)
	}
	if err := keyGetCmd.Flags().Set("project-id", "proj-123"); err != nil {
		t.Fatalf("set project-id: %v", err)
	}

	completions, directive := completeKeyID(keyGetCmd, []string{}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("unexpected directive: %v", directive)
	}
	if len(completions) != 0 {
		t.Errorf("expected no completions on error, got: %v", completions)
	}
}
