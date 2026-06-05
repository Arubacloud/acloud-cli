package cmd

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestElasticIPListCmd(t *testing.T) {
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
			args: []string{"network", "elasticip", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "eip-001", "my-eip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{
					Values: []types.ElasticIPResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "eip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "elasticip", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "elasticip", "list", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "eip-001", "my-eip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{
					Values: []types.ElasticIPResponse{
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
			args: []string{"network", "elasticip", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"network", "elasticip", "list", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", errorResponse(404, "Not Found", "resource not found"))
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

func TestElasticIPGetCmd(t *testing.T) {
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
				id, name := "eip-001", "my-eip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "eip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"network", "elasticip", "get", "eip-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestElasticIPCreateCmd(t *testing.T) {
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
			args: []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "ITBG-Bergamo", "--billing-period", "Hour"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "eip-001", "my-eip"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "eip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "elasticip", "create", "--project-id", "proj-123", "--region", "ITBG-Bergamo", "--billing-period", "Hour"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--billing-period", "Hour"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "ITBG-Bergamo", "--billing-period", "Hour"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/elasticIps", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "ITBG-Bergamo", "--billing-period", "Hour"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/elasticIps", errorResponse(404, "Not Found", "resource not found"))
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

func TestElasticIPUpdateCmd(t *testing.T) {
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
			args: []string{"network", "elasticip", "update", "eip-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "eip-001", "my-eip"
				state := types.StateActive
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "eip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"network", "elasticip", "update", "eip-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-Get server error",
			args: []string{"network", "elasticip", "update", "eip-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update server error",
			args: []string{"network", "elasticip", "update", "eip-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "eip-001", "my-eip"
				state := types.StateActive
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
					Status:   types.ResourceStatusResponse{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestElasticIPDeleteCmd(t *testing.T) {
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
			args: []string{"network", "elasticip", "delete", "eip-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "eip-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"network", "elasticip", "delete", "eip-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "eip-001", "my-eip"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "eip-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "elasticip", "delete", "eip-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"network", "elasticip", "delete", "eip-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestElasticIPCreateCmd_WithAddressAndTags(t *testing.T) {
	// Covers: Properties.Address and Metadata.Tags nil-guards in create response (lines 137-141).
	srv := newArubaTestServer(t)
	id, name := "eip-001", "my-eip"
	addr := "1.2.3.4"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:   &id,
			Name: &name,
			Tags: []string{"env=prod"},
		},
		Properties: types.ElasticIPPropertiesResponse{
			Address: &addr,
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "elasticip", "create", "--project-id", "proj-123", "--name", "my-eip", "--region", "ITBG-Bergamo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "1.2.3.4") {
		t.Errorf("expected address in output, got: %s", out)
	}
	if !strings.Contains(out, "env=prod") {
		t.Errorf("expected tags in output, got: %s", out)
	}
}

func TestElasticIPListCmd_AllOptionalFields(t *testing.T) {
	// Covers: LocationResponse, Address, Status.State nil-guards in list loop.
	srv := newArubaTestServer(t)
	id, name := "eip-001", "my-eip"
	addr := "1.2.3.4"
	state := types.StateActive
	region := types.Region("IT-BG")
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{
		Values: []types.ElasticIPResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.ElasticIPPropertiesResponse{
					Address: &addr,
				},
				Status: types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "elasticip", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "eip-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
	if !strings.Contains(out, "IT-BG") {
		t.Errorf("expected region in output, got: %s", out)
	}
	if !strings.Contains(out, "1.2.3.4") {
		t.Errorf("expected address in output, got: %s", out)
	}
	if !strings.Contains(out, "Active") {
		t.Errorf("expected status in output, got: %s", out)
	}
}

func TestElasticIPGetCmd_AllOptionalFields(t *testing.T) {
	// Covers: URI, LocationResponse, Address, BillingPlan.BillingPeriod detail block.
	id, name := "eip-001", "my-eip"
	uri := "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"
	addr := "1.2.3.4"
	state := types.StateActive
	region := types.Region("IT-BG")
	billingPeriod := types.BillingPeriodHour
	now := time.Now()
	createdBy := "user@example.com"
	makeResponse := func() types.ElasticIPResponse {
		return types.ElasticIPResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:               &id,
				Name:             &name,
				URI:              &uri,
				LocationResponse: &types.LocationResponse{Value: region},
				CreationDate:     &now,
				CreatedBy:        &createdBy,
				Tags:             []string{"env=test"},
			},
			Properties: types.ElasticIPPropertiesResponse{
				Address:           &addr,
				BillingPlanCommon: &types.BillingPlanCommon{BillingPeriod: &billingPeriod},
			},
			Status: types.ResourceStatusResponse{State: &state},
		}
	}

	t.Run("detail output with all optional fields", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, makeResponse()))
		out, err := runCmdCapture(srv.Client(), []string{"network", "elasticip", "get", "eip-001", "--project-id", "proj-123"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "eip-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
		if !strings.Contains(out, "1.2.3.4") {
			t.Errorf("expected address in output, got: %s", out)
		}
		if !strings.Contains(out, "IT-BG") {
			t.Errorf("expected region in output, got: %s", out)
		}
		if !strings.Contains(out, "Hour") {
			t.Errorf("expected billing period in output, got: %s", out)
		}
	})
}

func TestElasticIPListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "eip-001", "my-eip"
	region := types.Region("IT-BG")
	state := types.StateActive
	addr := "203.0.113.1"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{
		Values: []types.ElasticIPResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Properties: types.ElasticIPPropertiesResponse{Address: &addr},
				Status:     types.ResourceStatusResponse{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "elasticip", "list", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "eip-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestElasticIPGetCmd_FullDetail(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "eip-001", "my-eip"
	uri := "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"
	region := types.Region("IT-BG")
	state := types.StateActive
	addr := "203.0.113.1"
	createdBy := "test-user@example.com"
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=test"},
		},
		Properties: types.ElasticIPPropertiesResponse{Address: &addr},
		Status:     types.ResourceStatusResponse{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "elasticip", "get", "eip-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "eip-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

// =============================================================================
// Layer 1 — Validate() tests (pure-Go, no SDK, no httptest)
// =============================================================================

func validNetworkElasticIPCreateArgs() NetworkElasticIPCreateArgs {
	return NetworkElasticIPCreateArgs{
		ProjectID:     "proj-123",
		Name:          "my-eip",
		Region:        aruba.RegionITBGBergamo,
		BillingPeriod: aruba.BillingPeriodHour,
	}
}

func TestNetworkElasticIPCreateArgs_Validate(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*NetworkElasticIPCreateArgs)
		wantErr     bool
		errContains string
	}{
		{
			name:    "happy path",
			wantErr: false,
		},
		{
			name:        "name too short",
			mutate:      func(a *NetworkElasticIPCreateArgs) { a.Name = "ab" },
			wantErr:     true,
			errContains: "--name must be at least 3 characters",
		},
		{
			name:    "name minimum length 3",
			mutate:  func(a *NetworkElasticIPCreateArgs) { a.Name = "abc" },
			wantErr: false,
		},
		{
			name:        "name too long",
			mutate:      func(a *NetworkElasticIPCreateArgs) { a.Name = strings.Repeat("x", 65) },
			wantErr:     true,
			errContains: "--name must be at most 64 characters",
		},
		{
			name:        "invalid region",
			mutate:      func(a *NetworkElasticIPCreateArgs) { a.Region = "ZZ-Invalid" },
			wantErr:     true,
			errContains: "--region",
		},
		{
			name:        "invalid billing period",
			mutate:      func(a *NetworkElasticIPCreateArgs) { a.BillingPeriod = "Weekly" },
			wantErr:     true,
			errContains: "--billing-period",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := validNetworkElasticIPCreateArgs()
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
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNetworkElasticIPGetArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkElasticIPGetArgs{ProjectID: "p1", ID: "eip-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := NetworkElasticIPGetArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

func TestNetworkElasticIPUpdateArgs_Validate(t *testing.T) {
	t.Run("happy path with name", func(t *testing.T) {
		args := NetworkElasticIPUpdateArgs{ProjectID: "p1", ID: "eip-001", Name: "new-name"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("happy path with tags changed", func(t *testing.T) {
		args := NetworkElasticIPUpdateArgs{ProjectID: "p1", ID: "eip-001", TagsChanged: true}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("no name and no tags", func(t *testing.T) {
		args := NetworkElasticIPUpdateArgs{ProjectID: "p1", ID: "eip-001"}
		err := args.Validate()
		if err == nil {
			t.Fatal("expected error when no flags provided")
		}
		if !strings.Contains(err.Error(), "at least one") {
			t.Errorf("error %q does not contain 'at least one'", err.Error())
		}
	})
}

func TestNetworkElasticIPDeleteArgs_Validate(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		args := NetworkElasticIPDeleteArgs{ProjectID: "p1", ID: "eip-001"}
		if err := args.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("empty ID", func(t *testing.T) {
		args := NetworkElasticIPDeleteArgs{ProjectID: "p1", ID: ""}
		if err := args.Validate(); err == nil {
			t.Fatal("expected error for empty ID")
		}
	})
}

// =============================================================================
// Layer 2 — Operation function tests (httptest harness, bypasses RunE)
// =============================================================================

func TestNetworkElasticIPCreate_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "eip-new", "my-eip"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))

	out := captureStdout(func() {
		err := NetworkElasticIPCreate(context.Background(), srv.Client(), validNetworkElasticIPCreateArgs())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "eip-new") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkElasticIPCreate_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/elasticIps", errorResponse(500, "Internal Server Error", "quota exceeded"))

	err := NetworkElasticIPCreate(context.Background(), srv.Client(), validNetworkElasticIPCreateArgs())
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "creating Elastic IP") {
		t.Errorf("error %q does not contain 'creating Elastic IP'", err.Error())
	}
}

func TestNetworkElasticIPList_HappyPath(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "eip-001", "my-eip"
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{
		Values: []types.ElasticIPResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))

	out := captureStdout(func() {
		err := NetworkElasticIPList(context.Background(), srv.Client(), NetworkElasticIPListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "eip-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestNetworkElasticIPList_Empty(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{}))

	out := captureStdout(func() {
		err := NetworkElasticIPList(context.Background(), srv.Client(), NetworkElasticIPListArgs{
			ProjectID: "proj-123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No Elastic IPs found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}
