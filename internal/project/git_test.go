package project_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/internal/project"
)

// gitOut runs a git command in dir and returns its trimmed stdout.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
}

// gitRepo builds a repository with one commit on branch, wired to a local bare
// remote at `origin`. The remote is a directory, so no test touches the
// network (SPEC-020, the Risks note).
func gitRepo(t *testing.T, branch string) string {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "origin.git")
	runGit(t, filepath.Dir(remote), "init", "--bare", remote)

	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "tester")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "root")
	runGit(t, dir, "branch", "-M", branch)
	runGit(t, dir, "remote", "add", "origin", remote)
	return dir
}

// A clean tree commits nothing: false, nil and HEAD unchanged. A dirty tree
// commits the staging-all policy with the given message as the subject.
func TestCommitAll(t *testing.T) {
	dir := gitRepo(t, "main")
	before := gitOut(t, dir, "rev-parse", "HEAD")

	committed, err := project.CommitAll(dir, "chore(SPEC-020): checkpoint implementing")
	if err != nil {
		t.Fatal(err)
	}
	if committed {
		t.Error("a clean tree must not commit")
	}
	if after := gitOut(t, dir, "rev-parse", "HEAD"); after != before {
		t.Errorf("a clean tree moved HEAD from %s to %s", before, after)
	}

	if err := os.WriteFile(filepath.Join(dir, "note.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	committed, err = project.CommitAll(dir, "chore(SPEC-020): checkpoint implementing")
	if err != nil {
		t.Fatal(err)
	}
	if !committed {
		t.Fatal("a dirty tree must commit")
	}
	if got := gitOut(t, dir, "log", "-1", "--pretty=%s"); got != "chore(SPEC-020): checkpoint implementing" {
		t.Errorf("commit subject = %q", got)
	}
	if got := gitOut(t, dir, "status", "--porcelain"); got != "" {
		t.Errorf("the tree should be clean after a commit: %q", got)
	}
}

// A branch with no upstream is unpushed; a branch equal to its upstream is
// not; a new commit on top of the upstream is.
func TestHasUnpushed(t *testing.T) {
	dir := gitRepo(t, "spec/020-push-progress")

	if got, err := project.HasUnpushed(dir); err != nil || !got {
		t.Fatalf("no upstream: got %v, %v; want true, nil", got, err)
	}

	runGit(t, dir, "push", "-u", "origin", "spec/020-push-progress")
	if got, err := project.HasUnpushed(dir); err != nil || got {
		t.Fatalf("equal to upstream: got %v, %v; want false, nil", got, err)
	}

	if err := os.WriteFile(filepath.Join(dir, "more.txt"), []byte("more\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-m", "more")
	if got, err := project.HasUnpushed(dir); err != nil || !got {
		t.Fatalf("ahead of upstream: got %v, %v; want true, nil", got, err)
	}
}

// main and master are the default branch, whatever the case; a spec branch and
// a directory outside a repository are not.
func TestOnDefaultBranch(t *testing.T) {
	for _, branch := range []string{"main", "master", "MAIN"} {
		dir := gitRepo(t, branch)
		if !project.OnDefaultBranch(dir) {
			t.Errorf("%s should be the default branch", branch)
		}
	}
	dir := gitRepo(t, "spec/020-push-progress")
	if project.OnDefaultBranch(dir) {
		t.Error("a spec branch is not the default branch")
	}
	if project.OnDefaultBranch(t.TempDir()) {
		t.Error("outside a repository there is no default branch")
	}
}

// The opt-in is on/true/yes, case insensitive; absent and off leave it off.
func TestPushEnabled(t *testing.T) {
	for _, tc := range []struct {
		val  string
		want bool
	}{
		{"on", true},
		{"true", true},
		{"yes", true},
		{"ON", true},
		{"off", false},
		{"false", false},
		{"", false},
	} {
		front := "---\ntest: \"go test ./...\"\n"
		if tc.val != "" {
			front += "push: " + tc.val + "\n"
		}
		root := write(t, front+"---\n\n# Project\n", nil)
		p, err := project.Load(root)
		if err != nil {
			t.Fatal(err)
		}
		if got := p.PushEnabled(); got != tc.want {
			t.Errorf("push: %q -> %v, want %v", tc.val, got, tc.want)
		}
	}
}

// A GitHub remote is recognised from the URL git already has: the https and
// ssh forms both count, while a local bare path or no remote does not. Nothing
// is fetched, so no test touches the network (SPEC-022, decision 2).
func TestHasGitHubRemote(t *testing.T) {
	dir := gitRepo(t, "main")
	if project.HasGitHubRemote(dir) {
		t.Error("a local bare remote is not GitHub")
	}
	for _, url := range []string{
		"https://github.com/x/y.git",
		"git@github.com:x/y.git",
	} {
		runGit(t, dir, "remote", "set-url", "origin", url)
		if !project.HasGitHubRemote(dir) {
			t.Errorf("origin at %s should be recognised as GitHub", url)
		}
	}
	if project.HasGitHubRemote(t.TempDir()) {
		t.Error("a directory outside a repository has no remote")
	}
}
