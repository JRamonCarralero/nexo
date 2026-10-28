# ADR-0001: Go para el backend

[English](0001-go-for-the-backend.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Nexo necesita una API HTTP rápida, barata de ejecutar en una capa gratuita (Cloud Run) y mantenible durante años por un equipo pequeño. El punto fuerte profesional del autor es Go sobre GCP; su experiencia en frontend es más limitada. La otra opción seria era TypeScript en los dos lados, porque permite compartir tipos y esquemas de validación.

## Decisión

El backend se escribe en **Go** (Gin como router HTTP y pgx para PostgreSQL). Los tipos compartidos con el frontend se resuelven con una API definida primero en el contrato ([ADR-0003](0003-contract-first-api.es.md)) en lugar de con un lenguaje común.

## Consecuencias

**Positivas**

- Aprovecha la experiencia principal del autor: menos errores en la parte más crítica.
- Un único binario estático, poco consumo de memoria y arranques en frío rápidos: encaja con la capa gratuita de Cloud Run.
- Existe el SDK oficial de MCP para Go, y los motores de sincronización local-first leen de PostgreSQL, así que ninguna función futura depende del lenguaje del backend.

**Negativas / contrapartidas**

- Los tipos no se comparten de forma nativa con el frontend: hay que generarlos desde el contrato.
- Algunas reglas de validación pueden tener que expresarse dos veces (contrato y código Go).

## Alternativas consideradas

- **TypeScript con NestJS:** tipos y esquemas Zod compartidos, y NestJS ya era conocido; se descartó porque el backend es donde más pesa la experiencia del autor.
- **Backend solo con Firebase (Cloud Functions + Firestore):** el arranque más rápido, pero flojo para consultas relacionales y caro por lectura (ver [ADR-0004](0004-postgresql.es.md)).
