// Package plugins provides springx's third-party extensibility layer.
//
// # Overview
//
// A plugin is a compiled Go package that calls one or more of the Register*
// functions below inside its init() function, then is blank-imported by the
// host binary. This is the same pattern springx uses internally for its
// built-in hooks and is idiomatic Go — no shared-object files, no CGO, no
// runtime code loading.
//
// # Plugin layout (on disk)
//
// Every plugin directory under ~/.config/springx/plugins/<name>/ must contain
// a plugin.json manifest so springx can surface metadata for the
// `springx plugin list/info` commands and honour enable/disable state:
//
//	~/.config/springx/plugins/
//	└── aws/
//	    └── plugin.json     ← required manifest
//
// The manifest is purely declarative; the plugin's Go code is compiled into
// the springx binary at build time.
//
// # Interfaces
//
// Three interfaces cover the extension points:
//
//	TemplatePlugin      — contributes project presets (templates)
//	HookPlugin          — contributes post-generation automation steps
//	DependencyProvider  — contributes additional dependency groups to the picker
//
// A plugin struct may implement any combination of these interfaces.
//
// # Minimal plugin example
//
//	package myplugin
//
//	import "github.com/saireddy-shyamakura/springx/internal/plugins"
//
//	func init() { plugins.RegisterPlugin(&myPlugin{}) }
//
//	type myPlugin struct{}
//	func (p *myPlugin) Manifest() plugins.Manifest { ... }
//	func (p *myPlugin) Templates() []templates.Template { ... }
package plugins

import (
	"fmt"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/templates"
)

// ── Manifest ──────────────────────────────────────────────────────────────────

// Manifest carries the declarative metadata for a plugin. It mirrors the
// fields in plugin.json and is also returned by the Plugin interface so that
// fully in-process plugins do not need a file on disk.
type Manifest struct {
	// Name is the canonical identifier used in CLI commands and the registry.
	// Must be lowercase, URL-safe, and unique across all loaded plugins.
	Name string `json:"name"`

	// Version is the human-readable plugin version (e.g. "1.0.0").
	Version string `json:"version"`

	// Author is the plugin author or organisation.
	Author string `json:"author"`

	// Description is a short, one-line summary of what the plugin provides.
	Description string `json:"description"`

	// Homepage is an optional URL to the plugin's source or documentation.
	Homepage string `json:"homepage,omitempty"`

	// Tags is an optional list of searchable keywords.
	Tags []string `json:"tags,omitempty"`
}

// ── Core interface ────────────────────────────────────────────────────────────

// Plugin is the minimum contract every plugin must satisfy.
// A single struct may additionally implement TemplatePlugin, HookPlugin,
// and/or DependencyProvider.
type Plugin interface {
	// Manifest returns the plugin's static metadata.
	Manifest() Manifest
}

// ── Extension-point interfaces ────────────────────────────────────────────────

// TemplatePlugin contributes project template presets to springx.
// The returned templates are merged into the global templates registry and
// become available via `springx template list` and `springx new --template`.
type TemplatePlugin interface {
	Plugin
	// Templates returns the list of templates this plugin provides.
	// Each template must have a unique Name across the combined built-in +
	// plugin template set.
	Templates() []templates.Template
}

// HookPlugin contributes post-generation hooks to springx.
// Each hook is registered in the postgen global registry and becomes
// available via `springx new --hook <name>`.
type HookPlugin interface {
	Plugin
	// Hooks returns the list of hooks this plugin provides.
	Hooks() []postgen.Hook
}

// DependencyProvider contributes additional dependency groups to the live
// metadata that is shown in the Bubble Tea dependency picker.
// Groups are appended after the built-in Spring Initializr groups.
type DependencyProvider interface {
	Plugin
	// DependencyGroups returns the additional groups this plugin provides.
	DependencyGroups() []metadata.DependencyGroup
}

// ── In-process registry ───────────────────────────────────────────────────────

// entry wraps a registered plugin together with its enable/disable state.
type entry struct {
	plugin  Plugin
	enabled bool
}

// pluginRegistry is the global in-process registry populated by RegisterPlugin.
var pluginRegistry = map[string]*entry{}

// RegisterPlugin adds p to the global registry. It panics on duplicate names
// so mis-registrations surface at startup rather than silently at runtime.
// Call this from your plugin's init() function.
func RegisterPlugin(p Plugin) {
	m := p.Manifest()
	name := strings.ToLower(strings.TrimSpace(m.Name))
	if name == "" {
		panic("plugins: RegisterPlugin called with empty plugin name")
	}
	if _, exists := pluginRegistry[name]; exists {
		panic(fmt.Sprintf("plugins: duplicate plugin registration for %q", name))
	}
	pluginRegistry[name] = &entry{plugin: p, enabled: true}
}

// Registered returns all plugins that have been registered in the current
// process, regardless of their enable/disable state.
func Registered() []Plugin {
	result := make([]Plugin, 0, len(pluginRegistry))
	for _, e := range pluginRegistry {
		result = append(result, e.plugin)
	}
	sortPlugins(result)
	return result
}

// Enabled returns only the plugins that are currently enabled.
func Enabled() []Plugin {
	var result []Plugin
	for _, e := range pluginRegistry {
		if e.enabled {
			result = append(result, e.plugin)
		}
	}
	sortPlugins(result)
	return result
}

// SetEnabled flips the enable/disable state of the named plugin in memory.
// Persisting the state to disk is the responsibility of the caller (see
// loader.go). Returns an error if the plugin is not registered.
func SetEnabled(name string, enabled bool) error {
	e, ok := pluginRegistry[strings.ToLower(name)]
	if !ok {
		return fmt.Errorf("plugin %q is not registered", name)
	}
	e.enabled = enabled
	return nil
}

// Lookup returns the Plugin registered under name (case-insensitive).
func Lookup(name string) (Plugin, bool) {
	e, ok := pluginRegistry[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	return e.plugin, true
}

// IsEnabled reports whether the named plugin is registered and enabled.
func IsEnabled(name string) bool {
	e, ok := pluginRegistry[strings.ToLower(name)]
	return ok && e.enabled
}

// sortPlugins sorts a Plugin slice by manifest name (ascending).
func sortPlugins(ps []Plugin) {
	for i := 1; i < len(ps); i++ {
		key := ps[i]
		j := i - 1
		for j >= 0 && ps[j].Manifest().Name > key.Manifest().Name {
			ps[j+1] = ps[j]
			j--
		}
		ps[j+1] = key
	}
}
