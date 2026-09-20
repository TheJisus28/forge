package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/workflow"
)

// The machinery is served from the binary, so it works in a directory with
// no .forge/ at all.
func TestWorkflowCommand_PrintsTheWorkflow(t *testing.T) {
	dir := t.TempDir()
	out := mustRun(t, dir, "workflow")
	if !strings.Contains(out, "# Workflow") || !strings.Contains(out, "## States") {
		t.Fatalf("workflow output unexpected:\n%s", out)
	}
}

// The states table is generated from internal/workflow, in lifecycle order,
// and the same bytes are written on every run.
func TestWorkflowCommand_RendersTheStatesFromGo(t *testing.T) {
	dir := t.TempDir()
	first := mustRun(t, dir, "workflow")
	second := mustRun(t, dir, "workflow")
	if first != second {
		t.Error("two runs of forge workflow are not byte-identical")
	}

	last := -1
	for _, s := range workflow.All() {
		row := fmt.Sprintf("| `%s` | %s | %s |", s, workflow.Meaning(s), workflow.WaitingFor(s))
		at := strings.Index(first, row)
		if at < 0 {
			t.Errorf("workflow output missing the %s row %q:\n%s", s, row, first)
			continue
		}
		if at < last {
			t.Errorf("state %s is out of lifecycle order:\n%s", s, first)
		}
		last = at
	}

	if !strings.Contains(first, "| `contracting` |") {
		t.Errorf("workflow output does not render the renamed state:\n%s", first)
	}
	for _, retired := range []string{"specifying", "awaiting-approval"} {
		if strings.Contains(first, retired) {
			t.Errorf("workflow output still names the retired state %q:\n%s", retired, first)
		}
	}
}

func TestRolesCommand_ListsAndPrints(t *testing.T) {
	dir := t.TempDir()
	out := mustRun(t, dir, "roles")
	for _, role := range []string{"orchestrator", "architect", "implementer", "reviewer"} {
		if !strings.Contains(out, role) {
			t.Errorf("roles output missing %q:\n%s", role, out)
		}
	}
	one := mustRun(t, dir, "roles", "architect")
	if !strings.Contains(one, "# Architect") {
		t.Errorf("role output unexpected:\n%s", one)
	}
	if out, code := run(t, dir, "roles", "nope"); code == 0 {
		t.Errorf("an unknown role should fail:\n%s", out)
	}
}

func TestTemplateCommand_PrintsAndRejects(t *testing.T) {
	dir := t.TempDir()
	out := mustRun(t, dir, "template", "decision")
	if !strings.Contains(out, "## Decision") {
		t.Errorf("template output unexpected:\n%s", out)
	}
	if out, code := run(t, dir, "template", "nope"); code == 0 {
		t.Errorf("an unknown template should fail:\n%s", out)
	}
	if out, code := run(t, dir, "template"); code == 0 {
		t.Errorf("a template name is required:\n%s", out)
	}
}

// The existing-state survey lives in spec.md, so the spec template carries
// the section and the plan template no longer does (SPEC-015, decision 5).
func TestTemplate_SurveySection(t *testing.T) {
	dir := t.TempDir()
	spec := mustRun(t, dir, "template", "spec")
	if !strings.Contains(spec, "## Existing state") {
		t.Errorf("forge template spec should carry the Existing state section:\n%s", spec)
	}
	plan := mustRun(t, dir, "template", "plan")
	if strings.Contains(plan, "## Existing state") {
		t.Errorf("forge template plan should not carry the Existing state section:\n%s", plan)
	}
}

// A template planted in the repository is ignored: the templates are a
// standard that ships in the binary.
func TestNew_IgnoresAPlantedTemplate(t *testing.T) {
	dir := newRepo(t)
	write(t, filepath.Join(dir, ".forge", "kit", "templates", "spec.md"), "JUNK\n")
	mustRun(t, dir, "new", "Probe", "--capability", "workflow")

	specs, _ := filepath.Glob(filepath.Join(dir, ".forge", "specs", "*", "spec.md"))
	if len(specs) != 1 {
		t.Fatalf("expected one spec, got %d", len(specs))
	}
	body := read(t, specs[0])
	if strings.Contains(body, "JUNK") {
		t.Errorf("forge new used a planted template:\n%s", body)
	}
	if !strings.Contains(body, "## Contract") {
		t.Errorf("forge new did not use the embedded template:\n%s", body)
	}
}

// The generated agents carry the role text and never point at a file under
// .forge/, which no longer holds the machinery.
func TestInit_InlinesTheRolesIntoHostAdapters(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init")

	for _, rel := range []string{
		".claude/agents/forge-architect.md",
		".claude/agents/forge-implementer.md",
		".opencode/agents/forge-reviewer.md",
	} {
		body := read(t, filepath.Join(dir, rel))
		if strings.Contains(body, ".forge/kit") {
			t.Errorf("%s still points at .forge/kit:\n%s", rel, body)
		}
		if strings.Contains(body, "{{forge-role:") {
			t.Errorf("%s has an unfilled role marker:\n%s", rel, body)
		}
	}

	body := read(t, filepath.Join(dir, ".claude", "agents", "forge-architect.md"))
	if !strings.Contains(body, "# Architect") {
		t.Errorf("adapter does not carry the role text:\n%s", body)
	}
}

// forge update removes the .forge/kit/ directory older versions planted.
func TestUpdate_RemovesStaleKit(t *testing.T) {
	dir := newRepo(t)
	write(t, filepath.Join(dir, ".forge", "kit", "agents", "architect.md"), "old\n")
	mustRun(t, dir, "update")
	if _, err := os.Stat(filepath.Join(dir, ".forge", "kit")); !os.IsNotExist(err) {
		t.Errorf(".forge/kit should be removed by update, got %v", err)
	}
}
