# ADR-0008: Paginación por cursor por defecto

[English](0008-cursor-pagination.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

Las listas de tareas cambian en tiempo real mientras se hace scroll. La paginación por offset salta o repite elementos cuando se insertan o mueven filas, y se vuelve más lenta en páginas profundas. Además, `COUNT(*)` no es gratis en PostgreSQL.

## Decisión

- `Store.List` usa **paginación por cursor (keyset)** por defecto: `(columna de orden, id)` comparado con la última fila de la página anterior; se pide una fila extra para saber si hay más.
- El cursor es opaco (JSON en base64url) y solo vale para la ordenación que lo generó.
- El **offset** está disponible para listados pequeños de administración, nunca combinado con un cursor.
- **`Count`** es una llamada aparte, solo cuando la interfaz muestra un total; los contadores por grupo usan un único `GROUP BY`.

## Consecuencias

**Positivas**

- Páginas estables con datos en vivo y coste constante por página.
- Encaja con el scroll infinito y con los botones de «cargar más».

**Negativas / contrapartidas**

- No se puede «saltar a la página 7»; las columnas de orden no pueden admitir `NULL`.
- Los cursores dejan de valer si cambia la ordenación (se rechazan con `INVALID_CURSOR`).

## Alternativas consideradas

- **Offset en todas partes:** números de página familiares, pero inestable y lento con datos en vivo.
