// Command forge plants and drives a spec-driven workflow for coding agents.
package main

import (
	"os"

	"github.com/TheJisus28/forge/internal/cli"
)

// version is overwritten at release time via -ldflags.
var version = "dev"

func main() {
	cli.Version = version
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
