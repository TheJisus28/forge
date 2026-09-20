package cli_test

import (
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
