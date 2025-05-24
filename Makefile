CONTAINER_IMAGE=proxy-go
CONTAINER_TAG?=latest
PORT?=8080
.DEFAULT_GOAL := help

.PHONY: podman-build podman-run podman-clean help

podman-build: ## Build container image using Podman
	@echo "Building container image..."
	@podman build -t $(CONTAINER_IMAGE):$(CONTAINER_TAG) .
	@echo "✓ Container image built"

podman-run: podman-build ## Run container using Podman
	@echo "Starting container on port $(PORT)..."
	@podman run --rm -p $(PORT):8080 $(CONTAINER_IMAGE):$(CONTAINER_TAG)

podman-clean: ## Remove container image
	@echo "Removing container image..."
	@podman rmi $(CONTAINER_IMAGE):$(CONTAINER_TAG) 2>/dev/null || true
	@echo "✓ Container image removed"

help: ## Display available commands
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'