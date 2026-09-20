package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// UserName returns the configured git user, used as a default conductor.
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
