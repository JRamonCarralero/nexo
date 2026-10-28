# ADR-0006: Estructura por funcionalidades y adaptadores de plataforma

[English](0006-feature-structure-and-adapters.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

El backend está organizado por dominio, con adaptadores detrás de interfaces. Organizar el frontend por tipo de fichero (components/, hooks/, services/) no escala y deja que los SDK de terceros (Firebase) se extiendan por todas partes.

## Decisión

```
src/app/       arranque, router, providers y rutas
src/features/  una carpeta por funcionalidad, con su API pública en index.ts
src/shared/    api/ (cliente generado), platform/ (adaptadores: i18n, auth, tiempo real…), ui/
```
ESLint lo hace cumplir: es un error importar **Firebase fuera de `shared/platform`**, y también importar las **partes internas de una funcionalidad** en lugar de su `index.ts`.

## Consecuencias

**Positivas**

- Refleja los módulos del backend; las funcionalidades se mantienen cohesionadas.
- Cambiar de proveedor toca un único adaptador.

**Negativas / contrapartidas**

- Exige criterio para decidir qué es compartido y qué pertenece a una funcionalidad.

## Alternativas consideradas

- **Por tipo de fichero:** familiar, pero mezcla código sin relación según crece la aplicación.
