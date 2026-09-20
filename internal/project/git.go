package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Branch returns the current git branch, or "" outside a repository.
func Branch(root string) string {
	out, err := run(root, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	return out
}

var branchIDRe = regexp.MustCompile(`(?i)(?:^|/)spec[-/](\d+)`)

// SpecIDFromBranch reads the spec ID out of a branch like spec/004-slug.
func SpecIDFromBranch(branch string) string {
	if m := branchIDRe.FindStringSubmatch(branch); m != nil {
		return NormalizeID(m[1])
	}
	return ""
}

// RemoteSpecIDs returns the spec ids committed under `.forge/specs` on a git
// ref, or nothing when that ref does not exist. Nothing is fetched: only refs
// already present are read, because the network belongs to git and fetching is
// the user's call (SPEC-015, decision 7).
func RemoteSpecIDs(root, ref string) []string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	out, err := run(root, "git", "ls-tree", "-d", "--name-only", ref, Dir+"/specs/")
	if err != nil {
		return nil
	}
	var ids []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// A folder is `SPEC-NNN-slug`; the id is the first two dash-separated
		// parts, normalised the same way every other id is.
		parts := strings.SplitN(filepath.Base(line), "-", 3)
		if len(parts) < 2 {
			continue
		}
		id := NormalizeID(parts[0] + "-" + parts[1])
		if !idRe.MatchString(id) {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

// BranchName is the branch Forge suggests for a spec.
func BranchName(id, title string) string {
	num := strings.TrimPrefix(id, "SPEC-")
	return "spec/" + num + "-" + Slug(title)
}

// Current returns the spec the current branch is about, if any.
func (p *Project) Current() (*Spec, bool) {
	if id := SpecIDFromBranch(Branch(p.Root)); id != "" {
		return p.Spec(id)
	}
	return nil, false
}

// FetchAge reports how long ago the repository last fetched from a remote.
// The second result is false when that cannot be determined.
func FetchAge(root string) (time.Duration, bool) {
	gitDir, err := run(root, "git", "rev-parse", "--git-dir")
	if err != nil {
		return 0, false
	}
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(root, gitDir)
	}
	newest := time.Time{}
	for _, name := range []string{"FETCH_HEAD", "refs/remotes/origin/HEAD"} {
		st, err := os.Stat(filepath.Join(gitDir, name))
		if err != nil {
			continue
		}
		if st.ModTime().After(newest) {
			newest = st.ModTime()
		}
	}
	if newest.IsZero() {
		return 0, false
	}
	return time.Since(newest), true
}

// Fetch updates the remote refs. The network belongs to git, never to Forge.
func Fetch(root string) error {
	cmd := exec.Command("git", "fetch", "--quiet")
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// UserName returns the configured git user, used as a default orchestrator.
func UserName(root string) string {
	name, err := run(root, "git", "config", "user.name")
	if err != nil {
		return ""
	}
	return name
}

// GHUser returns the login of the authenticated gh user, or "" when gh is
// missing or not authenticated. The network belongs to gh, never to Forge.
func GHUser(root string) string {
	if !HasGH() {
		return ""
	}
	out, err := run(root, "gh", "api", "user", "--jq", ".login")
	if err != nil {
		return ""
	}
	return out
}

// Push publishes a branch to origin, setting its upstream. The network
// belongs to git.
func Push(root, branch string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// CommitAll stages the whole tree and commits it with message, returning
// whether it committed. A clean tree is checked first, so "nothing to commit"
// is a false and never an error (SPEC-020, decision 2). The tree is staged
// whole because a spec branch is single-purpose.
func CommitAll(root, message string) (bool, error) {
	status, err := run(root, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	if status == "" {
		return false, nil
	}
	if _, err := run(root, "git", "add", "-A"); err != nil {
		return false, err
	}
	if _, err := run(root, "git", "commit", "-m", message); err != nil {
		return false, err
	}
	return true, nil
}

// HasUnpushed reports whether the current branch has commits origin does not.
// A branch with no upstream has never been published, so it reports true rather
// than an error: that is what makes `forge push` publish a new branch (SPEC-020,
// decision 2 as corrected in tasks.md).
func HasUnpushed(root string) (bool, error) {
	if _, err := run(root, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}"); err != nil {
		return true, nil
	}
	out, err := run(root, "git", "rev-list", "--count", "@{upstream}..HEAD")
	if err != nil {
		return false, err
	}
	ahead, err := strconv.Atoi(out)
	if err != nil {
		return false, err
	}
	return ahead > 0, nil
}

// OnDefaultBranch reports whether the current branch is main or master. A
// detached HEAD or a branch outside a repository is not the default branch.
func OnDefaultBranch(root string) bool {
	switch strings.ToLower(strings.TrimSpace(Branch(root))) {
	case "main", "master":
		return true
	}
	return false
}

// HasGH reports whether the GitHub CLI is available for optional sync.
func HasGH() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// GH runs the GitHub CLI and returns its trimmed output.
func GH(root string, args ...string) (string, error) { return run(root, "gh", args...) }

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
