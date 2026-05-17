# GheymatChi

GheymatChi is a local-first web application for tracking product prices over time. Users will eventually add products from e-commerce websites, attach source URLs, review historical price changes, and compare prices against market rates such as USD and gold.

The MVP starts with a simple modular monolith: a Go backend, a separate Go worker process, a Next.js + TypeScript frontend, and SQLite for local development. PostgreSQL is the planned production database target, so database code and schema decisions should keep that migration path in mind.

## Local Development Goal

The project should run on a MacBook with simple local commands and Docker Compose. The current backend API exposes health checks, connects to a local SQLite database for readiness checks, and provides product CRUD plus product source URL management endpoints.

## Initial Stack

- Backend: Go
- API routing: Go standard library or a small router in a later phase
- Worker: separate Go process for scheduled price checks
- Frontend: Next.js with TypeScript
- Local MVP database: SQLite
- Future database target: PostgreSQL
- Local orchestration: Docker Compose when services exist

## Local Commands

Run the API directly:

```sh
make migrate
make api
```

Run database migrations:

```sh
make migrate
```

Run backend tests:

```sh
make test
```

Run with Docker Compose:

```sh
docker compose run --rm migrate
make docker-up
```

## Current Phase

Phase 6 adds product source URL management. It includes configuration loading from environment variables, structured logging, graceful shutdown, a local SQLite database at `data/gheymatchi.db`, and these endpoints:

- `GET /healthz`
- `GET /readyz`
- `POST /api/products`
- `GET /api/products`
- `GET /api/products/{id}`
- `PATCH /api/products/{id}`
- `DELETE /api/products/{id}`
- `POST /api/products/{product_id}/sources`
- `GET /api/products/{product_id}/sources`
- `PATCH /api/products/{product_id}/sources/{source_id}`
- `DELETE /api/products/{product_id}/sources/{source_id}`

No frontend application code, scraping, price history API, worker scheduler, market-rate behavior, alert evaluation, or notification sending are implemented yet.
