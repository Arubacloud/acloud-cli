package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestProjectListCmd(t *testing.T) {
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
				id, name := "proj-001", "my-project"
				srv.OnGet("/projects", jsonResponse(200, types.ProjectList{
					Values: []types.ProjectResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			args: []string{"management", "project", "list"},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "proj-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects", jsonResponse(200, types.ProjectList{}))
			},
			args: []string{"management", "project", "list"},
		},
		{
			name: "--output=json emits valid JSON",
			setupSrv: func(srv *arubaTestServer) {
				id, name := "proj-001", "my-project"
				srv.OnGet("/projects", jsonResponse(200, types.ProjectList{
					Values: []types.ProjectResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			args: []string{"management", "project", "list", "--output", "json"},
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
				srv.OnGet("/projects", errorResponse(500, "Internal Server Error", "boom"))
			},
			args:        []string{"management", "project", "list"},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects", errorResponse(404, "Not Found", "resource not found"))
			},
			args:        []string{"management", "project", "list"},
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

func TestProjectGetCmd(t *testing.T) {
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
				id, name := "proj-001", "my-project"
				srv.OnGet("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "proj-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"management", "project", "get", "proj-001"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestProjectCreateCmd(t *testing.T) {
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
			args: []string{"management", "project", "create", "--name", "my-project"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "proj-new", "my-project"
				srv.OnPost("/projects", jsonResponse(200, types.ProjectResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "proj-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"management", "project", "create"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name: "server error propagates",
			args: []string{"management", "project", "create", "--name", "my-project"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"management", "project", "create", "--name", "my-project"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects", errorResponse(404, "Not Found", "resource not found"))
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

func TestProjectUpdateCmd(t *testing.T) {
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
			args: []string{"management", "project", "update", "proj-001", "--description", "new desc"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "proj-001", "my-project"
				srv.OnGet("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "proj-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing flags",
			args:        []string{"management", "project", "update", "proj-001"},
			wantErr:     true,
			errContains: "at least one of",
		},
		{
			name: "pre-Get API error",
			args: []string{"management", "project", "update", "proj-001", "--description", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "Update server error",
			args: []string{"management", "project", "update", "proj-001", "--description", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "proj-001", "my-project"
				srv.OnGet("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestProjectDeleteCmd(t *testing.T) {
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
			args: []string{"management", "project", "delete", "proj-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "proj-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"management", "project", "delete", "proj-001", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "proj-001", "my-project"
				// Only register Get; if Delete is called the harness t.Errorf on the
				// unregistered DELETE route — stronger than the old t.Fatal guard.
				srv.OnGet("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "proj-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"management", "project", "delete", "proj-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			// SDK project.Delete low-level client doesn't parse error body into resp.Error,
			// so the title is not surfaced. Tracked upstream: Arubacloud/sdk-go.
			name: "API error propagates",
			args: []string{"management", "project", "delete", "proj-001", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404)",
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

// TestProjectGetCmd_FullDetail exercises all nil-guard branches in GET detail output.
func TestProjectGetCmd_FullDetail(t *testing.T) {
	t.Run("detail with all optional fields covers nil-guards", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "proj-001", "my-project"
		createdBy := "user@example.com"
		updatedBy := "admin@example.com"
		desc := "My test project"
		ts := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
		srv.OnGet("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:           &id,
				Name:         &name,
				Tags:         []string{"env=test"},
				CreationDate: &ts,
				CreatedBy:    &createdBy,
				UpdateDate:   &ts,
				UpdatedBy:    &updatedBy,
			},
			Properties: types.ProjectPropertiesResponse{
				Description: &desc,
			},
		}))
		out, err := runCmdCapture(srv.Client(), []string{"management", "project", "get", "proj-001"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "proj-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "user@example.com") {
			t.Errorf("expected createdBy in output, got: %s", out)
		}
		if !strings.Contains(out, "admin@example.com") {
			t.Errorf("expected updatedBy in output, got: %s", out)
		}
		if !strings.Contains(out, "My test project") {
			t.Errorf("expected description in output, got: %s", out)
		}
		if !strings.Contains(out, "env=test") {
			t.Errorf("expected tags in output, got: %s", out)
		}
	})

	t.Run("--output json hits early return", func(t *testing.T) {
		srv := newArubaTestServer(t)
		id, name := "proj-001", "my-project"
		srv.OnGet("/projects/proj-001", jsonResponse(200, types.ProjectResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
		out, err := runCmdCapture(srv.Client(), []string{"management", "project", "get", "proj-001", "--output", "json"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var result map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
			t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
		}
	})
}

// TestProjectListCmd_AllOptionalFields exercises nil-guard branches in LIST output.
func TestProjectListCmd_AllOptionalFields(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "proj-001", "my-project"
	desc := "My test project"
	srv.OnGet("/projects", jsonResponse(200, types.ProjectList{
		Values: []types.ProjectResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:   &id,
					Name: &name,
					Tags: []string{"env=test"},
				},
				Properties: types.ProjectPropertiesResponse{
					Description: &desc,
					Default:     true,
				},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"management", "project", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "proj-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestProjectListCmd_WithProjectData(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "proj-001", "my-project"
	srv.OnGet("/projects", jsonResponse(200, types.ProjectList{
		Values: []types.ProjectResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"management", "project", "list"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "proj-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
