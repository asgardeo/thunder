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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {within} from '@testing-library/react';
import withConfig from '../withConfig';

// Use vi.hoisted so these are available inside vi.mock factories
const {capturedProviderProps, mockGetClientId, mockGetServerUrl, mockGetClientUrl, mockGetScopes} = vi.hoisted(() => ({
  capturedProviderProps: {current: {} as Record<string, unknown>},
  mockGetClientId: vi.fn(),
  mockGetServerUrl: vi.fn(),
  mockGetClientUrl: vi.fn(),
  mockGetScopes: vi.fn(),
}));

function MockChild() {
  return <div data-testid="mock-child">Child</div>;
}
const WithConfigComponent = withConfig(MockChild);

vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      getClientId: mockGetClientId,
      getServerUrl: mockGetServerUrl,
      getClientUrl: mockGetClientUrl,
      getScopes: mockGetScopes,
    }),
  };
});

vi.mock('@asgardeo/react', () => ({
  AsgardeoProvider: ({
    children,
    /* eslint-disable react/require-default-props */
    baseUrl,
    clientId,
    afterSignInUrl,
    scopes,
    /* eslint-enable react/require-default-props */
  }: {
    children: React.ReactNode;
    baseUrl?: string;
    clientId?: string;
    afterSignInUrl?: string;
    scopes?: string[];
  }) => {
    capturedProviderProps.current = {baseUrl, clientId, afterSignInUrl, scopes};
    return (
      <div
        data-testid="asgardeo-provider"
        data-base-url={baseUrl}
        data-client-id={clientId}
        data-after-sign-in-url={afterSignInUrl}
        data-scopes={scopes ? JSON.stringify(scopes) : undefined}
      >
        {children}
      </div>
    );
  },
}));

describe('withConfig (thunder-develop)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    capturedProviderProps.current = {};
    import.meta.env.VITE_ASGARDEO_BASE_URL = 'https://env-base.example.com';
    import.meta.env.VITE_ASGARDEO_CLIENT_ID = 'env-client-id';
    import.meta.env.VITE_ASGARDEO_AFTER_SIGN_IN_URL = 'https://env-signin.example.com';
    mockGetScopes.mockReturnValue([]);
  });

  it('renders without crashing', async () => {
    mockGetClientId.mockReturnValue('client-id');
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    const {container} = await render(<WithConfigComponent />);
    expect(container).toBeInTheDocument();
  });

  it('renders the wrapped component', async () => {
    mockGetClientId.mockReturnValue('client-id');
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    const {container} = await render(<WithConfigComponent />);
    expect(within(container).getByTestId('mock-child')).toBeInTheDocument();
  });

  it('wraps with AsgardeoProvider', async () => {
    mockGetClientId.mockReturnValue('client-id');
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    const {container} = await render(<WithConfigComponent />);
    expect(within(container).getByTestId('asgardeo-provider')).toBeInTheDocument();
  });

  it('passes baseUrl from useConfig to AsgardeoProvider', async () => {
    mockGetServerUrl.mockReturnValue('https://custom-server.example.com');
    mockGetClientId.mockReturnValue('client-id');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.baseUrl).toBe('https://custom-server.example.com');
  });

  it('passes clientId from useConfig to AsgardeoProvider', async () => {
    mockGetClientId.mockReturnValue('custom-client-id');
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.clientId).toBe('custom-client-id');
  });

  it('passes afterSignInUrl from useConfig to AsgardeoProvider', async () => {
    mockGetClientUrl.mockReturnValue('https://custom-client.example.com');
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientId.mockReturnValue('client-id');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.afterSignInUrl).toBe('https://custom-client.example.com');
  });

  it('falls back to env VITE_ASGARDEO_BASE_URL when getServerUrl returns null', async () => {
    mockGetServerUrl.mockReturnValue(null);
    mockGetClientId.mockReturnValue('client-id');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.baseUrl).toBe('https://env-base.example.com');
  });

  it('falls back to env VITE_ASGARDEO_CLIENT_ID when getClientId returns null', async () => {
    mockGetClientId.mockReturnValue(null);
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.clientId).toBe('env-client-id');
  });

  it('falls back to env VITE_ASGARDEO_AFTER_SIGN_IN_URL when getClientUrl returns null', async () => {
    mockGetClientUrl.mockReturnValue(null);
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientId.mockReturnValue('client-id');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.afterSignInUrl).toBe('https://env-signin.example.com');
  });

  it('passes scopes when config returns a non-empty array', async () => {
    mockGetScopes.mockReturnValue(['openid', 'profile', 'email']);
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientId.mockReturnValue('client-id');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.scopes).toEqual(['openid', 'profile', 'email']);
  });

  it('does not pass scopes when config returns empty array', async () => {
    mockGetScopes.mockReturnValue([]);
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientId.mockReturnValue('client-id');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    await render(<WithConfigComponent />);
    expect(capturedProviderProps.current.scopes).toBeUndefined();
  });

  it('wraps different components correctly', async () => {
    function AnotherChild() {
      return <div data-testid="another-child">Another</div>;
    }
    const AnotherWrapped = withConfig(AnotherChild);
    mockGetServerUrl.mockReturnValue('https://server.example.com');
    mockGetClientId.mockReturnValue('client-id');
    mockGetClientUrl.mockReturnValue('https://client.example.com');

    const {container} = await render(<AnotherWrapped />);
    expect(within(container).getByTestId('another-child')).toBeInTheDocument();
  });
});
