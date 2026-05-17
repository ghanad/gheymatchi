# Product Requirements

GheymatChi helps users track product prices over time. A future user will be able to add a product, attach one or more source URLs, and view historical price changes.

The MVP should stay local-first and small. SQLite is the local database for the MVP, while PostgreSQL is the planned production database target.

## Initial Scope

- Track products and their source URLs in later phases.
- Store price history in later phases.
- Add a worker process for scheduled checks in a later phase.
- Show a frontend dashboard in a later phase.

## Out Of Scope For This Phase

- Product CRUD
- Source URL fetching
- Price parsing
- Market-rate tracking
- Alert rules
- Email or SMS notifications
- Authentication
