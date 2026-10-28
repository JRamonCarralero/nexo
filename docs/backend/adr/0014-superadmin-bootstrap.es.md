# ADR-0014: Bootstrap del superadmin de la plataforma

[English](0014-superadmin-bootstrap.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-23

## Contexto

La plataforma necesita una cuenta de operador en cada entorno, creada sin SQL manual y sin guardar contraseñas.

## Decisión

- Al arrancar, `internal/bootstrap` ejecuta tareas idempotentes. Si `BOOTSTRAP_SUPERADMIN_EMAIL` está definida y no hay superadmin, convierte en superadmin al usuario con ese email (solo si está verificado) o crea una fila preparada que se vincula en el primer login con el email verificado.
- Nunca degrada ni borra a nadie. Una cuenta existente sin verificar **no** se convierte en superadmin (se registra un aviso).
- El superadmin es un rol de **plataforma**, distinto del admin de un espacio de trabajo, y **no** se salta el aislamiento entre espacios (un modo soporte auditado se diseñará aparte).
- Los datos de desarrollo (seed) son un mecanismo distinto que nunca se ejecuta en producción.

## Consecuencias

**Positivas**

- Reproducible en cualquier entorno con una sola variable.
- Protegido frente a quien registre un email que no es suyo.

**Negativas / contrapartidas**

- El superadmin todavía no tiene poderes; se añadirán con funciones explícitas y auditadas.

## Alternativas consideradas

- **SQL manual en cada entorno:** propenso a errores y sin documentar.
- **Superadmin que se salta la RLS:** cómodo, pero el mayor agujero de seguridad posible.
