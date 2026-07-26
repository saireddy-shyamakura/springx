package plugins_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/metadata"
	"github.com/saireddy-shyamakura/springx/internal/plugins"
	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
	"github.com/saireddy-shyamakura/springx/internal/templates"
)

// ── helpers ───────────────────────────────────────────────────────────────────

// resetRegistry clears the in-process plugin registry between tests so
// RegisterPlugin does not panic on duplicate names.
// We access the unexported registry indirectly through the public API:
// there is no exported Reset(), so each test uses a fresh fake name.

// fakePlugin is a minimal Plugin that optionally implements all three
// extension-point interfaces.
type fakePlugin struct {
	name      string
	version   string
	tmplates  []templates.Template
	hooks     []postgen.Hook
	depGroups []metadata.DependencyGroup
}

func (f *fakePlugin) Manifest() plugins.Manifest {
	return plugins.Manifest{
		Name:        f.name,
		Version:     f.version,
		Author:      "test",
		Description: "Test plugin " + f.name,
	}
}
func (f *fakePlugin) Templates() []templates.Template              { return f.tmplates }
func (f *fakePlugin) Hooks() []postgen.Hook                        { return f.hooks }
func (f *fakePlugin) DependencyGroups() []metadata.DependencyGroup { return f.depGroups }

// fakeHook is a postgen.Hook that records whether it was called.
type fakeHook struct {
	name   string
	called bool
}

func (h *fakeHook) Name() string { return h.name }
func (h *fakeHook) Run(_ string, _ *prompt.ProjectConfig) error {
	h.called = true
	return nil
}

// uniqueName returns a name that is unlikely to collide with other tests.
func uniqueName(base string) string {
	return base + "-" + randomSuffix()
}

var suffixCounter int

func randomSuffix() string {
	suffixCounter++
	return string(rune('a'+suffixCounter%26)) + string(rune('a'+(suffixCounter/26)%26))
}

// ── Manifest JSON round-trip ──────────────────────────────────────────────────

func TestManifest_JSONRoundTrip(t *testing.T) {
	m := plugins.Manifest{
		Name:        "test-plugin",
		Version:     "2.1.0",
		Author:      "Alice",
		Description: "Does something useful.",
		Homepage:    "https://example.com",
		Tags:        []string{"cloud", "aws"},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got plugins.Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Name != m.Name || got.Version != m.Version || got.Author != m.Author {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "cloud" {
		t.Errorf("tags mismatch: %v", got.Tags)
	}
}

// ── RegisterPlugin / Lookup / Registered ─────────────────────────────────────

func TestRegisterPlugin_And_Lookup(t *testing.T) {
	name := uniqueName("lookup")
	p := &fakePlugin{name: name, version: "1.0.0"}
	plugins.RegisterPlugin(p)

	got, ok := plugins.Lookup(name)
	if !ok {
		t.Fatalf("Lookup(%q) returned false", name)
	}
	if got.Manifest().Name != name {
		t.Errorf("wrong plugin returned: want %q, got %q", name, got.Manifest().Name)
	}
}

func TestRegisterPlugin_PanicsOnDuplicate(t *testing.T) {
	name := uniqueName("dup")
	plugins.RegisterPlugin(&fakePlugin{name: name, version: "1.0.0"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration, got none")
		}
	}()
	plugins.RegisterPlugin(&fakePlugin{name: name, version: "2.0.0"})
}

func TestRegistered_ContainsAll(t *testing.T) {
	before := len(plugins.Registered())

	name1 := uniqueName("reg1")
	name2 := uniqueName("reg2")
	plugins.RegisterPlugin(&fakePlugin{name: name1, version: "1.0.0"})
	plugins.RegisterPlugin(&fakePlugin{name: name2, version: "1.0.0"})

	after := plugins.Registered()
	if len(after) != before+2 {
		t.Errorf("expected %d plugins, got %d", before+2, len(after))
	}
}

func TestRegistered_IsSorted(t *testing.T) {
	all := plugins.Registered()
	for i := 1; i < len(all); i++ {
		a := all[i-1].Manifest().Name
		b := all[i].Manifest().Name
		if a > b {
			t.Errorf("Registered() not sorted: %q > %q at positions %d,%d", a, b, i-1, i)
		}
	}
}

// ── Enable / disable (in-memory) ──────────────────────────────────────────────

func TestSetEnabled_DisableAndEnable(t *testing.T) {
	name := uniqueName("toggle")
	plugins.RegisterPlugin(&fakePlugin{name: name, version: "1.0.0"})

	if !plugins.IsEnabled(name) {
		t.Fatal("newly registered plugin should be enabled by default")
	}

	if err := plugins.SetEnabled(name, false); err != nil {
		t.Fatalf("SetEnabled false: %v", err)
	}
	if plugins.IsEnabled(name) {
		t.Error("plugin should be disabled after SetEnabled(false)")
	}

	// Enabled() should exclude it.
	for _, p := range plugins.Enabled() {
		if p.Manifest().Name == name {
			t.Error("disabled plugin appeared in Enabled()")
		}
	}

	if err := plugins.SetEnabled(name, true); err != nil {
		t.Fatalf("SetEnabled true: %v", err)
	}
	if !plugins.IsEnabled(name) {
		t.Error("plugin should be enabled after SetEnabled(true)")
	}
}

func TestSetEnabled_UnknownPlugin(t *testing.T) {
	err := plugins.SetEnabled("does-not-exist-xyz", false)
	if err == nil {
		t.Error("expected error for unknown plugin name, got nil")
	}
}

// ── Enable / disable (persistence) ───────────────────────────────────────────

func TestPersistDisabled_And_LoadDisabledIntoRegistry(t *testing.T) {
	dir := t.TempDir()
	plugins.PluginDirOverride = dir
	defer func() { plugins.PluginDirOverride = "" }()

	name := uniqueName("persist")
	plugins.RegisterPlugin(&fakePlugin{name: name, version: "1.0.0"})

	if err := plugins.PersistDisabled(name); err != nil {
		t.Fatalf("PersistDisabled: %v", err)
	}

	// Re-enable in memory so we can test that LoadDisabledIntoRegistry resets it.
	_ = plugins.SetEnabled(name, true)
	if !plugins.IsEnabled(name) {
		t.Fatal("expected plugin enabled in memory before load")
	}

	plugins.LoadDisabledIntoRegistry()

	if plugins.IsEnabled(name) {
		t.Error("LoadDisabledIntoRegistry should have disabled the plugin")
	}
}

func TestPersistEnabled_RemovesFromDisabledFile(t *testing.T) {
	dir := t.TempDir()
	plugins.PluginDirOverride = dir
	defer func() { plugins.PluginDirOverride = "" }()

	name := uniqueName("re-enable")
	plugins.RegisterPlugin(&fakePlugin{name: name, version: "1.0.0"})

	_ = plugins.PersistDisabled(name)
	_ = plugins.PersistEnabled(name) // remove from disabled list

	plugins.LoadDisabledIntoRegistry()

	if !plugins.IsEnabled(name) {
		t.Error("plugin should be enabled after PersistEnabled + reload")
	}
}

// ── DiscoverManifests ─────────────────────────────────────────────────────────

func TestDiscoverManifests_ReturnsManifestsFromDisk(t *testing.T) {
	dir := t.TempDir()
	plugins.PluginDirOverride = dir
	defer func() { plugins.PluginDirOverride = "" }()

	// Write two plugin manifests.
	writeManifest(t, dir, "alpha", `{"name":"alpha","version":"1.0.0","author":"test","description":"Alpha plugin"}`)
	writeManifest(t, dir, "beta", `{"name":"beta","version":"2.0.0","author":"test","description":"Beta plugin"}`)

	manifests, err := plugins.DiscoverManifests()
	if err != nil {
		t.Fatalf("DiscoverManifests: %v", err)
	}
	if len(manifests) < 2 {
		t.Fatalf("expected at least 2 manifests, got %d", len(manifests))
	}

	found := map[string]bool{}
	for _, m := range manifests {
		found[m.Name] = true
	}
	for _, want := range []string{"alpha", "beta"} {
		if !found[want] {
			t.Errorf("expected manifest %q in results", want)
		}
	}
}

func TestDiscoverManifests_SkipsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	plugins.PluginDirOverride = dir
	defer func() { plugins.PluginDirOverride = "" }()

	writeManifest(t, dir, "good", `{"name":"good","version":"1.0.0","author":"test","description":"Good"}`)
	writeManifest(t, dir, "bad", `{not valid json`)

	manifests, err := plugins.DiscoverManifests()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, m := range manifests {
		if m.Name == "bad" {
			t.Error("malformed manifest should have been skipped")
		}
	}
}

func TestDiscoverManifests_EmptyDirReturnsNil(t *testing.T) {
	dir := t.TempDir()
	plugins.PluginDirOverride = filepath.Join(dir, "nonexistent")
	defer func() { plugins.PluginDirOverride = "" }()

	manifests, err := plugins.DiscoverManifests()
	if err != nil {
		t.Fatalf("unexpected error for missing plugin dir: %v", err)
	}
	if len(manifests) != 0 {
		t.Errorf("expected 0 manifests, got %d", len(manifests))
	}
}

// ── Apply ─────────────────────────────────────────────────────────────────────

func TestApply_TemplatePlugin_AddsToBuiltIn(t *testing.T) {
	name := uniqueName("tpl-apply")
	tmplName := "test-template-" + name
	p := &fakePlugin{
		name:    name,
		version: "1.0.0",
		tmplates: []templates.Template{
			{Name: tmplName, Description: "test", Dependencies: []string{"web"}},
		},
	}
	plugins.RegisterPlugin(p)

	plugins.Apply(nil)

	_, err := templates.Get(tmplName)
	if err != nil {
		t.Errorf("template %q should be registered after Apply, got error: %v", tmplName, err)
	}
}

func TestApply_HookPlugin_RegistersInPostgen(t *testing.T) {
	name := uniqueName("hook-apply")
	hookName := "test-hook-" + name
	h := &fakeHook{name: hookName}
	p := &fakePlugin{
		name:    name,
		version: "1.0.0",
		hooks:   []postgen.Hook{h},
	}
	plugins.RegisterPlugin(p)

	plugins.Apply(nil)

	got, err := postgen.Lookup(hookName)
	if err != nil {
		t.Errorf("hook %q should be registered after Apply, got error: %v", hookName, err)
	}
	if got.Name() != hookName {
		t.Errorf("wrong hook: want %q, got %q", hookName, got.Name())
	}
}

func TestApply_DependencyProvider_AppendsGroups(t *testing.T) {
	name := uniqueName("dep-apply")
	groupName := "TestGroup-" + name
	p := &fakePlugin{
		name:    name,
		version: "1.0.0",
		depGroups: []metadata.DependencyGroup{
			{
				Name: groupName,
				Values: []metadata.DependencyValue{
					{ID: "test-dep", Name: "Test Dep", Description: "A test dependency"},
				},
			},
		},
	}
	plugins.RegisterPlugin(p)

	meta := &metadata.Metadata{}
	plugins.Apply(meta)

	found := false
	for _, g := range meta.Dependencies.Values {
		if g.Name == groupName {
			found = true
			if len(g.Values) != 1 || g.Values[0].ID != "test-dep" {
				t.Errorf("dependency group contents wrong: %+v", g.Values)
			}
			break
		}
	}
	if !found {
		t.Errorf("dependency group %q not found in metadata after Apply", groupName)
	}
}

func TestApply_SkipsDisabledPlugins(t *testing.T) {
	name := uniqueName("disabled-apply")
	tmplName := "disabled-template-" + name
	p := &fakePlugin{
		name:    name,
		version: "1.0.0",
		tmplates: []templates.Template{
			{Name: tmplName, Description: "test", Dependencies: []string{"web"}},
		},
	}
	plugins.RegisterPlugin(p)
	_ = plugins.SetEnabled(name, false)

	plugins.Apply(nil)

	_, err := templates.Get(tmplName)
	if err == nil {
		t.Errorf("disabled plugin's template %q should NOT be registered", tmplName)
	}
}

func TestApply_Idempotent(t *testing.T) {
	name := uniqueName("idempotent")
	tmplName := "idempotent-template-" + name
	p := &fakePlugin{
		name:    name,
		version: "1.0.0",
		tmplates: []templates.Template{
			{Name: tmplName, Description: "test", Dependencies: []string{"web"}},
		},
	}
	plugins.RegisterPlugin(p)

	// Apply twice — should not panic or double-register.
	plugins.Apply(nil)
	plugins.Apply(nil) // idempotent: duplicate guard prevents panic
}

// ── Summary ───────────────────────────────────────────────────────────────────

func TestSummary(t *testing.T) {
	name := uniqueName("summary")
	p := &fakePlugin{
		name:    name,
		version: "1.0.0",
		tmplates: []templates.Template{
			{Name: "t1"}, {Name: "t2"},
		},
		hooks: []postgen.Hook{
			&fakeHook{name: "h1-" + name},
		},
		depGroups: []metadata.DependencyGroup{
			{Name: "G1", Values: []metadata.DependencyValue{
				{ID: "d1"}, {ID: "d2"}, {ID: "d3"},
			}},
		},
	}
	plugins.RegisterPlugin(p)

	s := plugins.Summary(p)
	if s.Templates != 2 {
		t.Errorf("expected 2 templates, got %d", s.Templates)
	}
	if s.Hooks != 1 {
		t.Errorf("expected 1 hook, got %d", s.Hooks)
	}
	if s.DependencyGroups != 1 {
		t.Errorf("expected 1 dep group, got %d", s.DependencyGroups)
	}
	if s.Dependencies != 3 {
		t.Errorf("expected 3 deps, got %d", s.Dependencies)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeManifest(t *testing.T, baseDir, name, content string) {
	t.Helper()
	dir := filepath.Join(baseDir, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(content), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}
