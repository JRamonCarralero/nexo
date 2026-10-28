# Acceso a datos

[English](data-access.md) · **Español**

Cómo lee y escribe datos el backend: el `Store` genérico, el alcance de la petición, la seguridad por filas (Row-Level Security) y cómo añadir una tabla nueva.

## Paquetes

| Paquete | Contenido | Quién lo usa |
|---|---|---|
| [`platform/database`](../../backend/internal/platform/database/) | Contratos independientes de la tecnología: `Store[T]`, `Entity`, `Model`, `TenantModel`, `Query`, `Filter`, `Page`, `Transactor` | Servicios y repositorios |
| [`platform/database/postgres`](../../backend/internal/platform/database/postgres/) | Implementación de PostgreSQL: conexión, `TxManager`, `Store[T]`, traducción de errores | `internal/app` (montaje) y los ficheros `store_postgres.go` de los módulos |
| [`platform/database/postgres/pgtest`](../../backend/internal/platform/database/postgres/pgtest/) | PostgreSQL desechable para los tests de integración | Tests |
| [`platform/scope`](../../backend/internal/platform/scope/) | El usuario que actúa y el espacio de trabajo, en el `context` | Middlewares (lo fijan), capa de base de datos (lo lee) |

Cambiar de motor de base de datos supone escribir una implementación nueva de los contratos y reescribir las consultas a medida. Servicios y handlers no cambian.

## Entidades

Una entidad es un struct con etiquetas `db` que incluye `database.Model` (global) o `database.TenantModel` (pertenece a un espacio de trabajo):

```go
type Task struct {
	database.TenantModel            // id, workspace_id, created_at, updated_at
	Title     string     `db:"title"`
	CreatedBy *uuid.UUID `db:"created_by"` // opcional: se rellena con el usuario que actúa
	DeletedAt *time.Time `db:"deleted_at"` // opcional: activa el borrado lógico
}

func (Task) TableName() string    { return "tasks" }
func (Task) ResourceName() string { return "task" } // → TASK_NOT_FOUND
```

El Store gestiona estas columnas; quien lo llama nunca las escribe:

| Columna | Comportamiento |
|---|---|
| `id` | UUIDv7 generado al crear si está vacío (ordenado por tiempo: bueno para índices y cursores) |
| `workspace_id` | Sale del alcance de la petición. Toda lectura y escritura filtra por él |
| `created_at`, `updated_at` | Los pone la base de datos (`now()`); `updated_at` cambia en cada actualización |
| `created_by` | El usuario que actúa, del alcance; `NULL` para acciones del sistema (bootstrap, seed) |
| `deleted_at` | `Delete` marca la fila; las lecturas ignoran las filas marcadas |

## El Store

```go
store := postgres.NewStore[Task](txManager) // en internal/app

task := Task{Title: "Escribir la documentación"}
err := store.Create(ctx, &task)                       // task ya tiene id, fechas…
task, err = store.GetByID(ctx, id)                    // TASK_NOT_FOUND si no existe
task, err = store.Update(ctx, id, database.Changes{"title": "Nuevo título"})
err = store.Delete(ctx, id)
page, err := store.List(ctx, database.Query{
	Filters: []database.Filter{database.Eq("status_id", statusID)},
	OrderBy: "rank", Limit: 50, After: cursor,
})
n, err := store.Count(ctx, database.In("id", ids))
```

Los servicios dependen de `database.Store[T]` (la interfaz), nunca de `postgres.Store[T]`.

**Lo que no encaje limpiamente** (joins, JSONB, búsqueda de texto, actualizaciones masivas, reordenar…) es una consulta a medida con sqlc en el `store_postgres.go` del módulo. El Store no debe convertirse en un ORM casero.

### Paginación

- **Cursor (por defecto):** pasa `Page.NextCursor` como `Query.After` para obtener la página siguiente. Es estable aunque se creen o muevan filas, y su coste es constante. `OrderBy` debe ser una columna que no admita `NULL`; los empates se resuelven por `id`. Un cursor solo sirve con la ordenación que lo generó.
- **Offset:** `Query.Offset` para listados pequeños de administración. No se puede combinar con `After`.
- **Totales:** `Count` es una llamada aparte: pídelo solo cuando la interfaz muestre un total.

## Alcance de la petición y transacciones

Los middlewares guardan en el `context` el usuario que actúa y el espacio de trabajo (`scope.WithUser`, `scope.WithWorkspace`). El `TxManager` los publica en PostgreSQL al empezar cada transacción:

```sql
SELECT set_config('app.user_id', '<uuid>', true), set_config('app.workspace_id', '<uuid>', true)
```

- **Toda operación del Store se ejecuta en una transacción**, así que el alcance siempre llega a la base de datos.
- `TxManager.WithinTx(ctx, fn)` agrupa varias operaciones (incluso de módulos distintos) en una sola transacción. Las llamadas anidadas se unen a la exterior.
- Las consultas a medida deben ejecutarse dentro de `WithinTx` y usar `TxManager.Conn(ctx)`.

## Row-Level Security

El aislamiento entre espacios de trabajo no depende solo de que el código se acuerde de filtrar: lo garantiza PostgreSQL.

- La API se conecta como **`nexo_api`**, miembro de **`nexo_app`**. **No es dueño** de las tablas (los dueños se saltan las políticas). Las migraciones se ejecutan con el dueño.
- `nexo_user_id()` y `nexo_workspace_id()` devuelven el alcance de la transacción actual (`NULL` si no está fijado).
- Cada tabla de un espacio de trabajo tiene una política que compara `workspace_id` con `nexo_workspace_id()`.
- La RLS está **activada, no forzada**: el dueño (migraciones y funciones `SECURITY DEFINER`) debe saltársela ([ADR-0022](adr/0022-rls-without-force.es.md)). Un test falla si alguna tabla fuerza la RLS.

La base de datos **falla de forma segura**: sin alcance, o fuera de una transacción, las tablas de espacios de trabajo no devuelven filas, y escribir una fila en otro espacio se rechaza con `FORBIDDEN`.

### Añadir una tabla de espacio de trabajo

```sql
-- +goose Up
CREATE TABLE tasks (
    id           uuid PRIMARY KEY,
    workspace_id uuid NOT NULL REFERENCES workspaces (id),
    title        text NOT NULL,
    created_by   uuid,
    created_at   timestamptz NOT NULL,
    updated_at   timestamptz NOT NULL,
    deleted_at   timestamptz
);
CREATE INDEX tasks_workspace_idx ON tasks (workspace_id) WHERE deleted_at IS NULL;

ALTER TABLE tasks ENABLE ROW LEVEL SECURITY; -- Sin FORCE: el dueño debe seguir saltándose la RLS (ADR-0022).
CREATE POLICY tenant_isolation ON tasks
    USING (workspace_id = nexo_workspace_id())
    WITH CHECK (workspace_id = nexo_workspace_id());

-- +goose Down
DROP TABLE tasks;
```

`nexo_app` recibe automáticamente `SELECT, INSERT, UPDATE, DELETE` sobre las tablas nuevas (privilegios por defecto de la migración `00002`).

Después, añade los tests de aislamiento (obligatorios para toda entidad de espacio de trabajo): otro espacio no puede leer, modificar ni borrar las filas.

## Roles por entorno

| Entorno | Cómo se crean los roles |
|---|---|
| Local | [`db/local/roles.sql`](../../backend/db/local/roles.sql): al crear el volumen de Docker y en cada `make up` |
| Tests | `pgtest` ejecuta el mismo fichero en el contenedor de pruebas |
| Producción | Se crean con la infraestructura (nunca con la contraseña de local) |

## Tests

```go
// TestMain, una vez por paquete (build tag: integration)
testDB, err = pgtest.Start(ctx)

// testDB.Pool      → conecta como nexo_api: aplica RLS, como en producción
// testDB.OwnerPool → dueño de las tablas (nexo_owner, NO es superusuario, como en una
//                    base de datos gestionada): preparar datos o inspeccionar filas saltándose RLS
```

Mira [`store_integration_test.go`](../../backend/internal/platform/database/postgres/store_integration_test.go) para ver ejemplos, incluidos tests de RLS que ejecutan consultas sin ningún `WHERE`.
