# ADR-0023: Espacios de trabajo, membresías y roles

[English](0023-workspaces-and-roles.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-28

## Contexto

Todo en Nexo ocurre dentro de un espacio de trabajo. Las personas pertenecen a varios espacios con roles distintos, y el aislamiento entre ellos no debe depender solo del código de la aplicación.

## Decisión

- Un único módulo, **`workspace`**, es dueño de `workspaces` y de `memberships` (una membresía no tiene sentido fuera de su espacio).
- Tres roles: **admin** (gestiona el espacio, sus miembros e invitaciones), **member** (trabaja en él) e **invitado** (`guest`, colaborador externo; sus restricciones llegarán con los proyectos).
- **Crear un espacio** pasa por `nexo_create_workspace`, que inserta el espacio y su primer admin de forma atómica. Los miembros entran mediante invitaciones (siguiente PR); el rol de la API no puede insertar membresías.
- **Visibilidad (RLS):** cada usuario ve los espacios a los que pertenece, sus propias membresías en todos ellos y todas las membresías del espacio actual; quienes comparten un espacio ven el perfil de los demás.
- **Permisos (RLS + servicio):** solo los admins renombran el espacio y cambian roles; los admins pueden quitar a cualquiera; cualquiera puede salir. El servicio lo comprueba primero para dar errores claros (`WORKSPACE_ADMIN_REQUIRED`); la RLS aplica las mismas reglas si el servicio se saltara. Los permisos por columna solo dejan cambiar `name`, `role` y las fechas.
- **Al menos un admin:** un trigger rechaza degradar o quitar al último admin (`LAST_ADMIN`), bloqueando la fila del espacio para evitar carreras.
- El middleware de espacio de trabajo responde **404** a quien no es miembro, para no revelar que el espacio existe.
- Las funciones auxiliares que usan las políticas (`nexo_workspace_role`, `nexo_is_member`, `nexo_shares_workspace`) son `SECURITY DEFINER`, para que las políticas de `memberships` puedan leer `memberships` sin recursión.

## Consecuencias

**Positivas**

- El aislamiento y las reglas de administración se mantienen aunque haya un fallo en el servicio.
- Un espacio nunca puede quedarse sin admin.

**Negativas / contrapartidas**

- Las reglas de permisos existen dos veces (servicio y políticas) y hay que mantenerlas sincronizadas; los tests de integración cubren ambas.
- Las restricciones de los invitados se aplazan hasta que existan los proyectos.

## Alternativas consideradas

- **Módulo `membership` aparte:** las dos tablas cambian juntas, así que dos módulos se comunicarían por puertos sin beneficio.
- **Roles como tabla de permisos:** más flexible, pero prematuro con tres roles.
