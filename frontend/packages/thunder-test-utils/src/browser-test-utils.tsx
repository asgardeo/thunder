/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/* eslint-disable react-refresh/only-export-components */
import {useMemo, type ReactElement, type ReactNode} from 'react';
import {render as browserRender, renderHook as browserRenderHook} from 'vitest-browser-react';
import {MemoryRouter} from 'react-router';
import {OxygenUIThemeProvider} from '@wso2/oxygen-ui';
import {ConfigProvider} from '@thunder/shared-contexts';
import {LoggerProvider, LogLevel} from '@thunder/logger';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';

/**
 * Configuration options for Thunder test utilities
 */
export interface ThunderTestConfig {
  /**
   * Base path for the application (e.g., '/develop', '/gate')
   */
  base: string;
  /**
   * Client ID for the application
   */
  clientId: string;
  /**
   * Server hostname
   * @default 'localhost'
   */
  hostname?: string;
  /**
   * Server port
   * @default 8090
   */
  port?: number;
  /**
   * Whether to use HTTP only
   * @default false
   */
  httpOnly?: boolean;
}

interface ProvidersProps {
  children: ReactNode;
  queryClient?: QueryClient;
  config?: ThunderTestConfig;
  initialEntries?: string[];
}

// Default configuration for thunder-develop (backwards compatibility)
const defaultConfig: ThunderTestConfig = {
  base: '/develop',
  clientId: 'DEVELOP',
};

// Store the current config
let currentConfig: ThunderTestConfig = defaultConfig;

// Create a new QueryClient for each test to avoid shared state
function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

// Wrapper component with common providers
function Providers({children, queryClient = undefined, config = undefined, initialEntries = undefined}: ProvidersProps) {
  const testConfig = config ?? currentConfig;

  // Setup window.__THUNDER_RUNTIME_CONFIG__ for tests
  // Always set fresh config to avoid test pollution from previous tests
  if (typeof window !== 'undefined') {
    // eslint-disable-next-line no-underscore-dangle
    window.__THUNDER_RUNTIME_CONFIG__ = {
      client: {
        base: testConfig.base,
        client_id: testConfig.clientId,
      },
      server: {
        hostname: testConfig.hostname ?? 'localhost',
        port: testConfig.port ?? 8090,
        http_only: testConfig.httpOnly ?? false,
      },
    };
  }

  // Use useMemo to ensure the default QueryClient is only created once per mount,
  // preventing cache reset on re-renders when queryClient prop is not provided
  const client = useMemo(() => queryClient ?? createTestQueryClient(), [queryClient]);

  return (
    <MemoryRouter initialEntries={initialEntries}>
      <QueryClientProvider client={client}>
        <ConfigProvider>
          <LoggerProvider
            logger={{
              level: LogLevel.ERROR,
              transports: [],
            }}
          >
            <OxygenUIThemeProvider>{children}</OxygenUIThemeProvider>
          </LoggerProvider>
        </ConfigProvider>
      </QueryClientProvider>
    </MemoryRouter>
  );
}

/**
 * Configure the test utilities with app-specific settings
 * Call this in your test setup file before running tests
 *
 * Note: Configuration is set at the module level. Each app should call this
 * once in their test setup file. Tests within the same app share the same config.
 */
export function configureTestUtils(config: ThunderTestConfig): void {
  currentConfig = config;
}

/**
 * Reset test utilities to default configuration
 * Useful for cleanup in afterAll hooks or when switching between app configurations
 */
export function resetTestUtils(): void {
  currentConfig = defaultConfig;
}

// Custom render options that extend vitest-browser-react options
export interface ThunderRenderOptions {
  /**
   * Initial route entries for MemoryRouter
   */
  initialEntries?: string[];
  /**
   * Additional wrapper component (applied inside default providers)
   */
  wrapper?: React.ComponentType<{children: ReactNode}>;
}

/**
 * Custom render function for Vitest Browser Mode
 * Wraps components with all necessary providers and returns the native page object
 *
 * @example
 * ```typescript
 * import {render} from '@thunder/test-utils/browser';
 * import {page} from 'vitest/browser';
 *
 * await render(<MyComponent />);
 *
 * // Use page object for queries (native Vitest Browser Mode API)
 * const button = page.getByRole('button', {name: 'Submit'});
 * await button.click();
 *
 * const input = page.getByLabelText('Email');
 * await input.fill('test@example.com');
 *
 * // For multiple elements
 * const textboxes = page.getByRole('textbox').all();
 * expect(textboxes).toHaveLength(3);
 * ```
 */
export function render(ui: ReactElement, options?: ThunderRenderOptions) {
  const {wrapper: CustomWrapper, initialEntries} = options ?? {};

  const wrapper = ({children}: {children: ReactNode}) => {
    const content = CustomWrapper ? <CustomWrapper>{children}</CustomWrapper> : children;
    return (
      <Providers config={currentConfig} initialEntries={initialEntries}>
        {content}
      </Providers>
    );
  };

  // Render the component with providers
  return browserRender(ui, {wrapper});
}

/**
 * Alternative render function with providers
 * Alias for render to support different naming conventions
 */
export function renderWithProviders(ui: ReactElement, options?: ThunderRenderOptions) {
  return render(ui, options);
}

export interface ThunderRenderHookOptions<Props> {
  /**
   * Optional QueryClient instance for tests that need to manipulate cache
   */
  queryClient?: QueryClient;
  /**
   * Initial props for the hook
   */
  initialProps?: Props;
}

/**
 * Custom renderHook function for Vitest Browser Mode
 * Wraps hooks with necessary context providers for testing
 * Optionally accepts a queryClient for tests that need direct access to manipulate cache or spy on methods
 * Returns the hook result along with the queryClient instance for convenience
 */
export async function renderHook<Result, Props>(
  hook: (props?: Props) => Result,
  options?: ThunderRenderHookOptions<Props>,
) {
  const {queryClient: providedQueryClient, initialProps} = options ?? {};
  const queryClient = providedQueryClient ?? createTestQueryClient();

  const wrapper = ({children}: {children: ReactNode}) => <Providers queryClient={queryClient}>{children}</Providers>;

  const hookResult = await browserRenderHook(hook, {wrapper, initialProps});

  return {
    ...hookResult,
    queryClient,
  };
}

/**
 * Helper to get element by translation key
 * Useful when using mocked translations that return keys
 */
export function getByTranslationKey(container: HTMLElement, key: string) {
  return (
    container.querySelector(`[data-testid="${key}"]`) ??
    Array.from(container.querySelectorAll('*')).find((el) => el.textContent === key)
  );
}
