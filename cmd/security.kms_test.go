package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestKMSListCmd(t *testing.T) {
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
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{
					Values: []types.KmsResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			args: []string{"security", "kms", "list", "--project-id", "proj-123"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{}))
			},
			args: []string{"security", "kms", "list", "--project-id", "proj-123"},
		},
		{
			name: "--output json emits valid JSON",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{
					Values: []types.KmsResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			args: []string{"security", "kms", "list", "--project-id", "proj-123", "--output", "json"},
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", errorResponse(500, "Internal Server Error", "boom"))
			},
			args:        []string{"security", "kms", "list", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", errorResponse(404, "Not Found", "resource not found"))
			},
			args:        []string{"security", "kms", "list", "--project-id", "proj-123"},
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

func TestKMSGetCmd(t *testing.T) {
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
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success --output json",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
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
			name: "detail with all optional fields covers nil-guards",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				uri := "/projects/proj-123/providers/Aruba.Security/kms/kms-001"
				createdBy := "user@example.com"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{
						ID:               &id,
						Name:             &name,
						URI:              &uri,
						CreatedBy:        &createdBy,
						LocationResponse: &types.LocationResponse{Value: types.Region("IT-BG")},
						Tags:             []string{"env=prod"},
					},
					Status: types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
				if !strings.Contains(out, "IT-BG") {
					t.Errorf("expected region in output, got: %s", out)
				}
				if !strings.Contains(out, "env=prod") {
					t.Errorf("expected tag in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", errorResponse(404, "Not Found", "resource not found"))
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
			args := []string{"security", "kms", "get", "kms-001", "--project-id", "proj-123"}
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

func TestKMSCreateCmd(t *testing.T) {
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
			args: []string{"security", "kms", "create", "--project-id", "proj-123", "--name", "my-kms", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-new", "my-kms"
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"security", "kms", "create", "--project-id", "proj-123", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"security", "kms", "create", "--project-id", "proj-123", "--name", "my-kms"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: []string{"security", "kms", "create", "--project-id", "proj-123", "--name", "my-kms", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"security", "kms", "create", "--project-id", "proj-123", "--name", "my-kms", "--region", "ITBG-Bergamo"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", errorResponse(404, "Not Found", "resource not found"))
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

func TestKMSUpdateCmd(t *testing.T) {
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
			args: []string{"security", "kms", "update", "kms-001", "--project-id", "proj-123", "--name", "renamed"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success with tags covers RetaggedAs branch",
			args: []string{"security", "kms", "update", "kms-001", "--project-id", "proj-123", "--tags", "env=test"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, Tags: []string{"env=test"}},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing flags",
			args:        []string{"security", "kms", "update", "kms-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one of",
		},
		{
			name: "pre-Get API error",
			args: []string{"security", "kms", "update", "kms-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "Update server error",
			args: []string{"security", "kms", "update", "kms-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Security/kms/kms-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestKMSDeleteCmd(t *testing.T) {
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
			args: []string{"security", "kms", "delete", "kms-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"security", "kms", "delete", "kms-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "kms-001", "my-kms"
				// Only register Get; if Delete is called the harness t.Errorf on the
				// unregistered DELETE route — stronger than a manual t.Fatal guard.
				srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-001", jsonResponse(200, types.KmsResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "kms-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"security", "kms", "delete", "kms-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"security", "kms", "delete", "kms-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Security/kms/kms-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestKMSCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-001", "my-kms"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"security", "kms", "create",
		"--project-id", "proj-123",
		"--name", "my-kms",
		"--region", "ITBG-Bergamo",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKMSListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-001", "my-kms"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{
		Values: []types.KmsResponse{
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
	out, err := runCmdCapture(srv.Client(), []string{"security", "kms", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "kms-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func validKMSCreateArgs() SecurityKMSCreateArgs {
	return SecurityKMSCreateArgs{
		ProjectID:     "proj-123",
		Name:          "my-kms",
		Region:        aruba.RegionITBGBergamo,
		BillingPeriod: aruba.BillingPeriodHour,
		Tags:          []string{},
	}
}

func TestSecurityKMSCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*SecurityKMSCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid args",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *SecurityKMSCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *SecurityKMSCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *SecurityKMSCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *SecurityKMSCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		{
			name:        "invalid region",
			mutate:      func(a *SecurityKMSCreateArgs) { a.Region = aruba.Region("ZZ-Invalid") },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *SecurityKMSCreateArgs) { a.BillingPeriod = aruba.BillingPeriod("Quarterly") },
			wantErr:     true,
			errContains: "--billing-period",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validKMSCreateArgs()
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

func TestSecurityKMSCreateArgs_Validate_WrappedByConstructor(t *testing.T) {
	args := validKMSCreateArgs()
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

func TestSecurityKMSUpdateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        SecurityKMSUpdateArgs
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid with name",
			args:    SecurityKMSUpdateArgs{ProjectID: "proj-123", ID: "kms-001", Name: "new-name"},
			wantErr: false,
		},
		{
			name:    "valid with tags changed",
			args:    SecurityKMSUpdateArgs{ProjectID: "proj-123", ID: "kms-001", TagsChanged: true},
			wantErr: false,
		},
		{
			name:        "no fields changed",
			args:        SecurityKMSUpdateArgs{ProjectID: "proj-123", ID: "kms-001"},
			wantErr:     true,
			errContains: "at least one of",
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

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestSecurityKMSCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-new", "my-kms"
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := SecurityKMSCreate(context.Background(), srv.Client(), validKMSCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "kms-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSecurityKMSCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := SecurityKMSCreate(context.Background(), srv.Client(), validKMSCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating KMS") {
		t.Errorf("error %q does not contain 'creating KMS'", err.Error())
	}
}

func TestSecurityKMSList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-001", "my-kms"
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{
		Values: []types.KmsResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := SecurityKMSList(context.Background(), srv.Client(), SecurityKMSListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "kms-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSecurityKMSList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{}))

	out := captureStdout(func() {
		err := SecurityKMSList(context.Background(), srv.Client(), SecurityKMSListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No KMS resources found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestSecurityKMSList_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", errorResponse(500, "Internal Server Error", "boom"))

	err := SecurityKMSList(context.Background(), srv.Client(), SecurityKMSListArgs{
		ProjectID: "proj-123",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "listing KMS") {
		t.Errorf("error %q does not contain 'listing KMS'", err.Error())
	}
}
