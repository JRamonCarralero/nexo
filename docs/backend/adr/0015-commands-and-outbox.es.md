# ADR-0015: Comandos y eventos con outbox transaccional

[English](0015-commands-and-outbox.md) · **Español**

- **Estado:** Aceptado (aún no implementado)
- **Fecha:** 2026-10-22

## Contexto

El historial de actividad, las notificaciones, los avisos en tiempo real y, más adelante, las automatizaciones, los webhooks y los agentes de IA reaccionan a los mismos cambios. Las personas y los agentes deben seguir las mismas reglas.

## Decisión

- Toda escritura es un **comando** (`CreateTask`, `MoveTask`…) que comprueba permisos y reglas, cambia los datos y guarda un **evento** en `activity_events` **en la misma transacción** (patrón outbox).
- Un despachador (primero dentro del proceso, después Pub/Sub) entrega los eventos a los consumidores: historial de actividad, notificaciones y avisos en tiempo real; más adelante automatizaciones, webhooks y agentes.
- **No** es event sourcing: el estado vive en tablas normales; los eventos son registro y disparador.
- Esta capa de comandos es también lo que hará posible migrar a local-first más adelante.

## Consecuencias

**Positivas**

- Las funciones nuevas se enchufan como consumidores de eventos sin tocar el código existente.
- Nunca hay un cambio sin su evento ni un evento sin su cambio.

**Negativas / contrapartidas**

- Más piezas: despachador, consumidores idempotentes y reintentos.

## Alternativas consideradas

- **Llamar a los efectos secundarios directamente desde los servicios:** más simple, pero se pierden si algo falla y acopla los módulos.
- **Event sourcing completo:** potente, pero pesado para este producto.
