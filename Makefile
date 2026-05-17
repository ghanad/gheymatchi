.PHONY: api migrate test frontend docker-up docker-down

api:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/api

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
