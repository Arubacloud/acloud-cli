package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

const (
	restoreBkpURI = "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001"
	restoreVolURI = "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
)

func TestStorageRestoreCreateCmd(t *testing.T) {
	createArgs := []string{"storage", "restore", "bkp-001", "vol-001", "--project-id", "proj-123", "--name", "my-restore"}

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
				id, name := "rst-new", "my-restore"
				bkpURI := restoreBkpURI
				volURI := restoreVolURI
				// Pre-validation: GET backup
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{URI: &bkpURI},
				}))
				// Pre-validation: GET volume
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{URI: &volURI},
				}))
				// Create: POST restore (scoped to the backup)
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "restore", "bkp-001", "vol-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "backup pre-validation error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", errorResponse(404, "Not Found", "backup not found"))
			},
			wantErr: true, errContains: "getting backup",
		},
		{
			name: "volume pre-validation error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				bkpURI := restoreBkpURI
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{URI: &bkpURI},
				}))
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(404, "Not Found", "volume not found"))
			},
			wantErr: true, errContains: "getting volume",
		},
		{
			name: "API error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				bkpURI := restoreBkpURI
				volURI := restoreVolURI
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{URI: &bkpURI},
				}))
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{URI: &volURI},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				bkpURI := restoreBkpURI
				volURI := restoreVolURI
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
					Metadata: types.ResourceMetadataResponse{URI: &bkpURI},
				}))
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{URI: &volURI},
				}))
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", errorResponse(500, "Internal Server Error", "boom"))
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

func TestStorageRestoreListCmd(t *testing.T) {
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
			args: []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreListResponse{
					Values: []types.StorageRestoreResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreListResponse{}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreListResponse{
					Values: []types.StorageRestoreResponse{
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
			args: []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", errorResponse(404, "Not Found", "resource not found"))
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

func TestStorageRestoreGetCmd(t *testing.T) {
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
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", jsonResponse(200, types.StorageRestoreResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", errorResponse(500, "Internal Server Error", "boom"))
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
				[]string{"storage", "restore", "get", "bkp-001", "rst-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageRestoreDeleteCmd(t *testing.T) {
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
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:      "--dry-run: prints intent, does not call Delete",
			extraArgs: []string{"--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", jsonResponse(200, types.StorageRestoreResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", errorResponse(500, "Internal Server Error", "boom"))
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
			args := []string{"storage", "restore", "delete", "bkp-001", "rst-001", "--project-id", "proj-123", "--yes"}
			args = append(args, tc.extraArgs...)
			out, err := runCmdCapture(srv.Client(), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestStorageRestoreUpdateCmd(t *testing.T) {
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
			args: []string{"storage", "restore", "update", "bkp-001", "rst-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", jsonResponse(200, types.StorageRestoreResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", jsonResponse(200, types.StorageRestoreResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "rst-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"storage", "restore", "update", "bkp-001", "rst-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-GET error",
			args: []string{"storage", "restore", "update", "bkp-001", "rst-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"storage", "restore", "update", "bkp-001", "rst-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", jsonResponse(200, types.StorageRestoreResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001", errorResponse(500, "Internal Server Error", "boom"))
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

// TestStorageRestoreGetCmd_FullDetail exercises all nil-guard branches in GET detail output.
func TestStorageRestoreGetCmd_FullDetail(t *testing.T) {
	t.Run("detail with all optional fields covers nil-guards", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "rst-001", "my-restore"
		uri := "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001"
		createdBy := "user@example.com"
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		state := types.StateActive
		region := types.Region("IT-BG")
		destURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
		srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001",
			jsonResponse(200, types.StorageRestoreResponse{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					URI:              &uri,
					LocationResponse: &types.LocationResponse{Value: region},
					CreationDate:     &ts,
					CreatedBy:        &createdBy,
					Tags:             []string{"env=test"},
				},
				Status: types.ResourceStatusResponse{State: &state},
				Properties: types.StorageRestorePropertiesResponse{
					Destination: types.ReferenceResourceCommon{URI: destURI},
				},
			}))
		out, err := runCmdCapture(srv.Client(), []string{"storage", "restore", "get", "bkp-001", "rst-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "rst-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
		if !strings.Contains(out, "IT-BG") {
			t.Errorf("expected region in output, got: %s", out)
		}
		if !strings.Contains(out, "env=test") {
			t.Errorf("expected tags in output, got: %s", out)
		}
	})

	t.Run("--output json emits valid JSON", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "rst-001", "my-restore"
		srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/rst-001",
			jsonResponse(200, types.StorageRestoreResponse{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
			}))
		out, err := runCmdCapture(srv.Client(), []string{"storage", "restore", "get", "bkp-001", "rst-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
			t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
		}
	})
}

// TestStorageRestoreListCmd_AllOptionalFields exercises nil-guard branches in LIST output.
func TestStorageRestoreListCmd_AllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "rst-001", "my-restore"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores",
		jsonResponse(200, types.StorageRestoreListResponse{
			Values: []types.StorageRestoreResponse{
				{
					Metadata: types.ResourceMetadataResponse{
						ID:   &id,
						Name: &name,
					},
					Status: types.ResourceStatusResponse{State: &state},
				},
			},
		}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "rst-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestStorageRestoreCreateCmd_WithStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	backupURI := "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001"
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	id, name := "restore-001", "my-restore"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001", jsonResponse(200, types.StorageBackupResponse{
		Metadata: types.ResourceMetadataResponse{URI: &backupURI},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{URI: &volURI},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"storage", "restore", "bkp-001", "vol-001",
		"--project-id", "proj-123",
		"--name", "my-restore",
		"--region", "IT-BG",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStorageRestoreListCmd_WithStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "restore-001", "my-restore"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreListResponse{
		Values: []types.StorageRestoreResponse{
			{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				Status:   types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "restore-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestStorageRestoreGetCmd_WithAllOptionals(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "restore-001", "my-restore"
	uri := "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/restore-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores/restore-001", jsonResponse(200, types.StorageRestoreResponse{
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
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "restore", "get", "bkp-001", "restore-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "restore-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
