# ADR-0011: Autenticación del frontend tras un adaptador

[English](0011-frontend-authentication.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-11-09

## Contexto

La API confía en los ID tokens de Firebase ([auth del backend](../../backend/auth.es.md)). La app web debe iniciar sesión (email/contraseña y Google), mantenerla, enviar el token en cada petición y gestionar la verificación del email, que la API exige para aceptar invitaciones por email. Además queremos poder cambiar de proveedor de identidad sin tocar las pantallas, como hacemos en el backend.

## Decisión

- El SDK de Firebase (modular) se usa **solo** en `shared/platform/auth/firebase.ts`, tras una interfaz `AuthService` (estado, `getIdToken`, alta/inicio de sesión, Google, cierre de sesión, recuperar contraseña, verificación, `refreshUser`). ESLint prohíbe importar `firebase` en otro sitio.
- React lee el estado con `useSyncExternalStore` mediante `useAuth()`; el servicio se crea en `main.tsx` y se pasa a `App`, así los tests inyectan uno falso.
- El cliente HTTP recibe un **proveedor de token** y envía `Authorization: Bearer`. Ante `INVALID_TOKEN` reintenta **una vez** con un token renovado a la fuerza.
- Los **guards** de ruta (`requireAuth`, `redirectIfSignedIn`) se ejecutan en `beforeLoad` con el contexto del router. El parámetro `redirect` pasa por `safeRedirect` (solo rutas de la propia app) para evitar redirecciones abiertas.
- La caché de consultas se vacía cuando cambia el uid de la sesión.
- La configuración viene de `VITE_FIREBASE_*`. En desarrollo usa por defecto el emulador local; una **build de producción rechaza** la URL del emulador y exige la configuración del proyecto real.
- Tras el registro forzamos la renovación del token para que lleve el nombre, y el perfil lo crea la API en la primera petición.

## Consecuencias

**Positivas**
- Pantallas y tests no conocen Firebase; cambiar de proveedor afecta a una carpeta.
- Todo funciona en local con el emulador y sin cuenta en la nube.
- Un enlace de login manipulado no puede redirigir a otro sitio; una build con emulador no llega a producción por error.

**Negativas / contrapartidas**
- El SDK de Firebase añade peso al bundle (se carga con la app, porque todas las páginas necesitan auth).
- La interfaz cubre lo que usamos hoy; nuevos proveedores o MFA la ampliarán.

## Alternativas consideradas

- **Llamar a Firebase directamente desde los componentes:** más rápido, pero acopla pantallas y tests al SDK.
- **Cookies de sesión emitidas por nuestra API:** más trabajo de backend y sin beneficio ahora; la API ya es stateless con ID tokens.
