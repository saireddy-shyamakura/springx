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

## Running Tests

Run the full test suite:

```bash
go test -v ./...
```