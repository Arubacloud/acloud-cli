package logging

import (
	"log/slog"
	"os"
	"sync"
)

var (
	mu     sync.RWMutex
	global *slog.Logger
)

// Setup initialises the package-level slog logger from the resolved level and
// format strings. All output goes to stderr so it never corrupts -o json/yaml
// stdout output.
//
// level:  "debug" | "info" | "warn" | "error" (default: "info")
// format: "text" | "json"                      (default: "text")
func Setup(level, format string) {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}

	var h slog.Handler
	if format == "json" {
		h = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		h = slog.NewTextHandler(os.Stderr, opts)
	}

	mu.Lock()
	global = slog.New(h)
	mu.Unlock()
}

// L returns the package-level slog.Logger. If Setup has not yet been called,
// it falls back to the default slog logger.
func L() *slog.Logger {
	mu.RLock()
	l := global
	mu.RUnlock()
	if l == nil {
		return slog.Default()
	}
	return l
}

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
