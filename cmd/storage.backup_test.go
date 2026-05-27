package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestStorageBackupCreateCmd(t *testing.T) {
	createArgs := []string{"storage", "backup", "vol-001", "--project-id", "proj-123", "--name", "my-backup"}
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"

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
				// Pre-validation: GET volume
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{URI: &volURI},
				}))
				// Create: POST backup
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "backup", "vol-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "volume pre-validation error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(404, "Not Found", "volume not found"))
			},
			wantErr: true, errContains: "getting volume",
		},
		{
			name: "API error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{URI: &volURI},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{URI: &volURI},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "creating",
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

func TestStorageBackupListCmd(t *testing.T) {
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
			args: []string{"storage", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupList{
					Values: []types.StorageBackupResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"storage", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupList{}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"storage", "backup", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupList{
					Values: []types.StorageBackupResponse{
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
			name: "server error propagates",
			args: []string{"storage", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "backup", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
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

func TestStorageBackupGetCmd(t *testing.T) {
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
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
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "getting",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			out, err := runCmdCapture(srv.Client(),
				[]string{"storage", "backup", "get", "bkp-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageBackupDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupSrv    func(*arubaTestServer)
		extraArgs   []string
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "bkp-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:      "--dry-run: prints intent, does not call Delete",
			extraArgs: []string{"--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
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
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "deleting",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			args := []string{"storage", "backup", "delete", "bkp-001", "--project-id", "proj-123", "--yes"}
			args = append(args, tc.extraArgs...)
			out, err := runCmdCapture(srv.Client(), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageBackupUpdateCmd(t *testing.T) {
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
			args: []string{"storage", "backup", "update", "bkp-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
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
			name:        "no flags error",
			args:        []string{"storage", "backup", "update", "bkp-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-GET error",
			args: []string{"storage", "backup", "update", "bkp-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"storage", "backup", "update", "bkp-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "bkp-001", "my-backup"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestStorageBackupCreateCmd_WithRetentionAndBilling(t *testing.T) {
	srv := newArubaTestServer(t)
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	id, name := "bkp-001", "my-backup"
	state := types.StateActive
	period := types.BillingPeriodMonth
	retDays := 30
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{URI: &volURI},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.StorageBackupPropertiesResult{
			Type:          types.StorageBackupTypeFull,
			RetentionDays: &retDays,
			BillingPeriod: &period,
		},
		Status: types.ResourceStatus{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"storage", "backup", "vol-001",
		"--project-id", "proj-123",
		"--name", "my-backup",
		"--region", "IT-BG",
		"--type", "Full",
		"--retention-days", "30",
		"--billing-period", "Month",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageBackupListCmd_WithStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-001", "my-backup"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupList{
		Values: []types.StorageBackupResponse{
			{
				Metadata:   types.ResourceMetadataResponse{ID: &id, Name: &name},
				Properties: types.StorageBackupPropertiesResult{Type: types.StorageBackupTypeFull},
				Status:     types.ResourceStatus{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "backup", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bkp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestStorageBackupGetCmd_JSONOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-001", "my-backup"
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "backup", "get", "bkp-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestStorageBackupGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "bkp-001", "my-backup"
	uri := "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	createdBy := "test-user@example.com"
	period := types.BillingPeriodMonth
	retDays := 30
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.StorageBackupPropertiesResult{
			Type:          types.StorageBackupTypeFull,
			Origin:        types.ReferenceResource{URI: volURI},
			RetentionDays: &retDays,
			BillingPeriod: &period,
		},
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "backup", "get", "bkp-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bkp-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
