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
maintainers:
  - jesus
  - ana
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
		write(t, filepath.Join(base, "specs", name), body)
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
func findings(t *testing.T, p *project.Project, opt validate.Options) []string {
	t.Helper()
	var out []string
	for _, f := range validate.Run(p, opt) {
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
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: proposed\n---\n\n## Problem\n",
	})
	if got := findings(t, p, validate.Options{}); len(got) != 0 {
		t.Fatalf("unexpected errors: %v", got)
	}
}

func TestRun_BrokenReferences(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\n" +
			"parent: SPEC-404\ncovers: [AC1]\ndepends_on: [SPEC-999]\n---\n",
	})
	got := findings(t, p, validate.Options{})
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
	expectError(t, findings(t, p, validate.Options{}), "covers AC7")
}

// A parent whose children are closing while a promise has no owner is the
// failure the coverage matrix exists to catch.
func TestRun_UncoveredCriterionOnceAChildIsDone(t *testing.T) {
	parent := "---\nid: SPEC-001\ntitle: Parent\nstatus: accepted\n---\n\n" +
		"## Acceptance criteria\n\n- AC1: one\n- AC2: two\n"
	child := "---\nid: SPEC-002\ntitle: Child\nstatus: done\nparent: SPEC-001\n" +
		"covers: [AC1]\napproved_by: jesus\n---\n\n## Contract\n\nx\n"

	open := build(t, map[string]string{"SPEC-001-a.md": parent,
		"SPEC-002-b.md": strings.Replace(child, "status: done", "status: accepted", 1)})
	if got := findings(t, open, validate.Options{}); len(got) != 0 {
		t.Fatalf("while nothing is closed it is only a warning: %v", got)
	}

	closing := build(t, map[string]string{"SPEC-001-a.md": parent, "SPEC-002-b.md": child})
	expectError(t, findings(t, closing, validate.Options{}), "AC2 is not covered")
}

func TestRun_DependencyCycle(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\ndepends_on: [SPEC-002]\n---\n",
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: B\nstatus: accepted\ndepends_on: [SPEC-001]\n---\n",
	})
	expectError(t, findings(t, p, validate.Options{}), "dependency cycle")
}

func TestRun_ContractChangedAfterApproval(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: implementing\n" +
			"approved_by: jesus\ncontract_hash: staleHash\n---\n\n## Contract\n\nGET /things\n",
	})
	expectError(t, findings(t, p, validate.Options{}), "contract changed after jesus approved")
}

func TestRun_DependentBuildingOnAnOldContract(t *testing.T) {
	contract := "## Contract\n\nGET /things\n"
	hash := project.HashContract(contract)
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: API\nstatus: implementing\n" +
			"approved_by: jesus\ncontract_hash: " + hash + "\n---\n\n" + contract,
		"SPEC-002-b.md": "---\nid: SPEC-002\ntitle: UI\nstatus: implementing\n" +
			"depends_on: [SPEC-001@contract]\nagreed_contracts:\n  - \"SPEC-001:older\"\n---\n\n" +
			"## Contract\n\nx\n",
	})
	got := findings(t, p, validate.Options{})
	expectError(t, got, "SPEC-002 builds against SPEC-001's contract older")
}

func TestRun_ClaimedApprovalThatNeverHappened(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: planning\napproved_by: jesus\n" +
			"contract_hash: x\n---\n\n## Contract\n\nx\n",
	})
	got := findings(t, p, validate.Options{Approvers: []string{"ana"}})
	expectError(t, got, "did not approve this pull request")

	if got := findings(t, p, validate.Options{Approvers: []string{"jesus"}}); len(got) > 1 {
		t.Fatalf("a real approval should not be flagged: %v", got)
	}
}

func TestRun_ApprovalByAStranger(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: accepted\naccepted_by: pedro\n---\n",
	})
	expectError(t, findings(t, p, validate.Options{}), "not in maintainers")
}

func TestRun_SelfApproval(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: planning\nconductor: ana\n" +
			"approved_by: ana\ncontract_hash: x\n---\n\n## Contract\n\nx\n",
	})
	expectError(t, findings(t, p, validate.Options{}), "approved the contract they conducted")
}

func TestRun_MissingArtifacts(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: implementing\n---\n\n" +
			"## Contract\n\nGET /things\n",
	})
	expectError(t, findings(t, p, validate.Options{}), "implementing without")

	p = build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: awaiting-approval\n---\n\n" +
			"## Contract\n\n",
	})
	expectError(t, findings(t, p, validate.Options{}), "empty Contract section")
}

func TestRun_DoneWithoutArchiving(t *testing.T) {
	p := build(t, map[string]string{
		"SPEC-001-a.md": "---\nid: SPEC-001\ntitle: A\nstatus: done\napproved_by: jesus\n" +
			"---\n\n## Contract\n\nx\n",
	})
	write(t, filepath.Join(p.WipDirFor("SPEC-001"), "review.md"), "pass\n")
	p, err := project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	expectError(t, findings(t, p, validate.Options{}), "run forge archive")
}
