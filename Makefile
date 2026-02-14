.PHONY: help build-cli build-desktop build-all test-cli test-desktop test-all clean dev install

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build-cli: ## Build the CLI (boatman binary)
	@echo "Building CLI..."
	cd cli && go build -o boatman ./cmd/boatman
	@echo "✓ CLI built at cli/boatman"

build-desktop: build-cli ## Build the desktop app (requires CLI)
	@echo "Building desktop app..."
	cd desktop && wails build
	@echo "✓ Desktop app built at desktop/build/"

build-all: build-cli build-desktop ## Build both CLI and desktop

test-cli: ## Run CLI tests
	@echo "Running CLI tests..."
	cd cli && go test ./...

test-desktop: ## Run desktop Go tests
	@echo "Running desktop Go tests..."
	cd desktop && go test ./...

test-frontend: ## Run desktop frontend tests
	@echo "Running desktop frontend tests..."
	cd desktop/frontend && npm test

test-all: test-cli test-desktop ## Run all tests

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f cli/boatman
	rm -rf desktop/build/
	rm -rf cli/.worktrees/
	rm -rf desktop/.worktrees/
	@echo "✓ Clean complete"

dev: build-cli ## Start desktop app in dev mode
	@echo "Starting desktop in dev mode..."
	cd desktop && wails dev

install-cli: build-cli ## Install CLI to ~/bin
	@echo "Installing CLI to ~/bin..."
	mkdir -p ~/bin
	cp cli/boatman ~/bin/boatman
	@echo "✓ CLI installed to ~/bin/boatman"
	@echo "  Make sure ~/bin is in your PATH"

workspace-sync: ## Sync Go workspace
	go work sync

fmt: ## Format all Go code
	@echo "Formatting Go code..."
	cd cli && go fmt ./...
	cd desktop && go fmt ./...
	@echo "✓ Code formatted"

lint: ## Run linters
	@echo "Running linters..."
	cd cli && golangci-lint run || true
	cd desktop && golangci-lint run || true

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	cd cli && go mod download
	cd desktop && go mod download
	cd desktop/frontend && npm install
	@echo "✓ Dependencies downloaded"

# =============================================================================
# Version Management
# =============================================================================

bump-cli-patch: ## Bump CLI patch version (1.2.3 -> 1.2.4)
	./scripts/bump-version.sh cli patch

bump-cli-minor: ## Bump CLI minor version (1.2.3 -> 1.3.0)
	./scripts/bump-version.sh cli minor

bump-cli-major: ## Bump CLI major version (1.2.3 -> 2.0.0)
	./scripts/bump-version.sh cli major

bump-desktop-patch: ## Bump desktop patch version (1.0.0 -> 1.0.1)
	./scripts/bump-version.sh desktop patch

bump-desktop-minor: ## Bump desktop minor version (1.0.0 -> 1.1.0)
	./scripts/bump-version.sh desktop minor

bump-desktop-major: ## Bump desktop major version (1.0.0 -> 2.0.0)
	./scripts/bump-version.sh desktop major

show-versions: ## Show current versions
	@echo "Current versions:"
	@echo "  CLI:     $$(cat cli/VERSION 2>/dev/null || echo 'not set')"
	@echo "  Desktop: $$(grep '"version"' desktop/wails.json | sed 's/.*"version": "\(.*\)".*/v\1/')"

release: ## Interactive release wizard
	./scripts/release.sh

release-guide: ## Show release documentation
	@echo "Release Documentation:"
	@echo "  • RELEASE_SUMMARY.md - Quick reference"
	@echo "  • RELEASES.md        - Complete guide"
	@echo "  • VERSIONING.md      - Strategy details"
	@echo ""
	@echo "Quick commands:"
	@echo "  make show-versions   - Show current versions"
	@echo "  make release         - Interactive release wizard"
	@echo "  make bump-cli-minor  - Bump CLI minor version"
	@echo "  make bump-desktop-minor - Bump desktop minor version"
