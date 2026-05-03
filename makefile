.PHONY: dev-backend dev-frontend dev migrate-up migrate-down migrate-create



# Variables
BACKEND_DIR=apps/backend
FRONTEND_DIR=apps/frontend
DB_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable


# Development
dev-backend:
	@echo "Starting backend with Air..."
	cd $(BACKEND_DIR) && air

dev-frontend:
	@echo "Starting frontend..."
	cd $(FRONTEND_DIR) && pnpm dev

# Run both using a tool like 'concurrently' or simple backgrounding
# For a senior setup, we'll assume the user can use multiple terminals or we provide a single command
dev:
	@make -j 2 dev-backend dev-frontend

# Migrations (using golang-migrate)
# Install: brew install golang-migrate
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(BACKEND_DIR)/migrations -seq $$name

migrate-up:
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" down
