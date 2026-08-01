// Package config provides persistent user configuration for springx.
// Configuration is stored as YAML at ~/.config/springx/config.yaml (Linux/macOS)
// or %APPDATA%\springx\config.yaml (Windows).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/saireddy-shyamakura/springx/internal/validate"
)

// ConfigFilePathOverride can be set in unit tests to redirect config reads/writes
// to a temporary file path.
var ConfigFilePathOverride string

// Config holds user-defined defaults for springx project generation.
type Config struct {
	GroupID        string `yaml:"groupId"`
	ArtifactPrefix string `yaml:"artifactPrefix"`
	PackagePrefix  string `yaml:"packagePrefix"`
	JavaVersion    string `yaml:"javaVersion"`
	BuildTool      string `yaml:"buildTool"`
	Packaging      string `yaml:"packaging"`
	Language       string `yaml:"language"`
}

// DefaultConfig returns the baseline springx fallback configuration.
func DefaultConfig() Config {
	return Config{
		GroupID:     "com.example",
		JavaVersion: "21",
		BuildTool:   "maven-project",
		Packaging:   "jar",
		Language:    "java",
	}
}

// GetConfigPath returns the absolute path to the springx config.yaml file.
// On Linux/macOS: ~/.config/springx/config.yaml.
// On Windows: %APPDATA%/springx/config.yaml.
func GetConfigPath() (string, error) {
	if ConfigFilePathOverride != "" {
		return ConfigFilePathOverride, nil
	}

	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not resolve user config directory: %w", err)
	}

	return filepath.Join(userDir, "springx", "config.yaml"), nil
}

// Load reads configuration from file (if present) and applies environment variable overrides.
// Values read from the config file are validated; a config file that would
// inject into the download URL or the filesystem is rejected rather than
// silently trusted.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path, err := GetConfigPath()
	if err == nil {
		if data, readErr := os.ReadFile(path); readErr == nil {
			var fileCfg Config
			if unmarshalErr := yaml.Unmarshal(data, &fileCfg); unmarshalErr == nil {
				if err := fileCfg.validate(); err != nil {
					return nil, err
				}
				if fileCfg.GroupID != "" {
					cfg.GroupID = fileCfg.GroupID
				}
				if fileCfg.ArtifactPrefix != "" {
					cfg.ArtifactPrefix = fileCfg.ArtifactPrefix
				}
				if fileCfg.PackagePrefix != "" {
					cfg.PackagePrefix = fileCfg.PackagePrefix
				}
				if fileCfg.JavaVersion != "" {
					cfg.JavaVersion = fileCfg.JavaVersion
				}
				if fileCfg.BuildTool != "" {
					cfg.BuildTool = fileCfg.BuildTool
				}
				if fileCfg.Packaging != "" {
					cfg.Packaging = fileCfg.Packaging
				}
				if fileCfg.Language != "" {
					cfg.Language = fileCfg.Language
				}
			}
		}
	}

	// Environment variable overrides
	if val := os.Getenv("SPRINGX_GROUP_ID"); val != "" {
		if !validate.GroupIDValid(val) {
			return nil, fmt.Errorf("SPRINGX_GROUP_ID contains invalid characters: %q", val)
		}
		cfg.GroupID = val
	}
	if val := os.Getenv("SPRINGX_ARTIFACT_PREFIX"); val != "" {
		if !validate.ArtifactIDValid(val) {
			return nil, fmt.Errorf("SPRINGX_ARTIFACT_PREFIX contains invalid characters: %q", val)
		}
		cfg.ArtifactPrefix = val
	}
	if val := os.Getenv("SPRINGX_PACKAGE_PREFIX"); val != "" {
		if !validate.PackageNameValid(val) {
			return nil, fmt.Errorf("SPRINGX_PACKAGE_PREFIX is not a valid package prefix: %q", val)
		}
		cfg.PackagePrefix = val
	}
	if val := os.Getenv("SPRINGX_JAVA_VERSION"); val != "" {
		if !validate.JavaVersionValid(val) {
			return nil, fmt.Errorf("SPRINGX_JAVA_VERSION must be numeric: %q", val)
		}
		cfg.JavaVersion = val
	}
	if val := os.Getenv("SPRINGX_BUILD_TOOL"); val != "" {
		if !validate.BuildToolValid(val) {
			return nil, fmt.Errorf("SPRINGX_BUILD_TOOL contains invalid characters: %q", val)
		}
		cfg.BuildTool = val
	}
	if val := os.Getenv("SPRINGX_PACKAGING"); val != "" {
		if !validate.PackagingValid(val) {
			return nil, fmt.Errorf("SPRINGX_PACKAGING must be 'jar' or 'war': %q", val)
		}
		cfg.Packaging = val
	}
	if val := os.Getenv("SPRINGX_LANGUAGE"); val != "" {
		if !validate.LanguageValid(val) {
			return nil, fmt.Errorf("SPRINGX_LANGUAGE must be 'java', 'kotlin', or 'groovy': %q", val)
		}
		cfg.Language = val
	}

	return &cfg, nil
}

// validate checks a Config read from disk for unsafe values.
func (c *Config) validate() error {
	if c.GroupID != "" && !validate.GroupIDValid(c.GroupID) {
		return fmt.Errorf("config file: invalid groupId %q: use only letters, digits, '.', '_', '-' or ':'", c.GroupID)
	}
	if c.ArtifactPrefix != "" && !validate.ArtifactIDValid(c.ArtifactPrefix) {
		return fmt.Errorf("config file: invalid artifactPrefix %q: use only letters, digits, '.', '_', '-' or ':'", c.ArtifactPrefix)
	}
	if c.PackagePrefix != "" && !validate.PackageNameValid(c.PackagePrefix) {
		return fmt.Errorf("config file: invalid packagePrefix %q: must be a valid Java package prefix", c.PackagePrefix)
	}
	if c.JavaVersion != "" && !validate.JavaVersionValid(c.JavaVersion) {
		return fmt.Errorf("config file: invalid javaVersion %q: must be numeric", c.JavaVersion)
	}
	if c.BuildTool != "" && !validate.BuildToolValid(c.BuildTool) {
		return fmt.Errorf("config file: invalid buildTool %q: use only letters, digits, '.', '_', '-' or ':'", c.BuildTool)
	}
	if c.Packaging != "" && !validate.PackagingValid(c.Packaging) {
		return fmt.Errorf("config file: invalid packaging %q: must be 'jar' or 'war'", c.Packaging)
	}
	if c.Language != "" && !validate.LanguageValid(c.Language) {
		return fmt.Errorf("config file: invalid language %q: must be 'java', 'kotlin', or 'groovy'", c.Language)
	}
	return nil
}

// Save marshals the configuration to YAML and writes it to the config file path.
func Save(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Reset removes the configuration file if it exists.
func Reset() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}

	return nil
}

// Validate checks the configuration for obvious syntax or domain errors.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if cfg.Packaging != "" {
		p := strings.ToLower(cfg.Packaging)
		if p != "jar" && p != "war" {
			return fmt.Errorf("invalid packaging %q: must be 'jar' or 'war'", cfg.Packaging)
		}
	}

	if cfg.Language != "" {
		l := strings.ToLower(cfg.Language)
		if l != "java" && l != "kotlin" && l != "groovy" {
			return fmt.Errorf("invalid language %q: must be 'java', 'kotlin', or 'groovy'", cfg.Language)
		}
	}

	return nil
}
