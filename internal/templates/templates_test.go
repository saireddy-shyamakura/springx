package templates_test

import (
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/config"
	"github.com/saireddy-shyamakura/springx/internal/templates"
)

func TestList(t *testing.T) {
	all := templates.List()
	if len(all) < 6 {
		t.Errorf("expected at least 6 built-in templates, got %d", len(all))
	}

	names := make(map[string]bool)
	for _, tmpl := range all {
		names[tmpl.Name] = true
	}

	expected := []string{"rest-api", "jpa", "security", "microservice", "kafka", "ai"}
	for _, exp := range expected {
		if !names[exp] {
			t.Errorf("expected template %q in list, but it was missing", exp)
		}
	}
}

func TestGet_ValidCaseInsensitive(t *testing.T) {
	tests := []struct {
		input            string
		expectedName     string
		expectedDepCount int
	}{
		{input: "rest-api", expectedName: "rest-api", expectedDepCount: 4},
		{input: "REST-API", expectedName: "rest-api", expectedDepCount: 4},
		{input: "jpa", expectedName: "jpa", expectedDepCount: 5},
		{input: "  Security  ", expectedName: "security", expectedDepCount: 3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tmpl, err := templates.Get(tt.input)
			if err != nil {
				t.Fatalf("unexpected error looking up %q: %v", tt.input, err)
			}
			if tmpl.Name != tt.expectedName {
				t.Errorf("expected template name %q, got %q", tt.expectedName, tmpl.Name)
			}
			if len(tmpl.Dependencies) != tt.expectedDepCount {
				t.Errorf("expected %d dependencies, got %d", tt.expectedDepCount, len(tmpl.Dependencies))
			}
		})
	}
}

func TestGet_NotFound(t *testing.T) {
	_, err := templates.Get("non-existent-template")
	if err == nil {
		t.Error("expected error for non-existent template, got nil")
	}
}

func TestApplyTemplate_OverridesConfigDefaults(t *testing.T) {
	tmpl, err := templates.Get("rest-api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := config.DefaultConfig() // JavaVersion="21", BuildTool="maven-project", Packaging="jar"
	cfg.JavaVersion = "17"        // user preference — template should override this
	cfg.BuildTool = "gradle-project"

	deps := templates.ApplyTemplate(tmpl, &cfg)

	if cfg.JavaVersion != tmpl.Defaults.JavaVersion {
		t.Errorf("expected JavaVersion %q, got %q", tmpl.Defaults.JavaVersion, cfg.JavaVersion)
	}
	if cfg.BuildTool != tmpl.Defaults.BuildTool {
		t.Errorf("expected BuildTool %q, got %q", tmpl.Defaults.BuildTool, cfg.BuildTool)
	}
	if cfg.Packaging != tmpl.Defaults.Packaging {
		t.Errorf("expected Packaging %q, got %q", tmpl.Defaults.Packaging, cfg.Packaging)
	}
	if len(deps) != len(tmpl.Dependencies) {
		t.Errorf("expected %d dependencies, got %d", len(tmpl.Dependencies), len(deps))
	}
	for i, d := range tmpl.Dependencies {
		if deps[i] != d {
			t.Errorf("dependency[%d]: expected %q, got %q", i, d, deps[i])
		}
	}
}

func TestApplyTemplate_ReturnsCopy(t *testing.T) {
	tmpl, _ := templates.Get("kafka")
	cfg := config.DefaultConfig()

	deps := templates.ApplyTemplate(tmpl, &cfg)

	// Mutating the returned slice must not affect the template's stored list.
	if len(deps) == 0 {
		t.Fatal("expected non-empty dependency list")
	}
	original := deps[0]
	deps[0] = "mutated"

	fresh, _ := templates.Get("kafka")
	if fresh.Dependencies[0] != original {
		t.Errorf("ApplyTemplate returned a direct reference to template Dependencies; mutation affected source")
	}
}

func TestApplyTemplate_NilSafety(t *testing.T) {
	cfg := config.DefaultConfig()

	// nil template
	deps := templates.ApplyTemplate(nil, &cfg)
	if deps != nil {
		t.Errorf("expected nil deps for nil template, got %v", deps)
	}

	// nil config
	tmpl, _ := templates.Get("jpa")
	deps = templates.ApplyTemplate(tmpl, nil)
	if deps != nil {
		t.Errorf("expected nil deps for nil config, got %v", deps)
	}
}
