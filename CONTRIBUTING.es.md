# Contribuir a Nexo

[English](CONTRIBUTING.md) · **Español**

¡Gracias por tu interés! Nexo está en una fase muy temprana; antes de dedicar tiempo a un cambio grande, abre un *issue* para comentarlo.

## Acuerdo de contribución (CLA)

Para poder aceptar contribuciones externas, cada colaborador debe firmar un **acuerdo de licencia de contribución (CLA)**. Se firma una sola vez, con un clic, desde el propio pull request. El CLA permite al proyecto mantener la licencia AGPL-3.0 y, si fuera necesario, ofrecer otras licencias en el futuro.

Lee el [CLA](CLA.es.md) ([in English](CLA.md)). Un bot comenta en tu primer pull request: firma respondiendo con la frase que te pide. Tu firma queda registrada en la rama `cla-signatures` y cubre todas tus contribuciones. Conservas los derechos de autor de tu trabajo. El job del CI se describe en [docs/ci.es.md](docs/ci.es.md#cla).

## Flujo de trabajo

1. Crea una rama desde `main`.
2. Haz tus cambios con tests.
3. Comprueba que todo pasa: `make lint test` (y `make test-integration` si tocas el backend; necesita Docker).
4. Abre un pull request. El CI debe estar en verde para hacer merge.

## Convenciones

- **Commits:** [Conventional Commits](https://www.conventionalcommits.org/es/) — `feat(tasks): añade vista de tablero`, `fix(api): …`. Un *git hook* lo comprueba.
- **Contrato primero:** los cambios de API empiezan en `api/openapi.yaml`; después `make generate`. No edites el código generado a mano.
- **Arquitectura:** el backend es un monolito modular. Las reglas de dependencias se comprueban en `backend/arch_test.go`: un módulo no importa a otro, y solo `internal/app` conoce las implementaciones concretas.
- **Errores:** usa `apperror` en el backend; los mensajes para el usuario se traducen en el frontend a partir de `code`.
- **Decisiones:** una decisión técnica importante (una librería nueva, un patrón, una regla de seguridad…) se registra como ADR en `docs/backend/adr/` o `docs/frontend/adr/`, en inglés y en español. La plantilla está en el índice de cada carpeta.
- **Documentación:** mantén `docs/` al día con el código, en inglés (`*.md`) y en español (`*.es.md`).
