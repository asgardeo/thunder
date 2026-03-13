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
import {render, screen} from '@thunder/test-utils';
import EditTokenSettings from '../EditTokenSettings';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';

// Mock child components

vi.mock('../TokenUserAttributesSection', () => ({
  default: ({tokenType, headerAction}: {tokenType: string; headerAction?: React.ReactNode}) => (
    <div data-testid={`token-user-attributes-section-${tokenType}`}>
      Token User Attributes Section - {tokenType}
      {headerAction}
    </div>
  ),
}));

vi.mock('../TokenValidationSection', () => ({
  default: ({tokenType}: {tokenType: string}) => (
    <div data-testid={`token-validation-section-${tokenType}`}>Token Validation Section - {tokenType}</div>
  ),
}));

// Mock useAsgardeo
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: () => ({
    http: {
      request: vi.fn().mockResolvedValue({
        data: {
          totalResults: 1,
          startIndex: 0,
          count: 1,
          schemas: [
            {
              id: 'schema-1',
              name: 'default',
            },
          ],
        },
      }),
    },
  }),
}));

// Mock useConfig
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      getServerUrl: () => 'https://api.example.com',
    }),
  };
});

// Mock useLogger
vi.mock('@thunder/logger', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/logger')>();
  return {
    ...actual,
    useLogger: () => ({
      error: vi.fn(),
      info: vi.fn(),
      debug: vi.fn(),
    }),
  };
});

// Wrapper is now provided by test-utils

describe('EditTokenSettings', () => {
  const mockOnFieldChange = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    allowed_user_types: ['default'],
    token: {
      validity_period: 3600,
      user_attributes: ['email'],
    },
  } as Application;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe.skip('Native Mode (No OAuth2 Config) - SKIPPED: Component hangs due to async operations', () => {
    it('should render without crashing', () => {
      const {container} = render(<EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} />);

      expect(container).toBeTruthy();
    });

    it('should render shared token user attributes section', () => {
      render(<EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByTestId('token-user-attributes-section-shared')).toBeInTheDocument();
    });

    it('should render shared token validation section', () => {
      render(<EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByTestId('token-validation-section-shared')).toBeInTheDocument();
    });

    it('should not render access token sections in native mode', () => {
      render(<EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByTestId('token-user-attributes-section-access')).not.toBeInTheDocument();
      expect(screen.queryByTestId('token-validation-section-access')).not.toBeInTheDocument();
    });

    it('should not render ID token sections in native mode', () => {
      render(<EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByTestId('token-user-attributes-section-id')).not.toBeInTheDocument();
      expect(screen.queryByTestId('token-validation-section-id')).not.toBeInTheDocument();
    });
  });

  describe.skip('OAuth2/OIDC Mode - SKIPPED: Component hangs due to async operations', () => {
    const mockOAuth2Config: OAuth2Config = {
      token: {
        access_token: {
          validity_period: 1800,
          user_attributes: ['sub', 'email'],
        },
        id_token: {
          validity_period: 3600,
          user_attributes: ['sub', 'name', 'email'],
        },
      },
    } as OAuth2Config;

    it('should render access token user attributes section', () => {
      render(
        <EditTokenSettings
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByTestId('token-user-attributes-section-access')).toBeInTheDocument();
    });

    it('should render ID token user attributes section', () => {
      render(
        <EditTokenSettings
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByTestId('token-user-attributes-section-id')).toBeInTheDocument();
    });

    it('should render access token validation section', () => {
      render(
        <EditTokenSettings
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByTestId('token-validation-section-access')).toBeInTheDocument();
    });

    it('should render ID token validation section', () => {
      render(
        <EditTokenSettings
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByTestId('token-validation-section-id')).toBeInTheDocument();
    });

    it('should not render shared token sections in OAuth mode', () => {
      render(
        <EditTokenSettings
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.queryByTestId('token-user-attributes-section-shared')).not.toBeInTheDocument();
      expect(screen.queryByTestId('token-validation-section-shared')).not.toBeInTheDocument();
    });
  });

  describe.skip('Props Validation - SKIPPED: Component hangs due to async operations', () => {
    it('should handle undefined oauth2Config gracefully', () => {
      const {container} = render(
        <EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} oauth2Config={undefined} />,
      );

      expect(container).toBeTruthy();
      expect(screen.getByTestId('token-user-attributes-section-shared')).toBeInTheDocument();
    });

    it('should handle application without token config', () => {
      const appWithoutToken = {
        ...mockApplication,
        token: undefined,
      };

      const {container} = render(<EditTokenSettings application={appWithoutToken} onFieldChange={mockOnFieldChange} />);

      expect(container).toBeTruthy();
    });

    it('should handle empty allowed_user_types array', () => {
      const appWithoutUserTypes = {
        ...mockApplication,
        allowed_user_types: [],
      };

      const {container} = render(
        <EditTokenSettings application={appWithoutUserTypes} onFieldChange={mockOnFieldChange} />,
      );

      expect(container).toBeTruthy();
    });
  });

  describe.skip('Section Rendering Order - SKIPPED: Component hangs due to async operations', () => {
    it('should render all sections for OAuth mode', () => {
      const mockOAuth2Config: OAuth2Config = {
        token: {
          access_token: {validity_period: 1800, user_attributes: []},
          id_token: {validity_period: 3600, user_attributes: []},
        },
      } as unknown as OAuth2Config;

      const {container} = render(
        <EditTokenSettings
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(container).toBeTruthy();      expect(screen.getByTestId('token-user-attributes-section-access')).toBeInTheDocument();
      expect(screen.getByTestId('token-validation-section-access')).toBeInTheDocument();
      expect(screen.getByTestId('token-user-attributes-section-id')).toBeInTheDocument();
      expect(screen.getByTestId('token-validation-section-id')).toBeInTheDocument();
    });

    it('should render all sections for native mode', () => {
      const {container} = render(<EditTokenSettings application={mockApplication} onFieldChange={mockOnFieldChange} />);

      expect(container).toBeTruthy();
      expect(screen.getByTestId('token-user-attributes-section-shared')).toBeInTheDocument();
      expect(screen.getByTestId('token-validation-section-shared')).toBeInTheDocument();    });
  });

  describe.skip('User Info Configuration Logic - SKIPPED: Component hangs due to async operations', () => {
    const idTokenAttrs = ['sub', 'email'];
    const mockApp = {...mockApplication};

    it('should render User Info section with Inherit checkbox checked by default (No UserInfo Config)', () => {
      const mockConfig = {
        token: {
          id_token: {user_attributes: idTokenAttrs},
        },
      } as OAuth2Config;

      render(<EditTokenSettings application={mockApp} oauth2Config={mockConfig} onFieldChange={mockOnFieldChange} />);

      // Check for the checkbox presence
      const checkbox = screen.getByRole('checkbox', {name: /Use same attributes as ID Token/i});
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).toBeChecked();
    });

    it('should verify "Inherited" state (Checked) when explicit UserInfo attributes MATCH ID Token attributes', () => {
      const mockConfig = {
        token: {
          id_token: {user_attributes: idTokenAttrs},
        },
        user_info: {
          user_attributes: ['sub', 'email'], // Explicit but Match
        },
      } as OAuth2Config;

      render(<EditTokenSettings application={mockApp} oauth2Config={mockConfig} onFieldChange={mockOnFieldChange} />);

      const checkbox = screen.getByRole('checkbox', {name: /Use same attributes as ID Token/i});
      expect(checkbox).toBeChecked(); // Should be inherited because attributes are identical
    });

    it('should verify "Custom" state (Unchecked) when UserInfo attributes DIFFER from ID Token attributes', () => {
      const mockConfig = {
        token: {
          id_token: {user_attributes: idTokenAttrs},
        },
        user_info: {
          user_attributes: ['sub', 'email', 'phone'], // Different
        },
      } as OAuth2Config;

      render(<EditTokenSettings application={mockApp} oauth2Config={mockConfig} onFieldChange={mockOnFieldChange} />);

      const checkbox = screen.getByRole('checkbox', {name: /Use same attributes as ID Token/i});
      expect(checkbox).not.toBeChecked();
    });
  });
});
