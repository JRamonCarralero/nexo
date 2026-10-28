# Registro de decisiones de arquitectura (ADR) — Frontend

[English](README.md) · **Español**

Un ADR recoge una decisión importante: el contexto, qué se decidió, por qué y qué cuesta. Una vez aceptado no se edita: si la decisión cambia, un ADR nuevo sustituye al anterior y ambos quedan en el historial.

| ADR | Título | Estado | Fecha |
|---|---|---|---|
| [0001](0001-react-vite-spa.es.md) | Aplicación de una sola página con React y Vite | Aceptado | 2026-10-22 |
| [0002](0002-typescript-strict.es.md) | TypeScript estricto, fijado en la versión 6 | Aceptado | 2026-10-22 |
| [0003](0003-tanstack-router.es.md) | TanStack Router con rutas basadas en ficheros | Aceptado | 2026-10-22 |
| [0004](0004-state-management.es.md) | Estado del servidor en TanStack Query y de la interfaz en Zustand | Aceptado | 2026-10-22 |
| [0005](0005-generated-api-client.es.md) | Cliente de la API generado con orval | Aceptado | 2026-10-22 |
| [0006](0006-feature-structure-and-adapters.es.md) | Estructura por funcionalidades y adaptadores de plataforma | Aceptado | 2026-10-22 |
| [0007](0007-styling-and-components.es.md) | Tailwind CSS y componentes shadcn/ui | Aceptado (implementado en parte) | 2026-10-22 |
| [0008](0008-internationalization.es.md) | Internacionalización con i18next y errores traducidos por código | Aceptado | 2026-10-22 |
| [0009](0009-testing-strategy.es.md) | Estrategia de tests | Aceptado (implementado en parte) | 2026-10-22 |
| [0010](0010-node-lts-and-pnpm.es.md) | Node LTS y pnpm fijado | Aceptado | 2026-10-22 |

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
