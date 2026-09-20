package validate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/validate"
)

const config = `---
test: "go test ./..."
---

# Project
`

// build writes a project and returns it loaded.
func build(t *testing.T, specs map[string]string) *project.Project {
	t.Helper()
	root := t.TempDir()
	base := filepath.Join(root, project.Dir)
	if err := os.MkdirAll(filepath.Join(base, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(base, "README.md"), "# .forge\n")
	write(t, filepath.Join(base, "project.md"), config)
	for name, body := range specs {
		write(t, filepath.Join(base, "specs", strings.TrimSuffix(name, ".md"), "spec.md"), body)
	}
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// findings returns the error messages, so tests read like expectations.
func findings(t *testing.T, p *project.Project) []string {
	t.Helper()
	var out []string
	for _, f := range validate.Run(p) {
		if f.Severity == validate.Error {
			out = append(out, f.String())
		}
	}
	return out
}

func expectError(t *testing.T, got []string, want string) {
	t.Helper()
	for _, g := range got {
		if strings.Contains(g, want) {
			return
		}
	}
	t.Fatalf("expected an error containing %q, got %v", want, got)
}

func TestRun_CleanProject(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\ncapability: a\n---\n\n## Problem\n",
	})
	if got := findings(t, p); len(got) != 0 {
		t.Fatalf("unexpected errors: %v", got)
	}
}

// The specs delivered before `capability` existed must stay visible without
// failing the build: it is a warning, not an error.
func TestRun_WarnsMissingCapability(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\n---\n\n## Problem\n",
	})
	if !warns(t, p, "has no capability") {
		t.Fatal("a spec with no capability should warn")
	}
	if got := findings(t, p); len(got) != 0 {
		t.Fatalf("a missing capability must not fail the build: %v", got)
	}
}

// A present key that is not a lowercase slug is an error, and an empty value
// counts: the key exists, so it is not the missing-field warning.
func TestRun_RejectsInvalidCapability(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\ncapability: Guard\n---\n",
	})
	expectError(t, findings(t, p), `capability "Guard" is not a lowercase slug`)

	p = build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\ncapability: \"\"\n---\n",
	})
	expectError(t, findings(t, p), `capability "" is not a lowercase slug`)
}

func TestRun_BrokenReferences(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\n" +
			"parent: SPEC-404\ncovers: [AC1]\ndepends_on: [SPEC-999]\n---\n",
	})
	got := findings(t, p)
	expectError(t, got, "parent SPEC-404 does not exist")
	expectError(t, got, "depends on SPEC-999, which does not exist")
}

func TestRun_CoversSomethingTheParentNeverPromised(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: Parent\nstatus: accepted\n---\n\n" +
			"## Acceptance criteria\n\n- AC1: one\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: Child\nstatus: accepted\n" +
			"parent: SPEC-001\ncovers: [AC7]\n---\n",
	})
	expectError(t, findings(t, p), "covers AC7")
}

// A parent whose children are closing while a promise has no owner is the
// failure the coverage matrix exists to catch. Nobody has to authorize
// this check; it is just arithmetic on what was promised versus delivered.
func TestRun_UncoveredCriterionOnceAChildIsDone(t *testing.T) {
	parent := "---\nid: SPEC-001\ntitle: Parent\nstatus: accepted\n---\n\n" +
		"## Acceptance criteria\n\n- AC1: one\n- AC2: two\n"
	child := "---\nid: SPEC-002\ntitle: Child\nstatus: done\nparent: SPEC-001\n" +
		"covers: [AC1]\napproved_by: ana\n---\n\n## Contract\n\nx\n"

	open := build(t, map[string]string{"SPEC-001-a.md": parent,
		"SPEC-002-b.md": strings.Replace(child, "status: done", "status: accepted", 1)})
	if got := findings(t, open); len(got) != 0 {
		t.Fatalf("while nothing is closed it is only a warning: %v", got)
	}

	closing := build(t, map[string]string{"SPEC-001-a.md": parent, "SPEC-002-b.md": child})
	expectError(t, findings(t, closing), "AC2 is not covered")
}

func TestRun_DependencyCycle(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\ndepends_on: [SPEC-002]\n---\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: B\nstatus: accepted\ndepends_on: [SPEC-001]\n---\n",
	})
	expectError(t, findings(t, p), "dependency cycle")
}

func TestRun_SupersedesAMissingSpec(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\ncapability: a\n" +
			"supersedes: [SPEC-404]\n---\n",
	})
	expectError(t, findings(t, p), "supersedes SPEC-404, which does not exist")
}

func TestRun_SupersedesASpecThatIsNotDone(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\ncapability: a\n" +
			"supersedes: [SPEC-002]\n---\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: B\nstatus: accepted\ncapability: b\n---\n",
	})
	expectError(t, findings(t, p), "supersedes SPEC-002, which is not done")
}

func TestRun_SupersedesCycle(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\ncapability: a\n" +
			"supersedes: [SPEC-002]\n---\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: B\nstatus: accepted\ncapability: b\n" +
			"supersedes: [SPEC-001]\n---\n",
	})
	expectError(t, findings(t, p), "supersedes cycle")
}

// Two specs still in flight cannot both claim to replace the same delivered
// contract: one of them is wrong, and history would show two successors.
func TestRun_TwoLiveSpecsSupersedeTheSameTarget(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\ncapability: a\n" +
			"supersedes: [SPEC-003]\n---\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: B\nstatus: accepted\ncapability: b\n" +
			"supersedes: [SPEC-003]\n---\n",
		"SPEC-003-c.md": "---\nid: SPEC-003\ntitle: C\nstatus: done\ncapability: c\n" +
			"approved_by: ana\n---\n\n## Contract\n\nx\n",
	})
	s, _ := p.Spec("SPEC-003")
	write(t, filepath.Join(s.Dir(), "review.md"), "Verdict: pass\n")

	got := findings(t, p)
	expectError(t, got, "supersedes SPEC-003, which SPEC-002 also supersedes")
	expectError(t, got, "supersedes SPEC-003, which SPEC-001 also supersedes")
}

// A supersede that points at a delivered contract and loops over nothing is
// what the field is for; it must not produce a finding.
func TestRun_ValidSupersedeIsQuiet(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-old.md": "---\nid: SPEC-001\ntitle: Old\nstatus: done\ncapability: a\n" +
			"approved_by: ana\n---\n\n## Contract\n\nx\n",
		"SPEC-002-new.md": "---\nid: SPEC-002\ntitle: New\nstatus: proposed\ncapability: a\n" +
			"supersedes: [SPEC-001]\n---\n",
	})
	s, _ := p.Spec("SPEC-001")
	write(t, filepath.Join(s.Dir(), "review.md"), "Verdict: pass\n")

	if got := findings(t, p); len(got) != 0 {
		t.Fatalf("a valid supersede should be clean: %v", got)
	}
}

func TestRun_ContractChangedAfterApproval(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: implementing\n" +
			"approved_by: ana\ncontract_hash: staleHash\n---\n\n## Contract\n\nGET /things\n",
	})
	expectError(t, findings(t, p), "contract changed after ana approved")
}

func TestRun_DependentBuildingOnAnOldContract(t *testing.T) {
	contract := "## Contract\n\nGET /things\n"
	hash := project.HashContract(contract)
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: API\nstatus: implementing\n" +
			"approved_by: ana\ncontract_hash: " + hash + "\n---\n\n" + contract,
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: UI\nstatus: implementing\n" +
			"depends_on: [SPEC-001@contract]\nagreed_contracts:\n  - \"SPEC-001:older\"\n---\n\n" +
			"## Contract\n\nx\n",
	})
	got := findings(t, p)
	expectError(t, got, "SPEC-002 builds against SPEC-001's contract older")
}

func TestRun_MissingArtifacts(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: implementing\n---\n\n" +
			"## Contract\n\nGET /things\n",
	})
	expectError(t, findings(t, p), "implementing without")

	p = build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: awaiting-approval\n---\n\n" +
			"## Contract\n\n",
	})
	expectError(t, findings(t, p), "empty Contract section")
}

// A plan that does not survey what already exists is only a warning: it
// guides the agent to record reuse without blocking the build.
func TestRun_PlanWithoutExistingStateWarns(t *testing.T) {
	spec := "---\nid: SPEC-001\ntitle: A\nstatus: implementing\n---\n\n## Contract\n\nx\n"

	p := build(t, map[string]string{"SPEC-001-a.md": spec})
	s, _ := p.Spec("SPEC-001")
	plan := filepath.Join(s.Dir(), "plan.md")
	write(t, plan, "# Plan\n\n## Phase 1\n\n- Scope: x.\n")
	write(t, filepath.Join(s.Dir(), "tasks.md"), "# Tasks\n\n- [ ] Phase 1\n")
	p, err := project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if !warns(t, p, "Existing state") {
		t.Fatal("a plan with no Existing state section should warn")
	}

	write(t, plan, "# Plan\n\n## Existing state\n\n- reuses `calc.go`.\n\n## Phase 1\n")
	p, err = project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if warns(t, p, "Existing state") {
		t.Fatal("a plan that surveys the existing state should not warn")
	}
}

// An open question is a warning: it should be answered, not forgotten.
func TestRun_WarnsOpenQuestions(t *testing.T) {
	spec := "---\nid: SPEC-001\ntitle: A\nstatus: accepted\n---\n\n" +
		"## Open questions\n\n- OQ1: which store?\n"
	p := build(t, map[string]string{"SPEC-001-a.md": spec})
	if !warns(t, p, "open questions") {
		t.Fatal("an open question should warn")
	}
}

// warns reports whether any finding is a warning whose message contains want.
func warns(t *testing.T, p *project.Project, want string) bool {
	t.Helper()
	for _, f := range validate.Run(p) {
		if f.Severity == validate.Warning && strings.Contains(f.Message, want) {
			return true
		}
	}
	return false
}

// The spec folder is the durable record: a done spec keeps its review.
func TestRun_DoneRequiresReview(t *testing.T) {
	spec := "---\nid: SPEC-001\ntitle: A\nstatus: done\napproved_by: ana\n" +
		"---\n\n## Contract\n\nx\n"

	p := build(t, map[string]string{"SPEC-001-a.md": spec})
	s, _ := p.Spec("SPEC-001")
	expectError(t, findings(t, p), "is done without")

	write(t, filepath.Join(s.Dir(), "review.md"), "Verdict: pass\n")
	p, err := project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	if got := findings(t, p); len(got) != 0 {
		t.Fatalf("a done spec with a review should be clean: %v", got)
	}
}

// Anyone can accept or approve; nothing in Forge checks who they are. The
// attribution is informational, the way a git commit's author is.
func TestRun_AnyHandleCanAcceptOrApprove(t *testing.T) {
	hash := project.HashContract("x") // the section body, not the heading
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\n" +
			"accepted_by: pedro\n---\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: B\nstatus: planning\nconductor: ana\n" +
			"approved_by: ana\ncontract_hash: " + hash + "\n---\n\n## Contract\n\nx\n",
	})
	if got := findings(t, p); len(got) != 0 {
		t.Fatalf("no authorization checks should fire: %v", got)
	}
}
