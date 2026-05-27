package cmd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestSubnetListCmd(t *testing.T) {
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
			args: []string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sub-001", "my-subnet"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetList{
					Values: []types.SubnetResponse{
						{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
					},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sub-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "success empty",
			args: []string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetList{}))
			},
		},
		{
			name: "--output json emits valid JSON",
			args: []string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123", "--output", "json"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sub-001", "my-subnet"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetList{
					Values: []types.SubnetResponse{
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
			args: []string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			args: []string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", errorResponse(404, "Not Found", "resource not found"))
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

func TestSubnetGetCmd(t *testing.T) {
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
				id, name := "sub-001", "my-subnet"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sub-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", errorResponse(500, "Internal Server Error", "boom"))
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", errorResponse(404, "Not Found", "resource not found"))
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
			out, err := runCmdCapture(srv.Client(), []string{"network", "subnet", "get", "vpc-001", "sub-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
			if tc.assertOut != nil {
				tc.assertOut(t, out)
			}
		})
	}
}

func TestSubnetCreateCmd(t *testing.T) {
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
			args: []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet", "--region", "IT-BG"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sub-001", "my-subnet"
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sub-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--region", "IT-BG"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --region",
			args:        []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet"},
			wantErr:     true,
			errContains: "region",
		},
		{
			name: "server error propagates",
			args: []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet", "--region", "IT-BG"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", errorResponse(500, "Internal Server Error", "quota exceeded"))
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"network", "subnet", "create", "vpc-001", "--project-id", "proj-123", "--name", "my-subnet", "--region", "IT-BG"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", errorResponse(404, "Not Found", "resource not found"))
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

func TestSubnetUpdateCmd(t *testing.T) {
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
			args: []string{"network", "subnet", "update", "vpc-001", "sub-001", "--project-id", "proj-123", "--name", "new-name"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sub-001", "my-subnet"
				state := types.StateActive
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, LocationResponse: &types.LocationResponse{Value: region}},
					Status:   types.ResourceStatus{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sub-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name:        "no flags error",
			args:        []string{"network", "subnet", "update", "vpc-001", "sub-001", "--project-id", "proj-123"},
			wantErr:     true,
			errContains: "at least one",
		},
		{
			name: "pre-Get server error",
			args: []string{"network", "subnet", "update", "vpc-001", "sub-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", errorResponse(404, "Not Found", "resource not found"))
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
		{
			name: "update server error",
			args: []string{"network", "subnet", "update", "vpc-001", "sub-001", "--project-id", "proj-123", "--name", "x"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sub-001", "my-subnet"
				state := types.StateActive
				region := types.Region("IT-BG")
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name, LocationResponse: &types.LocationResponse{Value: region}},
					Status:   types.ResourceStatus{State: &state},
				}))
				srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", errorResponse(500, "Internal Server Error", "boom"))
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

func TestSubnetDeleteCmd(t *testing.T) {
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
			args: []string{"network", "subnet", "delete", "vpc-001", "sub-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, nil))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sub-001") {
					t.Errorf("expected ID in output, got: %s", out)
				}
			},
		},
		{
			name: "--dry-run: prints intent, does not call Delete",
			args: []string{"network", "subnet", "delete", "vpc-001", "sub-001", "--project-id", "proj-123", "--yes", "--dry-run"},
			setupSrv: func(srv *arubaTestServer) {
				id, name := "sub-001", "my-subnet"
				srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
					Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
				}))
			},
			assertOut: func(t *testing.T, out string) {
				if !strings.Contains(out, "sub-001") {
					t.Errorf("expected ID in dry-run output, got: %s", out)
				}
			},
		},
		{
			name: "server error propagates",
			args: []string{"network", "subnet", "delete", "vpc-001", "sub-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", errorResponse(500, "Internal Server Error", "resource in use"))
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			args: []string{"network", "subnet", "delete", "vpc-001", "sub-001", "--project-id", "proj-123", "--yes"},
			setupSrv: func(srv *arubaTestServer) {
				srv.OnDelete("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", errorResponse(404, "Not Found", "resource not found"))
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

func TestSplitRouteString(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"0.0.0.0/0:10.0.0.1", []string{"0.0.0.0/0", "10.0.0.1"}},
		{"192.168.1.0/24:192.168.1.1", []string{"192.168.1.0/24", "192.168.1.1"}},
		{"no-colon", []string{"no-colon"}},
		{"", []string{""}},
	}
	for _, c := range cases {
		got := splitRouteString(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitRouteString(%q): got %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("splitRouteString(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}

func TestSubnetGetCmd_WithCreationDateAndTags(t *testing.T) {
	// Covers: CreationDate, CreatedBy, Tags nil-guards in detail block (lines 228-235).
	srv := newArubaTestServer(t)
	id, name := "sub-001", "my-subnet"
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001"
	createdBy := "user@example.com"
	state := types.StateActive
	region := types.Region("IT-BG")
	now := time.Now()
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
			CreationDate:     &now,
			CreatedBy:        &createdBy,
			Tags:             []string{"env=prod"},
		},
		Properties: types.SubnetPropertiesResponse{
			Type:    types.SubnetTypeAdvanced,
			Network: &types.SubnetNetwork{Address: "10.1.0.0/24"},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "subnet", "get", "vpc-001", "sub-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "user@example.com") {
		t.Errorf("expected createdBy in output, got: %s", out)
	}
	if !strings.Contains(out, "env=prod") {
		t.Errorf("expected tags in output, got: %s", out)
	}
}

func TestSubnetUpdateCmd_WithTagsAndCIDR(t *testing.T) {
	// Covers: tags branch (subnet.RetaggedAs) line 353, cidr branch (subnet.WithCIDR) line 356.
	id, name := "sub-001", "my-subnet"
	state := types.StateActive
	region := types.Region("IT-BG")

	t.Run("update with tags", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:               &id,
				Name:             &name,
				LocationResponse: &types.LocationResponse{Value: region},
			},
			Status: types.ResourceStatus{State: &state},
		}))
		updName := "sub-001"
		srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
			Metadata: types.ResourceMetadataResponse{ID: &updName, Name: &name, Tags: []string{"foo=bar"}},
		}))
		out, err := runCmdCapture(srv.Client(), []string{
			"network", "subnet", "update", "vpc-001", "sub-001",
			"--project-id", "proj-123",
			"--tags", "foo=bar",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "sub-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})

	t.Run("update with cidr", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
			Metadata: types.ResourceMetadataResponse{
				ID:               &id,
				Name:             &name,
				LocationResponse: &types.LocationResponse{Value: region},
			},
			Status: types.ResourceStatus{State: &state},
		}))
		updName := "sub-001"
		srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
			Metadata:   types.ResourceMetadataResponse{ID: &updName, Name: &name},
			Properties: types.SubnetPropertiesResponse{Network: &types.SubnetNetwork{Address: "10.1.0.0/24"}},
		}))
		out, err := runCmdCapture(srv.Client(), []string{
			"network", "subnet", "update", "vpc-001", "sub-001",
			"--project-id", "proj-123",
			"--cidr", "10.1.0.0/24",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out, "sub-001") {
			t.Errorf("expected ID in output, got: %s", out)
		}
	})
}

func TestSubnetCreateCmd_AdvancedType(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sub-001", "my-subnet"
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
	}))
	err := runCmd(srv.Client(), []string{
		"network", "subnet", "create", "vpc-001",
		"--project-id", "proj-123",
		"--name", "my-subnet",
		"--region", "IT-BG",
		"--cidr", "10.0.1.0/24",
		"--dhcp-enabled",
		"--dhcp-dns", "8.8.8.8,8.8.4.4",
		"--dhcp-routes", "10.1.0.0/24:10.0.0.1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubnetListCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sub-001", "my-subnet"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetList{
		Values: []types.SubnetResponse{
			{
				Metadata: types.ResourceMetadataResponse{
					ID:               &id,
					Name:             &name,
					LocationResponse: &types.LocationResponse{Value: region},
				},
				Status: types.ResourceStatus{State: &state},
			},
		},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "subnet", "list", "vpc-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sub-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSubnetGetCmd_WithLocationAndStatus(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sub-001", "my-subnet"
	region := types.Region("IT-BG")
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "subnet", "get", "vpc-001", "sub-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sub-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSubnetGetCmd_WithDHCPAndNetwork(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sub-001", "my-subnet"
	region := types.Region("IT-BG")
	uri := "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{
			ID:               &id,
			Name:             &name,
			URI:              &uri,
			LocationResponse: &types.LocationResponse{Value: region},
		},
		Properties: types.SubnetPropertiesResponse{
			Type:    types.SubnetTypeAdvanced,
			Network: &types.SubnetNetwork{Address: "10.0.1.0/24"},
			DHCP: &types.SubnetDHCP{
				Enabled: true,
				Routes:  []types.SubnetDHCPRoute{{Address: "10.0.0.0/8", Gateway: "10.0.1.1"}},
				DNS:     []string{"8.8.8.8", "8.8.4.4"},
			},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{"network", "subnet", "get", "vpc-001", "sub-001", "--project-id", "proj-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sub-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}

func TestSubnetUpdateCmd_WithDHCP(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "sub-001", "my-subnet"
	state := types.StateActive
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Properties: types.SubnetPropertiesResponse{
			Type:    types.SubnetTypeAdvanced,
			Network: &types.SubnetNetwork{Address: "10.0.1.0/24"},
			DHCP: &types.SubnetDHCP{
				Enabled: true,
				Routes:  []types.SubnetDHCPRoute{{Address: "10.0.0.0/8", Gateway: "10.0.1.1"}},
				DNS:     []string{"8.8.8.8"},
			},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	updID, updName := "sub-001", "my-subnet"
	srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets/sub-001", jsonResponse(200, types.SubnetResponse{
		Metadata: types.ResourceMetadataResponse{ID: &updID, Name: &updName},
		Properties: types.SubnetPropertiesResponse{
			Network: &types.SubnetNetwork{Address: "10.0.1.0/24"},
		},
		Status: types.ResourceStatus{State: &state},
	}))
	out, err := runCmdCapture(srv.Client(), []string{
		"network", "subnet", "update", "vpc-001", "sub-001",
		"--project-id", "proj-123",
		"--dhcp-enabled",
		"--dhcp-routes", "192.168.0.0/16:10.0.1.254",
		"--dhcp-dns", "1.1.1.1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "sub-001") {
		t.Errorf("expected ID in output, got: %s", out)
	}
}
