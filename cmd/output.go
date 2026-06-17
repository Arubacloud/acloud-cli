package cmd

import (
	"github.com/Arubacloud/acloud-cli/internal/output"
)

// TableColumn is a type alias for output.TableColumn so existing cmd files
// can use the unqualified name without any import changes.
type TableColumn = output.TableColumn

// ListColumn is a type alias for output.ListColumn so cmd files can declare
// column definitions without importing internal/output directly.
type ListColumn[T any] = output.ListColumn[T]

// renderList builds and prints a typed list using the global --output flag.
// Items for which any filter returns false are excluded from the output.
func renderList[T any](obj any, cols []ListColumn[T], items []T, filters ...func(T) bool) {
	output.RenderList(resolveOutputFormat(), obj, cols, items, filters...)
}

// resolveOutputFormat reads the global --output flag and returns one of the
// five canonical format names defined in constants.go.
func resolveOutputFormat() string {
	raw, _ := rootCmd.PersistentFlags().GetString("output")
	return output.ResolveFormat(raw)
}

// PrintOutput is the single output primitive for all commands. obj is the full SDK
// response object used for json/yaml modes; headers+rows drive the table modes.
// Pass obj=nil when no rich object is available (e.g. delete confirmation rows).
func PrintOutput(obj any, headers []TableColumn, rows [][]string) {
	output.Print(resolveOutputFormat(), obj, headers, rows)
}

// renderGet renders a text/template string against data for the table-mode detail
// view of get commands. Call only after the JSON/YAML early-return path.
func renderGet(tmpl string, data any) error {
	return output.RenderGet(tmpl, data)
}
