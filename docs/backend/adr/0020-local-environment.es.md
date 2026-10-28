# ADR-0020: Entorno local con Docker Compose y emuladores

[English](0020-local-environment.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-23

## Contexto

La fase 1 es completamente local y gratuita. La máquina de desarrollo ya tiene un PostgreSQL nativo en el puerto 5432.

## Decisión

`make up` levanta **PostgreSQL 17 en el puerto 5433** (para convivir con uno nativo) y el **emulador de Firebase Auth** (puertos 9099 y 4000 para su interfaz). Mailpit y MinIO se añadirán cuando lleguen el email y el almacenamiento de archivos. Los servicios en la nube (Neon, Cloud Run, R2, Resend) son solo para el despliegue.

## Consecuencias

**Positivas**

- Coste cero y sin necesidad de cuentas para desarrollar.
- Los tests de integración usan las mismas imágenes.

**Negativas / contrapartidas**

- Hace falta Docker Desktop.

## Alternativas consideradas

- **Entorno de desarrollo compartido en la nube:** realista, pero cuesta dinero y necesita credenciales.
