package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func submitRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func submitGitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
}

// submit commits the pending work before it pushes, so the pull request never
// misses an uncommitted change (SPEC-020, decision 8). The gh and push seams
// are faked, so the test touches no network.
func TestSubmit_CommitsBeforePush(t *testing.T) {
	dir := t.TempDir()
	submitRunGit(t, dir, "init")
	submitRunGit(t, dir, "config", "user.email", "t@example.com")
	submitRunGit(t, dir, "config", "user.name", "tester")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	submitRunGit(t, dir, "add", "-A")
	submitRunGit(t, dir, "commit", "-m", "root")
	submitRunGit(t, dir, "branch", "-M", "main")

	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	var out, errOut bytes.Buffer
	if code := Main([]string{"init"}, &out, &errOut); code != 0 {
		t.Fatalf("init failed: %s", errOut.String())
	}
	if err := os.WriteFile(filepath.Join(dir, ".forge", "project.md"),
		[]byte("---\ntest: \"go test ./...\"\nguard: on\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errOut.Reset()
	if code := Main([]string{"new", "Something", "--capability", "workflow"}, &out, &errOut); code != 0 {
		t.Fatalf("new failed: %s", errOut.String())
	}
	submitRunGit(t, dir, "checkout", "-b", "spec/001-something")

	if err := os.WriteFile(filepath.Join(dir, ".forge", "note.md"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var pushed, statusAtPush string
	swap(t, &hasGH, func() bool { return true })
	swap(t, &pushBranch, func(root, branch string) error {
		pushed = branch
		statusAtPush = submitGitOut(t, root, "status", "--porcelain")
		return nil
	})
	swap(t, &ghRun, func(root string, args ...string) (string, error) {
		return "https://example.test/pull/1", nil
	})

	out.Reset()
	if err := cmdSubmit(nil, &out); err != nil {
		t.Fatalf("submit failed: %v", err)
	}
	if got := submitGitOut(t, dir, "log", "-1", "--pretty=%s"); got != "chore(SPEC-001): checkpoint proposed" {
		t.Errorf("commit subject = %q", got)
	}
	if pushed != "spec/001-something" {
		t.Errorf("pushed branch = %q", pushed)
	}
	if statusAtPush != "" {
		t.Errorf("the tree should be clean when submit pushes: %q", statusAtPush)
	}
}
