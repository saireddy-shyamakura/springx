// Package postgen provides a hook-based post-generation automation system.
// Each Hook performs a single, independent action on a freshly generated
// Spring Boot project (e.g. git init, Dockerfile creation, VS Code settings).
//
// Design goals:
//   - Adding a new hook requires changing only one file (the registration site).
//   - Each hook is independent; a failure in one does not abort the others.
//   - Progress is printed to stdout as each hook completes or fails.
//   - The caller decides which hooks to run by name; unknown names are rejected.
package postgen

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/saireddy-shyamakura/springx/internal/prompt"
)

// Hook is implemented by every post-generation action.
type Hook interface {
	// Name returns the canonical identifier used on the CLI (e.g. "git", "docker").
	Name() string

	// Run executes the hook against projectPath using the supplied project
	// configuration. It must be safe to call on any working directory; all
	// filesystem operations must be relative to projectPath.
	Run(projectPath string, cfg *prompt.ProjectConfig) error
}

// registry maps canonical hook names to their Hook implementations.
// All built-in hooks are registered via Register() calls in init() functions
// inside the hooks sub-package, keeping this file free of concrete imports.
var registry = map[string]Hook{}

// Register adds h to the global registry. It panics on duplicate names so
// mis-registrations surface immediately at startup rather than at runtime.
func Register(h Hook) {
	name := strings.ToLower(h.Name())
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("postgen: duplicate hook registration for %q", name))
	}
	registry[name] = h
}

// All returns every registered hook in a stable, alphabetically sorted order.
func All() []Hook {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	// Simple insertion sort — registry is tiny (<20 entries).
	for i := 1; i < len(names); i++ {
		key := names[i]
		j := i - 1
		for j >= 0 && names[j] > key {
			names[j+1] = names[j]
			j--
		}
		names[j+1] = key
	}
	hooks := make([]Hook, 0, len(names))
	for _, n := range names {
		hooks = append(hooks, registry[n])
	}
	return hooks
}

// Lookup returns the Hook registered under name (case-insensitive).
// Returns an error if the name is not found.
func Lookup(name string) (Hook, error) {
	h, ok := registry[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf(
			"unknown hook %q — available: %s",
			name,
			strings.Join(Names(), ", "),
		)
	}
	return h, nil
}

// Names returns all registered hook names in sorted order.
func Names() []string {
	all := All()
	names := make([]string, len(all))
	for i, h := range all {
		names[i] = h.Name()
	}
	return names
}

// HookResult records the outcome of a single hook execution.
type HookResult struct {
	Hook Hook
	Err  error
}

// RunOptions controls how RunHooks behaves.
type RunOptions struct {
	// Out is where progress lines are written. Defaults to os.Stdout.
	Out io.Writer

	// Config is the project configuration used during generation.
	Config *prompt.ProjectConfig

	// ProjectPath is the root directory of the generated project.
	ProjectPath string

	// Hooks is the ordered list of hooks to execute.
	Hooks []Hook
}

// RunHooks executes each hook in order, printing a progress line after each
// one. A failing hook is recorded but execution continues. The returned slice
// contains one entry per hook (in execution order); entries with a nil Err
// succeeded. If any hook failed the returned error is a combined summary.
func RunHooks(opts RunOptions) ([]HookResult, error) {
	out := opts.Out
	if out == nil {
		out = os.Stdout
	}

	results := make([]HookResult, 0, len(opts.Hooks))

	for _, h := range opts.Hooks {
		err := h.Run(opts.ProjectPath, opts.Config)
		results = append(results, HookResult{Hook: h, Err: err})
		if err != nil {
			fmt.Fprintf(out, "  ✗ %-20s — %v\n", h.Name(), err)
		} else {
			fmt.Fprintf(out, "  ✔ %-20s\n", h.Name())
		}
	}

	return results, summariseErrors(results)
}

// summariseErrors returns a combined error if any hook failed, or nil.
func summariseErrors(results []HookResult) error {
	var errs []error
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, fmt.Errorf("hook %q: %w", r.Hook.Name(), r.Err))
		}
	}
	return errors.Join(errs...)
}

// ResolveHooks converts a slice of hook names into their Hook implementations.
// Passing an empty/nil slice returns every registered hook (default behavior).
func ResolveHooks(names []string) ([]Hook, error) {
	if len(names) == 0 {
		return All(), nil
	}
	hooks := make([]Hook, 0, len(names))
	for _, n := range names {
		h, err := Lookup(n)
		if err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}
