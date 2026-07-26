# Contributing to springx

Thank you for your interest in contributing. This document covers everything
you need to get started.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Building](#building)
- [Testing](#testing)
- [Linting](#linting)
- [Coding Style](#coding-style)
- [Commit Messages](#commit-messages)
- [Opening a Pull Request](#opening-a-pull-request)
- [Reporting Bugs](#reporting-bugs)
- [Requesting Features](#requesting-features)

---

## Code of Conduct

Be respectful. Constructive criticism is welcome; personal attacks are not.

---

## Prerequisites

| Tool | Minimum version | Install |
|------|----------------|---------|
| Go | 1.24 | https://go.dev/dl/ |
| golangci-lint | 2.x | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| goreleaser | 2.x | https://goreleaser.com/install/ |
| git | any | system package manager |

---

## Project Structure

```
springx/
├── cmd/                    # Cobra CLI commands (root, new, config, template, plugin, version, dependencies)
├── internal/
│   ├── config/             # Persistent user configuration (~/.config/springx/config.yaml)
│   ├── extract/            # Zip Slip-safe archive extraction
│   ├── initializr/         # HTTP client for Spring Initializr download API
│   ├── metadata/           # HTTP client for Spring Initializr metadata API
│   ├── plugins/            # Plugin registry, manifest loader, enable/disable persistence
│   ├── postgen/            # Post-generation hook registry and runner
│   │   └── hooks/          # Built-in hook implementations (git, docker, vscode, etc.)
│   ├── prompt/             # Interactive text prompts (project name, group ID, etc.)
│   ├── templates/          # Built-in project template presets
│   └── ui/                 # Bubble Tea TUI: dependency picker and progress pipeline
├── plugins/
│   └── examples/aws/       # Example third-party plugin
├── main.go                 # Entry point; signal handling
├── Makefile                # Developer targets
├── .golangci.yml           # Linter configuration (golangci-lint v2)
└── .goreleaser.yaml        # Release pipeline (GoReleaser v2)
```

---

## Building

```bash
# Build the binary for the current platform
make build          # output: ./springx

# Install to $GOPATH/bin
make install

# Build a GoReleaser snapshot (all platforms, no publish)
make snapshot
```

To inject version metadata manually:

```bash
go build -ldflags "-X github.com/saireddy-shyamakura/springx/cmd.Version=v1.0.0 \
  -X github.com/saireddy-shyamakura/springx/cmd.Commit=$(git rev-parse --short HEAD) \
  -X github.com/saireddy-shyamakura/springx/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o springx .
```

---

## Testing

```bash
# Run the full test suite
make test

# Run with verbose output
make test-verbose

# Run with the race detector
make test-race

# Generate an HTML coverage report
make coverage

# Run benchmarks
make bench
```

Tests live next to the packages they test in `_test.go` files using the
`package foo_test` external-test pattern. The `internal/prompt` package is
excluded from automated tests because it requires an interactive TTY.

---

## Linting

```bash
# Run golangci-lint (requires v2.x)
make lint

# Run all quality checks in one command
make check          # gofmt + go vet + tests + lint
```

To install golangci-lint:

```bash
make dev-setup
```

All PRs must pass `make check` before review.

---

## Coding Style

- Follow standard Go idioms and the [Effective Go](https://go.dev/doc/effective_go) guide.
- All exported symbols must have a doc comment ending with a period.
- Error strings must be lower-case and must not end with punctuation (Go convention).
- Use `fmt.Errorf("...: %w", err)` to wrap errors; never discard them silently.
- HTTP requests must use `http.NewRequestWithContext` with a `context.Context`.
- Do not add global variables unless strictly necessary.
- Bubble Tea model methods must be value receivers (not pointer receivers) — this is the Bubble Tea convention that enables safe message passing.
- All filesystem operations in hooks must be relative to `projectPath`.
- Format code with `gofmt` before committing (`make fmt`).

---

## Commit Messages

springx uses [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short summary>

<optional body>
```

Common types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `perf`, `ci`.

Examples:

```
feat(hooks): add devcontainer hook for Java projects
fix(ui): restore cursor position after clearing search
docs(readme): add Scoop installation instructions
chore(deps): bump charmbracelet/bubbletea to v1.4.0
```

---

## Opening a Pull Request

1. Fork the repository and create a branch from `main`.
2. Make your changes and add tests for any new behaviour.
3. Run `make check` — all checks must pass.
4. Open a PR against `main` using the PR template.
5. Keep PRs focused: one concern per PR.

A maintainer will review your PR. Please respond to review comments promptly.

---

## Reporting Bugs

Open a [Bug Report](https://github.com/saireddy-shyamakura/springx/issues/new?template=bug_report.md)
and fill in the template. Include your OS, Go version, springx version
(`springx version`), and exact steps to reproduce.

---

## Requesting Features

Open a [Feature Request](https://github.com/saireddy-shyamakura/springx/issues/new?template=feature_request.md).
Describe the problem you are trying to solve and your proposed solution.
