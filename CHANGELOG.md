# Changelog

All notable changes to springx are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
springx adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] — 2024-01-01

### Added

- Interactive three-panel dependency picker TUI built with Bubble Tea.
  - Groups browser, dependency list with sticky group header, and permanent selected-items panel.
  - Instant `/` search with match highlighting and pre-search cursor restore.
  - Confirmation screen before generation.
  - Full keyboard navigation: arrows, vim keys, Tab group jump, PageUp/Down, Home/End.
- `springx new` — interactive Spring Boot project wizard.
  - Fetches live metadata from `start.spring.io` on every run.
  - `--template` flag to bootstrap from a named preset.
  - `--hook` flag to run specific post-generation hooks.
  - `--no-hooks` flag to skip all automation.
- Built-in project templates: `rest-api`, `jpa`, `security`, `microservice`, `kafka`, `ai`.
- Post-generation hooks running automatically after project extraction:
  - `git` — `git init` and initial commit.
  - `gitignore` — merges Spring Boot `.gitignore` entries without duplicates.
  - `readme` — appends a generation info section to `README.md`.
  - `wrapper` — verifies Maven/Gradle wrapper was included.
  - `vscode` — creates `.vscode/extensions.json` with Java + Spring Boot extensions.
  - `devcontainer` — creates `.devcontainer/devcontainer.json`.
  - `docker` — multi-stage `Dockerfile` using Eclipse Temurin and layered jars.
  - `compose` — `docker-compose.yml` with PostgreSQL/Redis when selected as dependencies.
- `springx template list` / `springx template info <name>` — browse and inspect templates.
- `springx config` — persistent defaults for Group ID, Java version, build tool, etc.
  - `init`, `show`, `edit`, `reset` sub-commands.
  - Environment variable overrides (`SPRINGX_GROUP_ID`, `SPRINGX_JAVA_VERSION`, etc.).
- `springx dependencies` — list all available Spring Boot dependencies from live metadata.
- `springx plugin` — manage compiled-in plugins (`list`, `info`, `enable`, `disable`).
- `springx version` — display version, commit, build date, Go version, and platform.
- Plugin system: contribute templates, hooks, and dependency groups via blank-imported Go packages.
- Example AWS plugin (`plugins/examples/aws`) — Lambda/S3 templates, SAM hook, AWS dependency group.
- `--verbose` and `--debug` persistent flags for detailed output.
- Full terminal lifecycle safety: alternate screen, raw mode, mouse mode, and cursor are restored on exit, Ctrl+C, SIGTERM, SIGHUP, and panic.
- Cross-platform binaries: Linux amd64/arm64, macOS amd64/arm64, Windows amd64.
- Zip Slip-protected archive extraction.
- Metadata caching — Spring Initializr is only contacted once per process.

[1.0.0]: https://github.com/saireddy-shyamakura/springx/releases/tag/v1.0.0
