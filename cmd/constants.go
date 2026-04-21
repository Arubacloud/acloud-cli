package cmd

// Resource state values returned by the API.
const (
	StateInCreation = "InCreation"
	StateUsed       = "Used"
	StateNotUsed    = "NotUsed"
)

// DateLayout is the display format used for all creation/modification timestamps.
const DateLayout = "02-01-2006 15:04:05"

// File permission modes.
const (
	FilePermConfig = 0600 // owner read/write only — for credential files
	FilePermDirAll = 0755 // owner rwx, group/other rx — for config directories
)

// Canonical output formats accepted by the global --output flag.
// resolveOutputFormat normalises all accepted aliases to one of these values.
const (
	OutputFormatTable     = "table"
	OutputFormatTableJSON = "table-json"
	OutputFormatTableYAML = "table-yaml"
	OutputFormatJSON      = "json"
	OutputFormatYAML      = "yaml"
)

// Aliases accepted by the --output flag and normalised to canonical names.
const (
	OutputAliasStd          = "std"
	OutputAliasStandard     = "standard"
	OutputAliasStdJSON      = "std-json"
	OutputAliasStandardJSON = "standard-json"
	OutputAliasStdYAML      = "std-yaml"
	OutputAliasStandardYAML = "standard-yaml"
)
