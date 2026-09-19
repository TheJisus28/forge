package initcmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/initcmd"
)

func TestInit_WritesKitAndStack(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := initcmd.Init(dir, initcmd.Options{}); err != nil {
		t.Fatal(err)
	}
	mustExist(t, dir, "AGENTS.md")
	mustExist(t, dir, "CLAUDE.md")
	mustExist(t, dir, "GEMINI.md")
	mustExist(t, dir, "forge/README.md")
	mustExist(t, dir, "forge/LIFECYCLE.md")
	mustExist(t, dir, "forge/agents/orchestrator.md")
	mustExist(t, dir, "forge/BOARD.md")
	mustExist(t, dir, "forge/playbooks/go.md")
	mustExist(t, dir, "forge/playbooks/python.md")
	mustExist(t, dir, "forge/templates/spec.md")
	mustExist(t, dir, "forge/templates/record.md")
	mustExist(t, dir, "forge/memory/constitution.md")
	mustExist(t, dir, "forge/memory/decisions.md")
	mustExist(t, dir, "forge/memory/adrs/README.md")
	mustExist(t, dir, "forge/backlog/.gitkeep")
	mustExist(t, dir, "forge/specs/.gitkeep")
	mustExist(t, dir, ".cursor/rules/forge.mdc")
	mustExist(t, dir, ".cursor/skills/spec-workflow/SKILL.md")
	mustExist(t, dir, ".claude/skills/spec-workflow/SKILL.md")
	mustExist(t, dir, ".github/copilot-instructions.md")
	body, err := os.ReadFile(filepath.Join(dir, "forge/memory/stack.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "- go") {
		t.Fatalf("stack.md:\n%s", body)
	}
	claude, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if !strings.Contains(string(claude), "@AGENTS.md") {
		t.Fatalf("CLAUDE.md should import AGENTS.md, got %s", claude)
	}
}

func TestInit_RefusesWithoutForce(t *testing.T) {
	dir := t.TempDir()
	if err := initcmd.Init(dir, initcmd.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := initcmd.Init(dir, initcmd.Options{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestInit_ForcePreservesMemory(t *testing.T) {
	dir := t.TempDir()
	if err := initcmd.Init(dir, initcmd.Options{}); err != nil {
		t.Fatal(err)
	}
	stack := filepath.Join(dir, "forge/memory/stack.md")
	custom := []byte("# custom stack\n")
	if err := os.WriteFile(stack, custom, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "forge/README.md"), []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := initcmd.Init(dir, initcmd.Options{Force: true}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(stack)
	if string(got) != string(custom) {
		t.Fatalf("stack overwritten: %s", got)
	}
	readme, _ := os.ReadFile(filepath.Join(dir, "forge/README.md"))
	if strings.TrimSpace(string(readme)) == "old" {
		t.Fatal("kit README was not rewritten")
	}
}

func TestInit_SkipCopilot(t *testing.T) {
	dir := t.TempDir()
	if err := initcmd.Init(dir, initcmd.Options{SkipCopilot: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".github/copilot-instructions.md")); !os.IsNotExist(err) {
		t.Fatalf("copilot file should be absent: %v", err)
	}
}

func mustExist(t *testing.T, root string, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
		t.Fatalf("missing %s: %v", rel, err)
	}
}
