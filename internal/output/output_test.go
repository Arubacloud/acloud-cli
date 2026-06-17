package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

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

// fakeRawMarshaler satisfies the rawMarshaler interface for testing.
type fakeRawMarshaler struct{ j, y []byte }

func (f fakeRawMarshaler) RawJSON() []byte { return f.j }
func (f fakeRawMarshaler) RawYAML() []byte { return f.y }

// ─── NormalizeHeaderKey ───────────────────────────────────────────────────────

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
		got := NormalizeHeaderKey(c.in)
		if got != c.want {
			t.Errorf("NormalizeHeaderKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ─── ResolveFormat ───────────────────────────────────────────────────────────

func TestResolveFormat(t *testing.T) {
	cases := []struct {
		raw, want string
	}{
		{"table", FormatTable},
		{"std", FormatTable},
		{"standard", FormatTable},
		{"", FormatTable},
		{"TABLE", FormatTable},
		{"table-json", FormatTableJSON},
		{"std-json", FormatTableJSON},
		{"standard-json", FormatTableJSON},
		{"TABLE-JSON", FormatTableJSON},
		{"table-yaml", FormatTableYAML},
		{"std-yaml", FormatTableYAML},
		{"standard-yaml", FormatTableYAML},
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"yaml", FormatYAML},
		{"YAML", FormatYAML},
		{"unknown-xyz", FormatTable}, // falls back to table
	}
	for _, c := range cases {
		got := ResolveFormat(c.raw)
		if got != c.want {
			t.Errorf("ResolveFormat(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

// ─── printJSON ────────────────────────────────────────────────────────────────

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

	t.Run("marshal error writes to stderr", func(t *testing.T) {
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
	})
}

// ─── printYAML ───────────────────────────────────────────────────────────────

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

	t.Run("marshal error writes to stderr", func(t *testing.T) {
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
	})
}

// ─── printTableJSON ──────────────────────────────────────────────────────────

func TestPrintTableJSON(t *testing.T) {
	t.Run("empty rows emits []", func(t *testing.T) {
		out := captureStdout(func() {
			printTableJSON([]TableColumn{{Header: "NAME", Width: 10}}, [][]string{})
		})
		if !strings.Contains(out, "[]") {
			t.Errorf("empty rows should print [], got: %s", out)
		}
	})

	t.Run("multiple rows", func(t *testing.T) {
		headers := []TableColumn{{Header: "NAME", Width: 10}, {Header: "ID", Width: 10}}
		rows := [][]string{{"alice", "id-1"}, {"bob", "id-2"}}
		out := captureStdout(func() { printTableJSON(headers, rows) })
		if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
			t.Errorf("output missing rows, got: %s", out)
		}
	})

	t.Run("short row handled gracefully", func(t *testing.T) {
		headers := []TableColumn{{Header: "NAME", Width: 10}, {Header: "ID", Width: 10}, {Header: "EXTRA", Width: 10}}
		rows := [][]string{{"alice", "id-1"}}
		out := captureStdout(func() { printTableJSON(headers, rows) })
		if !strings.Contains(out, "alice") {
			t.Errorf("short row output missing alice, got: %s", out)
		}
	})
}

// ─── rowsToRecords ───────────────────────────────────────────────────────────

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
	rows := [][]string{{"alice", "id-1"}}
	records := rowsToRecords(headers, rows)
	if len(records) != 1 {
		t.Errorf("expected 1 record, got %d", len(records))
	}
	if len(records[0].Content) != 4 { // 2 pairs × 2 nodes each
		t.Errorf("expected 4 content nodes (2 pairs), got %d", len(records[0].Content))
	}
}
