package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ── Disk paths ────────────────────────────────────────────────────────────────

// PluginDirOverride can be set in tests to redirect discovery to a temp path.
var PluginDirOverride string

// PluginDir returns the canonical plugin discovery directory.
// Default: ~/.config/springx/plugins/.
func PluginDir() (string, error) {
	if PluginDirOverride != "" {
		return PluginDirOverride, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user config directory: %w", err)
	}
	return filepath.Join(base, "springx", "plugins"), nil
}

// disabledFile returns the path of the JSON file that persists the set of
// disabled plugin names.
func disabledFile() (string, error) {
	dir, err := PluginDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "disabled.json"), nil
}

// ── Manifest discovery ────────────────────────────────────────────────────────

// DiscoverManifests walks the plugin directory and returns one Manifest for
// every sub-directory that contains a valid plugin.json. Directories without
// a manifest are silently skipped so that partially installed plugins do not
// break the CLI.
//
// Note: discovery reads metadata only. It does not load or execute plugin
// code — that happens at compile time via blank imports in the host binary.
func DiscoverManifests() ([]Manifest, error) {
	dir, err := PluginDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil // no plugins directory is not an error
	}
	if err != nil {
		return nil, fmt.Errorf("reading plugin directory %s: %w", dir, err)
	}

	disabled, err := loadDisabled()
	if err != nil {
		disabled = map[string]bool{} // non-fatal
	}

	var manifests []Manifest
	for _, de := range entries {
		if !de.IsDir() {
			continue
		}
		manifestPath := filepath.Join(dir, de.Name(), "plugin.json")
		m, err := readManifest(manifestPath)
		if err != nil {
			continue // silently skip malformed manifests
		}
		// Normalise name — directory name wins when manifest name is empty.
		if m.Name == "" {
			m.Name = strings.ToLower(de.Name())
		}
		// Sync enable/disable state from the disabled set.
		if disabled[strings.ToLower(m.Name)] {
			_ = SetEnabled(m.Name, false)
		}
		manifests = append(manifests, m)
	}
	return manifests, nil
}

// readManifest reads and validates a plugin.json file.
func readManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}
	if m.Name == "" {
		return Manifest{}, fmt.Errorf("%s: manifest must include a non-empty 'name' field", path)
	}
	return m, nil
}

// WriteManifest serialises m to <pluginDir>/<name>/plugin.json, creating
// the directory if necessary. This is used by plugin scaffolding tools and
// the example plugin, not by the runtime loader.
func WriteManifest(m Manifest) error {
	dir, err := PluginDir()
	if err != nil {
		return err
	}
	pluginDir := filepath.Join(dir, strings.ToLower(m.Name))
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644)
}

// ── Enable / disable persistence ──────────────────────────────────────────────

// loadDisabled reads the set of disabled plugin names from disk.
// Returns an empty map when the file does not exist.
func loadDisabled() (map[string]bool, error) {
	path, err := disabledFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]bool{}, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(names))
	for _, n := range names {
		out[strings.ToLower(n)] = true
	}
	return out, nil
}

// saveDisabled writes the current disabled set to disk.
func saveDisabled(disabled map[string]bool) error {
	path, err := disabledFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	names := make([]string, 0, len(disabled))
	for n, isDisabled := range disabled {
		if isDisabled {
			names = append(names, n)
		}
	}
	data, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// PersistEnabled marks name as enabled on disk.
func PersistEnabled(name string) error {
	disabled, err := loadDisabled()
	if err != nil {
		disabled = map[string]bool{}
	}
	delete(disabled, strings.ToLower(name))
	return saveDisabled(disabled)
}

// PersistDisabled marks name as disabled on disk.
func PersistDisabled(name string) error {
	disabled, err := loadDisabled()
	if err != nil {
		disabled = map[string]bool{}
	}
	disabled[strings.ToLower(name)] = true
	return saveDisabled(disabled)
}

// LoadDisabledIntoRegistry reads the on-disk disabled set and applies it to
// the in-process registry. Call this once at startup, after all plugins have
// been blank-imported (and thus registered).
func LoadDisabledIntoRegistry() {
	disabled, err := loadDisabled()
	if err != nil {
		return // non-fatal; all plugins remain enabled
	}
	for name := range disabled {
		_ = SetEnabled(name, false)
	}
}
