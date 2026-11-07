# ADR-0024: Invitaciones por email y por enlace

[English](0024-invitations.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-11-07

## Contexto

Solo se entra en un espacio de trabajo mediante una invitación ([ADR-0023](0023-workspaces-and-roles.es.md)). Una invitación es una credencial: quien la tiene puede entrar en el espacio. Viaja por email o por chat, puede reenviarse y su token aparece en una URL.

## Decisión

- **Dos tipos:**
  - **email**: para una persona, **de un solo uso**, caduca en **7 días**. Para aceptarla hace falta una cuenta con **el mismo email, verificado**: un enlace reenviado no le sirve a otra persona, y nadie puede registrar la dirección del invitado sin verificarla para quedarse con su plaza. Las cuentas de Google ya vienen verificadas.
  - **enlace**: se puede compartir y es **reutilizable** hasta que caduca (**30 días**) o se revoca. **Nunca puede dar el rol de admin** (lo impiden una restricción de la tabla, la función de creación y el servicio).
- **Tokens:** 256 bits aleatorios en base64url. Solo se guarda su **hash SHA-256**, así que una fuga de la base de datos no expone ninguna invitación utilizable. Los tokens por email solo llegan al buzón del destinatario; las URL de los enlaces se devuelven una única vez, al crearlos.
- **Solo los admins** crean, listan y revocan invitaciones (comprobación en el servicio y RLS). Volver a invitar a un email **revoca la invitación pendiente anterior**; invitar a alguien que ya es miembro se rechaza (`ALREADY_MEMBER`).
- **Aceptar** pasa por `nexo_accept_invitation` (`SECURITY DEFINER`): la persona aún no es miembro, así que la RLS no le dejaría crear su membresía. Es **idempotente** y nunca cambia el rol de quien ya es miembro.
- Una **vista previa** (`GET /invitations/{token}`) muestra el espacio, el rol, quién invita y el estado (pendiente, caducada, revocada, aceptada) antes de aceptar. Tanto la vista previa como la aceptación exigen haber iniciado sesión.
- Los tokens **nunca llegan a los logs**: el middleware de logs sustituye el parámetro `token` de la ruta por `[REDACTED]`. Los tokens mal formados se rechazan sin consultar la base de datos.
- Si el email no se puede enviar, la invitación se **revoca** y la API responde `EMAIL_DELIVERY_FAILED`, para que no exista una invitación que nadie ha recibido. El outbox transaccional ([ADR-0015](0015-commands-and-outbox.es.md)) lo convertirá en reintentos.
- Las invitaciones viven en el **módulo `workspace`**: son la forma de crear membresías.
- El email es **bilingüe** (español y después inglés), porque no se sabe el idioma de quien lo recibe.

## Consecuencias

**Positivas**

- Ni una fuga de la base de datos, ni un enlace reenviado, ni una cuenta sin verificar sirven para quedarse con una invitación por email.
- Las reglas se cumplen en la base de datos aunque se salte el servicio (los tests lo demuestran).

**Negativas / contrapartidas**

- Quien se registra con email y contraseña debe verificar su dirección antes de aceptar una invitación por email.
- Sin el outbox, un fallo temporal del SMTP obliga al admin a reintentar a mano.

## Alternativas consideradas

- **Guardar los tokens en claro:** búsquedas más simples, pero una fuga de la base de datos expondría todas las invitaciones pendientes.
- **Aceptar invitaciones por email sin email verificado:** menos fricción, pero quien tuviera un enlace reenviado podría registrar la dirección y entrar.
- **Permitir enlaces de admin:** cómodo, pero un enlace puede acabar en cualquier parte.
