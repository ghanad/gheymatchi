# Database

GheymatChi uses SQLite for the local MVP. PostgreSQL is the future production database target, so schema and data-access decisions should avoid SQLite-specific assumptions where a simple portable design is available.

## Storage Principles

- Keep database access isolated behind repositories or stores.
- Store prices without floating point values.
- Use integer values for IRR amounts.
- Use decimal-safe representations for currencies, rates, and gold units when those features are introduced.
- Keep migrations explicit and reviewable.

## Current Phase Boundary

No schema, migrations, or database connection are implemented in Phase 3. The `backend/migrations` directory remains a placeholder for the SQLite setup phase.
