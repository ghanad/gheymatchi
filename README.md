# GheymatChi

GheymatChi is a local-first web application for tracking product prices over time. Users will eventually add products from e-commerce websites, attach source URLs, review historical price changes, and compare prices against market rates such as USD and gold.

The MVP starts with a simple modular monolith: a Go backend, a separate Go worker process, a Next.js + TypeScript frontend, and SQLite for local development. PostgreSQL is the planned production database target, so database code and schema decisions should keep that migration path in mind.

## Local Development Goal

The project should run on a MacBook with simple local commands and Docker Compose once application services are introduced. Phase 2 only creates the repository skeleton and documentation; runnable backend and frontend services are intentionally deferred to later phases.

## Initial Stack

- Backend: Go
- API routing: Go standard library or a small router in a later phase
- Worker: separate Go process for scheduled price checks
- Frontend: Next.js with TypeScript
- Local MVP database: SQLite
- Future database target: PostgreSQL
- Local orchestration: Docker Compose when services exist

## Planned Commands

These commands are expected to exist in later phases:

```sh
docker compose up
```

```sh
go test ./...
```

```sh
npm run dev
```

## Current Phase

Phase 2 creates only the repository structure and initial documentation placeholders. No backend application code, frontend application code, migrations, scraping, product CRUD, source tracking, price history, worker scheduler, market rates, alerts, or notifications are implemented yet.
