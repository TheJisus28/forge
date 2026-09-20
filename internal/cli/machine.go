package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/TheJisus28/forge/internal/workflow"
	"github.com/TheJisus28/forge/kit"
)

// workflowFile names the embedded copy for errors, and workflowMarker is the
// placeholder it carries where the states table is rendered from Go.
const (
	workflowFile   = "kit/machine/WORKFLOW.md"
	workflowMarker = "<!-- forge:states -->"
)

// cmdWorkflow prints the process description from the binary. It never reads
// the project, so any agent, wrapped or not, can pull the same copy. The
// states table is rendered from internal/workflow, which is the only
// declaration of the states.
func cmdWorkflow(args []string, out io.Writer) error {
	fs := newFlagSet("workflow", "usage: forge workflow", out)
	if _, err := parseArgs(fs, args); err != nil {
		return err
	}
	data, err := kit.Workflow()
	if err != nil {
		return err
	}
	rendered, err := renderWorkflow(data)
	if err != nil {
		return err
	}
	_, err = out.Write(rendered)
	return err
}

// renderWorkflow replaces the states marker in the embedded workflow with a
// table built from workflow.All(), workflow.Meaning() and
// workflow.WaitingFor(). The Markdown stays prose; the table has one source.
// A workflow that lost the marker is an error naming the file, so the states
// cannot silently disappear from `forge workflow`.
func renderWorkflow(raw []byte) ([]byte, error) {
	if !bytes.Contains(raw, []byte(workflowMarker)) {
		return nil, fmt.Errorf("%s is missing the %s marker", workflowFile, workflowMarker)
	}
	rows := []string{"| State | Meaning | Who acts next |", "|---|---|---|"}
	for _, s := range workflow.All() {
		rows = append(rows, fmt.Sprintf("| `%s` | %s | %s |", s, workflow.Meaning(s), workflow.WaitingFor(s)))
	}
	return bytes.ReplaceAll(raw, []byte(workflowMarker), []byte(strings.Join(rows, "\n"))), nil
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
