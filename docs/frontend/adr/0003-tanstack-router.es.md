# ADR-0003: TanStack Router con rutas basadas en ficheros

[English](0003-tanstack-router.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Uno de los principios del producto es «todo es enlazable»: cada vista y cada filtro tienen URL. Los filtros viven en los parámetros de búsqueda de la URL, así que deben estar tipados y validados.

## Decisión

**TanStack Router** con **rutas basadas en ficheros** en `src/app/routes`. El plugin de Vite (y `tsr generate` en el CI) construye `routeTree.gen.ts`; las rutas se dividen en trozos de código automáticamente. El contexto del router lleva el `QueryClient`.

## Consecuencias

**Positivas**

- Parámetros de ruta y de búsqueda totalmente tipados: un enlace roto es un error de compilación.
- Carga de datos integrada que funciona con TanStack Query.

**Negativas / contrapartidas**

- Menos conocido que React Router y con menos ejemplos.
- Un fichero generado más que mantener al día (se comprueba en el CI).

## Alternativas consideradas

- **React Router:** el estándar y el que usa el CMS del autor, pero los parámetros de búsqueda no están tipados.
