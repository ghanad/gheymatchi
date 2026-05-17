.PHONY: api test docker-up docker-down

api:
	cd backend && go run ./cmd/api

test:
	cd backend && go test ./...

docker-up:
	docker compose up --build

docker-down:
	docker compose down
