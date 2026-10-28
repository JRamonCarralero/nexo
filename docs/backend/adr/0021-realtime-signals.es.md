# ADR-0021: Tiempo real mediante documentos de aviso en Firestore

[English](0021-realtime-signals.md) · **Español**

- **Estado:** Aceptado (aún no implementado)
- **Fecha:** 2026-10-22

## Contexto

Las conexiones WebSocket o SSE de larga duración mantienen ocupadas las instancias de Cloud Run y agotarían la capa gratuita con unos pocos usuarios conectados todo el día.

## Decisión

En cada cambio, el backend escribe un documento mínimo `signals/{projectId}` en Firestore; la web lo escucha con `onSnapshot` y vuelve a pedir los datos a la API. PostgreSQL sigue siendo la fuente de verdad. El mecanismo queda detrás de una interfaz `Notifier` (Go) y de un hook (React), para poder sustituirlo por WebSockets más adelante.

## Consecuencias

**Positivas**

- Cabe en la capa gratuita de Firestore y aprovecha la experiencia previa con él.

**Negativas / contrapartidas**

- Un servicio más; los clientes vuelven a pedir los datos en lugar de recibirlos.

## Alternativas consideradas

- **WebSockets/SSE en Cloud Run:** consume la capa gratuita.
- **Sondeo periódico:** lo más simple, menos tiempo real.
- **Ably / Pusher:** gestionados, pero otro proveedor.
