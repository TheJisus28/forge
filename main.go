package main

import (
	"fmt"
	"os"

	"github.com/TheJisus28/forge/internal/detect"
	"github.com/TheJisus28/forge/internal/initcmd"
	"github.com/TheJisus28/forge/internal/stackmd"
)

// version is overwritten at release time via -ldflags.
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}
	switch args[0] {
	case "help", "-h", "--help":
		printHelp()
		return nil
	case "version", "-v", "--version":
		fmt.Println(version)
		return nil
	case "init":
		return cmdInit(args[1:])
	case "detect":
		return cmdDetect(args[1:])
	default:
		return fmt.Errorf("unknown command %q (try forge help)", args[0])
	}
}

func cmdInit(args []string) error {
	opt := initcmd.Options{}
	dir := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--force":
			opt.Force = true
		case "--reset-memory":
			opt.ResetMemory = true
		case "--skip-copilot":
			opt.SkipCopilot = true
		case "--help", "-h":
			fmt.Print(`usage: forge init [dir] [--force] [--reset-memory] [--skip-copilot]

Copy the kit (forge/, AGENTS.md, CLAUDE.md, GEMINI.md, .cursor/, .claude/,
.github/) and write forge/memory/stack.md from repo manifests.

  --force          rewrite the kit if it already exists
  --reset-memory   also overwrite stack.md, decisions.md, constitution.md
  --skip-copilot   do not write .github/copilot-instructions.md
`)
			return nil
		default:
			if len(args[i]) > 0 && args[i][0] == '-' {
				return fmt.Errorf("unknown flag %s", args[i])
			}
			dir = args[i]
		}
	}
	if err := initcmd.Init(dir, opt); err != nil {
		return err
	}
	fmt.Printf("forge ready in %s\n", dir)
	return nil
}

func cmdDetect(args []string) error {
	dir := "."
	for _, a := range args {
		switch a {
		case "--help", "-h":
			fmt.Print("usage: forge detect [dir]\n")
			return nil
		default:
			if len(a) > 0 && a[0] == '-' {
				return fmt.Errorf("unknown flag %s", a)
			}
			dir = a
		}
	}
	st, err := detect.Dir(dir)
	if err != nil {
		return err
	}
	fmt.Print(stackmd.Render(st))
	return nil
}

func printHelp() {
	fmt.Print(`forge — initialize a spec-driven agent kit (Cursor, Claude Code, Codex, Gemini)

commands:
  init [dir]     copy the kit and detect the stack
  detect [dir]   print the stack.md that would be generated
  version
  help

init flags:
  --force --reset-memory --skip-copilot
`)
}
