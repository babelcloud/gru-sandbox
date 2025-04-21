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

# Default container name prefix
CONTAINER_PREFIX ?= gbox

.PHONY: build-all clean-all buildx-all help build-% buildx-% clean-% run-% start-% stop-% vnc-% test-% ttyd-%

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

# Pattern rule for running images interactively (ephemeral)
run-%: build-% ## Run image interactively (removes container on exit)
	@echo "Running image $(REGISTRY)/$(CONTAINER_PREFIX)-$*:latest interactively..."
	docker run --rm -it -P $(REGISTRY)/$(CONTAINER_PREFIX)-$*:latest $(CMD)

# Pattern rule for starting images detached (persistent)
start-%: build-% ## Start image detached with a persistent name (e.g., start-python)
	@CONTAINER_NAME=$(CONTAINER_PREFIX)-$* ; \
	 if [ -z "$(shell docker ps -q -f name=$(CONTAINER_PREFIX)-$*)" ]; then \
		echo "Starting container $$CONTAINER_NAME in background..."; \
		docker run -d --name $$CONTAINER_NAME -P $(REGISTRY)/$(CONTAINER_PREFIX)-$*:latest; \
		echo "Container $$CONTAINER_NAME started. Use 'make stop-$*' to stop and remove."; \
	else \
		echo "Container $$CONTAINER_NAME is already running."; \
	fi

# Pattern rule for stopping and removing detached containers
stop-%: ## Stop and remove the detached container (e.g., stop-python)
	@CONTAINER_NAME=$(CONTAINER_PREFIX)-$* ; \
	 if [ -n "$(shell docker ps -a -q -f name=$(CONTAINER_PREFIX)-$*)" ]; then \
		echo "Stopping and removing container $$CONTAINER_NAME..."; \
		docker stop $$CONTAINER_NAME > /dev/null; \
		docker rm $$CONTAINER_NAME > /dev/null; \
		echo "Container $$CONTAINER_NAME stopped and removed."; \
	else \
		echo "Container $$CONTAINER_NAME not found."; \
	fi

# Pattern rule for connecting via VNC to a running detached container
vnc-%: start-% ## Connect to the running detached container via VNC (macOS only) (e.g., vnc-python)
	@CONTAINER_NAME=$(CONTAINER_PREFIX)-$* ; \
	 if [ -z "$(shell docker ps -q -f name=$(CONTAINER_PREFIX)-$*)" ]; then \
		 echo "Error: Container $$CONTAINER_NAME is not running. Use 'make start-$*' first." >&2; \
		 exit 1; \
	 fi; \
	 echo "Attempting to connect to VNC in container $$CONTAINER_NAME..."; \
	 VNC_PASS=$$(docker exec $$CONTAINER_NAME cat /root/.vnc/plaintext_passwd 2>/dev/null); \
	 if [ -z "$$VNC_PASS" ]; then \
		 echo "Error: Could not retrieve VNC password from /root/.vnc/plaintext_passwd in $$CONTAINER_NAME." >&2; \
		 echo "       Make sure the entrypoint script ran correctly." >&2; \
		 exit 1; \
	 fi; \
	 VNC_PORT_MAPPING=$$(docker port $$CONTAINER_NAME 5900/tcp); \
	 if [ -z "$$VNC_PORT_MAPPING" ]; then \
		 echo "Error: Could not get port mapping for 5900/tcp in $$CONTAINER_NAME." >&2; \
		 echo "       Ensure the container exposed port 5900 and was run with -P or -p." >&2; \
		 exit 1; \
	 fi; \
	 HOST_PORT=$$(echo $$VNC_PORT_MAPPING | sed 's/.*://'); \
	 VNC_URL="vnc://localhost:$$HOST_PORT"; \
	 echo "  VNC URL:        $$VNC_URL"; \
	 echo "  VNC Password:   $(CYAN)$$VNC_PASS$(RESET)"; \
	 echo "Launching Screen Sharing (macOS)..."; \
	 open $$VNC_URL

# Pattern rule for connecting via ttyd to a running detached container
ttyd-%: start-% ## Open ttyd web terminal for the running container (e.g., ttyd-python)
	@CONTAINER_NAME=$(CONTAINER_PREFIX)-$* ; \
	 if [ -z "$(shell docker ps -q -f name=$(CONTAINER_PREFIX)-$*)" ]; then \
		 echo "Error: Container $$CONTAINER_NAME is not running. Use 'make start-$*' first." >&2; \
		 exit 1; \
	 fi; \
	 echo "Attempting to open ttyd for container $$CONTAINER_NAME..."; \
	 TTYD_PORT_MAPPING=$$(docker port $$CONTAINER_NAME 7681/tcp); \
	 if [ -z "$$TTYD_PORT_MAPPING" ]; then \
		 echo "Error: Could not get port mapping for 7681/tcp in $$CONTAINER_NAME." >&2; \
		 echo "       Ensure the container exposed port 7681 and was run with -P or -p." >&2; \
		 exit 1; \
	 fi; \
	 HOST_PORT=$$(echo $$TTYD_PORT_MAPPING | sed 's/.*://'); \
	 TTYD_URL="http://localhost:$$HOST_PORT"; \
	 echo "  Container Name: $$CONTAINER_NAME"; \
	 echo "  ttyd URL:       $$TTYD_URL"; \
	 echo "Launching default web browser..."; \
	 open $$TTYD_URL

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