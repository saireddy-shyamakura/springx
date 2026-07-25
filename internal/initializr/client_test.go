package initializr_test

import (
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/initializr"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		config   *prompt.ProjectConfig
		expected string
	}{
		{
			name: "Maven and Jar with single dependency",
			config: &prompt.ProjectConfig{
				ProjectName:  "bookstore",
				GroupID:      "com.saireddy",
				ArtifactID:   "bookstore",
				PackageName:  "com.saireddy.bookstore",
				BuildTool:    "maven-project",
				Packaging:    "jar",
				JavaVersion:  "21",
				Dependencies: []string{"web"},
			},
			expected: "https://start.spring.io/starter.zip?type=maven-project&language=java&groupId=com.saireddy&artifactId=bookstore&name=bookstore&packageName=com.saireddy.bookstore&packaging=jar&javaVersion=21&dependencies=web",
		},
		{
			name: "Gradle and War with multiple dependencies",
			config: &prompt.ProjectConfig{
				ProjectName:  "my-app",
				GroupID:      "org.test",
				ArtifactID:   "my-app",
				PackageName:  "org.test.myapp",
				BuildTool:    "gradle-project",
				Packaging:    "war",
				JavaVersion:  "17",
				Dependencies: []string{"web", "data-jpa", "postgresql", "lombok"},
			},
			expected: "https://start.spring.io/starter.zip?type=gradle-project&language=java&groupId=org.test&artifactId=my-app&name=my-app&packageName=org.test.myapp&packaging=war&javaVersion=17&dependencies=web,data-jpa,postgresql,lombok",
		},
		{
			name: "Gradle Kotlin and Jar",
			config: &prompt.ProjectConfig{
				ProjectName:  "demo",
				GroupID:      "com.example",
				ArtifactID:   "demo",
				PackageName:  "com.example.demo",
				BuildTool:    "gradle-project-kotlin",
				Packaging:    "jar",
				JavaVersion:  "24",
				Dependencies: []string{"web", "devtools"},
			},
			expected: "https://start.spring.io/starter.zip?type=gradle-project-kotlin&language=java&groupId=com.example&artifactId=demo&name=demo&packageName=com.example.demo&packaging=jar&javaVersion=24&dependencies=web,devtools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := initializr.BuildURL(tt.config)
			if actual != tt.expected {
				t.Errorf("expected URL:\n%s\ngot:\n%s", tt.expected, actual)
			}
		})
	}
}
