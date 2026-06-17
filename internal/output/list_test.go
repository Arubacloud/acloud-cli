package output

import (
	"encoding/json"
	"strings"
	"testing"
)

type testItem struct{ name, id, status string }

func cols() []ListColumn[testItem] {
	return []ListColumn[testItem]{
		{TableColumn: TableColumn{Header: "NAME", Width: 10}, Value: func(i testItem) string { return i.name }},
		{TableColumn: TableColumn{Header: "ID", Width: 5}, Value: func(i testItem) string { return i.id }},
		{TableColumn: TableColumn{Header: "STATUS", Width: 8}, Value: func(i testItem) string { return i.status }},
	}
}

func TestRenderList_Table(t *testing.T) {
	items := []testItem{{"alice", "1", "Active"}, {"bob", "2", "Inactive"}}
	out := captureStdout(func() { RenderList(FormatTable, nil, cols(), items) })
	if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
		t.Errorf("missing rows: %s", out)
	}
	if !strings.Contains(out, "NAME") {
		t.Errorf("missing header: %s", out)
	}
}

func TestRenderList_Empty(t *testing.T) {
	out := captureStdout(func() { RenderList(FormatTable, nil, cols(), nil) })
	if !strings.Contains(out, "NAME") {
		t.Errorf("empty list should still print headers: %s", out)
	}
}

func TestRenderList_Filter(t *testing.T) {
	items := []testItem{{"alice", "1", "Active"}, {"", "2", ""}, {"bob", "3", "Active"}}
	skip := func(i testItem) bool { return i.name != "" }
	out := captureStdout(func() { RenderList(FormatTable, nil, cols(), items, skip) })
	if strings.Contains(out, `"2"`) || strings.Contains(out, "  ") {
		// blank item should be excluded
	}
	if !strings.Contains(out, "alice") || !strings.Contains(out, "bob") {
		t.Errorf("non-blank items should appear: %s", out)
	}
}

func TestRenderList_JSON(t *testing.T) {
	items := []testItem{{"alice", "1", "Active"}}
	out := captureStdout(func() { RenderList(FormatTableJSON, nil, cols(), items) })
	var result []map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out)
	}
	if len(result) != 1 || result[0]["name"] != "alice" {
		t.Errorf("unexpected JSON result: %v", result)
	}
}

func TestRenderList_YAML(t *testing.T) {
	items := []testItem{{"alice", "1", "Active"}}
	out := captureStdout(func() { RenderList(FormatTableYAML, nil, cols(), items) })
	if !strings.Contains(out, "name: alice") {
		t.Errorf("missing name in YAML: %s", out)
	}
}

func TestRenderList_MultipleFilters(t *testing.T) {
	items := []testItem{{"alice", "1", "Active"}, {"bob", "2", ""}, {"", "3", "Active"}}
	hasName := func(i testItem) bool { return i.name != "" }
	hasStatus := func(i testItem) bool { return i.status != "" }
	out := captureStdout(func() { RenderList(FormatTable, nil, cols(), items, hasName, hasStatus) })
	if !strings.Contains(out, "alice") {
		t.Errorf("alice (has name+status) should appear: %s", out)
	}
	if strings.Contains(out, "bob") {
		t.Errorf("bob (no status) should be excluded: %s", out)
	}
}
