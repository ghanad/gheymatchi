ROOT_DIR := $(CURDIR)
GO_ENV := GOCACHE=$(ROOT_DIR)/.cache/go-build
DOCKER_UID := $(shell id -u)
DOCKER_GID := $(shell id -g)

.PHONY: api worker migrate seed test frontend-install frontend docker-migrate docker-seed docker-up docker-postgres-up docker-down

api:
	cd backend && $(GO_ENV) DB_PATH=../data/gheymatchi.db go run ./cmd/api

worker:
	cd backend && $(GO_ENV) DB_PATH=../data/gheymatchi.db WORKER_INTERVAL=5m go run ./cmd/worker

frontend:
	cd frontend && BACKEND_API_BASE_URL=http://localhost:8080 npm run dev

migrate:
	cd backend && $(GO_ENV) DB_PATH=../data/gheymatchi.db go run ./cmd/migrate

seed:
	cd backend && $(GO_ENV) DB_PATH=../data/gheymatchi.db go run ./cmd/seed

test:
	cd backend && $(GO_ENV) go test ./...

frontend-install:
	cd frontend && npm install

docker-migrate:
	DOCKER_UID=$(DOCKER_UID) DOCKER_GID=$(DOCKER_GID) docker compose run --rm migrate

docker-seed: docker-migrate
	DOCKER_UID=$(DOCKER_UID) DOCKER_GID=$(DOCKER_GID) docker compose run --rm seed

docker-up:
	DOCKER_UID=$(DOCKER_UID) DOCKER_GID=$(DOCKER_GID) docker compose up --build

docker-postgres-up:
	docker compose --profile postgres up --build postgres migrate-postgres api-postgres frontend-postgres worker-postgres

docker-down:
	docker compose down
