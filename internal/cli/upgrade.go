package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// forgeModule is the module the running binary was installed from, so the
// toolchain knows exactly what to fetch.
const forgeModule = "github.com/TheJisus28/forge"

// Test seams: the OS and toolchain calls are variables so tests can run the
// command without a real Go toolchain and without touching the network.
var (
	lookupGo   = exec.LookPath
	goInstall  = installModule
	goVersion  = moduleVersion
	executable = os.Executable
)

// cmdUpgrade installs a released forge into a private temp GOBIN and reports
// the version it is replacing and the version it installed. It never reads or
// writes a .forge/ project: refreshing the kit stays `forge update`.
func cmdUpgrade(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("upgrade", "usage: forge upgrade [version]", stdout)
	positional, err := parseArgs(fs, args)
	if err != nil {
		return err
	}
	var ref string
	switch {
	case len(positional) == 0:
		ref = "latest"
	case len(positional) == 1:
		ref = positional[0]
		if ref == "" {
			ref = "latest"
		}
	default:
		return fmt.Errorf("forge upgrade takes at most one version, for example: forge upgrade v0.2.0")
	}

	goBin, err := lookupGo("go")
	if err != nil {
		return fmt.Errorf("go is not on PATH; install Go 1.22+ (https://go.dev/dl) and run: go install %s@latest: %w", forgeModule, err)
	}

	destDir, err := os.MkdirTemp("", "forge-upgrade-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(destDir)
	fresh := filepath.Join(destDir, "forge"+exeSuffix())

	fmt.Fprintf(stdout, "forge %s: installing %s@%s\n", Version, forgeModule, ref)
	if err := goInstall(goBin, destDir, ref, stdout, stderr); err != nil {
		return err
	}

	installed := ref
	if v, err := goVersion(goBin, fresh); err == nil && v != "" {
		installed = v
	}
	fmt.Fprintf(stdout, "forge upgraded to %s\n", installed)
	return nil
}

// exeSuffix is the executable extension on the host, empty everywhere but
// Windows.
func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// installModule runs `go install <forgeModule>@<ref>` into destDir, which is
// also the working directory so the user's workspace and module files stay out
// of the call.
func installModule(goBin, destDir, ref string, stdout, stderr io.Writer) error {
	cmd := exec.Command(goBin, "install", forgeModule+"@"+ref)
	cmd.Env = append(os.Environ(), "GOBIN="+destDir)
	cmd.Dir = destDir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go install %s@%s failed: %w", forgeModule, ref, err)
	}
	return nil
}

// moduleVersion reads the module version out of the freshly installed binary
// with `go version -m`, the only source that needs no network. It returns an
// error when the version cannot be read; the caller falls back to the ref.
func moduleVersion(goBin, path string) (string, error) {
	out, err := exec.Command(goBin, "version", "-m", path).Output()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "mod" && fields[1] == forgeModule {
			return fields[2], nil
		}
	}
	return "", fmt.Errorf("%s: no version for %s in its build info", path, forgeModule)
}
