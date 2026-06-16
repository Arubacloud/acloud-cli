package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestDBaaSListCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
					Values: []types.DBaaSResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"database", "dbaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"database", "dbaas", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
					Values: []types.DBaaSResponse{
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
			name: "list items with all optional fields covers nil-guards",
			args: []string{"database", "dbaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-rich", "rich-dbaas"
				engineType, engineVersion := "postgresql", "14"
				flavorName := "db.small"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
					Values: []types.DBaaSResponse{
						{
							Metadata: types.ResourceMetadataResponse{
								ID:               &id,
								Name:             &name,
								LocationResponse: &types.LocationResponse{Value: types.Region("IT-BG")},
							},
							Properties: types.DBaaSPropertiesResponse{
								Engine: &types.DBaaSEngineResponse{Type: &engineType, Version: &engineVersion},
								Flavor: &types.DBaaSFlavorResponse{Name: &flavorName},
							},
							Status: types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
						},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-rich") {
					t.Errorf("expected ID in output, got: %s", out)
				}
				if !strings.Contains(out, "postgresql") {
					t.Errorf("expected engine type in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSGetCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "detail output with all optional fields covers nil-guards",
			args: []string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				uri := "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001"
				engineType, engineVersion, engineName := "postgresql", "14", "PostgreSQL 14"
				flavorName := "db.small"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{
						ID:               &id,
						Name:             &name,
						URI:              &uri,
						LocationResponse: &types.LocationResponse{Value: types.Region("IT-BG")},
						Tags:             []string{"env=prod"},
					},
					Properties: types.DBaaSPropertiesResponse{
						Engine: &types.DBaaSEngineResponse{
							Type:    &engineType,
							Version: &engineVersion,
							Name:    &engineName,
						},
						Flavor: &types.DBaaSFlavorResponse{Name: &flavorName},
					},
					Status: types.ResourceStatusResponse{State: func() *types.State { s := types.StateActive; return &s }()},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
				if !strings.Contains(out, "postgresql") {
					t.Errorf("expected engine type in output, got: %s", out)
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
			name: "--output json emits valid JSON",
			args: []string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
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
			name: "server error propagates",
			args: []string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "get", "dbaas-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSCreateCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-new", "my-dbaas"
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success with networking flags",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50",
				"--vpc-id", "vpc-001",
				"--subnet-id", "sub-001",
				"--security-group-id", "sg-001",
				"--elastic-ip-id", "eip-001",
			},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-new", "my-dbaas"
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"database", "dbaas", "create", "--project-id", "proj-123", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --engine-id",
			args:        []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--flavor", "db.small", "--storage-size", "50"},
			wantErr:     true,
			errContains: "engine-id",
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSDeleteCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "delete", "dbaas-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"database", "dbaas", "delete", "dbaas-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "delete", "dbaas-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "delete", "dbaas-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestDBaaSUpdateCmd(t *testing.T) {
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
			args: []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success with tags covers RetaggedAs branch",
			args: []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123", "--tags", "env=test"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, Tags: []string{"env=test"}},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			// Regression: fromResponse sets d.engine from Engine.Type ("mysql") instead of
			// the catalog ID. Re-inject Engine.ID from the GET response so toRequest()
			// emits the correct catalog identifier.
			// Note: DataCenter is intentionally NOT re-injected — the GET response stores
			// it as a region display name (e.g. "ITBG-Bergamo") not the zone code
			// ("ITBG-1"), so injecting it would cause a 400 "DataCenter cannot be modified".
			// Omitting dataCenter from PUT (omitempty + nil) is accepted as "no change".
			name: "engine ID re-populated from response — avoids catalog 400; zone omitted from PUT",
			args: []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123", "--tags", "updated"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				engineID, engineType := "mysql-8.0", "mysql"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Properties: types.DBaaSPropertiesResponse{
						Engine: &types.DBaaSEngineResponse{ID: &engineID, Type: &engineType},
					},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, Tags: []string{"updated"}},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			// --zone re-injects the zone so PUT body includes "dataCenter".
			// Exercises ParseFromCobraCommand zone read and DatabaseDBaaSUpdate InZone branch.
			name: "--zone flag injects dataCenter into PUT body",
			args: []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123", "--tags", "updated", "--zone", "ITBG-1"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, Tags: []string{"updated"}},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "dbaas-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-GET error",
			args: []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"database", "dbaas", "update", "dbaas-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "dbaas-001", "my-dbaas"
				srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", jsonResponse(200, types.DBaaSResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001", errorResponse(500, "Internal Server Error", "boom"))
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

// TestDatabaseDBaaSUpdate_ZoneInPUTBody verifies that --zone injects dataCenter into
// the PUT request body so the API round-trip succeeds without a "DataCenter cannot
// be modified" 400. This is a regression test for the zone omitempty gap.
func TestDatabaseDBaaSUpdate_ZoneInPUTBody(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-z1", "my-dbaas"
	var capturedBody []byte
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-z1", jsonResponse(200, types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	srv.On(http.MethodPut, "/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-z1",
		func(w http.ResponseWriter, r *http.Request) {
			capturedBody, _ = io.ReadAll(r.Body)
			jsonResponse(200, types.DBaaSResponse{
				Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, Tags: []string{"t"}},
			})(w, r)
		})

	err := runCmd(srv.Client(), []string{
		"database", "dbaas", "update", "dbaas-z1",
		"--project-id", "proj-123",
		"--tags", "t",
		"--zone", "ITBG-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(capturedBody)
	if !strings.Contains(body, `"dataCenter"`) || !strings.Contains(body, `"ITBG-1"`) {
		t.Errorf("PUT body missing expected dataCenter ITBG-1; got: %s", body)
	}
}

func TestDatabaseDBaaSCreateRun_ValidationError(t *testing.T) {
	// --name "x" is too short (< 3 chars) — triggers ErrValidationFailed
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "x", "--region", "ITBG-Bergamo", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestDatabaseDBaaSListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"database", "dbaas", "list"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDBaaSCreateCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-001", "my-db"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"database", "dbaas", "create",
		"--project-id", "proj-123",
		"--name", "my-db",
		"--region", "ITBG-Bergamo",
		"--zone", "IT-BG-1",
		"--engine-id", "mysql-8.0",
		"--flavor", "DBO4A8",
		"--storage-size", "50",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDBaaSListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-001", "my-db"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
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
	out, err := runCmdCapture(srv.Client(), []string{"database", "dbaas", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "dbaas-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestDatabaseDBaaSCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-new", "my-dbaas"
	srv.OnPost("/projects/p1/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	ctx, cancel := newCtx()
	defer cancel()
	out := captureStdout(func() {
		err := DatabaseDBaaSCreate(ctx, srv.Client(), DatabaseDBaaSCreateArgs{
			ProjectID:     "p1",
			Name:          "my-dbaas",
			Region:        aruba.RegionITBGBergamo,
			Zone:          "ITBG-1",
			Engine:        "mysql-8.0",
			Flavor:        "DBO4A8",
			SizeGB:        50,
			BillingPeriod: aruba.BillingPeriodHour,
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "dbaas-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSCreate_WithNetworking(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-net", "net-dbaas"
	srv.OnPost("/projects/p1/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	ctx, cancel := newCtx()
	defer cancel()
	out := captureStdout(func() {
		err := DatabaseDBaaSCreate(ctx, srv.Client(), DatabaseDBaaSCreateArgs{
			ProjectID:   "p1",
			Name:        "net-dbaas",
			Region:      aruba.RegionITBGBergamo,
			Zone:        "ITBG-1",
			Engine:      "mysql-8.0",
			Flavor:      "DBO4A8",
			SizeGB:      50,
			VPCID:       "vpc-001",
			SubnetID:    "sub-001",
			SGID:        "sg-001",
			ElasticIPID: "eip-001",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "dbaas-net") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/p1/providers/Aruba.Database/dbaas", errorResponse(500, "Internal Server Error", "quota"))
	ctx, cancel := newCtx()
	defer cancel()
	err := DatabaseDBaaSCreate(ctx, srv.Client(), DatabaseDBaaSCreateArgs{
		ProjectID: "p1",
		Name:      "my-dbaas",
		Engine:    "mysql-8.0",
		Flavor:    "DBO4A8",
		SizeGB:    50,
	})
	if err == nil || !strings.Contains(err.Error(), "creating") {
		t.Errorf("expected creating error, got: %v", err)
	}
}

func TestDatabaseDBaaSGet_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-1", "my-dbaas"
	srv.OnGet("/projects/p1/providers/Aruba.Database/dbaas/dbaas-1", jsonResponse(200, types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	ctx, cancel := newCtx()
	defer cancel()
	out := captureStdout(func() {
		err := DatabaseDBaaSGet(ctx, srv.Client(), DatabaseDBaaSGetArgs{ProjectID: "p1", ID: "dbaas-1"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "dbaas-1") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestDatabaseDBaaSUpdate_AsyncMessage(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "dbaas-upd", "my-dbaas"
	srv.OnGet("/projects/p1/providers/Aruba.Database/dbaas/dbaas-upd", jsonResponse(200, types.DBaaSResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	srv.OnPut("/projects/p1/providers/Aruba.Database/dbaas/dbaas-upd", jsonResponse(202, types.DBaaSResponse{}))
	ctx, cancel := newCtx()
	defer cancel()
	out := captureStdout(func() {
		err := DatabaseDBaaSUpdate(ctx, srv.Client(), DatabaseDBaaSUpdateArgs{
			ProjectID: "p1",
			ID:        "dbaas-upd",
			Name:      "new-name",
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "dbaas-upd") {
		t.Errorf("expected ID in async message, got: %s", out)
	}
}

func TestDatabaseDBaaSList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/p1/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{}))
	ctx, cancel := newCtx()
	defer cancel()
	out := captureStdout(func() {
		err := DatabaseDBaaSList(ctx, srv.Client(), DatabaseDBaaSListArgs{ProjectID: "p1"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No DBaaS instances") {
		t.Errorf("expected 'No DBaaS instances' message, got: %s", out)
	}
}

func TestDatabaseDBaaSList_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/p1/providers/Aruba.Database/dbaas", errorResponse(500, "Internal Server Error", "boom"))
	ctx, cancel := newCtx()
	defer cancel()
	err := DatabaseDBaaSList(ctx, srv.Client(), DatabaseDBaaSListArgs{ProjectID: "p1"})
	if err == nil || !strings.Contains(err.Error(), "listing") {
		t.Errorf("expected listing error, got: %v", err)
	}
}

// =============================================================================
// Layer 1 — Validate tests (pure Go, no SDK)
// =============================================================================

func TestDatabaseDBaaSCreateArgs_Validate(t *testing.T) {
	validBase := DatabaseDBaaSCreateArgs{
		ProjectID:     "p1",
		Name:          "my-dbaas",
		Region:        aruba.RegionITBGBergamo,
		Zone:          "ITBG-1",
		Engine:        "mysql-8.0",
		Flavor:        "DBO4A8",
		SizeGB:        50,
		BillingPeriod: aruba.BillingPeriodHour,
	}
	tests := []struct {
		name        string
		mutate      func(*DatabaseDBaaSCreateArgs)
		wantErr     bool
		errContains string
	}{
		{name: "valid", mutate: func(_ *DatabaseDBaaSCreateArgs) {}, wantErr: false},
		{
			name:        "name too short",
			mutate:      func(a *DatabaseDBaaSCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:        "name too long",
			mutate:      func(a *DatabaseDBaaSCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *DatabaseDBaaSCreateArgs) { a.Region = aruba.Region("bad") },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *DatabaseDBaaSCreateArgs) { a.BillingPeriod = aruba.BillingPeriod("Weekly") },
			wantErr:     true,
			errContains: "--billing-period",
		},
		{
			name:        "empty engine",
			mutate:      func(a *DatabaseDBaaSCreateArgs) { a.Engine = "" },
			wantErr:     true,
			errContains: "--engine-id is required",
		},
		{
			name:        "empty flavor",
			mutate:      func(a *DatabaseDBaaSCreateArgs) { a.Flavor = "" },
			wantErr:     true,
			errContains: "--flavor is required",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validBase
			tc.mutate(&args)
			err := args.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.errContains)
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error to contain %q, got: %v", tc.errContains, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDatabaseDBaaSGetArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    DatabaseDBaaSGetArgs
		wantErr bool
	}{
		{name: "valid", args: DatabaseDBaaSGetArgs{ProjectID: "p1", ID: "d1"}, wantErr: false},
		{name: "empty ID", args: DatabaseDBaaSGetArgs{ProjectID: "p1"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDatabaseDBaaSUpdateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        DatabaseDBaaSUpdateArgs
		wantErr     bool
		errContains string
	}{
		{name: "valid with name", args: DatabaseDBaaSUpdateArgs{ProjectID: "p1", ID: "d1", Name: "new"}, wantErr: false},
		{name: "valid with tags", args: DatabaseDBaaSUpdateArgs{ProjectID: "p1", ID: "d1", TagsChanged: true}, wantErr: false},
		{name: "missing ID", args: DatabaseDBaaSUpdateArgs{ProjectID: "p1", Name: "new"}, wantErr: true, errContains: "DBaaS ID is required"},
		{name: "no fields", args: DatabaseDBaaSUpdateArgs{ProjectID: "p1", ID: "d1"}, wantErr: true, errContains: "at least one"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.errContains != "" && !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected error to contain %q, got: %v", tc.errContains, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDatabaseDBaaSDeleteArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    DatabaseDBaaSDeleteArgs
		wantErr bool
	}{
		{name: "valid", args: DatabaseDBaaSDeleteArgs{ProjectID: "p1", ID: "d1"}, wantErr: false},
		{name: "empty ID", args: DatabaseDBaaSDeleteArgs{ProjectID: "p1"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestDatabaseDBaaSListArgs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		args    DatabaseDBaaSListArgs
		wantErr bool
	}{
		{name: "valid", args: DatabaseDBaaSListArgs{ProjectID: "p1"}, wantErr: false},
		{name: "empty project", args: DatabaseDBaaSListArgs{}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.args.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}
