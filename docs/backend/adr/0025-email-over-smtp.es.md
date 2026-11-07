# ADR-0025: Email por SMTP detrás de una interfaz Mailer

[English](0025-email-over-smtp.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-11-07

## Contexto

Las invitaciones son la primera función que envía emails; después vendrán las notificaciones y los resúmenes. En desarrollo no se deben enviar emails reales, en producción hace falta un proveedor fiable (en la decisión de alojamiento se eligió la capa gratuita de Resend) y el proveedor podría cambiar.

## Decisión

- Los módulos dependen de la interfaz **`email.Mailer`** (`platform/email`). La única implementación, `platform/email/smtp`, habla **SMTP** con la librería estándar.
- Un único adaptador sirve para todos los entornos:
  - **En local y en los tests: Mailpit** (Docker), que captura todos los mensajes y los muestra en <http://localhost:8025>. Los tests de integración envían a través de un contenedor de Mailpit y leen el mensaje con su API.
  - **En producción: el relé SMTP de Resend** (`smtp.resend.com`, puerto 465, TLS implícito y la clave de API como contraseña).
- Modos de TLS `none`, `starttls` y `tls`. La API se niega a arrancar en producción con `SMTP_TLS=none` o con un `APP_BASE_URL` que no sea HTTPS (los enlaces llevan tokens).
- Los mensajes son MIME con una parte de texto y una parte HTML opcional (`multipart/alternative`), en UTF-8, quoted-printable y con los asuntos en codificación Q.

## Consecuencias

**Positivas**

- Ningún SDK de proveedor en el código; cambiar de proveedor es un cambio de configuración, o un adaptador nuevo.
- Lo que se prueba en local es lo que se ejecuta en producción (mismo adaptador, mismo protocolo).

**Negativas / contrapartidas**

- SMTP no da el estado de la entrega (rebotes, aperturas) como la API HTTP de un proveedor. Si hiciera falta, se puede añadir un adaptador HTTP detrás de la misma interfaz.

## Alternativas consideradas

- **SDK HTTP de Resend:** más funciones, pero ata el código a un proveedor y necesita una simulación para el desarrollo local.
- **Escribir los emails en el log en lugar de enviarlos en local:** más simple, pero el HTML y la codificación quedarían sin probar.
