# ADR-0006: Multi-tenancy con esquema compartido y Row-Level Security

[English](0006-multi-tenancy-rls.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-23

## Contexto

Que los datos de un espacio de trabajo se filtren a otro es el peor fallo posible en un producto multi-tenant. Filtrar por `workspace_id` en el código es necesario pero frágil: basta con olvidar un `WHERE`.

## Decisión

- Una base de datos y un esquema; cada tabla de espacio de trabajo tiene `workspace_id`.
- **Row-Level Security** en cada tabla de espacio de trabajo (`ENABLE` + `FORCE`), con políticas que comparan `workspace_id` con `nexo_workspace_id()`.
- La API se conecta como **`nexo_api`** (miembro de **`nexo_app`**), que **no es dueño de las tablas** (los dueños se saltan la RLS). Las migraciones se ejecutan con el dueño.
- `TxManager` publica el alcance de la petición al empezar cada transacción (`set_config('app.user_id' | 'app.workspace_id', …, true)`), y toda operación del Store se ejecuta en una transacción.
- Los permisos por columna y las funciones `SECURITY DEFINER` cubren lo que la RLS no puede (ver [ADR-0013](0013-user-provisioning.es.md)).

## Consecuencias

**Positivas**

- Defensa en profundidad: el aislamiento se mantiene aunque el código olvide un filtro.
- La base de datos falla de forma segura: sin alcance, o sin transacción, no hay filas de ningún espacio.
- Barato de operar: un esquema y migraciones normales.

**Negativas / contrapartidas**

- Cada consulta debe ejecutarse dentro de una transacción con el alcance aplicado (un viaje de ida y vuelta extra por transacción).
- Dos roles de base de datos por entorno, y los tests necesitan una conexión de dueño y otra del rol de la API.
- Las políticas de RLS son SQL que hay que revisar y probar como el código (los tests de aislamiento son obligatorios para cada entidad).

## Alternativas consideradas

- **Base de datos o esquema por espacio:** aislamiento fuerte, pero las migraciones y las conexiones se multiplican con cada espacio.
- **Solo filtrar en la aplicación:** lo más simple, pero un solo fallo filtra datos.
