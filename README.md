# springx

A modern, fast, and interactive Go CLI tool for generating Spring Boot applications via Spring Initializr metadata.

## Features

- **Interactive TUI**: Built with Bubble Tea & Lipgloss for a full-screen dependency selection experience.
- **Dynamic Metadata**: Fetches live metadata (boot versions, Java versions, build tools, packaging options, dependencies) directly from `start.spring.io`.
- **Dependency Search**: Instant live filter search (`/`) across all Spring Boot dependency groups.
- **Configuration & Defaults**: Define permanent project defaults via configuration files or environment variables.

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

## Running Tests

Run the full test suite:

```bash
go test -v ./...
```