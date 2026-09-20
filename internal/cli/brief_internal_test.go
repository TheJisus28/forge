package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/TheJisus28/forge/internal/project"
)

// briefRepo builds a repository wired to a local bare origin, with a
// .forge/project.md, one local spec, and a commit on origin the working clone
// has never fetched. fetchOn adds the `fetch: on` opt-in. It returns the
// working directory and the remote-only commit, so a fetch is observable as
// `origin/main` advancing. The remote is a directory, so no test touches the
// network (SPEC-022, the Risks note).
func briefRepo(t *testing.T, fetchOn bool) (dir, remoteOnly string) {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "origin.git")
	submitRunGit(t, filepath.Dir(remote), "init", "--bare", remote)

	dir = t.TempDir()
	submitRunGit(t, dir, "init")
	submitRunGit(t, dir, "config", "user.email", "t@example.com")
	submitRunGit(t, dir, "config", "user.name", "tester")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("root\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	submitRunGit(t, dir, "add", "-A")
	submitRunGit(t, dir, "commit", "-m", "root")
	submitRunGit(t, dir, "branch", "-M", "main")
	submitRunGit(t, dir, "remote", "add", "origin", remote)
	submitRunGit(t, dir, "push", "-u", "origin", "main")

	// A second clone advances origin/main with a commit the working clone has
	// never seen, so only a fetch can move its origin/main ref.
	other := filepath.Join(t.TempDir(), "other")
	submitRunGit(t, filepath.Dir(other), "clone", "-b", "main", remote, other)
	submitRunGit(t, other, "config", "user.email", "t@example.com")
	submitRunGit(t, other, "config", "user.name", "tester")
	if err := os.WriteFile(filepath.Join(other, "remote.txt"), []byte("remote\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	submitRunGit(t, other, "add", "-A")
	submitRunGit(t, other, "commit", "-m", "remote-only")
	submitRunGit(t, other, "push", "origin", "main")
	remoteOnly = submitGitOut(t, other, "rev-parse", "HEAD")

	config := "---\ntest: \"go test ./...\"\nworking_language: en\nguard: on\n"
	if fetchOn {
		config += "fetch: on\n"
	}
	config += "---\n\n# Project\n\nA demo.\n"
	if err := os.MkdirAll(filepath.Join(dir, ".forge", "specs", "SPEC-001-local"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".forge", "project.md"), []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := "---\nid: SPEC-001\ntitle: Local spec\nstatus: proposed\ncapability: workflow\n---\n\n## Problem\n\nA local spec.\n"
	if err := os.WriteFile(filepath.Join(dir, ".forge", "specs", "SPEC-001-local", "spec.md"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, remoteOnly
}

// ageFetchHead makes the clone look like it last fetched daysAgo, so the
// brief's staleness note is exercised before a fetch clears it (SPEC-022,
// AC1). A fetch overwrites FETCH_HEAD, which is what makes the change visible.
func ageFetchHead(t *testing.T, dir string, daysAgo int) {
	t.Helper()
	path := filepath.Join(gitDir(t, dir), "FETCH_HEAD")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().AddDate(0, 0, -daysAgo)
	if err := os.Chtimes(path, stamp, stamp); err != nil {
		t.Fatal(err)
	}
}

// fetchHeadBytes snapshots FETCH_HEAD, or nil when it does not exist, so a
// before/after comparison proves whether a fetch ran.
func fetchHeadBytes(t *testing.T, dir string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(gitDir(t, dir), "FETCH_HEAD"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// gitDir resolves the clone's .git directory, which `git rev-parse` prints
// relative to the working directory.
func gitDir(t *testing.T, dir string) string {
	t.Helper()
	gd := submitGitOut(t, dir, "rev-parse", "--git-dir")
	if !filepath.IsAbs(gd) {
		gd = filepath.Join(dir, gd)
	}
	return gd
}

// briefCmd runs cmdBrief inside dir and returns its output and error, so a
// test can assert the command always exits 0 (SPEC-022, decision 5).
func briefCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	var out bytes.Buffer
	err = cmdBrief(args, &out)
	return out.String(), err
}

// With `fetch: on`, a GitHub remote and an authenticated gh, the brief fetches
// before it renders: origin/main advances to the remote-only commit and the
// staleness note is gone, while the local spec list is untouched (SPEC-022,
// AC1).
func TestBrief_FetchesWhenOptedIn(t *testing.T) {
	dir, remoteOnly := briefRepo(t, true)
	ageFetchHead(t, dir, 3)

	stale := submitGitOut(t, dir, "rev-parse", "origin/main")
	if stale == remoteOnly {
		t.Fatal("the fixture should start behind origin/main")
	}
	if age, ok := project.FetchAge(dir); !ok || age <= 48*time.Hour {
		t.Fatalf("the fixture should start stale: age %v, ok %v", age, ok)
	}

	swap(t, &githubRemote, func(string) bool { return true })
	swap(t, &ghUser, func(string) string { return "tester" })

	out, err := briefCmd(t, dir)
	if err != nil {
		t.Fatalf("brief failed: %v", err)
	}
	if got := submitGitOut(t, dir, "rev-parse", "origin/main"); got != remoteOnly {
		t.Errorf("origin/main = %s, want the remote-only commit %s", got, remoteOnly)
	}
	if strings.Contains(out, "last fetched") {
		t.Errorf("a successful fetch should clear the staleness note:\n%s", out)
	}
	if !strings.Contains(out, "SPEC-001") {
		t.Errorf("the brief should still list the local spec:\n%s", out)
	}
	if strings.Contains(out, "warning:") {
		t.Errorf("a successful fetch should not warn:\n%s", out)
	}
}

// Without the scalar the brief performs no gh call and no fetch: the seams
// fail the test if touched, and origin/main and FETCH_HEAD are byte-identical
// before and after (SPEC-022, AC2).
func TestBrief_OfflineByDefault(t *testing.T) {
	dir, _ := briefRepo(t, false)
	ageFetchHead(t, dir, 3)

	beforeOrigin := submitGitOut(t, dir, "rev-parse", "origin/main")
	beforeFetchHead := fetchHeadBytes(t, dir)

	swap(t, &githubRemote, func(string) bool {
		t.Error("githubRemote must not be called without fetch: on")
		return false
	})
	swap(t, &ghUser, func(string) string {
		t.Error("ghUser must not be called without fetch: on")
		return ""
	})

	out, err := briefCmd(t, dir)
	if err != nil {
		t.Fatalf("brief failed: %v", err)
	}
	if got := submitGitOut(t, dir, "rev-parse", "origin/main"); got != beforeOrigin {
		t.Errorf("origin/main moved from %s to %s without fetch: on", beforeOrigin, got)
	}
	if got := fetchHeadBytes(t, dir); !bytes.Equal(got, beforeFetchHead) {
		t.Error("FETCH_HEAD changed without fetch: on")
	}
	if strings.Contains(out, "warning:") {
		t.Errorf("the default is offline and silent:\n%s", out)
	}
}

// A fetch moves refs, never the working tree: with a dirty tree, the status,
// HEAD and branch are identical before and after, and no output suggests a
// pull, merge or rebase (SPEC-022, AC3).
func TestBrief_FetchDoesNotTouchWorkingTree(t *testing.T) {
	dir, _ := briefRepo(t, true)
	ageFetchHead(t, dir, 3)
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("uncommitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	beforeStatus := submitGitOut(t, dir, "status", "--porcelain")
	beforeHead := submitGitOut(t, dir, "rev-parse", "HEAD")
	beforeBranch := submitGitOut(t, dir, "rev-parse", "--abbrev-ref", "HEAD")

	swap(t, &githubRemote, func(string) bool { return true })
	swap(t, &ghUser, func(string) string { return "tester" })

	out, err := briefCmd(t, dir)
	if err != nil {
		t.Fatalf("brief failed: %v", err)
	}
	if got := submitGitOut(t, dir, "status", "--porcelain"); got != beforeStatus {
		t.Errorf("working tree changed:\nbefore:\n%s\nafter:\n%s", beforeStatus, got)
	}
	if got := submitGitOut(t, dir, "rev-parse", "HEAD"); got != beforeHead {
		t.Errorf("HEAD moved from %s to %s", beforeHead, got)
	}
	if got := submitGitOut(t, dir, "rev-parse", "--abbrev-ref", "HEAD"); got != beforeBranch {
		t.Errorf("branch changed from %s to %s", beforeBranch, got)
	}
	for _, notice := range []string{"pull", "merge", "rebase"} {
		if strings.Contains(strings.ToLower(out), notice) {
			t.Errorf("the output carries a %q notice:\n%s", notice, out)
		}
	}
}

// With no GitHub remote the fetch is skipped, gh is not called and the brief
// still renders with a warning; an unauthenticated gh warns the same way
// (SPEC-022, AC4).
func TestBrief_SkipsWithoutGitHub(t *testing.T) {
	dir, _ := briefRepo(t, true)
	ageFetchHead(t, dir, 3)
	beforeOrigin := submitGitOut(t, dir, "rev-parse", "origin/main")
	beforeFetchHead := fetchHeadBytes(t, dir)

	swap(t, &githubRemote, func(string) bool { return false })
	swap(t, &ghUser, func(string) string {
		t.Error("ghUser must not be called when there is no GitHub remote")
		return ""
	})

	out, err := briefCmd(t, dir)
	if err != nil {
		t.Fatalf("brief failed: %v", err)
	}
	if !strings.HasPrefix(out, "warning: no GitHub remote") {
		t.Errorf("the output should start with the no-GitHub warning:\n%s", out)
	}
	if !strings.Contains(out, "SPEC-001") {
		t.Errorf("the brief should still render:\n%s", out)
	}
	if got := submitGitOut(t, dir, "rev-parse", "origin/main"); got != beforeOrigin {
		t.Errorf("origin/main moved without a fetch: %s -> %s", beforeOrigin, got)
	}
	if got := fetchHeadBytes(t, dir); !bytes.Equal(got, beforeFetchHead) {
		t.Error("FETCH_HEAD changed without a fetch")
	}

	// The same skip with an authenticated remote but no gh login.
	dir2, _ := briefRepo(t, true)
	ageFetchHead(t, dir2, 3)
	swap(t, &githubRemote, func(string) bool { return true })
	swap(t, &ghUser, func(string) string { return "" })

	out2, err := briefCmd(t, dir2)
	if err != nil {
		t.Fatalf("brief failed: %v", err)
	}
	if !strings.HasPrefix(out2, "warning: gh is not authenticated") {
		t.Errorf("the output should start with the gh warning:\n%s", out2)
	}
}

// A failed fetch warns, still renders the brief and returns nil; the warning
// is prepended to the one text, so the --json payload stays a single line
// (SPEC-022, AC5 and decision 6).
func TestBrief_FetchFailureIsAWarning(t *testing.T) {
	dir, _ := briefRepo(t, true)
	ageFetchHead(t, dir, 3)
	// A missing remote makes git fetch fail without any network.
	submitRunGit(t, dir, "remote", "set-url", "origin", filepath.Join(dir, "missing.git"))

	swap(t, &githubRemote, func(string) bool { return true })
	swap(t, &ghUser, func(string) string { return "tester" })

	out, err := briefCmd(t, dir)
	if err != nil {
		t.Fatalf("a failed fetch should return nil: %v", err)
	}
	if !strings.HasPrefix(out, "warning: ") {
		t.Errorf("the output should start with the warning:\n%s", out)
	}
	if !strings.Contains(out, "git fetch failed") {
		t.Errorf("the warning should name the failed fetch:\n%s", out)
	}
	if !strings.Contains(out, "Rules that matter here") {
		t.Errorf("the brief should still render after a failed fetch:\n%s", out)
	}

	payload, err := briefCmd(t, dir, "--json")
	if err != nil {
		t.Fatalf("--json should also return nil: %v", err)
	}
	if strings.Contains(strings.TrimRight(payload, "\n"), "\n") {
		t.Errorf("the --json payload must stay one line:\n%s", payload)
	}
	if !strings.Contains(payload, `warning: git fetch failed`) {
		t.Errorf("the --json payload should carry the warning:\n%s", payload)
	}
}
