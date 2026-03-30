# Project Overview

Thunder is a lightweight user and identity management product. Go backend + React frontend in a monorepo. It provides authentication and authorization via OAuth2/OIDC, flexible orchestration flows, and individual auth mechanisms (password, passwordless, social login).

- [ARCHITECTURE.md](ARCHITECTURE.md)
- For build and running - Makefile and README.md.
- Documentation at /docs/content.

Thunder's login gate leverages v2 of the [Asgardeo JavaScript SDK](https://github.com/asgardeo/javascript), consumed via its published package in typical setups.
Clone the SDK repository only if you are developing or debugging the SDK itself, or testing Thunder against unreleased SDK changes.

## General Rules

- Do not add new dependencies without explicit approval.
- Do not modify CI/CD pipelines, GitHub Actions, or Makefiles without explicit approval.
- Do not edit mock files manually. Mocks are auto-generated via `make mockery`.
- Ensure all identity-related code aligns with relevant RFC specifications.
- Ensure proper error handling and logging at appropriate layers.
- Ensure that documentation is updated for any new features or significant changes.

## Git and PR Conventions

- Adhere to .github/pull_request_template.md

### Commit Messages
- Use short imperative sentences without conventional commit prefixes (no `feat:`, `fix:`, etc.).
- Reference the related issue or pull request when applicable (e.g., `Refs #123` or `Fixes #123`).

### One Commit Per Pull Request
- PRs are squash-merged, so the final commit history stays clean automatically.

## Contributing Guidelines

- [`docs/contributing/backend-guidelines.md`](docs/contributing/backend-guidelines.md) — Go backend: package structure, database patterns, error handling, service initialization, transactions, testing
- [`docs/contributing/frontend-guidelines.md`](docs/contributing/frontend-guidelines.md) — React/TypeScript: component patterns, testing, linting
- [`docs/AGENTS.md`](/docs/AGENTS.md) — Documentation authoring standards
