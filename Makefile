# MCKMT Makefile

.PHONY: help build test test-unit test-integration test-coverage mocks clean dev deps migrate proto clean-proto swagger

# Default target
help:
	@echo "MCKMT Development Commands:"
	@echo "  deps               - Start dependencies (PostgreSQL, Redis, etc.)"
	@echo "  dev                - Start full development environment (deps + hub + agent)"
	@echo "  build              - Build all binaries"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration tests only"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  mocks              - Generate mocks"
	@echo "  clean              - Clean build artifacts"
	@echo ""
	@echo "Database:"
	@echo "  migrate            - Run migrations"
	@echo "  migrate-status     - Check migration status"
	@echo "  create-migration   - Create new migration"
	@echo ""
	@echo "Code Generation:"
	@echo "  proto              - Generate protobuf code"
	@echo "  clean-proto        - Clean generated protobuf files"
	@echo "  swagger            - Generate Swagger docs"

# Build all binaries
build:
	@echo "Building binaries..."
	@mkdir -p bin
	go build -o bin/hub cmd/hub/main.go
	go build -o bin/agent cmd/agent/main.go
	go build -o bin/ctl cmd/ctl/main.go
	@echo "Build complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v ./internal/integration/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate mocks using go generate
mocks:
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated!"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Build Docker images
docker:
	@echo "Building Docker images..."
	docker build -f deployments/docker/Dockerfile.hub -t mckmt/hub:latest .
	docker build -f deployments/docker/Dockerfile.agent -t mckmt/agent:latest .

# Build specific Docker image
docker-hub:
	@echo "Building Hub Docker image..."
	docker build -f deployments/docker/Dockerfile.hub -t mckmt/hub:latest .

docker-agent:
	@echo "Building Agent Docker image..."
	docker build -f deployments/docker/Dockerfile.agent -t mckmt/agent:latest .

# Start dependencies only
deps:
	@echo "Starting dependencies..."
	cd deployments/docker && docker compose up -d postgres redis db-init prometheus grafana
	@echo "Dependencies started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

# Start full development environment
dev:
	@echo "Starting full development environment..."
	cd deployments/docker && docker compose up -d
	@echo "Full environment started!"
	@echo "Hub API: http://localhost:8080"
	@echo "Hub gRPC: localhost:8081"
	@echo "API Docs: http://localhost:8080/swagger/index.html"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Database migration commands (using golang-migrate CLI)
migrate:
	@echo "Running database migrations..."
	migrate -path ./migrations -database "postgres://mckmt:mckmt@localhost:5432/mckmt?sslmode=disable" up

migrate-status:
	@echo "Checking migration status..."
	migrate -path ./migrations -database "postgres://mckmt:mckmt@localhost:5432/mckmt?sslmode=disable" version

migrate-rollback:
	@echo "Rolling back last migration..."
	migrate -path ./migrations -database "postgres://mckmt:mckmt@localhost:5432/mckmt?sslmode=disable" down 1

migrate-force:
	@echo "Forcing migration version..."
	@read -p "Enter version number: " version; \
	migrate -path ./migrations -database "postgres://mckmt:mckmt@localhost:5432/mckmt?sslmode=disable" force $$version

# Create new migration files (manual process)
migrate-create:
	@echo "To create a new migration:"
	@echo "1. Create two files in migrations/ directory:"
	@echo "   - 000XXX_migration_name.up.sql"
	@echo "   - 000XXX_migration_name.down.sql"
	@echo "2. Use the next sequential number (XXX)"
	@echo "3. Add your SQL statements to the .up.sql file"
	@echo "4. Add rollback statements to the .down.sql file"

# Generate protobuf Go code
proto:
	@echo "Generating protobuf Go code..."
	@mkdir -p api/proto
	@PROTO_FILES=$$(find api/proto -name "*.proto"); \
	if [ -n "$$PROTO_FILES" ]; then \
		protoc --proto_path=. \
			--go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			--plugin=protoc-gen-go=$(shell go env GOPATH)/bin/protoc-gen-go \
			--plugin=protoc-gen-go-grpc=$(shell go env GOPATH)/bin/protoc-gen-go-grpc \
			$$PROTO_FILES; \
		echo "Protobuf code generated successfully!"; \
	else \
		echo "No proto files found in api/proto/"; \
	fi

# Clean generated protobuf files
clean-proto:
	@echo "Cleaning generated protobuf files..."
	@find api/proto -name "*.pb.go" -delete
	@echo "Protobuf files cleaned!"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/hub/main.go -o ./docs
	@echo "Swagger documentation generated in ./docs/"