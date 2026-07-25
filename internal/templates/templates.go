package templates

import (
	"fmt"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/config"
)

// TemplateDefaults specifies optional default parameters for a template.
type TemplateDefaults struct {
	BuildTool   string
	Packaging   string
	JavaVersion string
}

// Template defines a Spring Boot project preset configuration.
type Template struct {
	Name         string
	Description  string
	Dependencies []string
	Defaults     TemplateDefaults
}

// BuiltIn holds all built-in template presets.
// Adding a new template preset requires adding an entry to this single slice.
var BuiltIn = []Template{
	{
		Name:        "rest-api",
		Description: "REST API with validation and monitoring.",
		Dependencies: []string{
			"web",
			"validation",
			"actuator",
			"lombok",
		},
		Defaults: TemplateDefaults{
			JavaVersion: "21",
			BuildTool:   "maven-project",
			Packaging:   "jar",
		},
	},
	{
		Name:        "jpa",
		Description: "Relational database project with Spring Data JPA and Flyway migrations.",
		Dependencies: []string{
			"data-jpa",
			"postgresql",
			"flyway",
			"validation",
			"lombok",
		},
		Defaults: TemplateDefaults{
			JavaVersion: "21",
			BuildTool:   "maven-project",
			Packaging:   "jar",
		},
	},
	{
		Name:        "security",
		Description: "Spring Security project with OAuth2 client support.",
		Dependencies: []string{
			"security",
			"oauth2-client",
			"validation",
		},
		Defaults: TemplateDefaults{
			JavaVersion: "21",
			BuildTool:   "maven-project",
			Packaging:   "jar",
		},
	},
	{
		Name:        "microservice",
		Description: "Microservice template with Spring Cloud OpenFeign and Actuator.",
		Dependencies: []string{
			"web",
			"actuator",
			"cloud-feign",
			"cloud-config-client",
			"validation",
			"lombok",
		},
		Defaults: TemplateDefaults{
			JavaVersion: "21",
			BuildTool:   "maven-project",
			Packaging:   "jar",
		},
	},
	{
		Name:        "kafka",
		Description: "Event-driven microservice using Apache Kafka.",
		Dependencies: []string{
			"kafka",
			"actuator",
			"web",
		},
		Defaults: TemplateDefaults{
			JavaVersion: "21",
			BuildTool:   "maven-project",
			Packaging:   "jar",
		},
	},
	{
		Name:        "ai",
		Description: "Reserved for future Spring AI support.",
		Dependencies: []string{
			"web",
			"actuator",
		},
		Defaults: TemplateDefaults{
			JavaVersion: "21",
			BuildTool:   "maven-project",
			Packaging:   "jar",
		},
	},
}

// List returns all registered template presets.
func List() []Template {
	return BuiltIn
}

// Get finds a template by name using case-insensitive matching.
func Get(name string) (*Template, error) {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, t := range BuiltIn {
		if strings.ToLower(t.Name) == n {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("template %q not found. Run 'springx template list' to view available presets", name)
}

// ApplyTemplate merges a template's defaults into the provided config and
// returns the template's dependency list. Only non-empty template fields
// overwrite the corresponding config fields, so a user's own configuration
// is always the base layer and the template acts as a targeted override.
//
// The returned []string is the ordered dependency list from the template.
// The caller is responsible for seeding this into the project configuration
// before launching any interactive prompts.
func ApplyTemplate(tmpl *Template, cfg *config.Config) (deps []string) {
	if tmpl == nil || cfg == nil {
		return nil
	}
	if tmpl.Defaults.JavaVersion != "" {
		cfg.JavaVersion = tmpl.Defaults.JavaVersion
	}
	if tmpl.Defaults.BuildTool != "" {
		cfg.BuildTool = tmpl.Defaults.BuildTool
	}
	if tmpl.Defaults.Packaging != "" {
		cfg.Packaging = tmpl.Defaults.Packaging
	}
	return append([]string(nil), tmpl.Dependencies...)
}
