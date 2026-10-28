# ADR-0016: Campos personalizados en JSONB

[English](0016-custom-fields-jsonb.md) · **Español**

- **Estado:** Aceptado (aún no implementado)
- **Fecha:** 2026-10-22

## Contexto

Nexo debe ser muy configurable (campos personalizados, tipos de tarea) sin cambiar el esquema para cada espacio de trabajo.

## Decisión

Los valores personalizados se guardan en una columna **`custom_fields JSONB`**, descrita por una tabla `field_definitions` e indexada con GIN. El modelo de datos se prepara desde el MVP; la función llega en la fase 2.

## Consecuencias

**Positivas**

- Rápido, filtrable con índices y sin cambios de esquema por espacio.

**Negativas / contrapartidas**

- La comprobación de tipos de los valores se hace en la aplicación, no en los tipos de columna.

## Alternativas consideradas

- **Tabla entidad-atributo-valor:** flexible pero lenta y compleja de consultar.
- **Columnas dinámicas (`ALTER TABLE`):** inviable con miles de espacios de trabajo.
