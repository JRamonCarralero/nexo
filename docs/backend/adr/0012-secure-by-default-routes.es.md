# ADR-0012: Rutas seguras por defecto, derivadas del contrato

[English](0012-secure-by-default-routes.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-23

## Contexto

Olvidarse de proteger una ruta nueva es un fallo de seguridad clásico. El generador de código no marca en el envoltorio de Gin qué operaciones requieren autenticación.

## Decisión

- El contrato aplica `bearerAuth` de forma **global**; una operación solo es pública si declara `security: []` (hoy solo `/health`).
- Al arrancar, el router lee el contrato embebido y construye el conjunto de rutas públicas.
- El middleware de autenticación exige token en **todas las rutas que no estén en ese conjunto**, incluidas las que ni siquiera están en el contrato.

## Consecuencias

**Positivas**

- Olvidarse de declarar una ruta nunca la hace pública.
- Las rutas públicas se ven y se revisan en el propio contrato.

**Negativas / contrapartidas**

- Hacer pública una ruta exige cambiar el contrato (es intencionado).

## Alternativas consideradas

- **Middleware por ruta en el código:** fácil de olvidar.
- **Middleware de validación desde el contrato (kin-openapi):** también valida los cuerpos, pero es más pesado; puede añadirse más adelante.
