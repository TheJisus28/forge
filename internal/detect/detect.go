package detect

import (
	"os"
	"path/filepath"
	"strings"
)

// Stack is the detected runtime of a repository. Empty fields mean unknown.
type Stack struct {
	Languages       []string
	Playbooks       []string
	Runtime         string
	LanguageVersion string
	Build           string
	PackageManager  string
	Test            string
	Dev             string
}

// Dir inspects root for well-known manifest files. Several languages may
// appear in a monorepo; playbooks only include the ones the kit ships.
func Dir(root string) (Stack, error) {
	var s Stack
	seen := map[string]bool{}

	add := func(lang string) {
		if seen[lang] {
			return
		}
		seen[lang] = true
		s.Languages = append(s.Languages, lang)
	}

	has := func(name string) bool {
		_, err := os.Stat(filepath.Join(root, name))
		return err == nil
	}

	read := func(name string) string {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return ""
		}
		return string(b)
	}

	if has("go.mod") {
		add("go")
		s.Build = first(s.Build, "go")
		s.Test = first(s.Test, "go test ./...")
		s.Dev = first(s.Dev, "go run .")
		mod := read("go.mod")
		switch {
		case strings.Contains(mod, "github.com/labstack/echo"):
			s.Runtime = first(s.Runtime, "echo")
		case strings.Contains(mod, "github.com/go-chi/chi"):
			s.Runtime = first(s.Runtime, "chi")
		case strings.Contains(mod, "github.com/gin-gonic/gin"):
			s.Runtime = first(s.Runtime, "gin")
		}
	}

	if has("package.json") {
		add("node")
		pkg := read("package.json")
		if has("tsconfig.json") || strings.Contains(pkg, `"typescript"`) {
			add("typescript")
		}
		switch {
		case has("pnpm-lock.yaml"):
			s.PackageManager = "pnpm"
			s.Build = first(s.Build, "pnpm")
			s.Test = first(s.Test, "pnpm test")
			s.Dev = first(s.Dev, "pnpm dev")
		case has("yarn.lock"):
			s.PackageManager = "yarn"
			s.Build = first(s.Build, "yarn")
			s.Test = first(s.Test, "yarn test")
			s.Dev = first(s.Dev, "yarn dev")
		default:
			s.PackageManager = "npm"
			s.Build = first(s.Build, "npm")
			s.Test = first(s.Test, "npm test")
			s.Dev = first(s.Dev, "npm run dev")
		}
		switch {
		case strings.Contains(pkg, `"next"`):
			s.Runtime = first(s.Runtime, "next")
		case strings.Contains(pkg, `"express"`):
			s.Runtime = first(s.Runtime, "express")
		case strings.Contains(pkg, `"fastify"`):
			s.Runtime = first(s.Runtime, "fastify")
		}
	}

	gradle := has("build.gradle") || has("build.gradle.kts")
	if gradle || has("pom.xml") {
		add("java")
		body := read("build.gradle") + read("build.gradle.kts") + read("pom.xml")
		if strings.Contains(body, "spring-boot") || strings.Contains(body, "org.springframework.boot") {
			s.Runtime = first(s.Runtime, "spring-boot")
		}
		if gradle {
			s.Build = first(s.Build, "gradle")
			if has("gradlew") || has("gradlew.bat") {
				s.Test = first(s.Test, "./gradlew test")
				s.Dev = first(s.Dev, "./gradlew bootRun")
			} else {
				s.Test = first(s.Test, "gradle test")
				s.Dev = first(s.Dev, "gradle bootRun")
			}
		} else {
			s.Build = first(s.Build, "maven")
			s.Test = first(s.Test, "mvn test")
			s.Dev = first(s.Dev, "mvn spring-boot:run")
		}
	}

	if has("pyproject.toml") || has("requirements.txt") {
		add("python")
		body := read("pyproject.toml") + read("requirements.txt")
		switch {
		case containsFold(body, "fastapi"):
			s.Runtime = first(s.Runtime, "fastapi")
		case containsFold(body, "django"):
			s.Runtime = first(s.Runtime, "django")
		case containsFold(body, "flask"):
			s.Runtime = first(s.Runtime, "flask")
		}
		switch {
		case has("uv.lock"):
			s.Build = first(s.Build, "uv")
			s.Test = first(s.Test, "uv run pytest")
		case has("poetry.lock"):
			s.Build = first(s.Build, "poetry")
			s.Test = first(s.Test, "poetry run pytest")
		default:
			s.Build = first(s.Build, "pip")
			s.Test = first(s.Test, "pytest")
		}
		if has("manage.py") {
			s.Dev = first(s.Dev, "python manage.py runserver")
		}
	}
	if has("Cargo.toml") {
		add("rust")
	}
	if has("pubspec.yaml") {
		add("dart")
	}

	for _, lang := range s.Languages {
		switch lang {
		case "java", "go", "node", "python":
			s.Playbooks = appendUnique(s.Playbooks, lang)
		}
	}

	return s, nil
}

func containsFold(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), needle)
}

func first(cur, next string) string {
	if cur != "" {
		return cur
	}
	return next
}

func appendUnique(dst []string, v string) []string {
	for _, x := range dst {
		if x == v {
			return dst
		}
	}
	return append(dst, v)
}
