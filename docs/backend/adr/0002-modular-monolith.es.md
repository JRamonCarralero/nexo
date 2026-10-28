# ADR-0002: Monolito modular

[English](0002-modular-monolith.md) · **Español**

- **Estado:** Aceptado
- **Fecha:** 2026-10-22

## Contexto

El producto crecerá hasta tener muchos dominios (espacios de trabajo, proyectos, tareas, comentarios, notificaciones, automatizaciones, integraciones, agentes de IA). Con un solo desarrollador, los microservicios añaden coste operativo sin beneficio, pero un monolito clásico por capas tiende a convertirse en una maraña donde todo depende de todo.

## Decisión

Un **monolito modular**: un único binario dividido por dominio en `internal/modules/<nombre>`. Cada módulo tiene sus propias capas (handler → servicio → repositorio) y expone una superficie pequeña al resto. Reglas:

- Un módulo **nunca importa a otro módulo**. Cuando necesita algo de otro, declara una interfaz mínima (puerto) e `internal/app` conecta la implementación.
- Cada tabla tiene un único módulo dueño, el único que puede escribir en ella.
- La infraestructura vive en `internal/platform` y no conoce el negocio.
- Las reglas se comprueban en `backend/arch_test.go` (imports entre nuestros paquetes) y con `depguard` en golangci-lint (librerías de terceros).

## Consecuencias

**Positivas**

- Fácil de desplegar y depurar: un proceso, una base de datos y transacciones reales entre módulos.
- Las fronteras son explícitas y están probadas, así que sobreviven al crecimiento.
- Un módulo podría extraerse a un servicio más adelante sin rediseñarlo.

**Negativas / contrapartidas**

- Exige disciplina: las necesidades entre módulos pasan por puertos, que es más código que una llamada directa.
- Las pantallas que combinan varios módulos necesitan una capa de lectura (`internal/query`) en lugar de joins libres en los servicios.

## Alternativas consideradas

- **Monolito por capas** (handlers/, services/, repositories/ para todo): el más rápido para empezar, envejece mal.
- **Arquitectura hexagonal o clean pura:** dominio limpio pero mucha ceremonia en Go para un MVP mayoritariamente CRUD.
- **Microservicios:** coste sin beneficio con este tamaño.
