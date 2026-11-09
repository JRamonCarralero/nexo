# ADR-0012: Primitivas de UI propias y react-hook-form con Zod

[English](0012-ui-primitives-and-forms.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-11-09

## Contexto

El [ADR-0007](0007-styling-and-components.es.md) preveía componentes shadcn/ui (Radix). Las primeras pantallas (inicio de sesión, formularios, una tabla de miembros, algunas alertas) necesitan botones, campos, selectores, alertas y tarjetas, y formularios con validación que siga el idioma actual.

## Decisión

- Las primeras pantallas usan **componentes pequeños escritos a mano** en `shared/ui` (`Button`, `Input`, `Select`, `Field`, `Alert`, `Card`, `Spinner`) sobre elementos nativos, que ya son accesibles. Radix, dnd-kit y cmdk se añadirán cuando una pantalla necesite un diálogo, un menú, arrastrar y soltar o la paleta de comandos.
- `Field` conecta etiqueta, ayuda y error con el control (`id`, `aria-invalid`, `aria-describedby`); los errores se anuncian con `role="alert"`.
- Los formularios usan **react-hook-form + Zod** (`zodResolver`). Los esquemas usan **claves de traducción como mensajes**, traducidas al mostrarse, así siguen el idioma.
- Las acciones destructivas piden confirmación en línea; de momento no hace falta un modal.

## Consecuencias

**Positivas**
- Sin dependencias sin usar; los controles nativos se comportan bien con teclado y tecnologías de apoyo.
- Los mensajes de validación siempre están en el idioma actual.

**Negativas / contrapartidas**
- Mantenemos los componentes; si crecen, pasaremos al conjunto basado en Radix del ADR-0007.
- El `select` nativo es menos personalizable que un listbox propio.

## Alternativas consideradas

- **Adoptar shadcn/ui ya:** trae primitivas de overlays que aún no usamos.
- **Formularios no controlados con validación manual:** más código y mensajes inconsistentes.
