package kit

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// machineRoot is the embedded tree of machinery: the workflow, the roles and
// the templates. It is never planted in a repository; it is served by the
// binary through `forge workflow`, `forge roles` and `forge template`.
const machineRoot = "machine"

// Workflow returns the process description an agent reads: machine/WORKFLOW.md.
func Workflow() ([]byte, error) {
	return FS.ReadFile(machineRoot + "/WORKFLOW.md")
}

// Roles lists the role names, sorted.
func Roles() []string {
	return list("roles")
}

// Role returns one role's instructions by name.
func Role(name string) ([]byte, error) {
	return read("roles", name)
}

// Template returns one file template, the same copy `forge new` uses.
func Template(name string) ([]byte, error) {
	return read("templates", name)
}

// read resolves a name against the files known to exist in dir, so a name
// that escapes the tree or does not exist fails instead of reading anything.
func read(dir, name string) ([]byte, error) {
	name = strings.TrimSuffix(strings.TrimSpace(name), ".md")
	known := list(dir)
	if name == "" {
		return nil, fmt.Errorf("which %s? choose one of %s", kind(dir), strings.Join(known, ", "))
	}
	for _, k := range known {
		if name == k {
			return FS.ReadFile(machineRoot + "/" + dir + "/" + k + ".md")
		}
	}
	return nil, fmt.Errorf("unknown %s %q; choose one of %s", kind(dir), name, strings.Join(known, ", "))
}

func list(dir string) []string {
	entries, err := fs.ReadDir(FS, machineRoot+"/"+dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".md"))
	}
	sort.Strings(names)
	return names
}

func kind(dir string) string {
	return strings.TrimSuffix(dir, "s")
}
