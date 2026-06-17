package output

import (
	"strings"
	"testing"
)

func TestRenderGet_Basic(t *testing.T) {
	tmpl := "\nVPC Details:\n============\nID:   {{.ID}}\nName: {{.Name}}\n"
	data := struct{ ID, Name string }{"vpc-001", "my-vpc"}
	out := captureStdout(func() { _ = RenderGet(tmpl, data) })
	if !strings.Contains(out, "vpc-001") || !strings.Contains(out, "my-vpc") {
		t.Errorf("unexpected output: %q", out)
	}
	if !strings.Contains(out, "VPC Details:") {
		t.Errorf("missing title: %q", out)
	}
}

func TestRenderGet_EmptyField(t *testing.T) {
	tmpl := "Zone: {{.Zone}}\n"
	data := struct{ Zone string }{""}
	out := captureStdout(func() { _ = RenderGet(tmpl, data) })
	if !strings.Contains(out, "Zone: ") {
		t.Errorf("empty field should still print label: %q", out)
	}
}

func TestRenderGet_InvalidTemplate(t *testing.T) {
	err := RenderGet("{{.Unclosed", nil)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestRenderGet_SliceField(t *testing.T) {
	tmpl := "Tags: {{.Tags}}\n"
	data := struct{ Tags string }{"[e2e-test storage]"}
	out := captureStdout(func() { _ = RenderGet(tmpl, data) })
	if !strings.Contains(out, "[e2e-test storage]") {
		t.Errorf("unexpected output: %q", out)
	}
}
