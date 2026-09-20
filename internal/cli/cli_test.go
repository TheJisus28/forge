package cli_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/cli"
	"github.com/TheJisus28/forge/internal/workflow"
)

// runGit runs a git command in dir and fails the test when it fails.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// run executes a command inside dir and returns its output and exit code.
func run(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	before, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(before)

	var out, errOut bytes.Buffer
	code := cli.Main(args, &out, &errOut)
	return out.String() + errOut.String(), code
}

func mustRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, code := run(t, dir, args...)
	if code != 0 {
		t.Fatalf("forge %s failed: %s", strings.Join(args, " "), out)
	}
	return out
}

const projectConfig = `---
test: "go test ./..."
dev: "go run ."
working_language: en
guard: on
---

# Project

A demo.
`

// newRepo returns an initialized, onboarded project.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustRun(t, dir, "init")
	write(t, filepath.Join(dir, ".forge", "project.md"), projectConfig)
	return dir
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

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// docsSection returns the text under the heading whose line starts with
// heading, up to the next heading. It scopes a docs check to the command it
// documents instead of the whole page.
func docsSection(body, heading string) string {
	var b strings.Builder
	found := false
	for _, line := range strings.SplitAfter(body, "\n") {
		if !found {
			if strings.HasPrefix(line, heading) {
				found = true
			}
			continue
		}
		if strings.HasPrefix(line, "##") {
			break
		}
		b.WriteString(line)
	}
	return b.String()
}

// doneSpec writes a delivered spec straight to disk: the capability view only
// reads the file, so a test needs no full lifecycle to have something done.
func doneSpec(t *testing.T, dir, id, title, capability string, supersedes ...string) string {
	t.Helper()
	body := "---\nid: " + id + "\ntitle: " + title + "\nstatus: done\ncapability: " + capability + "\n"
	if len(supersedes) > 0 {
		body += "supersedes: [" + strings.Join(supersedes, ", ") + "]\n"
	}
	body += "---\n\n## Contract\n\nDelivered.\n"
	path := filepath.Join(dir, ".forge", "specs", id, "spec.md")
	write(t, path, body)
	return path
}

// tree reads every file under dir into one comparable string, so a test can
// prove that a command wrote nothing.
func tree(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b.WriteString(rel)
		b.WriteString("\n")
		b.Write(data)
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func TestInit_PlantsTheKitAndKeepsYourContent(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init")

	for _, rel := range []string{
		"AGENTS.md", "CLAUDE.md",
		".forge/README.md", ".forge/project.md",
		".forge/specs/README.md", ".forge/conventions/README.md",
		".claude/agents/forge-implementer.md", ".claude/skills/forge-onboard/SKILL.md",
		".claude/settings.json",
		".opencode/agents/forge-implementer.md", ".opencode/plugins/forge-guard.js",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("missing %s", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".forge", "kit")); err == nil {
		t.Error(".forge/kit should not be planted: the machinery ships in the binary")
	}
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows")); err == nil {
		t.Error("CI workflows should only be planted with --ci github")
	}

	settings := read(t, filepath.Join(dir, ".claude", "settings.json"))
	for _, want := range []string{
		"SessionStart", "forge brief --json", "PreToolUse", "Write|Edit|Bash", "forge guard",
	} {
		if !strings.Contains(settings, want) {
			t.Errorf("settings.json missing %q:\n%s", want, settings)
		}
	}

	// The project's own memory survives a second init and an update.
	write(t, filepath.Join(dir, ".forge", "project.md"), projectConfig)
	mustRun(t, dir, "init", "--force")
	mustRun(t, dir, "update")
	if got := read(t, filepath.Join(dir, ".forge", "project.md")); !strings.Contains(got, "A demo.") {
		t.Fatalf("project.md was overwritten:\n%s", got)
	}
}

// An existing AGENTS.md or CLAUDE.md belongs to the project: init and update
// must not delete the instructions they hold. Only --force rewrites them.
func TestInit_KeepsExistingRootPointers(t *testing.T) {
	dir := t.TempDir()
	const agents = "# My own agent rules\n\nDo not touch.\n"
	const claude = "# My own Claude rules\n"
	write(t, filepath.Join(dir, "AGENTS.md"), agents)
	write(t, filepath.Join(dir, "CLAUDE.md"), claude)

	mustRun(t, dir, "init")
	for name, want := range map[string]string{"AGENTS.md": agents, "CLAUDE.md": claude} {
		if got := read(t, filepath.Join(dir, name)); got != want {
			t.Errorf("init overwrote %s:\n%s", name, got)
		}
	}

	mustRun(t, dir, "update")
	for name, want := range map[string]string{"AGENTS.md": agents, "CLAUDE.md": claude} {
		if got := read(t, filepath.Join(dir, name)); got != want {
			t.Errorf("update overwrote %s:\n%s", name, got)
		}
	}

	mustRun(t, dir, "init", "--force")
	if got := read(t, filepath.Join(dir, "AGENTS.md")); got == agents {
		t.Error("--force should overwrite an existing AGENTS.md")
	}
}

func TestInit_NoGuardAndCI(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init", "--no-guard", "--ci", "github")
	settings := read(t, filepath.Join(dir, ".claude", "settings.json"))
	if strings.Contains(settings, "forge guard") {
		t.Error("--no-guard should not install the PreToolUse hook")
	}
	if _, err := os.Stat(filepath.Join(dir, ".opencode/plugins/forge-guard.js")); !os.IsNotExist(err) {
		t.Error("--no-guard should not plant the opencode guard plugin")
	}
	if _, err := os.Stat(filepath.Join(dir, ".github/workflows/forge-validate.yml")); err != nil {
		t.Error("--ci github should plant the workflows")
	}
}

// --by wins, so a teammate can name themselves without a gh account.
func TestActor_UsesTheFlag(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "Thing", "--capability", "workflow")
	mustRun(t, dir, "accept", "SPEC-001", "--by", "octocat")
	body := read(t, filepath.Join(dir, ".forge", "specs", "SPEC-001-thing", "spec.md"))
	if !strings.Contains(body, "accepted_by: octocat") {
		t.Errorf("accepted_by should be the flag value:\n%s", body)
	}
}

// A spec has no capability unless --capability says which part of the
// system it touches; failing must not leave a half-created folder behind.
func TestNew_RequiresCapability(t *testing.T) {
	dir := newRepo(t)

	out, code := run(t, dir, "new", "Thing")
	if code == 0 {
		t.Fatalf("new without --capability should fail:\n%s", out)
	}
	if !strings.Contains(out, "--capability <name>") {
		t.Errorf("the error should say how to pass a capability:\n%s", out)
	}
	specs, _ := filepath.Glob(filepath.Join(dir, ".forge", "specs", "*", "spec.md"))
	if len(specs) != 0 {
		t.Errorf("nothing should be created without a capability: %v", specs)
	}
}

// The field reaches the file, and an undeclared capability is a warning to
// fix a typo, never a reason to refuse a genuinely new one.
func TestNew_WritesCapabilityAndWarnsOnANewName(t *testing.T) {
	dir := newRepo(t)

	out := mustRun(t, dir, "new", "Guest access", "--capability", "guard")
	if !strings.Contains(out, "no existing spec declares") {
		t.Errorf("an undeclared capability should warn:\n%s", out)
	}
	body := read(t, filepath.Join(dir, ".forge", "specs", "SPEC-001-guest-access", "spec.md"))
	if !strings.Contains(body, "capability: guard") {
		t.Errorf("the spec should carry the capability:\n%s", body)
	}

	out = mustRun(t, dir, "new", "More access", "--capability", "guard")
	if strings.Contains(out, "no existing spec declares") {
		t.Errorf("a declared capability should not warn:\n%s", out)
	}
}

// Accepting work is one gate: `forge new` names `forge accept` and no intake
// pull request, and the pages that describe the loop agree (SPEC-015,
// decision 3).
func TestNew_DescribesOneGate(t *testing.T) {
	dir := newRepo(t)

	out := mustRun(t, dir, "new", "A change", "--capability", "workflow")
	if !strings.Contains(out, "forge accept") {
		t.Errorf("forge new should name the one gate:\n%s", out)
	}
	if strings.Contains(out, "intake") {
		t.Errorf("forge new should not describe an intake pull request:\n%s", out)
	}

	accepted := mustRun(t, dir, "accept", "SPEC-001", "--by", "ana")
	if strings.Contains(accepted, "intake") {
		t.Errorf("forge accept should not describe an intake pull request:\n%s", accepted)
	}

	for _, page := range []string{
		"../../docs/teams.md",
		"../../docs/cli.md",
		"../../kit/forge/specs/README.md",
		"../../kit/claude/skills/forge-work/SKILL.md",
	} {
		if strings.Contains(strings.ToLower(read(t, page)), "intake") {
			t.Errorf("%s should not describe an intake pull request", page)
		}
	}
}

// A number taken on main is confirmed by renumbering when the spec is
// accepted: the folder, the id and the history line all carry it
// (SPEC-015, decision 7).
func TestAccept_RenumbersWhenTakenOnMain(t *testing.T) {
	dir := newRepo(t)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "tester")

	// main already carries SPEC-001, committed before this branch existed.
	write(t, filepath.Join(dir, ".forge", "specs", "SPEC-001-taken", "spec.md"),
		"---\nid: SPEC-001\ntitle: Taken on main\nstatus: done\ncapability: workflow\n---\n\n## Contract\n\nShipped.\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "main")
	runGit(t, dir, "branch", "-M", "main")

	// The branch also created SPEC-001: the race decision 7 names.
	runGit(t, dir, "checkout", "-b", "spec/001-local")
	if err := os.RemoveAll(filepath.Join(dir, ".forge", "specs")); err != nil {
		t.Fatal(err)
	}
	mustRun(t, dir, "new", "Local change", "--capability", "workflow")

	accepted := mustRun(t, dir, "accept", "SPEC-001", "--by", "ana")
	if !strings.Contains(accepted, "SPEC-002") {
		t.Errorf("accept should report the confirmed id:\n%s", accepted)
	}

	specs := filepath.Join(dir, ".forge", "specs")
	renamed := filepath.Join(specs, "SPEC-002-local-change", "spec.md")
	if _, err := os.Stat(renamed); err != nil {
		t.Fatalf("the spec should move to SPEC-002: %v", err)
	}
	if _, err := os.Stat(filepath.Join(specs, "SPEC-001-local-change")); !os.IsNotExist(err) {
		t.Errorf("the provisional folder should be gone, stat err = %v", err)
	}
	body := read(t, renamed)
	for _, want := range []string{
		"id: SPEC-002",
		"status: accepted",
		"accepted_by: ana",
		"renumbered from SPEC-001: taken on main",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the confirmed spec should contain %q:\n%s", want, body)
		}
	}
	if out := mustRun(t, dir, "status"); !strings.Contains(out, "SPEC-002") {
		t.Errorf("status should show the confirmed id:\n%s", out)
	}
}

// A referenced spec cannot be renumbered, because the reference would break:
// accept refuses and leaves the fixing to `forge renumber` (SPEC-015,
// decision 7).
func TestAccept_RefusesWhenReferenced(t *testing.T) {
	dir := newRepo(t)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "tester")

	write(t, filepath.Join(dir, ".forge", "specs", "SPEC-001-taken", "spec.md"),
		"---\nid: SPEC-001\ntitle: Taken on main\nstatus: done\ncapability: workflow\n---\n\n## Contract\n\nShipped.\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "main")
	runGit(t, dir, "branch", "-M", "main")

	runGit(t, dir, "checkout", "-b", "spec/001-local")
	if err := os.RemoveAll(filepath.Join(dir, ".forge", "specs")); err != nil {
		t.Fatal(err)
	}
	mustRun(t, dir, "new", "Local change", "--capability", "workflow")
	write(t, filepath.Join(dir, ".forge", "specs", "SPEC-002-dependent", "spec.md"),
		"---\nid: SPEC-002\ntitle: Dependent\nstatus: proposed\ncapability: workflow\ndepends_on: [SPEC-001]\n---\n")

	out, code := run(t, dir, "accept", "SPEC-001", "--by", "ana")
	if code == 0 {
		t.Fatalf("accept should refuse while the id is referenced:\n%s", out)
	}
	if !strings.Contains(out, "referenced by") || !strings.Contains(out, "SPEC-002") {
		t.Errorf("the refusal should name the reference:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".forge", "specs", "SPEC-001-local-change")); err != nil {
		t.Errorf("the spec should not have been renumbered: %v", err)
	}
}

// Two specs accepted from the same base share an id; `forge validate` reports
// the duplicate and `forge renumber` is the migration (SPEC-015, decision 7).
func TestRenumber_ResolvesTheRace(t *testing.T) {
	dir := newRepo(t)
	specs := filepath.Join(dir, ".forge", "specs")
	write(t, filepath.Join(specs, "SPEC-020-mine", "spec.md"),
		"---\nid: SPEC-020\ntitle: Mine\nstatus: proposed\ncapability: workflow\n---\n")
	write(t, filepath.Join(specs, "SPEC-020-theirs", "spec.md"),
		"---\nid: SPEC-020\ntitle: Theirs\nstatus: proposed\ncapability: workflow\n---\n")

	out, code := run(t, dir, "validate")
	if code == 0 || !strings.Contains(out, "duplicate id") ||
		!strings.Contains(out, "run forge renumber") {
		t.Fatalf("validate should report the duplicate:\n%s", out)
	}

	if out := mustRun(t, dir, "renumber", "SPEC-020"); !strings.Contains(out, "SPEC-021") {
		t.Errorf("renumber should report the new id:\n%s", out)
	}
	if out, code := run(t, dir, "validate"); code != 0 {
		t.Errorf("validate should pass after the renumber:\n%s", out)
	}
	if moved, _ := filepath.Glob(filepath.Join(specs, "SPEC-021-*")); len(moved) != 1 {
		t.Errorf("exactly one spec should be renumbered, got %v", moved)
	}
	if remaining, _ := filepath.Glob(filepath.Join(specs, "SPEC-020-*")); len(remaining) != 1 {
		t.Errorf("exactly one SPEC-020 should remain, got %v", remaining)
	}
}

// Without gh, or with --dry-run, submit prints the commands instead of
// touching the network.
func TestSubmit_DryRunPrintsCommands(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "tester")
	write(t, filepath.Join(dir, "README.md"), "hi\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "init")

	mustRun(t, dir, "init")
	write(t, filepath.Join(dir, ".forge", "project.md"), projectConfig)
	mustRun(t, dir, "new", "Something", "--capability", "workflow")
	runGit(t, dir, "checkout", "-b", "spec/001-something")

	out := mustRun(t, dir, "submit", "--dry-run")
	if !strings.Contains(out, "git push -u origin spec/001-something") {
		t.Errorf("missing git push hint:\n%s", out)
	}
	if !strings.Contains(out, "gh pr create --base main") {
		t.Errorf("missing gh pr create hint:\n%s", out)
	}
}

// The whole point of the tool, end to end.
func TestLifecycle(t *testing.T) {
	dir := newRepo(t)
	specs := filepath.Join(dir, ".forge", "specs")

	mustRun(t, dir, "new", "Saved card payments", "--capability", "payments")
	path := filepath.Join(specs, "SPEC-001-saved-card-payments", "spec.md")
	body := read(t, path)
	if !strings.Contains(body, "status: proposed") {
		t.Fatalf("a new spec starts proposed:\n%s", body)
	}
	write(t, path, strings.Replace(body, "- AC1: ...\n- AC2: ...",
		"- AC1: a saved card can be reused", 1))

	if out, code := run(t, dir, "start", "SPEC-001"); code == 0 {
		t.Fatalf("starting work nobody accepted should fail: %s", out)
	}
	// Anyone can accept; there is no list to be a stranger to.
	mustRun(t, dir, "accept", "SPEC-001", "--by", "jesus")
	mustRun(t, dir, "start", "SPEC-001", "--by", "ana")

	// The template ships guidance in the Contract section; strip it so the
	// contract is genuinely empty and approve has nothing to freeze.
	body = read(t, path)
	empty := regexp.MustCompile(`(?s)## Contract\n.*?\n## Out of scope`).
		ReplaceAllString(body, "## Contract\n\n## Out of scope")
	write(t, path, empty)
	if out, code := run(t, dir, "approve", "SPEC-001", "--by", "ana"); code == 0 {
		t.Fatalf("an empty contract must not be approvable: %s", out)
	}
	body = read(t, path)
	write(t, path, strings.Replace(body, "## Contract\n", "## Contract\n\nGET /cards\n", 1))
	// The orchestrator approving their own contract is fine: there is no
	// separate approver role to ask, and the record still says it was ana.
	mustRun(t, dir, "approve", "SPEC-001", "--by", "ana")
	if !strings.Contains(read(t, path), "contract_hash:") {
		t.Error("approving should fingerprint the contract")
	}

	specDir := filepath.Join(specs, "SPEC-001-saved-card-payments")
	write(t, filepath.Join(specDir, "plan.md"), "# Plan\n\n## Existing state\n\nNone yet.\n")
	write(t, filepath.Join(specDir, "tasks.md"), "# Tasks\n\n- [ ] Phase 1\n")
	mustRun(t, dir, "advance", "SPEC-001", "--to", "implementing")
	if out, code := run(t, dir, "archive", "SPEC-001"); code == 0 {
		t.Fatalf("archiving before the review should fail: %s", out)
	}
	mustRun(t, dir, "advance", "SPEC-001", "--to", "reviewing")

	write(t, filepath.Join(specDir, "tasks.md"),
		"# Tasks\n\n- [x] Phase 1\n\n## Proposed conventions\n\nErrors use an envelope.\n")
	write(t, filepath.Join(specDir, "review.md"), "Verdict: pass\n")
	if out, code := run(t, dir, "archive", "SPEC-001"); code == 0 {
		t.Fatalf("an undecided convention should block archiving: %s", out)
	}
	write(t, filepath.Join(specDir, "tasks.md"),
		"# Tasks\n\n- [x] Phase 1\n\n## Proposed conventions\n\nNone.\n")
	mustRun(t, dir, "archive", "SPEC-001")

	if !strings.Contains(read(t, path), "status: done") {
		t.Error("the spec should be done")
	}
	if _, err := os.Stat(filepath.Join(specDir, "review.md")); err != nil {
		t.Error("archiving should keep the spec folder and its review")
	}
	if out, code := run(t, dir, "validate"); code != 0 {
		t.Fatalf("the finished project should validate:\n%s", out)
	}
	// A later session must see what was delivered, so it can reuse it.
	brief := mustRun(t, dir, "brief")
	if !strings.Contains(brief, "already delivered") || !strings.Contains(brief, "SPEC-001") {
		t.Errorf("the brief should surface delivered work to reuse:\n%s", brief)
	}
}

// The shipped templates carry their Proposed conventions guidance as an HTML
// comment, so a spec that proposed nothing archives without anyone deleting
// the template's own text; a real proposal under the comment still blocks and
// is named instead of the comment (SPEC-019).
func TestArchive_IgnoresTemplateConventionsComment(t *testing.T) {
	dir := newRepo(t)
	specDir := filepath.Join(dir, ".forge", "specs", "SPEC-001-a-change")
	write(t, filepath.Join(specDir, "spec.md"),
		"---\nid: SPEC-001\ntitle: A change\nstatus: reviewing\ncapability: workflow\n---\n\n## Contract\n\nx\n")
	write(t, filepath.Join(specDir, "review.md"), "Verdict: pass\n")
	tasks := mustRun(t, dir, "template", "tasks")

	write(t, filepath.Join(specDir, "tasks.md"),
		strings.Replace(tasks, "\nNone.\n", "\n- Errors use an envelope.\n", 1))
	out, code := run(t, dir, "archive", "SPEC-001")
	if code == 0 {
		t.Fatalf("a real proposal should block archiving:\n%s", out)
	}
	if !strings.Contains(out, "Errors use an envelope.") {
		t.Errorf("the proposal should be the reported line:\n%s", out)
	}
	if strings.Contains(out, "Patterns you had to decide") {
		t.Errorf("the template comment should not be read as a proposal:\n%s", out)
	}

	write(t, filepath.Join(specDir, "tasks.md"), tasks)
	mustRun(t, dir, "archive", "SPEC-001")
	if body := read(t, filepath.Join(specDir, "spec.md")); !strings.Contains(body, "status: done") {
		t.Errorf("the shipped template default should not block archiving:\n%s", body)
	}
}

// The CLI reads English section headings only: a proposal under the retired
// Spanish heading is invisible, while one under `Proposed conventions` blocks
// archiving (AC4, SPEC-018 decision 6).
func TestArchive_ReadsOnlyEnglishConventionHeading(t *testing.T) {
	prepare := func(t *testing.T) (string, string) {
		t.Helper()
		dir := newRepo(t)
		specDir := filepath.Join(dir, ".forge", "specs", "SPEC-001-a-change")
		write(t, filepath.Join(specDir, "spec.md"),
			"---\nid: SPEC-001\ntitle: A change\nstatus: reviewing\ncapability: workflow\n---\n\n## Contract\n\nx\n")
		write(t, filepath.Join(specDir, "review.md"), "Verdict: pass\n")
		return dir, specDir
	}

	t.Run("English heading blocks", func(t *testing.T) {
		dir, specDir := prepare(t)
		write(t, filepath.Join(specDir, "tasks.md"),
			"# Tasks\n\n- [x] Phase 1\n\n## Proposed conventions\n\nErrors use an envelope.\n")
		out, code := run(t, dir, "archive", "SPEC-001")
		if code == 0 {
			t.Fatalf("a proposal under the English heading should block archiving:\n%s", out)
		}
		if !strings.Contains(out, "Errors use an envelope.") {
			t.Errorf("the proposal should be the reported line:\n%s", out)
		}
	})

	t.Run("Spanish heading is invisible", func(t *testing.T) {
		dir, specDir := prepare(t)
		write(t, filepath.Join(specDir, "tasks.md"),
			"# Tasks\n\n- [x] Phase 1\n\n## Convenciones propuestas\n\nErrors use an envelope.\n")
		mustRun(t, dir, "archive", "SPEC-001")
		if body := read(t, filepath.Join(specDir, "spec.md")); !strings.Contains(body, "status: done") {
			t.Errorf("a proposal under the Spanish heading should not block archiving:\n%s", body)
		}
	})
}

func TestHierarchyAndDependencies(t *testing.T) {
	dir := newRepo(t)
	specs := filepath.Join(dir, ".forge", "specs")

	mustRun(t, dir, "new", "Notifications", "--capability", "notifications")
	parent := filepath.Join(specs, "SPEC-001-notifications", "spec.md")
	write(t, parent, strings.Replace(read(t, parent), "- AC1: ...\n- AC2: ...",
		"- AC1: delivery\n- AC2: history", 1))

	if out, code := run(t, dir, "new", "UI", "--capability", "notifications", "--parent", "SPEC-001", "--covers", "AC9"); code == 0 {
		t.Fatalf("covering a criterion nobody promised should fail: %s", out)
	}
	mustRun(t, dir, "new", "API", "--capability", "notifications", "--parent", "SPEC-001", "--covers", "AC1")
	mustRun(t, dir, "new", "UI", "--capability", "notifications", "--parent", "SPEC-001", "--covers", "AC2")

	ui := filepath.Join(specs, "SPEC-003-ui", "spec.md")
	write(t, ui, strings.Replace(read(t, ui), "status: proposed",
		"status: proposed\ndepends_on: [SPEC-002@contract]", 1))

	mustRun(t, dir, "accept", "SPEC-002", "--by", "jesus")
	mustRun(t, dir, "accept", "SPEC-003", "--by", "jesus")

	out, code := run(t, dir, "start", "SPEC-003")
	if code == 0 {
		t.Fatal("the UI must wait for the API contract")
	}
	if !strings.Contains(out, "no approved contract") {
		t.Errorf("the reason should be the contract: %s", out)
	}
	if !strings.Contains(out, "SPEC-002") {
		t.Errorf("it should say what is ready instead: %s", out)
	}

	// The parent is never started directly.
	mustRun(t, dir, "accept", "SPEC-001", "--by", "jesus")
	if out, code := run(t, dir, "start", "SPEC-001"); code == 0 ||
		!strings.Contains(out, "children") {
		t.Fatalf("a parent has no code of its own: %s", out)
	}

	// Approving the API contract unblocks the UI without waiting for code.
	mustRun(t, dir, "start", "SPEC-002", "--by", "ana")
	api := filepath.Join(specs, "SPEC-002-api", "spec.md")
	write(t, api, strings.Replace(read(t, api), "## Contract\n", "## Contract\n\nGET /n\n", 1))
	mustRun(t, dir, "approve", "SPEC-002", "--by", "jesus")
	mustRun(t, dir, "start", "SPEC-003", "--by", "jose")

	if !strings.Contains(read(t, ui), "agreed_contracts") {
		t.Error("starting against a contract should record which version")
	}
	status := mustRun(t, dir, "status", "SPEC-001")
	if !strings.Contains(status, "AC1") || !strings.Contains(status, "SPEC-002") {
		t.Errorf("the coverage matrix should show who delivers what:\n%s", status)
	}
}

func TestBriefAndGuard(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init")

	out := mustRun(t, dir, "brief")
	if !strings.Contains(out, "not onboarded") {
		t.Errorf("an unconfigured project should say so first:\n%s", out)
	}
	write(t, filepath.Join(dir, ".forge", "project.md"), projectConfig)

	out = mustRun(t, dir, "brief", "--json")
	if !strings.Contains(out, `"hookEventName":"SessionStart"`) ||
		!strings.Contains(out, "additionalContext") {
		t.Errorf("the hook payload is wrong:\n%s", out)
	}

	out = mustRun(t, dir, "guard", "--explain")
	if !strings.Contains(out, "would deny") {
		t.Errorf("without a spec branch, product code is denied:\n%s", out)
	}

	write(t, filepath.Join(dir, ".forge", "project.md"),
		strings.Replace(projectConfig, "guard: on", "guard: off", 1))
	if out := mustRun(t, dir, "guard", "--explain"); !strings.Contains(out, "") ||
		strings.Contains(out, "would deny") {
		t.Errorf("a disabled guard must stay out of the way: %s", out)
	}
}

// `forge guard --file` is the hook-free mode other agents (opencode) call.
func TestGuardFileMode(t *testing.T) {
	dir := newRepo(t)

	if out, code := run(t, dir, "guard", "--file", "src/example"); code == 0 ||
		!strings.Contains(out, "no spec") {
		t.Fatalf("product code without a spec should be denied: %q", out)
	}
	for _, rel := range []string{
		".forge/specs/SPEC-001.md",
		".opencode/plugins/forge-guard.js",
		"opencode.json",
		"CHANGELOG.md",
	} {
		if out, code := run(t, dir, "guard", "--file", rel); code != 0 {
			t.Errorf("%s is process paperwork and must be allowed: %s", rel, out)
		}
	}
	for _, rel := range []string{
		"internal/cli/guard.go",
		"main.go",
		"docs/customizing.md",
		"go.mod",
	} {
		if out, code := run(t, dir, "guard", "--file", rel); code != 1 ||
			!strings.Contains(out, "no spec") {
			t.Errorf("%s is product code and must be denied: %q", rel, out)
		}
	}
}

// The guard also refuses what would land on the default branch, reading the
// command segment by segment so a later mention of main is not a push to it.
func TestGuardCommand(t *testing.T) {
	dir := newRepo(t)

	for _, cmd := range []string{
		"gh pr merge 3", "git push origin main", "git push origin HEAD:master",
		"git push origin main:feature", "git -C . push origin master",
	} {
		if out, code := run(t, dir, "guard", "--command", cmd); code == 0 {
			t.Errorf("%q should be denied: %s", cmd, out)
		}
	}
	for _, cmd := range []string{
		"go test ./...", "git push -u origin spec/004-x", "git push origin feature/main", "git status",
		"git push -u origin my-branch && gh pr create --base main",
		"git push -u origin my-branch ; echo main",
		"git push -u origin my-branch || gh pr create --base main",
		"echo git push origin main",
	} {
		if out, code := run(t, dir, "guard", "--command", cmd); code != 0 {
			t.Errorf("%q should be allowed: %s", cmd, out)
		}
	}

	if out := mustRun(t, dir, "guard", "--command", "git push -u origin my-branch && gh pr create --base main", "--explain"); !strings.Contains(out, "would allow") {
		t.Errorf("the compound push should explain as allowed: %s", out)
	}
	if out := mustRun(t, dir, "guard", "--command", "git push origin main", "--explain"); !strings.Contains(out, "would deny") {
		t.Errorf("pushing main should explain as denied: %s", out)
	}
}

// A bare `git push` follows the current branch: denied on the default branch,
// allowed on a feature branch (SPEC-014 AC3).
func TestGuardCommand_BarePushFollowsTheBranch(t *testing.T) {
	dir := newRepo(t)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "T")
	runGit(t, dir, "checkout", "-b", "main")
	write(t, filepath.Join(dir, "README.md"), "# x\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")

	if out, code := run(t, dir, "guard", "--command", "git push"); code == 0 {
		t.Errorf("a bare push on main should be denied: %s", out)
	}
	runGit(t, dir, "checkout", "-b", "feature")
	if out, code := run(t, dir, "guard", "--command", "git push"); code != 0 {
		t.Errorf("a bare push on a feature branch should be allowed: %s", out)
	}
}

// Only a person accepts a spec into the queue: the agent guard denies the
// command, inside a compound too, and `guard: off` puts it back
// (SPEC-015, decision 8).
func TestGuard_DeniesForgeAcceptForAgents(t *testing.T) {
	dir := newRepo(t)

	if out, code := run(t, dir, "guard", "--command", "forge accept SPEC-020"); code == 0 ||
		!strings.Contains(out, "only a person accepts") {
		t.Fatalf("forge accept should be denied for an agent: %q", out)
	}
	if out, code := run(t, dir, "guard", "--command", "cd .forge && forge accept SPEC-020"); code == 0 ||
		!strings.Contains(out, "only a person accepts") {
		t.Errorf("a compound forge accept should be denied too: %q", out)
	}
	if out, code := run(t, dir, "guard", "--command", `forge new "x" --capability workflow`); code != 0 {
		t.Errorf("forge new should be allowed: %s", out)
	}
	if out := mustRun(t, dir, "guard", "--explain", "--command", "forge accept SPEC-020"); !strings.Contains(out, "would deny") {
		t.Errorf("explain should say would deny: %s", out)
	}

	write(t, filepath.Join(dir, ".forge", "project.md"),
		strings.Replace(projectConfig, "guard: on", "guard: off", 1))
	if out, code := run(t, dir, "guard", "--command", "forge accept SPEC-020"); code != 0 {
		t.Errorf("guard off should allow forge accept: %s", out)
	}
}

// The guard reports the contract phase by its current name whichever retired
// name the frontmatter still carries, because the spec loads as `contracting`
// and the denial names `forge approve` (SPEC-015, decision 6).
func TestGuard_NamesContracting(t *testing.T) {
	dir := newRepo(t)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "T")
	write(t, filepath.Join(dir, "README.md"), "# x\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "checkout", "-b", "spec/001-thing")

	spec := filepath.Join(dir, ".forge", "specs", "SPEC-001-thing", "spec.md")
	for _, status := range []string{"contracting", "specifying", "awaiting-approval"} {
		write(t, spec, "---\nid: SPEC-001\ntitle: Thing\nstatus: "+status+
			"\ncapability: workflow\n---\n\n## Contract\n\nGET /things\n")

		out, code := run(t, dir, "guard", "--file", "src/x")
		if code == 0 {
			t.Fatalf("a %s spec should deny product code: %s", status, out)
		}
		if !strings.Contains(out, "contracting") || !strings.Contains(out, "forge approve") {
			t.Errorf("the %s denial should name contracting and forge approve: %q", status, out)
		}
		if strings.Contains(out, "specifying") || strings.Contains(out, "awaiting-approval") {
			t.Errorf("the %s denial should not name a retired state: %q", status, out)
		}
	}
}

// forge migrate rewrites the retired status scalars to contracting, leaves the
// body including ## History byte-identical, writes nothing under --dry-run,
// and reports a clean no-op on a second run (SPEC-015, decision 6).
func TestMigrate_RewritesRetiredStatus(t *testing.T) {
	dir := newRepo(t)

	specBody := func(id, status string) string {
		return "---\nid: " + id + "\ntitle: " + id + "\nstatus: " + status +
			"\ncapability: workflow\n---\n\n## Problem\n\nSomething.\n\n" +
			"## History\n\nWritten by `forge`. Do not edit by hand.\n" +
			"- 2026-09-20  accepted  by ana\n" +
			"- 2026-09-20  " + status + "  by ana\n"
	}
	specs := map[string]string{
		filepath.Join(dir, ".forge", "specs", "SPEC-001-old", "spec.md"): "specifying",
		filepath.Join(dir, ".forge", "specs", "SPEC-002-old", "spec.md"): "awaiting-approval",
	}
	for path, status := range specs {
		id := strings.TrimSuffix(filepath.Base(filepath.Dir(path)), "-old")
		write(t, path, specBody(id, status))
	}

	historyOf := func(path string) string {
		body := read(t, path)
		i := strings.Index(body, "## History")
		if i < 0 {
			t.Fatalf("%s has no History section:\n%s", path, body)
		}
		return body[i:]
	}
	beforeHistory := map[string]string{}
	for path := range specs {
		beforeHistory[path] = historyOf(path)
	}

	before := tree(t, filepath.Join(dir, ".forge"))
	out := mustRun(t, dir, "migrate", "--dry-run")
	for path := range specs {
		rel, _ := filepath.Rel(dir, path)
		if !strings.Contains(out, filepath.ToSlash(rel)) {
			t.Errorf("--dry-run should list %s:\n%s", rel, out)
		}
	}
	if after := tree(t, filepath.Join(dir, ".forge")); after != before {
		t.Errorf("forge migrate --dry-run wrote to .forge:\nbefore:\n%s\nafter:\n%s", before, after)
	}

	mustRun(t, dir, "migrate")
	for path := range specs {
		got := read(t, path)
		if !strings.Contains(got, "status: contracting") {
			t.Errorf("%s should say contracting:\n%s", path, got)
		}
		if historyOf(path) != beforeHistory[path] {
			t.Errorf("%s rewrote ## History:\nbefore:\n%s\nafter:\n%s",
				path, beforeHistory[path], historyOf(path))
		}
	}

	out = mustRun(t, dir, "migrate")
	if !strings.Contains(out, "nothing to migrate") {
		t.Errorf("a current tree should print nothing to migrate: %q", out)
	}
}

// Open questions block the contract: a design built on them is the mistake.
func TestApprove_RefusesOpenQuestions(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "Thing", "--capability", "workflow")
	mustRun(t, dir, "accept", "SPEC-001", "--by", "ana")
	mustRun(t, dir, "start", "SPEC-001", "--by", "ana")

	spec := filepath.Join(dir, ".forge", "specs", "SPEC-001-thing", "spec.md")
	body := read(t, spec)
	body = strings.Replace(body, "## Contract\n", "## Contract\n\nGET /things\n", 1)
	write(t, spec, strings.Replace(body, "## Open questions\n",
		"## Open questions\n\n- OQ1: which store?\n", 1))

	if out, code := run(t, dir, "approve", "SPEC-001", "--by", "ana"); code == 0 {
		t.Fatalf("approve should refuse while questions are open: %s", out)
	}
}

// Approval reads a contract written straight from contracting: there is no
// separate move to a review state, and the approver and the fingerprint are
// still frozen (SPEC-015, decision 2).
func TestApprove_StraightFromContracting(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "Thing", "--capability", "workflow")
	mustRun(t, dir, "accept", "SPEC-001", "--by", "ana")
	mustRun(t, dir, "start", "SPEC-001", "--by", "ana")

	spec := filepath.Join(dir, ".forge", "specs", "SPEC-001-thing", "spec.md")
	body := read(t, spec)
	write(t, spec, strings.Replace(body, "## Contract\n", "## Contract\n\nGET /things\n", 1))

	mustRun(t, dir, "approve", "SPEC-001", "--by", "jesus")

	got := read(t, spec)
	if !strings.Contains(got, "status: planning") {
		t.Errorf("approval should move the spec straight to planning:\n%s", got)
	}
	if !strings.Contains(got, "approved_by: jesus") || !strings.Contains(got, "contract_hash:") {
		t.Errorf("approval should record the approver and the fingerprint:\n%s", got)
	}
	if strings.Contains(got, "awaiting-approval") {
		t.Errorf("history should not name the retired state:\n%s", got)
	}
}

// The detail shows how far the tasks have gone, straight from tasks.md.
func TestStatusShowsTaskProgress(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "Thing", "--capability", "workflow")
	mustRun(t, dir, "accept", "SPEC-001", "--by", "ana")
	mustRun(t, dir, "start", "SPEC-001", "--by", "ana")

	write(t, filepath.Join(dir, ".forge", "specs", "SPEC-001-thing", "tasks.md"),
		"# Tasks\n\n- [x] Phase 1\n- [ ] Phase 2\n")

	out := mustRun(t, dir, "status", "SPEC-001")
	if !strings.Contains(out, "tasks") || !strings.Contains(out, "1/2") {
		t.Errorf("status should show task progress:\n%s", out)
	}
}

// Both the per-spec status and the session brief name the capability, so an
// agent starts knowing which part of the system it is about.
func TestStatusAndBriefShowCapability(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "Guest access", "--capability", "guard")

	out := mustRun(t, dir, "status", "SPEC-001")
	if !strings.Contains(out, "capability") || !strings.Contains(out, "guard") {
		t.Errorf("status should show the capability:\n%s", out)
	}

	// The brief only names the current spec, which is the one on the branch.
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "tester")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "init")
	runGit(t, dir, "checkout", "-b", "spec/001-guest-access")

	out = mustRun(t, dir, "brief")
	if !strings.Contains(out, "capability") || !strings.Contains(out, "guard") {
		t.Errorf("the brief should show the current spec's capability:\n%s", out)
	}
}

// A spec that replaces a delivered one shows the link from both ends: what
// it supersedes, and what supersedes it.
func TestStatusShowsSupersedes(t *testing.T) {
	dir := newRepo(t)
	specs := filepath.Join(dir, ".forge", "specs")

	mustRun(t, dir, "new", "Old contract", "--capability", "workflow")
	old := filepath.Join(specs, "SPEC-001-old-contract", "spec.md")
	mustRun(t, dir, "accept", "SPEC-001", "--by", "ana")
	mustRun(t, dir, "start", "SPEC-001", "--by", "ana")
	write(t, old, strings.Replace(read(t, old), "## Contract\n", "## Contract\n\nGET /old\n", 1))
	mustRun(t, dir, "approve", "SPEC-001", "--by", "ana")

	oldDir := filepath.Dir(old)
	write(t, filepath.Join(oldDir, "plan.md"), "# Plan\n\n## Existing state\n\nNone yet.\n")
	write(t, filepath.Join(oldDir, "tasks.md"),
		"# Tasks\n\n- [x] Phase 1\n\n## Proposed conventions\n\nNone.\n")
	mustRun(t, dir, "advance", "SPEC-001", "--to", "implementing")
	mustRun(t, dir, "advance", "SPEC-001", "--to", "reviewing")
	write(t, filepath.Join(oldDir, "review.md"), "Verdict: pass\n")
	mustRun(t, dir, "archive", "SPEC-001")

	mustRun(t, dir, "new", "New contract", "--capability", "workflow")
	newPath := filepath.Join(specs, "SPEC-002-new-contract", "spec.md")
	write(t, newPath, strings.Replace(read(t, newPath), "status: proposed",
		"status: proposed\nsupersedes: [SPEC-001]", 1))

	out := mustRun(t, dir, "status", "SPEC-002")
	if !strings.Contains(out, "supersedes") || !strings.Contains(out, "SPEC-001") {
		t.Errorf("status should show what the spec supersedes:\n%s", out)
	}
	out = mustRun(t, dir, "status", "SPEC-001")
	if !strings.Contains(out, "superseded by") || !strings.Contains(out, "SPEC-002") {
		t.Errorf("status should show what supersedes the spec:\n%s", out)
	}
}

// The derived view lists each capability with the done contracts under it,
// alphabetically, and says nothing about in-flight work.
func TestCapabilities_ListsByCapability(t *testing.T) {
	dir := newRepo(t)
	doneSpec(t, dir, "SPEC-001", "Greeting", "payments")
	doneSpec(t, dir, "SPEC-002", "Deleting", "workflow")
	write(t, filepath.Join(dir, ".forge", "specs", "SPEC-003", "spec.md"),
		"---\nid: SPEC-003\ntitle: Proposed\nstatus: proposed\ncapability: workflow\n---\n")

	out := mustRun(t, dir, "capabilities")
	for _, want := range []string{"payments", "SPEC-001", "Greeting", "workflow", "SPEC-002"} {
		if !strings.Contains(out, want) {
			t.Errorf("capabilities output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "SPEC-003") {
		t.Errorf("a spec that is not done is not part of the view:\n%s", out)
	}
	if strings.Index(out, "payments") > strings.Index(out, "workflow") {
		t.Errorf("capabilities should be alphabetical:\n%s", out)
	}
}

// A superseded contract stays visible, marked with who replaces it.
func TestCapabilities_MarksSuperseded(t *testing.T) {
	dir := newRepo(t)
	doneSpec(t, dir, "SPEC-001", "Old", "workflow")
	doneSpec(t, dir, "SPEC-002", "New", "workflow", "SPEC-001")

	out := mustRun(t, dir, "capabilities")
	if !strings.Contains(out, "(superseded by SPEC-002)") {
		t.Errorf("the superseded contract should say by whom:\n%s", out)
	}
	if !strings.Contains(out, "SPEC-001") {
		t.Errorf("the superseded contract should not be omitted:\n%s", out)
	}
}

// With a name, only that capability is shown; an unknown one is an error.
func TestCapabilities_OneName(t *testing.T) {
	dir := newRepo(t)
	doneSpec(t, dir, "SPEC-001", "Payments", "payments")
	doneSpec(t, dir, "SPEC-002", "Workflow", "workflow")

	out := mustRun(t, dir, "capabilities", "payments")
	if !strings.Contains(out, "SPEC-001") || strings.Contains(out, "SPEC-002") {
		t.Errorf("a named capability should show only its contracts:\n%s", out)
	}
	if out, code := run(t, dir, "capabilities", "nope"); code == 0 {
		t.Errorf("an unknown capability should fail:\n%s", out)
	}
}

// The view is derived and read-only: the same tree prints identical bytes,
// and no file under .forge changes.
func TestCapabilities_IsDeterministicAndWritesNothing(t *testing.T) {
	dir := newRepo(t)
	doneSpec(t, dir, "SPEC-001", "Old", "workflow")
	doneSpec(t, dir, "SPEC-002", "New", "workflow", "SPEC-001")

	before := tree(t, filepath.Join(dir, ".forge"))
	first := mustRun(t, dir, "capabilities")
	second := mustRun(t, dir, "capabilities")
	if first != second {
		t.Errorf("the same tree should print identical bytes:\n%q\n%q", first, second)
	}
	after := tree(t, filepath.Join(dir, ".forge"))
	if before != after {
		t.Errorf("forge capabilities wrote to .forge:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// The brief mirrors the command's summary, so a session starts knowing the
// current shape instead of reading every contract.
func TestBrief_ShowsCapabilitySummary(t *testing.T) {
	dir := newRepo(t)
	doneSpec(t, dir, "SPEC-001", "Payments", "payments")
	doneSpec(t, dir, "SPEC-002", "First", "workflow")
	doneSpec(t, dir, "SPEC-003", "Second", "workflow")
	doneSpec(t, dir, "SPEC-004", "Replacement", "payments", "SPEC-001")

	brief := mustRun(t, dir, "brief")
	if !strings.Contains(brief, "capabilities:") {
		t.Fatalf("the brief should carry the capability summary:\n%s", brief)
	}
	if !strings.Contains(brief, "2 current contracts") {
		t.Errorf("workflow has two current contracts:\n%s", brief)
	}
	if !strings.Contains(brief, "1 current contract") {
		t.Errorf("payments has one current contract:\n%s", brief)
	}
}

// Repository-root paperwork is process files, not product code: the unused
// community boilerplate is gone and the contributor guide lives in AGENTS.md.
// The removed names are assembled from fragments so no live file spells them
// out; the durable record under .forge/specs/ is the only place they may stay.
func TestRepositoryPaperwork(t *testing.T) {
	contributing := "CONTRIB" + "UTING.md"
	security := "SECUR" + "ITY.md"
	conduct := "CODE_OF_" + "CONDUCT.md"

	for _, rel := range []string{contributing, security, conduct} {
		if _, err := os.Stat(filepath.Join("..", "..", rel)); !os.IsNotExist(err) {
			t.Errorf("%s should be gone from the tree", rel)
		}
	}

	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(readme), contributing) {
		t.Error("README.md should not reference the removed contributor guide")
	}
	if !strings.Contains(string(readme), "AGENTS.md") {
		t.Error("README.md should point contributors at AGENTS.md")
	}

	agents, err := os.ReadFile("../../AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"## Contributing", "Adding a command", "Supporting another agent"} {
		if !strings.Contains(string(agents), want) {
			t.Errorf("AGENTS.md should contain %q", want)
		}
	}
}

// The guard's definition of root paperwork is written down where the guard is
// documented, so the code and the pages cannot drift apart.
func TestDocs_DescribeProcessFiles(t *testing.T) {
	doc, err := os.ReadFile("../../docs/customizing.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Markdown", "LICENSE"} {
		if !strings.Contains(string(doc), want) {
			t.Errorf("docs/customizing.md should contain %q", want)
		}
	}
}

// The CLI reference documents the flag that opens a spec.
func TestDocs_DescribeCapability(t *testing.T) {
	doc, err := os.ReadFile("../../docs/cli.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc), "--capability") {
		t.Error("docs/cli.md should document --capability")
	}
}

// The CLI reference documents the derived capability view.
func TestDocs_DescribeCapabilities(t *testing.T) {
	doc, err := os.ReadFile("../../docs/cli.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc), "forge capabilities") {
		t.Error("docs/cli.md should document forge capabilities")
	}
}

// The pages an agent or a person reads must not keep a second copy of the
// state machine: an arrow chain between two states, or a table row whose first
// cell is a state. The names come from the binary, so a page cannot keep a
// list that diverged from it (AC1, AC5).
func TestDocs_DoNotRestateTheStateMachine(t *testing.T) {
	pages := []string{"../../AGENTS.md", "../../kit/AGENTS.md", "../../README.md"}
	docs, err := filepath.Glob("../../docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	pages = append(pages, docs...)

	names := make([]string, 0, len(workflow.All()))
	for _, s := range workflow.All() {
		names = append(names, regexp.QuoteMeta(string(s)))
	}
	alt := strings.Join(names, "|")
	chain := regexp.MustCompile(`\b(?:` + alt + `)\s*(?:→|↔)\s*(?:` + alt + `)\b`)
	row := regexp.MustCompile("(?m)^\\|\\s*`(?:" + alt + ")`\\s*\\|")

	for _, page := range pages {
		body := read(t, page)
		if m := chain.FindString(body); m != "" {
			t.Errorf("%s restates the state machine as an arrow chain: %q", page, m)
		}
		if m := row.FindString(body); m != "" {
			t.Errorf("%s restates the state machine as a table row: %q", page, m)
		}
	}
}

// docs/ explains the process and points at the binary for the machine instead
// of holding an ordered state list (AC5).
func TestDocs_WorkflowPointsAtForgeWorkflow(t *testing.T) {
	if !strings.Contains(read(t, "../../docs/workflow.md"), "forge workflow") {
		t.Error("docs/workflow.md should point at `forge workflow` for the states and transitions")
	}
}

// `forge start` creates the spec folder and records fingerprints; planning
// writes `plan.md` and `tasks.md` after approval. The page must match the
// command and not claim start creates them (AC3).
func TestDocs_ForgeStartDoesNotCreatePlanningFiles(t *testing.T) {
	start := docsSection(read(t, "../../docs/cli.md"), "### `forge start")
	if start == "" {
		t.Fatal("docs/cli.md should document `forge start`")
	}

	claim := regexp.MustCompile("(?i)creates?[^.\\n]{0,60}`(plan\\.md|tasks\\.md)`")
	for _, m := range claim.FindAllStringIndex(start, -1) {
		if strings.HasSuffix(strings.ToLower(start[:m[0]]), "not ") {
			continue
		}
		t.Errorf("docs/cli.md claims `forge start` creates a planning file: %q", start[m[0]:m[1]])
	}
}

// The driver is recorded under one key: `forge start` writes `orchestrator`
// and never the retired `conductor`, and `forge status` names it (AC2).
func TestStart_RecordsOneOrchestratorName(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "A change", "--capability", "workflow")
	path := filepath.Join(dir, ".forge", "specs", "SPEC-001-a-change", "spec.md")

	mustRun(t, dir, "accept", "SPEC-001", "--by", "jesus")
	mustRun(t, dir, "start", "SPEC-001", "--by", "ana")

	body := read(t, path)
	if !strings.Contains(body, "orchestrator: ana") {
		t.Errorf("start should record the orchestrator:\n%s", body)
	}
	if strings.Contains(body, "conductor:") {
		t.Errorf("start should not write the legacy conductor key:\n%s", body)
	}
	if out := mustRun(t, dir, "status", "SPEC-001"); !strings.Contains(out, "orchestrator   ana") {
		t.Errorf("status should name the orchestrator:\n%s", out)
	}
}

// One name for the driver in every page a reader or an agent opens, and in the
// roles the binary lists. The `.forge/` records are history and are not
// scanned (AC2).
func TestDocs_DoNotSayConductor(t *testing.T) {
	pages := []string{
		"../../AGENTS.md",
		"../../kit/AGENTS.md",
		"../../kit/machine/WORKFLOW.md",
	}
	docs, err := filepath.Glob("../../docs/*.md")
	if err != nil {
		t.Fatal(err)
	}
	pages = append(pages, docs...)
	for _, page := range pages {
		if strings.Contains(read(t, page), "conductor") {
			t.Errorf("%s should say orchestrator, not conductor", page)
		}
	}

	if out := mustRun(t, t.TempDir(), "roles"); strings.Contains(out, "conductor") {
		t.Errorf("forge roles should not list a conductor:\n%s", out)
	}
}

// Headings are fixed English; `working_language` governs the prose inside a
// spec, decision or convention, never the headings the CLI reads. The page
// names the English headings and makes no claim that a translated heading is
// accepted (AC4, SPEC-018 decision 6).
func TestDocs_CustomizingDoesNotAcceptTranslatedHeadings(t *testing.T) {
	doc := strings.Join(strings.Fields(read(t, "../../docs/customizing.md")), " ")

	alias := regexp.QuoteMeta("Criterios de aceptación")
	claim := regexp.MustCompile(`(?i)accept(?:s|ed)?[^.;]{0,120}` + alias +
		`|` + alias + `[^.;]{0,120}accept(?:s|ed)?`)
	for _, m := range claim.FindAllString(doc, -1) {
		lower := strings.ToLower(m)
		if strings.Contains(lower, "not accept") ||
			strings.Contains(lower, "no longer accept") ||
			strings.Contains(lower, "never accept") {
			continue
		}
		t.Errorf("docs/customizing.md claims the parser accepts a translated heading: %q", m)
	}

	for _, want := range []string{
		"Acceptance criteria", "Contract", "Open questions",
		"Proposed conventions", "Existing state",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("docs/customizing.md should name the fixed English heading %q", want)
		}
	}
}

func TestUnknownCommandAndMissingProject(t *testing.T) {
	dir := t.TempDir()
	if out, code := run(t, dir, "frobnicate"); code == 0 || !strings.Contains(out, "unknown command") {
		t.Fatalf("out=%q code=%d", out, code)
	}
	if out, code := run(t, dir, "status"); code == 0 || !strings.Contains(out, "forge init") {
		t.Fatalf("outside a project the error should point at init: %q", out)
	}
}
