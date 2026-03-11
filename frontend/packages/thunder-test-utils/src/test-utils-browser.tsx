/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
import {render, renderHook as browserRenderHook, type ComponentRenderOptions, type RenderHookOptions} from 'vitest-browser-react';
import {MemoryRouter} from 'react-router';
import {OxygenUIThemeProvider} from '@wso2/oxygen-ui';
import {ConfigProvider} from '@thunder/shared-contexts';
import {LoggerProvider, LogLevel} from '@thunder/logger';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import type {ThunderTestConfig} from './test-utils';

interface ProvidersProps {
  children: ReactNode;
  queryClient?: QueryClient;
  config?: ThunderTestConfig;
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
function Providers({children, queryClient = undefined, config = undefined}: ProvidersProps) {
  const testConfig = config ?? currentConfig;

  // Setup window.__THUNDER_RUNTIME_CONFIG__ for tests
  // eslint-disable-next-line no-underscore-dangle
  if (typeof window !== 'undefined' && !window.__THUNDER_RUNTIME_CONFIG__) {
    // eslint-disable-next-line no-underscore-dangle
    window.__THUNDER_RUNTIME_CONFIG__ = {
      brand: {
        product_name: 'Thunder',
      },
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
    <MemoryRouter>
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
 */
export function configureTestUtils(config: ThunderTestConfig): void {
  currentConfig = config;
}

// Custom render function that includes providers (async for Vitest Browser Mode)
async function customRender(ui: ReactElement, options?: ComponentRenderOptions) {
  const {wrapper: UserWrapper, ...restOptions} = options ?? {};
  const wrapper = ({children}: {children: ReactNode}) => {
    const content = UserWrapper ? <UserWrapper>{children}</UserWrapper> : children;
    return <Providers config={currentConfig}>{content}</Providers>;
  };
  return render(ui, {wrapper, ...restOptions});
}

/**
 * Alternative render function with providers (async for Vitest Browser Mode)
 * Alias for customRender to support different naming conventions
 */
export async function renderWithProviders(ui: ReactElement, options?: ComponentRenderOptions) {
  return customRender(ui, options ?? {});
}

interface RenderHookWithQueryClientOptions<Props> extends Omit<RenderHookOptions<Props>, 'wrapper'> {
  queryClient?: QueryClient;
  wrapper?: React.ComponentType<{children: ReactNode}>;
}

/**
 * Custom renderHook function that includes providers (async for Vitest Browser Mode)
 * Wraps hooks with necessary context providers for testing
 * Optionally accepts a queryClient for tests that need direct access to manipulate cache or spy on methods
 * Returns the queryClient instance for convenience
 */
export async function renderHook<Result, Props>(
  hook: (props?: Props) => Result,
  options?: RenderHookWithQueryClientOptions<Props>,
) {
  const {queryClient: providedQueryClient, wrapper: UserWrapper, ...restOptions} = options ?? {};
  const queryClient = providedQueryClient ?? createTestQueryClient();

  const wrapper = ({children}: {children: ReactNode}) => {
    const content = UserWrapper ? <UserWrapper>{children}</UserWrapper> : children;
    return <Providers queryClient={queryClient}>{content}</Providers>;
  };

  return {
    ...(await browserRenderHook(hook, {wrapper, ...restOptions})),
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

export default customRender;
