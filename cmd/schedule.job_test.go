package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestJobListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockJobsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with results",
			setupMock: func(m *mockJobsClient) {
				id, name := "job-001", "my-job"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.JobList], error) {
					return &types.Response[types.JobList]{
						StatusCode: 200,
						Data: &types.JobList{
							Values: []types.JobResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockJobsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.JobList], error) {
					return &types.Response[types.JobList]{StatusCode: 200, Data: &types.JobList{}}, nil
				}
			},
		},
		{
			name: "--output=json emits valid JSON",
			setupMock: func(m *mockJobsClient) {
				id, name := "job-001", "my-job"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.JobList], error) {
					return &types.Response[types.JobList]{
						StatusCode: 200,
						Data: &types.JobList{
							Values: []types.JobResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
							},
						},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				var result map[string]any
				if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &result); err != nil {
					t.Errorf("output is not valid JSON: %v\noutput: %s", err, out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockJobsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.JobList], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockJobsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.JobList], error) {
					return &types.Response[types.JobList]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockJobsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"schedule", "job", "list", "--project-id", "proj-123"}
			if tc.name == "--output=json emits valid JSON" {
				args = append(args, "--output", "json")
			}
			out, err := runCmdCapture(newMockClient(withSchedule(&mockScheduleClient{jobsClient: m})), args)
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
		setupMock   func(*mockJobsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			setupMock: func(m *mockJobsClient) {
				id, name := "job-001", "my-job"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return &types.Response[types.JobResponse]{
						StatusCode: 200,
						Data:       &types.JobResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockJobsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockJobsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return &types.Response[types.JobResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockJobsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withSchedule(&mockScheduleClient{jobsClient: m})),
				[]string{"schedule", "job", "get", "job-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestJobCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockJobsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success",
			args: []string{"schedule", "job", "create", "--project-id", "proj-123", "--name", "my-job", "--region", "IT-BG", "--job-type", "OneShot", "--schedule-at", "2026-06-01T10:00:00Z"},
			setupMock: func(m *mockJobsClient) {
				id, name := "job-new", "my-job"
				m.createFn = func(_ context.Context, _ string, _ types.JobRequest, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return &types.Response[types.JobResponse]{
						StatusCode: 200,
						Data:       &types.JobResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-new") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"schedule", "job", "create", "--project-id", "proj-123", "--region", "IT-BG", "--job-type", "OneShot", "--schedule-at", "2026-06-01T10:00:00Z"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --job-type",
			args:        []string{"schedule", "job", "create", "--project-id", "proj-123", "--name", "my-job", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "job-type",
		},
		{
			name: "SDK error propagates",
			args: []string{"schedule", "job", "create", "--project-id", "proj-123", "--name", "my-job", "--region", "IT-BG", "--job-type", "OneShot", "--schedule-at", "2026-06-01T10:00:00Z"},
			setupMock: func(m *mockJobsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.JobRequest, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return nil, fmt.Errorf("quota exceeded")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"schedule", "job", "create", "--project-id", "proj-123", "--name", "my-job", "--region", "IT-BG", "--job-type", "OneShot", "--schedule-at", "2026-06-01T10:00:00Z"},
			setupMock: func(m *mockJobsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.JobRequest, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return &types.Response[types.JobResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockJobsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			out, err := runCmdCapture(newMockClient(withSchedule(&mockScheduleClient{jobsClient: m})), tc.args)
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
		setupMock   func(*mockJobsClient)
		wantErr     bool
		errContains string
		assertOut   func(*testing.T, string)
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockJobsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			setupMock: func(m *mockJobsClient) {
				id, name := "job-001", "my-job"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.JobResponse], error) {
					return &types.Response[types.JobResponse]{StatusCode: 200, Data: &types.JobResponse{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}}}, nil
				}
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					t.Fatal("Delete must not be called in --dry-run mode")
					return nil, nil
				}
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "job-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockJobsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("resource in use")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockJobsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockJobsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			args := []string{"schedule", "job", "delete", "job-001", "--project-id", "proj-123", "--yes"}
			if tc.name == "--dry-run: prints intent, does not call Delete" {
				args = append(args, "--dry-run")
			}
			out, err := runCmdCapture(newMockClient(withSchedule(&mockScheduleClient{jobsClient: m})), args)
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}
