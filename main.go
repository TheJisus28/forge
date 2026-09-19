// Command forge plants and drives a spec-driven workflow for coding agents.
package main

import (
	"os"
	"runtime/debug"

	"github.com/TheJisus28/forge/internal/cli"
)

// version is baked into the GoReleaser binaries via -ldflags. It stays
// "dev" for `go install pkg@version`, which compiles locally and never
// runs our ldflags; resolveVersion falls back to the module version Go
// itself records in the build info for that case.
var version = "dev"

func main() {
	cli.Version = resolveVersion()
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok &&
		info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
