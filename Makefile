.PHONY: run build test lint migrate swagger clean help

# Variables
BINARY_NAME=payment-gateway
DOCKER_COMPOSE=docker compose

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Start all services with Docker Compose
	$(DOCKER_COMPOSE) up --build

down: ## Stop all services with Docker Compose
	$(DOCKER_COMPOSE) down

run-bg: ## Start services in background
	$(DOCKER_COMPOSE) up -d --build

test: ## Run all tests
	$(DOCKER_COMPOSE) run --rm app go test ./... -v -race -cover

test-unit: ## Run unit tests only
	$(DOCKER_COMPOSE) run --rm app go test ./internal/domain/... ./internal/usecase/... -v

test-integration: ## Run integration tests only
	$(DOCKER_COMPOSE) run --rm app go test ./internal/adapter/... -v -tags=integration

lint: ## Run linter
	$(DOCKER_COMPOSE) run --rm app golangci-lint run ./...

migrate-up: ## Run migrations up
	$(DOCKER_COMPOSE) run --rm app go run scripts/migrate.go up

migrate-down: ## Run migrations down
	$(DOCKER_COMPOSE) run --rm app go run scripts/migrate.go down

swagger: ## Generate Swagger docs
	$(DOCKER_COMPOSE) run --rm app swag init -g cmd/api/main.go -o docs/swagger

clean: ## Remove containers and volumes
	$(DOCKER_COMPOSE) down -v --remove-orphans

logs: ## Show Logs
	$(DOCKER_COMPOSE) logs -f app

ps: ## Show running containers
	$(DOCKER_COMPOSE) ps




