.PHONY: help build run test docker-up docker-down docker-logs migrate-up migrate-down clean

ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

DOCKER_COMPOSE=docker-compose
MIGRATE=$(shell go env GOPATH)/bin/migrate

POSTGRES_USER ?= root
POSTGRES_PASSWORD ?= root
POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_DB ?= root

TEST_POSTGRES_DB ?= root_test

DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
TEST_DATABASE_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(TEST_POSTGRES_DB)?sslmode=disable

help: ## Показать помощь
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'

docker-up: ## Запустить все сервисы через docker-compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up --build -d && docker compose logs -f
	@echo "Services started. App available at http://localhost:${HTTP_PORT}"

docker-down: ## Остановить все сервисы
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down
	@echo "Services stopped."

docker-down-volumes: ## Остановить все сервисы и удалить volumes
	@echo "Stopping services and removing volumes..."
	$(DOCKER_COMPOSE) down -v
	@echo "Services stopped and volumes removed."

docker-restart: ## Перезапустить приложение
	@echo "Restarting app..."
	$(DOCKER_COMPOSE) restart app
	@echo "App restarted."

migrate-install:
	@echo "Installing migrate..."
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up: migrate-install ## Применить все миграции
	@echo "Running migrations..."
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" up
	@echo "Migrations applied."

migrate-down: migrate-install ## Откатить все миграции
	@echo "Rolling back migrations..."
	$(MIGRATE) -path migrations -database "$(DATABASE_URL)" down
	@echo "Migrations rolled back."

test-db-migrate: migrate-install
	@echo "Running migrations on test database..."
	@$(MIGRATE) -path migrations -database "$(TEST_DATABASE_URL)" up
	@echo "Test database migrated."

test-db-drop:
	@echo "Dropping test database..."
	@psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/postgres" \
		-c "DROP DATABASE IF EXISTS $(TEST_POSTGRES_DB);" 2>/dev/null || true
	@echo "Test database dropped."


test-db-create:
	@echo "Creating test database..."
	@psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/postgres" \
		-c "DROP DATABASE IF EXISTS $(TEST_POSTGRES_DB);" 2>/dev/null || true
	@psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/postgres" \
		-c "CREATE DATABASE $(TEST_POSTGRES_DB);"
	@echo "Test database '$(TEST_POSTGRES_DB)' created."

test: test-db-create test-db-migrate ## Запустить тесты
	@echo "Running integration tests with coverage..."
	@cd test && TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -v -cover -coverpkg=../pkg/repo/pg,../pkg/domain,../pkg/usecase
	@$(MAKE) test-db-drop