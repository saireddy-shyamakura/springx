package postgen_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/saireddy-shyamakura/springx/internal/postgen"
	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

// ── Test hook implementations ─────────────────────────────────────────────────

type okHook struct{ name string }

func (h *okHook) Name() string { return h.name }
func (h *okHook) Run(_ string, _ *prompt.ProjectConfig) error { return nil }

type failHook struct{ name string }

func (h *failHook) Name() string { return h.name }
func (h *failHook) Run(_ string, _ *prompt.ProjectConfig) error {
	return errors.New("hook intentionally failed")
}

// ── Register / Lookup / Names ─────────────────────────────────────────────────

func TestRegister_And_Lookup(t *testing.T) {
	postgen.Register(&okHook{name: "test-ok-1"})
	h, err := postgen.Lookup("test-ok-1")
	if err != nil {
		t.Fatalf("Lookup('test-ok-1') returned error: %v", err)
	}
	if h.Name() != "test-ok-1" {
		t.Errorf("expected name 'test-ok-1', got %q", h.Name())
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	postgen.Register(&okHook{name: "test-case-hook"})
	for _, name := range []string{"TEST-CASE-HOOK", "Test-Case-Hook"} {
		if _, err := postgen.Lookup(name); err != nil {
			t.Errorf("Lookup(%q) should be case-insensitive: %v", name, err)
		}
	}
}

func TestLookup_UnknownReturnsError(t *testing.T) {
	_, err := postgen.Lookup("definitely-does-not-exist-xyz")
	if err == nil {
		t.Error("expected error for unknown hook, got nil")
	}
	if !strings.Contains(err.Error(), "available") {
		t.Errorf("error should list available hooks, got: %v", err)
	}
}

func TestNames_ReturnsSortedList(t *testing.T) {
	// Names() should return alphabetically sorted names.
	names := postgen.Names()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("Names() not sorted: %v", names)
			break
		}
	}
}

// ── ResolveHooks ──────────────────────────────────────────────────────────────

func TestResolveHooks_EmptyReturnsAll(t *testing.T) {
	all := postgen.All()
	resolved, err := postgen.ResolveHooks(nil)
	if err != nil {
		t.Fatalf("ResolveHooks(nil) error: %v", err)
	}
	if len(resolved) != len(all) {
		t.Errorf("expected %d hooks, got %d", len(all), len(resolved))
	}
}

func TestResolveHooks_SpecificNames(t *testing.T) {
	postgen.Register(&okHook{name: "resolve-a"})
	postgen.Register(&okHook{name: "resolve-b"})

	resolved, err := postgen.ResolveHooks([]string{"resolve-a", "resolve-b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved) != 2 {
		t.Errorf("expected 2 resolved hooks, got %d", len(resolved))
	}
}

func TestResolveHooks_UnknownNameReturnsError(t *testing.T) {
	_, err := postgen.ResolveHooks([]string{"unknown-hook-xyz"})
	if err == nil {
		t.Error("expected error for unknown hook name, got nil")
	}
}

// ── RunHooks ──────────────────────────────────────────────────────────────────

func TestRunHooks_AllSucceed(t *testing.T) {
	hooks := []postgen.Hook{
		&okHook{name: "run-ok-a"},
		&okHook{name: "run-ok-b"},
	}
	results, err := postgen.RunHooks(postgen.RunOptions{
		ProjectPath: t.TempDir(),
		Config:      &prompt.ProjectConfig{},
		Hooks:       hooks,
		Out:         io.Discard,
	})
	if err != nil {
		t.Errorf("expected no error when all hooks succeed, got: %v", err)
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("hook %q reported error: %v", r.Hook.Name(), r.Err)
		}
	}
}

func TestRunHooks_OneFailureContinues(t *testing.T) {
	hooks := []postgen.Hook{
		&okHook{name: "run-before-fail"},
		&failHook{name: "run-fail-middle"},
		&okHook{name: "run-after-fail"},
	}
	results, err := postgen.RunHooks(postgen.RunOptions{
		ProjectPath: t.TempDir(),
		Config:      &prompt.ProjectConfig{},
		Hooks:       hooks,
		Out:         io.Discard,
	})
	// Overall error should be non-nil (one hook failed).
	if err == nil {
		t.Error("expected combined error when a hook fails, got nil")
	}
	// All three hooks should have a result entry.
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
	// First and last should have succeeded.
	if results[0].Err != nil {
		t.Errorf("first hook should have succeeded: %v", results[0].Err)
	}
	if results[2].Err != nil {
		t.Errorf("third hook should have succeeded: %v", results[2].Err)
	}
	// Middle should have failed.
	if results[1].Err == nil {
		t.Error("middle hook should have failed")
	}
}

func TestRunHooks_EmptyList(t *testing.T) {
	results, err := postgen.RunHooks(postgen.RunOptions{
		ProjectPath: t.TempDir(),
		Config:      &prompt.ProjectConfig{},
		Hooks:       nil,
		Out:         io.Discard,
	})
	if err != nil {
		t.Errorf("empty hook list should not error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
