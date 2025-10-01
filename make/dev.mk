# Development Environment
.PHONY: *

deps: ## Start core dependencies (PostgreSQL, Redis, DB init)
	@echo "Starting core dependencies..."
	@cd deployments/docker && docker compose up -d postgres redis db-init
	@echo "✅ Core dependencies started!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo ""
	@echo "To start monitoring stack (Prometheus + Grafana), run:"
	@echo "  make deps-monitoring"

deps-monitoring: ## Start monitoring stack (Prometheus, Grafana)
	@echo "Starting monitoring stack..."
	@cd deployments/docker && docker compose up -d prometheus grafana
	@echo "✅ Monitoring stack started!"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"

deps-oidc: ## Start OIDC provider (Keycloak)
	@echo "Starting OIDC provider (Keycloak)..."
	@cd deployments/docker && docker compose up -d keycloak keycloak-db
	@echo "✅ Keycloak started!"
	@echo "Keycloak: http://localhost:8082"
	@echo "Admin Console: http://localhost:8082/admin (admin/admin123)"
	@echo ""
	@echo "To set up Keycloak with MCKMT configuration, run:"
	@echo "  make setup-oidc"


dev: ## Start full development environment (deps + hub + agent)
	@echo "Starting full development environment..."
	@cd deployments/docker && docker compose up -d
	@echo "✅ Full environment started!"
	@echo "Hub API: http://localhost:8080"
	@echo "Hub gRPC: localhost:8081"
	@echo "API Docs: http://localhost:8080/swagger/index.html"
	@echo "Prometheus: http://localhost:9090"

dev-stop: ## Stop and remove development containers (keeps volumes)
	@echo "Stopping development environment..."
	@cd deployments/docker && docker compose down
	@echo "✅ Development environment stopped!"
	@echo "Volumes preserved. To cleanup everything including volumes, run: make dev-clean"

dev-clean: ## Stop and remove development containers and volumes
	@echo "Stopping development environment and removing volumes..."
	@cd deployments/docker && docker compose down -v
	@echo "✅ Development environment stopped and volumes removed!"
	@echo "All data has been deleted."
