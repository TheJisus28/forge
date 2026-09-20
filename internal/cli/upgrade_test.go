package cli_test

import (
	"os"
	"strings"
	"testing"
)

func TestHelp_ListsUpgrade(t *testing.T) {
	dir := t.TempDir()
	out := mustRun(t, dir, "help")
	if !strings.Contains(out, "forge upgrade") {
		t.Fatalf("help should list forge upgrade:\n%s", out)
	}
}

func TestDocs_DocumentUpgrade(t *testing.T) {
	doc, err := os.ReadFile("../../docs/cli.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(doc), "forge upgrade") {
		t.Fatal("docs/cli.md should document forge upgrade")
	}
}
