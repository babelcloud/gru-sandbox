# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/babelcloud/gbox-api-server:latest

# Platform specific variables
PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_BUILDER ?= gbox-builder

# Show help
.PHONY: help
help: ## Show this help message
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' Makefile | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Run the API server in development mode
.PHONY: dev
dev: ## Run the API server in development mode
	go run main.go

# Build the docker image locally
.PHONY: docker-build
docker-build: ## Build docker image locally
	docker build -t ${IMG} .

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
		--push \
		--builder $$BUILDER \
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
	docker run -p 28080:28080 -v /var/run/docker.sock:/var/run/docker.sock ${IMG}
