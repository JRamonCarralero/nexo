# ADR-0026: Datos de ejemplo para desarrollo local, creados a través de los servicios

[English](0026-dev-seed-data.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-11-14

## Contexto

Probar la app, enseñarla y ejecutar los tests end-to-end necesitan una base de datos con datos creíbles: personas que pueden iniciar sesión, espacios con miembros de distintos roles y una invitación pendiente. Escribirlos como SQL duplicaría reglas de negocio (ids, membresías, hashes de tokens, Row-Level Security) y se desincronizaría al evolucionar el esquema. Además, el seed nunca debe tocar producción ([ADR-0014](0014-superadmin-bootstrap.es.md)).

## Decisión

- `make db-seed` ejecuta `cmd/seed`, que construye la misma aplicación que la API y llama a `devseed.Run` (paquete `internal/devseed`). `make db-reset` lo borra todo, migra y carga el seed.
- El seed pasa por los **servicios reales**: las personas se convierten en usuarios como en un primer login (`Resolve`), los espacios los crea su administrador, los miembros entran **aceptando invitaciones por enlace** (que después se revocan) y luego reciben su rol, y se deja una invitación por email pendiente. Así los datos cumplen las mismas reglas y la RLS que cualquier cosa creada desde la app.
- Las cuentas se crean en el **emulador de Firebase Auth** (`firebase.EmulatorAccounts`) con el email verificado y una contraseña conocida, para poder iniciar sesión. Ese adaptador se niega a funcionar sin host de emulador.
- `cmd/seed` **se niega a ejecutarse** salvo con `APP_ENV=development` y `FIREBASE_AUTH_EMULATOR_HOST` definido. Los emails de invitación se descartan (`email.Discard`).
- Es **idempotente**: si un espacio ya existe (por nombre, para su administrador) se omite.
- `devseed` es código de ensamblaje: ni los módulos ni la plataforma pueden importarlo (lo impone `arch_test.go`); `app.Seed` le entrega los servicios.

## Consecuencias

**Positivas**
- Una única fuente de verdad para las reglas; el seed falla de forma visible si una regla cambia.
- Los mismos datos sirven para pruebas manuales, demos y la suite e2e.
- Seguro por construcción: solo desarrollo, solo emulador.

**Negativas / contrapartidas**
- Es más lento que el SQL directo (un puñado de transacciones).
- Hay que actualizar el seed cuando una funcionalidad cambie cómo entra la gente o cómo se crean los espacios.

## Alternativas consideradas

- **Fixtures SQL cargados con psql:** rápidos, pero duplican invariantes y se saltan la RLS y el flujo de tokens.
- **Llamar a la API HTTP desde un script:** exige un servidor en marcha e ID tokens reales; los servicios dan las mismas garantías con menos piezas.
