package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// completionCacheOverride is an in-memory override used in tests to prevent
// cross-test cache pollution. When non-nil, disk I/O is bypassed entirely.
var completionCacheOverride map[string][]string
var completionCacheOverrideMu sync.Mutex

// completionCacheReset clears the in-memory cache override, forcing the next
// call to completionCacheGet to miss. Called from resetClientState during tests.
func completionCacheReset() {
	completionCacheOverrideMu.Lock()
	defer completionCacheOverrideMu.Unlock()
	completionCacheOverride = make(map[string][]string)
}

// completionCacheEntry is the on-disk shape of a completion cache file.
type completionCacheEntry struct {
	Entries  []string `json:"entries"`
	CachedAt string   `json:"cachedAt"`
}

// xdgCacheDir returns $XDG_CACHE_HOME, falling back to ~/.cache (#179).
func xdgCacheDir() (string, error) {
	if d := os.Getenv("XDG_CACHE_HOME"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache"), nil
}

// completionCachePath returns the cache file path for a given resource key.
// The key should be stable and safe for use in a filename (no slashes).
func completionCachePath(key string) (string, error) {
	dir, err := xdgCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "acloud", "completion", key+".json"), nil
}

// completionTTL returns the configured TTL from ACLOUD_COMPLETION_CACHE_TTL
// (default 60 s) (#179).
func completionTTL() time.Duration {
	if raw := os.Getenv("ACLOUD_COMPLETION_CACHE_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			return d
		}
	}
	return 60 * time.Second
}

// completionCacheGet returns cached entries for key if they are within the TTL.
// Returns nil, nil on any miss (file absent, expired, or unreadable).
// When the in-memory override is active (set by completionCacheReset in tests),
// disk I/O is bypassed.
func completionCacheGet(key string) ([]string, error) {
	completionCacheOverrideMu.Lock()
	override := completionCacheOverride
	completionCacheOverrideMu.Unlock()
	if override != nil {
		if entries, ok := override[key]; ok {
			return entries, nil
		}
		return nil, nil
	}

	path, err := completionCachePath(key)
	if err != nil {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	var entry completionCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, nil
	}
	cachedAt, err := time.Parse(time.RFC3339, entry.CachedAt)
	if err != nil || time.Since(cachedAt) > completionTTL() {
		return nil, nil
	}
	return entry.Entries, nil
}

// completionCachePut writes entries to the cache for key. Errors are silently
// discarded — a failed write degrades to no-cache, not a hard failure.
// When the in-memory override is active (set by completionCacheReset in tests),
// disk I/O is bypassed.
func completionCachePut(key string, entries []string) {
	completionCacheOverrideMu.Lock()
	override := completionCacheOverride
	completionCacheOverrideMu.Unlock()
	if override != nil {
		completionCacheOverrideMu.Lock()
		completionCacheOverride[key] = entries
		completionCacheOverrideMu.Unlock()
		return
	}

	path, err := completionCachePath(key)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	entry := completionCacheEntry{
		Entries:  entries,
		CachedAt: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0600)
}

// cacheKey builds a canonical cache key from a resource name and additional
// identifiers (project ID, parent resource IDs). Segments are joined with "-"
// after stripping any path-unsafe characters.
func cacheKey(parts ...string) string {
	safe := make([]string, 0, len(parts))
	for _, p := range parts {
		// Replace non-alphanumeric runs with "_" to keep filenames safe.
		var b strings.Builder
		prevSafe := false
		for _, r := range p {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
				b.WriteRune(r)
				prevSafe = true
			} else if prevSafe {
				b.WriteRune('_')
				prevSafe = false
			}
		}
		s := strings.Trim(b.String(), "_-")
		if s != "" {
			safe = append(safe, s)
		}
	}
	return strings.Join(safe, "-")
}

// filterCompletions returns entries whose ID portion starts with prefix.
// If prefix is empty all entries are returned. The ID portion is the text
// before the first tab (the description separator used by Cobra).
func filterCompletions(entries []string, prefix string) []string {
	if prefix == "" {
		return entries
	}
	out := entries[:0:0] // nil-safe empty slice
	for _, e := range entries {
		id := e
		if idx := strings.IndexByte(e, '\t'); idx >= 0 {
			id = e[:idx]
		}
		if strings.HasPrefix(id, prefix) {
			out = append(out, e)
		}
	}
	return out
}
