package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestBlockStorageListCmd(t *testing.T) {
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
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
					Values: []types.BlockStorageResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
					Values: []types.BlockStorageResponse{
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
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", errorResponse(404, "Not Found", "resource not found"))
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

func TestBlockStorageGetCmd(t *testing.T) {
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
				id, name := "vol-001", "my-volume"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(500, "Internal Server Error", "boom"))
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
				[]string{"storage", "blockstorage", "get", "vol-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestBlockStorageCreateCmd(t *testing.T) {
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
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-new", "my-vol"
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--region", "ITBG-Bergamo", "--size", "10"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --size",
			args:        []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo"},
			wantErr:     true,
			errContains: "size",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockStorages", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockStorages", errorResponse(500, "Internal Server Error", "boom"))
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

func TestBlockStorageDeleteCmd(t *testing.T) {
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
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:      "--dry-run: prints intent, does not call Delete",
			extraArgs: []string{"--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(500, "Internal Server Error", "boom"))
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
			args := []string{"storage", "blockstorage", "delete", "vol-001", "--project-id", "proj-123", "--yes"}
			args = append(args, tc.extraArgs...)
			out, err := runCmdCapture(srv.Client(), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestBlockStorageUpdateCmd(t *testing.T) {
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
			args: []string{"storage", "blockstorage", "update", "vol-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				state := types.StateNotUsed
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "vol-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "pre-GET error",
			args: []string{"storage", "blockstorage", "update", "vol-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"storage", "blockstorage", "update", "vol-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				state := types.StateNotUsed
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestBlockStorageListCmd_AllOptionalFields(t *testing.T) {
	// Covers: Status.State nil-guard in list loop (lines 194-195).
	srv := newArubaTestServer(t)
	id, name := "vol-001", "my-volume"
	state := types.StateActive
	region := types.Region("IT-BG")
	bootable := true
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
		Values: []types.BlockStorageResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.BlockStoragePropertiesResponse{
					SizeGB:   50,
					Bootable: &bootable,
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "blockstorage", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "vol-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "IT-BG") {
		t.Errorf("expected region in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestBlockStorageGetCmd_AllOptionalFields(t *testing.T) {
	// Covers: URI, LocationResponse, Bootable, Status.State, CreationDate, CreatedBy, Tags detail block.
	// Also covers the JSON early-return path (PrintOutput(vol, nil, nil)).
	id, name := "vol-001", "my-volume"
	uri := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	createdBy := "user@example.com"
	state := types.StateActive
	region := types.Region("IT-BG")
	bootable := true
	now := time.Now()
	makeResponse := func() types.BlockStorageResponse {
		return types.BlockStorageResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:               &id,
				Name:             &name,
				URI:              &uri,
				LocationResponse: &types.LocationResponse{Value: region},
				CreationDate:     &now,
				CreatedBy:        &createdBy,
				Tags:             []string{"env=test"},
			},
			Properties: types.BlockStoragePropertiesResponse{
				SizeGB:   50,
				Bootable: &bootable,
			},
			Status: types.ResourceStatusResponse{State: &state},
		}
	}

	t.Run("detail output with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"storage", "blockstorage", "get", "vol-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "vol-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "IT-BG") {
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
		srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"storage", "blockstorage", "get", "vol-001", "--project-id", "proj-123", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
			t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
		}
	})
}

func TestBlockStorageUpdateCmd_NoFlagsError(t *testing.T) {
	// Covers: "at least one" error branch (line 296).
	srv := newArubaTestServer(t)
	_, err := runCmdCapture(srv.Client(), []string{"storage", "blockstorage", "update", "vol-001", "--project-id", "proj-123"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "at least one") {
		t.Errorf("expected 'at least one' in error, got: %v", err)
	}
}

func TestBlockStorageCreateCmd_WithSnapshotID(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vol-002", "my-volume"
	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	err := runCmd(srv.Client(), []string{
		"storage", "blockstorage", "create",
		"--project-id", "proj-123",
		"--name", "my-volume",
		"--region", "IT-BG",
		"--size", "50",
		"--snapshot-id", "snap-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
