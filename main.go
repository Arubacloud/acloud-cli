package main

import "acloud/cmd"

// Version is set at build time via ldflags
var Version string

func main() {
	cmd.SetVersion(Version)
	cmd.Execute()
}
