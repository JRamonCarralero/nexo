# ADR-0018: Estrategia de tests

[English](0018-testing-strategy.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

La seguridad (aislamiento entre espacios), el acceso a datos y los adaptadores importan más que un porcentaje de cobertura. Los tests deben ejecutarse igual en local y en el CI.

## Decisión

- **Unitarios** (`make test`, sin Docker): servicios con mocks de **mockery** (testify), tests por tabla, el router con adaptadores falsos y **`arch_test.go`**.
- **Integración** (`make test-integration`, build tag `integration`): **PostgreSQL** real y el **emulador de Firebase Auth** con **testcontainers**; un contenedor por paquete, con roles y migraciones aplicados, y dos conexiones (rol de la API y dueño).
- **Tests de aislamiento obligatorios** para toda entidad de espacio de trabajo, incluidas consultas sin `WHERE` para demostrar que la RLS funciona por sí sola.
- Siempre con `-race`; golangci-lint con depguard. Sin porcentaje de cobertura global obligatorio.

## Consecuencias

**Positivas**

- Las partes de más riesgo se prueban contra infraestructura real.
- Los mismos comandos se ejecutan en local y en el CI.

**Negativas / contrapartidas**

- Los tests de integración necesitan Docker y tardan segundos por paquete.
- La imagen del emulador se construye fuera de testcontainers porque los ayudantes de credenciales de Docker en local pueden romper la construcción.

## Alternativas consideradas

- **Mocks de la base de datos:** rápidos, pero no demuestran la RLS, las restricciones ni el SQL.
