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

// ─── redactVPNTunnelSecrets ───────────────────────────────────────────────────

func TestRedactVPNTunnelSecrets(t *testing.T) {
	secret := "super-secret"
	cloud := "cloud-site"
	id := "vpn-redact-1"

	// Build a VPNTunnel via a mock GET so we get a real *aruba.VPNTunnel.
	resp := types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id},
		Properties: types.VPNTunnelPropertiesResponse{
			VPNClientSettingsCommon: &types.VPNClientSettingsCommon{
				PSK: &types.PSKSettingsCommon{
					CloudSite: &cloud,
					Secret:    &secret,
				},
			},
		},
	}
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-redact-1", jsonResponse(200, resp))

	ctx, cancel := newCtx()
	defer cancel()
	tunnel, err := srv.Client().FromNetwork().VPNTunnels().Get(ctx, aruba.URI("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-redact-1"))
	if err != nil {
		t.Fatalf("unexpected error fetching tunnel: %v", err)
	}

	redactVPNTunnelSecrets(tunnel)

	if tunnel.Raw().Properties.VPNClientSettingsCommon.PSK.Secret != nil {
		t.Error("expected PSK.Secret to be nil after redaction")
	}
	if tunnel.Raw().Properties.VPNClientSettingsCommon.PSK.CloudSite == nil {
		t.Error("expected non-secret PSK fields to be preserved")
	}

	// nil tunnel — must not panic
	redactVPNTunnelSecrets(nil)
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
	active := types.StateActive
	getResp := types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &active},
		Properties: types.VPNTunnelPropertiesResponse{
			IPConfigurationsCommon: &types.IPConfigurationsCommon{
				VPC:      &types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001"},
				PublicIP: &types.ReferenceResourceCommon{URI: "/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001"},
				Subnet:   &types.SubnetInfoCommon{Name: "my-subnet", CIDR: "10.0.0.0/24"},
			},
			VPNClientSettingsCommon: &types.VPNClientSettingsCommon{
				IKE: &types.IKESettingsCommon{
					Lifetime:    3600,
					Encryption:  &enc,
					Hash:        &hash,
					DHGroup:     &dhGroup,
					DPDAction:   &dpdAction,
					DPDInterval: 10,
					DPDTimeout:  30,
				},
				ESP: &types.ESPSettingsCommon{
					Lifetime:   1800,
					Encryption: &espEnc,
					Hash:       &espHash,
					PFS:        &pfs,
				},
				PSK: &types.PSKSettingsCommon{
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
	id, name := "vpn-002", "tunnel-2"
	active := types.StateActive
	getResp := types.VPNTunnelResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &active},
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

// ─── *Ref helpers ─────────────────────────────────────────────────────────────

func TestRefHelpers(t *testing.T) {
	uriStr := func(r aruba.Ref) string { return r.URI() }
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"project", uriStr(projectRef("p1")), "/projects/p1"},
		{"cloudserver", uriStr(cloudServerRef("p1", "cs1")),
			"/projects/p1/providers/Aruba.Compute/cloudServers/cs1"},
		{"keypair", uriStr(keypairRef("p1", "kp1")),
			"/projects/p1/providers/Aruba.Compute/keyPairs/kp1"},
		{"volume", uriStr(volumeRef("p1", "v1")),
			"/projects/p1/providers/Aruba.Storage/blockStorages/v1"},
		{"loadbalancer", uriStr(loadBalancerRef("p1", "lb1")),
			"/projects/p1/providers/Aruba.Network/loadBalancers/lb1"},
		{"securitygroup", uriStr(securityGroupRef("p1", "vpc1", "sg1")),
			"/projects/p1/providers/Aruba.Network/vpcs/vpc1/security-groups/sg1"},
		{"database", uriStr(databaseRef("p1", "d1", "mydb")),
			"/projects/p1/providers/Aruba.Database/dbaas/d1/databases/mydb"},
		{"grant", uriStr(grantRef("p1", "d1", "mydb", "g1")),
			"/projects/p1/providers/Aruba.Database/dbaas/d1/databases/mydb/grants/g1"},
		{"kms", uriStr(kmsRef("p1", "k1")),
			"/projects/p1/providers/Aruba.Security/kms/k1"},
		{"job", uriStr(jobRef("p1", "j1")),
			"/projects/p1/providers/Aruba.Schedule/jobs/j1"},
		{"restore", uriStr(restoreRef("p1", "b1", "r1")),
			"/projects/p1/providers/Aruba.Storage/backups/b1/restores/r1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.got != c.want {
				t.Fatalf("got %q, want %q", c.got, c.want)
			}
		})
	}
}

// ─── extractIDFromURI ─────────────────────────────────────────────────────────

func TestExtractIDFromURI(t *testing.T) {
	cases := []struct{ in, want string }{
		{"/projects/p1/providers/Aruba.Network/elasticIps/eip-42", "eip-42"},
		{"/projects/p1/", "p1"},  // trailing slash trimmed
		{"plain-id", "plain-id"}, // no slash
		{"", ""},
	}
	for _, c := range cases {
		if got := extractIDFromURI(c.in); got != c.want {
			t.Errorf("extractIDFromURI(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ─── projectWrapper ───────────────────────────────────────────────────────────

func TestProjectWrapper(t *testing.T) {
	id := "p1"
	name := "alpha"
	t.Run("success", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/p1", jsonResponse(200, types.ProjectResponse{
			Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		}))
		ctx, cancel := newCtx()
		defer cancel()
		p, err := projectWrapper(ctx, srv.Client(), "p1")
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if p.ID() != "p1" {
			t.Fatalf("ID=%q want p1", p.ID())
		}
	})

	t.Run("api error normalised", func(t *testing.T) {
		srv := newArubaTestServer(t)
		srv.OnGet("/projects/p404", errorResponse(404, "Not Found", "no such project"))
		ctx, cancel := newCtx()
		defer cancel()
		_, err := projectWrapper(ctx, srv.Client(), "p404")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("err does not mention status 404: %v", err)
		}
	})
}

// ─── printJSON / printYAML ────────────────────────────────────────────────────

type fakeRawMarshaler struct{ j, y []byte }

func (f fakeRawMarshaler) RawJSON() []byte { return f.j }
func (f fakeRawMarshaler) RawYAML() []byte { return f.y }

func TestPrintJSON(t *testing.T) {
	t.Run("nil emits {}", func(t *testing.T) {
		out := captureStdout(func() { printJSON(nil) })
		if strings.TrimSpace(out) != "{}" {
			t.Fatalf("got %q, want {}", out)
		}
	})

	t.Run("rawMarshaler written verbatim", func(t *testing.T) {
		payload := []byte(`{"id":"x1"}`)
		out := captureStdout(func() { printJSON(fakeRawMarshaler{j: payload}) })
		if !strings.Contains(out, `"id":"x1"`) {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("rawMarshaler empty bytes emits {}", func(t *testing.T) {
		out := captureStdout(func() { printJSON(fakeRawMarshaler{j: []byte{}}) })
		if strings.TrimSpace(out) != "{}" {
			t.Fatalf("got %q, want {}", out)
		}
	})

	t.Run("plain struct marshalled", func(t *testing.T) {
		out := captureStdout(func() { printJSON(map[string]any{"hello": "world"}) })
		if !strings.Contains(out, `"hello"`) || !strings.Contains(out, `"world"`) {
			t.Fatalf("got %q", out)
		}
	})
}

func TestPrintYAML(t *testing.T) {
	t.Run("nil emits {}", func(t *testing.T) {
		out := captureStdout(func() { printYAML(nil) })
		if strings.TrimSpace(out) != "{}" {
			t.Fatalf("got %q, want {}", out)
		}
	})

	t.Run("rawMarshaler written verbatim", func(t *testing.T) {
		payload := []byte("id: x1\n")
		out := captureStdout(func() { printYAML(fakeRawMarshaler{y: payload}) })
		if !strings.Contains(out, "id: x1") {
			t.Fatalf("got %q", out)
		}
	})

	t.Run("rawMarshaler empty bytes emits {}", func(t *testing.T) {
		out := captureStdout(func() { printYAML(fakeRawMarshaler{y: []byte{}}) })
		if strings.TrimSpace(out) != "{}" {
			t.Fatalf("got %q, want {}", out)
		}
	})

	t.Run("plain struct converted to YAML", func(t *testing.T) {
		out := captureStdout(func() { printYAML(map[string]any{"hello": "world"}) })
		if !strings.Contains(out, "hello: world") {
			t.Fatalf("got %q", out)
		}
	})
}
