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
import {render, screen} from '@testing-library/react';
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

  it('renders AsgardeoProvider with config values', () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://test-server.example.com');
    expect(provider).toHaveAttribute('data-client-id', 'test-client-id');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://test-client.example.com');
  });

  it('falls back to environment variables when config returns null', () => {
    mockGetClientId.mockReturnValue(null);
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    expect(provider).toHaveAttribute('data-client-id', 'default-client-id');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
  });

  it('renders App component', () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');

    render(<AppWithConfig />);

    expect(screen.getByTestId('app')).toBeInTheDocument();
  });

  it('uses config value for baseUrl when available', () => {
    mockGetServerUrl.mockReturnValue('https://config-server.example.com');
    mockGetClientId.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://config-server.example.com');
  });

  it('uses config value for clientId when available', () => {
    mockGetClientId.mockReturnValue('config-client-id');
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-client-id', 'config-client-id');
  });

  it('uses config value for afterSignInUrl when available', () => {
    mockGetClientUrl.mockReturnValue('https://config-client.example.com');
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientId.mockReturnValue(null);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://config-client.example.com');
  });

  it('falls back to environment variables when config returns undefined', () => {
    mockGetClientId.mockReturnValue(undefined);
    mockGetServerUrl.mockReturnValue(undefined);
    mockGetClientUrl.mockReturnValue(undefined);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    expect(provider).toHaveAttribute('data-client-id', 'default-client-id');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
  });

  it('handles mixed config values and fallbacks - scenario 1', () => {
    mockGetServerUrl.mockReturnValue('https://config-server.example.com');
    mockGetClientId.mockReturnValue(undefined);
    mockGetClientUrl.mockReturnValue('https://config-client.example.com');

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://config-server.example.com');
    expect(provider).toHaveAttribute('data-client-id', 'default-client-id');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://config-client.example.com');
  });

  it('handles mixed config values and fallbacks - scenario 2', () => {
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientId.mockReturnValue('config-client-id');
    mockGetClientUrl.mockReturnValue(null);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    expect(provider).toHaveAttribute('data-client-id', 'config-client-id');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
  });

  it('uses config value for scopes when available', () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');
    mockGetScopes.mockReturnValue(['openid', 'profile', 'email', 'system']);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-scopes', '["openid","profile","email","system"]');
  });

  it('does not pass scopes prop when config returns empty array', () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');
    mockGetScopes.mockReturnValue([]);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).not.toHaveAttribute('data-scopes');
  });

  it('passes scopes when config has scopes', () => {
    mockGetClientId.mockReturnValue('test-client-id');
    mockGetServerUrl.mockReturnValue('https://test-server.example.com');
    mockGetClientUrl.mockReturnValue('https://test-client.example.com');
    mockGetScopes.mockReturnValue(['openid', 'profile']);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-scopes', '["openid","profile"]');
  });

  it('handles scopes from config with other fallbacks', () => {
    mockGetClientId.mockReturnValue(null);
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientUrl.mockReturnValue(null);
    mockGetScopes.mockReturnValue(['openid', 'profile', 'email']);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://default-base.example.com');
    expect(provider).toHaveAttribute('data-client-id', 'default-client-id');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://default-signin.example.com');
    expect(provider).toHaveAttribute('data-scopes', '["openid","profile","email"]');
  });

  it('properly evaluates falsy values for config options', () => {
    // Test that falsy values (null, undefined, empty string, etc.) are properly handled
    // Empty strings are truthy in JavaScript, so they will be used as-is
    mockGetClientId.mockReturnValue('');
    mockGetServerUrl.mockReturnValue('');
    mockGetClientUrl.mockReturnValue('');
    mockGetScopes.mockReturnValue([]);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    // Empty strings are truthy, so they will be passed through (not fallback to env vars)
    expect(provider).toHaveAttribute('data-base-url', '');
    expect(provider).toHaveAttribute('data-client-id', '');
    expect(provider).toHaveAttribute('data-after-sign-in-url', '');
    expect(provider).not.toHaveAttribute('data-scopes');
  });

  it('handles all config values as truthy strings', () => {
    mockGetClientId.mockReturnValue('client-123');
    mockGetServerUrl.mockReturnValue('https://server.test');
    mockGetClientUrl.mockReturnValue('https://client.test');
    mockGetScopes.mockReturnValue(['scope1', 'scope2', 'scope3']);

    render(<AppWithConfig />);

    const provider = screen.getByTestId('asgardeo-provider');
    expect(provider).toHaveAttribute('data-base-url', 'https://server.test');
    expect(provider).toHaveAttribute('data-client-id', 'client-123');
    expect(provider).toHaveAttribute('data-after-sign-in-url', 'https://client.test');
    expect(provider).toHaveAttribute('data-scopes', '["scope1","scope2","scope3"]');
  });
});
