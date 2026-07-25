package postgen_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"

	// Side-effect import: registers all built-in hooks.
	_ "github.com/saireddy-shyamakura/springx/internal/postgen/hooks"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// stubHook is a controllable Hook used in unit tests.
type stubHook struct {
	name    string
	callSeq *[]string // pointer so multiple stubs share the same sequence slice
	err     error     // returned by Run if non-nil
}

func (s *stubHook) Name() string { return s.name }
func (s *stubHook) Run(_ string, _ *prompt.ProjectConfig) error {
	*s.callSeq = append(*s.callSeq, s.name)
	return s.err
}

func defaultCfg() *prompt.ProjectConfig {
	return &prompt.ProjectConfig{
		ProjectName: "demo",
		GroupID:     "com.example",
		ArtifactID:  "demo",
		JavaVersion: "21",
		BuildTool:   "maven-project",
		Packaging:   "jar",
	}
}

// ── registry tests ────────────────────────────────────────────────────────────

func TestBuiltInHooksAreRegistered(t *testing.T) {
	names := postgen.Names()
	if len(names) == 0 {
		t.Fatal("expected at least one registered hook, got none")
	}

	want := []string{
		"compose",
		"devcontainer",
		"docker",
		"git",
		"gitignore",
		"readme",
		"vscode",
		"wrapper",
	}
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	for _, w := range want {
		if !nameSet[w] {
			t.Errorf("expected built-in hook %q to be registered", w)
		}
	}
}

func TestAll_ReturnsSortedOrder(t *testing.T) {
	all := postgen.All()
	for i := 1; i < len(all); i++ {
		if all[i-1].Name() > all[i].Name() {
			t.Errorf("All() is not sorted: %q > %q", all[i-1].Name(), all[i].Name())
		}
	}
}

func TestLookup_KnownHook(t *testing.T) {
	h, err := postgen.Lookup("git")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h.Name() != "git" {
		t.Errorf("expected hook name %q, got %q", "git", h.Name())
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	_, err := postgen.Lookup("GIT")
	if err != nil {
		t.Fatalf("Lookup should be case-insensitive, got error: %v", err)
	}
}

func TestLookup_UnknownHook(t *testing.T) {
	_, err := postgen.Lookup("does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown hook, got nil")
	}
}

func TestResolveHooks_EmptyReturnsAll(t *testing.T) {
	hooks, err := postgen.ResolveHooks(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	all := postgen.All()
	if len(hooks) != len(all) {
		t.Errorf("expected %d hooks, got %d", len(all), len(hooks))
	}
}

func TestResolveHooks_ByName(t *testing.T) {
	hooks, err := postgen.ResolveHooks([]string{"git", "docker"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hooks))
	}
}

func TestResolveHooks_UnknownName(t *testing.T) {
	_, err := postgen.ResolveHooks([]string{"git", "phantom"})
	if err == nil {
		t.Fatal("expected error for unknown hook name")
	}
}

// ── RunHooks: execution order ─────────────────────────────────────────────────

func TestRunHooks_ExecutionOrder(t *testing.T) {
	seq := &[]string{}
	hooks := []postgen.Hook{
		&stubHook{name: "alpha", callSeq: seq},
		&stubHook{name: "beta", callSeq: seq},
		&stubHook{name: "gamma", callSeq: seq},
	}

	var buf bytes.Buffer
	_, err := postgen.RunHooks(postgen.RunOptions{
		ProjectPath: t.TempDir(),
		Config:      defaultCfg(),
		Hooks:       hooks,
		Out:         &buf,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"alpha", "beta", "gamma"}
	for i, got := range *seq {
		if got != want[i] {
			t.Errorf("execution order[%d]: want %q, got %q", i, want[i], got)
		}
	}
}

// ── RunHooks: failure isolation ───────────────────────────────────────────────

func TestRunHooks_FailureDoesNotAbortRemaining(t *testing.T) {
	seq := &[]string{}
	hooks := []postgen.Hook{
		&stubHook{name: "first", callSeq: seq},
		&stubHook{name: "boom", callSeq: seq, err: errors.New("simulated failure")},
		&stubHook{name: "third", callSeq: seq}, // must still run
	}

	var buf bytes.Buffer
	results, _ := postgen.RunHooks(postgen.RunOptions{
		ProjectPath: t.TempDir(),
		Config:      defaultCfg(),
		Hooks:       hooks,
		Out:         &buf,
	})

	// All three hooks must have been called.
	if len(*seq) != 3 {
		t.Errorf("expected 3 hooks executed, got %d (%v)", len(*seq), *seq)
	}

	// Only "boom" should have failed.
	failCount := 0
	for _, r := range results {
		if r.Err != nil {
			failCount++
			if r.Hook.Name() != "boom" {
				t.Errorf("unexpected failure in hook %q", r.Hook.Name())
			}
		}
	}
	if failCount != 1 {
		t.Errorf("expected 1 failure, got %d", failCount)
	}
}

func TestRunHooks_CombinedErrorOnAnyFailure(t *testing.T) {
	seq := &[]string{}
	hooks := []postgen.Hook{
		&stubHook{name: "a", callSeq: seq, err: errors.New("err-a")},
		&stubHook{name: "b", callSeq: seq, err: errors.New("err-b")},
	}

	var buf bytes.Buffer
	_, err := postgen.RunHooks(postgen.RunOptions{
		ProjectPath: t.TempDir(),
		Config:      defaultCfg(),
		Hooks:       hooks,
		Out:         &buf,
	})
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(err.Error(), "err-a") || !strings.Contains(err.Error(), "err-b") {
		t.Errorf("combined error should mention both failures, got: %v", err)
	}
}

func TestRunHooks_ProgressOutput(t *testing.T) {
	seq := &[]string{}
	hooks := []postgen.Hook{
		&stubHook{name: "ok", callSeq: seq},
		&stubHook{name: "fail", callSeq: seq, err: errors.New("oops")},
	}

	var buf bytes.Buffer
	postgen.RunHooks(postgen.RunOptions{ //nolint:errcheck
		ProjectPath: t.TempDir(),
		Config:      defaultCfg(),
		Hooks:       hooks,
		Out:         &buf,
	})

	out := buf.String()
	if !strings.Contains(out, "✔") {
		t.Error("expected ✔ for successful hook")
	}
	if !strings.Contains(out, "✗") {
		t.Error("expected ✗ for failed hook")
	}
}

// ── Individual hook unit tests ────────────────────────────────────────────────

func TestGitignoreHook_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("gitignore")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Run(dir, defaultCfg()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf(".gitignore not created: %v", err)
	}
	if !strings.Contains(string(data), "target/") {
		t.Error("expected 'target/' in generated .gitignore")
	}
}

func TestGitignoreHook_Idempotent(t *testing.T) {
	dir := t.TempDir()
	h, _ := postgen.Lookup("gitignore")

	_ = h.Run(dir, defaultCfg())
	firstData, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))

	_ = h.Run(dir, defaultCfg()) // second run
	secondData, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))

	if string(firstData) != string(secondData) {
		t.Error("gitignore hook is not idempotent: content changed on second run")
	}
}

func TestReadmeHook_CreatesAndAppends(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("readme")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Run(dir, defaultCfg()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatalf("README.md not created: %v", err)
	}
	if !strings.Contains(string(data), "springx-generated") {
		t.Error("expected springx-generated section in README.md")
	}
}

func TestReadmeHook_Idempotent(t *testing.T) {
	dir := t.TempDir()
	h, _ := postgen.Lookup("readme")

	_ = h.Run(dir, defaultCfg())
	first, _ := os.ReadFile(filepath.Join(dir, "README.md"))
	_ = h.Run(dir, defaultCfg())
	second, _ := os.ReadFile(filepath.Join(dir, "README.md"))

	if string(first) != string(second) {
		t.Error("readme hook is not idempotent: content changed on second run")
	}
}

func TestVscodeHook_CreatesExtensionsJSON(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("vscode")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Run(dir, defaultCfg()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".vscode", "extensions.json"))
	if err != nil {
		t.Fatalf("extensions.json not created: %v", err)
	}
	if !strings.Contains(string(data), "vscjava.vscode-java-pack") {
		t.Error("expected vscjava.vscode-java-pack in extensions.json")
	}
}

func TestDevcontainerHook_CreatesDevcontainerJSON(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("devcontainer")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Run(dir, defaultCfg()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, ".devcontainer", "devcontainer.json"))
	if err != nil {
		t.Fatalf("devcontainer.json not created: %v", err)
	}
	if !strings.Contains(string(data), "8080") {
		t.Error("expected port 8080 in devcontainer.json")
	}
}

func TestDockerHook_MavenDockerfile(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("docker")
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultCfg()
	cfg.BuildTool = "maven-project"
	if err := h.Run(dir, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("Dockerfile not created: %v", err)
	}
	if !strings.Contains(string(data), "mvnw") {
		t.Error("expected mvnw in Maven Dockerfile")
	}
}

func TestDockerHook_GradleDockerfile(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("docker")
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultCfg()
	cfg.BuildTool = "gradle-project"
	if err := h.Run(dir, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		t.Fatalf("Dockerfile not created: %v", err)
	}
	if !strings.Contains(string(data), "gradlew") {
		t.Error("expected gradlew in Gradle Dockerfile")
	}
}

func TestComposeHook_PostgresAndRedis(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("compose")
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultCfg()
	cfg.Dependencies = []string{"postgresql", "redis"}
	if err := h.Run(dir, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("docker-compose.yml not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "postgres") {
		t.Error("expected postgres service in docker-compose.yml")
	}
	if !strings.Contains(content, "redis") {
		t.Error("expected redis service in docker-compose.yml")
	}
}

func TestComposeHook_NoDepsSkipsFile(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("compose")
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultCfg()
	cfg.Dependencies = []string{"web", "actuator"}
	if err := h.Run(dir, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Error("docker-compose.yml should not be created when no infra deps are selected")
	}
}

func TestWrapperHook_MissingMvnw(t *testing.T) {
	dir := t.TempDir()
	h, err := postgen.Lookup("wrapper")
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultCfg()
	cfg.BuildTool = "maven-project"
	// No mvnw in the temp dir — hook should return an error.
	if err := h.Run(dir, cfg); err == nil {
		t.Error("expected error when mvnw is missing, got nil")
	}
}

func TestWrapperHook_PresentMvnw(t *testing.T) {
	dir := t.TempDir()
	// Simulate Spring Initializr placing mvnw.
	if err := os.WriteFile(filepath.Join(dir, "mvnw"), []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatal(err)
	}
	h, _ := postgen.Lookup("wrapper")
	cfg := defaultCfg()
	cfg.BuildTool = "maven-project"
	if err := h.Run(dir, cfg); err != nil {
		t.Errorf("unexpected error when mvnw is present: %v", err)
	}
}
