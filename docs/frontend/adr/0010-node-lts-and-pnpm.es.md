# ADR-0010: Node LTS y pnpm fijado

[English](0010-node-lts-and-pnpm.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Algunas dependencias (jsdom 30) exigen versiones recientes de Node, y los entornos local y de CI deben coincidir.

## Decisión

- **Node 24 LTS**, fijado en `.nvmrc`; el CI lo lee (`node-version-file`). No se usan versiones mayores impares (sin soporte a largo plazo); los saltos de versión mayor se hacen a mano junto con `@types/node` y la imagen del emulador.
- **pnpm**, fijado mediante `packageManager` en `package.json`; en el CI se instala con `--frozen-lockfile`.

## Consecuencias

**Positivas**

- Compilaciones reproducibles; las mismas versiones en todas partes.

**Negativas / contrapartidas**

- Las versiones mayores de Node necesitan una actualización deliberada.

## Alternativas consideradas

- **npm / Yarn:** pnpm es más rápido y más estricto con las dependencias.
