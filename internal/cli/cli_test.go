package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/cli"
)

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

func TestInit_PlantsTheKitAndKeepsYourContent(t *testing.T) {
	dir := t.TempDir()
	mustRun(t, dir, "init")

	for _, rel := range []string{
		"AGENTS.md", "CLAUDE.md",
		".forge/README.md", ".forge/project.md", ".forge/kit/WORKFLOW.md",
		".forge/kit/agents/orchestrator.md", ".forge/kit/templates/spec.md",
		".forge/specs/README.md", ".forge/conventions/README.md",
		".claude/agents/forge-implementer.md", ".claude/skills/forge-onboard/SKILL.md",
		".claude/settings.json",
		".opencode/agents/forge-implementer.md", ".opencode/plugins/forge-guard.js",
		".gitignore",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("missing %s", rel)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows")); err == nil {
		t.Error("CI workflows should only be planted with --ci github")
	}

	settings := read(t, filepath.Join(dir, ".claude", "settings.json"))
	for _, want := range []string{"SessionStart", "forge brief --json", "PreToolUse", "forge guard"} {
		if !strings.Contains(settings, want) {
			t.Errorf("settings.json missing %q:\n%s", want, settings)
		}
	}
	if !strings.Contains(read(t, filepath.Join(dir, ".gitignore")), ".forge/BOARD.md") {
		t.Error("the generated board should be gitignored")
	}

	// The project's own memory survives a second init and an update.
	write(t, filepath.Join(dir, ".forge", "project.md"), projectConfig)
	mustRun(t, dir, "init", "--force")
	mustRun(t, dir, "update")
	if got := read(t, filepath.Join(dir, ".forge", "project.md")); !strings.Contains(got, "A demo.") {
		t.Fatalf("project.md was overwritten:\n%s", got)
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

// The whole point of the tool, end to end.
func TestLifecycle(t *testing.T) {
	dir := newRepo(t)
	specs := filepath.Join(dir, ".forge", "specs")

	mustRun(t, dir, "new", "Saved card payments")
	path := filepath.Join(specs, "SPEC-001-saved-card-payments.md")
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

	if out, code := run(t, dir, "approve", "SPEC-001", "--by", "ana"); code == 0 {
		t.Fatalf("an empty contract must not be approvable: %s", out)
	}
	body = read(t, path)
	write(t, path, strings.Replace(body, "## Contract\n", "## Contract\n\nGET /cards\n", 1))
	mustRun(t, dir, "advance", "SPEC-001", "--to", "awaiting-approval", "--by", "ana")
	// The conductor approving their own contract is fine: there is no
	// separate approver role to ask, and the record still says it was ana.
	mustRun(t, dir, "approve", "SPEC-001", "--by", "ana")
	if !strings.Contains(read(t, path), "contract_hash:") {
		t.Error("approving should fingerprint the contract")
	}

	write(t, filepath.Join(dir, ".forge", "wip", "SPEC-001", "plan.md"), "# Plan\n")
	mustRun(t, dir, "advance", "SPEC-001", "--to", "implementing")
	if out, code := run(t, dir, "archive", "SPEC-001"); code == 0 {
		t.Fatalf("archiving before the review should fail: %s", out)
	}
	mustRun(t, dir, "advance", "SPEC-001", "--to", "reviewing")

	write(t, filepath.Join(dir, ".forge", "wip", "SPEC-001", "changes.md"),
		"# Changes\n\n## Proposed conventions\n\nErrors use an envelope.\n")
	write(t, filepath.Join(dir, ".forge", "wip", "SPEC-001", "review.md"), "Verdict: pass\n")
	if out, code := run(t, dir, "archive", "SPEC-001"); code == 0 {
		t.Fatalf("an undecided convention should block archiving: %s", out)
	}
	write(t, filepath.Join(dir, ".forge", "wip", "SPEC-001", "changes.md"),
		"# Changes\n\n## Proposed conventions\n\nNone.\n")
	mustRun(t, dir, "archive", "SPEC-001")

	if !strings.Contains(read(t, path), "status: done") {
		t.Error("the spec should be done")
	}
	if _, err := os.Stat(filepath.Join(dir, ".forge", "wip", "SPEC-001")); !os.IsNotExist(err) {
		t.Error("archiving should remove the scaffolding")
	}
	if out, code := run(t, dir, "validate"); code != 0 {
		t.Fatalf("the finished project should validate:\n%s", out)
	}
}

func TestHierarchyAndDependencies(t *testing.T) {
	dir := newRepo(t)
	specs := filepath.Join(dir, ".forge", "specs")

	mustRun(t, dir, "new", "Notifications")
	parent := filepath.Join(specs, "SPEC-001-notifications.md")
	write(t, parent, strings.Replace(read(t, parent), "- AC1: ...\n- AC2: ...",
		"- AC1: delivery\n- AC2: history", 1))

	if out, code := run(t, dir, "new", "UI", "--parent", "SPEC-001", "--covers", "AC9"); code == 0 {
		t.Fatalf("covering a criterion nobody promised should fail: %s", out)
	}
	mustRun(t, dir, "new", "API", "--parent", "SPEC-001", "--covers", "AC1")
	mustRun(t, dir, "new", "UI", "--parent", "SPEC-001", "--covers", "AC2")

	ui := filepath.Join(specs, "SPEC-003-ui.md")
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
	api := filepath.Join(specs, "SPEC-002-api.md")
	write(t, api, strings.Replace(read(t, api), "## Contract\n", "## Contract\n\nGET /n\n", 1))
	mustRun(t, dir, "advance", "SPEC-002", "--to", "awaiting-approval")
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
	} {
		if out, code := run(t, dir, "guard", "--file", rel); code != 0 {
			t.Errorf("%s is process paperwork and must be allowed: %s", rel, out)
		}
	}
}

func TestBoardIsGeneratedAndDisposable(t *testing.T) {
	dir := newRepo(t)
	mustRun(t, dir, "new", "Something")
	mustRun(t, dir, "board")
	board := read(t, filepath.Join(dir, ".forge", "BOARD.md"))
	if !strings.Contains(board, "SPEC-001") || !strings.Contains(board, "Do not edit by hand") {
		t.Errorf("board:\n%s", board)
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
