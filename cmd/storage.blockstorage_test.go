package cmd

import (
	"encoding/json"
	"strings"
	"testing"

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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages", jsonResponse(200, types.BlockStorageList{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages", jsonResponse(200, types.BlockStorageList{}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages", jsonResponse(200, types.BlockStorageList{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "blockstorage", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages", errorResponse(404, "Not Found", "resource not found"))
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(500, "Internal Server Error", "boom"))
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
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockstorages", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockstorages", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			args: []string{"storage", "blockstorage", "create", "--project-id", "proj-123", "--name", "my-vol", "--region", "ITBG-Bergamo", "--size", "10"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockstorages", errorResponse(500, "Internal Server Error", "boom"))
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
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, nil))
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(500, "Internal Server Error", "boom"))
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
				state := "NotUsed"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatus{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"storage", "blockstorage", "update", "vol-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "vol-001", "my-volume"
				state := "NotUsed"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", jsonResponse(200, types.BlockStorageResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatus{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/blockstorages/vol-001", errorResponse(500, "Internal Server Error", "boom"))
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
