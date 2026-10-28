# ADR-0010: Logs estructurados: una línea por petición

[English](0010-structured-logging.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Los logs se leerán en Google Cloud Logging. Registrar el mismo error en cada capa genera ruido, y los errores deben poder rastrearse desde el aviso de un usuario hasta la línea del log.

## Decisión

- `log/slog` (librería estándar). JSON con los nombres de campo de Cloud Logging (`severity`, `message`) fuera de desarrollo; texto legible en desarrollo.
- Un **trace id** por petición: se toma de `traceparent` o de `X-Cloud-Trace-Context`, o se genera; se devuelve en `X-Trace-Id` y en cada Problem, y se añade a todo log escrito con el contexto de la petición.
- **Una línea de log por petición**, escrita por el middleware de logs: `INFO` normalmente, `WARNING` para 4xx sin detalles internos y `ERROR` para 5xx con la ubicación y la cadena de causas. El health check se registra en `DEBUG`.

## Consecuencias

**Positivas**

- Sin logs duplicados; cada petición es un único registro que se puede buscar.
- Un usuario puede dar el código de seguimiento que ve en la interfaz para encontrar la línea exacta del log.

**Negativas / contrapartidas**

- El código que quiera escribir logs adicionales debe usar el contexto de la petición para conservar el trace id.
- El trazado distribuido (OpenTelemetry) todavía no está montado.

## Alternativas consideradas

- **zap / zerolog:** más rápidos, pero una dependencia más; slog es suficiente.
- **Registrar los errores donde ocurren:** simple, pero duplica entradas entre capas.
