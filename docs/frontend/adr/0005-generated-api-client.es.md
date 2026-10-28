# ADR-0005: Cliente de la API generado con orval

[English](0005-generated-api-client.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

El contrato de la API es la fuente de verdad ([ADR-0003](../../backend/adr/0003-contract-first-api.es.md) del backend). Escribir a mano las llamadas y los tipos haría que se desviaran de él.

## Decisión

- **orval** genera los tipos, las funciones de petición y los **hooks de TanStack Query** a partir de `api/openapi.yaml` en `src/shared/api/generated` (nunca se editan a mano; se comprueba en el CI).
- Todas las peticiones pasan por un único **mutator** (`http-client.ts`): devuelve el cuerpo si todo va bien y, si no, lanza un **`ApiError`** con el `code`, el estado y el `traceId` del backend (`NETWORK_ERROR` y `UNEXPECTED_RESPONSE` cuando no hay cuerpo Problem).
- `ApiError` es el tipo de error de todos los hooks generados.

## Consecuencias

**Positivas**

- Los cambios de contrato rompen la compilación donde el frontend ya no encaja.
- Los errores se tratan igual en todas partes.

**Negativas / contrapartidas**

- El código generado es extenso y sigue las convenciones de orval.

## Alternativas consideradas

- **openapi-typescript + openapi-fetch:** más ligero, pero sin hooks de consulta generados.
- **Cliente escrito a mano:** se desvía del contrato.
