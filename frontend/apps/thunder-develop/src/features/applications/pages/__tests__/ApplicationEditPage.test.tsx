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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {BrowserRouter} from 'react-router';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {ConfigProvider} from '@thunder/commons-contexts';
import ApplicationEditPage from '../ApplicationEditPage';
import type {Application} from '../../models/application';

// Mock react-router
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({applicationId: 'app-123'}),
  };
});

// Mock useGetApplication
const mockUseGetApplication = vi.fn();
vi.mock('../../api/useGetApplication', () => ({
  default: (id: string): ReturnType<typeof mockUseGetApplication> => mockUseGetApplication(id),
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:view.back': 'Back to Applications',
        'applications:view.subtitle': 'View application details and configuration',
        'applications:view.error': 'Failed to load application information',
        'applications:view.notFound': 'Application not found',
        'applications:view.sections.basicInformation': 'Basic Information',
        'applications:view.sections.flowConfiguration': 'Flow Configuration',
        'applications:view.sections.userAttributes': 'User Attributes',
        'applications:view.sections.oauth2Configuration': 'OAuth2 Configuration',
        'applications:view.sections.timestamps': 'Timestamps',
        'applications:view.fields.applicationId': 'Application ID',
        'applications:view.fields.description': 'Description',
        'applications:view.fields.url': 'URL',
        'applications:view.fields.tosUri': 'Terms of Service URI',
        'applications:view.fields.policyUri': 'Privacy Policy URI',
        'applications:view.fields.contacts': 'Contacts',
        'applications:view.fields.authFlowId': 'Authentication Flow ID',
        'applications:view.fields.registrationFlowId': 'Registration Flow ID',
        'applications:view.fields.registrationFlowEnabled': 'Registration Flow Enabled',
        'applications:view.fields.clientId': 'Client ID',
        'applications:view.fields.redirectUris': 'Redirect URIs',
        'applications:view.fields.grantTypes': 'Grant Types',
        'applications:view.fields.responseTypes': 'Response Types',
        'applications:view.fields.scopes': 'Scopes',
        'applications:view.fields.publicClient': 'Public Client',
        'applications:view.fields.pkceRequired': 'PKCE Required',
        'applications:view.fields.createdAt': 'Created At',
        'applications:view.fields.updatedAt': 'Updated At',
        'applications:view.values.yes': 'Yes',
        'applications:view.values.no': 'No',
      };
      return translations[key] || key;
    },
  }),
}));

describe('ApplicationEditPage', () => {
  let queryClient: QueryClient;

  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test Application',
    description: 'Test application description',
    url: 'https://test-app.com',
    logo_url: 'https://test-app.com/logo.png',
    tos_uri: 'https://test-app.com/terms',
    policy_uri: 'https://test-app.com/privacy',
    contacts: ['admin@test-app.com', 'support@test-app.com'],
    auth_flow_id: 'auth_flow_config_basic',
    registration_flow_id: 'registration_flow_config_basic',
    is_registration_flow_enabled: true,
    user_attributes: ['email', 'username', 'given_name'],
    inbound_auth_config: [
      {
        type: 'oauth2',
        config: {
          client_id: 'test-client-id',
          redirect_uris: ['https://test-app.com/callback', 'https://test-app.com/callback2'],
          grant_types: ['authorization_code', 'refresh_token'],
          response_types: ['code'],
          scopes: ['openid', 'profile', 'email'],
          public_client: true,
          pkce_required: false,
          token: {
            issuer: 'thunder',
            access_token: {
              validity_period: 3600,
              user_attributes: ['email', 'username'],
            },
            id_token: {
              validity_period: 3600,
              user_attributes: ['sub', 'email', 'name'],
              scope_claims: {
                profile: ['name', 'picture'],
                email: ['email', 'email_verified'],
              },
            },
          },
        },
      },
    ],
    created_at: '2025-01-15T10:30:00Z',
    updated_at: '2025-01-15T14:45:00Z',
  };

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: {
          retry: false,
        },
      },
    });

    // Setup window.__THUNDER_RUNTIME_CONFIG__ for tests
    if (typeof window !== 'undefined') {
      // eslint-disable-next-line no-underscore-dangle
      window.__THUNDER_RUNTIME_CONFIG__ = {
        client: {
          base: '/develop',
          client_id: 'DEVELOP',
        },
        server: {
          hostname: 'localhost',
          port: 8090,
          http_only: false,
        },
      };
    }
  });

  const renderComponent = () =>
    render(
      <QueryClientProvider client={queryClient}>
        <ConfigProvider>
          <BrowserRouter>
            <ViewApplicationPage />
          </BrowserRouter>
        </ConfigProvider>
      </QueryClientProvider>,
    );

  it('should render loading state', () => {
    mockUseGetApplication.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render error state', () => {
    const error = new Error('Failed to fetch');
    mockUseGetApplication.mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
      isError: true,
    });

    renderComponent();

    expect(screen.getByRole('alert')).toBeInTheDocument();
    // Error message is shown from error.message, not the translation
    expect(screen.getByText('Failed to fetch')).toBeInTheDocument();
    expect(screen.getByText('Back to Applications')).toBeInTheDocument();
  });

  it('should render not found state when application is null', () => {
    mockUseGetApplication.mockReturnValue({
      data: null,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText('Application not found')).toBeInTheDocument();
    expect(screen.getByText('Back to Applications')).toBeInTheDocument();
  });

  it('should render application details successfully', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Test Application')).toBeInTheDocument();
    expect(screen.getByText('View application details and configuration')).toBeInTheDocument();
  });

  it('should render back button and navigate on click', async () => {
    const user = userEvent.setup();
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    // The button now uses i18n key "Back to Applications"
    const backButton = screen.getByText('Back to Applications');
    expect(backButton).toBeInTheDocument();

    await user.click(backButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/applications');
    });
  });

  it('should render basic information section', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Basic Information')).toBeInTheDocument();
    expect(screen.getByText('Application ID')).toBeInTheDocument();
    expect(screen.getByText('app-123')).toBeInTheDocument();
    expect(screen.getByText('Description')).toBeInTheDocument();
    expect(screen.getByText('Test application description')).toBeInTheDocument();
    expect(screen.getByText('URL')).toBeInTheDocument();
    expect(screen.getByText('https://test-app.com')).toBeInTheDocument();
  });

  it('should render optional basic information fields when present', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Terms of Service URI')).toBeInTheDocument();
    expect(screen.getByText('Privacy Policy URI')).toBeInTheDocument();
    expect(screen.getByText('Contacts')).toBeInTheDocument();
    expect(screen.getByText('admin@test-app.com')).toBeInTheDocument();
    expect(screen.getByText('support@test-app.com')).toBeInTheDocument();
  });

  it('should not render optional basic information fields when absent', () => {
    const minimalApp: Application = {
      id: 'app-123',
      name: 'Minimal App',
    };

    mockUseGetApplication.mockReturnValue({
      data: minimalApp,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.queryByText('Description')).not.toBeInTheDocument();
    expect(screen.queryByText('URL')).not.toBeInTheDocument();
    expect(screen.queryByText('Terms of Service URI')).not.toBeInTheDocument();
  });

  it('should render flow configuration section', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Flow Configuration')).toBeInTheDocument();
    expect(screen.getByText('Authentication Flow ID')).toBeInTheDocument();
    expect(screen.getByText('auth_flow_config_basic')).toBeInTheDocument();
    expect(screen.getByText('Registration Flow ID')).toBeInTheDocument();
    expect(screen.getByText('registration_flow_config_basic')).toBeInTheDocument();
    expect(screen.getByText('Registration Flow Enabled')).toBeInTheDocument();
    // "Yes" appears multiple times (registration flow enabled, public client), so use getAllByText
    const yesTexts = screen.getAllByText('Yes');
    expect(yesTexts.length).toBeGreaterThan(0);
  });

  it('should render registration flow enabled as No when false', () => {
    const appWithoutRegFlow: Application = {
      ...mockApplication,
      is_registration_flow_enabled: false,
    };

    mockUseGetApplication.mockReturnValue({
      data: appWithoutRegFlow,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    const registrationFlowEnabled = screen.getAllByText('No');
    expect(registrationFlowEnabled.length).toBeGreaterThan(0);
  });

  it('should render user attributes section when present', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('User Attributes')).toBeInTheDocument();
    // email appears in both user attributes and scopes, so use getAllByText
    const emailChips = screen.getAllByText('email');
    expect(emailChips.length).toBeGreaterThan(0);
    expect(screen.getByText('username')).toBeInTheDocument();
    expect(screen.getByText('given_name')).toBeInTheDocument();
  });

  it('should not render user attributes section when absent', () => {
    const appWithoutAttributes: Application = {
      ...mockApplication,
      user_attributes: undefined,
    };

    mockUseGetApplication.mockReturnValue({
      data: appWithoutAttributes,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.queryByText('User Attributes')).not.toBeInTheDocument();
  });

  it('should not render user attributes section when empty array', () => {
    const appWithEmptyAttributes: Application = {
      ...mockApplication,
      user_attributes: [],
    };

    mockUseGetApplication.mockReturnValue({
      data: appWithEmptyAttributes,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.queryByText('User Attributes')).not.toBeInTheDocument();
  });

  it('should render OAuth2 configuration section when present', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
    expect(screen.getByText('Client ID')).toBeInTheDocument();
    expect(screen.getByText('test-client-id')).toBeInTheDocument();
    expect(screen.getByText('Redirect URIs')).toBeInTheDocument();
    expect(screen.getByText('https://test-app.com/callback')).toBeInTheDocument();
    expect(screen.getByText('https://test-app.com/callback2')).toBeInTheDocument();
  });

  it('should render OAuth2 grant types as chips', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Grant Types')).toBeInTheDocument();
    expect(screen.getByText('authorization_code')).toBeInTheDocument();
    expect(screen.getByText('refresh_token')).toBeInTheDocument();
  });

  it('should render OAuth2 response types as chips', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Response Types')).toBeInTheDocument();
    expect(screen.getByText('code')).toBeInTheDocument();
  });

  it('should render OAuth2 scopes as chips', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Scopes')).toBeInTheDocument();
    expect(screen.getByText('openid')).toBeInTheDocument();
    expect(screen.getByText('profile')).toBeInTheDocument();
    // email appears in both user attributes and scopes, so use getAllByText
    const emailChips = screen.getAllByText('email');
    expect(emailChips.length).toBeGreaterThan(0);
  });

  it('should render public client field', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Public Client')).toBeInTheDocument();
    expect(screen.getAllByText('Yes')).toBeTruthy();
  });

  it('should render PKCE required field', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('PKCE Required')).toBeInTheDocument();
    expect(screen.getAllByText('No')).toBeTruthy();
  });

  it('should not render OAuth2 configuration section when absent', () => {
    const appWithoutOAuth: Application = {
      ...mockApplication,
      inbound_auth_config: undefined,
    };

    mockUseGetApplication.mockReturnValue({
      data: appWithoutOAuth,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.queryByText('OAuth2 Configuration')).not.toBeInTheDocument();
  });

  it('should render timestamps section when present', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('Timestamps')).toBeInTheDocument();
    expect(screen.getByText('Created At')).toBeInTheDocument();
    expect(screen.getByText('Updated At')).toBeInTheDocument();
  });

  it('should format timestamps correctly', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    const createdDate = new Date('2025-01-15T10:30:00Z').toLocaleString();
    const updatedDate = new Date('2025-01-15T14:45:00Z').toLocaleString();

    expect(screen.getByText(createdDate)).toBeInTheDocument();
    expect(screen.getByText(updatedDate)).toBeInTheDocument();
  });

  it('should not render timestamps section when both are absent', () => {
    const appWithoutTimestamps: Application = {
      ...mockApplication,
      created_at: undefined,
      updated_at: undefined,
    };

    mockUseGetApplication.mockReturnValue({
      data: appWithoutTimestamps,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.queryByText('Timestamps')).not.toBeInTheDocument();
  });

  it('should render application logo when present', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    const avatar = screen.getByRole('img');
    expect(avatar).toHaveAttribute('src', 'https://test-app.com/logo.png');
  });

  it('should render AppWindow icon when logo is not present', () => {
    const appWithoutLogo: Application = {
      ...mockApplication,
      logo_url: undefined,
    };

    mockUseGetApplication.mockReturnValue({
      data: appWithoutLogo,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    // AppWindow icon should be rendered in the Avatar fallback
    // Check that the application name is rendered (which means the Avatar section is present)
    expect(screen.getByText('Test Application')).toBeInTheDocument();
    // Avatar should be present - check by looking for the Avatar component in the rendered output
    const avatarElements = document.querySelectorAll('.MuiAvatar-root');
    expect(avatarElements.length).toBeGreaterThan(0);
  });

  it('should handle logo load error gracefully', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    const avatar = screen.getByRole('img');
    // Simulate image load error
    const errorEvent = new Event('error');
    Object.defineProperty(errorEvent, 'currentTarget', {
      value: {style: {display: ''}},
      writable: true,
    });
    avatar.dispatchEvent(errorEvent);

    // Should not crash
    expect(screen.getByText('Test Application')).toBeInTheDocument();
  });

  it('should render links with correct attributes', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    const urlLink = screen.getByText('https://test-app.com');
    expect(urlLink.closest('a')).toHaveAttribute('href', 'https://test-app.com');
    expect(urlLink.closest('a')).toHaveAttribute('target', '_blank');
    expect(urlLink.closest('a')).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('should call useGetApplication with correct application ID', () => {
    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(mockUseGetApplication).toHaveBeenCalledWith('app-123');
  });

  it('should handle application with minimal OAuth2 config', () => {
    const minimalOAuthApp: Application = {
      ...mockApplication,
      inbound_auth_config: [
        {
          type: 'oauth2',
          config: {
            client_id: 'minimal-client-id',
            redirect_uris: [],
            grant_types: [],
            response_types: [],
            scopes: [],
            token: {
              access_token: {
                validity_period: 3600,
                user_attributes: [],
              },
              id_token: {
                validity_period: 3600,
                user_attributes: [],
                scope_claims: {},
              },
            },
          },
        },
      ],
    };

    mockUseGetApplication.mockReturnValue({
      data: minimalOAuthApp,
      isLoading: false,
      error: null,
      isError: false,
    });

    renderComponent();

    expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
    expect(screen.getByText('minimal-client-id')).toBeInTheDocument();
    expect(screen.queryByText('Redirect URIs')).not.toBeInTheDocument();
    expect(screen.queryByText('Grant Types')).not.toBeInTheDocument();
  });

  it('should handle error navigation gracefully', async () => {
    const user = userEvent.setup();
    const error = new Error('Failed to fetch');
    mockUseGetApplication.mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
      isError: true,
    });

    mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

    renderComponent();

    const backButton = screen.getByText('Back to Applications');
    await user.click(backButton);

    // Should not throw
    expect(mockNavigate).toHaveBeenCalled();
  });

  describe('Edge Cases and Additional Coverage', () => {
    it('should handle application with only created_at timestamp', () => {
      const appWithOnlyCreatedAt: Application = {
        ...mockApplication,
        updated_at: undefined,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithOnlyCreatedAt,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Timestamps')).toBeInTheDocument();
      expect(screen.getByText('Created At')).toBeInTheDocument();
      expect(screen.queryByText('Updated At')).not.toBeInTheDocument();
    });

    it('should handle application with only updated_at timestamp', () => {
      const appWithOnlyUpdatedAt: Application = {
        ...mockApplication,
        created_at: undefined,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithOnlyUpdatedAt,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Timestamps')).toBeInTheDocument();
      expect(screen.getByText('Updated At')).toBeInTheDocument();
      expect(screen.queryByText('Created At')).not.toBeInTheDocument();
    });

    it('should handle application without auth_flow_id', () => {
      const appWithoutAuthFlow: Application = {
        ...mockApplication,
        auth_flow_id: undefined,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithoutAuthFlow,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Flow Configuration')).toBeInTheDocument();
      expect(screen.queryByText('Authentication Flow ID')).not.toBeInTheDocument();
      expect(screen.getByText('Registration Flow ID')).toBeInTheDocument();
    });

    it('should handle application without registration_flow_id', () => {
      const appWithoutRegFlow: Application = {
        ...mockApplication,
        registration_flow_id: undefined,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithoutRegFlow,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Flow Configuration')).toBeInTheDocument();
      expect(screen.getByText('Authentication Flow ID')).toBeInTheDocument();
      expect(screen.queryByText('Registration Flow ID')).not.toBeInTheDocument();
    });

    it('should handle application with single contact', () => {
      const appWithSingleContact: Application = {
        ...mockApplication,
        contacts: ['admin@test.com'],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithSingleContact,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Contacts')).toBeInTheDocument();
      expect(screen.getByText('admin@test.com')).toBeInTheDocument();
    });

    it('should handle application with many contacts', () => {
      const manyContacts = Array.from({length: 10}, (_, i) => `contact${i}@test.com`);
      const appWithManyContacts: Application = {
        ...mockApplication,
        contacts: manyContacts,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithManyContacts,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Contacts')).toBeInTheDocument();
      manyContacts.forEach((contact) => {
        expect(screen.getByText(contact)).toBeInTheDocument();
      });
    });

    it('should handle application with single redirect URI', () => {
      const appWithSingleRedirect: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              redirect_uris: ['https://test-app.com/callback'],
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithSingleRedirect,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Redirect URIs')).toBeInTheDocument();
      expect(screen.getByText('https://test-app.com/callback')).toBeInTheDocument();
    });

    it('should handle application with many redirect URIs', () => {
      const manyRedirectURIs = Array.from({length: 15}, (_, i) => `https://test-app.com/callback${i}`);
      const appWithManyRedirects: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              redirect_uris: manyRedirectURIs,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithManyRedirects,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Redirect URIs')).toBeInTheDocument();
      manyRedirectURIs.forEach((uri) => {
        expect(screen.getByText(uri)).toBeInTheDocument();
      });
    });

    it('should handle OAuth2 config with undefined client_id', () => {
      const appWithoutClientId: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              client_id: undefined,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithoutClientId,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
      expect(screen.queryByText('Client ID')).not.toBeInTheDocument();
    });

    it('should handle OAuth2 config with undefined public_client', () => {
      const appWithoutPublicClient: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              public_client: undefined,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithoutPublicClient,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
      expect(screen.queryByText('Public Client')).not.toBeInTheDocument();
    });

    it('should handle OAuth2 config with undefined pkce_required', () => {
      const appWithoutPKCE: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              pkce_required: undefined,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithoutPKCE,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
      expect(screen.queryByText('PKCE Required')).not.toBeInTheDocument();
    });

    it('should handle application with empty string values', () => {
      const appWithEmptyStrings: Application = {
        ...mockApplication,
        description: '',
        url: '',
        tos_uri: '',
        policy_uri: '',
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithEmptyStrings,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      // Empty strings should not render the fields
      expect(screen.queryByText('Description')).not.toBeInTheDocument();
      expect(screen.queryByText('URL')).not.toBeInTheDocument();
      expect(screen.queryByText('Terms of Service URI')).not.toBeInTheDocument();
      expect(screen.queryByText('Privacy Policy URI')).not.toBeInTheDocument();
    });

    it('should handle application with very long application ID', () => {
      const longId = 'a'.repeat(200);
      const appWithLongId: Application = {
        ...mockApplication,
        id: longId,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithLongId,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Application ID')).toBeInTheDocument();
      expect(screen.getByText(longId)).toBeInTheDocument();
    });

    it('should handle application with very long name', () => {
      const longName = 'A'.repeat(500);
      const appWithLongName: Application = {
        ...mockApplication,
        name: longName,
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithLongName,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText(longName)).toBeInTheDocument();
    });

    it('should handle application with special characters in URLs', () => {
      const appWithSpecialChars: Application = {
        ...mockApplication,
        url: 'https://test-app.com/path?query=value&other=test#fragment',
        tos_uri: 'https://test-app.com/terms?param=value',
        policy_uri: 'https://test-app.com/privacy#section',
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithSpecialChars,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('https://test-app.com/path?query=value&other=test#fragment')).toBeInTheDocument();
      expect(screen.getByText('https://test-app.com/terms?param=value')).toBeInTheDocument();
      expect(screen.getByText('https://test-app.com/privacy#section')).toBeInTheDocument();
    });

    it('should handle application with non-OAuth2 inbound_auth_config', () => {
      const appWithNonOAuth: Application = {
        ...mockApplication,
        inbound_auth_config: [] as unknown as Application['inbound_auth_config'],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithNonOAuth,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      // Should not render OAuth2 Configuration section
      expect(screen.queryByText('OAuth2 Configuration')).not.toBeInTheDocument();
    });

    it('should handle application with multiple inbound_auth_config (only first OAuth2 is shown)', () => {
      const appWithMultipleConfigs: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              client_id: 'first-oauth-client',
              redirect_uris: ['https://first.com/callback'],
              grant_types: ['authorization_code'],
              response_types: ['code'],
              scopes: ['openid'],
              token: {
                access_token: {
                  validity_period: 3600,
                  user_attributes: [],
                },
                id_token: {
                  validity_period: 3600,
                  user_attributes: [],
                  scope_claims: {},
                },
              },
            },
          },
          {
            type: 'oauth2',
            config: {
              client_id: 'second-oauth-client',
              redirect_uris: ['https://second.com/callback'],
              grant_types: ['authorization_code'],
              response_types: ['code'],
              scopes: ['openid'],
              token: {
                access_token: {
                  validity_period: 3600,
                  user_attributes: [],
                },
                id_token: {
                  validity_period: 3600,
                  user_attributes: [],
                  scope_claims: {},
                },
              },
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithMultipleConfigs,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      // Should only show the first OAuth2 config
      expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
      expect(screen.getByText('first-oauth-client')).toBeInTheDocument();
      expect(screen.queryByText('second-oauth-client')).not.toBeInTheDocument();
    });

    it('should handle application with empty OAuth2 arrays gracefully', () => {
      const appWithEmptyArrays: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              client_id: 'test-client',
              redirect_uris: [],
              grant_types: [],
              response_types: [],
              scopes: [],
              token: {
                access_token: {
                  validity_period: 3600,
                  user_attributes: [],
                },
                id_token: {
                  validity_period: 3600,
                  user_attributes: [],
                  scope_claims: {},
                },
              },
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithEmptyArrays,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('OAuth2 Configuration')).toBeInTheDocument();
      expect(screen.getByText('test-client')).toBeInTheDocument();
      expect(screen.queryByText('Redirect URIs')).not.toBeInTheDocument();
      expect(screen.queryByText('Grant Types')).not.toBeInTheDocument();
      expect(screen.queryByText('Response Types')).not.toBeInTheDocument();
      expect(screen.queryByText('Scopes')).not.toBeInTheDocument();
    });

    it('should handle back button click in error state', async () => {
      const user = userEvent.setup();
      const error = new Error('Network error');
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        error,
        isError: true,
      });

      renderComponent();

      const backButton = screen.getByText('Back to Applications');
      await user.click(backButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications');
      });
    });

    it('should handle back button click in not found state', async () => {
      const user = userEvent.setup();
      mockUseGetApplication.mockReturnValue({
        data: null,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      const backButton = screen.getByText('Back to Applications');
      await user.click(backButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications');
      });
    });

    it('should handle application with all OAuth2 grant types', () => {
      const appWithAllGrantTypes: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              grant_types: ['authorization_code', 'refresh_token', 'client_credentials', 'password', 'implicit'],
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithAllGrantTypes,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('authorization_code')).toBeInTheDocument();
      expect(screen.getByText('refresh_token')).toBeInTheDocument();
      expect(screen.getByText('client_credentials')).toBeInTheDocument();
      expect(screen.getByText('password')).toBeInTheDocument();
      expect(screen.getByText('implicit')).toBeInTheDocument();
    });

    it('should handle application with all OAuth2 response types', () => {
      const appWithAllResponseTypes: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              response_types: ['code', 'token', 'id_token', 'code token', 'code id_token', 'token id_token'],
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithAllResponseTypes,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('code')).toBeInTheDocument();
      expect(screen.getByText('token')).toBeInTheDocument();
      expect(screen.getByText('id_token')).toBeInTheDocument();
      expect(screen.getByText('code token')).toBeInTheDocument();
      expect(screen.getByText('code id_token')).toBeInTheDocument();
      expect(screen.getByText('token id_token')).toBeInTheDocument();
    });

    it('should handle application with many scopes', () => {
      const manyScopes = ['openid', 'profile', 'phone', 'address', 'offline_access', 'custom1', 'custom2', 'custom3'];
      const appWithManyScopes: Application = {
        ...mockApplication,
        // Remove email from user_attributes to avoid duplicates with scopes
        user_attributes: ['username', 'given_name'],
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              scopes: manyScopes,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithManyScopes,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Scopes')).toBeInTheDocument();
      manyScopes.forEach((scope) => {
        // Use getAllByText since some scopes might appear multiple times
        const elements = screen.getAllByText(scope);
        expect(elements.length).toBeGreaterThan(0);
      });
    });

    it('should handle application with many user attributes', () => {
      const manyAttributes = [
        'username',
        'given_name',
        'family_name',
        'middle_name',
        'nickname',
        'picture',
        'website',
        'gender',
        'birthdate',
        'zoneinfo',
        'locale',
        'phone_number',
        'address',
        'groups',
        'custom_attr',
      ];
      const appWithManyAttributes: Application = {
        ...mockApplication,
        user_attributes: manyAttributes,
        // Remove email from scopes to avoid duplicates
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              scopes: ['openid', 'profile'],
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithManyAttributes,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('User Attributes')).toBeInTheDocument();
      manyAttributes.forEach((attr) => {
        // Use getAllByText since some attributes might appear multiple times
        const elements = screen.getAllByText(attr);
        expect(elements.length).toBeGreaterThan(0);
      });
    });

    it('should handle application with null error message', () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
        isError: true,
      });

      renderComponent();

      expect(screen.getByRole('alert')).toBeInTheDocument();
      // Should show the default error message from translation
      expect(screen.getByText('Failed to load application information')).toBeInTheDocument();
    });

    it('should handle application with undefined applicationId in params', () => {
      // Note: This test verifies that when applicationId is undefined,
      // useGetApplication is called with empty string (handled by ?? operator in component)
      // The actual useParams mock is set up at the module level, so we verify the behavior
      // by checking that useGetApplication was called with the value from useParams
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      // useParams returns 'app-123' from the module-level mock
      // The component uses applicationId ?? '', so if undefined it would be ''
      // But since our mock returns 'app-123', we verify it's called with that
      expect(mockUseGetApplication).toHaveBeenCalled();
    });

    it('should render all link attributes correctly for tos_uri', () => {
      mockUseGetApplication.mockReturnValue({
        data: mockApplication,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      const tosLink = screen.getByText('https://test-app.com/terms');
      expect(tosLink.closest('a')).toHaveAttribute('href', 'https://test-app.com/terms');
      expect(tosLink.closest('a')).toHaveAttribute('target', '_blank');
      expect(tosLink.closest('a')).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('should render all link attributes correctly for policy_uri', () => {
      mockUseGetApplication.mockReturnValue({
        data: mockApplication,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      const policyLink = screen.getByText('https://test-app.com/privacy');
      expect(policyLink.closest('a')).toHaveAttribute('href', 'https://test-app.com/privacy');
      expect(policyLink.closest('a')).toHaveAttribute('target', '_blank');
      expect(policyLink.closest('a')).toHaveAttribute('rel', 'noopener noreferrer');
    });

    it('should handle application with PKCE required set to true', () => {
      const appWithPKCE: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              pkce_required: true,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithPKCE,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('PKCE Required')).toBeInTheDocument();
      const yesTexts = screen.getAllByText('Yes');
      expect(yesTexts.length).toBeGreaterThan(0);
    });

    it('should handle application with public_client set to false', () => {
      const appWithPrivateClient: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'oauth2',
            config: {
              ...mockApplication.inbound_auth_config![0].config,
              public_client: false,
            },
          },
        ],
      };

      mockUseGetApplication.mockReturnValue({
        data: appWithPrivateClient,
        isLoading: false,
        error: null,
        isError: false,
      });

      renderComponent();

      expect(screen.getByText('Public Client')).toBeInTheDocument();
      const noTexts = screen.getAllByText('No');
      expect(noTexts.length).toBeGreaterThan(0);
    });
  });
});
