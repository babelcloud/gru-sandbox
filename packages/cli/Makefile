# Makefile for gbox CLI

# Variable definitions
BINARY_NAME := gbox
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DIR := build
MAIN_FILE := main.go

# Default platform (based on current system)
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# LDFLAGS for embedding version information
LDFLAGS := -ldflags "-X github.com/babelcloud/gru-sandbox/packages/cli/internal/version.Version=$(VERSION) \
                     -X github.com/babelcloud/gru-sandbox/packages/cli/internal/version.BuildTime=$(BUILD_TIME) \
                     -X github.com/babelcloud/gru-sandbox/packages/cli/internal/version.Commit=$(COMMIT)"

# Supported platforms
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64

.PHONY: all clean help binary binary-all install e2e-setup e2e e2e-env-check e2e-cleanup e2e-%

# Default target
all: help ## Show this help message (default target)

# Help information
help: ## Show this help message
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Check if test API server is running
e2e-env-check: ## Check if test API server is running
	@if ! curl -s http://localhost:28081/api/v1/boxes > /dev/null; then \
		echo "Error: Test API server is not running. Please run 'make e2e-setup' first."; \
		exit 1; \
	fi

# Setup test environment
e2e-setup: ## Setup e2e test environment (build and start test API server)
	@echo "Setting up e2e test environment..."
	@echo "Starting e2e test API server..."
	@SERVICE_PORT=28081 \
	 PREFIX=test \
	 docker compose -f ../../manifests/docker/docker-compose.yml up -d --build
	@echo "Waiting for e2e test API server to be ready..."
	@for i in $$(seq 1 30); do \
		if curl -s http://localhost:28081/api/v1/boxes > /dev/null; then \
			echo "E2E test API server is ready!"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "Error: E2E test API server failed to start within 30 seconds."; \
	exit 1

# Run test for a specific command using pattern matching
.PHONY: e2e-%
e2e-%: e2e-env-check ## Run test for a specific command (e.g. make e2e-create)
	@cd e2e && ./run-test --$*.exp

# Run tests for all commands
e2e: e2e-env-check ## Run tests for all commands
	@cd e2e && ./run-test

# Clean build directory
clean: ## Clean the build directory
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleaning completed"

# Build binary for a single platform
binary: ## Build binary for the current platform (GOOS/GOARCH)
	@echo "Building $(BINARY_NAME) binary ($(GOOS)/$(GOARCH))..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build binaries for all supported platforms
binary-all: ## Build binaries for all supported platforms
	@echo "Building binaries for all supported platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d- -f1); \
		arch=$$(echo $$platform | cut -d- -f2); \
		outfile="$(BUILD_DIR)/$(BINARY_NAME)-$$platform"; \
		echo "Building $$outfile..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$outfile $(MAIN_FILE); \
	done
	@echo "All platform binaries have been built"

# Clean up test environment
e2e-cleanup: ## Clean up e2e test environment and stop test API server
	@echo "Cleaning up e2e test environment..."
	@PREFIX=test \
	 SERVICE_PORT=28081 \
	 docker compose -f ../../manifests/docker/docker-compose.yml down
	@echo "E2E test environment cleaned up"
