package cmd

import (
	"github.com/Arubacloud/acloud-cli/internal/output"
)

// TableColumn is a type alias for output.TableColumn so existing cmd files
// can use the unqualified name without any import changes.
type TableColumn = output.TableColumn

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
