.PHONY: help build test lint clean install run dev

# Variables
BINARY_NAME=mcp-go
MAIN_PATH=./cmd/mcp-go
BUILD_DIR=./bin
GO=go
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags for optimization
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)"
BUILD_FLAGS=-trimpath

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary with optimizations
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GO) build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

test: ## Run tests with race detection and coverage
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out -covermode=atomic ./...

test-coverage: test ## Run tests and show coverage report
	@echo "Coverage report:"
	$(GO) tool cover -func=coverage.out

test-coverage-html: test ## Run tests and open coverage in browser
	@echo "Opening coverage in browser..."
	$(GO) tool cover -html=coverage.out

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

clean: ## Clean build artifacts and cache
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out
	$(GO) clean -cache -testcache -modcache

install: build ## Install binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

run: build ## Build and run the server
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode with race detection
	@echo "Running in development mode..."
	$(GO) run -race $(MAIN_PATH)

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	$(GO) mod tidy

mod-verify: ## Verify go modules
	@echo "Verifying go modules..."
	$(GO) mod verify

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download

.DEFAULT_GOAL := help

