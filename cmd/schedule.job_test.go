package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestJobListCmd(t *testing.T) {
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
			args: []string{"schedule", "job", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
					Values: []types.JobResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
					Values: []types.JobResponse{
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
			name: "--output yaml emits valid YAML",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123", "--output", "yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
					Values: []types.JobResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if strings.TrimSpace(out) == "{}" || strings.TrimSpace(out) == "" {
					t.Errorf("yaml output is empty or {}, got: %s", out)
				}
			},
		},
		{
			name: "--output table-json emits valid JSON array",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123", "--output", "table-json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
					Values: []types.JobResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				var rows []map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &rows); err != nil {
					t.Errorf("table-json output is not a valid JSON array: %v\noutput: %s", err, out)
				}
				if len(rows) == 0 {
					t.Errorf("expected at least one row in table-json output, got none")
				}
			},
		},
		{
			name: "--output table-yaml emits non-empty YAML",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123", "--output", "table-yaml"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
					Values: []types.JobResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if strings.TrimSpace(out) == "" {
					t.Errorf("table-yaml output is empty, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", errorResponse(404, "Not Found", "resource not found"))
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

func TestJobGetCmd(t *testing.T) {
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
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"schedule", "job", "get", "job-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestJobCreateCmd(t *testing.T) {
	baseShotArgs := []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "ITBG-Bergamo",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
		"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
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
			name: "success OneShot",
			args: baseShotArgs,
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-new", "my-job"
				srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        removeFlag(baseShotArgs, "--name", "my-job"),
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --job-type",
			args:        removeFlag(baseShotArgs, "--job-type", "OneShot"),
			wantErr:     true,
			errContains: "job-type",
		},
		{
			name:        "missing required flag --step-resource-uri",
			args:        removeFlag(baseShotArgs, "--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001"),
			wantErr:     true,
			errContains: "step-resource-uri",
		},
		{
			name: "server error propagates",
			args: baseShotArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: baseShotArgs,
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", errorResponse(404, "Not Found", "resource not found"))
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

func TestJobDeleteCmd(t *testing.T) {
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
			args: []string{"schedule", "job", "delete", "job-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"schedule", "job", "delete", "job-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"schedule", "job", "delete", "job-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"schedule", "job", "delete", "job-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestJobUpdateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success --name",
			args: []string{"schedule", "job", "update", "job-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"schedule", "job", "update", "job-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-GET error",
			args: []string{"schedule", "job", "update", "job-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update error",
			args: []string{"schedule", "job", "update", "job-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestJobCreateCmd_Recurring(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupSrv    func(*arubaTestServer)
		wantErr     bool
		errContains string
	}{
		{
			name: "success Recurring",
			args: []string{
				"schedule", "job", "create",
				"--project-id", "proj-123",
				"--name", "my-recurring-job",
				"--region", "ITBG-Bergamo",
				"--job-type", "Recurring",
				"--cron", "0 * * * *",
				"--execute-until", "2027-01-01T00:00:00Z",
				"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
			},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-new", "my-recurring-job"
				srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
		},
		{
			name: "Recurring missing --cron",
			args: []string{
				"schedule", "job", "create",
				"--project-id", "proj-123",
				"--name", "my-job",
				"--region", "ITBG-Bergamo",
				"--job-type", "Recurring",
				"--execute-until", "2027-01-01T00:00:00Z",
				"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
			},
			wantErr:     true,
			errContains: "cron",
		},
		{
			name: "Recurring missing --execute-until",
			args: []string{
				"schedule", "job", "create",
				"--project-id", "proj-123",
				"--name", "my-job",
				"--region", "ITBG-Bergamo",
				"--job-type", "Recurring",
				"--cron", "0 * * * *",
				"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
			},
			wantErr:     true,
			errContains: "execute-until",
		},
		{
			name: "invalid --job-type",
			args: []string{
				"schedule", "job", "create",
				"--project-id", "proj-123",
				"--name", "my-job",
				"--region", "ITBG-Bergamo",
				"--job-type", "Invalid",
				"--schedule-at", "2026-06-01T10:00:00Z",
				"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
			},
			wantErr:     true,
			errContains: "job-type",
		},
		{
			name: "OneShot missing --schedule-at",
			args: []string{
				"schedule", "job", "create",
				"--project-id", "proj-123",
				"--name", "my-job",
				"--region", "ITBG-Bergamo",
				"--job-type", "OneShot",
				"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
			},
			wantErr:     true,
			errContains: "schedule-at",
		},
		{
			name: "success with step resource",
			args: []string{
				"schedule", "job", "create",
				"--project-id", "proj-123",
				"--name", "my-job",
				"--region", "ITBG-Bergamo",
				"--job-type", "OneShot",
				"--schedule-at", "2026-06-01T10:00:00Z",
				"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
				"--step-action-uri", "poweroff",
			},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-new", "my-job"
				srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newArubaTestServer(t)
			if tc.setupSrv != nil {
				tc.setupSrv(srv)
			}
			_, err := runCmdCapture(srv.Client(), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestJobCreateCmd_Disabled(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	err := runCmd(srv.Client(), []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "ITBG-Bergamo",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
		"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
		"--enabled=false",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJobCreateCmd_InvalidScheduleAt(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "ITBG-Bergamo",
		"--job-type", "OneShot",
		"--schedule-at", "not-a-date",
		"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
	})
	if err == nil {
		t.Fatal("expected error for invalid --schedule-at, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --schedule-at") {
		t.Errorf("expected 'invalid --schedule-at' in error, got: %v", err)
	}
}

func TestJobCreateCmd_InvalidExecuteUntil(t *testing.T) {
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "ITBG-Bergamo",
		"--job-type", "Recurring",
		"--cron", "0 0 * * *",
		"--execute-until", "not-a-date",
		"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
	})
	if err == nil {
		t.Fatal("expected error for invalid --execute-until, got nil")
	}
	if !strings.Contains(err.Error(), "invalid --execute-until") {
		t.Errorf("expected 'invalid --execute-until' in error, got: %v", err)
	}
}

func TestJobCreateCmd_WithEnabledAndLocation(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	region := types.Region("ITBG-Bergamo")
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Properties: types.JobPropertiesResponse{
			Enabled: true,
			JobType: types.JobTypeOneShot,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "ITBG-Bergamo",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
		"--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJobListCmd_WithEnabledAndLocation(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	region := types.Region("ITBG-Bergamo")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
		Values: []types.JobResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.JobPropertiesResponse{
					Enabled: true,
					JobType: types.JobTypeOneShot,
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"schedule", "job", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestJobGetCmd_JSONOutput(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"schedule", "job", "get", "job-001", "--project-id", "proj-123", "--output", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "{") {
		t.Errorf("expected JSON output, got: %s", out)
	}
}

func TestJobGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	uri := "/projects/proj-123/providers/Aruba.Schedule/jobs/job-001"
	region := types.Region("ITBG-Bergamo")
	state := types.StateActive
	createdBy := "test-user@example.com"
	schedAt := "2026-06-01T10:00:00Z"
	execUntil := "2027-01-01T00:00:00Z"
	cronExpr := "0 0 * * *"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.JobPropertiesResponse{
			Enabled:      true,
			JobType:      types.JobTypeRecurring,
			ScheduleAt:   &schedAt,
			ExecuteUntil: &execUntil,
			Cron:         &cronExpr,
		},
		Status: types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"schedule", "job", "get", "job-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestJobUpdateCmd_WithEnabledAndTags(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	updID, updName := "job-001", "my-job"
	srv.OnPut("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &updID,
			Name: &updName,
			Tags: []string{"env=prod"},
		},
		Properties: types.JobPropertiesResponse{Enabled: true},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"schedule", "job", "update", "job-001",
		"--project-id", "proj-123",
		"--enabled",
		"--tags", "env=prod",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobCreateRun_ValidationError(t *testing.T) {
	// --name "x" is too short (< 3 chars) — triggers ErrValidationFailed
	srv := newArubaTestServer(t)
	err := runCmd(srv.Client(), []string{"schedule", "job", "create", "--project-id", "proj-123", "--name", "x", "--region", "ITBG-Bergamo", "--job-type", "OneShot", "--schedule-at", "2026-06-01T10:00:00Z", "--step-resource-uri", "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "checking args") {
		t.Errorf("expected 'checking args', got: %v", err)
	}
}

func TestScheduleJobListRun_NoProjectID(t *testing.T) {
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
	err := runCmd(srv.Client(), []string{"schedule", "job", "list"})
	if err == nil {
		t.Fatal("expected error")
	}
}

// =============================================================================
// Layer 1 — Validate tests (pure Go, no SDK, no HTTP)
// =============================================================================

func validScheduleJobCreateArgs() ScheduleJobCreateArgs {
	return ScheduleJobCreateArgs{
		ProjectID:       "proj-123",
		Name:            "my-job",
		Region:          aruba.RegionITBGBergamo,
		JobType:         aruba.JobTypeOneShot,
		Enabled:         true,
		Tags:            []string{},
		ShotTime:        "2026-06-01T10:00:00Z",
		StepResourceURI: "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001",
	}
}

func TestJobCreateCmd_RequestBodyContainsSteps(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	stepURI := "/projects/proj-123/providers/Aruba.Compute/cloudServers/cs-001"
	stepAction := "/actions/powerOff"

	var capturedBody []byte
	srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		jsonResponse(200, types.JobResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		})(w, r)
	})

	err := runCmd(srv.Client(), []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "ITBG-Bergamo",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
		"--step-resource-uri", stepURI,
		"--step-action-uri", stepAction,
		"--step-http-verb", "POST",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := string(capturedBody)
	if !strings.Contains(body, "\"steps\"") {
		t.Fatalf("request body missing steps array: %s", body)
	}
	if !strings.Contains(body, stepURI) {
		t.Fatalf("request body missing step resource URI %q: %s", stepURI, body)
	}
	if !strings.Contains(body, stepAction) {
		t.Fatalf("request body missing step action URI %q: %s", stepAction, body)
	}
}

func TestScheduleJobCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*ScheduleJobCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid OneShot args",
			wantErr: false,
		},
		{
			name: "valid Recurring args",
			mutate: func(a *ScheduleJobCreateArgs) {
				a.JobType = aruba.JobTypeRecurring
				a.ShotTime = ""
				a.CronExpr = "0 * * * *"
				a.EndTime = "2027-01-01T00:00:00Z"
			},
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *ScheduleJobCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *ScheduleJobCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *ScheduleJobCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:    "name maximum length 64",
			mutate:  func(a *ScheduleJobCreateArgs) { a.Name = strings.Repeat("x", 64) },
			wantErr: false,
		},
		{
			name:        "invalid region",
			mutate:      func(a *ScheduleJobCreateArgs) { a.Region = aruba.Region("ZZ-Invalid") },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid job-type",
			mutate:      func(a *ScheduleJobCreateArgs) { a.JobType = aruba.JobType("WeeklyBatch") },
			wantErr:     true,
			errContains: "--job-type",
		},
		// OneShot / Recurring mutual-exclusion checks.
		{
			name: "OneShot missing ShotTime",
			mutate: func(a *ScheduleJobCreateArgs) {
				a.JobType = aruba.JobTypeOneShot
				a.ShotTime = ""
			},
			wantErr:     true,
			errContains: "--schedule-at is required for OneShot",
		},
		{
			name: "Recurring missing CronExpr",
			mutate: func(a *ScheduleJobCreateArgs) {
				a.JobType = aruba.JobTypeRecurring
				a.ShotTime = ""
				a.CronExpr = ""
				a.EndTime = "2027-01-01T00:00:00Z"
			},
			wantErr:     true,
			errContains: "--cron is required for Recurring",
		},
		{
			name: "Recurring missing EndTime",
			mutate: func(a *ScheduleJobCreateArgs) {
				a.JobType = aruba.JobTypeRecurring
				a.ShotTime = ""
				a.CronExpr = "0 * * * *"
				a.EndTime = ""
			},
			wantErr:     true,
			errContains: "--execute-until is required for Recurring",
		},
		{
			name: "Recurring missing both CronExpr and EndTime",
			mutate: func(a *ScheduleJobCreateArgs) {
				a.JobType = aruba.JobTypeRecurring
				a.ShotTime = ""
				a.CronExpr = ""
				a.EndTime = ""
			},
			wantErr:     true,
			errContains: "--cron is required for Recurring",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validScheduleJobCreateArgs()
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

func TestScheduleJobCreateArgs_Validate_WrappedByConstructor(t *testing.T) {
	args := validScheduleJobCreateArgs()
	args.Name = "x" // too short — triggers ErrValidationFailed when wrapped by constructor
	err := args.Validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	wrapped := errors.Join(ErrValidationFailed, err)
	if !errors.Is(wrapped, ErrValidationFailed) {
		t.Errorf("expected ErrValidationFailed in chain, got: %v", wrapped)
	}
}

func TestScheduleJobUpdateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		args        ScheduleJobUpdateArgs
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid with name",
			args:    ScheduleJobUpdateArgs{ProjectID: "proj-123", ID: "job-001", Name: "new-name"},
			wantErr: false,
		},
		{
			name:    "valid with enabled set",
			args:    ScheduleJobUpdateArgs{ProjectID: "proj-123", ID: "job-001", EnabledSet: true},
			wantErr: false,
		},
		{
			name:    "valid with tags changed",
			args:    ScheduleJobUpdateArgs{ProjectID: "proj-123", ID: "job-001", TagsChanged: true},
			wantErr: false,
		},
		{
			name:        "no fields changed",
			args:        ScheduleJobUpdateArgs{ProjectID: "proj-123", ID: "job-001"},
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

func TestScheduleJobCreate_HappyPath_OneShot(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-new", "my-job"
	srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ScheduleJobCreate(context.Background(), srv.Client(), validScheduleJobCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobCreate_HappyPath_Recurring(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-recur", "my-recurring-job"
	srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	args := validScheduleJobCreateArgs()
	args.Name = "my-recurring-job"
	args.JobType = aruba.JobTypeRecurring
	args.ShotTime = ""
	args.CronExpr = "0 * * * *"
	args.EndTime = "2027-01-01T00:00:00Z"

	out := captureStdout(func() {
		err := ScheduleJobCreate(context.Background(), srv.Client(), args)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-recur") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Schedule/jobs", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := ScheduleJobCreate(context.Background(), srv.Client(), validScheduleJobCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating job") {
		t.Errorf("error %q does not contain 'creating job'", err.Error())
	}
}

func TestScheduleJobList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
		Values: []types.JobResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := ScheduleJobList(context.Background(), srv.Client(), ScheduleJobListArgs{ProjectID: "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{}))

	out := captureStdout(func() {
		err := ScheduleJobList(context.Background(), srv.Client(), ScheduleJobListArgs{ProjectID: "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No jobs found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestScheduleJobList_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", errorResponse(500, "Internal Server Error", "boom"))

	err := ScheduleJobList(context.Background(), srv.Client(), ScheduleJobListArgs{ProjectID: "proj-123"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "listing jobs") {
		t.Errorf("error %q does not contain 'listing jobs'", err.Error())
	}
}

func TestScheduleJobGet_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ScheduleJobGet(context.Background(), srv.Client(), ScheduleJobGetArgs{
			ProjectID: "proj-123",
			ID:        "job-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobGet_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(404, "Not Found", "resource not found"))

	err := ScheduleJobGet(context.Background(), srv.Client(), ScheduleJobGetArgs{
		ProjectID: "proj-123",
		ID:        "job-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "getting job") {
		t.Errorf("error %q does not contain 'getting job'", err.Error())
	}
}

func TestScheduleJobDelete_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, nil))

	out := captureStdout(func() {
		err := ScheduleJobDelete(context.Background(), srv.Client(), ScheduleJobDeleteArgs{
			ProjectID: "proj-123",
			ID:        "job-001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobDelete_DryRun(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	// Only register Get; if Delete is called the harness t.Errorf on the unregistered route.
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := ScheduleJobDelete(context.Background(), srv.Client(), ScheduleJobDeleteArgs{
			ProjectID: "proj-123",
			ID:        "job-001",
			DryRun:    true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in dry-run output, got: %s", out)
	}
}

func TestScheduleJobDelete_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnDelete("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(500, "Internal Server Error", "in use"))

	err := ScheduleJobDelete(context.Background(), srv.Client(), ScheduleJobDeleteArgs{
		ProjectID: "proj-123",
		ID:        "job-001",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "deleting job") {
		t.Errorf("error %q does not contain 'deleting job'", err.Error())
	}
}

func TestScheduleJobUpdate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	newName := "renamed-job"
	srv.OnPut("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", jsonResponse(200, types.JobResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &newName},
	}))

	out := captureStdout(func() {
		err := ScheduleJobUpdate(context.Background(), srv.Client(), ScheduleJobUpdateArgs{
			ProjectID: "proj-123",
			ID:        "job-001",
			Name:      "renamed-job",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "job-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestScheduleJobUpdate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs/job-001", errorResponse(404, "Not Found", "not found"))

	err := ScheduleJobUpdate(context.Background(), srv.Client(), ScheduleJobUpdateArgs{
		ProjectID: "proj-123",
		ID:        "job-001",
		Name:      "x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "getting job") {
		t.Errorf("error %q does not contain 'getting job'", err.Error())
	}
}
