SHELL := /bin/bash
DATABASE_URL ?= postgres://lingora:lingora@localhost:5432/lingora?sslmode=disable

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: dev
dev: ## Run the api
	go run ./cmd/api

.PHONY: build
build: ## Compile every Go package
	go build ./...

.PHONY: test
test: ## Run unit tests with the race detector
	go test -race ./...

.PHONY: check
check: ## Format check, vet and build
	test -z "$$(gofmt -l .)" && go vet ./... && go build ./...

.PHONY: sqlc
sqlc: ## Regenerate sqlc code from db/query
	sqlc vet && sqlc generate

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

.PHONY: migrate-create
migrate-create: ## Create a migration pair: make migrate-create name=add_foo
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: up
up: ## Start local Postgres and Redis
	docker compose up -d

.PHONY: down
down: ## Stop local infrastructure
	docker compose down
