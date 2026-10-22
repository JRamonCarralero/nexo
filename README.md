# Nexo

**English** · [Español](README.es.md)

> *Nexo*: from the Latin *nexus*, **bond, connection**. The point where people, projects and the AI agents working with them come together.

Nexo is a project manager for everyone: very simple to use without configuring anything, and highly customizable when the work calls for it. Fast, keyboard-first and designed from day one to work hand in hand with AI agents.

**Status:** 🌱 early development (project skeleton). Not usable yet.

## Stack

| Part | Technology |
|---|---|
| API contract | OpenAPI 3 ([`api/openapi.yaml`](api/openapi.yaml)), the source of truth for server and client |
| Backend | Go · Gin · oapi-codegen · PostgreSQL (pgx, sqlc, goose) |
| Frontend | React · Vite · TypeScript · TanStack Router & Query · Tailwind · i18next · orval |
| Infrastructure | Docker Compose locally · GCP (Cloud Run) planned |

## Structure

```
api/          OpenAPI contract
backend/      Go API (modular monolith)
  cmd/api/              server entry point
  internal/app/         dependency wiring
  internal/httpserver/  router, middlewares and generated code
  internal/modules/     business modules
  internal/platform/    infrastructure: errors, logging, config, database…
  db/migrations/        SQL migrations
frontend/     React web app
  src/app/              bootstrap, router and routes
  src/features/         features
  src/shared/           API client, adapters, shared UI
```

## Getting started

Requirements: Go 1.26+, Node 22.12+, pnpm 10, Docker and `make`.

```bash
make setup      # dependencies, tools and git hooks; creates backend/.env
make up         # Postgres in Docker (port 5433, to coexist with a local Postgres on 5432)
make db-migrate # applies the migrations
make dev-api    # API at http://localhost:8080
make dev-web    # web at http://localhost:5173 (in another terminal)
```

`make help` lists every command.

## Documentation

- [Continuous integration (CI)](docs/ci.md)

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md). Contributions require accepting the Contributor License Agreement (CLA).

## License

[GNU Affero General Public License v3.0](LICENSE).
