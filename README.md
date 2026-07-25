# springx

A modern, fast, and interactive Go CLI tool for generating Spring Boot applications via Spring Initializr metadata.

## Features

- **Professional multi-panel TUI** — group browser, live dependency list, and persistent selected-items panel side-by-side
- **Instant search** — press `/` to filter across names, IDs, descriptions, and group names with match highlighting
- **Group navigation** — `tab` / `shift+tab` jumps between dependency groups without scrolling
- **Project templates** — one-command presets (`rest-api`, `jpa`, `kafka`, …) with pre-selected dependencies
- **Post-generation hooks** — automatic git init, Dockerfile, VS Code settings, Dev Container, docker-compose, and more
- **Live Spring Initializr metadata** — boot versions, Java versions, build tools fetched from `start.spring.io`
- **User configuration** — persistent defaults via `~/.config/springx/config.yaml` or environment variables
- **Plugin system** — extend springx with custom templates, hooks, and dependency groups without forking

---

## Terminal UI

The dependency picker is a three-panel full-screen interface with a sticky group header, live search result count, confirmation screen, and inline progress pipeline.

```
 springx                                           Spring Boot 3.5.0
 ──────────────────────────────────────────────────────────────────
 Search: postgres  Found 1 dependency  (esc to clear)   (/ or Ctrl+F)
 ──────────────────────────────────────────────────────────────────
╭────────────────╮╔══════════════════════════════╗╭──────────────────╮
│ Groups         ││ Dependencies                 ││ Selected (2)     │
│                ││ ▸ Data  ← sticky header      ││                  │
│   Developer… │││   [ ] Spring Data JPA        ││ ✓ Spring Web     │
│   Web          ││   [x] PostgreSQL Driver      ││   Web            │
│ > Data         ││   [ ] MySQL Driver           ││ ✓ PostgreSQL…    │
│   Security     ││   [ ] MongoDB                ││   Data           │
│   Messaging    ││                              ││                  │
╰────────────────╯╚══════════════════════════════╝╰──────────────────╯
 ↑↓ move • ←→/tab groups • Home/End first/last • PgUp/PgDn page
 • space select • / search • enter confirm • ? help • q quit
 Metadata loaded │ Boot 3.5.0 │ Java 21 │ Template jpa │ Selected 2
```

### Keyboard shortcuts

#### Navigation

| Key | Action |
|---|---|
| `↑` / `↓` or `k` / `j` | Move cursor up / down |
| `Home` / `g` | Jump to first dependency |
| `End` / `G` | Jump to last dependency |
| `PgUp` / `Ctrl+U` | Page up (8 rows) |
| `PgDn` / `Ctrl+D` | Page down (8 rows) |
| `Tab` / `→` / `l` | Next group |
| `Shift+Tab` / `←` / `h` | Previous group |

#### Selection & search

| Key | Action |
|---|---|
| `Space` | Toggle selection on cursor row |
| `/` or `Ctrl+F` | Open search |
| `Esc` | Clear search |
| `Ctrl+Backspace` | Clear entire search query |
| `Enter` | Open confirmation screen |

#### General

| Key | Action |
|---|---|
| `?` | Open / close keyboard shortcut help |
| `q` / `Ctrl+C` | Quit |

### Search

Press `/` or `Ctrl+F` to enter search mode. Results filter instantly across dependency names, IDs, descriptions, and group names. Matching characters are highlighted in gold. The bar shows `Found N dependencies` or `No matching dependencies`. Press `Esc` to clear.

### Confirmation screen

Pressing `Enter` opens a confirmation screen showing Spring Boot version, Java version, active template, and the full list of selected dependencies (with their group names). Press `Y` or `Enter` to generate, `N` or `Esc` to go back.

### Progress pipeline

After confirmation the UI switches to a linear progress view:

```
╭──────────────────────────────────────────────────╮
│ Generating Spring Boot project                   │
│                                                  │
│ ✔  Downloading from Spring Initializr  demo.zip  │
│ ✔  Extracting project                            │
│ ●  Running post-generation hooks                 │
│ ○  Done                                          │
╰──────────────────────────────────────────────────╯
```

Each step is independent. A failed step is marked with `✗` and execution continues.

### Error display

Network or generation errors are shown as formatted boxes instead of raw stack traces:

```
╭──────────────────────────────────────────────────────────╮
│ ❌  Unable to fetch Spring Initializr metadata.          │
│                                                          │
│ Reason:                                                  │
│   dial tcp: connection refused                           │
│                                                          │
│ Suggestions:                                             │
│   • Check your internet connection.                      │
│   • Verify that start.spring.io is reachable.            │
│   • Try again in a few seconds.                          │
╰──────────────────────────────────────────────────────────╯
```

---

## Installation

Build the binary using Go standard tooling:

```bash
go build -o springx main.go
```

---

## Usage

### Generate a New Spring Boot Project

```bash
springx new
```

Walks through interactive prompts for project parameters (Name, Group ID, Artifact ID, Package Name, Build Tool, Packaging, Java Version) and launches the interactive Bubble Tea dependency selector.

### List Dependencies

```bash
springx dependencies
# or
springx deps
```

Lists all available Spring Boot dependencies grouped by category.

---

## Configuration & Developer Experience

You can store permanent defaults so you don't have to re-enter common values (like `groupId` or `javaVersion`) every time you initialize a project.

### Configuration File Locations

- **Linux / macOS**: `~/.config/springx/config.yaml`
- **Windows**: `%APPDATA%\springx\config.yaml`

### Configuration Commands

- **Initialize Config**:
  ```bash
  springx config init
  ```
- **Show Active Configuration**:
  ```bash
  springx config show
  ```
- **Edit Configuration**:
  ```bash
  springx config edit
  ```
  *(Opens the config file in your `$EDITOR` or `nano`/`notepad`)*
- **Reset Configuration**:
  ```bash
  springx config reset
  ```

### Example `config.yaml`

```yaml
groupId: com.saireddy
packagePrefix: com.saireddy
javaVersion: 21
buildTool: maven-project
packaging: jar
language: java
```

### Environment Variable Overrides

Environment variables take highest precedence and override configuration file values:

| Environment Variable | Description | Example |
|---|---|---|
| `SPRINGX_GROUP_ID` | Default Group ID | `com.mycompany` |
| `SPRINGX_ARTIFACT_PREFIX` | Prefix prepended to Artifact ID | `service-` |
| `SPRINGX_PACKAGE_PREFIX` | Prefix prepended to Package Name | `com.mycompany.service` |
| `SPRINGX_JAVA_VERSION` | Default Java Version | `21` |
| `SPRINGX_BUILD_TOOL` | Default Build Tool ID or Name | `maven-project` |
| `SPRINGX_PACKAGING` | Default Packaging | `jar` |
| `SPRINGX_LANGUAGE` | Default Language | `java` |

---

---

## Project Templates

Use `--template` to bootstrap from an opinionated preset instead of selecting every dependency manually.

```bash
springx new --template rest-api
springx new --template jpa
springx new --template kafka
```

### Available Templates

| Template | Description | Dependencies |
|---|---|---|
| `rest-api` | REST API with validation and monitoring | web, validation, actuator, lombok |
| `jpa` | Relational database with Spring Data JPA and Flyway | data-jpa, postgresql, flyway, validation, lombok |
| `security` | Spring Security with OAuth2 client support | security, oauth2-client, validation |
| `microservice` | Spring Cloud microservice with Feign and Config | web, actuator, cloud-feign, cloud-config-client, validation, lombok |
| `kafka` | Event-driven service with Apache Kafka | kafka, actuator, web |
| `ai` | Reserved for future Spring AI support | web, actuator |

### Template Commands

```bash
# List all templates
springx template list

# Show details for a specific template
springx template info rest-api
```

You may still modify any field (including dependencies) after a template is applied.

---

## Post-Generation Hooks

After a project is generated and extracted, springx automatically runs a set of **post-generation hooks** that prepare the project for development. Each hook is independent — a failure in one does not abort the others.

### Running Hooks

```bash
# Run all hooks (default behaviour)
springx new

# Run only specific hooks
springx new --hook git --hook docker --hook vscode

# Skip all hooks
springx new --no-hooks
```

### Built-in Hooks

| Hook | Description |
|---|---|
| `git` | Runs `git init` and creates an initial commit |
| `gitignore` | Ensures `.gitignore` exists with all standard Spring Boot entries; merges into an existing file without duplicating entries |
| `readme` | Appends a `## springx Generation Info` section to `README.md` with template, Java version, build tool, and generation timestamp |
| `wrapper` | Verifies that `mvnw` (Maven) or `gradlew` (Gradle) exists — warns if Spring Initializr did not include it |
| `vscode` | Generates `.vscode/extensions.json` recommending `vscjava.vscode-java-pack` and `vmware.vscode-spring-boot` |
| `devcontainer` | Generates `.devcontainer/devcontainer.json` with the correct Java image, port forwarding, and VS Code extensions |
| `docker` | Generates a multi-stage, production-ready `Dockerfile` using Eclipse Temurin images and Spring Boot layered jars |
| `compose` | Generates `docker-compose.yml` when `postgresql` or `redis` dependencies are selected; includes health checks and volume mounts |

### Progress Output

```
Running post-generation hooks:
  ✔ git
  ✔ gitignore
  ✔ readme
  ✔ wrapper
  ✔ vscode
  ✔ docker
  ✔ compose

  ✔ Completed

Your project is ready at: ./my-service
```

Failures are reported inline and summarised at the end without aborting the remaining hooks.

---

## Plugin System

springx supports third-party plugins that add templates, hooks, and dependency groups — all without modifying the core codebase.

### How plugins work

Plugins are compiled Go packages. A plugin struct registers itself in its `init()` function using `plugins.RegisterPlugin`, then the package is blank-imported into the host binary. This is the same pattern springx uses for its built-in hooks and is idiomatic Go — no shared objects, no CGO, no runtime code loading.

```
~/.config/springx/plugins/
└── aws/
    └── plugin.json     ← manifest (name, version, author, description)
```

The manifest is read at runtime for `plugin list` / `plugin info` display and for persisting enable/disable state. The Go code itself is compiled in at build time.

### Plugin commands

```bash
springx plugin list              # list all registered plugins
springx plugin info aws          # detailed info: templates, hooks, dep groups
springx plugin enable aws        # enable a previously disabled plugin
springx plugin disable aws       # disable without removing
```

### Interfaces

A plugin struct may implement any combination of three extension-point interfaces on top of the base `Plugin` interface:

| Interface | Method | What it contributes |
|---|---|---|
| `TemplatePlugin` | `Templates() []templates.Template` | Project presets available via `--template` |
| `HookPlugin` | `Hooks() []postgen.Hook` | Post-generation automation steps |
| `DependencyProvider` | `DependencyGroups() []metadata.DependencyGroup` | Extra groups in the dependency picker |

### Authoring a plugin

**1. Create your package**

```go
// plugins/myplugin/myplugin.go
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
        Description: "Adds custom templates for my stack.",
    }
}

// TemplatePlugin
func (p *myPlugin) Templates() []templates.Template {
    return []templates.Template{
        {
            Name:        "my-template",
            Description: "A custom project preset.",
            Dependencies: []string{"web", "actuator"},
            Defaults: templates.TemplateDefaults{
                JavaVersion: "21",
                BuildTool:   "maven-project",
                Packaging:   "jar",
            },
        },
    }
}
```

**2. Blank-import in main.go**

```go
import _ "github.com/saireddy-shyamakura/springx/plugins/myplugin"
```

**3. Write a manifest** at `~/.config/springx/plugins/myplugin/plugin.json`:

```json
{
  "name": "myplugin",
  "version": "1.0.0",
  "author": "Your Name",
  "description": "Adds custom templates for my stack.",
  "homepage": "https://github.com/you/springx-myplugin"
}
```

### Example plugin — AWS

`plugins/examples/aws` is a fully working example plugin. It contributes:

- **Templates**: `aws-lambda` (Spring Cloud Function on Lambda), `aws-s3` (Spring Cloud AWS S3)
- **Hook**: `aws-sam` — generates `template.yaml` for AWS SAM deployment
- **Dependency group**: `AWS` — Spring Cloud AWS starters (S3, SQS, SNS, Secrets Manager, Parameter Store, Lambda)

Activate it by blank-importing in `main.go`:

```go
import _ "github.com/saireddy-shyamakura/springx/plugins/examples/aws"
```

Then use it:

```bash
springx new --template aws-lambda
springx new --template aws-s3 --hook aws-sam
springx plugin info aws
```

### Plugin enable/disable

Disabled plugins are persisted to `~/.config/springx/plugins/disabled.json`. The file is a JSON array of plugin names:

```json
["aws", "another-plugin"]
```

Enable/disable state is loaded at the start of every `springx new` run and respected by all commands.

---

## Running Tests

Run the full test suite:

```bash
go test -v ./...
```