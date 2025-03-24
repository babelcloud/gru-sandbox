# Get version from environment variable or git commit
VERSION ?= $(shell git rev-parse --short HEAD)

# Distribution directory
DIST_DIR := dist
DIST_PACKAGE := $(DIST_DIR)/gbox-$(VERSION).tar.gz

# Check and enable pnpm via corepack
.PHONY: check-pnpm
check-pnpm: ## Check and enable pnpm via corepack
	@if ! command -v pnpm &> /dev/null; then \
		echo "Enabling pnpm via corepack..."; \
		corepack enable; \
		corepack prepare pnpm@latest --activate; \
	fi

# Show help
.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Build all components
.PHONY: build
build: check-pnpm ## Build all components
	@echo "Building components..."
	@make -C packages/api-server docker-build
	@cd packages/mcp-server && pnpm install && pnpm build

# Create distribution package
.PHONY: dist
dist: build ## Create distribution package
	@echo "Creating distribution package version $(VERSION)..."
	@rm -rf $(DIST_DIR)
	@mkdir -p $(DIST_DIR)

	# Create directory structure
	@mkdir -p $(DIST_DIR)/bin
	@mkdir -p $(DIST_DIR)/manifests
	@mkdir -p $(DIST_DIR)/packages/mcp-server

	# Copy files maintaining directory structure
	@cp -r bin/* $(DIST_DIR)/bin/
	@cp -r manifests/* $(DIST_DIR)/manifests/
	@rsync -av --exclude='node_modules' packages/mcp-server/ $(DIST_DIR)/packages/mcp-server/
	@cp LICENSE README.md $(DIST_DIR)/

	# Create tar.gz package
	@cd $(DIST_DIR) && tar -czf gbox-$(VERSION).tar.gz *
	@echo "Distribution package created: $(DIST_PACKAGE)"

# Build and push docker images
.PHONY: docker-push
docker-push: ## Build and push docker images
	@echo "Building and pushing docker images..."
	@make -C packages/api-server docker-push

# Clean distribution files
.PHONY: clean
clean: ## Clean distribution files
	@echo "Cleaning distribution files..."
	@rm -rf $(DIST_DIR)

# Default target
.DEFAULT_GOAL := help 