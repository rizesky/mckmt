# Code Generation
.PHONY: *

proto: ## Generate protobuf code
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto
	@echo "✅ Protobuf code generated"

clean-proto: ## Clean generated protobuf files
	@echo "Cleaning generated protobuf files..."
	@find . -name "*.pb.go" -delete
	@echo "✅ Generated protobuf files cleaned"

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/hub/main.go -o docs --parseDependency --parseInternal
	@echo "✅ Swagger documentation generated"
