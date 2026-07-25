package initializr

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

// BuildURL constructs the Spring Initializr URL using the provided ProjectConfig.
func BuildURL(config *prompt.ProjectConfig) string {
	depsParam := strings.Join(config.Dependencies, ",")
	return fmt.Sprintf(
		"https://start.spring.io/starter.zip?type=%s&language=java&groupId=%s&artifactId=%s&name=%s&packageName=%s&packaging=%s&javaVersion=%s&dependencies=%s",
		config.BuildTool,
		config.GroupID,
		config.ArtifactID,
		config.ProjectName,
		config.PackageName,
		config.Packaging,
		config.JavaVersion,
		depsParam,
	)
}

// Download fetches a Spring Boot project ZIP from Spring Initializr based on
// the provided configuration and returns the local filename of the saved ZIP.
func Download(config *prompt.ProjectConfig) (string, error) {
	url := BuildURL(config)

	filename := config.ProjectName + ".zip"

	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("spring initializr returned %s", resp.Status)
	}

	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filename, nil
}
