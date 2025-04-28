# ANSI Color Variables
BOLD    := $(shell tput bold)
GREEN   := $(shell tput setaf 2)
BLUE    := $(shell tput setaf 4)
RESET   := $(shell tput sgr0)

# Image URL to use all building/pushing image targets
IMG ?= babelcloud/gbox-api-server:latest

# Common prefix for Go packages
MODULE_PREFIX := github.com/babelcloud/gbox

# Version information variables
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Version package path
VERSION_PKG := $(MODULE_PREFIX)/packages/api-server/internal/misc/service

# Platform specific variables
PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?= gbox-builder

# Binary build variables
BINARY_NAME ?= api-server
BUILD_DIR ?= build
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# LDFLAGS for embedding version information
LDFLAGS := -ldflags "-X $(VERSION_PKG).Version=$(VERSION) \
                    -X $(VERSION_PKG).BuildTime=$(BUILD_TIME) \
                    -X $(VERSION_PKG).CommitID=$(COMMIT_ID)"

# Show help
.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FSv = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Run the API server in development mode
.PHONY: dev
dev: ## Run the API server in development mode
	@DEBUG=true go run $(LDFLAGS) ./cmd/app/main.go

# Build binary for current platform
.PHONY: binary
binary: ## Build binary for current platform
	@echo "Building binary for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/app/main.go
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build binaries for all platforms
.PHONY: binary-all
binary-all: ## Build binaries for all platforms
	@echo "Building binaries for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in linux darwin; do \
		for arch in amd64 arm64; do \
			echo "Building for $$platform/$$arch..."; \
			CGO_ENABLED=0 GOOS=$$platform GOARCH=$$arch go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$platform-$$arch ./cmd/app/main.go; \
		done; \
	done
	@echo "Binaries built in $(BUILD_DIR)"

# Build the docker image locally
.PHONY: docker-build
docker-build: ## Build docker image locally
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_ID=$(COMMIT_ID) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		-t ${IMG} .

# Build multi-architecture docker image
.PHONY: docker-buildx
docker-buildx: ## Build multi-architecture docker image
	@if docker buildx ls | grep -q "default"; then \
		echo "Using default builder..."; \
		BUILDER="default"; \
	elif docker buildx ls | grep -q "$(BUILDX_BUILDER)"; then \
		echo "Using existing $(BUILDX_BUILDER) builder..."; \
		BUILDER="$(BUILDX_BUILDER)"; \
	else \
		echo "Creating new $(BUILDX_BUILDER) builder..."; \
		docker buildx create --name $(BUILDX_BUILDER) --driver docker-container --bootstrap; \
		BUILDER="$(BUILDX_BUILDER)"; \
	fi
	docker buildx build --platform $(PLATFORMS) \
		-t ${IMG} \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT_ID=$(COMMIT_ID) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--push \
		--builder "$$BUILDER" \
		-f Dockerfile \
		.

# Push the docker image
.PHONY: docker-push
docker-push: docker-buildx ## Build and push multi-architecture docker image

# Check and create Docker socket symlink if needed
.PHONY: check-docker-socket
check-docker-socket: ## Check and create Docker socket symlink if needed
	@if [ ! -L /var/run/docker.sock ]; then \
		echo "Creating Docker socket symlink..."; \
		sudo ln -sf $(HOME)/.docker/run/docker.sock /var/run/docker.sock; \
	fi

# Run the docker container locally
.PHONY: docker-run
docker-run: docker-build check-docker-socket ## Run docker container locally
	docker run -p 28080:28080 --rm -v /var/run/docker.sock:/var/run/docker.sock ${IMG}

# Serve the vision test HTML page locally (Foreground) and open browser (macOS)
.PHONY: serve-vision-test
serve-vision-test: ## Serve vision-test.html on port 8070 (Foreground), open browser (macOS)
	@bash -c 'sleep 0.3 && open http://localhost:8070/vision-test.html || echo "Failed to open browser automatically."' &
	@echo "Serving vision-test.html at: $(BOLD)$(BLUE)http://localhost:8070/vision-test.html$(RESET)"
	@echo "Press $(BOLD)$(GREEN)Ctrl+C$(RESET) to stop the server."
	@cd internal/browser/testdata && python -m http.server -b 127.0.0.1 8070
