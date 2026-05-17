# GheymatChi

GheymatChi is a local-first web application for tracking product prices over time. Users will eventually add products from e-commerce websites, attach source URLs, review historical price changes, and compare prices against market rates such as USD and gold.

The MVP starts with a simple modular monolith: a Go backend, a separate Go worker process, a Next.js + TypeScript frontend, and SQLite for local development. PostgreSQL is the planned production database target, so database code and schema decisions should keep that migration path in mind.

## Local Development Goal

The project should run on a MacBook with simple local commands and Docker Compose. The current backend API skeleton is intentionally small: it exposes health checks and connects to a local SQLite database for readiness checks.

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

Phase 4 adds SQLite setup and migrations. It includes configuration loading from environment variables, structured logging, graceful shutdown, a local SQLite database at `data/gheymatchi.db`, and these endpoints:

- `GET /healthz`
- `GET /readyz`

No frontend application code, scraping, product CRUD API, source tracking API, price history API, worker scheduler, market-rate behavior, alert evaluation, or notification sending are implemented yet.
