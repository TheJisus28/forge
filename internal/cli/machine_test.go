package cli_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/workflow"
	"github.com/TheJisus28/forge/kit"
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

// The roles tell the implementer and the orchestrator how work reaches the
// remote between phases (SPEC-020, AC5).
func TestRoles_DocumentTheCheckpoint(t *testing.T) {
	dir := t.TempDir()
	if implementer := mustRun(t, dir, "roles", "implementer"); !strings.Contains(implementer, "forge push") {
		t.Errorf("the implementer role should run forge push after a phase:\n%s", implementer)
	}
	if orchestrator := mustRun(t, dir, "roles", "orchestrator"); !strings.Contains(orchestrator, "push: on") {
		t.Errorf("the orchestrator role should name the push: on opt-in:\n%s", orchestrator)
	}
}

// A project that opts in refreshes its remote refs at session start, and the
// roles say it is a fetch only (SPEC-022, AC6).
func TestRoles_DocumentTheOptInFetch(t *testing.T) {
	dir := t.TempDir()
	for _, role := range []string{"orchestrator", "architect"} {
		out := mustRun(t, dir, "roles", role)
		if !strings.Contains(out, "fetch: on") {
			t.Errorf("the %s role should name the fetch: on opt-in:\n%s", role, out)
		}
		if !strings.Contains(out, "a fetch only") {
			t.Errorf("the %s role should say the refresh is a fetch only:\n%s", role, out)
		}
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

// A shipped template must not carry a literal `AC<digit>`: a fresh file
// copied from it would read as covered by `forge check`, the SPEC-019 trap
// one artifact over. The tasks and review templates name the criteria only by
// placeholder (`ACn`, `<criterion ids>`) and keep their example rows in a
// comment, out of the section the check reads (SPEC-021, decision 4).
func TestTemplates_CarryNoRealCriterionId(t *testing.T) {
	digit := regexp.MustCompile(`\bAC\d+\b`)

	tasks, err := kit.Template("tasks")
	if err != nil {
		t.Fatalf("Template(tasks): %v", err)
	}
	if m := digit.FindString(string(tasks)); m != "" {
		t.Errorf("the tasks template carries a real criterion id %q; use a placeholder", m)
	}
	if !strings.Contains(string(tasks), "Moves:") {
		t.Errorf("the tasks template should ask each phase to name the criteria it moves:\n%s", tasks)
	}

	review, err := kit.Template("review")
	if err != nil {
		t.Fatalf("Template(review): %v", err)
	}
	if m := digit.FindString(string(review)); m != "" {
		t.Errorf("the review template carries a real criterion id %q; use a placeholder", m)
	}
	if !strings.Contains(string(review), "## Acceptance criteria") || !strings.Contains(string(review), "ACn") {
		t.Errorf("the review template should keep its example under the Acceptance criteria heading as ACn:\n%s", review)
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
