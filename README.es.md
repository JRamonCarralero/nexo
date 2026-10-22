# Nexo

[English](README.md) · **Español**

> *Nexo*: del latín *nexus*, **unión, vínculo**. El punto donde se conectan las personas, los proyectos y los agentes de IA que trabajan con ellos.

Nexo es un gestor de proyectos para todo el mundo: muy simple de usar sin configurar nada, y ampliamente personalizable cuando el trabajo lo requiere. Rápido, centrado en el teclado y diseñado desde el principio para trabajar codo con codo con agentes de IA.

**Estado:** 🌱 en desarrollo temprano (esqueleto del proyecto). Todavía no es usable.

## Stack

| Parte | Tecnología |
|---|---|
| Contrato de la API | OpenAPI 3 ([`api/openapi.yaml`](api/openapi.yaml)), fuente de verdad del servidor y del cliente |
| Backend | Go · Gin · oapi-codegen · PostgreSQL (pgx, sqlc, goose) |
| Frontend | React · Vite · TypeScript · TanStack Router y Query · Tailwind · i18next · orval |
| Infraestructura | Docker Compose en local · GCP (Cloud Run) previsto |

## Estructura

```
api/          Contrato OpenAPI
backend/      API en Go (monolito modular)
  cmd/api/              arranque del servidor
  internal/app/         composición de dependencias
  internal/httpserver/  router, middlewares y código generado
  internal/modules/     módulos de negocio
  internal/platform/    infraestructura: errores, logs, config, base de datos…
  db/migrations/        migraciones SQL
frontend/     Aplicación web en React
  src/app/              arranque, router y rutas
  src/features/         funcionalidades
  src/shared/           cliente de la API, adaptadores, UI común
```

## Puesta en marcha

Requisitos: Go 1.26+, Node 22.12+, pnpm 10, Docker y `make`.

```bash
make setup      # dependencias, herramientas y git hooks; crea backend/.env
make up         # Postgres en Docker (puerto 5433, para convivir con un Postgres local en 5432)
make db-migrate # aplica las migraciones
make dev-api    # API en http://localhost:8080
make dev-web    # web en http://localhost:5173 (en otra terminal)
```

`make help` lista todos los comandos.

## Documentación

- [Integración continua (CI)](docs/ci.es.md)

## Contribuir

Lee [CONTRIBUTING.es.md](CONTRIBUTING.es.md). Las contribuciones requieren aceptar el acuerdo de contribución (CLA).

## Licencia

[GNU Affero General Public License v3.0](LICENSE).
