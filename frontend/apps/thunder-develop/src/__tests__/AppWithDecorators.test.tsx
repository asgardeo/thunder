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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {ReactNode} from 'react';
import AppWithConfig from '../AppWithConfig';

const mockGetClientId = vi.fn();
const mockGetServerUrl = vi.fn();
const mockGetClientUrl = vi.fn();
const mockGetScopes = vi.fn();

// Mock the useConfig hook
vi.mock('@thunder/shared-contexts', () => ({
  useConfig: () => ({
    getClientId: mockGetClientId,
    getServerUrl: mockGetServerUrl,
    getClientUrl: mockGetClientUrl,
    getScopes: mockGetScopes,
  }),
}));

// Mock AsgardeoProvider
interface MockAsgardeoProviderProps {
  children: ReactNode;
  baseUrl?: string | null;
  clientId?: string | null;
  afterSignInUrl?: string | null;
  scopes?: string[];
}

vi.mock('@asgardeo/react', () => ({
  AsgardeoProvider: ({
    children,
    baseUrl = null,
    clientId = null,
    afterSignInUrl = null,
    scopes = undefined,
  }: MockAsgardeoProviderProps) => (
    <div
      data-testid="asgardeo-provider"
      data-base-url={baseUrl}
      data-client-id={clientId}
      data-after-sign-in-url={afterSignInUrl}
      data-scopes={scopes ? JSON.stringify(scopes) : undefined}
    >
      {children}
    </div>
  ),
  useAsgardeo: () => ({
    http: {
      request: vi.fn().mockResolvedValue({data: {language: 'en-US', translations: {}}}),
    },
  }),
}));

// Mock App component
vi.mock('../App', () => ({
  default: () => <div data-testid="app">App Component</div>,
}));

// Mock theme
vi.mock('@thunder/ui', () => ({
  theme: {
    palette: {
      mode: 'light',
    },
    typography: {
      fontWeightBold: 700,
    },
  },
}));

describe('AppWithConfig', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Set up default environment variables
    import.meta.env.VITE_ASGARDEO_BASE_URL = 'https://default-base.example.com';
    import.meta.env.VITE_ASGARDEO_CLIENT_ID = 'default-client-id';
    import.meta.env.VITE_ASGARDEO_AFTER_SIGN_IN_URL = 'https://default-signin.example.com';
    // Default to empty scopes
    mockGetScopes.mockReturnValue([]);
  });

  it('renders AsgardeoProvider with config values', async () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://test-server.example.com');
    await expect.element(provider).toHaveAttribute('data-client-id', 'test-client-id');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://test-client.example.com');
  });

  it('falls back to environment variables when config returns null', async () => {
    mockGetClientId.mockReturnValue(null);
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    await expect.element(provider).toHaveAttribute('data-client-id', 'default-client-id');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
  });

  it('renders App component', async () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');

    await render(<AppWithConfig />);

    await expect.element(page.getByTestId('app')).toBeInTheDocument();
  });

  it('uses config value for baseUrl when available', async () => {
    mockGetServerUrl.mockReturnValue('https://config-server.example.com');
    mockGetClientId.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://config-server.example.com');
  });

  it('uses config value for clientId when available', async () => {
    mockGetClientId.mockReturnValue('config-client-id');
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-client-id', 'config-client-id');
  });

  it('uses config value for afterSignInUrl when available', async () => {
    mockGetClientUrl.mockReturnValue('https://config-client.example.com');
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientId.mockReturnValue(null);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://config-client.example.com');
  });

  it('falls back to environment variables when config returns undefined', async () => {
    mockGetClientId.mockReturnValue(undefined);
    mockGetServerUrl.mockReturnValue(undefined);
    mockGetClientUrl.mockReturnValue(undefined);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    await expect.element(provider).toHaveAttribute('data-client-id', 'default-client-id');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
  });

  it('handles mixed config values and fallbacks - scenario 1', async () => {
    mockGetServerUrl.mockReturnValue('https://config-server.example.com');
    mockGetClientId.mockReturnValue(undefined);
    mockGetClientUrl.mockReturnValue('https://config-client.example.com');

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://config-server.example.com');
    await expect.element(provider).toHaveAttribute('data-client-id', 'default-client-id');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://config-client.example.com');
  });

  it('handles mixed config values and fallbacks - scenario 2', async () => {
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientId.mockReturnValue('config-client-id');
    mockGetClientUrl.mockReturnValue(null);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    await expect.element(provider).toHaveAttribute('data-client-id', 'config-client-id');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
  });

  it('uses config value for scopes when available', async () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');
    mockGetScopes.mockReturnValue(['openid', 'profile', 'email', 'system']);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-scopes', '["openid","profile","email","system"]');
  });

  it('does not pass scopes prop when config returns empty array', async () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');
    mockGetScopes.mockReturnValue([]);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).not.toHaveAttribute('data-scopes');
  });

  it('passes scopes when config has scopes', async () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');
    mockGetScopes.mockReturnValue(['openid', 'profile']);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-scopes', '["openid","profile"]');
  });

  it('handles scopes from config with other fallbacks', async () => {
    mockGetClientId.mockReturnValue(null);
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);
    mockGetScopes.mockReturnValue(['openid', 'profile', 'email']);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    await expect.element(provider).toHaveAttribute('data-client-id', 'default-client-id');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
    await expect.element(provider).toHaveAttribute('data-scopes', '["openid","profile","email"]');
  });

  it('properly evaluates falsy values for config options', async () => {
    // Test that falsy values (null, undefined, empty string, etc.) are properly handled
    // Empty strings are truthy in JavaScript, so they will be used as-is
    mockGetClientId.mockReturnValue('');
    mockGetServerUrl.mockReturnValue('');
    mockGetClientUrl.mockReturnValue('');
    mockGetScopes.mockReturnValue([]);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    // Empty strings are truthy, so they will be passed through (not fallback to env vars)
    await expect.element(provider).toHaveAttribute('data-base-url', '');
    await expect.element(provider).toHaveAttribute('data-client-id', '');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', '');
    await expect.element(provider).not.toHaveAttribute('data-scopes');
  });

  it('handles all config values as truthy strings', async () => {
    mockGetClientId.mockReturnValue('client-123');
    mockGetServerUrl.mockReturnValue('https://server.test');
    mockGetClientUrl.mockReturnValue('https://client.test');
    mockGetScopes.mockReturnValue(['scope1', 'scope2', 'scope3']);

    await render(<AppWithConfig />);

    const provider = page.getByTestId('asgardeo-provider');
    await expect.element(provider).toHaveAttribute('data-base-url', 'https://server.test');
    await expect.element(provider).toHaveAttribute('data-client-id', 'client-123');
    await expect.element(provider).toHaveAttribute('data-after-sign-in-url', 'https://client.test');
    await expect.element(provider).toHaveAttribute('data-scopes', '["scope1","scope2","scope3"]');
  });
});
