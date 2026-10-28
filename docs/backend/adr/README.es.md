# Registro de decisiones de arquitectura (ADR) — Backend

[English](README.md) · **Español**

Un ADR recoge una decisión importante: el contexto, qué se decidió, por qué y qué cuesta. Una vez aceptado no se edita: si la decisión cambia, un ADR nuevo sustituye al anterior y ambos quedan en el historial.

| ADR | Título | Estado | Fecha |
|---|---|---|---|
| [0001](0001-go-for-the-backend.es.md) | Go para el backend | Aceptado | 2026-10-22 |
| [0002](0002-modular-monolith.es.md) | Monolito modular | Aceptado | 2026-10-22 |
| [0003](0003-contract-first-api.es.md) | API definida primero en el contrato (OpenAPI) | Aceptado | 2026-10-22 |
| [0004](0004-postgresql.es.md) | PostgreSQL como base de datos principal | Aceptado | 2026-10-22 |
| [0005](0005-data-access-store.es.md) | Acceso a datos independiente de la tecnología: Store genérico y consultas a medida | Aceptado | 2026-10-22 |
| [0006](0006-multi-tenancy-rls.es.md) | Multi-tenancy con esquema compartido y Row-Level Security | Aceptado | 2026-10-23 |
| [0007](0007-uuidv7-identifiers.es.md) | Identificadores UUIDv7 generados en Go | Aceptado | 2026-10-22 |
| [0008](0008-cursor-pagination.es.md) | Paginación por cursor por defecto | Aceptado | 2026-10-22 |
| [0009](0009-errors-problem-details.es.md) | Errores de aplicación y Problem Details (RFC 9457) | Aceptado | 2026-10-22 |
| [0010](0010-structured-logging.es.md) | Logs estructurados: una línea por petición | Aceptado | 2026-10-22 |
| [0011](0011-firebase-authentication.es.md) | Firebase Authentication detrás de una interfaz Authenticator | Aceptado | 2026-10-23 |
| [0012](0012-secure-by-default-routes.es.md) | Rutas seguras por defecto, derivadas del contrato | Aceptado | 2026-10-23 |
| [0013](0013-user-provisioning.es.md) | Alta de usuarios mediante funciones SECURITY DEFINER | Aceptado | 2026-10-23 |
| [0014](0014-superadmin-bootstrap.es.md) | Bootstrap del superadmin de la plataforma | Aceptado | 2026-10-23 |
| [0015](0015-commands-and-outbox.es.md) | Comandos y eventos con outbox transaccional | Aceptado (aún no implementado) | 2026-10-22 |
| [0016](0016-custom-fields-jsonb.es.md) | Campos personalizados en JSONB | Aceptado (aún no implementado) | 2026-10-22 |
| [0017](0017-workspace-in-url-path.es.md) | Espacio de trabajo en la ruta de la URL | Aceptado (aún no implementado) | 2026-10-23 |
| [0018](0018-testing-strategy.es.md) | Estrategia de tests | Aceptado | 2026-10-22 |
| [0019](0019-migrations-goose.es.md) | Migraciones SQL con goose, embebidas en el binario | Aceptado | 2026-10-22 |
| [0020](0020-local-environment.es.md) | Entorno local con Docker Compose y emuladores | Aceptado | 2026-10-23 |
| [0021](0021-realtime-signals.es.md) | Tiempo real mediante documentos de aviso en Firestore | Aceptado (aún no implementado) | 2026-10-22 |

## Cómo escribir un ADR nuevo

1. Copia la plantilla de abajo en `NNNN-titulo-corto.md` (siguiente número libre) y su gemelo en español `NNNN-titulo-corto.es.md`.
2. Empieza como **Propuesto**; pásalo a **Aceptado** al hacer merge.
3. Para cambiar una decisión, escribe un ADR nuevo y marca el anterior como **Sustituido por ADR-NNNN**.
4. Añádelo a la tabla de arriba.

## Plantilla

```markdown
# ADR-NNNN: Título

[English](NNNN-titulo-corto.md) · **Español**

- **Estado:** Propuesto | Aceptado | Aceptado (aún no implementado) | Sustituido por ADR-NNNN
- **Fecha:** AAAA-MM-DD

## Contexto
Qué problema o necesidad obliga a decidir.

## Decisión
Qué se decidió, dicho de forma directa.

## Consecuencias
**Positivas**
- …

**Negativas / contrapartidas**
- …

## Alternativas consideradas
- **Opción:** por qué no se eligió.
```
