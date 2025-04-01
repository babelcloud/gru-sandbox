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

# Test configurations
TEST_API_IMAGE := babelcloud/gbox-api-server:test
SKIP_BUILD ?= false

# LDFLAGS for embedding version information
LDFLAGS := -ldflags "-X github.com/babelcloud/gru-sandbox/packages/cli/internal/version.Version=$(VERSION) \
                     -X github.com/babelcloud/gru-sandbox/packages/cli/internal/version.BuildTime=$(BUILD_TIME) \
                     -X github.com/babelcloud/gru-sandbox/packages/cli/internal/version.Commit=$(COMMIT)"

# Supported platforms
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

.PHONY: all clean help binary binary-all install test-setup test-box test-boxes test-env-check test-cleanup

# Default target
all: help

# Help information
help:
	@echo "Available commands:"
	@echo "  make binary        - Build binary for the current platform"
	@echo "  make binary-all    - Build binaries for all supported platforms"
	@echo "  make clean         - Clean the build directory"
	@echo "  make install       - Install to local system (current platform only)"
	@echo "  make test-setup    - Setup test environment (build and start test API server)"
	@echo "  make test-box CMD=<cmd>  - Run test for a specific box command"
	@echo "  make test-boxes    - Run tests for all box commands"
	@echo "  make test-cleanup  - Clean up test environment and stop test API server"
	@echo "  make help          - Display this help information"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS               - Target operating system (default: $(GOOS))"
	@echo "  GOARCH             - Target architecture (default: $(GOARCH))"
	@echo "  VERSION            - Version number (default: $(VERSION))"
	@echo "  SKIP_BUILD         - Skip building test API server image (default: false)"

# Check if test API server is running
test-env-check:
	@if ! curl -s http://localhost:28081/api/v1/boxes > /dev/null; then \
		echo "Error: Test API server is not running. Please run 'make test-setup' first."; \
		exit 1; \
	fi

# Setup test environment
test-setup:
	@echo "Setting up test environment..."
	@if [ "$(SKIP_BUILD)" = "false" ]; then \
		echo "Building test API server image..."; \
		cd ../api-server && make docker-build IMG=$(TEST_API_IMAGE) || exit 1; \
	else \
		if ! docker inspect $(TEST_API_IMAGE) > /dev/null 2>&1; then \
			echo "Error: Test API server image not found. Please run without SKIP_BUILD=true first."; \
			exit 1; \
		fi; \
	fi
	@echo "Starting test API server..."
	@API_SERVER_IMG_TAG=test docker compose -f ../../manifests/docker/docker-compose.test.yml up -d
	@echo "Waiting for test API server to be ready..."
	@for i in $$(seq 1 30); do \
		if curl -s http://localhost:28081/api/v1/boxes > /dev/null; then \
			echo "Test API server is ready!"; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "Error: Test API server failed to start within 30 seconds."; \
	exit 1

# Run test for a specific box command
test-box: test-env-check
	@if [ -z "$(CMD)" ]; then \
		echo "Error: Please specify the command to test using CMD=<cmd>"; \
		echo "Example: make test-box CMD=box_create"; \
		exit 1; \
	fi
	@cd exp-test && ./run_test --$(CMD)

# Run tests for all box commands
test-boxes: test-env-check
	@cd exp-test && ./run_test

# Clean build directory
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleaning completed"

# Build binary for a single platform
binary:
	@echo "Building $(BINARY_NAME) binary ($(GOOS)/$(GOARCH))..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build binaries for all supported platforms
binary-all:
	@echo "Building binaries for all supported platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d- -f1); \
		arch=$$(echo $$platform | cut -d- -f2); \
		if [ "$$os" = "windows" ]; then \
			suffix=".exe"; \
		else \
			suffix=""; \
		fi; \
		outfile="$(BUILD_DIR)/$(BINARY_NAME)-$$platform$$suffix"; \
		echo "Building $$outfile..."; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $$outfile $(MAIN_FILE); \
	done
	@echo "All platform binaries have been built"

# Clean up test environment
test-cleanup:
	@echo "Cleaning up test environment..."
	@API_SERVER_IMG_TAG=test docker compose -f ../../manifests/docker/docker-compose.test.yml down
	@echo "Test environment cleaned up"
