package output

import (
	"os"
	"text/template"
)

// RenderGet executes a text/template string against data and writes the result
// to stdout. Intended for the table-mode detail view of get commands; JSON/YAML
// modes use PrintOutput before this is reached.
func RenderGet(tmpl string, data any) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, data)
}
