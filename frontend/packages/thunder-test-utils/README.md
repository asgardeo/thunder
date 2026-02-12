# @thunder/test-utils

Shared testing utilities for ⚡️ Thunder applications. Provides common test setup, render helpers, and mocks for consistent testing across all Thunder apps.

## Features

- **Unified Test Setup** - Common test configuration for Vitest, jsdom, and React Testing Library
- **Custom Render Functions** - Pre-configured render with all necessary providers (QueryClient, Router, Config, Logger, Theme)
- **App Configuration** - Configurable settings for different Thunder apps (develop, gate)
- **Ready-to-Use Mocks** - Common mocks for i18n, IntersectionObserver, ResizeObserver, and more
- **Re-exported Utilities** - Convenient re-exports from @testing-library/react and user-event

## Installation

Since this is a workspace package, add it to your app's `package.json`:

```json
{
  "devDependencies": {
    "@thunder/test-utils": "workspace:^"
  }
}
```

Then install dependencies from the root:

```bash
pnpm install
```

## Quick Start

### 1. Configure Vitest

In your app's `vitest.config.ts`:

```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
  },
});
```

### 2. Create Test Setup File

Create `src/test/setup.ts` in your app:

```typescript
// Import shared test setup from @thunder/test-utils
import '@thunder/test-utils/setup';
import { configureTestUtils } from '@thunder/test-utils';

// Configure for your app (example for thunder-gate)
configureTestUtils({
  base: '/gate',
  clientId: 'GATE',
});
```

For `thunder-develop`, you can skip `configureTestUtils` as it uses the default configuration:

```typescript
// Import shared test setup (defaults to '/develop' and 'DEVELOP')
import '@thunder/test-utils/setup';
```

### 3. Write Tests

```tsx
import { describe, it, expect } from 'vitest';
import { renderWithProviders, screen } from '@thunder/test-utils';
import { MyComponent } from './MyComponent';

describe('MyComponent', () => {
  it('renders correctly', () => {
    renderWithProviders(<MyComponent />);
    expect(screen.getByText('Hello')).toBeInTheDocument();
  });
});
```

## API Reference

### Entry Points

The package exposes three entry points:

- **`@thunder/test-utils`** - Main entry with render functions and re-exports
- **`@thunder/test-utils/setup`** - Test setup (import in setup file)
- **`@thunder/test-utils/mocks`** - Mock implementations

### Main Exports (`@thunder/test-utils`)

#### `render`

Default custom render function that wraps components with all providers.

```tsx
import { render } from '@thunder/test-utils';

const { container } = render(<MyComponent />);
```

#### `renderWithProviders`

Alias for `render` with explicit naming.

```tsx
import { renderWithProviders } from '@thunder/test-utils';

renderWithProviders(<MyComponent />);
```

#### `renderHook`

Custom renderHook function with providers. Returns the QueryClient instance for direct cache manipulation.

```tsx
import { renderHook } from '@thunder/test-utils';

const { result, queryClient } = renderHook(() => useMyHook());

// Access the queryClient for cache manipulation
queryClient.setQueryData(['key'], mockData);
```

#### `configureTestUtils`

Configure test utilities with app-specific settings. Call this in your test setup file.

```typescript
import { configureTestUtils } from '@thunder/test-utils';

configureTestUtils({
  base: '/gate',        // Base path for the app
  clientId: 'GATE',     // Client ID for the app
  hostname: 'localhost', // Optional: server hostname (default: 'localhost')
  port: 8090,           // Optional: server port (default: 8090)
  httpOnly: false,      // Optional: use HTTP only (default: false)
});
```

#### `getByTranslationKey`

Helper to find elements by translation key when using mocked translations.

```tsx
import { getByTranslationKey } from '@thunder/test-utils';

const element = getByTranslationKey(container, 'users.title');
```

#### Re-exports

All exports from `@testing-library/react` are re-exported for convenience:

```tsx
import { screen, waitFor, within, fireEvent } from '@thunder/test-utils';
```

Additionally, `userEvent` from `@testing-library/user-event`:

```tsx
import { userEvent } from '@thunder/test-utils';

const user = userEvent.setup();
await user.click(button);
```

### Setup (`@thunder/test-utils/setup`)

Import this in your test setup file. It provides:

- **Jest-DOM matchers** - `toBeInTheDocument()`, `toHaveClass()`, etc.
- **i18n initialization** - Pre-configured with all Thunder translations
- **Automatic cleanup** - Cleanup after each test
- **Browser API mocks**:
  - `IntersectionObserver`
  - `ResizeObserver`
  - `HTMLMediaElement` (play, pause, load)
  - CSS variable handling for jsdom
- **Asgardeo mocks** - Mock implementation of `@asgardeo/react`

```typescript
// In your test setup file
import '@thunder/test-utils/setup';
```

### Mocks (`@thunder/test-utils/mocks`)

#### `mockUseTranslation`

Mock implementation of `useTranslation` hook that returns translation keys.

```typescript
import { vi } from 'vitest';
import { mockUseTranslation } from '@thunder/test-utils/mocks';

vi.mock('react-i18next', () => ({
  useTranslation: mockUseTranslation,
}));
```

#### `mockUseLanguage`

Mock implementation of `useLanguage` hook.

```typescript
import { mockUseLanguage } from '@thunder/test-utils/mocks';
```

#### `mockUseDataGridLocaleText`

Mock implementation of `useDataGridLocaleText` hook for DataGrid components.

```typescript
import { mockUseDataGridLocaleText } from '@thunder/test-utils/mocks';
```

## Configuration Options

### ThunderTestConfig

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `base` | `string` | `/develop` | Base path for the application |
| `clientId` | `string` | `DEVELOP` | Client ID for the application |
| `hostname` | `string` | `localhost` | Server hostname |
| `port` | `number` | `8090` | Server port |
| `httpOnly` | `boolean` | `false` | Whether to use HTTP only |

## Usage Examples

### Testing Components with React Query

```tsx
import { describe, it, expect } from 'vitest';
import { renderWithProviders, screen, waitFor } from '@thunder/test-utils';
import { UserList } from './UserList';

describe('UserList', () => {
  it('displays users after loading', async () => {
    renderWithProviders(<UserList />);

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument();
    });
  });
});
```

### Testing Hooks with QueryClient Access

```tsx
import { describe, it, expect } from 'vitest';
import { renderHook, waitFor } from '@thunder/test-utils';
import { useUsers } from './useUsers';

describe('useUsers', () => {
  it('fetches users', async () => {
    const { result, queryClient } = renderHook(() => useUsers());

    // Pre-populate cache if needed
    queryClient.setQueryData(['users'], [{ id: 1, name: 'John' }]);

    await waitFor(() => {
      expect(result.current.data).toHaveLength(1);
    });
  });
});
```

### Testing with User Interactions

```tsx
import { describe, it, expect, vi } from 'vitest';
import { renderWithProviders, screen, userEvent } from '@thunder/test-utils';
import { LoginForm } from './LoginForm';

describe('LoginForm', () => {
  it('submits the form', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();

    renderWithProviders(<LoginForm onSubmit={onSubmit} />);

    await user.type(screen.getByLabelText('Email'), 'test@example.com');
    await user.type(screen.getByLabelText('Password'), 'password123');
    await user.click(screen.getByRole('button', { name: 'Sign In' }));

    expect(onSubmit).toHaveBeenCalled();
  });
});
```

### Testing Components that Include Their Own Router

For components like `App` that include their own router, import the raw `render` directly from `@testing-library/react` to avoid router nesting:

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/react'; // Use raw render, not @thunder/test-utils
import App from './App';

// Mock app routes
vi.mock('./config/appRoutes', () => ({
  default: [],
}));

describe('App', () => {
  it('renders without crashing', () => {
    // App includes its own BrowserRouter, so we use the raw render
    // from @testing-library/react instead of @thunder/test-utils
    const { container } = render(<App />);
    expect(container).toBeInTheDocument();
  });
});
```

> **Note:** The `render` export from `@thunder/test-utils` wraps components with `MemoryRouter`. For components that include their own router (like the main `App` component), you must import `render` directly from `@testing-library/react`.

## Providers Included

The custom render functions wrap components with the following providers:

1. **MemoryRouter** - React Router for navigation
2. **QueryClientProvider** - TanStack Query with retry disabled for tests
3. **ConfigProvider** - Thunder configuration context
4. **LoggerProvider** - Thunder logger with ERROR level for minimal test output
5. **OxygenUIThemeProvider** - WSO2 Oxygen UI theming

## App-Specific Setup

### Thunder Develop

```typescript
// src/test/setup.ts
import '@thunder/test-utils/setup';
// Uses default config: base='/develop', clientId='DEVELOP'
```

### Thunder Gate

```typescript
// src/test/setup.ts
import '@thunder/test-utils/setup';
import { configureTestUtils } from '@thunder/test-utils';

configureTestUtils({
  base: '/gate',
  clientId: 'GATE',
});
```

## Troubleshooting

### Tests Failing with Provider Errors

Make sure you're using `renderWithProviders` or the default `render` from `@thunder/test-utils`, not the raw render from `@testing-library/react`:

```tsx
// ✅ Good
import { render } from '@thunder/test-utils';

// ❌ Avoid (unless you have a specific reason)
import { render } from '@testing-library/react';
```

### i18n Not Initialized

Ensure your test setup file imports the shared setup:

```typescript
import '@thunder/test-utils/setup';
```

### QueryClient State Leaking Between Tests

Each test gets a fresh QueryClient by default. If you need to pre-populate the cache, use the returned `queryClient` from `renderHook`:

```tsx
const { result, queryClient } = renderHook(() => useMyHook());
queryClient.setQueryData(['key'], mockData);
```

### CSS Variable Errors in Tests

The setup file includes a patch for CSS variable handling in jsdom. If you see CSS-related errors, make sure you're importing the setup file.

## Contributing

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for development setup and contribution guidelines.

## License

Apache-2.0
