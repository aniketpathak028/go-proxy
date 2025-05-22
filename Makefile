BINARY_NAME=proxy-go
GO=go
BUILD_DIR=build
PKG_LIST=$(shell go list ./... 2>/dev/null)
LOG_FILE=/tmp/$(BINARY_NAME)-build.log
CONTAINER_IMAGE=proxy-go
CONTAINER_TAG=latest
.DEFAULT_GOAL := help

.PHONY: build clean run test help deps tidy podman-build podman-run podman-clean

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download -x=false 2>/dev/null
	@echo "✓ Dependencies downloaded"

tidy: ## Tidy up go.mod and go.sum
	@echo "Tidying up dependencies..."
	@$(GO) mod tidy -v=false 2>/dev/null
	@echo "✓ Dependencies tidied"

build: deps ## Build the proxy server
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -v=false -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go 2>$(LOG_FILE) || (cat $(LOG_FILE) && exit 1)
	@echo "✓ Build successful"

clean: ## Remove built binaries and artifacts
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR) 2>/dev/null || true
	@$(GO) clean -cache -testcache 2>/dev/null
	@echo "✓ Clean completed"

run: build ## Build and run the proxy server
	@echo "running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

test: deps ## Run all tests
	@echo "running tests..."
	@$(GO) test -v $(PKG_LIST)

vet: ## Run go vet
	@echo "running go vet..."
	@$(GO) vet $(PKG_LIST)

help: ## Display available commands
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

podman-build: ## Build container image using Podman
	@echo "Building container image..."
	@podman build -t $(CONTAINER_IMAGE):$(CONTAINER_TAG) .
	@echo "✓ Container image built"

podman-run: podman-build ## Run container using Podman
	@echo "Starting container..."
	@podman run --rm -p 8080:8080 $(CONTAINER_IMAGE):$(CONTAINER_TAG)

podman-clean: ## Remove container image
	@echo "Removing container image..."
	@podman rmi $(CONTAINER_IMAGE):$(CONTAINER_TAG) 2>/dev/null || true
	@echo "✓ Container image removed"