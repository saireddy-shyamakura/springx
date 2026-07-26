// Package initializr communicates with Spring Initializr to download
// generated Spring Boot project archives.
package initializr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

// MetadataBaseURL is the base URL used by BuildURL and Download.
// Override in tests to point at a local httptest.Server.
var MetadataBaseURL = "https://start.spring.io/starter.zip?"

// BuildURL constructs the full Spring Initializr download URL from cfg.
func BuildURL(cfg *prompt.ProjectConfig) string {
	depsParam := strings.Join(cfg.Dependencies, ",")
	return fmt.Sprintf(
		"%stype=%s&language=java&groupId=%s&artifactId=%s&name=%s&packageName=%s&packaging=%s&javaVersion=%s&dependencies=%s",
		MetadataBaseURL,
		cfg.BuildTool,
		cfg.GroupID,
		cfg.ArtifactID,
		cfg.ProjectName,
		cfg.PackageName,
		cfg.Packaging,
		cfg.JavaVersion,
		depsParam,
	)
}

// Download fetches a Spring Boot project ZIP from Spring Initializr and
// writes it to the current working directory. It returns the filename of the
// saved ZIP on success.
func Download(cfg *prompt.ProjectConfig) (string, error) {
	url := BuildURL(cfg)
	filename := cfg.ProjectName + ".zip"

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request to Spring Initializr failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spring Initializr returned HTTP %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create output file %s: %w", filename, err)
	}
	defer file.Close() //nolint:errcheck

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", fmt.Errorf("failed to write ZIP to %s: %w", filename, err)
	}

	return filename, nil
}
