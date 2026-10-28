# ADR-0007: Identificadores UUIDv7 generados en Go

[English](0007-uuidv7-identifiers.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Los ids deben poder mostrarse en URLs sin problema, ser únicos entre entornos y funcionar bien con índices B-tree y con la paginación por cursor. PostgreSQL 17 no tiene una función nativa de UUIDv7.

## Decisión

El id de toda entidad es un **UUIDv7**, generado en Go (`uuid.NewV7`) por el Store o por el servicio antes de llamar a una función de la base de datos.

## Consecuencias

**Positivas**

- Ordenados por tiempo: las inserciones van al final del índice y ordenar por id equivale a ordenar por creación, lo que simplifica los cursores.
- No se pueden adivinar como los enteros secuenciales y son únicos globalmente.

**Negativas / contrapartidas**

- 16 bytes en lugar de 8, y menos legibles que los números (las tareas tendrán claves cortas como `NEX-42`).
- Revelan la hora aproximada de creación.

## Alternativas consideradas

- **Enteros secuenciales:** compactos, pero adivinables y problemáticos entre entornos.
- **UUIDv4:** aleatorios, lo que fragmenta los índices y no da un orden natural.
