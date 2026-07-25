package plugins

import (
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/templates"
)

// Apply iterates over all enabled plugins and pushes their contributions into
// the three global registries:
//
//   - TemplatePlugin  → templates.BuiltIn (appended)
//   - HookPlugin      → postgen registry (via postgen.Register)
//   - DependencyProvider → the supplied *metadata.Metadata (groups appended)
//
// Apply is idempotent within a single process lifetime because duplicate
// registrations panic in the underlying registries; callers must call it
// exactly once. Pass a nil meta when no live metadata is available
// (e.g. during plugin list commands that do not fetch metadata).
func Apply(meta *metadata.Metadata) {
	for _, p := range Enabled() {
		applyOne(p, meta)
	}
}

// applyOne applies a single plugin's contributions.
func applyOne(p Plugin, meta *metadata.Metadata) {
	if tp, ok := p.(TemplatePlugin); ok {
		for _, t := range tp.Templates() {
			if !templateAlreadyRegistered(t.Name) {
				templates.BuiltIn = append(templates.BuiltIn, t)
			}
		}
	}

	if hp, ok := p.(HookPlugin); ok {
		for _, h := range hp.Hooks() {
			// postgen.Register panics on duplicates — guard against
			// double-apply by checking first.
			if _, err := postgen.Lookup(h.Name()); err != nil {
				postgen.Register(h)
			}
		}
	}

	if dp, ok := p.(DependencyProvider); ok && meta != nil {
		for _, g := range dp.DependencyGroups() {
			if !depGroupAlreadyPresent(meta, g.Name) {
				meta.Dependencies.Values = append(meta.Dependencies.Values, g)
			}
		}
	}
}

// templateAlreadyRegistered reports whether a template with the given name
// (case-insensitive) already exists in templates.BuiltIn.
func templateAlreadyRegistered(name string) bool {
	n := strings.ToLower(name)
	for _, t := range templates.BuiltIn {
		if strings.ToLower(t.Name) == n {
			return true
		}
	}
	return false
}

// depGroupAlreadyPresent reports whether a dependency group with groupName
// already appears in meta.
func depGroupAlreadyPresent(meta *metadata.Metadata, groupName string) bool {
	n := strings.ToLower(groupName)
	for _, g := range meta.Dependencies.Values {
		if strings.ToLower(g.Name) == n {
			return true
		}
	}
	return false
}

// Summary returns a human-readable count of what a plugin contributes.
func Summary(p Plugin) ContributionSummary {
	s := ContributionSummary{}
	if tp, ok := p.(TemplatePlugin); ok {
		s.Templates = len(tp.Templates())
	}
	if hp, ok := p.(HookPlugin); ok {
		s.Hooks = len(hp.Hooks())
	}
	if dp, ok := p.(DependencyProvider); ok {
		for _, g := range dp.DependencyGroups() {
			s.DependencyGroups++
			s.Dependencies += len(g.Values)
		}
	}
	return s
}

// ContributionSummary reports how many items a plugin contributes per type.
type ContributionSummary struct {
	Templates        int
	Hooks            int
	DependencyGroups int
	Dependencies     int
}
