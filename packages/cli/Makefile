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
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

.PHONY: all clean help binary binary-all install

# Default target
all: help

# Help information
help:
	@echo "Available commands:"
	@echo "  make binary        - Build binary for the current platform"
	@echo "  make binary-all    - Build binaries for all supported platforms"
	@echo "  make clean         - Clean the build directory"
	@echo "  make install       - Install to local system (current platform only)"
	@echo "  make help          - Display this help information"
	@echo ""
	@echo "Environment variables:"
	@echo "  GOOS               - Target operating system (default: $(GOOS))"
	@echo "  GOARCH             - Target architecture (default: $(GOARCH))"
	@echo "  VERSION            - Version number (default: $(VERSION))"

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
