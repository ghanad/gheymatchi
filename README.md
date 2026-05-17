# GheymatChi

GheymatChi is a local-first web application for tracking product prices over time. Users will eventually add products from e-commerce websites, attach source URLs, review historical price changes, and compare prices against market rates such as USD and gold.

The MVP starts with a simple modular monolith: a Go backend, a separate Go worker process, a Next.js + TypeScript frontend, and SQLite for local development. PostgreSQL is the planned production database target, so database code and schema decisions should keep that migration path in mind.

## Local Development Goal

The project should run on a MacBook with simple local commands and Docker Compose. The current backend API skeleton is intentionally small: it exposes health checks only and does not connect to a database yet.

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
make api
```

Run backend tests:

```sh
make test
```

Run with Docker Compose:

```sh
make docker-up
```

## Current Phase

Phase 3 creates the minimal Go API service under `backend/`. It includes configuration loading from environment variables, structured logging, graceful shutdown, and these endpoints:

- `GET /healthz`
- `GET /readyz`

No frontend application code, migrations, database connection, scraping, product CRUD, source tracking, price history, worker scheduler, market rates, alerts, or notifications are implemented yet.
