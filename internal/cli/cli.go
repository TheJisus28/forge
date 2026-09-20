// Package cli parses commands and prints results. Every command is a thin
// wrapper: the rules live in internal/workflow, internal/project and
// internal/validate.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Version is set at release time from main.
var Version = "dev"

const usage = `forge — spec-driven agentic development, in files you own

  forge init [dir]              plant the kit in a repository
  forge update                  refresh the kit, never touching your content
  forge upgrade [version]       upgrade the forge binary itself

  forge new "<title>" --capability <name>  open a spec (proposed)
  forge accept <id>             into the queue (defaults --by to git user.name)
  forge start <id>              begin the work: checks dependencies
  forge approve <id>            the contract is right; code can start
  forge advance <id> --to <state>
  forge archive <id>            distil the spec, last commit of the pull request
  forge submit [id]             push the branch and open the pull request

  forge status [id]             what is open, who is waiting, what blocks
  forge brief                   the short state an agent reads at session start
  forge validate                exit 1 when the project is inconsistent
  forge sync [id]               read the pull request state through gh
  forge renumber <id>           resolve a duplicate id
  forge guard                   no product code without a spec (hook or --file)
  forge workflow                print the workflow the process follows
  forge roles [name]            print the role names, or one role
  forge template <name>         print a file template (spec, plan, decision...)
  forge version

Run any command with --help for its flags.
`

// Main runs a command and returns the process exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	cleanupStaleBinary()
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return 0
	}
	cmd, rest := args[0], args[1:]
	var err error
	switch cmd {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	case "version", "-v", "--version":
		fmt.Fprintln(stdout, Version)
		return 0
	case "init":
		err = cmdInit(rest, stdout)
	case "update":
		err = cmdUpdate(rest, stdout)
	case "upgrade":
		err = cmdUpgrade(rest, stdout, stderr)
	case "new":
		err = cmdNew(rest, stdout)
	case "accept":
		err = cmdAccept(rest, stdout)
	case "start":
		err = cmdStart(rest, stdout)
	case "approve":
		err = cmdApprove(rest, stdout)
	case "advance":
		err = cmdAdvance(rest, stdout)
	case "archive":
		err = cmdArchive(rest, stdout)
	case "status":
		err = cmdStatus(rest, stdout)
	case "brief":
		err = cmdBrief(rest, stdout)
	case "validate":
		return cmdValidate(rest, stdout, stderr)
	case "sync":
		err = cmdSync(rest, stdout)
	case "submit":
		err = cmdSubmit(rest, stdout)
	case "renumber":
		err = cmdRenumber(rest, stdout)
	case "guard":
		err = cmdGuard(rest, stdout)
	case "workflow":
		err = cmdWorkflow(rest, stdout)
	case "roles":
		err = cmdRoles(rest, stdout)
	case "template":
		err = cmdTemplate(rest, stdout)
	default:
		err = fmt.Errorf("unknown command %q; run forge help", cmd)
	}
	if err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		fmt.Fprintf(stderr, "forge: %v\n", err)
		return 1
	}
	return 0
}

// newFlagSet builds a flag set that prints its help to stdout.
func newFlagSet(name, help string, out io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(out)
	fs.Usage = func() {
		fmt.Fprintln(out, help)
		fs.PrintDefaults()
	}
	return fs
}

// parseArgs parses flags that appear before, after or between positional
// arguments, because `forge new "title" --parent X` is how people type.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var positional []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		positional = append(positional, rest[0])
		args = rest[1:]
	}
	return positional, nil
}

func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
