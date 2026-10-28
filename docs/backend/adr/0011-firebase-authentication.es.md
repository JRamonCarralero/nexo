# ADR-0011: Firebase Authentication detrás de una interfaz Authenticator

[English](0011-firebase-authentication.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-23

## Contexto

Construir una autenticación segura (contraseñas, inicio con Google, verificación de email, restablecimiento) es caro y arriesgado. El autor conoce Firebase, cuya capa gratuita cubre 50.000 usuarios activos al mes. El proveedor podría cambiar más adelante.

## Decisión

- La web inicia sesión con **Firebase Authentication** y envía el **ID token** como token bearer.
- El backend lo verifica a través de la interfaz **`Authenticator`** (`platform/auth`); el SDK de Firebase solo vive en `platform/auth/firebase`.
- En local y en los tests, el **emulador de Firebase Auth** (proyecto `demo-nexo`, en Docker) sustituye al servicio real. La API se niega a arrancar en producción con el emulador configurado.
- Nexo identifica a las personas por su **email**: los tokens sin email se rechazan.

## Consecuencias

**Positivas**

- Nexo no guarda contraseñas; los métodos de inicio de sesión vienen incluidos.
- Cambiar de proveedor supone un paquete adaptador nuevo.
- Los tests de integración se ejecutan contra el emulador real, sin mocks del SDK.

**Negativas / contrapartidas**

- Dependencia de Google para la identidad; las cuentas viven fuera de nuestra base de datos.
- La revocación de tokens no se comprueba en cada petición (necesitaría una llamada extra).

## Alternativas consideradas

- **Autenticación propia (p. ej. Better Auth):** control total, pero más superficie de seguridad que mantener.
- **Auth0 / Clerk:** buenos productos, pero de pago más allá de capas gratuitas pequeñas.
