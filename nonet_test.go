package main

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// Forge is local-first: it reads and writes files, and nothing else. The
// network belongs to git and gh, which the user or CI invokes explicitly.
// This test is the promise, enforced.
func TestBinaryMakesNoNetworkCalls(t *testing.T) {
	banned := map[string]string{
		"net":                 "raw sockets",
		"net/http":            "HTTP",
		"net/rpc":             "RPC",
		"net/smtp":            "email",
		"crypto/tls":          "TLS connections",
		"golang.org/x/net":    "networking",
		"github.com/go-resty": "HTTP clients",
	}
	fset := token.NewFileSet()
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "kit" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			name := strings.Trim(imp.Path.Value, `"`)
			for prefix, what := range banned {
				if name == prefix || strings.HasPrefix(name, prefix+"/") {
					t.Errorf("%s imports %s (%s); Forge must not reach the network itself",
						path, name, what)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
