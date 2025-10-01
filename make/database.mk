# Database Management
.PHONY: *

migrate: migrate-up ## Run migrations

migrate-status: ## Check migration status
	@echo "Checking migration status..."
	@migrate -path migrations -database "$(DATABASE_URL)" version

migrate-up: ## Run pending migrations
	@echo "Running pending migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ Migrations completed"

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "✅ Migration rolled back"

create-migration: ## Create new migration (usage: make create-migration NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make create-migration NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✅ Migration created: migrations/*_$(NAME).sql"
