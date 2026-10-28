# ADR-0009: Estrategia de tests

[English](0009-testing-strategy.md) · **Español**

- **Estado:** Aceptado (implementado en parte)
- **Fecha:** 2026-10-22

## Contexto

Los tests deben comprobar el comportamiento que ve el usuario, ejecutarse rápido en el CI y no depender de un backend en marcha.

## Decisión

- **Vitest** + **Testing Library** (jsdom) para unidades y componentes, simulando `fetch`. *Implementado.*
- **MSW** con mocks generados por orval para tests de componentes más completos, y **Playwright** + **axe** para los flujos de punta a punta y la accesibilidad en el CI cuando exista el primer flujo real (iniciar sesión + crear una tarea). *Previsto.*

## Consecuencias

**Positivas**

- Respuesta rápida; tests escritos desde el punto de vista del usuario.

**Negativas / contrapartidas**

- jsdom no es un navegador real: algunos comportamientos solo se ven con Playwright.

## Alternativas consideradas

- **Jest:** más lento y necesita configuración extra con Vite.
- **Cypress:** bueno, pero Playwright es más rápido y cubre más navegadores.
