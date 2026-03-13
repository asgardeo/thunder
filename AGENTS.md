# AGENTS.md

Instructions for all AI coding agents (Claude, Codex, Gemini, Cursor, Copilot, etc.) working on this project.

## Project Overview

Thunder is a lightweight user and identity management product. Go backend + React frontend in a monorepo. It provides authentication and authorization via OAuth2/OIDC, flexible orchestration flows, and individual auth mechanisms (password, passwordless, social login).

## Contributing Guidelines

Read the relevant guide before working in that area:

- [`docs/contributing/backend-guidelines.md`](docs/contributing/backend-guidelines.md) — Go backend: package structure, database patterns, error handling, service initialization, transactions, testing
- [`docs/contributing/frontend-guidelines.md`](docs/contributing/frontend-guidelines.md) — React/TypeScript: component patterns, testing, linting
- [`docs/AGENTS.md`](/docs/AGENTS.md) — Documentation authoring standards

## General Rules

### Do NOT
- Add comments, docstrings, or type annotations to code you did not change.
- Refactor, "improve", or clean up surrounding code when fixing a bug or adding a feature.
- Add error handling for scenarios that cannot happen.
- Create new files unless absolutely necessary. Prefer editing existing files.
- Add new dependencies without explicit approval.
- Modify CI/CD pipelines, GitHub Actions, or Makefiles without explicit approval.
- Over-engineer. No premature abstractions, no feature flags, no backwards-compatibility shims.
- Generate or modify mock files manually. Mocks are auto-generated via `make mockery`.
- Add `// removed`, `// deprecated`, or placeholder comments for deleted code. Just delete it.
- Rename unused variables to `_` prefixed names. If unused, remove entirely — unless required by an interface, callback, or framework signature.
- Create fallback tests with mock/hardcoded data when original tests fail. Fix the actual failing tests instead of replacing them.

### Do
- Keep changes minimal and focused on the task requested.
- Follow existing patterns in the codebase. Match the style of surrounding code.
- Ensure all identity-related code aligns with relevant RFC specifications.
- Write tests for new features and bug fixes (target 80%+ coverage).
- Ensure proper error handling and logging at appropriate layers.
- Promote code reusability and define constants where applicable.

## Git and PR Conventions

### Commit Messages
- Use short imperative sentences without conventional commit prefixes (no `feat:`, `fix:`, etc.).
- Examples: "Add batch fetch support for users and groups", "Fix build failure caused by missing zod peer dep"

### One Commit Per PR
- PRs are squash-merged, so the final commit history stays clean automatically.
- Keep individual commits in the PR reasonably organized, but do not worry about squashing manually.

## File Naming Conventions

| Type | Convention | Example |
|------|-----------|---------|
| Go source files | `snake_case.go` | `error_constants.go` |
| Pure TypeScript (`.ts`) | `camelCase.ts` | `useCreateApplication.ts` |
| React components (`.tsx`) | `PascalCase.tsx` | `MenuButton.tsx` |

## Build and Test Commands

| Command | Description |
|---------|-------------|
| `make build` | Build everything |
| `make build_backend` | Build Go backend only |
| `make build_frontend` | Build frontend only |
| `make test_unit` | Run Go unit tests |
| `make test_integration` | Run Go integration tests |
| `make mockery` | Regenerate mocks (run after changing interfaces) |
| `make lint` | Run golangci-lint |
| `cd frontend && pnpm install` | Install frontend dependencies |
| `cd frontend && pnpm build` | Build all frontend packages and apps |
| `cd frontend && pnpm test` | Run frontend tests |

## Additional References

- REST API design: https://wso2.com/whitepapers/wso2-rest-apis-design-guidelines/
- Secure coding: https://security.docs.wso2.com/en/latest/security-guidelines/secure-engineering-guidelines/secure-coding-guidlines/general-recommendations-for-secure-coding/
