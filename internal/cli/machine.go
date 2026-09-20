package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/TheJisus28/forge/kit"
)

// cmdWorkflow prints the process description from the binary. It never reads
// the project, so any agent, wrapped or not, can pull the same copy.
func cmdWorkflow(args []string, out io.Writer) error {
	fs := newFlagSet("workflow", "usage: forge workflow", out)
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}
	data, err := kit.Workflow()
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

// cmdRoles prints the role names, or one role's instructions by name.
func cmdRoles(args []string, out io.Writer) error {
	fs := newFlagSet("roles", "usage: forge roles [name]", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		fmt.Fprintln(out, strings.Join(kit.Roles(), "\n"))
		return nil
	}
	data, err := kit.Role(rest[0])
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}

// cmdTemplate prints one file template, so a decision or a convention is
// created from the binary's copy instead of a planted file.
func cmdTemplate(args []string, out io.Writer) error {
	fs := newFlagSet("template", "usage: forge template <name>", out)
	rest, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return fmt.Errorf("which template? run: forge template <spec|plan|tasks|review|decision|convention>")
	}
	data, err := kit.Template(rest[0])
	if err != nil {
		return err
	}
	_, err = out.Write(data)
	return err
}
