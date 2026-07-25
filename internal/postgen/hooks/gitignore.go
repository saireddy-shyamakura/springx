package hooks

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func init() {
	postgen.Register(&gitignoreHook{})
}

type gitignoreHook struct{}

func (g *gitignoreHook) Name() string { return "gitignore" }

// springBootIgnore is the canonical set of entries every Spring Boot project
// should have. Entries already present in the file are not duplicated.
var springBootIgnore = []string{
	"HELP.md",
	".gradle/",
	"build/",
	"!gradle/wrapper/gradle-wrapper.jar",
	"!**/src/main/**/build/",
	"!**/src/test/**/build/",
	"### STS ###",
	".apt_generated",
	".classpath",
	".factorypath",
	".project",
	".settings",
	".springBeans",
	".sts4-cache",
	"bin/",
	"!**/src/main/**/bin/",
	"!**/src/test/**/bin/",
	"### IntelliJ IDEA ###",
	".idea",
	"*.iws",
	"*.iml",
	"*.ipr",
	"out/",
	"!**/src/main/**/out/",
	"!**/src/test/**/out/",
	"### VS Code ###",
	".vscode/",
	"### Maven ###",
	"target/",
	"pom.xml.tag",
	"pom.xml.releaseBackup",
	"pom.xml.versionsBackup",
	"pom.xml.next",
	"release.properties",
	"dependency-reduced-pom.xml",
	"buildNumber.properties",
	".mvn/timing.properties",
	".mvn/wrapper/maven-wrapper.jar",
	"### OS ###",
	".DS_Store",
	"Thumbs.db",
}

func (g *gitignoreHook) Run(projectPath string, cfg *prompt.ProjectConfig) error {
	path := filepath.Join(projectPath, ".gitignore")

	existing := ""
	if data, err := os.ReadFile(path); err == nil {
		existing = string(data)
	}

	var additions []string
	for _, entry := range springBootIgnore {
		if !strings.Contains(existing, entry) {
			additions = append(additions, entry)
		}
	}

	if len(additions) == 0 {
		return nil // nothing to add
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add a blank line separator before the new block when merging into an
	// existing file so the result remains human-readable.
	if existing != "" && !strings.HasSuffix(existing, "\n\n") {
		_, _ = f.WriteString("\n")
	}

	_, err = f.WriteString(strings.Join(additions, "\n") + "\n")
	return err
}
