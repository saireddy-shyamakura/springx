package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func init() {
	postgen.Register(&wrapperHook{})
}

type wrapperHook struct{}

func (w *wrapperHook) Name() string { return "wrapper" }

func (w *wrapperHook) Run(projectPath string, cfg *prompt.ProjectConfig) error {
	isMaven := strings.Contains(strings.ToLower(cfg.BuildTool), "maven")
	isGradle := strings.Contains(strings.ToLower(cfg.BuildTool), "gradle")

	if isMaven {
		return w.check(projectPath, "mvnw", "Maven wrapper (mvnw)")
	}
	if isGradle {
		return w.check(projectPath, "gradlew", "Gradle wrapper (gradlew)")
	}

	// Unknown build tool — nothing to verify.
	return nil
}

func (w *wrapperHook) check(projectPath, filename, label string) error {
	p := filepath.Join(projectPath, filename)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return fmt.Errorf(
			"%s not found at %s — the project may not have been generated correctly",
			label, p,
		)
	}
	return nil
}
