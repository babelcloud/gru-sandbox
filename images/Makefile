# Docker image names
REGISTRY ?= babelcloud
PLATFORMS ?= linux/amd64,linux/arm64

# Default Python version
PYTHON_VERSION ?= 3.13

# Default Node.js version for TypeScript
NODE_VERSION ?= 20.12

# ANSI color codes
CYAN := \033[36m
RESET := \033[0m

# Get all directories containing Dockerfile
IMAGES := $(wildcard */Dockerfile)
IMAGES := $(IMAGES:/Dockerfile=)

# Get all environment variables with BUILD_ARG_ prefix and convert them to build args
BUILD_ARGS := $(foreach v,$(filter BUILD_ARG_%,$(.VARIABLES)),--build-arg $(subst BUILD_ARG_,,$(v))=$($(v)))

# Default command for run target
CMD ?= /bin/bash

.PHONY: build-all clean-all buildx-all help build-% buildx-% clean-% run-% test-%

# Default target
.DEFAULT_GOAL := help

# Build all images locally
build-all: $(addprefix build-,$(IMAGES)) ## Build all images locally

# Build all images for multiple architectures
buildx-all: $(addprefix buildx-,$(IMAGES)) ## Build all images for multiple architectures

# Clean all images
clean-all: $(addprefix clean-,$(IMAGES)) ## Clean all images

# Pattern rule for local builds
build-%: %/Dockerfile ## Build image locally (e.g., build-python)
	docker build \
		$(BUILD_ARGS) \
		-t $(REGISTRY)/gbox-$*:latest \
		-f $< \
		$*

# Pattern rule for multi-arch builds
buildx-%: %/Dockerfile ## Build image for multiple architectures (e.g., buildx-python)
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		$(BUILD_ARGS) \
		-t $(REGISTRY)/gbox-$*:latest \
		-f $< \
		$*

# Pattern rule for cleaning specific image
clean-%: ## Clean specific image (e.g., clean-python)
	docker rmi $(REGISTRY)/gbox-$*:latest || true

# Pattern rule for running images
run-%: ## Run image with custom command (e.g., run-python CMD="python -c 'print(1+1)'")
	docker run --rm -it $(REGISTRY)/gbox-$*:latest $(CMD)

# Pattern rule for testing images
test-%: build-% ## Test image if test.sh exists (e.g., test-python)
	@if [ -f $*/test.sh ]; then \
		echo "$(CYAN)Running tests for $*...$(RESET)"; \
		cd $* && ./test.sh; \
	else \
		echo "$(CYAN)No test script found for $*$(RESET)"; \
	fi

# Show help
help: ## Show this help message
	@echo "Available targets:"
	@echo ""
	@echo "General targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | grep -v '%' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-25s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "Pattern rules (replace % with image name):"
	@grep -E '^[a-zA-Z0-9_-]+-%:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-25s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "Available images:"
	@for img in $(IMAGES); do \
		printf "  $(CYAN)%s$(RESET)\n" "$$img"; \
	done 