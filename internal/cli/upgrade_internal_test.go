package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// swap replaces a package-level seam for the duration of one test.
func swap[T any](t *testing.T, target *T, value T) {
	t.Helper()
	orig := *target
	*target = value
	t.Cleanup(func() { *target = orig })
}

// installCall records what the fake toolchain was asked to do.
type installCall struct {
	goBin   string
	destDir string
	ref     string
}

// stubToolchain replaces every seam the command touches with a fake that
// records the call and writes a dummy binary, so no test runs a real
// `go install` and none touches the network.
func stubToolchain(t *testing.T, version string, versionErr error) *installCall {
	t.Helper()
	call := &installCall{}
	swap(t, &lookupGo, func(file string) (string, error) {
		if file != "go" {
			t.Errorf("looked up %q, want go", file)
		}
		return "/fake/bin/go", nil
	})
	swap(t, &executable, func() (string, error) {
		exe := filepath.Join(t.TempDir(), "forge"+exeSuffix())
		if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
			t.Fatal(err)
		}
		return exe, nil
	})
	swap(t, &goInstall, func(goBin, destDir, ref string, stdout, stderr io.Writer) error {
		call.goBin, call.destDir, call.ref = goBin, destDir, ref
		return os.WriteFile(filepath.Join(destDir, "forge"+exeSuffix()), []byte("fresh"), 0o755)
	})
	swap(t, &goVersion, func(goBin, path string) (string, error) {
		return version, versionErr
	})
	return call
}

func TestUpgrade_InstallsLatestIntoATempDir(t *testing.T) {
	call := stubToolchain(t, "v0.3.0", nil)
	exe, err := executable()
	if err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	if err := cmdUpgrade(nil, &out, &errOut); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if call.ref != "latest" {
		t.Errorf("ref = %q, want latest", call.ref)
	}
	if got := filepath.Dir(exe); got == call.destDir {
		t.Errorf("installed into the executable's own directory %s", got)
	}
	if !strings.Contains(out.String(), "installing github.com/TheJisus28/forge@latest") {
		t.Errorf("output does not announce the install:\n%s", out.String())
	}
}

func TestUpgrade_UsesTheGivenVersionVerbatim(t *testing.T) {
	call := stubToolchain(t, "v0.2.0", nil)

	var out, errOut bytes.Buffer
	if err := cmdUpgrade([]string{"v0.2.0"}, &out, &errOut); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if call.ref != "v0.2.0" {
		t.Errorf("ref = %q, want v0.2.0", call.ref)
	}
}

func TestUpgrade_EmptyArgumentMeansLatest(t *testing.T) {
	call := stubToolchain(t, "", nil)

	var out, errOut bytes.Buffer
	if err := cmdUpgrade([]string{""}, &out, &errOut); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if call.ref != "latest" {
		t.Errorf("ref = %q, want latest", call.ref)
	}
}

func TestUpgrade_WithoutGoExplainsHowToInstallIt(t *testing.T) {
	swap(t, &lookupGo, func(string) (string, error) { return "", exec.ErrNotFound })
	installed := false
	swap(t, &goInstall, func(goBin, destDir, ref string, stdout, stderr io.Writer) error {
		installed = true
		return nil
	})

	var out, errOut bytes.Buffer
	err := cmdUpgrade(nil, &out, &errOut)
	if err == nil {
		t.Fatal("upgrade without go should fail")
	}
	if !strings.Contains(err.Error(), "Go 1.22+") {
		t.Errorf("error does not tell the user to install Go:\n%v", err)
	}
	if !strings.Contains(err.Error(), "go install github.com/TheJisus28/forge@latest") {
		t.Errorf("error does not give the manual command:\n%v", err)
	}
	if installed {
		t.Error("go install must not run when go is not on PATH")
	}
}

func TestUpgrade_PrintsOldAndNewVersions(t *testing.T) {
	swap(t, &Version, "dev")
	stubToolchain(t, "v0.3.0", nil)

	var out, errOut bytes.Buffer
	if err := cmdUpgrade(nil, &out, &errOut); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "dev") {
		t.Errorf("output does not show the version being replaced:\n%s", got)
	}
	if !strings.Contains(got, "upgraded to v0.3.0") {
		t.Errorf("output does not show the installed version:\n%s", got)
	}
}

func TestUpgrade_FallsBackToTheRefWhenVersionIsUnreadable(t *testing.T) {
	swap(t, &Version, "dev")
	stubToolchain(t, "", errors.New("cannot read the build info"))

	var out, errOut bytes.Buffer
	if err := cmdUpgrade(nil, &out, &errOut); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	if !strings.Contains(out.String(), "upgraded to latest") {
		t.Errorf("output should fall back to the requested ref:\n%s", out.String())
	}
}

func TestUpgrade_ReplacesTheRunningBinary(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "forge"+exeSuffix())
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	swap(t, &lookupGo, func(string) (string, error) { return "/fake/bin/go", nil })
	swap(t, &executable, func() (string, error) { return exe, nil })
	swap(t, &goInstall, func(goBin, destDir, ref string, stdout, stderr io.Writer) error {
		return os.WriteFile(filepath.Join(destDir, "forge"+exeSuffix()), []byte("new"), 0o755)
	})
	swap(t, &goVersion, func(goBin, path string) (string, error) { return "v9.9.9", nil })

	var out, errOut bytes.Buffer
	if err := cmdUpgrade(nil, &out, &errOut); err != nil {
		t.Fatalf("upgrade failed: %v", err)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("running binary = %q, want the freshly installed one", got)
	}
	if !strings.Contains(out.String(), "upgraded to v9.9.9") {
		t.Errorf("output does not report the replacement:\n%s", out.String())
	}
}

func TestUpgrade_WhenRenameFailsTheBinaryIsUnchanged(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "forge"+exeSuffix())
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	swap(t, &lookupGo, func(string) (string, error) { return "/fake/bin/go", nil })
	swap(t, &executable, func() (string, error) { return exe, nil })
	swap(t, &goInstall, func(goBin, destDir, ref string, stdout, stderr io.Writer) error {
		return os.WriteFile(filepath.Join(destDir, "forge"+exeSuffix()), []byte("new"), 0o755)
	})
	swap(t, &renameFile, func(string, string) error { return errors.New("access is denied") })

	var out, errOut bytes.Buffer
	err := cmdUpgrade(nil, &out, &errOut)
	if err == nil {
		t.Fatal("a failed rename should fail the upgrade")
	}
	if !strings.Contains(err.Error(), "cannot replace the running forge binary") {
		t.Errorf("error is not the actionable rename sentence:\n%v", err)
	}
	if !strings.Contains(err.Error(), "go install github.com/TheJisus28/forge@latest") {
		t.Errorf("error does not give the manual command:\n%v", err)
	}
	got, readErr := os.ReadFile(exe)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "old" {
		t.Errorf("running binary = %q, want it left unchanged", got)
	}
}

func TestCleanupStaleBinary_RemovesTheSidecar(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "forge"+exeSuffix())
	old := exe + ".old"
	if err := os.WriteFile(old, []byte("stale"), 0o755); err != nil {
		t.Fatal(err)
	}
	swap(t, &executable, func() (string, error) { return exe, nil })

	cleanupStaleBinary()
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("sidecar should be gone, stat err = %v", err)
	}
}

func TestCleanupStaleBinary_IgnoresRemoveErrors(t *testing.T) {
	swap(t, &executable, func() (string, error) { return "/does/not/matter/forge", nil })
	swap(t, &removeFile, func(string) error { return errors.New("still locked by the running process") })

	cleanupStaleBinary()
}

func TestUpgrade_OutsideAProjectCreatesNoForgeDirectory(t *testing.T) {
	stubToolchain(t, "v0.3.0", nil)

	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatal(err)
		}
	})

	var out, errOut bytes.Buffer
	if code := Main([]string{"upgrade"}, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr:\n%s", code, errOut.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".forge")); !os.IsNotExist(err) {
		t.Errorf("upgrade must not create .forge/, stat err = %v", err)
	}
}

func TestUpgrade_RejectsMoreThanOneVersion(t *testing.T) {
	swap(t, &lookupGo, func(string) (string, error) {
		t.Error("argument validation must run before looking up go")
		return "", nil
	})

	var out, errOut bytes.Buffer
	err := cmdUpgrade([]string{"v0.1.0", "v0.2.0"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "at most one version") {
		t.Fatalf("err = %v, want the too-many-arguments sentence", err)
	}
}
