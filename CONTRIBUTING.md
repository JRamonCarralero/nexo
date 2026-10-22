# Contributing to Nexo

**English** · [Español](CONTRIBUTING.es.md)

Thanks for your interest! Nexo is at a very early stage; before spending time on a large change, please open an issue to discuss it.

## Contributor License Agreement (CLA)

To accept external contributions, every contributor must sign a **Contributor License Agreement (CLA)**. You sign it once, with a single click, from the pull request itself. The CLA lets the project keep the AGPL-3.0 license and, if ever needed, offer other licenses in the future.

> The CLA text and the bot that manages it will be published before the first external contribution is accepted.

## Workflow

1. Create a branch from `main`.
2. Make your changes, with tests.
3. Check that everything passes: `make lint test` (plus `make test-integration` if you touch the backend; it needs Docker).
4. Open a pull request. CI must be green to merge.

## Conventions

- **Commits:** [Conventional Commits](https://www.conventionalcommits.org/) — `feat(tasks): add board view`, `fix(api): …`. A git hook checks it.
- **Contract first:** API changes start in `api/openapi.yaml`, then `make generate`. Never edit generated code by hand.
- **Architecture:** the backend is a modular monolith. Dependency rules are checked in `backend/arch_test.go`: a module never imports another one, and only `internal/app` knows the concrete implementations.
- **Errors:** use `apperror` in the backend; user-facing messages are translated in the frontend from `code`.
