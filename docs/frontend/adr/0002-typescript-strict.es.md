# ADR-0002: TypeScript estricto, fijado en la versión 6

[English](0002-typescript-strict.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

El frontend consume un cliente de API generado y totalmente tipado. TypeScript 7 (la reescritura nativa en Go) acababa de publicarse, pero las herramientas de lint todavía no la soportaban.

## Decisión

TypeScript en **modo estricto** con comprobaciones extra (`noUncheckedIndexedAccess`, `noUnusedLocals`, `verbatimModuleSyntax`), el alias `@/` para `src/` y **TypeScript 6** fijado hasta que typescript-eslint soporte la versión 7.

## Consecuencias

**Positivas**

- Los cambios de contrato aparecen como errores de compilación.
- Unas herramientas estables.

**Negativas / contrapartidas**

- De momento se pierde el compilador más rápido de TypeScript 7.

## Alternativas consideradas

- **TypeScript 7 desde ya:** más rápido, pero el lint dejaría de funcionar.
