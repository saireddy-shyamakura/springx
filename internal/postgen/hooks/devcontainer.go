package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func init() {
	postgen.Register(&devcontainerHook{})
}

type devcontainerHook struct{}

func (d *devcontainerHook) Name() string { return "devcontainer" }

// devcontainerJSON mirrors the Dev Containers spec for devcontainer.json.
// Only the fields relevant to a Spring Boot project are included.
type devcontainerJSON struct {
	Features     map[string]any    `json:"features,omitempty"`
	Settings     map[string]any    `json:"settings,omitempty"`
	RemoteEnv    map[string]string `json:"remoteEnv,omitempty"`
	Extensions   []string          `json:"extensions,omitempty"`
	ForwardPorts []int             `json:"forwardPorts,omitempty"`
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	PostCreate   string            `json:"postCreateCommand,omitempty"`
}

func (d *devcontainerHook) Run(projectPath string, cfg *prompt.ProjectConfig) error {
	dir := filepath.Join(projectPath, ".devcontainer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Map user-selected Java version to a Microsoft devcontainer base image tag.
	javaTag := javaImageTag(cfg.JavaVersion)

	dc := devcontainerJSON{
		Name:  fmt.Sprintf("%s Dev Container", cfg.ProjectName),
		Image: fmt.Sprintf("mcr.microsoft.com/devcontainers/java:%s", javaTag),
		Features: map[string]any{
			"ghcr.io/devcontainers/features/java:1": map[string]any{
				"version":       cfg.JavaVersion,
				"installMaven":  true,
				"installGradle": true,
			},
		},
		ForwardPorts: []int{8080},
		Extensions: []string{
			"vscjava.vscode-java-pack",
			"vmware.vscode-spring-boot",
		},
		Settings: map[string]any{
			"java.configuration.updateBuildConfiguration": "automatic",
		},
		PostCreate: buildPostCreate(cfg),
		RemoteEnv: map[string]string{
			"SPRING_PROFILES_ACTIVE": "dev",
		},
	}

	data, err := json.MarshalIndent(dc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(filepath.Join(dir, "devcontainer.json"), data, 0o644)
}

// javaImageTag maps a Java version string to the devcontainer image tag.
func javaImageTag(version string) string {
	switch version {
	case "21":
		return "21-bullseye"
	case "17":
		return "17-bullseye"
	case "11":
		return "11-bullseye"
	default:
		return "21-bullseye"
	}
}

// buildPostCreate returns a shell command to run after the container is created.
func buildPostCreate(cfg *prompt.ProjectConfig) string {
	if isMavenProject(cfg.BuildTool) {
		return "./mvnw -q dependency:resolve"
	}
	return "./gradlew -q dependencies"
}
