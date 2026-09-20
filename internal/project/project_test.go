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
		dir := filepath.Join(base, "specs", strings.TrimSuffix(name, ".md"))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const config = `---
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
capability: notifications
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
		t.Error("a project with a test command is configured")
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
	if s.Capability != "notifications" {
		t.Errorf("capability = %q, want notifications", s.Capability)
	}
}

// supersedes reads back with the same id forgiveness as parent and depends_on.
func TestLoad_ReadsSupersedes(t *testing.T) {
	root := write(t, config, map[string]string{
		"SPEC-001-old.md": api,
		"SPEC-004-new.md": "---\nid: SPEC-004\ntitle: Replacement\nstatus: proposed\n" +
			"capability: workflow\nsupersedes: [SPEC-002, spec-3]\n---\n",
	})
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := p.Spec("SPEC-004")
	if !ok {
		t.Fatal("SPEC-004 should load")
	}
	if len(s.Supersedes) != 2 || s.Supersedes[0] != "SPEC-002" || s.Supersedes[1] != "SPEC-003" {
		t.Errorf("supersedes = %v, want [SPEC-002 SPEC-003]", s.Supersedes)
	}
}

// The capability view is the delivered specs grouped by capability: names
// alphabetical, contracts by number, and a superseded contract still listed
// but marked with who replaces it.
func TestCapabilities_GroupsOrdersAndMarksSuperseded(t *testing.T) {
	root := write(t, config, map[string]string{
		"SPEC-001-alpha.md":   "---\nid: SPEC-001\ntitle: Alpha\nstatus: done\ncapability: workflow\n---\n",
		"SPEC-002-bravo.md":   "---\nid: SPEC-002\ntitle: Bravo\nstatus: done\ncapability: payments\n---\n",
		"SPEC-003-charlie.md": "---\nid: SPEC-003\ntitle: Charlie\nstatus: done\ncapability: workflow\n---\n",
		"SPEC-004-delta.md":   "---\nid: SPEC-004\ntitle: Delta\nstatus: done\ncapability: payments\nsupersedes: [SPEC-002]\n---\n",
		"SPEC-005-echo.md":    "---\nid: SPEC-005\ntitle: Echo\nstatus: proposed\ncapability: workflow\n---\n",
		"SPEC-006-foxtrot.md": "---\nid: SPEC-006\ntitle: Foxtrot\nstatus: done\n---\n",
		"SPEC-007-golf.md":    "---\nid: SPEC-007\ntitle: Golf\nstatus: dropped\ncapability: workflow\nsupersedes: [SPEC-003]\n---\n",
	})
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}

	groups := p.Capabilities()
	if len(groups) != 2 {
		t.Fatalf("capabilities = %d, want 2: %+v", len(groups), groups)
	}
	if groups[0].Name != "payments" || groups[1].Name != "workflow" {
		t.Fatalf("capabilities should be in name order: %+v", groups)
	}

	payments := groups[0].Contracts
	if len(payments) != 2 || payments[0].ID != "SPEC-002" || payments[1].ID != "SPEC-004" {
		t.Fatalf("payments contracts = %+v, want [SPEC-002 SPEC-004]", payments)
	}
	if got := payments[0].SupersededBy; len(got) != 1 || got[0] != "SPEC-004" {
		t.Errorf("SPEC-002 superseded by = %v, want [SPEC-004]", got)
	}
	if payments[0].Current() {
		t.Error("SPEC-002 is superseded by a done spec, so it is not current")
	}
	if !payments[1].Current() {
		t.Error("SPEC-004 is not superseded, so it is current")
	}

	workflow := groups[1].Contracts
	if len(workflow) != 2 || workflow[0].ID != "SPEC-001" || workflow[1].ID != "SPEC-003" {
		t.Fatalf("workflow contracts = %+v, want [SPEC-001 SPEC-003]", workflow)
	}
	if !workflow[1].Current() {
		t.Error("a dropped spec must not bury the contract it supersedes")
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
	s.Orchestrator = "ana"
	s.Capability = "notifications"
	s.Supersedes = []string{"SPEC-002"}
	s.Agreed["SPEC-002"] = "abc123"
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	again, err := project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, _ := again.Spec("SPEC-003")
	if reloaded.Status != workflow.Specifying || reloaded.Orchestrator != "ana" {
		t.Errorf("not persisted: %+v", reloaded)
	}
	if reloaded.Capability != "notifications" {
		t.Errorf("capability not persisted: %q", reloaded.Capability)
	}
	if len(reloaded.Supersedes) != 1 || reloaded.Supersedes[0] != "SPEC-002" {
		t.Errorf("supersedes not persisted: %v", reloaded.Supersedes)
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

	// An empty list removes the key rather than writing `supersedes: []`.
	reloaded.Supersedes = nil
	if err := reloaded.Save(); err != nil {
		t.Fatal(err)
	}
	cleared, err := project.Load(p.Root)
	if err != nil {
		t.Fatal(err)
	}
	without, _ := cleared.Spec("SPEC-003")
	if without.Doc().Has("supersedes") {
		t.Error("an empty supersedes list should remove the key")
	}
}

// A spec written before the key was renamed still reads its driver from the
// legacy `conductor`, and the next save writes the single `orchestrator` key
// and drops the old one (SPEC-018, decision 4).
func TestSave_MigratesLegacyConductorKey(t *testing.T) {
	root := write(t, config, map[string]string{
		"SPEC-001-legacy.md": "---\nid: SPEC-001\ntitle: Legacy\nstatus: planning\n" +
			"capability: workflow\nconductor: ana\n---\n\n## Contract\n\nx\n",
	})
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := p.Spec("SPEC-001")
	if !ok {
		t.Fatal("SPEC-001 should load")
	}
	if s.Orchestrator != "ana" {
		t.Fatalf("legacy conductor should load as the orchestrator: %q", s.Orchestrator)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "orchestrator: ana") {
		t.Errorf("the save should write orchestrator: ana:\n%s", body)
	}
	if strings.Contains(body, "conductor:") {
		t.Errorf("the save should delete the legacy conductor key:\n%s", body)
	}
}

// The CLI reads English section headings only. A body that uses the retired
// Spanish aliases yields no criteria, contract or questions, so a second
// language cannot become a second set of headings to read (AC4, SPEC-018
// decision 6).
func TestSpec_ReadsEnglishSectionHeadingsOnly(t *testing.T) {
	root := write(t, config, map[string]string{
		"SPEC-001-spanish.md": "---\nid: SPEC-001\ntitle: Spanish\nstatus: accepted\n" +
			"capability: workflow\n---\n\n" +
			"## Criterios de aceptación\n\n- AC1: something\n\n" +
			"## Contrato\n\nGET /something\n\n" +
			"## Preguntas abiertas\n\n- OQ1: something?\n",
	})
	p, err := project.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := p.Spec("SPEC-001")
	if !ok {
		t.Fatal("SPEC-001 should load")
	}
	if got := s.Criteria(); len(got) != 0 {
		t.Errorf("a Spanish-only body should yield no criteria: %v", got)
	}
	if got := s.Contract(); got != "" {
		t.Errorf("a Spanish-only body should yield no contract: %q", got)
	}
	if got := s.OpenQuestions(); got != "" {
		t.Errorf("a Spanish-only body should yield no open questions: %q", got)
	}
}

func TestValidCapability(t *testing.T) {
	if !project.ValidCapability("guard") {
		t.Error("guard is a valid capability")
	}
	for _, bad := range []string{"Guard", "", "with_underscore", "guard!"} {
		if project.ValidCapability(bad) {
			t.Errorf("%q is not a valid capability", bad)
		}
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
	if got := project.SpecDirName("SPEC-004", "Saved card"); got != "SPEC-004-saved-card" {
		t.Errorf("SpecDirName = %q", got)
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
