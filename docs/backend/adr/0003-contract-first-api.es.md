# ADR-0003: API definida primero en el contrato (OpenAPI)

[English](0003-contract-first-api.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

El backend (Go) y el frontend (TypeScript) no pueden compartir tipos directamente. Además, la API será pública más adelante (integraciones, servidor MCP, CLI), así que debe ser estable y estar documentada.

## Decisión

`api/openapi.yaml` es la **fuente de verdad**. A partir de él:

- **oapi-codegen** genera el servidor Go: modelos, rutas de Gin y una interfaz de **servidor estricto** (los handlers reciben peticiones tipadas y devuelven respuestas tipadas o un error).
- **orval** genera el cliente TypeScript y los hooks de TanStack Query.
- El contrato va embebido en el binario; el router lo lee para saber qué rutas son públicas ([ADR-0012](0012-secure-by-default-routes.es.md)).

El CI falla si el código generado no está al día, pasa el lint del contrato (Redocly) y detecta cambios incompatibles frente a `main` (oasdiff).

## Consecuencias

**Positivas**

- Una definición y dos lados tipados: un cambio de contrato que rompe el frontend no compila.
- Los handlers solo hacen `return err`; los detalles de transporte están en un solo sitio.
- La API pública está documentada desde el primer día y los cambios incompatibles se detectan en el CI.

**Negativas / contrapartidas**

- Cada endpoint empieza en YAML, que es más lento que escribir el handler directamente.
- El código generado tiene sus particularidades (p. ej. tipos de handler con el mismo nombre necesitan alias para incrustarse juntos).

## Alternativas consideradas

- **Primero el código** (generar OpenAPI desde anotaciones en Go): más rápido de escribir, pero el contrato pasa a ser un subproducto y se desvía.
- **gRPC / Connect:** tipado fuerte, pero encaja peor con una API REST pública y con clientes en el navegador.
