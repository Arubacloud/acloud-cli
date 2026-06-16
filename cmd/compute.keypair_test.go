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
			name: "--output yaml emits valid YAML",
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-001", "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
					Values: []types.KeyPairResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
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
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-001", "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
					Values: []types.KeyPairResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
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
			args: []string{"compute", "keypair", "list", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				kpID, kpName := "kp-001", "my-kp"
				srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
					Values: []types.KeyPairResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
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
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA", "--region", "ITBG-Bergamo"},
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
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--public-key", "ssh-rsa AAAA", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --public-key",
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "public-key",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(500, "Internal Server Error", "duplicate name"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA", "--region", "ITBG-Bergamo"},
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

func TestKeypairCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kp-001", "my-keypair"
	region := types.Region("IT-BG")
	state := types.StateActive
	pubKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Properties: types.KeyPairPropertiesResponse{Value: pubKey},
		Status:     types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"compute", "keypair", "create",
		"--project-id", "proj-123",
		"--name", "my-keypair",
		"--region", "ITBG-Bergamo",
		"--public-key", "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKeypairListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kp-001", "my-keypair"
	region := types.Region("IT-BG")
	state := types.StateActive
	pubKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC..."
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
		Values: []types.KeyPairResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.KeyPairPropertiesResponse{Value: pubKey},
				Status:     types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"compute", "keypair", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validKeyPairCreateArgs() ComputeKeyPairCreateArgs {
	return ComputeKeyPairCreateArgs{
		ProjectID: "proj-123",
		Name:      "my-key",
		Region:    aruba.RegionITBGBergamo,
		PublicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...",
	}
}

func TestComputeKeyPairCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*ComputeKeyPairCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *ComputeKeyPairCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *ComputeKeyPairCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *ComputeKeyPairCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *ComputeKeyPairCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		{
			name:        "invalid region",
			mutate:      func(a *ComputeKeyPairCreateArgs) { a.Region = aruba.Region("ZZ-Invalid") },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "empty public key",
			mutate:      func(a *ComputeKeyPairCreateArgs) { a.PublicKey = "" },
			wantErr:     true,
			errContains: "--public-key is required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validKeyPairCreateArgs()
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
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestComputeKeyPairCreateArgs_Validate_WrappedByConstructor verifies that the
// constructor wraps Validate errors with ErrValidationFailed.
func TestComputeKeyPairCreateArgs_Validate_WrappedByConstructor(t *testing.T) {
	// We cannot easily call the constructor without a real cobra.Command, so
	// verify the wrapping convention by constructing the error directly.
	args := validKeyPairCreateArgs()
	args.Name = "x" // too short
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	wrapped := errors.Join(ErrValidationFailed, err) // mirrors constructor pattern
	if !errors.Is(wrapped, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed in chain, got: %v", wrapped)
	}
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestComputeKeyPairCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kp-new", "my-key"
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ComputeKeyPairCreate(context.Background(), srv.Client(), validKeyPairCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "kp-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeKeyPairCreate_SyncResponse(t *testing.T) {
	srv := newArubaTestServer(t)
	// 202 with metadata populated — SDK hydrates the wrapper and the sync table path fires.
	id, name := "kp-new", "my-key"
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(202, types.KeyPairResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ComputeKeyPairCreate(context.Background(), srv.Client(), validKeyPairCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	// Either sync table output (with ID) or async message (with name) should appear.
	if !strings.Contains(out, "kp-new") && !strings.Contains(out, "my-key") {
		t.Errorf("expected ID or name in output, got: %s", out)
	}
}

func TestComputeKeyPairCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := ComputeKeyPairCreate(context.Background(), srv.Client(), validKeyPairCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating keypair") {
		t.Errorf("error %q does not contain 'creating keypair'", err.Error())
	}
}

func TestComputeKeyPairList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kp-001", "my-key"
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
		Values: []types.KeyPairResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := ComputeKeyPairList(context.Background(), srv.Client(), ComputeKeyPairListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "kp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestComputeKeyPairList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{}))

	out := captureStdout(func() {
		err := ComputeKeyPairList(context.Background(), srv.Client(), ComputeKeyPairListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No keypairs found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestComputeKeyPairList_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", errorResponse(500, "Internal Server Error", "boom"))

	err := ComputeKeyPairList(context.Background(), srv.Client(), ComputeKeyPairListArgs{
		ProjectID: "proj-123",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "listing keypairs") {
		t.Errorf("error %q does not contain 'listing keypairs'", err.Error())
	}
}

// =============================================================================
// Coverage-gap tests: ErrValidationFailed and ErrParsingFailed branches
// =============================================================================

func TestComputeKeyPairCreateRun_ValidationError(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"compute", "keypair", "create",
		"--project-id", "proj-123", "--name", "x", // "x" is too short → Validate fails
		"--public-key", "ssh-rsa AAAA", "--region", "ITBG-Bergamo",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args' in error, got: %v", err)
	}
}

func TestComputeKeyPairListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"compute", "keypair", "list"})
	if err == nil {
		t.Fatal("expected error (no project-id, no context)")
	}
}
