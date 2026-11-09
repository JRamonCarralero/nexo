# ADR-0012: Hand-written UI primitives and react-hook-form with Zod

**English** · [Español](0012-ui-primitives-and-forms.es.md)

- **Status:** Accepted
- **Date:** 2026-11-09

## Context

[ADR-0007](0007-styling-and-components.md) planned shadcn/ui (Radix) components. The first screens (sign in, forms, a members table, a few alerts) need buttons, fields, selects, alerts and cards, and need forms with validation that follows the current language.

## Decision

- The first screens use **small hand-written components** in `shared/ui` (`Button`, `Input`, `Select`, `Field`, `Alert`, `Card`, `Spinner`) built on native elements, which are accessible out of the box. Radix, dnd-kit and cmdk are added when a screen needs a dialog, menu, drag and drop or the command palette.
- `Field` wires label, hint and error to the control (`id`, `aria-invalid`, `aria-describedby`); errors are announced with `role="alert"`.
- Forms use **react-hook-form + Zod** (`zodResolver`). Schemas use **translation keys as messages**, translated when rendered, so they follow the language.
- Destructive actions ask for confirmation inline; no modal is needed yet.

## Consequences

**Positive**
- No unused dependencies; native controls behave well with keyboard and assistive technology.
- Validation messages are always in the current language.

**Negative / trade-offs**
- Components are ours to maintain; if they grow, we move to the Radix-based set of ADR-0007.
- Native `select` is less customisable than a custom listbox.

## Alternatives considered

- **Adopt shadcn/ui now:** pulls in overlay primitives we do not use yet.
- **Uncontrolled forms with manual validation:** more code and inconsistent messages.
