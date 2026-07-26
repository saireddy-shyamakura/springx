package prompt_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/config"
	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

var mockMetadataJSON = `{
  "bootVersion": {
    "default": "4.1.0",
    "values": [
      {"id": "4.1.0", "name": "4.1.0"}
    ]
  },
  "javaVersion": {
    "default": "21",
    "values": [
      {"id": "24", "name": "24"},
      {"id": "21", "name": "21"},
      {"id": "17", "name": "17"}
    ]
  },
  "packaging": {
    "default": "jar",
    "values": [
      {"id": "jar", "name": "Jar"},
      {"id": "war", "name": "War"}
    ]
  },
  "language": {
    "default": "java",
    "values": [
      {"id": "java", "name": "Java"}
    ]
  },
  "type": {
    "default": "maven-project",
    "values": [
      {
        "id": "gradle-project",
        "name": "Gradle - Groovy",
        "action": "/starter.zip"
      },
      {
        "id": "gradle-project-kotlin",
        "name": "Gradle - Kotlin",
        "action": "/starter.zip"
      },
      {
        "id": "maven-project",
        "name": "Maven",
        "action": "/starter.zip"
      },
      {
        "id": "maven-build",
        "name": "Maven POM",
        "action": "/pom.xml"
      }
    ]
  }
}`

func getMockMetadata(t *testing.T) *metadata.Metadata {
	var meta metadata.Metadata
	if err := json.Unmarshal([]byte(mockMetadataJSON), &meta); err != nil {
		t.Fatalf("failed to unmarshal mock metadata: %v", err)
	}
	return &meta
}

func TestPromptForConfigIO_Success(t *testing.T) {
	// Simulate user entering:
	// 1. Project Name: "   my-cool-project   " (should trim to "my-cool-project")
	// 2. Group ID: "" (should default to "com.example")
	// 3. Artifact ID: "" (should default to project name: "my-cool-project")
	// 4. Package Name: "   com.custom.pkg   " (should trim to "com.custom.pkg")
	// 5. Build Tool: "2" (should select Gradle - Kotlin, which has ID gradle-project-kotlin)
	// 6. Packaging: "war" (case-insensitive name, should select War, which has ID war)
	// 7. JavaVersion: "" (should default to default ID "21")
	input := strings.Join([]string{
		"   my-cool-project   ",
		"",
		"",
		"   com.custom.pkg   ",
		"2",
		"war",
		"",
	}, "\n") + "\n"

	var out bytes.Buffer
	in := strings.NewReader(input)

	meta := getMockMetadata(t)
	cfg, err := prompt.PromptForConfigIO(in, &out, meta, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProjectName != "my-cool-project" {
		t.Errorf("expected ProjectName to be 'my-cool-project', got %q", cfg.ProjectName)
	}
	if cfg.GroupID != "com.example" {
		t.Errorf("expected GroupID to be default 'com.example', got %q", cfg.GroupID)
	}
	if cfg.ArtifactID != "my-cool-project" {
		t.Errorf("expected ArtifactID to be default 'my-cool-project', got %q", cfg.ArtifactID)
	}
	if cfg.PackageName != "com.custom.pkg" {
		t.Errorf("expected PackageName to be 'com.custom.pkg', got %q", cfg.PackageName)
	}
	if cfg.BuildTool != "gradle-project-kotlin" {
		t.Errorf("expected BuildTool to be 'gradle-project-kotlin', got %q", cfg.BuildTool)
	}
	if cfg.Packaging != "war" {
		t.Errorf("expected Packaging to be 'war', got %q", cfg.Packaging)
	}
	if cfg.JavaVersion != "21" {
		t.Errorf("expected JavaVersion to be '21', got %q", cfg.JavaVersion)
	}
}

func TestPromptForConfigIO_RejectsEmptyProjectName(t *testing.T) {
	// Simulate user entering:
	// 1. Project Name: "" (empty, should be rejected)
	// 2. Project Name: "   " (whitespace, should be rejected)
	// 3. Project Name: "realproject"
	// 4-7. Enter for everything else to accept defaults
	input := strings.Join([]string{
		"",
		"   ",
		"realproject",
		"",
		"",
		"",
		"",
		"",
		"",
	}, "\n") + "\n"

	var out bytes.Buffer
	in := strings.NewReader(input)

	meta := getMockMetadata(t)
	cfg, err := prompt.PromptForConfigIO(in, &out, meta, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ProjectName != "realproject" {
		t.Errorf("expected ProjectName to be 'realproject', got %q", cfg.ProjectName)
	}

	// Verify that the output contained the error message at least twice
	outStr := out.String()
	count := strings.Count(outStr, "Value cannot be empty. Please try again.")
	if count != 2 {
		t.Errorf("expected rejection message 2 times, got it %d times. Output:\n%s", count, outStr)
	}
}

func TestPromptForConfigIO_InvalidSelectionReprompt(t *testing.T) {
	// Simulate user entering:
	// 1. Project Name: "myproject"
	// 2-4. Enter for group, artifact, package
	// 5. Build tool: "invalid" (should prompt again), then "4" (invalid number since only 3 project options exist), then "maven-project" (valid ID)
	// 6-7. Enter for packaging, java
	input := strings.Join([]string{
		"myproject",
		"",
		"",
		"",
		"invalid",
		"4",
		"maven-project",
		"",
		"",
	}, "\n") + "\n"

	var out bytes.Buffer
	in := strings.NewReader(input)

	meta := getMockMetadata(t)
	cfg, err := prompt.PromptForConfigIO(in, &out, meta, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BuildTool != "maven-project" {
		t.Errorf("expected BuildTool to be 'maven-project', got %q", cfg.BuildTool)
	}

	outStr := out.String()
	count := strings.Count(outStr, "Invalid choice. Please try again.")
	if count != 2 {
		t.Errorf("expected invalid choice message 2 times, got %d times. Output:\n%s", count, outStr)
	}
}

func TestPrintSummaryFormat(t *testing.T) {
	projCfg := &prompt.ProjectConfig{
		ProjectName:  "bookstore",
		GroupID:      "com.saireddy",
		ArtifactID:   "bookstore",
		PackageName:  "com.saireddy.bookstore",
		BuildTool:    "maven-project",
		Packaging:    "jar",
		JavaVersion:  "21",
		Dependencies: []string{"web", "data-jpa"},
	}

	var out bytes.Buffer
	meta := getMockMetadata(t)
	prompt.PrintSummary(&out, projCfg, meta)

	expected := `
----------------------------------
Project Configuration
----------------------------------
Project Name : bookstore
Group ID     : com.saireddy
Artifact ID  : bookstore
Package Name : com.saireddy.bookstore
Build Tool   : Maven
Packaging    : Jar
Java Version : 21
Dependencies : web, data-jpa
----------------------------------
`

	// Trim whitespace from lines to compare structure and spacing around colons
	actualLines := strings.Split(strings.TrimSpace(out.String()), "\n")
	expectedLines := strings.Split(strings.TrimSpace(expected), "\n")

	if len(actualLines) != len(expectedLines) {
		t.Fatalf("line count mismatch: got %d, expected %d. Output:\n%q", len(actualLines), len(expectedLines), out.String())
	}

	for i := range expectedLines {
		if actualLines[i] != expectedLines[i] {
			t.Errorf("line %d mismatch:\nexpected: %q\ngot:      %q", i, expectedLines[i], actualLines[i])
		}
	}
}

func TestPromptForConfigIO_WithUserConfig(t *testing.T) {
	// User enters project name "demo" and accepts all pre-configured defaults by hitting Enter
	input := "demo\n\n\n\n\n\n\n"
	var out bytes.Buffer
	in := strings.NewReader(input)

	meta := getMockMetadata(t)
	userCfg := &config.Config{
		GroupID:        "com.custom",
		ArtifactPrefix: "service-",
		PackagePrefix:  "com.custom.pkg",
		JavaVersion:    "17",
		BuildTool:      "gradle-project-kotlin",
		Packaging:      "war",
	}

	projectCfg, err := prompt.PromptForConfigIO(in, &out, meta, userCfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if projectCfg.GroupID != "com.custom" {
		t.Errorf("expected GroupID 'com.custom', got %q", projectCfg.GroupID)
	}
	if projectCfg.ArtifactID != "service-demo" {
		t.Errorf("expected ArtifactID 'service-demo', got %q", projectCfg.ArtifactID)
	}
	if projectCfg.PackageName != "com.custom.pkg.service-demo" {
		t.Errorf("expected PackageName 'com.custom.pkg.service-demo', got %q", projectCfg.PackageName)
	}
	if projectCfg.BuildTool != "gradle-project-kotlin" {
		t.Errorf("expected BuildTool 'gradle-project-kotlin', got %q", projectCfg.BuildTool)
	}
	if projectCfg.Packaging != "war" {
		t.Errorf("expected Packaging 'war', got %q", projectCfg.Packaging)
	}
	if projectCfg.JavaVersion != "17" {
		t.Errorf("expected JavaVersion '17', got %q", projectCfg.JavaVersion)
	}
}
