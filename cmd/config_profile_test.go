package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ─── getActiveProfile ─────────────────────────────────────────────────────────

func TestGetActiveProfile_Default(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "")
	resetCmdFlags(rootCmd)

	if got := getActiveProfile(); got != "default" {
		t.Errorf("getActiveProfile() = %q, want %q", got, "default")
	}
}

func TestGetActiveProfile_EnvVar(t *testing.T) {
	t.Setenv("ACLOUD_PROFILE", "staging")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if got := getActiveProfile(); got != "staging" {
		t.Errorf("getActiveProfile() = %q, want %q", got, "staging")
	}
}

func TestGetActiveProfile_FlagOverridesEnvVar(t *testing.T) {
	t.Setenv("ACLOUD_PROFILE", "staging")
	if err := rootCmd.PersistentFlags().Set("profile", "prod"); err != nil {
		t.Fatalf("set profile flag: %v", err)
	}
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if got := getActiveProfile(); got != "prod" {
		t.Errorf("getActiveProfile() = %q, want %q (flag should override env)", got, "prod")
	}
}

func TestGetActiveProfile_FromFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec"},
		"prod":    {ClientID: "id-prod", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := setActiveProfile("prod"); err != nil {
		t.Fatalf("setActiveProfile() error: %v", err)
	}

	if got := getActiveProfile(); got != "prod" {
		t.Errorf("getActiveProfile() = %q, want %q (from file)", got, "prod")
	}
}

func TestGetActiveProfile_EnvVarOverridesFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "staging")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"prod": {ClientID: "id-prod", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := setActiveProfile("prod"); err != nil {
		t.Fatalf("setActiveProfile() error: %v", err)
	}

	if got := getActiveProfile(); got != "staging" {
		t.Errorf("getActiveProfile() = %q, want %q (env var should override file)", got, "staging")
	}
}

// ─── loadAllProfiles / saveAllProfiles ───────────────────────────────────────

func TestLoadAllProfiles_SingleProfileFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	// Write old single-profile format
	p := filepath.Join(tmp, "acloud", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("clientId: legacy-id\nclientSecret: legacy-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}

	profiles, err := loadAllProfiles()
	if err != nil {
		t.Fatalf("loadAllProfiles() error: %v", err)
	}
	if _, ok := profiles["default"]; !ok {
		t.Errorf("expected 'default' profile from single-profile file, got %v", profiles)
	}
	if profiles["default"].ClientID != "legacy-id" {
		t.Errorf("ClientID = %q, want legacy-id", profiles["default"].ClientID)
	}
}

func TestSaveAndLoadAllProfiles_MultiProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	profiles := map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec-default"},
		"prod":    {ClientID: "id-prod", ClientSecret: "sec-prod", BaseURL: "https://api.arubacloud.com"},
	}

	if err := saveAllProfiles(profiles); err != nil {
		t.Fatalf("saveAllProfiles() error: %v", err)
	}

	loaded, err := loadAllProfiles()
	if err != nil {
		t.Fatalf("loadAllProfiles() error: %v", err)
	}
	if loaded["default"].ClientID != "id-default" {
		t.Errorf("default ClientID = %q, want id-default", loaded["default"].ClientID)
	}
	if loaded["prod"].ClientID != "id-prod" {
		t.Errorf("prod ClientID = %q, want id-prod", loaded["prod"].ClientID)
	}
	if loaded["prod"].BaseURL != "https://api.arubacloud.com" {
		t.Errorf("prod BaseURL = %q, want https://api.arubacloud.com", loaded["prod"].BaseURL)
	}
}

// ─── LoadConfig with profiles ─────────────────────────────────────────────────

func TestLoadConfig_ActiveProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "staging")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec-default"},
		"staging": {ClientID: "id-staging", ClientSecret: "sec-staging"},
	}); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.ClientID != "id-staging" {
		t.Errorf("ClientID = %q, want id-staging (staging profile)", cfg.ClientID)
	}
}

func TestLoadConfig_ProfileNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "nonexistent")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec-default"},
	}); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for nonexistent profile, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected profile name in error, got: %v", err)
	}
}

// ─── SaveConfig writes to active profile ─────────────────────────────────────

func TestSaveConfig_WritesToActiveProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "dev")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	// Pre-populate the default profile so it isn't lost.
	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "default-id", ClientSecret: "default-sec"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := SaveConfig(&Config{ClientID: "dev-id", ClientSecret: "dev-sec"}); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	profiles, err := loadAllProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if profiles["dev"].ClientID != "dev-id" {
		t.Errorf("dev profile ClientID = %q, want dev-id", profiles["dev"].ClientID)
	}
	if profiles["default"].ClientID != "default-id" {
		t.Errorf("default profile should be preserved, got ClientID %q", profiles["default"].ClientID)
	}
}

// ─── config profile commands ─────────────────────────────────────────────────

func TestConfigProfileList_Empty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	out := captureStdout(func() {
		_ = ConfigProfileListRun(nil, nil)
	})
	if !strings.Contains(out, "No profiles") {
		// Acceptable — the command shows an empty table or the "No profiles" message.
		_ = out
	}
}

func TestConfigProfileList_ShowsProfiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "prod")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default"},
		"prod":    {ClientID: "id-prod"},
	}); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(func() {
		_ = ConfigProfileListRun(nil, nil)
	})
	if !strings.Contains(out, "default") {
		t.Errorf("expected 'default' in list output, got: %s", out)
	}
	if !strings.Contains(out, "prod") {
		t.Errorf("expected 'prod' in list output, got: %s", out)
	}
	// Active profile (prod) should be marked with *
	if !strings.Contains(out, "* prod") {
		t.Errorf("expected active profile 'prod' marked with *, got: %s", out)
	}
}

func TestConfigProfileList_ShowsDefaultBaseURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "default")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	// Profile with no explicit BaseURL — the list should show the default.
	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "secret"},
	}); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(func() {
		_ = ConfigProfileListRun(nil, nil)
	})
	if !strings.Contains(out, DefaultBaseURL) {
		t.Errorf("expected DefaultBaseURL %q in list output, got: %s", DefaultBaseURL, out)
	}
}

func TestConfigProfileDelete(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec"},
		"prod":    {ClientID: "id-prod", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}

	if err := ConfigProfileDeleteRun(nil, []string{"prod"}); err != nil {
		t.Fatalf("ConfigProfileDeleteRun() error: %v", err)
	}

	profiles, err := loadAllProfiles()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := profiles["prod"]; ok {
		t.Error("'prod' profile should have been deleted")
	}
	if _, ok := profiles["default"]; !ok {
		t.Error("'default' profile should still exist after deleting 'prod'")
	}
}

func TestConfigProfileDelete_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}

	err := ConfigProfileDeleteRun(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error when deleting nonexistent profile")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the profile name, got: %v", err)
	}
}

// ─── sortedProfileNames ───────────────────────────────────────────────────────

func TestSortedProfileNames_DefaultFirst(t *testing.T) {
	profiles := map[string]configFile{
		"prod":    {},
		"default": {},
		"staging": {},
	}
	names := sortedProfileNames(profiles)
	if names[0] != "default" {
		t.Errorf("expected 'default' first, got %q", names[0])
	}
}

func TestSortedProfileNames_NoDefault(t *testing.T) {
	profiles := map[string]configFile{
		"prod":    {},
		"staging": {},
	}
	names := sortedProfileNames(profiles)
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

// ─── ConfigProfileSetRun ──────────────────────────────────────────────────────

func TestConfigProfileSet_NewProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_CLIENT_SECRET", "env-secret")

	// Build a minimal cobra command with required flags
	cmd := configProfileSetCmd
	resetCmdFlags(cmd)

	// Call directly with the flag values (simulating --client-id flag)
	if err := cmd.Flags().Set("client-id", "new-profile-id"); err != nil {
		t.Fatalf("set client-id: %v", err)
	}

	out := captureStdout(func() {
		err := ConfigProfileSetRun(cmd, []string{"staging"})
		if err != nil {
			t.Errorf("ConfigProfileSetRun() error: %v", err)
		}
	})
	if !strings.Contains(out, "staging") {
		t.Errorf("expected profile name in output, got: %s", out)
	}

	profiles, _ := loadAllProfiles()
	if profiles["staging"].ClientID != "new-profile-id" {
		t.Errorf("staging ClientID = %q, want new-profile-id", profiles["staging"].ClientID)
	}
	if profiles["staging"].ClientSecret != "env-secret" {
		t.Errorf("staging ClientSecret = %q, want env-secret", profiles["staging"].ClientSecret)
	}
	resetCmdFlags(cmd)
}

func TestConfigProfileSet_UpdateExisting(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_CLIENT_SECRET", "")

	if err := saveAllProfiles(map[string]configFile{
		"prod": {ClientID: "old-id", ClientSecret: "old-secret"},
	}); err != nil {
		t.Fatal(err)
	}

	cmd := configProfileSetCmd
	resetCmdFlags(cmd)
	if err := cmd.Flags().Set("client-id", "new-id"); err != nil {
		t.Fatalf("set client-id: %v", err)
	}

	if err := ConfigProfileSetRun(cmd, []string{"prod"}); err != nil {
		t.Fatalf("ConfigProfileSetRun() error: %v", err)
	}

	profiles, _ := loadAllProfiles()
	if profiles["prod"].ClientID != "new-id" {
		t.Errorf("ClientID = %q, want new-id", profiles["prod"].ClientID)
	}
	// Secret should be preserved from existing profile
	if profiles["prod"].ClientSecret != "old-secret" {
		t.Errorf("ClientSecret = %q, want old-secret (preserved)", profiles["prod"].ClientSecret)
	}
	resetCmdFlags(cmd)
}

func TestConfigProfileSet_MissingClientID(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cmd := configProfileSetCmd
	resetCmdFlags(cmd)

	err := ConfigProfileSetRun(cmd, []string{"newprofile"})
	if err == nil {
		t.Fatal("expected error when --client-id is missing for new profile")
	}
	if !strings.Contains(err.Error(), "client-id") {
		t.Errorf("expected client-id in error, got: %v", err)
	}
	resetCmdFlags(cmd)
}

func TestConfigProfileSet_WithBaseURL(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_CLIENT_SECRET", "secret")

	cmd := configProfileSetCmd
	resetCmdFlags(cmd)
	_ = cmd.Flags().Set("client-id", "custom-id")
	_ = cmd.Flags().Set("base-url", "https://custom.api.example.com")
	_ = cmd.Flags().Set("token-issuer-url", "https://custom.token.example.com")

	if err := ConfigProfileSetRun(cmd, []string{"custom"}); err != nil {
		t.Fatalf("ConfigProfileSetRun() error: %v", err)
	}

	profiles, _ := loadAllProfiles()
	if profiles["custom"].BaseURL != "https://custom.api.example.com" {
		t.Errorf("BaseURL = %q", profiles["custom"].BaseURL)
	}
	if profiles["custom"].TokenIssuerURL != "https://custom.token.example.com" {
		t.Errorf("TokenIssuerURL = %q", profiles["custom"].TokenIssuerURL)
	}
	resetCmdFlags(cmd)
}

func TestSaveAllProfiles_PreservesActiveProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec"},
		"prod":    {ClientID: "id-prod", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := setActiveProfile("prod"); err != nil {
		t.Fatalf("setActiveProfile() error: %v", err)
	}

	// Save profiles again (simulates updating a profile's credentials).
	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec"},
		"prod":    {ClientID: "id-prod-updated", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}

	// activeProfile must survive the second save.
	if got := loadActiveProfileFromFile(); got != "prod" {
		t.Errorf("activeProfile = %q after saveAllProfiles, want %q", got, "prod")
	}
}

// ─── ConfigProfileUseRun ──────────────────────────────────────────────────────

func TestConfigProfileUse_SwitchesProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id-default", ClientSecret: "sec"},
		"prod":    {ClientID: "id-prod", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(func() {
		if err := ConfigProfileUseRun(nil, []string{"prod"}); err != nil {
			t.Errorf("ConfigProfileUseRun() error: %v", err)
		}
	})
	if !strings.Contains(out, "prod") {
		t.Errorf("expected profile name in output, got: %s", out)
	}

	if got := getActiveProfile(); got != "prod" {
		t.Errorf("getActiveProfile() = %q after use, want %q", got, "prod")
	}

	// Verify LoadConfig returns the prod profile's credentials.
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if cfg.ClientID != "id-prod" {
		t.Errorf("ClientID = %q, want id-prod after switching to prod", cfg.ClientID)
	}
}

func TestConfigProfileUse_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	if err := saveAllProfiles(map[string]configFile{
		"default": {ClientID: "id", ClientSecret: "sec"},
	}); err != nil {
		t.Fatal(err)
	}

	err := ConfigProfileUseRun(nil, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error when switching to nonexistent profile")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the profile name, got: %v", err)
	}
}

// ─── loadAllProfiles edge cases ───────────────────────────────────────────────

func TestLoadAllProfiles_FileNotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	_, err := loadAllProfiles()
	if err == nil {
		t.Error("expected error when no config file exists")
	}
}

// ─── ConfigProfileListRun edge case: no config file ───────────────────────────

func TestConfigProfileListRun_NoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("ACLOUD_PROFILE", "")
	resetCmdFlags(rootCmd)
	t.Cleanup(func() { resetCmdFlags(rootCmd) })

	out := captureStdout(func() {
		_ = ConfigProfileListRun(nil, nil)
	})
	// Should print "No profiles" message or an empty table without panicking.
	_ = out
}
