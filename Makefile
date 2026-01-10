# Root Makefile for Home Assistant Blueprints
# Orchestrates blueprint validation, Go tool builds, and documentation checks

#------------------------------------------------------------------------------
# Configuration
#------------------------------------------------------------------------------

# Colors for terminal output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

# Tool paths
VALIDATE_TOOL := scripts/validate-blueprint-go/build/validate-blueprint
HA_WS_CLIENT := scripts/ha-ws-client-go/build/ha-ws-client

#------------------------------------------------------------------------------
# Default target
#------------------------------------------------------------------------------

.PHONY: all
all: build

#------------------------------------------------------------------------------
# Blueprint Validation
#------------------------------------------------------------------------------

# Validate all blueprints
.PHONY: validate
validate: build-validate
	@echo "$(GREEN)Validating all blueprints...$(NC)"
	./$(VALIDATE_TOOL) --all

# Validate a single blueprint
.PHONY: validate-single
validate-single: build-validate
	@if [ -z "$(FILE)" ]; then \
		echo "$(RED)Usage: make validate-single FILE=path/to/blueprint.yaml$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)Validating $(FILE)...$(NC)"
	./$(VALIDATE_TOOL) $(FILE)

#------------------------------------------------------------------------------
# Go Tools - Setup & Dependencies
#------------------------------------------------------------------------------

# Initialize all Go tools
.PHONY: go-init
go-init:
	@echo "$(GREEN)Initializing Go dependencies...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) init
	cd scripts/ha-ws-client-go && $(MAKE) init

# Install Go development tools
.PHONY: go-tools
go-tools:
	@echo "$(GREEN)Installing Go development tools...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) tools
	cd scripts/ha-ws-client-go && $(MAKE) tools

#------------------------------------------------------------------------------
# Go Tools - Building
#------------------------------------------------------------------------------

# Build all Go tools
.PHONY: build
build: build-validate build-ha-ws-client

# Build validate-blueprint tool
.PHONY: build-validate
build-validate:
	@cd scripts/validate-blueprint-go && $(MAKE) build

# Build ha-ws-client tool
.PHONY: build-ha-ws-client
build-ha-ws-client:
	@cd scripts/ha-ws-client-go && $(MAKE) build

#------------------------------------------------------------------------------
# Go Tools - Testing
#------------------------------------------------------------------------------

# Run all Go tests
.PHONY: go-test
go-test:
	@echo "$(GREEN)Running Go tests...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) test
	cd scripts/ha-ws-client-go && $(MAKE) test

# Run Go tests with race detection
.PHONY: go-test-race
go-test-race:
	@echo "$(GREEN)Running Go tests with race detection...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) test-race
	cd scripts/ha-ws-client-go && $(MAKE) test-race

# Run Go tests with coverage
.PHONY: go-test-cover
go-test-cover:
	@echo "$(GREEN)Running Go tests with coverage...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) test-cover
	cd scripts/ha-ws-client-go && $(MAKE) test-cover

#------------------------------------------------------------------------------
# Go Tools - Code Quality
#------------------------------------------------------------------------------

# Format all Go code
.PHONY: go-format
go-format:
	@echo "$(GREEN)Formatting Go code...$(NC)"
	@export PATH="$$HOME/go/bin:$$PATH"; \
	cd scripts/validate-blueprint-go && $(MAKE) format; \
	cd ../ha-ws-client-go && $(MAKE) format

# Lint all Go code (with auto-fix)
.PHONY: go-lint
go-lint:
	@echo "$(GREEN)Linting Go code...$(NC)"
	@export PATH="$$HOME/go/bin:$$PATH"; \
	cd scripts/validate-blueprint-go && $(MAKE) lint-fix; \
	cd ../ha-ws-client-go && $(MAKE) lint-fix

# Run go vet on all Go code
.PHONY: go-vet
go-vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) vet
	cd scripts/ha-ws-client-go && $(MAKE) vet

# Run all Go checks (format, lint, vet, test)
.PHONY: go-check
go-check:
	@echo "$(GREEN)Running all Go checks...$(NC)"
	@export PATH="$$HOME/go/bin:$$PATH"; \
	cd scripts/validate-blueprint-go && $(MAKE) check; \
	cd ../ha-ws-client-go && $(MAKE) check

# Run security audit with govulncheck
.PHONY: go-audit
go-audit:
	@echo "$(GREEN)Running security audit...$(NC)"
	@export PATH="$$HOME/go/bin:$$PATH"; \
	cd scripts/validate-blueprint-go && $(MAKE) audit; \
	cd ../ha-ws-client-go && $(MAKE) audit

#------------------------------------------------------------------------------
# Documentation
#------------------------------------------------------------------------------

# Check documentation with Biome
.PHONY: docs-check
docs-check:
	@echo "$(GREEN)Checking documentation...$(NC)"
	cd docs && npm run check

# Lint documentation
.PHONY: docs-lint
docs-lint:
	@echo "$(GREEN)Linting documentation...$(NC)"
	cd docs && npm run lint

# Lint and fix documentation issues
.PHONY: docs-lint-fix
docs-lint-fix:
	@echo "$(GREEN)Linting and fixing documentation...$(NC)"
	cd docs && npm run lint:fix

# Format documentation
.PHONY: docs-format
docs-format:
	@echo "$(GREEN)Formatting documentation...$(NC)"
	cd docs && npm run format

# Fix all documentation issues (lint + format)
.PHONY: docs-fix
docs-fix:
	@echo "$(GREEN)Fixing all documentation issues...$(NC)"
	cd docs && npm run fix

#------------------------------------------------------------------------------
# Cleanup
#------------------------------------------------------------------------------

# Clean all build artifacts
.PHONY: clean
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) clean
	cd scripts/ha-ws-client-go && $(MAKE) clean

# Clean everything including module cache
.PHONY: clean-all
clean-all:
	@echo "$(GREEN)Cleaning everything...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) clean-all
	cd scripts/ha-ws-client-go && $(MAKE) clean-all

#------------------------------------------------------------------------------
# Quality Checks
#------------------------------------------------------------------------------

# Run all checks (Go + blueprints + docs)
.PHONY: check
check: go-check validate docs-check
	@echo "$(GREEN)All checks passed!$(NC)"

# Run all checks including security audit
.PHONY: check-all
check-all: go-check go-audit validate docs-check
	@echo "$(GREEN)All checks including security audit passed!$(NC)"

# Quick check (no tests)
.PHONY: check-quick
check-quick: go-format go-lint go-vet validate
	@echo "$(GREEN)Quick checks passed!$(NC)"

# Pre-commit check (format + lint with auto-fix + tests)
.PHONY: pre-commit
pre-commit: go-format go-lint go-test validate docs-fix
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

#------------------------------------------------------------------------------
# Development
#------------------------------------------------------------------------------

# Setup development environment
.PHONY: setup
setup:
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@echo "Installing pre-commit..."
	pip install pre-commit
	pre-commit install
	@echo "Installing Go tools..."
	$(MAKE) go-tools
	@echo "Building Go tools..."
	$(MAKE) build
	@echo "Installing docs dependencies..."
	cd docs && npm install
	@echo "$(GREEN)Development environment ready!$(NC)"

# Update dependencies
.PHONY: update
update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	cd scripts/validate-blueprint-go && $(MAKE) update
	cd scripts/ha-ws-client-go && $(MAKE) update
	cd docs && npm update

#------------------------------------------------------------------------------
# Help
#------------------------------------------------------------------------------

.PHONY: help
help:
	@echo "$(GREEN)Home Assistant Blueprints - Root Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Setup:$(NC)"
	@echo "  make setup          Setup development environment (pre-commit, Go tools, docs)"
	@echo "  make go-init        Download Go dependencies"
	@echo "  make go-tools       Install Go development tools"
	@echo ""
	@echo "$(YELLOW)Blueprint Validation:$(NC)"
	@echo "  make validate       Validate all blueprints"
	@echo "  make validate-single FILE=path/to/blueprint.yaml"
	@echo "                      Validate a single blueprint"
	@echo ""
	@echo "$(YELLOW)Building:$(NC)"
	@echo "  make build          Build all Go tools"
	@echo "  make build-validate Build validate-blueprint tool"
	@echo "  make build-ha-ws-client"
	@echo "                      Build ha-ws-client tool"
	@echo ""
	@echo "$(YELLOW)Go Testing:$(NC)"
	@echo "  make go-test        Run all Go tests"
	@echo "  make go-test-race   Run tests with race detection"
	@echo "  make go-test-cover  Run tests with coverage"
	@echo ""
	@echo "$(YELLOW)Go Code Quality:$(NC)"
	@echo "  make go-format      Format all Go code"
	@echo "  make go-lint        Lint all Go code (with auto-fix)"
	@echo "  make go-vet         Run go vet on all Go code"
	@echo "  make go-check       Run all Go checks (format, lint, vet, test)"
	@echo "  make go-audit       Run security audit with govulncheck"
	@echo ""
	@echo "$(YELLOW)Documentation:$(NC)"
	@echo "  make docs-check     Check documentation with Biome"
	@echo "  make docs-lint      Lint documentation"
	@echo "  make docs-lint-fix  Lint and fix documentation"
	@echo "  make docs-format    Format documentation"
	@echo "  make docs-fix       Fix all documentation issues"
	@echo ""
	@echo "$(YELLOW)Quality Checks:$(NC)"
	@echo "  make check          Run all checks (Go + blueprints + docs)"
	@echo "  make check-all      Run all checks including security audit"
	@echo "  make check-quick    Quick check (no tests)"
	@echo "  make pre-commit     Pre-commit check (format + lint + test)"
	@echo ""
	@echo "$(YELLOW)Maintenance:$(NC)"
	@echo "  make clean          Clean build artifacts"
	@echo "  make clean-all      Clean everything including module cache"
	@echo "  make update         Update dependencies"
	@echo ""
	@echo "$(YELLOW)Help:$(NC)"
	@echo "  make help           Show this help message"
