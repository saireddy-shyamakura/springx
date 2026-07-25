package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func init() {
	postgen.Register(&vscodeHook{})
}

type vscodeHook struct{}

func (v *vscodeHook) Name() string { return "vscode" }

// extensionsJSON is the schema for .vscode/extensions.json.
type extensionsJSON struct {
	Recommendations []string `json:"recommendations"`
}

func (v *vscodeHook) Run(projectPath string, cfg *prompt.ProjectConfig) error {
	dir := filepath.Join(projectPath, ".vscode")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	ext := extensionsJSON{
		Recommendations: []string{
			"vscjava.vscode-java-pack",
			"vmware.vscode-spring-boot",
			"redhat.vscode-xml",
			"redhat.java",
			"vscjava.vscode-maven",
			"vscjava.vscode-gradle",
		},
	}

	data, err := json.MarshalIndent(ext, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(filepath.Join(dir, "extensions.json"), data, 0644)
}
