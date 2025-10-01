# Build System
.PHONY: *

build: ## Build all binaries
	@echo "Building all binaries..."
	@$(MAKE) build-hub build-agent

build-hub: ## Build hub binary
	@echo "Building Hub binary..."
	@mkdir -p bin
	@go build -o bin/hub -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" cmd/hub/main.go
	@echo "✅ Hub binary built: bin/hub"

build-agent: ## Build agent binary
	@echo "Building Agent binary..."
	@mkdir -p bin
	@go build -o bin/agent -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" cmd/agent/main.go
	@echo "✅ Agent binary built: bin/agent"

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf coverage.out coverage.html
	@go clean -cache
	@echo "✅ Build artifacts cleaned"
