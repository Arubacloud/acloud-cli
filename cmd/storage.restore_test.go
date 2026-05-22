package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

const (
	restoreBkpURI = "/projects/proj-123/providers/Aruba.Storage/backups/bkp-001"
	restoreVolURI = "/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001"
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(404, "Not Found", "volume not found"))
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreList{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreList{}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"storage", "restore", "list", "bkp-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "rst-001", "my-restore"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreList{
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
