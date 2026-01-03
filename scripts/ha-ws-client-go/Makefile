# ha-ws-client-go Makefile
# Supports cross-compilation for Raspberry Pi and other ARM devices

#------------------------------------------------------------------------------
# Project-specific Configuration
#------------------------------------------------------------------------------

# Required variables for shared.mk
BINARY_NAME := ha-ws-client
PROJECT_NAME := ha-ws-client-go
GO_LOCAL_PATH := github.com/home-assistant-blueprints/ha-ws-client-go

# Optional overrides
VERSION ?= 1.5.1
BUILD_DIR := build
BUILD_PATH := ./cmd/ha-ws-client

#------------------------------------------------------------------------------
# Include shared Makefile
#------------------------------------------------------------------------------

include ../shared.mk

#------------------------------------------------------------------------------
# Project-specific Targets
#------------------------------------------------------------------------------

# Docker build (for Home Assistant add-on)
.PHONY: docker
docker:
	docker build -t ha-ws-client:$(VERSION) .

#------------------------------------------------------------------------------
# Help (extends base help)
#------------------------------------------------------------------------------

.PHONY: help
help: help-base
	@echo ""
	@echo "$(YELLOW)Project-specific:$(NC)"
	@echo "  make docker         Build Docker image for Home Assistant add-on"
