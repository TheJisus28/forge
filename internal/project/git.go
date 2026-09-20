package project

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

// SpecRef is one spec folder on a git ref: its id and the `SPEC-NNN-slug`
// folder that carries it. The folder is the identity — the same id under a
// different folder is a different spec (SPEC-023, decision 1).
type SpecRef struct {
	ID  string
	Dir string
}

// RemoteRefs returns every remote-tracking ref, or nothing outside a
// repository. One git call reads the refs git already has; nothing is fetched
// (SPEC-023, decision 2).
func RemoteRefs(root string) []string {
	out, err := run(root, "git", "for-each-ref", "--format=%(refname)", "refs/remotes/")
	if err != nil {
		return nil
	}
	var refs []string
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			refs = append(refs, line)
		}
	}
	sort.Strings(refs)
	return refs
}

// RemoteSpecDirs returns the `SPEC-NNN-slug` folders under `.forge/specs/` on
// a git ref, deduped and sorted. A ref that is not there, or carries no spec,
// yields nothing. Nothing is fetched: only refs already present are read
// (SPEC-023, decision 2).
func RemoteSpecDirs(root, ref string) []string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	out, err := run(root, "git", "ls-tree", "-d", "--name-only", ref, Dir+"/specs/")
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var dirs []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		base := filepath.Base(line)
		if dirID(base) == "" || seen[base] {
			continue
		}
		seen[base] = true
		dirs = append(dirs, base)
	}
	sort.Strings(dirs)
	return dirs
}

// dirID reads the normalised id out of a `SPEC-NNN-slug` folder name, or "".
func dirID(dir string) string {
	parts := strings.SplitN(dir, "-", 3)
	if len(parts) < 2 {
		return ""
	}
	id := NormalizeID(parts[0] + "-" + parts[1])
	if !idRe.MatchString(id) {
		return ""
	}
	return id
}

// RemoteSpecIDs returns the spec ids committed under `.forge/specs` on a git
// ref, or nothing when that ref does not exist. It is the id view over
// RemoteSpecDirs; nothing is fetched (SPEC-015, decision 7).
func RemoteSpecIDs(root, ref string) []string {
	var ids []string
	seen := map[string]bool{}
	for _, dir := range RemoteSpecDirs(root, ref) {
		id := dirID(dir)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids
}

// RemoteSpecRefs returns every spec folder on every remote-tracking ref,
// deduped by folder. It never falls back to a local branch: `forge new` uses
// it so a branch keeps minting its provisional number while `forge accept`
// does the confirming (SPEC-023, decision 1). Nothing is fetched.
func RemoteSpecRefs(root string) []SpecRef {
	return collectSpecRefs(root, RemoteRefs(root))
}

// SharedSpecRefs returns the spec folders the id confirmation reads: every
// remote-tracking ref, or the local `main` ref when no remote-tracking ref
// carries a spec, so a repository with no remote keeps the SPEC-015
// behaviour. A missing ref yields nothing and never fails. Nothing is fetched
// (SPEC-023, decisions 2, 3 and 8).
func SharedSpecRefs(root string) []SpecRef {
	out := RemoteSpecRefs(root)
	if len(out) == 0 {
		out = collectSpecRefs(root, []string{"main"})
	}
	return out
}

// collectSpecRefs unions the spec folders on refs, deduped by folder and
// sorted.
func collectSpecRefs(root string, refs []string) []SpecRef {
	seen := map[string]bool{}
	var out []SpecRef
	for _, ref := range refs {
		for _, dir := range RemoteSpecDirs(root, ref) {
			if seen[dir] {
				continue
			}
			seen[dir] = true
			out = append(out, SpecRef{ID: dirID(dir), Dir: dir})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out
}

// HasGitHubRemote reports whether any configured remote points at GitHub. It
// reads only the remote list git already has, so it touches no network
// (SPEC-022, decision 2).
func HasGitHubRemote(root string) bool {
	out, err := run(root, "git", "remote")
	if err != nil {
		return false
	}
	for _, name := range strings.Split(out, "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		url, err := run(root, "git", "remote", "get-url", name)
		if err != nil {
			continue
		}
		if strings.Contains(url, "github.com") {
			return true
		}
	}
	return false
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
