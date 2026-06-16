package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ─── xdgCacheDir ─────────────────────────────────────────────────────────────

func TestXDGCacheDir_EnvVar(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	got, err := xdgCacheDir()
	if err != nil {
		t.Fatalf("xdgCacheDir() error: %v", err)
	}
	if got != tmp {
		t.Errorf("xdgCacheDir() = %q, want %q", got, tmp)
	}
}

func TestXDGCacheDir_FallbackToHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", tmp)

	got, err := xdgCacheDir()
	if err != nil {
		t.Fatalf("xdgCacheDir() error: %v", err)
	}
	want := filepath.Join(tmp, ".cache")
	if got != want {
		t.Errorf("xdgCacheDir() = %q, want %q", got, want)
	}
}

// ─── completionCachePath ─────────────────────────────────────────────────────

func TestCompletionCachePath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	got, err := completionCachePath("vpc-proj-123")
	if err != nil {
		t.Fatalf("completionCachePath() error: %v", err)
	}
	want := filepath.Join(tmp, "acloud", "completion", "vpc-proj-123.json")
	if got != want {
		t.Errorf("completionCachePath() = %q, want %q", got, want)
	}
}

// ─── completionTTL ───────────────────────────────────────────────────────────

func TestCompletionTTL_Default(t *testing.T) {
	t.Setenv("ACLOUD_COMPLETION_CACHE_TTL", "")
	if got := completionTTL(); got != 60*time.Second {
		t.Errorf("completionTTL() = %v, want 60s", got)
	}
}

func TestCompletionTTL_EnvVar(t *testing.T) {
	t.Setenv("ACLOUD_COMPLETION_CACHE_TTL", "5m")
	if got := completionTTL(); got != 5*time.Minute {
		t.Errorf("completionTTL() = %v, want 5m", got)
	}
}

func TestCompletionTTL_InvalidFallsBack(t *testing.T) {
	t.Setenv("ACLOUD_COMPLETION_CACHE_TTL", "not-a-duration")
	if got := completionTTL(); got != 60*time.Second {
		t.Errorf("completionTTL() with invalid env = %v, want 60s fallback", got)
	}
}

func TestCompletionTTL_ZeroFallsBack(t *testing.T) {
	t.Setenv("ACLOUD_COMPLETION_CACHE_TTL", "0s")
	if got := completionTTL(); got != 60*time.Second {
		t.Errorf("completionTTL() with 0s env = %v, want 60s fallback", got)
	}
}

// ─── completionCacheGet / completionCachePut (disk path) ─────────────────────

func TestCompletionCache_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	// Disable the in-memory override so we exercise disk I/O.
	completionCacheOverrideMu.Lock()
	prev := completionCacheOverride
	completionCacheOverride = nil
	completionCacheOverrideMu.Unlock()
	t.Cleanup(func() {
		completionCacheOverrideMu.Lock()
		completionCacheOverride = prev
		completionCacheOverrideMu.Unlock()
	})

	entries := []string{"vol-001\tmy-volume", "vol-002\tother-volume"}
	completionCachePut("blockstorage-proj-123", entries)

	got, err := completionCacheGet("blockstorage-proj-123")
	if err != nil {
		t.Fatalf("completionCacheGet() error: %v", err)
	}
	if len(got) != 2 || got[0] != entries[0] || got[1] != entries[1] {
		t.Errorf("round-trip mismatch: got %v, want %v", got, entries)
	}
}

func TestCompletionCacheGet_Miss(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	completionCacheOverrideMu.Lock()
	prev := completionCacheOverride
	completionCacheOverride = nil
	completionCacheOverrideMu.Unlock()
	t.Cleanup(func() {
		completionCacheOverrideMu.Lock()
		completionCacheOverride = prev
		completionCacheOverrideMu.Unlock()
	})

	got, err := completionCacheGet("nonexistent-key")
	if err != nil || got != nil {
		t.Errorf("expected nil, nil on cache miss; got %v, %v", got, err)
	}
}

func TestCompletionCacheGet_Expired(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	t.Setenv("ACLOUD_COMPLETION_CACHE_TTL", "1ms")
	completionCacheOverrideMu.Lock()
	prev := completionCacheOverride
	completionCacheOverride = nil
	completionCacheOverrideMu.Unlock()
	t.Cleanup(func() {
		completionCacheOverrideMu.Lock()
		completionCacheOverride = prev
		completionCacheOverrideMu.Unlock()
	})

	// Write a cache file with an old timestamp.
	path := filepath.Join(tmp, "acloud", "completion", "stale-key.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	stale := completionCacheEntry{
		Entries:  []string{"id-001\tname"},
		CachedAt: time.Now().Add(-10 * time.Second).UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(stale)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	got, err := completionCacheGet("stale-key")
	if err != nil || got != nil {
		t.Errorf("expected nil on expired entry; got %v, %v", got, err)
	}
}

func TestCompletionCacheGet_CorruptJSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	completionCacheOverrideMu.Lock()
	prev := completionCacheOverride
	completionCacheOverride = nil
	completionCacheOverrideMu.Unlock()
	t.Cleanup(func() {
		completionCacheOverrideMu.Lock()
		completionCacheOverride = prev
		completionCacheOverrideMu.Unlock()
	})

	path := filepath.Join(tmp, "acloud", "completion", "bad-key.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not json {{{"), 0600); err != nil {
		t.Fatal(err)
	}

	got, err := completionCacheGet("bad-key")
	if err != nil || got != nil {
		t.Errorf("expected nil on corrupt JSON; got %v, %v", got, err)
	}
}

func TestCompletionCacheGet_BadTimestamp(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	completionCacheOverrideMu.Lock()
	prev := completionCacheOverride
	completionCacheOverride = nil
	completionCacheOverrideMu.Unlock()
	t.Cleanup(func() {
		completionCacheOverrideMu.Lock()
		completionCacheOverride = prev
		completionCacheOverrideMu.Unlock()
	})

	path := filepath.Join(tmp, "acloud", "completion", "badts-key.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	bad := completionCacheEntry{Entries: []string{"id\tname"}, CachedAt: "not-a-time"}
	data, _ := json.Marshal(bad)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	got, _ := completionCacheGet("badts-key")
	if got != nil {
		t.Errorf("expected nil on bad timestamp, got %v", got)
	}
}

// ─── cacheKey ────────────────────────────────────────────────────────────────

func TestCacheKey(t *testing.T) {
	cases := []struct {
		parts []string
		want  string
	}{
		{[]string{"vpc", "proj-123"}, "vpc-proj-123"},
		{[]string{"blockstorage", "p1"}, "blockstorage-p1"},
		{[]string{"database", "proj/abc", "dbaas-01"}, "database-proj_abc-dbaas-01"},
		{[]string{"grant", "p1", "d1", "db name"}, "grant-p1-d1-db_name"},
		{[]string{"", "p1"}, "p1"},
		{[]string{"project"}, "project"},
	}
	for _, c := range cases {
		got := cacheKey(c.parts...)
		if got != c.want {
			t.Errorf("cacheKey(%v) = %q, want %q", c.parts, got, c.want)
		}
	}
}

// ─── filterCompletions ───────────────────────────────────────────────────────

func TestFilterCompletions(t *testing.T) {
	entries := []string{"aaa-001\tname-a", "bbb-002\tname-b", "aaa-003\tname-c"}

	t.Run("empty prefix returns all", func(t *testing.T) {
		got := filterCompletions(entries, "")
		if len(got) != 3 {
			t.Errorf("expected 3, got %d", len(got))
		}
	})

	t.Run("prefix match", func(t *testing.T) {
		got := filterCompletions(entries, "aaa")
		if len(got) != 2 {
			t.Errorf("expected 2 for prefix 'aaa', got %d: %v", len(got), got)
		}
	})

	t.Run("no match", func(t *testing.T) {
		got := filterCompletions(entries, "zzz")
		if len(got) != 0 {
			t.Errorf("expected 0 for prefix 'zzz', got %d", len(got))
		}
	})

	t.Run("entry without tab uses full string", func(t *testing.T) {
		got := filterCompletions([]string{"bare-id"}, "bare")
		if len(got) != 1 || got[0] != "bare-id" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("nil entries returns nil-safe empty", func(t *testing.T) {
		got := filterCompletions(nil, "x")
		if got == nil {
			// filterCompletions uses entries[:0:0] — result for nil input with non-empty prefix
			// is an empty non-nil slice or nil; either is acceptable.
		}
		if len(got) != 0 {
			t.Errorf("expected 0 results for nil entries, got %d", len(got))
		}
	})
}

// ─── in-memory override path ─────────────────────────────────────────────────

func TestCompletionCacheOverride_HitAndMiss(t *testing.T) {
	// Activate the in-memory override (mimics what resetClientState does in tests).
	completionCacheReset()
	t.Cleanup(func() {
		completionCacheOverrideMu.Lock()
		completionCacheOverride = nil
		completionCacheOverrideMu.Unlock()
	})

	// Miss before any Put.
	got, _ := completionCacheGet("k1")
	if got != nil {
		t.Errorf("expected miss before Put, got %v", got)
	}

	// Put then Get.
	completionCachePut("k1", []string{"id-1\tname-1"})
	got, _ = completionCacheGet("k1")
	if len(got) != 1 || got[0] != "id-1\tname-1" {
		t.Errorf("in-memory round-trip failed: %v", got)
	}
}

// ─── cacheKey special characters ─────────────────────────────────────────────

func TestCacheKey_SpecialChars(t *testing.T) {
	// Slashes, spaces, and other unsafe chars are stripped/replaced.
	got := cacheKey("res", "proj/with spaces!")
	if strings.Contains(got, "/") || strings.Contains(got, " ") || strings.Contains(got, "!") {
		t.Errorf("cacheKey produced unsafe filename: %q", got)
	}
	if got == "" {
		t.Errorf("cacheKey returned empty string for non-empty input")
	}
}
