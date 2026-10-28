# ADR-0013: Alta de usuarios mediante funciones SECURITY DEFINER

[English](0013-user-provisioning.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-23

## Contexto

Un usuario debe crearse en su primera petición autenticada, antes de que la API conozca su id de Nexo, así que la RLS (que funciona por usuario) no puede permitirlo. Además, la RLS filtra filas, no columnas: una política que deje a cada usuario actualizar su fila le permitiría poner `is_superadmin`.

## Decisión

- Los usuarios se crean **automáticamente en el middleware de autenticación** en la primera petición autenticada (justo a tiempo).
- El rol de la API **no puede insertar usuarios**: los únicos puntos de entrada son dos funciones **`SECURITY DEFINER`** propiedad del dueño de la tabla: `nexo_resolve_user` y `nexo_bootstrap_superadmin`.
- **Permisos por columna**: el rol de la API solo puede actualizar `name`, `avatar_url` y `updated_at`.
- Una cuenta preparada de antemano solo se vincula a una identidad cuando el **email está verificado**; si no, el login recibe `EMAIL_IN_USE`.

## Consecuencias

**Positivas**

- No existe el estado intermedio de «con sesión en Firebase pero desconocido para Nexo».
- Las operaciones privilegiadas son acotadas, tienen nombre y se pueden probar.
- Nadie puede hacerse superadmin a sí mismo, ni siquiera con un fallo en la API.

**Negativas / contrapartidas**

- Parte de la lógica vive en PL/pgSQL, menos conocido y que necesita tests de integración.
- Una consulta extra por petición autenticada (se podrá cachear más adelante).

## Alternativas consideradas

- **`POST /me` explícito tras el login:** un paso más que puede fallar.
- **Un segundo rol de base de datos con BYPASSRLS para tareas del sistema:** más poder en manos del mismo proceso.
