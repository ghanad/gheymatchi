# GheymatChi

GheymatChi is a local-first web application for tracking product prices over time. Users will eventually add products from e-commerce websites, attach source URLs, review historical price changes, and compare prices against market rates such as USD and gold.

The MVP starts with a simple modular monolith: a Go backend, a Next.js + TypeScript frontend, and SQLite for local development. PostgreSQL is the planned production database target, so database code and schema decisions should keep that migration path in mind.

## Local Development Goal

The project should run on a MacBook with simple local commands and Docker Compose. The current backend API exposes health checks, connects to a local SQLite database for readiness checks, and provides product CRUD, product source URL management, manual price history endpoints, and product alert rule management. The frontend provides the initial dashboard shell and a simple alert rule management page.

## Initial Stack

- Backend: Go
- API routing: Go standard library or a small router in a later phase
- Worker: planned separate Go process for scheduled price checks
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

Run the worker directly:

```sh
make worker
```

Run the worker with the first real external source adapter enabled:

```sh
cd backend
DB_PATH=../data/gheymatchi.db PRICE_FETCHER=digikala go run ./cmd/worker
```

Run the frontend directly:

```sh
cd frontend
npm install
BACKEND_API_BASE_URL=http://localhost:8080 npm run dev
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

Phase 18 adds the first real external price source adapter. The worker still uses the deterministic mock fetcher by default for local development and Docker Compose. Setting `PRICE_FETCHER=digikala` enables the Digikala adapter for supported product URLs such as `https://www.digikala.com/product/dkp-1234567/...`. The adapter fetches Digikala's public product JSON endpoint with a per-request timeout, a conservative per-source delay, parser-level error handling, and compact raw metadata storage instead of storing full responses.

The backend currently includes configuration loading from environment variables, structured logging, graceful shutdown, a local SQLite database at `data/gheymatchi.db`, and these endpoints:

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
- `POST /api/products/{product_id}/sources/{source_id}/price-points`
- `GET /api/products/{product_id}/price-points`
- `POST /api/products/{product_id}/alerts`
- `GET /api/products/{product_id}/alerts`
- `PATCH /api/products/{product_id}/alerts/{alert_id}`
- `DELETE /api/products/{product_id}/alerts/{alert_id}`
- `POST /api/market-rates`
- `GET /api/market-rates/latest`
- `GET /api/market-rates/history`
- `GET /api/notifications`

Product management UI, charts, manual market-rate storage, derived price display, alert rules, worker-side alert evaluation, a notification log, and one opt-in real Digikala price source adapter are implemented. Multiple source adapters, broad scraping, and real email/SMS notification sending are not implemented yet.
