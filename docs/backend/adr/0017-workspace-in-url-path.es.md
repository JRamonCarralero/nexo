# ADR-0017: Espacio de trabajo en la ruta de la URL

[English](0017-workspace-in-url-path.md) · **Español**

- **Estado:** Aceptado (aún no implementado)
- **Fecha:** 2026-10-23

## Contexto

La mayoría de las operaciones ocurren dentro de un espacio de trabajo. La API necesita saber cuál y comprobar que el usuario pertenece a él antes de llegar al handler.

## Decisión

Los endpoints de un espacio de trabajo cuelgan de **`/api/v1/workspaces/{workspaceId}/…`**. Un middleware valida la membresía y pone el espacio de trabajo en el alcance de la petición, que después llega a la RLS.

## Consecuencias

**Positivas**

- Explícito y enlazable: cada recurso tiene una URL completa.
- La membresía se comprueba en un solo sitio antes de ejecutar ningún handler.

**Negativas / contrapartidas**

- Rutas más largas.

## Alternativas consideradas

- **Cabecera `X-Workspace-Id`:** rutas más cortas, pero el contexto queda oculto y es más fácil equivocarse.
