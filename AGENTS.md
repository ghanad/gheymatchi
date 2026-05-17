# GheymatChi Agent Guide

## Product Context

GheymatChi is a web application for tracking prices over time. Users will add products from e-commerce websites, attach one or more source URLs, and view historical price changes. The application will also track market rates such as USD and gold so product prices can eventually be compared or converted across units.

The long-term product goal includes alerting. Users should eventually be able to define thresholds, such as a target product price or market-rate condition, and receive notifications by email or SMS when those thresholds are reached.

## Initial Stack

- Backend: Go.
- Frontend: Next.js with TypeScript.
- Local MVP database: SQLite.
- Future production database target: PostgreSQL.
- Development model: local-first development on a MacBook with simple native commands before adding operational complexity.
- Background work: a separate Go worker process for scheduled price checks.

SQLite is temporary for the local MVP. Design schemas, data access, and migrations so the project can move to PostgreSQL later without rewriting the application.

## Architecture Principles

- Build a modular monolith first.
- Keep API and worker binaries separate.
- Keep business logic outside HTTP handlers.
- Keep database access behind store or repository interfaces.
- Keep modules clean and explicit, but avoid premature service boundaries.
- Avoid microservices for the MVP.
- Prefer simple, boring technology over clever abstractions.
- Keep backend and frontend clearly separated.

The backend should be organized around domain modules such as products, sources, prices, market rates, alerts, notifications, scheduling, configuration, and database access. HTTP handlers should coordinate request parsing and response writing, while application and domain code should own business behavior.

## Safety And Scraping Rules

- Do not bypass CAPTCHAs, logins, bot protections, payment walls, or other access controls.
- Respect website terms, robots policies, and reasonable request rates.
- Prefer official APIs where available.
- Do not attempt to evade detection or website protections.
- Add per-source rate limiting before broad or repeated fetching behavior is introduced.
- Use clear user-agent and request behavior when source fetching is implemented.

## Coding Rules

- Keep changes small and focused.
- Complete the current phase only; do not implement future features early.
- Add tests for business logic where useful.
- Do not hard-code secrets, API keys, credentials, or tokens.
- Use clear errors and logging.
- Do not store prices as floating point values.
- Use integers for IRR amounts and decimal-safe representations for currencies, rates, and gold units.
- Keep database code isolated behind repositories or stores.
- Prefer code that can work with SQLite now and PostgreSQL later.
- Commit or clearly summarize changes after each phase.

## MVP Phases

1. Project bootstrap.
2. Backend skeleton.
3. SQLite schema.
4. Product CRUD.
5. Source URL tracking.
6. Price history.
7. Worker scheduler.
8. Frontend dashboard.
9. Market rates.
10. Alert rules.
11. Notifications.

Future agents should read this file before making changes. When in doubt, choose the smallest useful implementation that advances the current phase while preserving the modular monolith architecture.
