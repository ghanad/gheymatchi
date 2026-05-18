.PHONY: api worker migrate test frontend docker-up docker-down

api:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/api

worker:
	cd backend && DB_PATH=../data/gheymatchi.db WORKER_INTERVAL=5m go run ./cmd/worker

frontend:
	cd frontend && BACKEND_API_BASE_URL=http://localhost:8080 npm run dev

migrate:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/migrate

test:
	cd backend && go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
