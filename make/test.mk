# Testing System
.PHONY: *

test: ## Run all tests
	@echo "Running all tests..."
	@go test ./... -v

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test $(shell go list ./... | grep -v /integration) -v

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test ./internal/integration/... -v

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

coverage.html: test-coverage ## Generate HTML coverage report

mocks: ## Generate mocks
	@echo "Generating mocks..."
	@go generate ./...
	@echo "✅ Mocks generated"
