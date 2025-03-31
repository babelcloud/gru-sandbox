# Get version from environment variable or git commit
VERSION ?= $(shell git rev-parse --short HEAD)

# Distribution directory
DIST_DIR := dist
DIST_PACKAGE := $(DIST_DIR)/gbox-$(VERSION).tar.gz

# Function to get git commit hash for a path
define get_git_hash
$(shell git log --pretty=tformat:"%h" -n1 -- $(1))
endef

# Image tags
API_SERVER_TAG := $(call get_git_hash,packages/api-server)
PY_IMG_TAG := $(call get_git_hash,images/python)
TS_IMG_TAG := $(call get_git_hash,images/typescript)

# Function to write env var to file (usage: $(call write_env,FILE,VAR,VALUE))
define write_env
	@echo "$(2)=$(3)" > $(1)/.env
endef

define append_env
	@echo "$(2)=$(3)" >> $(1)/.env
endef

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
	@cd packages/mcp-server && pnpm install && pnpm build

# Build docker images
.PHONY: build-images
build-images: ## Build all docker images
	@echo "Building all docker images..."
	@make -C images build-all

.PHONY: build-image-%
build-image-%: ## Build specific docker image (e.g., build-image-python)
	@echo "Building docker image $*..."
	@make -C images build-$*

run-container-%: ## Run specific docker image (e.g., run-container-python)
	@echo "Running docker container $*..."
	@make -C images run-$*

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

	# Generate .env files
	$(call write_env,$(DIST_DIR)/manifests/docker,API_SERVER_IMG_TAG,$(API_SERVER_TAG))
	$(call write_env,$(DIST_DIR)/packages/mcp-server,PY_IMG_TAG,$(PY_IMG_TAG))
	$(call append_env,$(DIST_DIR)/packages/mcp-server,TS_IMG_TAG,$(TS_IMG_TAG))

	# Create tar.gz package
	@cd $(DIST_DIR) && tar -czf gbox-$(VERSION).tar.gz *
	@cd $(DIST_DIR) && sha256sum gbox-$(VERSION).tar.gz > gbox-$(VERSION).tar.gz.sha256
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

api-dev: ## Start api server
	@echo "Starting api server..."
	@make -C packages/api-server dev

api: ## Start api server with docker compose
	@cd manifests/docker && docker compose up --build

mcp-dev: ## Start mcp server
	@echo "Starting mcp server..."
	@cd packages/mcp-server && pnpm dev

mcp-inspect: ## Start mcp server
	@echo "Starting mcp server..."
	@cd packages/mcp-server && pnpm inspect

mcp: build ## Start mcp server with distribution package
	@echo "Starting mcp server with distribution package..."
	@cd packages/mcp-server && pnpm inspect:dist

# Default target
.DEFAULT_GOAL := help 