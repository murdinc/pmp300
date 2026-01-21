.PHONY: build install clean test help

BINARY_NAME=pmp300
BUILD_DIR=build
GO=go

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the CLI tool
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME)"

install: ## Install the CLI tool to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install .
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "✓ Clean complete"

test: ## Run tests
	@echo "Running tests..."
	@$(GO) test -v ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "✓ Format complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@$(GO) vet ./...
	@echo "✓ Vet complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@echo "✓ Dependencies downloaded"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@$(GO) mod tidy
	@echo "✓ Tidy complete"

all: deps fmt vet build ## Download deps, format, vet, and build

.DEFAULT_GOAL := help
