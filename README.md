# GheymatChi

GheymatChi is a local-first web application for tracking product prices over time. Users will eventually add products from e-commerce websites, attach source URLs, review historical price changes, and compare prices against market rates such as USD and gold.

The MVP starts with a simple modular monolith: a Go backend, a Next.js + TypeScript frontend, and SQLite for local development. PostgreSQL is the planned production database target, so database code and schema decisions should keep that migration path in mind.

## Local Development Goal

The project runs on a MacBook with simple local commands and Docker Compose. The current backend API exposes health checks, connects to a local SQLite database for readiness checks, and provides product CRUD, product source URL management, manual price history endpoints, market-rate endpoints, product alert rule management, and a notification log. The frontend provides the dashboard, product detail, alert, notification, and settings screens.

## Initial Stack

- Backend: Go
- API routing: Go standard library or a small router in a later phase
- Worker: planned separate Go process for scheduled price checks
- Frontend: Next.js with TypeScript
- Local MVP database: SQLite
- Future database target: PostgreSQL
- Local orchestration: Docker Compose

## Prerequisites

- Go 1.22 or newer
- Node.js 20 or newer
- Docker Desktop, if using Docker Compose
- `make`

## Environment Files

Example environment files are committed without secrets:

```sh
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
```

The Makefile sets the local defaults directly, so copying these files is optional for the current MVP. Keep real secrets out of `.env` files.

## Local Commands

Install frontend dependencies once after a clean checkout:

```sh
make frontend-install
```

Create or update the local SQLite database:

```sh
make migrate
```

Add local demo data for testing the frontend:

```sh
make seed
```

Run the API, worker, and frontend in separate terminal tabs:

```sh
make api
make worker
make frontend
```

Run backend tests:

```sh
make test
```

The frontend will be available at `http://localhost:3000`; the API listens on `http://localhost:8080`.

## Docker Compose

Docker Compose is optional. The native `make api`, `make worker`, and `make frontend` workflow above remains the primary simple local development path.

From a clean checkout, start the local stack:

```sh
make docker-up
```

This starts the API, worker, and frontend, and applies SQLite migrations before the API and worker start. The SQLite database is stored under `data/gheymatchi.db`, which is mounted into the containers so data persists across container restarts.

To add optional demo data, run:

```sh
make docker-seed
```

Stop the Compose services:

```sh
make docker-down
```

To run the opt-in PostgreSQL stack added in Phase 24:

```sh
make docker-postgres-up
```

The default Compose stack still uses SQLite. PostgreSQL uses `DB_DRIVER=postgres` and `DB_DSN=postgres://gheymatchi:gheymatchi@postgres:5432/gheymatchi?sslmode=disable`.

## External Price Fetching

The worker uses the deterministic mock fetcher by default for local development and Docker Compose. To opt into the Digikala adapter for a supported product URL, run the worker directly with:

```sh
cd backend
DB_PATH=../data/gheymatchi.db PRICE_FETCHER=digikala go run ./cmd/worker
```

## Email Notifications

Dry-run notifications remain the default. In dry-run mode, the worker records alert notifications and marks them sent without contacting an external provider.

To enable SMTP email locally, configure the worker environment:

```sh
NOTIFICATION_PROVIDER=smtp
NOTIFICATION_EMAIL_TO=you@example.com
NOTIFICATION_MAX_ATTEMPTS=3
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=your-username
SMTP_PASSWORD=your-password-or-api-key
SMTP_FROM=alerts@example.com
```

Do not commit real SMTP credentials. In the current single-user MVP, `NOTIFICATION_EMAIL_TO` is a single local recipient. Per-user email addresses belong to the later authentication phase.

With Docker Compose, pass those variables to the `worker` service through your shell or a local uncommitted override file. Failed sends keep the notification pending until `NOTIFICATION_MAX_ATTEMPTS` is reached, then the notification is marked failed with the latest non-secret error.

## Current Phase

Phase 24 adds opt-in PostgreSQL connection support, PostgreSQL migrations, and Docker Compose services while keeping SQLite as the default local development database.

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

Product management UI, charts, manual market-rate storage, derived price display, alert rules, worker-side alert evaluation, a notification log, optional SMTP email notifications, basic authentication, opt-in PostgreSQL support, and one opt-in real Digikala price source adapter are implemented. Multiple source adapters, broad scraping, SMS notification sending, and automated SQLite-to-PostgreSQL data copy are not implemented yet.
