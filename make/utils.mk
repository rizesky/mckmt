# Utility Scripts
# This file contains utility commands for MCKMT

.PHONY: *

# Kind cluster management
kind-create: ## Create Kind clusters with MCKMT agents (usage: make kind-create COUNT=3)
	@echo "🚀 Creating Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh create $(or $(COUNT),3)

kind-stop: ## Stop all Kind clusters
	@echo "🛑 Stopping Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh stop

kind-status: ## Show Kind cluster and agent status
	@echo "📊 Checking Kind cluster status..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh status

kind-list: ## List all Kind clusters
	@echo "📋 Listing Kind clusters..."
	@chmod +x scripts/kind-clusters.sh
	@./scripts/kind-clusters.sh list

# Dashboard access
dashboard: ## Open MCKMT monitoring dashboard
	@echo "🌐 Opening MCKMT dashboard..."
	@chmod +x scripts/open-dashboard.sh
	@./scripts/open-dashboard.sh

# Keyclock setup for OIDC
setup-keycloak: ## Configure Keycloak with MCKMT settings
	@echo "🔐 Setting up OIDC with Keycloak..."
	@chmod +x scripts/setup-keycloak.sh
	@./scripts/setup-keycloak.sh

# Kustomize testing
test-kustomize: ## Test Kustomize configurations
	@echo "🧪 Testing Kustomize configurations..."
	@chmod +x scripts/test-kustomize.sh
	@./scripts/test-kustomize.sh

test-kustomize-demo: ## Test demo Kustomize configuration
	@echo "🧪 Testing demo Kustomize configuration..."
	@chmod +x scripts/test-kustomize.sh
	@./scripts/test-kustomize.sh demo

test-kustomize-production: ## Test production Kustomize configuration
	@echo "🧪 Testing production Kustomize configuration..."
	@chmod +x scripts/test-kustomize.sh
	@./scripts/test-kustomize.sh production
