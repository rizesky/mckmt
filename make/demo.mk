# Demo Environments
# This file contains demo-specific targets for different scenarios

.PHONY: *

demo-oidc: ## Start complete demo with OIDC authentication (hub + monitoring + keycloak + kind clusters)
	@echo "🚀 Starting OIDC Demo Environment..."
	@echo "1. Starting core dependencies..."
	@cd deployments/docker && docker compose up -d postgres redis db-init
	@echo "2. Starting monitoring stack..."
	@cd deployments/docker && docker compose up -d prometheus grafana
	@echo "3. Starting OIDC provider..."
	@cd deployments/docker && docker compose up -d keycloak keycloak-db
	@echo "4. Setting up Keycloak..."
	@chmod +x scripts/setup-keycloak.sh
	@./scripts/setup-keycloak.sh
	@echo "5. Starting hub with OIDC configuration..."
	@cd deployments/docker && MCKMT_CONFIG_FILE=configs/demo/hub-oidc.yaml docker compose up -d hub
	@echo "6. Starting Kind clusters with agents..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh create 3
	@echo ""
	@echo "🎉 OIDC Demo Environment Ready!"
	@echo ""
	@echo "📋 Services:"
	@echo "   • MCKMT Hub: http://localhost:8080"
	@echo "   • Keycloak: http://localhost:8082 (admin/admin123)"
	@echo "   • Keycloak Admin: http://localhost:8082/admin"
	@echo "   • Prometheus: http://localhost:9090"
	@echo "   • Grafana: http://localhost:3000 (admin/admin)"
	@echo "👥 Test Users:"
	@echo "   • admin@mckmt.local / admin123 (admin role)"
	@echo "   • test@mckmt.local / test123 (viewer role)"
	@echo ""
	@echo "🔐 Test OIDC Login:"
	@echo "   Visit: http://localhost:8080/auth/oidc/login"
	@echo ""
	@echo "📊 Check Hub Status:"
	@echo "   curl http://localhost:8080/health"

demo-password: ## Start complete demo with password authentication (hub + monitoring + kind clusters, no OIDC)
	@echo "🚀 Starting Password Demo Environment..."
	@echo "1. Starting core dependencies..."
	@cd deployments/docker && docker compose up -d postgres redis db-init
	@echo "2. Starting monitoring stack..."
	@cd deployments/docker && docker compose up -d prometheus grafana
	@echo "3. Starting hub with password authentication..."
	@cd deployments/docker && MCKMT_CONFIG_FILE=configs/demo/hub-password.yaml docker compose up -d hub
	@echo "4. Starting Kind clusters with agents..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh create 3
	@echo ""
	@echo "🎉 Password Demo Environment Ready!"
	@echo ""
	@echo "📋 Services:"
	@echo "   • MCKMT Hub: http://localhost:8080"
	@echo "   • Prometheus: http://localhost:9090"
	@echo "   • Grafana: http://localhost:3000 (admin/admin)"
	@echo ""
	@echo "☸️  Kind Clusters:"
	@echo "   • mckmt-cluster-1: kind get kubeconfig --name mckmt-cluster-1"
	@echo "   • mckmt-cluster-2: kind get kubeconfig --name mckmt-cluster-2"
	@echo "   • mckmt-cluster-3: kind get kubeconfig --name mckmt-cluster-3"
	@echo ""
	@echo "🔐 Authentication:"
	@echo "   • Username/Password authentication enabled"
	@echo "   • OIDC authentication disabled"
	@echo ""
	@echo "📝 Register a user:"
	@echo "   POST http://localhost:8080/api/v1/auth/register"
	@echo ""
	@echo "📊 Check Hub Status:"
	@echo "   curl http://localhost:8080/health"

demo-stop: ## Stop all demo services
	@echo "🛑 Stopping demo services..."
	@cd deployments/docker && docker compose down
	@echo "🛑 Stopping Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh stop
	@echo "✅ Demo services stopped"

demo-clean: ## Clean up demo environment (containers, volumes, binaries, kind clusters)
	@echo "🧹 Cleaning demo environment..."
	@cd deployments/docker && docker compose down -v
	@echo "🧹 Cleaning Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh stop
	@docker volume prune -f
	@$(MAKE) clean
	@echo "✅ Demo environment cleaned"

demo-status: ## Show demo environment status (containers, clusters, services)
	@echo "📊 Demo Environment Status:"
	@echo ""
	@echo "🐳 Docker Containers:"
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep mckmt || echo "   No MCKMT containers running"
	@echo ""
	@echo "☸️  Kind Clusters:"
	@./scripts/kind-clusters.sh list 2>/dev/null || echo "   No Kind clusters found"
	@echo ""
	@echo "🌐 Services:"
	@echo "   • MCKMT Hub: $(shell curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null || echo "not running")"
	@echo "   • Keycloak: $(shell curl -s -o /dev/null -w "%{http_code}" http://localhost:8082 2>/dev/null || echo "not running")"
	@echo "   • Prometheus: $(shell curl -s -o /dev/null -w "%{http_code}" http://localhost:9090 2>/dev/null || echo "not running")"
	@echo "   • Grafana: $(shell curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 2>/dev/null || echo "not running")"

# Cluster management targets
clusters-setup: ## Setup Kind clusters with agents (usage: make clusters-setup COUNT=3)
	@echo "🚀 Setting up Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh create $(or $(COUNT),3)

clusters-stop: ## Stop Kind clusters (usage: make clusters-stop COUNT=3)
	@echo "🛑 Stopping Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh stop
