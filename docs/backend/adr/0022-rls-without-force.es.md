# ADR-0022: Row-Level Security sin FORCE; el dueño ejecuta las funciones privilegiadas

[English](0022-rls-without-force.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-28

## Contexto

El [ADR-0006](0006-multi-tenancy-rls.es.md) activó la RLS con `ENABLE` **y** `FORCE`. `FORCE` somete incluso al **dueño de la tabla** a las políticas. Las funciones privilegiadas `SECURITY DEFINER` ([ADR-0013](0013-user-provisioning.es.md)) se ejecutan como dueño y deben ver y escribir filas que el usuario que actúa no puede, así que con `FORCE` fallaban: crear un usuario daba *«new row violates row-level security policy»*.

No se notó porque en local y en los primeros tests el dueño era **superusuario**, y los superusuarios siempre se saltan la RLS. En bases de datos gestionadas (Neon, Cloud SQL) el dueño no es superusuario, así que habría fallado en producción.

## Decisión

- Las tablas usan **`ENABLE ROW LEVEL SECURITY` sin `FORCE`**. La migración `00004` lo quita de `users`.
- La API sigue sin conectarse nunca como dueño (usa `nexo_api`), así que la RLS se aplica a todas sus consultas exactamente igual que antes. Solo las sesiones del dueño se la saltan: las migraciones y las funciones `SECURITY DEFINER`, que es lo que necesitan.
- **Los tests de integración ejecutan las migraciones con un dueño que no es superusuario** (`nexo_owner`), como en una base de datos gestionada.
- Dos tests de regresión: el dueño de los tests no es superusuario y ninguna tabla fuerza la RLS.

## Consecuencias

**Positivas**

- Las funciones privilegiadas funcionan igual en los tests y en producción.
- No se pierde protección: el rol de la API no es el dueño.

**Negativas / contrapartidas**

- Cualquier sesión como dueño lo ve todo: sus credenciales deben quedar fuera de la API (ya lo están).

## Alternativas consideradas

- **Mantener FORCE y dar BYPASSRLS al dueño:** requiere un superusuario para concederlo, que a menudo no está disponible en bases de datos gestionadas.
- **Mantener FORCE y añadir políticas para el dueño:** más SQL y fácil de equivocarse.
