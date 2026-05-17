.PHONY: api migrate test docker-up docker-down

api:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/api

migrate:
	cd backend && DB_PATH=../data/gheymatchi.db go run ./cmd/migrate

test:
	cd backend && go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
