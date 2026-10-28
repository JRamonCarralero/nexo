# ADR-0004: Estado del servidor en TanStack Query y de la interfaz en Zustand

[English](0004-state-management.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

La mayor parte del estado de un gestor de proyectos es una copia en caché de datos del servidor que otras personas cambian a la vez. Mezclarlo con el estado de la interfaz en un único store global lleva a datos desactualizados y a sincronizar a mano.

## Decisión

- **Datos del servidor** (tareas, proyectos…): **TanStack Query**, que guarda en caché, reintenta y gestiona las actualizaciones optimistas. Los avisos en tiempo real invalidan las consultas. Los errores 4xx no se reintentan.
- **Estado de la interfaz** (paneles abiertos, selecciones): **Zustand**, solo cuando hay que compartirlo entre componentes; si no, estado local del componente.
- Los datos del servidor nunca se copian a Zustand.

## Consecuencias

**Positivas**

- Una única fuente de verdad para cada tipo de estado; sin sincronización manual.
- Los hooks generados (orval) se conectan directamente con TanStack Query.

**Negativas / contrapartidas**

- Dos librerías que aprender.

## Alternativas consideradas

- **Redux Toolkit / RTK Query:** potente, pero con más código repetitivo.
- **Todo en un store global:** datos desactualizados y caché gestionada a mano.
