package kit_test

import (
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJisus28/forge/kit"
)

// A new top-level folder under kit/ needs its own pattern in the go:embed
// line, otherwise it silently never reaches the user's repository.
func TestFS_EmbedsEveryFileOnDisk(t *testing.T) {
	embedded := map[string]bool{}
	err := fs.WalkDir(kit.FS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			embedded[p] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(embedded) == 0 {
		t.Fatal("kit.FS is empty")
	}

	err = filepath.WalkDir(".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasSuffix(p, ".go") {
			return nil
		}
		if rel := filepath.ToSlash(p); !embedded[rel] {
			t.Errorf("file on disk is not embedded: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
