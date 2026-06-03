package cmd

import (
	"encoding/json"
	"strings"
	"testing"

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
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
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
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50",
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
			args:        []string{"database", "dbaas", "create", "--project-id", "proj-123", "--region", "IT-BG", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --engine-id",
			args:        []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--zone", "ITBG-1", "--flavor", "db.small", "--storage-size", "50"},
			wantErr:     true,
			errContains: "engine-id",
		},
		{
			name: "server error propagates",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"database", "dbaas", "create", "--project-id", "proj-123", "--name", "my-dbaas", "--region", "IT-BG", "--zone", "ITBG-1", "--engine-id", "postgres14", "--flavor", "db.small", "--storage-size", "50"},
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
		"--region", "IT-BG",
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
