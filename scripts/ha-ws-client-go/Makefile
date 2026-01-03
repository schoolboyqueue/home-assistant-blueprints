# ha-ws-client-go Makefile
# Supports cross-compilation for Raspberry Pi and other ARM devices

BINARY_NAME=ha-ws-client
VERSION?=1.3.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build directory
BUILD_DIR=build

# Tool versions (for reproducible builds)
GOLANGCI_LINT_VERSION=v1.62.2
GOFUMPT_VERSION=v0.7.0

# Colors for terminal output
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

# Default target
.PHONY: all
all: build

#------------------------------------------------------------------------------
# Setup & Dependencies
#------------------------------------------------------------------------------

# Initialize go modules
.PHONY: init
init:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify

# Install development tools
.PHONY: tools
tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@echo "Installing gofumpt $(GOFUMPT_VERSION)..."
	@go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)
	@echo "Installing goimports..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)All tools installed!$(NC)"

# Check if required tools are installed
.PHONY: check-tools
check-tools:
	@which golangci-lint > /dev/null || (echo "$(RED)golangci-lint not found. Run 'make tools' to install.$(NC)" && exit 1)
	@which gofumpt > /dev/null || (echo "$(RED)gofumpt not found. Run 'make tools' to install.$(NC)" && exit 1)
	@which goimports > /dev/null || (echo "$(RED)goimports not found. Run 'make tools' to install.$(NC)" && exit 1)
	@echo "$(GREEN)All required tools are installed.$(NC)"

#------------------------------------------------------------------------------
# Formatting
#------------------------------------------------------------------------------

# Format code with gofmt (basic)
.PHONY: fmt
fmt:
	@echo "$(GREEN)Formatting code with gofmt...$(NC)"
	$(GOFMT) ./...

# Format code with gofumpt (stricter formatting)
.PHONY: fumpt
fumpt:
	@echo "$(GREEN)Formatting code with gofumpt...$(NC)"
	@gofumpt -l -w .

# Format imports with goimports
.PHONY: imports
imports:
	@echo "$(GREEN)Organizing imports with goimports...$(NC)"
	@goimports -local github.com/home-assistant-blueprints/ha-ws-client-go -w .

# Format everything (recommended)
.PHONY: format
format: fumpt imports
	@echo "$(GREEN)All formatting complete!$(NC)"

# Check formatting without making changes
.PHONY: format-check
format-check:
	@echo "$(GREEN)Checking code formatting...$(NC)"
	@test -z "$$(gofumpt -l . 2>&1)" || (echo "$(RED)Code is not properly formatted. Run 'make format' to fix.$(NC)" && gofumpt -l . && exit 1)
	@test -z "$$(goimports -local github.com/home-assistant-blueprints/ha-ws-client-go -l . 2>&1)" || (echo "$(RED)Imports are not properly organized. Run 'make format' to fix.$(NC)" && exit 1)
	@echo "$(GREEN)Formatting check passed!$(NC)"

#------------------------------------------------------------------------------
# Linting
#------------------------------------------------------------------------------

# Run golangci-lint
.PHONY: lint
lint:
	@echo "$(GREEN)Running golangci-lint...$(NC)"
	golangci-lint run

# Run golangci-lint with auto-fix
.PHONY: lint-fix
lint-fix:
	@echo "$(GREEN)Running golangci-lint with auto-fix...$(NC)"
	golangci-lint run --fix

# Run go vet
.PHONY: vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	$(GOVET) ./...

# Fast lint (only essential checks)
.PHONY: lint-fast
lint-fast:
	@echo "$(GREEN)Running fast lint...$(NC)"
	golangci-lint run --fast

#------------------------------------------------------------------------------
# Testing
#------------------------------------------------------------------------------

# Run tests
.PHONY: test
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v ./...

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "$(GREEN)Running tests with race detection...$(NC)"
	$(GOTEST) -race -v ./...

# Run tests with coverage
.PHONY: test-cover
test-cover:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Run tests with short flag (skip long-running tests)
.PHONY: test-short
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	$(GOTEST) -short -v ./...

# Run benchmarks
.PHONY: bench
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

#------------------------------------------------------------------------------
# Quality Checks
#------------------------------------------------------------------------------

# Run all checks (format check, lint, vet, test)
.PHONY: check
check: format-check lint vet test
	@echo "$(GREEN)All checks passed!$(NC)"

# Quick check (no tests)
.PHONY: check-quick
check-quick: format-check lint-fast vet
	@echo "$(GREEN)Quick checks passed!$(NC)"

# Pre-commit check (format, lint-fix, test)
.PHONY: pre-commit
pre-commit: format lint-fix test
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

# CI check (strict, no fixes)
.PHONY: ci
ci: format-check lint vet test-race
	@echo "$(GREEN)CI checks passed!$(NC)"

#------------------------------------------------------------------------------
# Building
#------------------------------------------------------------------------------

# Build for current platform
.PHONY: build
build:
	@echo "$(GREEN)Building for current platform...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ha-ws-client
	@echo "$(GREEN)Built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build with race detector (for development)
.PHONY: build-race
build-race:
	@echo "$(GREEN)Building with race detector...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -race $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-race ./cmd/ha-ws-client

# Build for Linux AMD64 (typical server/container)
.PHONY: build-linux-amd64
build-linux-amd64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ha-ws-client

# Build for Linux ARM64 (Raspberry Pi 4, newer 64-bit Pi)
.PHONY: build-linux-arm64
build-linux-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/ha-ws-client

# Build for Linux ARM v7 (Raspberry Pi 2/3, older 32-bit)
.PHONY: build-linux-armv7
build-linux-armv7:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv7 ./cmd/ha-ws-client

# Build for Linux ARM v6 (Raspberry Pi Zero, Pi 1)
.PHONY: build-linux-armv6
build-linux-armv6:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv6 ./cmd/ha-ws-client

# Build for macOS AMD64
.PHONY: build-darwin-amd64
build-darwin-amd64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ha-ws-client

# Build for macOS ARM64 (Apple Silicon)
.PHONY: build-darwin-arm64
build-darwin-arm64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/ha-ws-client

# Build for Windows AMD64
.PHONY: build-windows-amd64
build-windows-amd64:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/ha-ws-client

# Build all Raspberry Pi variants
.PHONY: build-pi
build-pi: build-linux-arm64 build-linux-armv7 build-linux-armv6
	@echo "$(GREEN)Built all Raspberry Pi variants:$(NC)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)-linux-arm*

# Build all platforms
.PHONY: build-all
build-all: build-linux-amd64 build-linux-arm64 build-linux-armv7 build-linux-armv6 build-darwin-amd64 build-darwin-arm64 build-windows-amd64
	@echo "$(GREEN)Built all platforms:$(NC)"
	@ls -lh $(BUILD_DIR)/

#------------------------------------------------------------------------------
# Cleanup
#------------------------------------------------------------------------------

# Clean build artifacts
.PHONY: clean
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Clean everything including module cache
.PHONY: clean-all
clean-all: clean
	@echo "$(GREEN)Cleaning module cache...$(NC)"
	$(GOCMD) clean -modcache

#------------------------------------------------------------------------------
# Development
#------------------------------------------------------------------------------

# Install the binary locally
.PHONY: install
install: build
	@echo "$(GREEN)Installing to GOPATH/bin...$(NC)"
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Show binary sizes
.PHONY: sizes
sizes:
	@echo "Binary sizes:"
	@ls -lh $(BUILD_DIR)/ 2>/dev/null || echo "No binaries built yet. Run 'make build-all' first."

# Development: build and run
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Watch mode: rebuild on changes (requires entr)
.PHONY: watch
watch:
	@which entr > /dev/null || (echo "$(RED)entr not found. Install with: brew install entr$(NC)" && exit 1)
	find . -name "*.go" | entr -c make build

# Docker build (for Home Assistant add-on)
.PHONY: docker
docker:
	docker build -t ha-ws-client:$(VERSION) .

# Generate dependency graph (requires graphviz)
.PHONY: deps-graph
deps-graph:
	@which dot > /dev/null || (echo "$(RED)graphviz not found. Install with: brew install graphviz$(NC)" && exit 1)
	$(GOCMD) mod graph | modgraphviz | dot -Tpng -o deps.png
	@echo "$(GREEN)Dependency graph generated: deps.png$(NC)"

# Tidy go modules
.PHONY: tidy
tidy:
	@echo "$(GREEN)Tidying go modules...$(NC)"
	$(GOMOD) tidy

# Update dependencies
.PHONY: update
update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Security audit
.PHONY: audit
audit:
	@echo "$(GREEN)Running security audit...$(NC)"
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

#------------------------------------------------------------------------------
# Help
#------------------------------------------------------------------------------

.PHONY: help
help:
	@echo "$(GREEN)ha-ws-client-go Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Setup:$(NC)"
	@echo "  make init           Download dependencies"
	@echo "  make tools          Install development tools (golangci-lint, gofumpt, goimports)"
	@echo "  make check-tools    Verify all required tools are installed"
	@echo ""
	@echo "$(YELLOW)Formatting:$(NC)"
	@echo "  make fmt            Format with gofmt (basic)"
	@echo "  make fumpt          Format with gofumpt (stricter)"
	@echo "  make imports        Organize imports with goimports"
	@echo "  make format         Format everything (recommended)"
	@echo "  make format-check   Check formatting without changes"
	@echo ""
	@echo "$(YELLOW)Linting:$(NC)"
	@echo "  make lint           Run golangci-lint"
	@echo "  make lint-fix       Run golangci-lint with auto-fix"
	@echo "  make lint-fast      Run fast lint (essential checks only)"
	@echo "  make vet            Run go vet"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  make test           Run all tests"
	@echo "  make test-race      Run tests with race detection"
	@echo "  make test-cover     Run tests with coverage report"
	@echo "  make test-short     Run short tests only"
	@echo "  make bench          Run benchmarks"
	@echo ""
	@echo "$(YELLOW)Quality Checks:$(NC)"
	@echo "  make check          Run all checks (format, lint, vet, test)"
	@echo "  make check-quick    Quick check (no tests)"
	@echo "  make pre-commit     Pre-commit check (format, lint-fix, test)"
	@echo "  make ci             CI check (strict, no fixes)"
	@echo ""
	@echo "$(YELLOW)Building:$(NC)"
	@echo "  make build          Build for current platform"
	@echo "  make build-pi       Build for all Raspberry Pi variants"
	@echo "  make build-all      Build for all platforms"
	@echo "  make build-race     Build with race detector"
	@echo ""
	@echo "$(YELLOW)Platform-specific builds:$(NC)"
	@echo "  make build-linux-amd64     Linux x86_64"
	@echo "  make build-linux-arm64     Linux ARM64 (Pi 4 64-bit)"
	@echo "  make build-linux-armv7     Linux ARMv7 (Pi 2/3 32-bit)"
	@echo "  make build-linux-armv6     Linux ARMv6 (Pi Zero/1)"
	@echo "  make build-darwin-amd64    macOS Intel"
	@echo "  make build-darwin-arm64    macOS Apple Silicon"
	@echo "  make build-windows-amd64   Windows x86_64"
	@echo ""
	@echo "$(YELLOW)Maintenance:$(NC)"
	@echo "  make clean          Remove build artifacts"
	@echo "  make clean-all      Remove build artifacts and module cache"
	@echo "  make tidy           Tidy go modules"
	@echo "  make update         Update dependencies"
	@echo "  make audit          Run security audit"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  make install        Install binary to GOPATH/bin"
	@echo "  make run ARGS=...   Build and run with arguments"
	@echo "  make watch          Watch mode (requires entr)"
	@echo "  make sizes          Show binary sizes"
