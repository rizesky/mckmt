# MCKMT Makefile
# ===================================
# Global config
APP_NAME    := mckmt
VERSION     ?= latest
ENV         ?= dev
DATABASE_URL ?= postgres://mckmt:mckmt123@localhost:5432/mckmt?sslmode=disable

# Include modular makefiles
include make/build.mk
include make/test.mk
include make/database.mk
include make/docker.mk
include make/dev.mk
include make/demo.mk
include make/codegen.mk
include make/utils.mk

# Default target
.DEFAULT_GOAL := help

# Help system (self-documenting make)
.PHONY: help
help: ## Show this help
	@echo "Available make commands:"
	@echo ""
	@for file in $(MAKEFILE_LIST); do \
		grep -E '^[a-zA-Z_-]+:.*?## .*$$' "$$file" | \
		awk -F: 'BEGIN {FS = ":.*?## "} {target=$$1; gsub(/^[ \t]+|[ \t]+$$/, "", target); printf "\033[36m%-20s\033[0m %s\n", target, $$2}'; \
	done | sort
	@echo ""
	@echo "Environment Variables:"
	@echo "  APP_NAME    = $(APP_NAME)"
	@echo "  VERSION     = $(VERSION)"
	@echo "  ENV         = $(ENV)"
	@echo "  DATABASE_URL = $(DATABASE_URL)"