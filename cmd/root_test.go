package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func TestSetVersion(t *testing.T) {
	orig := rootCmd.Version
	t.Cleanup(func() { rootCmd.Version = orig })

	SetVersion("v1.2.3")
	if rootCmd.Version != "v1.2.3" {
		t.Fatalf("expected v1.2.3, got %q", rootCmd.Version)
	}

	SetVersion("")
	if rootCmd.Version != "dev" {
		t.Fatalf("expected dev for empty version, got %q", rootCmd.Version)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// setupMockConfig creates a temporary config file for testing
func setupMockConfig(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".acloud.yaml")

	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}

	// Create mock config
	config := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("Failed to create mock config: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}

	return configPath, cleanup
}

func TestGetArubaClient(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	// Clear cache before test
	resetClientState()

	client, err := GetArubaClient()
	if err != nil {
		t.Fatalf("GetArubaClient() error = %v", err)
	}

	if client == nil {
		t.Fatal("GetArubaClient() returned nil client")
	}
}

func TestGetArubaClient_Caching(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	// Clear cache
	resetClientState()

	client1, err1 := GetArubaClient()
	if err1 != nil {
		t.Fatalf("GetArubaClient() error = %v", err1)
	}

	// Second call should use cached client
	client2, err2 := GetArubaClient()
	if err2 != nil {
		t.Fatalf("GetArubaClient() second call error = %v", err2)
	}

	// Should be the same instance (cached)
	if client1 != client2 {
		t.Error("GetArubaClient() should return cached client on second call")
	}
}

func TestGetArubaClient_NoConfig(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	// Create temporary directory without config
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}()

	client, err := GetArubaClient()
	if err == nil {
		t.Error("GetArubaClient() should return error when config doesn't exist")
	}
	if client != nil {
		t.Error("GetArubaClient() should return nil client when config doesn't exist")
	}

	// Verify error message
	errMsg := err.Error()
	if !contains(errMsg, "failed to load configuration") {
		t.Errorf("Expected error about failed to load configuration, got: %v", err)
	}
}

func TestGetArubaClient_EmptyCredentials(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	// Create temporary directory
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}()

	// Create config with empty credentials
	config := &Config{
		ClientID:     "",
		ClientSecret: "",
	}
	err := SaveConfig(config)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	client, err := GetArubaClient()
	if err == nil {
		t.Error("GetArubaClient() should return error when credentials are empty")
	}
	if client != nil {
		t.Error("GetArubaClient() should return nil client when credentials are empty")
	}

	// Verify error message
	errMsg := err.Error()
	if !contains(errMsg, "client ID or client secret not configured") {
		t.Errorf("Expected error about credentials not configured, got: %v", err)
	}
}

func TestPrintTable(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 10},
		{Header: "ID", Width: 10},
		{Header: "STATUS", Width: 10},
	}

	rows := [][]string{
		{"test-name", "test-id-123", "Active"},
		{"another-name", "another-id-456", "Inactive"},
	}

	// Should not panic
	PrintTable(headers, rows)
}

func TestPrintTable_EmptyRows(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 10},
		{Header: "ID", Width: 10},
	}

	rows := [][]string{}

	// Should not panic with empty rows
	PrintTable(headers, rows)
}

func TestPrintTable_LongValues(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 5},
	}

	rows := [][]string{
		{"very-long-name-that-exceeds-width"},
	}

	// Should truncate long values
	PrintTable(headers, rows)
}

func TestPrintTable_MismatchedColumns(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 10},
		{Header: "ID", Width: 10},
	}

	rows := [][]string{
		{"test-name"}, // Missing second column
		{"another-name", "id-123", "extra-column"}, // Extra column
	}

	// Should not panic with mismatched columns
	PrintTable(headers, rows)
}

func TestNormalizeHeaderKey(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"NAME", "name"},
		{"ID", "id"},
		{"PUBLIC_KEY", "public_key"},
		{"CREATION DATE", "creation_date"},
		{"RAM(GB)", "ram_gb"},
		{"HD(GB)", "hd_gb"},
		{"CPU", "cpu"},
		{"  spaced  ", "spaced"},
		{"multi  space__key", "multi_space_key"},
		{"abc123", "abc123"},
	}
	for _, c := range cases {
		got := normalizeHeaderKey(c.in)
		if got != c.want {
			t.Errorf("normalizeHeaderKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// captureStdout captures os.Stdout during f() and returns the output.
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestResolveOutputFormat(t *testing.T) {
	cases := []struct {
		flag string
		want string
	}{
		{"table", "table"},
		{"std", "table"},
		{"standard", "table"},
		{"", "table"},
		{"TABLE", "table"},
		{"table-json", "table-json"},
		{"std-json", "table-json"},
		{"standard-json", "table-json"},
		{"TABLE-JSON", "table-json"},
		{"table-yaml", "table-yaml"},
		{"std-yaml", "table-yaml"},
		{"standard-yaml", "table-yaml"},
		{"json", "json"},
		{"JSON", "json"},
		{"yaml", "yaml"},
		{"YAML", "yaml"},
	}
	for _, c := range cases {
		rootCmd.PersistentFlags().Set("output", c.flag)
		got := resolveOutputFormat()
		if got != c.want {
			t.Errorf("resolveOutputFormat() with flag=%q = %q, want %q", c.flag, got, c.want)
		}
	}
	rootCmd.PersistentFlags().Set("output", "table")
}

func TestPrintTable_TableJSONFormat(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 20},
		{Header: "ID", Width: 30},
		{Header: "RAM(GB)", Width: 10},
	}
	rows := [][]string{
		{"server-01", "abc123", "4"},
		{"server-02", "def456", "8"},
	}

	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "table-json")
		PrintTable(headers, rows)
		rootCmd.PersistentFlags().Set("output", "table")
	})

	var result []map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result))
	}
	if result[0]["name"] != "server-01" {
		t.Errorf("expected name=server-01, got %q", result[0]["name"])
	}
	if result[0]["ram_gb"] != "4" {
		t.Errorf("expected ram_gb=4, got %q", result[0]["ram_gb"])
	}
	if result[1]["id"] != "def456" {
		t.Errorf("expected id=def456, got %q", result[1]["id"])
	}
}

func TestPrintTable_TableJSONFormat_Empty(t *testing.T) {
	headers := []TableColumn{{Header: "NAME", Width: 10}}
	rows := [][]string{}

	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "table-json")
		PrintTable(headers, rows)
		rootCmd.PersistentFlags().Set("output", "table")
	})

	var result []map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("empty output is not valid JSON: %v\noutput: %s", err, out)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestPrintTable_TableYAMLFormat(t *testing.T) {
	headers := []TableColumn{
		{Header: "NAME", Width: 20},
		{Header: "ID", Width: 30},
		{Header: "CREATION DATE", Width: 15},
	}
	rows := [][]string{
		{"project-a", "id-001", "26-03-2026"},
		{"project-b", "id-002", "27-03-2026"},
	}

	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "table-yaml")
		PrintTable(headers, rows)
		rootCmd.PersistentFlags().Set("output", "table")
	})

	if !strings.Contains(out, "name: project-a") {
		t.Errorf("YAML output missing 'name: project-a':\n%s", out)
	}
	if !strings.Contains(out, "id: id-001") {
		t.Errorf("YAML output missing 'id: id-001':\n%s", out)
	}
	if !strings.Contains(out, "creation_date: 26-03-2026") {
		t.Errorf("YAML output missing 'creation_date: 26-03-2026':\n%s", out)
	}
	if !strings.Contains(out, "- ") {
		t.Errorf("YAML output missing list indicators:\n%s", out)
	}
}

func TestPrintOutput_FullJSON(t *testing.T) {
	type testProject struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	obj := testProject{ID: "proj-001", Name: "my-project"}
	headers := []TableColumn{{Header: "NAME", Width: 20}, {Header: "ID", Width: 30}}
	rows := [][]string{{"my-project", "proj-001"}}

	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "json")
		PrintOutput(obj, headers, rows)
		rootCmd.PersistentFlags().Set("output", "table")
	})

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("full JSON output is not valid JSON: %v\noutput: %s", err, out)
	}
	if result["id"] != "proj-001" {
		t.Errorf("expected id=proj-001, got %q", result["id"])
	}
	if result["name"] != "my-project" {
		t.Errorf("expected name=my-project, got %q", result["name"])
	}
}

func TestPrintOutput_FullYAML(t *testing.T) {
	type testProject struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	obj := testProject{ID: "proj-001", Name: "my-project"}

	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "yaml")
		PrintOutput(obj, nil, nil)
		rootCmd.PersistentFlags().Set("output", "table")
	})

	if !strings.Contains(out, "id: proj-001") {
		t.Errorf("YAML output missing 'id: proj-001':\n%s", out)
	}
	if !strings.Contains(out, "name: my-project") {
		t.Errorf("YAML output missing 'name: my-project':\n%s", out)
	}
}

func TestPrintOutput_FullJSON_NilObject(t *testing.T) {
	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "json")
		PrintOutput(nil, nil, nil)
		rootCmd.PersistentFlags().Set("output", "table")
	})
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected {} for nil object, got %q", out)
	}
}

func TestPrintOutput_AliasStdJSON(t *testing.T) {
	headers := []TableColumn{{Header: "ID", Width: 10}}
	rows := [][]string{{"abc"}}

	out := captureStdout(func() {
		rootCmd.PersistentFlags().Set("output", "std-json")
		PrintTable(headers, rows)
		rootCmd.PersistentFlags().Set("output", "table")
	})

	var result []map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("std-json alias output is not valid JSON: %v\noutput: %s", err, out)
	}
	if len(result) != 1 || result[0]["id"] != "abc" {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestGetArubaClient_DebugFlagChange(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	// Clear cache
	resetClientState()

	// First call without debug
	rootCmd.PersistentFlags().Set("debug", "false")
	client1, err1 := GetArubaClient()
	if err1 != nil {
		t.Fatalf("GetArubaClient() error = %v", err1)
	}

	// Change debug flag
	rootCmd.PersistentFlags().Set("debug", "true")

	// Second call should create new client (cache invalidated)
	client2, err2 := GetArubaClient()
	if err2 != nil {
		t.Fatalf("GetArubaClient() second call error = %v", err2)
	}

	// Should be different instances (cache invalidated due to debug flag change)
	if client1 == client2 {
		t.Error("GetArubaClient() should return new client when debug flag changes")
	}
}

func TestGetArubaClient_CredentialChange(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	if os.Getenv("USERPROFILE") != "" {
		os.Setenv("USERPROFILE", tmpDir)
	}
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
		// Clear client cache
		resetClientState()
	}()

	// Create initial config
	config1 := &Config{
		ClientID:     "client-1",
		ClientSecret: "secret-1",
	}
	err := SaveConfig(config1)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Clear cache
	resetClientState()

	// First call
	client1, err1 := GetArubaClient()
	if err1 != nil {
		t.Fatalf("GetArubaClient() error = %v", err1)
	}

	// Change credentials
	config2 := &Config{
		ClientID:     "client-2",
		ClientSecret: "secret-2",
	}
	err = SaveConfig(config2)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Second call should create new client (cache invalidated due to credential change)
	client2, err2 := GetArubaClient()
	if err2 != nil {
		t.Fatalf("GetArubaClient() second call error = %v", err2)
	}

	// Should be different instances (cache invalidated due to credential change)
	if client1 == client2 {
		t.Error("GetArubaClient() should return new client when credentials change")
	}
}

func TestFmtAPIError_And_APIErrFromResp(t *testing.T) {
	tests := []struct {
		name    string
		errFunc func() error
		want    string
		notWant string
	}{
		{
			name:    "fmtAPIError nil title and detail",
			errFunc: func() error { return fmtAPIError(404, nil, nil) },
			want:    "API error (status 404)",
			notWant: ":",
		},
		{
			name:    "fmtAPIError with title only",
			errFunc: func() error { return fmtAPIError(500, strPtr("Bad Gateway"), nil) },
			want:    "API error (status 500): Bad Gateway",
		},
		{
			name:    "fmtAPIError with detail only",
			errFunc: func() error { return fmtAPIError(502, nil, strPtr("upstream timeout")) },
			want:    "API error (status 502) \u2014 upstream timeout",
		},
		{
			name:    "fmtAPIError with title and detail",
			errFunc: func() error { return fmtAPIError(404, strPtr("Not Found"), strPtr("resource not found")) },
			want:    "API error (status 404): Not Found \u2014 resource not found",
		},
		{
			name:    "apiErrFromResp nil errResp",
			errFunc: func() error { return apiErrFromResp(404, nil) },
			want:    "API error (status 404)",
		},
		{
			name:    "apiErrFromResp 2xx returns nil",
			errFunc: func() error { return apiErrFromResp(200, nil) },
		},
		{
			name: "apiErrFromResp with errResp",
			errFunc: func() error {
				return apiErrFromResp(404, &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")})
			},
			want: "API error (status 404): Not Found \u2014 resource not found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.errFunc()
			if tc.want == "" {
				if err != nil {
					t.Errorf("expected nil error, got %q", err.Error())
				}
				return
			}
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %q, want substring %q", err.Error(), tc.want)
			}
			if tc.notWant != "" && strings.Contains(err.Error(), tc.notWant) {
				t.Errorf("got %q, but should not contain %q", err.Error(), tc.notWant)
			}
		})
	}
}

func TestMsgHelpers(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"msgCreated", msgCreated("VPC", "vpc-1"), "VPC 'vpc-1' created successfully."},
		{"msgCreatedAsync", msgCreatedAsync("KaaS", "k-1"), "KaaS 'k-1' creation initiated."},
		{"msgUpdated", msgUpdated("Subnet", "sn-1"), "Subnet 'sn-1' updated successfully."},
		{"msgUpdatedAsync", msgUpdatedAsync("CloudServer", "cs-1"), "CloudServer 'cs-1' update initiated."},
		{"msgDeleted", msgDeleted("VPN tunnel", "tun-1"), "VPN tunnel 'tun-1' deleted successfully."},
		{"msgAction", msgAction("Snapshot", "snap-1", "restored"), "Snapshot 'snap-1' restored successfully."},
		{"msgDryRun", msgDryRun("Block storage", "vol-1"), "vol-1"},
	}
	for _, c := range cases {
		if !strings.Contains(c.got, c.want) {
			t.Errorf("%s: got %q, want substring %q", c.name, c.got, c.want)
		}
	}
}

func TestListParams(t *testing.T) {
	// Build a temp command with limit/offset flags (same as all list commands)
	cmd := &cobra.Command{}
	cmd.Flags().Int32("limit", 0, "")
	cmd.Flags().Int32("offset", 0, "")

	// nil when both flags are zero
	p := listParams(cmd)
	if p != nil {
		t.Errorf("expected nil params for zero flags, got %+v", p)
	}

	// non-nil when limit is set
	if err := cmd.Flags().Set("limit", "10"); err != nil {
		t.Fatalf("setting limit: %v", err)
	}
	p = listParams(cmd)
	if p == nil || p.Limit == nil || *p.Limit != 10 {
		t.Errorf("expected limit=10, got %+v", p)
	}

	// non-nil when offset is set
	if err := cmd.Flags().Set("limit", "0"); err != nil {
		t.Fatalf("resetting limit: %v", err)
	}
	if err := cmd.Flags().Set("offset", "5"); err != nil {
		t.Fatalf("setting offset: %v", err)
	}
	p = listParams(cmd)
	if p == nil || p.Offset == nil || *p.Offset != 5 {
		t.Errorf("expected offset=5, got %+v", p)
	}
}

func TestResolveOutputFormat_AllFormats(t *testing.T) {
	cases := []struct{ flag, want string }{
		{"json", OutputFormatJSON},
		{"yaml", OutputFormatYAML},
		{"table-json", OutputFormatTableJSON},
		{"table-yaml", OutputFormatTableYAML},
		{"table", OutputFormatTable},
		{"std", OutputFormatTable},
		{"standard", OutputFormatTable},
		{"", OutputFormatTable},
		{"std-json", OutputFormatTableJSON},
		{"std-yaml", OutputFormatTableYAML},
		{"unknown-xyz", OutputFormatTable}, // falls back to table
	}
	for _, c := range cases {
		resetCmdFlags(rootCmd)
		if c.flag != "" {
			_ = rootCmd.PersistentFlags().Set("output", c.flag)
		}
		got := resolveOutputFormat()
		if got != c.want {
			t.Errorf("resolveOutputFormat() with --output=%q: got %q, want %q", c.flag, got, c.want)
		}
	}
	resetCmdFlags(rootCmd)
}

func TestPrintJSON_NilObject(t *testing.T) {
	out := captureStdout(func() { printJSON(nil) })
	if !strings.Contains(out, "{}") {
		t.Errorf("printJSON(nil) should print {}, got: %s", out)
	}
}

func TestPrintYAML_NilObject(t *testing.T) {
	out := captureStdout(func() { printYAML(nil) })
	if !strings.Contains(out, "{}") {
		t.Errorf("printYAML(nil) should print {}, got: %s", out)
	}
}

func TestPrintYAML_ValidObject(t *testing.T) {
	obj := map[string]string{"key": "value"}
	out := captureStdout(func() { printYAML(obj) })
	if !strings.Contains(out, "key") || !strings.Contains(out, "value") {
		t.Errorf("printYAML() output missing key/value, got: %s", out)
	}
}

func TestPrintTableJSON_Empty(t *testing.T) {
	out := captureStdout(func() {
		printTableJSON([]TableColumn{{Header: "NAME", Width: 10}}, [][]string{})
	})
	if !strings.Contains(out, "[]") {
		t.Errorf("printTableJSON with empty rows should print [], got: %s", out)
	}
}

func TestRowsToRecords(t *testing.T) {
	headers := []TableColumn{{Header: "NAME", Width: 10}, {Header: "ID", Width: 10}}
	rows := [][]string{{"alice", "id-1"}, {"bob", "id-2"}}
	records := rowsToRecords(headers, rows)
	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestRowsToRecords_ShortRow(t *testing.T) {
	headers := []TableColumn{{Header: "NAME", Width: 10}, {Header: "ID", Width: 10}, {Header: "EXTRA", Width: 10}}
	rows := [][]string{{"alice", "id-1"}} // fewer columns than headers
	records := rowsToRecords(headers, rows)
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
	// Should only have 2 key-value pairs (NAME and ID), not 3
	if len(records[0].Content) != 4 { // 2 pairs * 2 nodes each
		t.Errorf("expected 4 content nodes (2 pairs), got %d", len(records[0].Content))
	}
}

func TestPrintTableJSON_MultipleRows(t *testing.T) {
	headers := []TableColumn{{Header: "NAME", Width: 10}, {Header: "ID", Width: 10}}
	rows := [][]string{{"alice", "id-1"}, {"bob", "id-2"}}
	out := captureStdout(func() {
		printTableJSON(headers, rows)
	})
	if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
		t.Errorf("printTableJSON output missing rows, got: %s", out)
	}
}

func TestPrintTableJSON_ShortRow(t *testing.T) {
	headers := []TableColumn{{Header: "NAME", Width: 10}, {Header: "ID", Width: 10}, {Header: "EXTRA", Width: 10}}
	rows := [][]string{{"alice", "id-1"}} // fewer columns than headers
	out := captureStdout(func() {
		printTableJSON(headers, rows)
	})
	if !strings.Contains(out, "alice") {
		t.Errorf("printTableJSON short row output missing alice, got: %s", out)
	}
}

func TestResolveOutputFormat_Unknown(t *testing.T) {
	rootCmd.PersistentFlags().Set("output", "invalid-format")
	got := resolveOutputFormat()
	if got != OutputFormatTable {
		t.Errorf("resolveOutputFormat() with unknown flag = %q, want %q", got, OutputFormatTable)
	}
	rootCmd.PersistentFlags().Set("output", "table")
}

func TestPrintJSON_MarshalError(t *testing.T) {
	// channels cannot be marshalled to JSON — triggers the error branch
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	printJSON(make(chan int))
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "error marshalling to JSON") {
		t.Errorf("expected marshal error message, got: %s", buf.String())
	}
}

func TestPrintYAML_MarshalError(t *testing.T) {
	// channels cannot be marshalled to JSON — triggers the error branch in printYAML
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	printYAML(make(chan int))
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	if !strings.Contains(buf.String(), "error marshalling to JSON for YAML conversion") {
		t.Errorf("expected marshal error message, got: %s", buf.String())
	}
}
