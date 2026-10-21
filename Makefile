# Nexo — development commands. Run `make help` to list them.

SHELL := bash
.DEFAULT_GOAL := help

BACKEND  := backend
FRONTEND := frontend
MIGRATIONS := db/migrations
DB_URL ?= postgres://nexo:nexo@localhost:5433/nexo?sslmode=disable
GOOSE := cd $(BACKEND) && go tool goose -dir $(MIGRATIONS) postgres "$(DB_URL)"
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: help
help: ## List the available commands
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## ---------- Setup ----------

.PHONY: setup
setup: ## Install dependencies, tools and git hooks
	@test -f $(BACKEND)/.env || cp .env.example $(BACKEND)/.env
	cd $(BACKEND) && go mod download
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/evilmartians/lefthook@latest
	cd $(FRONTEND) && pnpm install
	lefthook install

## ---------- Local environment ----------

.PHONY: up
up: ## Start local services (Postgres) and wait until they are healthy
	docker compose up -d --wait

.PHONY: down
down: ## Stop local services (data is kept)
	docker compose down

.PHONY: dev-api
dev-api: ## Run the API (http://localhost:8080)
	cd $(BACKEND) && go run -ldflags "-X main.version=$(VERSION)" ./cmd/api

.PHONY: dev-web
dev-web: ## Run the frontend (http://localhost:5173)
	cd $(FRONTEND) && pnpm dev

## ---------- Database ----------

.PHONY: db-migrate
db-migrate: ## Apply pending migrations
	$(GOOSE) up

.PHONY: db-rollback
db-rollback: ## Roll back the last migration
	$(GOOSE) down

.PHONY: db-status
db-status: ## Show migration status
	$(GOOSE) status

.PHONY: db-new
db-new: ## Create a migration: make db-new name=create_tasks
	@test -n "$(name)" || (echo "usage: make db-new name=<name>" && exit 1)
	$(GOOSE) create $(name) sql

.PHONY: db-reset
db-reset: ## Drop everything and migrate again (seed will be added here)
	$(GOOSE) reset
	$(GOOSE) up

## ---------- Code generation ----------

.PHONY: generate
generate: ## Regenerate code from the contract (Go server + TS client)
	cd $(BACKEND) && go generate ./...
	cd $(FRONTEND) && pnpm generate

## ---------- Quality ----------

.PHONY: lint
lint: ## Lint backend and frontend
	cd $(BACKEND) && golangci-lint run
	cd $(FRONTEND) && pnpm lint && pnpm format:check && pnpm typecheck

.PHONY: fmt
fmt: ## Format backend and frontend
	cd $(BACKEND) && gofmt -w .
	cd $(FRONTEND) && pnpm format

.PHONY: test
test: ## Unit tests (no Docker), including architecture rules
	cd $(BACKEND) && go test -race ./...
	cd $(FRONTEND) && pnpm test

.PHONY: test-integration
test-integration: ## Backend integration tests (needs Docker)
	cd $(BACKEND) && go test -race -tags integration ./...

.PHONY: test-all
test-all: test test-integration ## All tests

## ---------- Build ----------

.PHONY: build
build: ## Build the API binary and the frontend bundle
	cd $(BACKEND) && go build -ldflags "-X main.version=$(VERSION)" -o bin/api ./cmd/api
	cd $(FRONTEND) && pnpm build

.PHONY: docker-build
docker-build: ## Build the API container image
	docker build --build-arg VERSION=$(VERSION) -t nexo-api:$(VERSION) $(BACKEND)
