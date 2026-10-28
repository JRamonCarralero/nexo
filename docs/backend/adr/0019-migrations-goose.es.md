# ADR-0019: Migraciones SQL con goose, embebidas en el binario

[English](0019-migrations-goose.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

El esquema incluye cosas que solo el SQL expresa bien: políticas de RLS, roles, permisos y funciones. Las migraciones deben ejecutarse igual en local, en los tests y en el despliegue.

## Decisión

Migraciones en SQL puro en `backend/db/migrations`, ejecutadas con **goose** (`make db-migrate`, `db-rollback`, `db-new`, `db-reset`). Van embebidas con `go:embed`, así que los tests las aplican sin depender del directorio de trabajo. Las migraciones se ejecutan con el dueño de las tablas; `db/local/roles.sql` crea los roles de login de local.

## Consecuencias

**Positivas**

- SQL completo, diferencias legibles y subida y bajada en el mismo fichero.

**Negativas / contrapartidas**

- Sin diferencias automáticas desde los modelos: las migraciones se escriben a mano.

## Alternativas consideradas

- **Atlas / migraciones automáticas de un ORM:** cómodas, pero incómodas para políticas, roles y funciones.
