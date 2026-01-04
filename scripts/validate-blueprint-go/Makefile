# validate-blueprint-go Makefile
# Supports cross-compilation for multiple platforms

#------------------------------------------------------------------------------
# Project-specific Configuration
#------------------------------------------------------------------------------

# Required variables for shared.mk
BINARY_NAME := validate-blueprint
PROJECT_NAME := validate-blueprint-go
GO_LOCAL_PATH := github.com/home-assistant-blueprints/validate-blueprint-go

# Optional overrides
VERSION ?= 1.7.0
BUILD_DIR := build
BUILD_PATH := .

#------------------------------------------------------------------------------
# Include shared Makefile
#------------------------------------------------------------------------------

include ../shared.mk

#------------------------------------------------------------------------------
# Project-specific Targets
#------------------------------------------------------------------------------

# Run validator on all blueprints
.PHONY: run-all
run-all: build
	./$(BUILD_DIR)/$(BINARY_NAME) --all

#------------------------------------------------------------------------------
# Help (extends base help)
#------------------------------------------------------------------------------

.PHONY: help
help: help-base
	@echo ""
	@echo "$(YELLOW)Project-specific:$(NC)"
	@echo "  make run-all        Run validator on all blueprints"
