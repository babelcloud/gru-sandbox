# Docker image names
REGISTRY ?= babelcloud
PLATFORMS ?= linux/amd64,linux/arm64

# Default Python version
PYTHON_VERSION ?= 3.13

# Default Tags for base images
PYTHON_TAG ?= latest
PLAYWRIGHT_TAG ?= latest
TTYD_VERSION ?= 1.7.7 # Default ttyd version

# Default Node.js version for TypeScript
NODE_VERSION ?= 20.12

# ANSI color codes
CYAN := \033[36m
RESET := \033[0m
YELLOW := \033[33m

# Get all directories containing Dockerfile
IMAGES := $(wildcard */Dockerfile)
IMAGES := $(IMAGES:/Dockerfile=)

# Get all environment variables with BUILD_ARG_ prefix and convert them to build args
# Pass common ARGs always
BUILD_ARGS := --build-arg PYTHON_VERSION=$(PYTHON_VERSION) \
              --build-arg PYTHON_TAG=$(PYTHON_TAG) \
              --build-arg PLAYWRIGHT_TAG=$(PLAYWRIGHT_TAG) \
              --build-arg TTYD_VERSION=$(TTYD_VERSION)
BUILD_ARGS += $(foreach v,$(filter BUILD_ARG_%,$(.VARIABLES)),--build-arg $(subst BUILD_ARG_,,$(v)))

# Default command for run target
CMD ?= /bin/bash

# Default container name prefix
CONTAINER_PREFIX ?= gbox-test

.PHONY: build-all clean-all buildx-all help build-% buildx-% clean-% run-% start-% stop-% viewer test-% ttyd-%

# Default target
.DEFAULT_GOAL := help

# Build all images locally
build-all: $(addprefix build-,$(IMAGES)) ## Build all images locally
	@:

# Build all images for multiple architectures
buildx-all: $(addprefix buildx-,$(IMAGES)) ## Build all images for multiple architectures
	@:

# Clean all images
clean-all: $(addprefix clean-,$(IMAGES)) ## Clean all images
	@:

# Stop and remove all containers matching the prefix
stop-all: $(addprefix stop-,$(IMAGES)) ## Stop and remove all containers started by 'make start-*' (for known images)
	@:

# Pattern rule for local builds
build-%: %/Dockerfile ## Build image locally (e.g., build-python)
	docker build \
		$(BUILD_ARGS) \
		-t ${REGISTRY}/gbox-$*:latest \
		-f $< \
		$*

# Pattern rule for multi-arch builds
buildx-%: %/Dockerfile ## Build image for multiple architectures (e.g., buildx-python)
	docker buildx build \
		--platform $(PLATFORMS) \
		--push \
		$(BUILD_ARGS) \
		-t ${REGISTRY}/gbox-$*:latest \
		-f $< \
		$*

# Pattern rule for cleaning specific image
clean-%: ## Clean specific image (e.g., clean-python)
	docker rmi ${REGISTRY}/gbox-$*:latest || true

# Pattern rule for running images interactively (ephemeral)
run-%: build-% ## Run image interactively (removes container on exit)
	@echo "Running image ${REGISTRY}/gbox-$*:latest interactively..."
	docker run --rm -it -P ${REGISTRY}/gbox-$*:latest $(CMD)

# Pattern rule for starting images detached (persistent)
start-%: build-% ## Start image detached with a persistent name (e.g., start-python)
	@CONTAINER_NAME='${CONTAINER_PREFIX}-$*' ; \
	 if ! docker container inspect "$${CONTAINER_NAME}" > /dev/null 2>&1; then \
		echo "Starting container $${CONTAINER_NAME} from image ${REGISTRY}/gbox-$*:latest..." ; \
		docker run -d --name "$${CONTAINER_NAME}" -P ${REGISTRY}/gbox-$*:latest; \
		echo "Container $${CONTAINER_NAME} started. Use 'make stop-$*' to stop and remove." ; \
	else \
		echo "Container $${CONTAINER_NAME} already exists (may be stopped or running)." ; \
	fi

# Pattern rule for stopping and removing detached containers
stop-%: ## Stop and remove the detached container (e.g., stop-python)
	@CONTAINER_NAME='${CONTAINER_PREFIX}-$*' ; \
	 if docker container inspect "$${CONTAINER_NAME}" > /dev/null 2>&1; then \
		echo "Stopping and removing container $${CONTAINER_NAME}..." ; \
		docker stop "$${CONTAINER_NAME}" > /dev/null 2>&1 ; \
		docker rm "$${CONTAINER_NAME}" > /dev/null 2>&1 ; \
		echo "Container $${CONTAINER_NAME} stopped and removed." ; \
	else \
		echo "Container $${CONTAINER_NAME} not found." ; \
	fi

# Pattern rule for opening noVNC web interface and showing stream URL for a running detached container
viewer: start-viewer ## Open the noVNC web interface and show stream URL for the running container (e.g., novnc-viewer)
	@./scripts/viewer.sh "${CONTAINER_PREFIX}-viewer"

# Pattern rule for connecting via ttyd to a running detached container
ttyd-%: start-% ## Open ttyd web terminal for the running container (e.g., ttyd-python)
	@CONTAINER_NAME='${CONTAINER_PREFIX}-$*' ; \
	 if ! docker container inspect "$${CONTAINER_NAME}" > /dev/null 2>&1; then \
		 echo "Error: Container $${CONTAINER_NAME} is not running. Use 'make start-$*' first." >&2 ; \
		 exit 1 ; \
	 fi ; \
	 echo "Attempting to open ttyd for container $${CONTAINER_NAME}..." ; \
	 TTYD_PORT_MAPPING=$$(docker port "$${CONTAINER_NAME}" 7681/tcp) ; \
	 if [ -z "$$TTYD_PORT_MAPPING" ]; then \
		 echo "Error: Could not get port mapping for 7681/tcp in $${CONTAINER_NAME}." >&2 ; \
		 echo "       Ensure the container exposed port 7681 and was run with -P or -p." >&2 ; \
		 exit 1 ; \
	 fi ; \
	 HOST_PORT=$$(echo $$TTYD_PORT_MAPPING | sed 's/.*://') ; \
	 TTYD_URL="http://localhost:$$HOST_PORT" ; \
	 echo "  Container Name: $${CONTAINER_NAME}" ; \
	 echo "  ttyd URL:       $(CYAN)$$TTYD_URL$(RESET)" ; \
	 echo "Launching default web browser..." ; \
	 open $$TTYD_URL

# Pattern rule for testing images
test-%: start-% ## Test box-$* image using the running container $(CONTAINER_NAME)
	@CONTAINER_NAME='${CONTAINER_PREFIX}-$*' ; \
	 TEST_SCRIPT="$*/test.sh"; \
	 if [ -f "$$TEST_SCRIPT" ]; then \
		echo "$(CYAN)Running tests for $* against container $${CONTAINER_NAME}...$(RESET)" ; \
		cd $* && ./test.sh "$${CONTAINER_NAME}" ; \
	else \
		echo "$(CYAN)No test script found for $* ($$TEST_SCRIPT)$(RESET)" ; \
		echo "$(YELLOW)Checking contents of $*...$(RESET)" ; \
		ls -l "$*" ; \
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

# --- Explicit Build Dependencies --- #
# These ensure images are built in the correct order when using build-all
build-playwright: build-python
build-viewer: build-playwright