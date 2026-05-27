package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

const snapshotVolID = "vol-001"
const snapshotVolURI = "/projects/proj-123/providers/Aruba.Storage/blockStorages/" + snapshotVolID

func TestSnapshotListCmd(t *testing.T) {
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
			args: []string{"storage", "snapshot", "list", "--project-id", "proj-123", "--volume-id", snapshotVolID},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "snap-001", "my-snapshot"
				volURI := snapshotVolURI
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotList{
					Values: []types.SnapshotResponse{
						{
							Metadata:   types.ResourceMetadataResponse{ID: &id, Name: &name},
							Properties: types.SnapshotPropertiesResponse{Volume: &types.VolumeInfo{URI: &volURI}},
						},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"storage", "snapshot", "list", "--project-id", "proj-123", "--volume-id", snapshotVolID},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotList{}))
			},
		},
		{
			name: "--output=json emits valid JSON",
			args: []string{"storage", "snapshot", "list", "--project-id", "proj-123", "--volume-id", snapshotVolID, "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "snap-001", "my-snapshot"
				volURI := snapshotVolURI
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotList{
					Values: []types.SnapshotResponse{
						{
							Metadata:   types.ResourceMetadataResponse{ID: &id, Name: &name},
							Properties: types.SnapshotPropertiesResponse{Volume: &types.VolumeInfo{URI: &volURI}},
						},
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
			args: []string{"storage", "snapshot", "list", "--project-id", "proj-123", "--volume-id", snapshotVolID},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr: true, errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"storage", "snapshot", "list", "--project-id", "proj-123", "--volume-id", snapshotVolID},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", errorResponse(404, "Not Found", "resource not found"))
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

func TestSnapshotGetCmd(t *testing.T) {
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
				id, name := "snap-001", "my-snapshot"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", errorResponse(500, "Internal Server Error", "boom"))
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
				[]string{"storage", "snapshot", "get", "snap-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSnapshotCreateCmd(t *testing.T) {
	createArgs := []string{
		"storage", "snapshot", "create",
		"--project-id", "proj-123",
		"--name", "my-snapshot",
		"--region", "IT-BG",
		"--volume-id", snapshotVolID,
	}
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
				id, name := "snap-new", "my-snapshot"
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"storage", "snapshot", "create", "--project-id", "proj-123", "--region", "IT-BG", "--volume-id", snapshotVolID},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --volume-id",
			args:        []string{"storage", "snapshot", "create", "--project-id", "proj-123", "--name", "my-snapshot", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "volume-id",
		},
		{
			name: "API error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/snapshots", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			args: createArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Storage/snapshots", errorResponse(500, "Internal Server Error", "boom"))
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

func TestSnapshotDeleteCmd(t *testing.T) {
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
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:      "--dry-run: prints intent, does not call Delete",
			extraArgs: []string{"--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "snap-001", "my-snapshot"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr: true, errContains: "API error (status 404): Not Found",
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", errorResponse(500, "Internal Server Error", "boom"))
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
			args := []string{"storage", "snapshot", "delete", "snap-001", "--project-id", "proj-123", "--yes"}
			args = append(args, tc.extraArgs...)
			out, err := runCmdCapture(srv.Client(), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSnapshotUpdateCmd(t *testing.T) {
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
			args: []string{"storage", "snapshot", "update", "snap-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "snap-001", "my-snapshot"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success with tags covers RetaggedAs branch",
			args: []string{"storage", "snapshot", "update", "snap-001", "--project-id", "proj-123", "--tags", "env=test"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "snap-001", "my-snapshot"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, Tags: []string{"env=test"}},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "snap-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"storage", "snapshot", "update", "snap-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-GET error",
			args: []string{"storage", "snapshot", "update", "snap-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"storage", "snapshot", "update", "snap-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "snap-001", "my-snapshot"
				srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestSnapshotCreateCmd_Verbose(t *testing.T) {
	id, name := "snap-new", "my-snapshot"
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"storage", "snapshot", "create",
		"--project-id", "proj-123",
		"--name", "my-snapshot",
		"--region", "IT-BG",
		"--volume-id", snapshotVolID,
		"--verbose",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "my-snapshot") {
		t.Errorf("expected name in verbose output, got: %s", out)
	}
}

func TestSnapshotCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "snap-001", "my-snapshot"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"storage", "snapshot", "create",
		"--project-id", "proj-123",
		"--name", "my-snapshot",
		"--region", "IT-BG",
		"--volume-id", "vol-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSnapshotListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "snap-001", "my-snapshot"
	region := types.Region("IT-BG")
	state := types.StateActive
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotList{
		Values: []types.SnapshotResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.SnapshotPropertiesResponse{
					Volume: &types.VolumeInfo{URI: &volURI},
				},
				Status: types.ResourceStatus{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "snapshot", "list", "--project-id", "proj-123", "--volume-id", "vol-001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "snap-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSnapshotGetCmd_JSONOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "snap-001", "my-snapshot"
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "snapshot", "get", "snap-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestSnapshotGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "snap-001", "my-snapshot"
	uri := "/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001"
	volURI := "/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	sizeGB := int32(50)
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots/snap-001", jsonResponse(200, types.SnapshotResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
		},
		Properties: types.SnapshotPropertiesResponse{
			SizeGB: &sizeGB,
			Volume: &types.VolumeInfo{URI: &volURI},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"storage", "snapshot", "get", "snap-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "snap-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
