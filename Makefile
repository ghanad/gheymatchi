.PHONY: api worker migrate seed test frontend-install frontend docker-migrate docker-seed docker-up docker-down

api:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/api

worker:
	cd backend && DB_PATH=../data/gheymatchi.db WORKER_INTERVAL=5m go run ./cmd/worker

frontend:
	cd frontend && BACKEND_API_BASE_URL=http://localhost:8080 npm run dev

migrate:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/migrate

seed:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/seed

test:
	cd backend && go test ./...

frontend-install:
	cd frontend && npm install

docker-migrate:
	docker compose run --rm migrate

docker-seed:
	docker compose run --rm seed

docker-up: docker-migrate
	docker compose up --build

docker-down:
	docker compose down
