# Frontend Coding Guidelines

React/TypeScript frontend guidelines for the Thunder project.

## Tech Stack

- React 19 with Vite and TypeScript.
- pnpm workspace managed by Nx.
- Testing: Vitest with `@testing-library/react`.
- Linting: ESLint. Each package may include its own lint configuration extending the shared base.
- Two apps: `thunder-gate` (auth UI), `thunder-develop` (console).

## Conventions

- Follow existing component patterns. Check sibling components before creating new ones.
- Place tests in `__tests__/` directories next to the source files.
- Use global mocks for commonly reused dependencies to avoid duplication across tests.
- Place manual mocks in `__mocks__/` directories.
- Test file names should match the component: `ComponentName.test.tsx`.
- Ensure test descriptions accurately match what the test body asserts.
- Use React Query patterns consistent with existing hooks in the codebase.
