package templates_test

import (
	"strings"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/config"
	"github.com/saireddy-shyamakura/springx/internal/templates"
)

// ── List ──────────────────────────────────────────────────────────────────────

func TestList_ReturnsBuiltIns(t *testing.T) {
	all := templates.List()
	if len(all) == 0 {
		t.Fatal("expected at least one built-in template, got none")
	}
}

func TestList_EachHasName(t *testing.T) {
	for _, tmpl := range templates.List() {
		if tmpl.Name == "" {
			t.Errorf("template has empty name: %+v", tmpl)
		}
	}
}

func TestList_EachHasDescription(t *testing.T) {
	for _, tmpl := range templates.List() {
		if tmpl.Description == "" {
			t.Errorf("template %q has empty description", tmpl.Name)
		}
	}
}

func TestList_EachHasDependencies(t *testing.T) {
	for _, tmpl := range templates.List() {
		if len(tmpl.Dependencies) == 0 {
			t.Errorf("template %q has no dependencies", tmpl.Name)
		}
	}
}

func TestList_NamesAreUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, tmpl := range templates.List() {
		lower := strings.ToLower(tmpl.Name)
		if seen[lower] {
			t.Errorf("duplicate template name: %q", tmpl.Name)
		}
		seen[lower] = true
	}
}

// ── Get ───────────────────────────────────────────────────────────────────────

func TestGet_ExistingTemplate(t *testing.T) {
	tmpl, err := templates.Get("rest-api")
	if err != nil {
		t.Fatalf("Get('rest-api') returned error: %v", err)
	}
	if tmpl.Name != "rest-api" {
		t.Errorf("expected name 'rest-api', got %q", tmpl.Name)
	}
}

func TestGet_CaseInsensitive(t *testing.T) {
	for _, name := range []string{"REST-API", "Rest-Api", "REST-api"} {
		if _, err := templates.Get(name); err != nil {
			t.Errorf("Get(%q) should be case-insensitive, got error: %v", name, err)
		}
	}
}

func TestGet_UnknownTemplateReturnsError(t *testing.T) {
	_, err := templates.Get("nonexistent-template-xyz")
	if err == nil {
		t.Error("expected error for unknown template name, got nil")
	}
}

func TestGet_ErrorMessageSuggestsListCommand(t *testing.T) {
	_, err := templates.Get("invalid")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "template list") && !strings.Contains(err.Error(), "springx") {
		t.Errorf("error message should mention list command, got: %v", err)
	}
}

// ── ApplyTemplate ─────────────────────────────────────────────────────────────

func TestApplyTemplate_SetsJavaVersion(t *testing.T) {
	tmpl, _ := templates.Get("rest-api")
	cfg := &config.Config{}
	templates.ApplyTemplate(tmpl, cfg)
	if cfg.JavaVersion != tmpl.Defaults.JavaVersion {
		t.Errorf("expected JavaVersion %q, got %q", tmpl.Defaults.JavaVersion, cfg.JavaVersion)
	}
}

func TestApplyTemplate_SetsBuildTool(t *testing.T) {
	tmpl, _ := templates.Get("jpa")
	cfg := &config.Config{}
	templates.ApplyTemplate(tmpl, cfg)
	if cfg.BuildTool != tmpl.Defaults.BuildTool {
		t.Errorf("expected BuildTool %q, got %q", tmpl.Defaults.BuildTool, cfg.BuildTool)
	}
}

func TestApplyTemplate_ReturnsDependencies(t *testing.T) {
	tmpl, _ := templates.Get("jpa")
	cfg := &config.Config{}
	deps := templates.ApplyTemplate(tmpl, cfg)
	if len(deps) == 0 {
		t.Error("expected non-empty dependency list from jpa template")
	}
	// jpa template must include data-jpa
	found := false
	for _, d := range deps {
		if d == "data-jpa" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'data-jpa' in jpa template deps, got: %v", deps)
	}
}

func TestApplyTemplate_OverwritesWithTemplateDefaults(t *testing.T) {
	// ApplyTemplate applies template defaults unconditionally — templates are
	// authoritative for the fields they specify. Users can still change values
	// interactively after the template is applied.
	tmpl, _ := templates.Get("rest-api")
	cfg := &config.Config{JavaVersion: "17"}
	templates.ApplyTemplate(tmpl, cfg)
	// Template default wins (rest-api defaults to Java 21).
	if cfg.JavaVersion != "21" {
		t.Errorf("expected template default '21' to overwrite '17', got %q", cfg.JavaVersion)
	}
}

func TestApplyTemplate_NilInputsAreSafe(t *testing.T) {
	deps := templates.ApplyTemplate(nil, nil)
	if deps != nil {
		t.Errorf("expected nil deps for nil inputs, got %v", deps)
	}
}

// ── Known templates exist ─────────────────────────────────────────────────────

func TestKnownTemplatesExist(t *testing.T) {
	for _, name := range []string{"rest-api", "jpa", "security", "microservice", "kafka", "ai"} {
		if _, err := templates.Get(name); err != nil {
			t.Errorf("built-in template %q not found: %v", name, err)
		}
	}
}
