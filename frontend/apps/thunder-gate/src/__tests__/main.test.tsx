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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';

// Mock ReactDOM
const mockRender = vi.fn();
const mockCreateRoot = vi.fn(() => ({
  render: mockRender,
}));
vi.mock('react-dom/client', () => ({
  createRoot: mockCreateRoot,
}));

// Mock AppWithConfig
vi.mock('../AppWithConfig', () => ({
  default: () => <div>App</div>,
}));

// Mock i18next
vi.mock('i18next', () => ({
  default: {
    use: vi.fn().mockReturnThis(),
    init: vi.fn().mockResolvedValue(undefined),
  },
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  initReactI18next: {},
}));

// Mock @thunder/i18n
vi.mock('@thunder/i18n/locales/en-US', () => ({
  default: {
    common: {},
    navigation: {},
    users: {},
    userTypes: {},
    integrations: {},
    applications: {},
    dashboard: {},
    auth: {},
    mfa: {},
    social: {},
    consent: {},
    errors: {},
  },
}));

// Mock @thunder/shared-contexts
vi.mock('@thunder/shared-contexts', () => ({
  ConfigProvider: ({children}: {children: React.ReactNode}) => children,
}));

// Mock @thunder/logger/react
vi.mock('@thunder/logger/react', () => ({
  LoggerProvider: ({children}: {children: React.ReactNode}) => children,
  LogLevel: {DEBUG: 0, INFO: 1, WARN: 2, ERROR: 3},
}));

// Mock @tanstack/react-query
vi.mock('@tanstack/react-query', () => ({
  // eslint-disable-next-line prefer-arrow-callback, func-names
  QueryClient: vi.fn().mockImplementation(function () {}),
  QueryClientProvider: ({children}: {children: React.ReactNode}) => children,
}));

// Mock @tanstack/react-query-devtools
vi.mock('@tanstack/react-query-devtools', () => ({
  ReactQueryDevtools: () => null,
}));

// Mock CSS import
vi.mock('../index.css', () => ({}));

describe('main', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clean up any existing root element
    const existingRoot = document.getElementById('root');
    if (existingRoot) {
      existingRoot.remove();
    }
    // Create a mock root element
    const root = document.createElement('div');
    root.id = 'root';
    document.body.appendChild(root);
  });

  afterEach(() => {
    // Clean up the root element
    const root = document.getElementById('root');
    if (root) {
      root.remove();
    }
  });

  it('should have a root element in the document', () => {
    const rootElement = document.getElementById('root');
    expect(rootElement).toBeInTheDocument();
  });

  it('should call createRoot and render when imported', async () => {
    // Reset modules to ensure clean test isolation for dynamic imports
    vi.resetModules();

    // Import main to trigger the render
    await import('../main');

    // Wait for async operations (i18n init)
    await vi.waitFor(() => {
      expect(mockCreateRoot).toHaveBeenCalled();
    });

    expect(mockRender).toHaveBeenCalled();
  });
});
