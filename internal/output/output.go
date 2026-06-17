package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// TableColumn defines a column in a fixed-width text table.
type TableColumn struct {
	Header string
	Width  int
}

// rawMarshaler is satisfied by every SDK wrapper and List[T] in sdk-go v1.0.0.
// Print and its helpers use it to delegate serialisation to the SDK so the
// output shape matches the wire representation exactly.
type rawMarshaler interface {
	RawJSON() []byte
	RawYAML() []byte
}

// Canonical output format names accepted by ResolveFormat.
const (
	FormatTable     = "table"
	FormatTableJSON = "table-json"
	FormatTableYAML = "table-yaml"
	FormatJSON      = "json"
	FormatYAML      = "yaml"
)

// ResolveFormat normalises raw (the value of the --output flag) to one of the
// five canonical format names above. Unknown values fall back to FormatTable.
func ResolveFormat(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case FormatTable, "std", "standard", "":
		return FormatTable
	case FormatTableJSON, "std-json", "standard-json":
		return FormatTableJSON
	case FormatTableYAML, "std-yaml", "standard-yaml":
		return FormatTableYAML
	case FormatJSON:
		return FormatJSON
	case FormatYAML:
		return FormatYAML
	default:
		fmt.Fprintf(os.Stderr, "warning: unknown --output value %q, falling back to %s\n", raw, FormatTable)
		return FormatTable
	}
}

// Print dispatches to the appropriate rendering function based on format.
// format must be one of the canonical FormatXxx constants (call ResolveFormat first).
// obj is the full SDK response object used for json/yaml modes; headers+rows drive
// the table modes. Pass obj=nil when no rich object is available.
func Print(format string, obj any, headers []TableColumn, rows [][]string) {
	switch format {
	case FormatJSON:
		printJSON(obj)
	case FormatYAML:
		printYAML(obj)
	case FormatTableJSON:
		printTableJSON(headers, rows)
	case FormatTableYAML:
		printTableYAML(headers, rows)
	default:
		printTable(headers, rows)
	}
}

// printJSON serialises obj as JSON to stdout.
// SDK wrappers satisfy rawMarshaler and are written via RawJSON().
// Anonymous structs fall through to json.MarshalIndent.
// Emits {} when obj is nil so machine consumers always receive valid JSON.
func printJSON(obj any) {
	if obj == nil {
		fmt.Println("{}")
		return
	}
	if m, ok := obj.(rawMarshaler); ok {
		b := m.RawJSON()
		if len(b) == 0 {
			fmt.Println("{}")
			return
		}
		fmt.Println(string(b))
		return
	}
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling to JSON: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

// printYAML serialises obj as YAML to stdout.
// SDK wrappers satisfy rawMarshaler and are written via RawYAML().
// Anonymous structs fall through to json.Marshal → YAML conversion.
// Emits {} when obj is nil.
func printYAML(obj any) {
	if obj == nil {
		fmt.Println("{}")
		return
	}
	if m, ok := obj.(rawMarshaler); ok {
		b := m.RawYAML()
		if len(b) == 0 {
			fmt.Println("{}")
			return
		}
		os.Stdout.Write(b)
		return
	}
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling to JSON for YAML conversion: %v\n", err)
		return
	}
	if len(jsonBytes) == 0 {
		fmt.Println("{}")
		return
	}
	var intermediate any
	if err := json.Unmarshal(jsonBytes, &intermediate); err != nil {
		fmt.Fprintf(os.Stderr, "error converting JSON to YAML intermediate: %v\n", err)
		return
	}
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	_ = enc.Encode(intermediate)
	_ = enc.Close()
}

// printTableJSON emits the table rows as an ordered JSON array of flat snake_case
// objects. Column order is preserved by hand-building the JSON rather than using
// json.Marshal on a map (which loses ordering).
func printTableJSON(headers []TableColumn, rows [][]string) {
	keys := make([]string, len(headers))
	for i, col := range headers {
		keys[i] = NormalizeHeaderKey(col.Header)
	}
	fmt.Print("[")
	for ri, row := range rows {
		if ri > 0 {
			fmt.Print(",")
		}
		fmt.Print("\n  {")
		for ki, key := range keys {
			if ki >= len(row) {
				break
			}
			if ki > 0 {
				fmt.Print(",")
			}
			keyJSON, _ := json.Marshal(key)
			valJSON, _ := json.Marshal(row[ki])
			fmt.Printf("\n    %s: %s", keyJSON, valJSON)
		}
		fmt.Print("\n  }")
	}
	if len(rows) > 0 {
		fmt.Print("\n")
	}
	fmt.Println("]")
}

// printTableYAML emits the table rows as a YAML sequence of flat snake_case mappings.
// yaml.Node is used to preserve column order.
func printTableYAML(headers []TableColumn, rows [][]string) {
	records := rowsToRecords(headers, rows)
	var doc yaml.Node
	doc.Kind = yaml.SequenceNode
	for i := range records {
		doc.Content = append(doc.Content, &records[i])
	}
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	_ = enc.Encode(&doc)
	_ = enc.Close()
}

// printTable emits a fixed-width text table to stdout.
// Values longer than the column Width are truncated with "...".
func printTable(headers []TableColumn, rows [][]string) {
	formatStr := ""
	headerValues := make([]any, len(headers))
	for i, col := range headers {
		formatStr += fmt.Sprintf("%%-%ds ", col.Width)
		headerValues[i] = col.Header
	}
	formatStr += "\n"
	fmt.Printf(formatStr, headerValues...)

	for _, row := range rows {
		rowValues := make([]any, len(row))
		for i, val := range row {
			if len(headers) > i && len(val) > headers[i].Width {
				val = val[:headers[i].Width-3] + "..."
			}
			rowValues[i] = val
		}
		fmt.Printf(formatStr, rowValues...)
	}
}

// NormalizeHeaderKey converts a display-oriented table header (e.g. "RAM(GB)",
// "CREATION DATE", "PUBLIC_KEY") into a snake_case key for JSON/YAML serialisation.
// Non-alphanumeric runs collapse into a single underscore.
func NormalizeHeaderKey(header string) string {
	var sb strings.Builder
	prevWasAlnum := false
	for _, r := range header {
		switch {
		case r >= 'A' && r <= 'Z':
			sb.WriteRune(r + ('a' - 'A'))
			prevWasAlnum = true
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			sb.WriteRune(r)
			prevWasAlnum = true
		default:
			if prevWasAlnum {
				sb.WriteRune('_')
				prevWasAlnum = false
			}
		}
	}
	return strings.Trim(sb.String(), "_")
}

// rowsToRecords converts tabular headers+rows into an ordered list of yaml.Node
// mappings suitable for YAML serialisation with preserved column order.
func rowsToRecords(headers []TableColumn, rows [][]string) []yaml.Node {
	keys := make([]string, len(headers))
	for i, col := range headers {
		keys[i] = NormalizeHeaderKey(col.Header)
	}

	records := make([]yaml.Node, 0, len(rows))
	for _, row := range rows {
		var mapping yaml.Node
		mapping.Kind = yaml.MappingNode
		for i, key := range keys {
			if i >= len(row) {
				break
			}
			var k, v yaml.Node
			k.Kind = yaml.ScalarNode
			k.Value = key
			v.Kind = yaml.ScalarNode
			v.Value = row[i]
			mapping.Content = append(mapping.Content, &k, &v)
		}
		records = append(records, mapping)
	}
	return records
}
