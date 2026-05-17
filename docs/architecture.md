# Architecture

GheymatChi starts as a modular monolith. The backend should be organized around clear internal modules, with HTTP handlers coordinating request parsing and response writing while application and domain code own business behavior.

The API service and worker service should be separate Go binaries. They can share internal packages where appropriate, but they should have separate entry points under `backend/cmd/api` and `backend/cmd/worker`.

## Principles

- Keep backend and frontend clearly separated.
- Keep business logic outside HTTP handlers.
- Keep database access behind store or repository interfaces.
- Use SQLite for the local MVP.
- Preserve a practical path to PostgreSQL later.
- Avoid microservices for the MVP.
- Avoid unnecessary abstractions before behavior exists.

## Current Phase Boundary

Phase 6 adds a focused source module under `backend/internal/source`. HTTP handlers parse requests and write responses, while validation and SQLite persistence live in the source package behind a store interface.

Price history, scraping, worker scheduling, frontend implementation, market rates, alerts, and notifications remain out of scope.
