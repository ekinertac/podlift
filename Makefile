.PHONY: build run test clean help install

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -X 'github.com/ekinertac/podlift/cmd/podlift/commands.Version=$(VERSION)' \
          -X 'github.com/ekinertac/podlift/cmd/podlift/commands.Commit=$(COMMIT)' \
          -X 'github.com/ekinertac/podlift/cmd/podlift/commands.Date=$(DATE)'

# Build the binary
build:
	@echo "Building podlift..."
	@go build -ldflags "$(LDFLAGS)" -o bin/podlift ./cmd/podlift
	@echo "✓ Built bin/podlift"

# Run locally
run: build
	@./bin/podlift

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "✓ Cleaned"

# Install to system
install: build
	@echo "Installing podlift..."
	@sudo cp bin/podlift /usr/local/bin/
	@echo "✓ Installed to /usr/local/bin/podlift"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	@echo "✓ Linted"

# Run all checks
check: fmt lint test
	@echo "✓ All checks passed"

# Show help
help:
	@echo "podlift Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          Build the binary"
	@echo "  make run             Build and run"
	@echo "  make test            Run tests"
	@echo "  make test-coverage   Run tests with coverage"
	@echo "  make clean           Clean build artifacts"
	@echo "  make install         Install to /usr/local/bin"
	@echo "  make fmt             Format code"
	@echo "  make lint            Lint code"
	@echo "  make check           Run all checks"
	@echo ""
	@echo "Build variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
	@echo "  DATE=$(DATE)"

