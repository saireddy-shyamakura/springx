# springx

> A fast, interactive Spring Boot project generator for the terminal.

[![CI](https://github.com/saireddy-shyamakura/springx/actions/workflows/ci.yml/badge.svg)](https://github.com/saireddy-shyamakura/springx/actions/workflows/ci.yml)
[![Release](https://github.com/saireddy-shyamakura/springx/actions/workflows/release.yml/badge.svg)](https://github.com/saireddy-shyamakura/springx/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/saireddy-shyamakura/springx)](https://goreportcard.com/report/github.com/saireddy-shyamakura/springx)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

springx wraps [Spring Initializr](https://start.spring.io) in a professional
terminal UI. It fetches live metadata, lets you browse and search all available
dependencies in a three-panel dashboard, and runs post-generation automation
(git init, Dockerfile, VS Code settings, and more) immediately after creating
your project.

---

## Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Terminal UI](#terminal-ui)
- [Features](#features)
- [Templates](#templates)
- [Configuration](#configuration)
- [Post-Generation Hooks](#post-generation-hooks)
- [Plugin System](#plugin-system)
- [CLI Reference](#cli-reference)
- [Troubleshooting](#troubleshooting)
- [FAQ](#faq)
- [Contributing](#contributing)
- [License](#license)

---

## Installation

### go install (recommended)

Requires Go 1.21 or later.

```bash
go install github.com/saireddy-shyamakura/springx@latest
```

### Homebrew (macOS / Linux)

```bash
brew install saireddy-shyamakura/tap/springx
```

### Scoop (Windows)

```powershell
scoop bucket add springx https://github.com/saireddy-shyamakura/scoop-bucket
scoop install springx
```

### Pre-built binaries

Download the latest release archive for your platform from the
[Releases page](https://github.com/saireddy-shyamakura/springx/releases),
extract it, and place the `springx` binary somewhere on your `$PATH`.

| Platform       | Archive                                      |
|----------------|----------------------------------------------|
| Linux amd64    | `springx_<version>_linux_amd64.tar.gz`       |
| Linux arm64    | `springx_<version>_linux_arm64.tar.gz`       |
| macOS amd64    | `springx_<version>_darwin_amd64.tar.gz`      |
| macOS arm64    | `springx_<version>_darwin_arm64.tar.gz`      |
| Windows amd64  | `springx_<version>_windows_amd64.zip`        |

SHA-256 checksums are provided in `springx_<version>_checksums.txt`.

### Build from source

```bash
git clone https://github.com/saireddy-shyamakura/springx.git
cd springx
make build          # produces ./springx
make install        # installs to $GOPATH/bin
```

---

## Quick Start

```bash
# Launch the interactive project wizard
springx new

# Bootstrap a REST API project immediately
springx new --template rest-api

# Skip post-generation hooks
springx new --no-hooks

# Check your version
springx version
```

The wizard walks you through:

1. **Project name** — the output directory name
2. **Group ID** — Maven/Gradle group (e.g. `com.example`)
3. **Artifact ID** — Maven/Gradle artifact (defaults to project name)
4. **Package name** — base Java package
5. **Build tool** — Maven or Gradle
6. **Packaging** — JAR or WAR
7. **Java version** — pulled live from start.spring.io
8. **Dependencies** — interactive TUI picker (see below)

After selection a confirmation screen summarises everything before generation
begins. Press **F5** or navigate to **Y — Generate** to proceed.

---

## Terminal UI

The dependency picker is a full-screen, three-panel dashboard:

```
 springx                                          Spring Boot 3.5.4
 ──────────────────────────────────────────────────────────────────
 Search                                                      Ctrl+F
 ╔══════════════════════════════════╗
 ║ ❯ postgres                       ║  Found 3 dependencies  Esc to clear
 ╚══════════════════════════════════╝
╔══════════════════╗╭────────────────────────────────────╮╭──────────────────╮
║ Groups           ║│ Dependencies                       ││ Selected (3)     │
║                  ││   Data                             ││ Web              │
║   Web            ││   ─────────────────────────────── ││  ✓ Spring Web    │
║   Messaging      ││ ❯ [x] PostgreSQL Driver            ││ Data             │
║ ❯ Data           ││   [ ] Spring Data JPA              ││  ✓ PostgreSQL    │
║   Security       ││   [ ] Spring Data R2DBC            ││  ✓ Flyway        │
║   AI             ││                                    ││                  │
╚══════════════════╝╰────────────────────────────────────╯╰──────────────────╯
 Template: jpa   Boot: 3.5.4   Java: 21   Selected: 3   Filter: postgres
 ──────────────────────────────────────────────────────────────────
 ↑↓ Move  ←→ Panels  Tab Group  Space Toggle  / Search  Esc Clear  F5 Generate
```

### Keyboard reference

#### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `Home` / `g` | Jump to first dependency |
| `End` / `G` | Jump to last dependency |
| `PgUp` / `Ctrl+U` | Page up (8 rows) |
| `PgDn` / `Ctrl+D` | Page down (8 rows) |
| `Tab` | Jump to next dependency group |
| `Shift+Tab` | Jump to previous dependency group |
| `←` / `h` | Focus previous panel |
| `→` / `l` | Focus next panel |

#### Selection

| Key | Action |
|-----|--------|
| `Space` | Toggle dependency on/off |
| `F5` | Open confirmation screen |

#### Search

| Key | Action |
|-----|--------|
| `/` or `Ctrl+F` | Open search box |
| `Esc` | Clear search, restore cursor position |
| `Ctrl+Backspace` | Clear entire query |
| `Enter` (in search) | Exit search box, keep filter active |

#### General

| Key | Action |
|-----|--------|
| `?` | Toggle keyboard reference overlay |
| `q` / `Ctrl+C` | Quit without generating |

### Confirmation screen

**F5** opens the confirmation screen showing all project settings and the
selected dependencies grouped by category. Use **Tab** or **← →** to move
between the **Y — Generate** and **N — Cancel** buttons. Press **Enter** to
activate the focused button. Press **Esc** or **n** to go back.

Generation never happens accidentally — it always requires an explicit
confirmation step.

---

## Features

- **Live metadata** — boot versions, Java versions, and dependency lists are
  fetched directly from `start.spring.io` on every run
- **Three-panel dashboard** — groups browser, dependency list with sticky
  group header, and a permanent selected-items panel
- **Instant search** — press `/` to filter across names, IDs, descriptions,
  and group names; matched characters are highlighted
- **Pre-search cursor restore** — pressing `Esc` clears the filter *and*
  returns the cursor to exactly where it was before searching
- **Project templates** — one-flag presets pre-select all the right
  dependencies for common project types
- **Post-generation hooks** — git init, Dockerfile, VS Code settings,
  Dev Container, docker-compose, gitignore, and README enrichment run
  automatically after extraction
- **Plugin system** — extend springx with custom templates, hooks, and
  dependency groups compiled into the binary
- **Persistent config** — store your Group ID, Java version, and build tool
  preference so you never have to re-type them
- **Clean terminal lifecycle** — alternate screen, raw mode, mouse mode, and
  cursor are always fully restored on exit, Ctrl+C, or panic
- **Cross-platform** — Linux, macOS, and Windows; amd64 and arm64

---

## Templates

Templates are opinionated dependency presets. Apply one with `--template`:

```bash
springx new --template rest-api
springx new --template jpa
springx new --template kafka
```

You can still add or remove dependencies after the template is applied — the
picker opens with the template's dependencies pre-selected.

| Template | Description | Key dependencies |
|----------|-------------|-----------------|
| `rest-api` | REST API with validation and monitoring | web, validation, actuator, lombok |
| `jpa` | Relational database with JPA and Flyway | data-jpa, postgresql, flyway, validation, lombok |
| `security` | Spring Security with OAuth2 | security, oauth2-client, validation |
| `microservice` | Spring Cloud microservice | web, actuator, cloud-feign, cloud-config-client, lombok |
| `kafka` | Event-driven service | kafka, actuator, web |
| `ai` | Spring AI starter (reserved) | web, actuator |

```bash
# List all templates
springx template list

# Show full details for a template
springx template info jpa
```

---

## Configuration

Store persistent defaults so you never re-enter the same Group ID or Java
version.

**Config file locations:**

| Platform | Path |
|----------|------|
| Linux / macOS | `~/.config/springx/config.yaml` |
| Windows | `%APPDATA%\springx\config.yaml` |

```bash
springx config init    # create with defaults
springx config show    # view active config
springx config edit    # open in $EDITOR
springx config reset   # delete (reverts to built-in defaults)
```

**Example `config.yaml`:**

```yaml
groupId: com.acme
artifactPrefix: svc-
packagePrefix: com.acme.services
javaVersion: "21"
buildTool: maven-project
packaging: jar
language: java
```

**Environment variable overrides** (highest precedence):

| Variable | Description |
|----------|-------------|
| `SPRINGX_GROUP_ID` | Default Group ID |
| `SPRINGX_ARTIFACT_PREFIX` | Prefix prepended to Artifact ID |
| `SPRINGX_PACKAGE_PREFIX` | Base package prefix |
| `SPRINGX_JAVA_VERSION` | Default Java version |
| `SPRINGX_BUILD_TOOL` | Default build tool ID |
| `SPRINGX_PACKAGING` | Default packaging (`jar` or `war`) |
| `SPRINGX_LANGUAGE` | Default language (`java`, `kotlin`, `groovy`) |

---

## Post-Generation Hooks

After a project is extracted, springx runs automation hooks to get it
development-ready. Each hook is independent — a failure in one does not abort
the rest.

```bash
springx new                          # run all hooks (default)
springx new --hook git --hook docker # run specific hooks only
springx new --no-hooks               # skip all hooks
```

**Built-in hooks:**

| Hook | What it does |
|------|--------------|
| `git` | `git init` + initial commit with a descriptive message |
| `gitignore` | Ensures `.gitignore` has all standard Spring Boot entries; merges without duplicating |
| `readme` | Appends a generation info section to `README.md` |
| `wrapper` | Verifies `mvnw` / `gradlew` was included by Spring Initializr |
| `vscode` | Creates `.vscode/extensions.json` with Java + Spring Boot extensions |
| `devcontainer` | Creates `.devcontainer/devcontainer.json` with the right Java image |
| `docker` | Multi-stage `Dockerfile` using Eclipse Temurin + Spring Boot layered jars |
| `compose` | `docker-compose.yml` with PostgreSQL or Redis when those deps are selected |

---

## Plugin System

Plugins are compiled Go packages that extend springx with:

- **Templates** — additional project presets
- **Hooks** — additional post-generation steps
- **Dependency groups** — additional entries in the dependency picker

### Using a plugin

1. Blank-import the plugin package in `main.go`:

   ```go
   import _ "github.com/saireddy-shyamakura/springx/plugins/examples/aws"
   ```

2. Write a manifest to `~/.config/springx/plugins/<name>/plugin.json`:

   ```json
   {
     "name": "aws",
     "version": "1.0.0",
     "author": "springx contributors",
     "description": "AWS templates, SAM hook, and Spring Cloud AWS dependencies."
   }
   ```

3. Rebuild: `make build`

### Plugin commands

```bash
springx plugin list            # list all registered plugins
springx plugin info aws        # full detail — templates, hooks, dep groups
springx plugin enable aws      # re-enable a disabled plugin
springx plugin disable aws     # disable without removing
```

### Authoring a plugin

```go
package myplugin

import (
    "github.com/saireddy-shyamakura/springx/internal/plugins"
    "github.com/saireddy-shyamakura/springx/internal/templates"
)

func init() { plugins.RegisterPlugin(&myPlugin{}) }

type myPlugin struct{}

func (p *myPlugin) Manifest() plugins.Manifest {
    return plugins.Manifest{
        Name:        "myplugin",
        Version:     "1.0.0",
        Author:      "Your Name",
        Description: "Adds custom templates.",
    }
}

func (p *myPlugin) Templates() []templates.Template {
    return []templates.Template{{
        Name:         "my-stack",
        Description:  "Custom project preset.",
        Dependencies: []string{"web", "actuator"},
        Defaults:     templates.TemplateDefaults{JavaVersion: "21", BuildTool: "maven-project"},
    }}
}
```

See [`plugins/examples/aws`](plugins/examples/aws) for a complete working example
that contributes templates, a hook, and a dependency group.

---

## CLI Reference

```
springx [command] [flags]

Commands:
  new           Create a new Spring Boot project
  template      View and inspect project templates
  config        Manage default configuration
  dependencies  List available Spring Boot dependencies
  plugin        Manage plugins
  version       Show version and build information

Global flags:
  -v, --verbose   Show more output
      --debug     Show debug-level output
  -h, --help      Help for any command
```

Run `springx <command> --help` for full details on any command.

---

## Troubleshooting

### `metadata fetch failed` / `connection refused`

springx needs to reach `start.spring.io` on startup.

- Check your internet connection
- If behind a corporate proxy, set `HTTPS_PROXY`
- Verify `https://start.spring.io` is reachable in your browser

### The terminal is corrupted after exiting

This should not happen with v1.0.0. If it does, run `reset` in your shell.
springx restores the terminal on normal exit, Ctrl+C, SIGTERM, SIGHUP, and
on internal panics. Please [file an issue](https://github.com/saireddy-shyamakura/springx/issues)
with your terminal emulator and OS.

### `git commit` fails in the git hook

The git hook configures a temporary identity (`springx@local`) when no global
git config is present — this covers most CI environments. If it still fails,
either configure a global git identity or use `--no-hooks`.

### Generated project does not compile

springx passes your dependency IDs directly to Spring Initializr. If an ID is
invalid or incompatible with your chosen Boot version, Initializr will return an
error. Use `springx dependencies` to see valid IDs.

---

## FAQ

**Does springx require Docker?**
No. Docker is only used by the optional `docker` and `compose` post-generation
hooks, which you can skip with `--no-hooks`.

**Can I use springx offline?**
springx currently requires an internet connection to fetch metadata and download
the generated project. An `--offline` mode using a local cache is planned for a
future release.

**Where is the downloaded ZIP kept if extraction fails?**
The ZIP is preserved in the current directory and the error screen shows its
path so you can extract it manually.

**Does springx support Kotlin or Groovy projects?**
Yes — select the language in the interactive prompts. The `SPRINGX_LANGUAGE`
environment variable can set a default.

**How do I uninstall springx?**
Remove the binary (`which springx`) and optionally delete
`~/.config/springx/` to remove configuration and plugin data.

---

## Contributing

Contributions are welcome. Please open an issue before starting significant
work so we can discuss the approach.

```bash
# Clone and set up
git clone https://github.com/saireddy-shyamakura/springx.git
cd springx
make dev-setup      # installs golangci-lint

# Run the full check suite
make check          # fmt + vet + lint + test

# Build a local binary
make build

# Run tests only
make test

# Run benchmarks
make bench
```

**Commit style:** [Conventional Commits](https://www.conventionalcommits.org/)
(`feat:`, `fix:`, `chore:`, `docs:`, etc.)

All PRs must pass CI (fmt, vet, lint, test on Linux/macOS/Windows) before merge.

---

## License

[MIT](LICENSE) © springx contributors
