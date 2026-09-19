package detect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TheJisus28/forge/internal/detect"
)

func TestDir_JavaGradleSpring(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "build.gradle", "plugins { id 'org.springframework.boot' version '4.1.0' }\n")
	write(t, dir, "gradlew", "#!/bin/sh\n")
	st, err := detect.Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := st.Languages; len(got) != 1 || got[0] != "java" {
		t.Fatalf("languages=%v", got)
	}
	if st.Runtime != "spring-boot" {
		t.Fatalf("runtime=%q", st.Runtime)
	}
	if st.Test != "./gradlew test" {
		t.Fatalf("test=%q", st.Test)
	}
	if len(st.Playbooks) != 1 || st.Playbooks[0] != "java" {
		t.Fatalf("playbooks=%v", st.Playbooks)
	}
}

func TestDir_GoEcho(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x\n\nrequire github.com/labstack/echo/v4 v4.0.0\n")
	st, err := detect.Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if st.Languages[0] != "go" || st.Runtime != "echo" {
		t.Fatalf("%+v", st)
	}
	if st.Playbooks[0] != "go" {
		t.Fatalf("playbooks=%v", st.Playbooks)
	}
}

func TestDir_NodePnpmTypeScript(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "package.json", `{"devDependencies":{"typescript":"5.0.0","next":"15.0.0"}}`)
	write(t, dir, "pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	write(t, dir, "tsconfig.json", "{}\n")
	st, err := detect.Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(st.Languages, "node") || !contains(st.Languages, "typescript") {
		t.Fatalf("languages=%v", st.Languages)
	}
	if st.PackageManager != "pnpm" || st.Runtime != "next" {
		t.Fatalf("%+v", st)
	}
}

func TestDir_PythonUvFastAPI(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "pyproject.toml", "[project]\ndependencies = [\"FastAPI>=0.115\"]\n")
	write(t, dir, "uv.lock", "version = 1\n")
	st, err := detect.Dir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(st.Languages, "python") || !contains(st.Playbooks, "python") {
		t.Fatalf("%+v", st)
	}
	if st.Runtime != "fastapi" || st.Build != "uv" || st.Test != "uv run pytest" {
		t.Fatalf("%+v", st)
	}
}

func TestDir_Empty(t *testing.T) {
	st, err := detect.Dir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Languages) != 0 {
		t.Fatalf("languages=%v", st.Languages)
	}
}

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
