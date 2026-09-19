package project_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/project"
	"github.com/TheJisus28/forge/internal/workflow"
)

// write builds a project on disk: the config plus one file per spec body.
func write(t *testing.T, config string, specs map[string]string) string {
	t.Helper()
	root := t.TempDir()
	base := filepath.Join(root, project.Dir)
	if err := os.MkdirAll(filepath.Join(base, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "README.md"), []byte("# .forge\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "project.md"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, body := range specs {
		if err := os.WriteFile(filepath.Join(base, "specs", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const config = `---
maintainers:
  - jesus
  - ana
test: "go test ./..."
---

# Project
`

const parent = `---
id: SPEC-001
title: Notifications
status: accepted
---

## Acceptance criteria

- AC1: real time delivery
- AC2: mark as read
- AC3: paginated history
`

const api = `---
id: SPEC-002
title: Notifications API
status: done
parent: SPEC-001
covers: [AC1, AC2]
approved_by: jesus
contract_hash: abc123
---

## Contract

GET /notifications
`

const ui = `---
id: SPEC-003
title: Notifications UI
status: accepted
parent: SPEC-001
covers: [AC3]
depends_on: [SPEC-002@contract]
needs:
  - a design system
---

## Contract
`

func load(t *testing.T) *project.Project {
	t.Helper()
	root := write(t, config, map[string]string{
		"SPEC-001-notifications.md": parent,
		"SPEC-002-api.md":           api,
		"SPEC-003-ui.md":            ui,
	})
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoad_ReadsSpecsAndConfig(t *testing.T) {
	p := load(t)
	if len(p.Specs) != 3 {
		t.Fatalf("specs = %d", len(p.Specs))
	}
	if !p.Configured() {
		t.Error("a project with maintainers and a test command is configured")
	}
	if !p.IsMaintainer("ANA") || p.IsMaintainer("pedro") {
		t.Error("maintainer lookup should be case insensitive and closed")
	}
	if p.AllowSelfApproval() {
		t.Error("with two maintainers, self approval is off unless asked for")
	}
	if !p.GuardEnabled() {
		t.Error("the guard is on unless the project turns it off")
	}
	s, ok := p.Spec("2")
	if !ok || s.ID != "SPEC-002" {
		t.Fatalf("ids should be forgiving: %v %v", s, ok)
	}
	if s.Status != workflow.Done || s.ApprovedBy != "jesus" {
		t.Errorf("frontmatter not read: %+v", s)
	}
}

func TestCoverage(t *testing.T) {
	p := load(t)
	root, _ := p.Spec("SPEC-001")
	rows := p.Coverage(root)
	if len(rows) != 3 {
		t.Fatalf("criteria = %d", len(rows))
	}
	if len(rows[0].By) != 1 || rows[0].By[0].ID != "SPEC-002" {
		t.Errorf("AC1 should be covered by SPEC-002: %v", rows[0].By)
	}
	if len(rows[2].By) != 1 || rows[2].By[0].ID != "SPEC-003" {
		t.Errorf("AC3 should be covered by SPEC-003: %v", rows[2].By)
	}
}

func TestCoverage_DetectsAnUncoveredPromise(t *testing.T) {
	root := write(t, config, map[string]string{
		"SPEC-001-notifications.md": parent,
		"SPEC-002-api.md":           api, // covers AC1 and AC2, nobody covers AC3
	})
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	spec, _ := p.Spec("SPEC-001")
	uncovered := 0
	for _, row := range p.Coverage(spec) {
		if len(row.By) == 0 {
			uncovered++
			if row.Criterion.ID != "AC3" {
				t.Errorf("wrong criterion reported: %s", row.Criterion.ID)
			}
		}
	}
	if uncovered != 1 {
		t.Fatalf("uncovered criteria = %d, want 1", uncovered)
	}
}

func TestDependencies(t *testing.T) {
	p := load(t)
	ui, _ := p.Spec("SPEC-003")
	if len(ui.Deps) != 1 || ui.Deps[0].Level != "contract" {
		t.Fatalf("deps = %+v", ui.Deps)
	}
	// SPEC-002 has an approved contract, so a @contract dependency is met
	// even though its code is not what SPEC-003 waits for.
	if ok, why := p.DepSatisfied(ui.Deps[0]); !ok {
		t.Errorf("contract level dependency should be satisfied: %s", why)
	}
	blockers := p.Blockers(ui)
	if len(blockers) != 1 || !strings.Contains(blockers[0], "design system") {
		t.Errorf("the unresolved need should be the only blocker: %v", blockers)
	}

	parentSpec, _ := p.Spec("SPEC-001")
	if got := p.Blockers(parentSpec); len(got) != 1 || !strings.Contains(got[0], "children") {
		t.Errorf("a parent is not started directly: %v", got)
	}
}

func TestDepSatisfied_WaitsForDoneByDefault(t *testing.T) {
	p := load(t)
	ok, _ := p.DepSatisfied(project.Dep{ID: "SPEC-003"})
	if ok {
		t.Error("a plain dependency needs the other spec to be done")
	}
	if ok, why := p.DepSatisfied(project.Dep{ID: "SPEC-404"}); ok ||
		!strings.Contains(why, "does not exist") {
		t.Errorf("unknown dependency: %v %q", ok, why)
	}
}

func TestContractDrift(t *testing.T) {
	p := load(t)
	s, _ := p.Spec("SPEC-002")
	if !s.ContractChanged() {
		t.Error("a stored hash that does not match the contract is drift")
	}
	s.ContractHash = project.HashContract(s.Contract())
	if s.ContractChanged() {
		t.Error("a matching hash is not drift")
	}
	// Whitespace is not a change of contract.
	s.ContractHash = project.HashContract("GET    /notifications\n\n")
	if s.ContractChanged() {
		t.Error("reformatting should not count as drift")
	}
}

func TestSaveRoundTrip(t *testing.T) {
	p := load(t)
	s, _ := p.Spec("SPEC-003")
	s.SetStatus(workflow.Specifying, "ana", "started")
	s.Conductor = "ana"
	s.Agreed["SPEC-002"] = "abc123"
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	again, err := project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, _ := again.Spec("SPEC-003")
	if reloaded.Status != workflow.Specifying || reloaded.Conductor != "ana" {
		t.Errorf("not persisted: %+v", reloaded)
	}
	if reloaded.Agreed["SPEC-002"] != "abc123" {
		t.Errorf("agreed contracts lost: %v", reloaded.Agreed)
	}
	if !strings.Contains(reloaded.Doc().Section("History"), "started") {
		t.Errorf("history not appended: %q", reloaded.Doc().Section("History"))
	}
	if len(reloaded.Deps) != 1 || reloaded.Deps[0].String() != "SPEC-002@contract" {
		t.Errorf("dependencies not preserved: %+v", reloaded.Deps)
	}
}

func TestNextNumAndNaming(t *testing.T) {
	p := load(t)
	if got := p.NextNum(); got != 4 {
		t.Errorf("next number = %d", got)
	}
	if got := project.FormatID(4); got != "SPEC-004" {
		t.Errorf("FormatID = %q", got)
	}
	cases := map[string]string{
		"Pago con tarjeta guardada": "pago-con-tarjeta-guardada",
		"  Notifications  API  ":    "notifications-api",
		"¿Qué pasa?":                "que-pasa",
		"":                          "untitled",
	}
	for in, want := range cases {
		if got := project.Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
	if got := project.FileName("SPEC-004", "Saved card"); got != "SPEC-004-saved-card.md" {
		t.Errorf("FileName = %q", got)
	}
}

func TestFind_NotInitialized(t *testing.T) {
	if _, err := project.Find(t.TempDir()); err != project.ErrNotInitialized {
		t.Fatalf("err = %v", err)
	}
}

func TestSpecIDFromBranch(t *testing.T) {
	cases := map[string]string{
		"spec/004-saved-card": "SPEC-004",
		"spec-4-thing":        "SPEC-004",
		"feature/whatever":    "",
		"main":                "",
	}
	for branch, want := range cases {
		if got := project.SpecIDFromBranch(branch); got != want {
			t.Errorf("SpecIDFromBranch(%q) = %q, want %q", branch, got, want)
		}
	}
}
