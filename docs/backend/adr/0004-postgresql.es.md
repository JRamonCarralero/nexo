# ADR-0004: PostgreSQL como base de datos principal

[English](0004-postgresql.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Un gestor de proyectos necesita filtros combinados (un responsable U otro, prioridad, rangos de fechas, etiquetas), búsqueda de texto completo, datos relacionales (dependencias, membresías, informes) y costes predecibles. El autor conoce bien Firestore por su trabajo.

## Decisión

**PostgreSQL** es la fuente de verdad de todos los datos de negocio (17 en local y en tests; Neon o Cloud SQL al desplegar). Las extensiones `citext`, `unaccent` y `pg_trgm` se activan desde la primera migración. Firestore puede usarse como **canal de avisos** para el tiempo real ([ADR-0021](0021-realtime-signals.es.md)), nunca como almacén de datos.

## Consecuencias

**Positivas**

- Consultas ricas, joins, transacciones y búsqueda de texto completo (sin distinguir tildes).
- JSONB da flexibilidad de esquema donde hace falta (campos personalizados, [ADR-0016](0016-custom-fields-jsonb.es.md)).
- La Row-Level Security convierte el aislamiento entre espacios en una garantía de la base de datos ([ADR-0006](0006-multi-tenancy-rls.es.md)).
- Coste predecible: Firestore cobra por documento leído (abrir un tablero de 200 tareas = 200 lecturas).

**Negativas / contrapartidas**

- Sin tiempo real ni modo sin conexión de serie: hay que añadirlos por encima.
- Una instancia gestionada no es gratis en GCP (Cloud SQL); la opción gratuita es un proveedor externo (Neon).

## Alternativas consideradas

- **Firestore:** tiempo real y modo sin conexión incluidos, y ya conocido, pero con filtros OR/rangos limitados, sin búsqueda de texto y con precio por lectura.
- **MySQL:** viable, pero con peor soporte de JSON y sin Row-Level Security.
