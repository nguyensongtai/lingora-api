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

.PHONY: openapi
openapi: ## Sinh lại type Go từ openapi.yaml
	oapi-codegen -config oapi-codegen.yaml openapi.yaml

.PHONY: migrate-up
migrate-up: ## Apply all pending migrations
	migrate -path db/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Roll back the last migration
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

.PHONY: migrate-create
migrate-create: ## Create a migration pair: make migrate-create name=add_foo
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: seed
seed: ## Nạp dữ liệu mẫu cho môi trường dev (cần docker compose up -d)
	docker compose exec -T postgres psql -U lingora -d lingora -v ON_ERROR_STOP=1 < db/seed/001_sample_courses.sql
	docker compose exec -T postgres psql -U lingora -d lingora -v ON_ERROR_STOP=1 < db/seed/002_curriculum.sql
	docker compose exec -T postgres psql -U lingora -d lingora -v ON_ERROR_STOP=1 < db/seed/004_vocabulary.sql

.PHONY: verify-vocab
verify-vocab: ## Đối chiếu bậc CEFR của từ vựng với dữ liệu nguồn (cần mạng)
	python3 db/seed/verify-levels.py

.PHONY: seed-demo
seed-demo: ## Nạp tiến độ mẫu cho demo@lingora.vn (tạo tài khoản đó trước)
	docker compose exec -T postgres psql -U lingora -d lingora -v ON_ERROR_STOP=1 < db/seed/003_demo_progress.sql

.PHONY: createuser
createuser: ## Tạo tài khoản: make createuser email=... role=admin
	@read -rs -p "Mật khẩu: " PW && echo && printf '%s' "$$PW" | go run ./cmd/createuser -email "$(email)" -role "$(role)"

.PHONY: token
token: ## In ra access token admin để gọi thử API (cần JWT_SECRET)
	go run ./cmd/devtoken

.PHONY: up
up: ## Start local Postgres and Redis
	docker compose up -d

.PHONY: down
down: ## Stop local infrastructure
	docker compose down

.PHONY: test-integration
## Chạy cả integration test; cần docker compose up trước.
test-integration:
	TEST_DATABASE_URL="$(DATABASE_URL)" REDIS_URL="$(REDIS_URL)" go test -race ./...
