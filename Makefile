# podlift Makefile

.PHONY: help build test install clean release verify bump-patch bump-minor bump-major bump-alpha bump-beta

# Default target
.DEFAULT_GOAL := help

# Version from git or default
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X github.com/ekinertac/podlift/cmd/podlift/commands.Version=$(VERSION) -X github.com/ekinertac/podlift/cmd/podlift/commands.Commit=$(COMMIT)

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building podlift $(VERSION)..."
	@go build -ldflags "$(LDFLAGS)" -o bin/podlift ./cmd/podlift
	@echo "✓ Built bin/podlift"

install: build ## Install to /usr/local/bin
	@echo "Installing to /usr/local/bin..."
	@sudo cp bin/podlift /usr/local/bin/
	@echo "✓ Installed"

test: ## Run all tests
	@echo "Running tests..."
	@go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage.out coverage.html
	@echo "✓ Cleaned"

release: ## Build release binaries for all platforms
	@./scripts/build-release.sh $(VERSION)

verify: ## Run all verification checks
	@./scripts/verify.sh

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "✓ Vet passed"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "✓ Dependencies downloaded"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@go mod tidy
	@echo "✓ Tidied"

dev: ## Build and install for development
	@$(MAKE) build
	@$(MAKE) install
	@echo "✓ Development build installed"

all: fmt vet test build ## Run fmt, vet, test, and build
	@echo "✓ All checks passed"

bump-patch: ## Bump patch version (1.0.0 -> 1.0.1) and create release
	@./scripts/bump-version.sh patch

bump-minor: ## Bump minor version (1.0.0 -> 1.1.0) and create release
	@./scripts/bump-version.sh minor

bump-major: ## Bump major version (1.0.0 -> 2.0.0) and create release
	@./scripts/bump-version.sh major

bump-alpha: ## Create alpha pre-release (1.0.2 -> 1.0.3-alpha.1)
	@./scripts/bump-version.sh alpha

bump-beta: ## Create beta pre-release (1.0.2 -> 1.0.3-beta.1)
	@./scripts/bump-version.sh beta
