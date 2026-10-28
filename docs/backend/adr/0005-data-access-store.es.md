# ADR-0005: Acceso a datos independiente de la tecnología: Store genérico y consultas a medida

[English](0005-data-access-store.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

En su trabajo, el autor usa una capa CRUD genérica para que cambiar de base de datos toque lo mínimo. Muchas entidades de Nexo son CRUD puro, pero muchas pantallas necesitan consultas que no encajan en un CRUD (reordenar, filtros combinados, JSONB, búsqueda). Además, todos los repositorios deben aplicar las mismas reglas transversales: filtro por espacio de trabajo, borrado lógico, fechas, transacciones y traducción de errores.

## Decisión

- `platform/database` contiene **contratos independientes de la tecnología**: `Store[T]`, `Entity`, `Model`/`TenantModel`, `Query`, `Filter`, `Page`, `Transactor`. Los servicios solo dependen de ellos.
- `platform/database/postgres` contiene todo lo de PostgreSQL/pgx: conexión, `TxManager`, traducción de errores y el **`Store[T]`** genérico (con genéricos de Go, no con `table string`).
- El Store **centraliza las reglas transversales**: ids UUIDv7, fechas, `created_by`, `workspace_id` desde el alcance de la petición, borrado lógico, una transacción por operación y traducción a `apperror`.
- Lo que no encaje limpiamente es una **consulta a medida** en el `store_postgres.go` del módulo (está previsto usar sqlc para ellas). El Store no debe convertirse en un ORM casero.

Guía detallada: [data-access.es.md](../data-access.es.md).

## Consecuencias

**Positivas**

- Poco código repetido en las entidades CRUD, y los repositorios no pueden olvidar el filtro por espacio de trabajo.
- Cambiar de motor supone una implementación nueva del Store y reescribir las consultas a medida; servicios y handlers no cambian.
- Toda la potencia de PostgreSQL donde importa (JSONB, RLS, búsqueda).

**Negativas / contrapartidas**

- Las consultas a medida son específicas del motor y habría que reescribirlas si cambia.
- El mapeo por reflexión se comprueba en ejecución, no al compilar (lo mitigan los tests y que `NewStore` falle con entidades inválidas).

## Alternativas consideradas

- **ORM (GORM, ent):** menos SQL que escribir, pero funciones del mínimo común y consultas ocultas.
- **Solo sqlc:** totalmente tipado, pero cada operación CRUD escrita a mano y las reglas transversales repetidas en cada consulta.
- **CRUD genérico con `table string` y `map[string]any`:** lo que se usa en el trabajo; sin seguridad de tipos.
