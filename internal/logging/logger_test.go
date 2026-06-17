package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestSetup_DefaultIsInfo(t *testing.T) {
	Setup("info", "text")
	l := L()
	if !l.Enabled(nil, slog.LevelInfo) {
		t.Error("expected INFO to be enabled")
	}
	if l.Enabled(nil, slog.LevelDebug) {
		t.Error("expected DEBUG to be disabled at INFO level")
	}
}

func TestSetup_Debug(t *testing.T) {
	Setup("debug", "text")
	l := L()
	if !l.Enabled(nil, slog.LevelDebug) {
		t.Error("expected DEBUG to be enabled")
	}
}

func TestSetup_Warn(t *testing.T) {
	Setup("warn", "text")
	l := L()
	if !l.Enabled(nil, slog.LevelWarn) {
		t.Error("expected WARN to be enabled")
	}
	if l.Enabled(nil, slog.LevelInfo) {
		t.Error("expected INFO to be disabled at WARN level")
	}
}

func TestSetup_Error(t *testing.T) {
	Setup("error", "text")
	l := L()
	if !l.Enabled(nil, slog.LevelError) {
		t.Error("expected ERROR to be enabled")
	}
	if l.Enabled(nil, slog.LevelWarn) {
		t.Error("expected WARN to be disabled at ERROR level")
	}
}

func TestSetup_UnknownLevelDefaultsToInfo(t *testing.T) {
	Setup("verbose", "text")
	l := L()
	if !l.Enabled(nil, slog.LevelInfo) {
		t.Error("expected INFO for unknown level")
	}
	if l.Enabled(nil, slog.LevelDebug) {
		t.Error("expected DEBUG disabled for unknown level")
	}
}

func TestSetup_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	// Override global with a JSON handler writing to buf for output verification.
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	mu.Lock()
	global = slog.New(slog.NewJSONHandler(&buf, opts))
	mu.Unlock()

	L().Debug("test message", "key", "value")

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("expected log output, got none")
	}
	var record map[string]any
	if err := json.Unmarshal([]byte(line), &record); err != nil {
		t.Fatalf("log output is not valid JSON: %v\noutput: %s", err, line)
	}
	if record["msg"] != "test message" {
		t.Errorf("expected msg='test message', got: %v", record["msg"])
	}
}

func TestL_FallsBackToDefault(t *testing.T) {
	mu.Lock()
	global = nil
	mu.Unlock()

	l := L()
	if l == nil {
		t.Error("expected non-nil logger even before Setup")
	}
}
