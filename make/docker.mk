# Docker Management
.PHONY: * 

docker: docker-hub docker-agent ## Build all Docker images

docker-hub: ## Build Hub Docker image
	@echo "Building Hub Docker image..."
	@docker build -f deployments/docker/Dockerfile.hub -t $(APP_NAME)/hub:$(VERSION) .
	@echo "✅ Hub Docker image built: $(APP_NAME)/hub:$(VERSION)"

docker-agent: ## Build Agent Docker image
	@echo "Building Agent Docker image..."
	@docker build -f deployments/docker/Dockerfile.agent -t $(APP_NAME)/agent:$(VERSION) .
	@echo "✅ Agent Docker image built: $(APP_NAME)/agent:$(VERSION)"

docker-push: ## Push Docker images to registry
	@echo "Pushing Docker images..."
	@docker push $(APP_NAME)/hub:$(VERSION)
	@docker push $(APP_NAME)/agent:$(VERSION)
	@echo "✅ Docker images pushed"
