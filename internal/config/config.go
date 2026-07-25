package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
// On Linux/macOS: ~/.config/springx/config.yaml
// On Windows: %APPDATA%/springx/config.yaml
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
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path, err := GetConfigPath()
	if err == nil {
		if data, readErr := os.ReadFile(path); readErr == nil {
			var fileCfg Config
			if unmarshalErr := yaml.Unmarshal(data, &fileCfg); unmarshalErr == nil {
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
		cfg.GroupID = val
	}
	if val := os.Getenv("SPRINGX_ARTIFACT_PREFIX"); val != "" {
		cfg.ArtifactPrefix = val
	}
	if val := os.Getenv("SPRINGX_PACKAGE_PREFIX"); val != "" {
		cfg.PackagePrefix = val
	}
	if val := os.Getenv("SPRINGX_JAVA_VERSION"); val != "" {
		cfg.JavaVersion = val
	}
	if val := os.Getenv("SPRINGX_BUILD_TOOL"); val != "" {
		cfg.BuildTool = val
	}
	if val := os.Getenv("SPRINGX_PACKAGING"); val != "" {
		cfg.Packaging = val
	}
	if val := os.Getenv("SPRINGX_LANGUAGE"); val != "" {
		cfg.Language = val
	}

	return &cfg, nil
}

// Save marshals the configuration to YAML and writes it to the config file path.
func Save(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
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
