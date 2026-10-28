# ADR-0007: Tailwind CSS y componentes shadcn/ui

[English](0007-styling-and-components.md) · **Español**

- **Estado:** Aceptado (implementado en parte)
- **Fecha:** 2026-10-22

## Contexto

Nexo es para todo el mundo, incluidas personas no técnicas: debe ser accesible (teclado, contraste) y coherente, sin construir una librería de componentes desde cero.

## Decisión

- **Tailwind CSS 4** (plugin de Vite) para los estilos; el orden de las clases lo normaliza el plugin de Prettier. *Implementado.*
- **shadcn/ui** (primitivas de Radix, copiadas en `shared/ui`), **dnd-kit** para arrastrar y soltar y **cmdk** para la paleta de comandos `Ctrl+K`. *Se añadirán con las primeras pantallas.*

## Consecuencias

**Positivas**

- Primitivas accesibles (Radix) sin una dependencia pesada: los componentes son código nuestro.
- Iteración rápida con clases utilitarias.

**Negativas / contrapartidas**

- Listas largas de clases en el marcado.
- Los componentes copiados los actualizamos nosotros.

## Alternativas consideradas

- **MUI / Ant Design:** completas, pero pesadas y difíciles de adaptar a un aspecto propio.
