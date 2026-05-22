package cmd

// helpers_coverage_test.go covers helper functions and nil branches that are
// missed by the command-level tests: *FromRaw(nil), redactVPNTunnelSecrets,
// vpnTunnelReattachSettings (via update command), apiErrFromV2 (extra paths),
// and confirmDelete non-interactive mode.

import (
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ─── *FromRaw nil branches ────────────────────────────────────────────────────
// Each *FromRaw function has two branches: nil input → nil, non-nil → Raw().
// The non-nil branch is exercised by the command tests; the nil branch is not.

func TestFromRaw_Nil(t *testing.T) {
	t.Run("containerRegistryFromRaw", func(t *testing.T) {
		if containerRegistryFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("kaasFromRaw", func(t *testing.T) {
		if kaasFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("databaseBackupFromRaw", func(t *testing.T) {
		if databaseBackupFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("databaseFromRaw", func(t *testing.T) {
		if databaseFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("dbaasFromRaw", func(t *testing.T) {
		if dbaasFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("userFromRaw", func(t *testing.T) {
		if userFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("projectFromRaw", func(t *testing.T) {
		if projectFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("backupFromRaw (storage)", func(t *testing.T) {
		if backupFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("restoreFromRaw", func(t *testing.T) {
		if restoreFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("snapshotFromRaw", func(t *testing.T) {
		if snapshotFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("volumeFromRaw", func(t *testing.T) {
		if volumeFromRaw(nil) != nil {
			t.Error("expected nil")
		}
	})
}

// ─── redactVPNTunnelSecrets ───────────────────────────────────────────────────

func TestRedactVPNTunnelSecrets(t *testing.T) {
	secret := "super-secret"
	cloud := "cloud-site"

	tunnels := []types.VPNTunnelResponse{
		{
			Properties: types.VPNTunnelPropertiesResponse{
				VPNClientSettings: &types.VPNClientSettings{
					PSK: &types.PSKSettings{
						CloudSite: &cloud,
						Secret:    &secret,
					},
				},
			},
		},
		// nil VPNClientSettings — must not panic
		{},
		// non-nil settings but nil PSK — must not panic
		{
			Properties: types.VPNTunnelPropertiesResponse{
				VPNClientSettings: &types.VPNClientSettings{},
			},
		},
	}

	redactVPNTunnelSecrets(tunnels)

	if tunnels[0].Properties.VPNClientSettings.PSK.Secret != nil {
		t.Error("expected PSK.Secret to be nil after redaction")
	}
	if tunnels[0].Properties.VPNClientSettings.PSK.CloudSite == nil {
		t.Error("expected non-secret PSK fields to be preserved")
	}
}

// ─── vpnTunnelReattachSettings (via update command) ───────────────────────────
// Tests that the update command survives a GET response that has full
// IKE/ESP/PSK/IPConfig settings — exercising the reattach helper branches.

func TestVPNTunnelUpdate_ReattachSettings(t *testing.T) {
	enc := types.IKEEncryption("aes256")
	hash := types.IKEHash("sha1")
	dhGroup := types.IKEDHGroup("1")
	dpdAction := types.IKEDPDAction("restart")
	espEnc := types.ESPEncryption("aes256")
	espHash := types.ESPHash("sha1")
	pfs := types.ESPPFSGroup("enable")
	cloudSite := "psk-cloud"
	secret := "psk-secret"

	id, name := "vpn-001", "my-tunnel"
	active := "Active"
	getResp := types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatus{State: &active},
		Properties: types.VPNTunnelPropertiesResponse{
			IPConfigurations: &types.IPConfigurations{
				VPC:      &types.ReferenceResource{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"},
				PublicIP: &types.ReferenceResource{URI: "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"},
				Subnet:   &types.SubnetInfo{Name: "my-subnet", CIDR: "10.0.0.0/24"},
			},
			VPNClientSettings: &types.VPNClientSettings{
				IKE: &types.IKESettings{
					Lifetime:    3600,
					Encryption:  &enc,
					Hash:        &hash,
					DHGroup:     &dhGroup,
					DPDAction:   &dpdAction,
					DPDInterval: 10,
					DPDTimeout:  30,
				},
				ESP: &types.ESPSettings{
					Lifetime:   1800,
					Encryption: &espEnc,
					Hash:       &espHash,
					PFS:        &pfs,
				},
				PSK: &types.PSKSettings{
					CloudSite: &cloudSite,
					Secret:    &secret,
				},
			},
		},
	}

	updatedName := "updated-tunnel"
	putResp := types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &updatedName},
	}

	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001", jsonResponse(200, getResp))
	srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001", jsonResponse(200, putResp))

	out, err := runCmdCapture(srv.Client(), []string{
		"network", "vpntunnel", "update", "vpn-001",
		"--project-id", "proj-123",
		"--name", "updated-tunnel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "updated-tunnel") {
		t.Errorf("expected updated name in output, got: %s", out)
	}
}

func TestVPNTunnelUpdate_NilIPConfig(t *testing.T) {
	// vpnTunnelReattachSettings with nil IPConfigurations and nil VPNClientSettings
	id, name, active := "vpn-002", "tunnel-2", "Active"
	getResp := types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatus{State: &active},
	}
	updatedName := "tunnel-2-updated"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-002", jsonResponse(200, getResp))
	srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-002", jsonResponse(200, types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &updatedName},
	}))

	_, err := runCmdCapture(srv.Client(), []string{
		"network", "vpntunnel", "update", "vpn-002",
		"--project-id", "proj-123",
		"--name", "tunnel-2-updated",
	})
	if err != nil {
		t.Fatalf("unexpected error with nil IPConfig: %v", err)
	}
}

// ─── apiErrFromV2 extra branches ──────────────────────────────────────────────

func TestAPIErrFromV2_Extra(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		if apiErrFromV2(nil) != nil {
			t.Error("expected nil for nil input")
		}
	})

	t.Run("non-HTTP error is returned as-is", func(t *testing.T) {
		orig := errors.New("some non-http error")
		got := apiErrFromV2(orig)
		if got != orig {
			t.Errorf("expected same error, got %v", got)
		}
	})

	t.Run("HTTP error with validation errors", func(t *testing.T) {
		title := "Validation Error"
		detail := "bad input"
		httpErr := &aruba.HTTPError{
			StatusCode: 422,
			ErrResp: &types.ErrorResponse{
				Title:  &title,
				Detail: &detail,
				Errors: []types.ValidationError{
					{Field: "name", Message: "too long"},
					{Message: "general error"},
					{Field: "region"},
				},
			},
		}
		got := apiErrFromV2(httpErr)
		if got == nil {
			t.Fatal("expected non-nil error")
		}
		msg := got.Error()
		if !strings.Contains(msg, "422") {
			t.Errorf("expected status code in message, got: %s", msg)
		}
		if !strings.Contains(msg, "validation") {
			t.Errorf("expected validation details in message, got: %s", msg)
		}
		if !strings.Contains(msg, "name: too long") {
			t.Errorf("expected field:message in output, got: %s", msg)
		}
	})

	t.Run("HTTP error with empty validation slice", func(t *testing.T) {
		title := "Bad Request"
		httpErr := &aruba.HTTPError{
			StatusCode: 400,
			ErrResp:    &types.ErrorResponse{Title: &title, Errors: []types.ValidationError{}},
		}
		got := apiErrFromV2(httpErr)
		if got == nil {
			t.Fatal("expected non-nil error")
		}
		if strings.Contains(got.Error(), "validation") {
			t.Errorf("no validation details expected for empty errors slice, got: %s", got.Error())
		}
	})

	t.Run("HTTP error nil ErrResp", func(t *testing.T) {
		httpErr := &aruba.HTTPError{StatusCode: 503}
		got := apiErrFromV2(httpErr)
		if got == nil {
			t.Fatal("expected non-nil error")
		}
		if !strings.Contains(got.Error(), "503") {
			t.Errorf("expected status in message, got: %s", got.Error())
		}
	})
}

// ─── confirmDelete non-interactive mode ───────────────────────────────────────
// When stdin is a pipe (not a TTY), Stat() reports ModeCharDevice unset →
// confirmDelete returns an error requiring --yes.

func TestConfirmDelete_NonInteractive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("stdin ModeCharDevice detection is unreliable on Windows in test processes")
	}
	// Replace os.Stdin with a real pipe so ModeCharDevice is unset regardless
	// of how the test runner sets up its stdin (e.g. GitHub Actions).
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	_, err = confirmDelete("vpc", "vpc-001")
	if err == nil {
		t.Fatal("expected error in non-interactive mode")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("expected --yes mention in error, got: %s", err.Error())
	}
}

// ─── *ListPayload nil branches ────────────────────────────────────────────────
// Each *ListPayload function has a defensive return nil when the type assertion
// fails. A zero-value *aruba.List has raw=nil, so the type assertion always
// fails → return nil is exercised.

func TestListPayload_Nil(t *testing.T) {
	t.Run("csListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.CloudServer]
		if csListPayload(&l) != nil {
			t.Error("expected nil from csListPayload with nil raw")
		}
	})
	t.Run("keypairListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.KeyPair]
		if keypairListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("containerRegistryListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.ContainerRegistry]
		if containerRegistryListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("kaasListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.KaaS]
		if kaasListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("databaseBackupListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.DBaaSBackup]
		if databaseBackupListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("databaseListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.Database]
		if databaseListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("dbaasListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.DBaaS]
		if dbaasListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("userListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.User]
		if userListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("projectListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.Project]
		if projectListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("elasticIPListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.ElasticIP]
		if elasticIPListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("loadBalancerListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.LoadBalancer]
		if loadBalancerListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("securityGroupListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.SecurityGroup]
		if securityGroupListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("securityRuleListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.SecurityRule]
		if securityRuleListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("subnetListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.Subnet]
		if subnetListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("vpcListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.VPC]
		if vpcListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("vpcPeeringListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.VPCPeering]
		if vpcPeeringListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("vpcPeeringRouteListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.VPCPeeringRoute]
		if vpcPeeringRouteListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("vpnRouteListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.VPNRoute]
		if vpnRouteListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("vpnTunnelListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.VPNTunnel]
		if vpnTunnelListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("jobListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.Job]
		if jobListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("kmsListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.KMS]
		if kmsListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("backupListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.StorageBackup]
		if backupListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("volumeListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.BlockStorage]
		if volumeListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("restoreListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.StorageRestore]
		if restoreListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
	t.Run("snapshotListPayload", func(t *testing.T) {
		var l aruba.List[*aruba.Snapshot]
		if snapshotListPayload(&l) != nil {
			t.Error("expected nil")
		}
	})
}
