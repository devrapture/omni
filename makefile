-include .env
-include apps/backend/.env
export

# Variables
BACKEND_DIR=apps/backend
FRONTEND_DIR=apps/frontend
DB_URL=$(DATABASE_URL)
ATLAS_ENV=DATABASE_URL="$(DATABASE_URL)" ATLAS_DEV_URL="$(ATLAS_DEV_URL)"
LOCAL_DB_NAME?=omni
ATLAS_DEV_DB_NAME?=omni-atlas

# ── Phony Targets ─────────────────────────────────────────────
.PHONY: dev dev-backend dev-frontend db-up db-down db-logs db-init db-reset \
        help migrate migrate-diff migrate-local migrate-staging migrate-prod \
        status status-staging status-prod dry-run-staging dry-run-prod lint

# ── Development ───────────────────────────────────────────────
dev-backend: ## Start backend with Air
	@echo "Starting backend with Air..."
	cd $(BACKEND_DIR) && air

dev-frontend: ## Start frontend
	@echo "Starting frontend..."
	cd $(FRONTEND_DIR) && pnpm dev

dev: ## Run both backend and frontend (use multiple terminals or backgrounding)
	@make -j 2 dev-backend dev-frontend

# ── Database ──────────────────────────────────────────────────
db-up: ## Start Postgres with Docker Compose
	@echo "Starting Postgres with Docker Compose..."
	docker compose up -d postgres

db-down: ## Stop Postgres container
	@echo "Stopping Postgres container..."
	docker compose down

db-logs: ## Follow Postgres logs
	docker compose logs -f postgres

db-init: db-up ## Create local and Atlas shadow databases if they do not exist
	@docker compose exec -T postgres psql -U postgres -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$(LOCAL_DB_NAME)'" | grep -q 1 || \
	docker compose exec -T postgres psql -U postgres -d postgres -c 'CREATE DATABASE "$(LOCAL_DB_NAME)";'
	@docker compose exec -T postgres psql -U postgres -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname = '$(ATLAS_DEV_DB_NAME)'" | grep -q 1 || \
	docker compose exec -T postgres psql -U postgres -d postgres -c 'CREATE DATABASE "$(ATLAS_DEV_DB_NAME)";'

db-reset: ## Recreate Postgres volume and rerun init scripts
	@echo "Recreating Postgres volume..."
	docker compose down -v
	docker compose up -d postgres

# ── Help ────────────────────────────────────────────────────
help: ## Show available commands
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Generate + Lint + Apply ───────────────────────────────────
migrate: db-init ## Generate, lint and apply locally (NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate NAME=add_users_table"; \
		exit 1; \
	fi
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate diff $(NAME) --env local
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate lint --env local --latest 1
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate apply --env local

# ── Generate ──────────────────────────────────────────────────
migrate-diff: db-init ## Generate migration (NAME=add_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-diff NAME=add_users_table"; \
		exit 1; \
	fi
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate diff $(NAME) --env local

# ── Apply ─────────────────────────────────────────────────────
migrate-local: ## Apply migrations locally
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate apply --env local

migrate-staging: ## Apply migrations to staging
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate apply --env staging

migrate-prod: ## Apply migrations to production
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate apply --env prod

# ── Status ────────────────────────────────────────────────────
status: ## Migration status (local)
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate status --env local

status-staging: ## Migration status (staging)
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate status --env staging

status-prod: ## Migration status (production)
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate status --env prod

# ── Dry Run ───────────────────────────────────────────────────
dry-run-staging: ## Dry run migrations on staging
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate apply --env staging --dry-run

dry-run-prod: ## Dry run migrations on production
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate apply --env prod --dry-run

# ── Lint ──────────────────────────────────────────────────────
lint: db-init ## Lint latest migration file
	cd $(BACKEND_DIR) && $(ATLAS_ENV) atlas migrate lint --env local --latest 1
