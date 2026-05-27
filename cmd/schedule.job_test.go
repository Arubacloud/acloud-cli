package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

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
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobList{
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
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobList{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"schedule", "job", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "job-001", "my-job"
				srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobList{
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
		"--region", "IT-BG",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
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
				"--region", "IT-BG",
				"--job-type", "Recurring",
				"--cron", "0 * * * *",
				"--execute-until", "2027-01-01T00:00:00Z",
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
				"--region", "IT-BG",
				"--job-type", "Recurring",
				"--execute-until", "2027-01-01T00:00:00Z",
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
				"--region", "IT-BG",
				"--job-type", "Recurring",
				"--cron", "0 * * * *",
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
				"--region", "IT-BG",
				"--job-type", "Invalid",
				"--schedule-at", "2026-06-01T10:00:00Z",
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
				"--region", "IT-BG",
				"--job-type", "OneShot",
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
				"--region", "IT-BG",
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
		"--region", "IT-BG",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
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
		"--region", "IT-BG",
		"--job-type", "OneShot",
		"--schedule-at", "not-a-date",
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
		"--region", "IT-BG",
		"--job-type", "Recurring",
		"--cron", "0 0 * * *",
		"--execute-until", "not-a-date",
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
	region := types.Region("IT-BG")
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
		Status: types.ResourceStatus{State: &state},
	}))
	err := runCmd(srv.Client(), []string{
		"schedule", "job", "create",
		"--project-id", "proj-123",
		"--name", "my-job",
		"--region", "IT-BG",
		"--job-type", "OneShot",
		"--schedule-at", "2026-06-01T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJobListCmd_WithEnabledAndLocation(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "job-001", "my-job"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobList{
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
				Status: types.ResourceStatus{State: &state},
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
	region := types.Region("IT-BG")
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
		Status: types.ResourceStatus{State: &state},
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
