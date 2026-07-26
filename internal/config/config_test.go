package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/config"
)

func setupTempConfig(t *testing.T) string {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	config.ConfigFilePathOverride = path
	t.Cleanup(func() {
		config.ConfigFilePathOverride = ""
	})
	return path
}

func TestLoad_Defaults(t *testing.T) {
	setupTempConfig(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	def := config.DefaultConfig()
	if cfg.GroupID != def.GroupID {
		t.Errorf("expected GroupID %q, got %q", def.GroupID, cfg.GroupID)
	}
	if cfg.JavaVersion != def.JavaVersion {
		t.Errorf("expected JavaVersion %q, got %q", def.JavaVersion, cfg.JavaVersion)
	}
}

func TestSaveAndLoad(t *testing.T) {
	path := setupTempConfig(t)

	customCfg := &config.Config{
		GroupID:        "com.saireddy",
		PackagePrefix:  "com.saireddy",
		ArtifactPrefix: "microservice-",
		JavaVersion:    "21",
		BuildTool:      "maven",
		Packaging:      "jar",
		Language:       "java",
	}

	if err := config.Save(customCfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected config file to exist at %s", path)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.GroupID != "com.saireddy" {
		t.Errorf("expected GroupID 'com.saireddy', got %q", loaded.GroupID)
	}
	if loaded.ArtifactPrefix != "microservice-" {
		t.Errorf("expected ArtifactPrefix 'microservice-', got %q", loaded.ArtifactPrefix)
	}
}

func TestEnvVarOverrides(t *testing.T) {
	setupTempConfig(t)

	// Set env vars
	if err := os.Setenv("SPRINGX_GROUP_ID", "org.override"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	if err := os.Setenv("SPRINGX_JAVA_VERSION", "25"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	t.Cleanup(func() {
		os.Unsetenv("SPRINGX_GROUP_ID")     //nolint:errcheck
		os.Unsetenv("SPRINGX_JAVA_VERSION") //nolint:errcheck
	})

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.GroupID != "org.override" {
		t.Errorf("expected GroupID override 'org.override', got %q", loaded.GroupID)
	}
	if loaded.JavaVersion != "25" {
		t.Errorf("expected JavaVersion override '25', got %q", loaded.JavaVersion)
	}
}

func TestValidate(t *testing.T) {
	valid := &config.Config{
		Packaging: "jar",
		Language:  "java",
	}
	if err := config.Validate(valid); err != nil {
		t.Errorf("expected valid config to pass, got: %v", err)
	}

	invalidPkg := &config.Config{
		Packaging: "invalid-packaging",
	}
	if err := config.Validate(invalidPkg); err == nil {
		t.Error("expected error for invalid packaging, got nil")
	}

	invalidLang := &config.Config{
		Language: "python",
	}
	if err := config.Validate(invalidLang); err == nil {
		t.Error("expected error for invalid language, got nil")
	}
}

func TestReset(t *testing.T) {
	path := setupTempConfig(t)

	cfg := &config.Config{GroupID: "com.test"}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if err := config.Reset(); err != nil {
		t.Fatalf("failed to reset config: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected config file %s to be deleted", path)
	}
}
