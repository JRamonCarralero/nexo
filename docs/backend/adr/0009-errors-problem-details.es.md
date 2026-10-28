# ADR-0009: Errores de aplicación y Problem Details (RFC 9457)

[English](0009-errors-problem-details.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Los errores deben tener sentido para tres públicos: el frontend (que traduce los mensajes a español e inglés), quien lee los logs y los clientes de la futura API pública. El CMS del autor ya usaba un tipo `AppError`.

## Decisión

- `platform/apperror` define el tipo de error: **tipo** (corresponde a un único estado HTTP), **código** estable y legible por máquina (`TASK_NOT_FOUND`), mensaje para desarrolladores, **detalles** por campo, **ubicación** en el código y la **causa** envuelta (`Unwrap`, `errors.Is/As`).
- Constructores con nombre (`NotFound`, `Validation`, `Conflict`, `Forbidden`, `Unauthenticated`, `Unavailable`, `Internal`).
- Los handlers solo hacen `return err`. Un middleware lo convierte a **Problem Details (RFC 9457)** con el código y un `traceId`.
- Las respuestas **5xx** nunca muestran mensajes internos; la causa queda en los logs.
- `platform/database/postgres` traduce los errores de PostgreSQL (sin filas, único, clave foránea, rechazo de RLS) a errores de aplicación.

## Consecuencias

**Positivas**

- El frontend traduce a partir del código, así que los mensajes están localizados y no dependen del texto del backend.
- Un formato de error estándar para la API pública.
- Los módulos nunca tratan con estados HTTP ni con códigos de PostgreSQL.

**Negativas / contrapartidas**

- Los códigos forman parte del contrato de la API: renombrar uno es un cambio incompatible.
- Cada código nuevo necesita su traducción en el frontend.

## Alternativas consideradas

- **Errores normales de Go + JSON improvisado en cada handler:** incoherente y filtra detalles internos.
- **Solo códigos de estado al estilo gRPC:** demasiado generales para los mensajes de la interfaz.
