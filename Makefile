# ── Variables ─────────────────────────────────────────────────────────────────
BINARY     := springx
MODULE     := github.com/saireddy-shyamakura/springx
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS    := -s -w \
              -X $(MODULE)/cmd.Version=$(VERSION) \
              -X $(MODULE)/cmd.Commit=$(COMMIT) \
              -X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)

# Test packages — excludes prompt (requires interactive TTY) and cmd (cobra wiring).
TEST_PKGS  := ./internal/config/... \
              ./internal/metadata/... \
              ./internal/plugins/... \
              ./internal/templates/... \
              ./internal/extract/... \
              ./internal/initializr/... \
              ./internal/postgen/... \
              ./internal/ui/...

.DEFAULT_GOAL := help

# ── Primary targets ───────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the springx binary for the current platform
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

.PHONY: install
install: ## Install springx to $GOPATH/bin
	go install -ldflags "$(LDFLAGS)" .

.PHONY: run
run: build ## Build and run springx new
	./$(BINARY) new

# ── Quality ───────────────────────────────────────────────────────────────────

.PHONY: fmt
fmt: ## Run gofmt on all Go source files
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Check formatting without modifying files (used in CI)
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files need formatting:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (install with: make dev-setup)
	golangci-lint run --timeout=5m

.PHONY: check
check: fmt-check vet test lint ## Run all quality checks (fmt + vet + test + lint)

# ── Testing ───────────────────────────────────────────────────────────────────

.PHONY: test
test: ## Run the full test suite
	go test -count=1 -timeout=120s $(TEST_PKGS)

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	go test -v -count=1 -timeout=120s $(TEST_PKGS)

.PHONY: test-race
test-race: ## Run tests with the race detector enabled
	go test -race -count=1 -timeout=120s $(TEST_PKGS)

.PHONY: coverage
coverage: ## Run tests and produce an HTML coverage report
	go test -count=1 -coverprofile=coverage.out $(TEST_PKGS)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

.PHONY: bench
bench: ## Run all benchmarks
	go test -bench=. -benchmem -run='^$$' ./internal/...

# ── Release ───────────────────────────────────────────────────────────────────

.PHONY: snapshot
snapshot: ## Build a GoReleaser snapshot (no publish, no git tag required)
	goreleaser release --snapshot --clean

.PHONY: release-dry-run
release-dry-run: ## Dry-run GoReleaser (checks config, does not publish)
	goreleaser check

# ── Development setup ─────────────────────────────────────────────────────────

.PHONY: dev-setup
dev-setup: ## Install development tools (golangci-lint)
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Done."

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: clean
clean: ## Remove build artifacts
	rm -f $(BINARY) $(BINARY).exe coverage.out coverage.html
	rm -rf dist/
