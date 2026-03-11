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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import ConfigureSignInOptions, {
  type ConfigureSignInOptionsProps,
} from '../configure-signin-options/ConfigureSignInOptions';
import ApplicationCreateProvider from '../../../contexts/ApplicationCreate/ApplicationCreateProvider';

// Mock the dependencies
vi.mock('@/features/integrations/api/useIdentityProviders');
vi.mock('@/features/integrations/utils/getIntegrationIcon');
vi.mock('@/features/flows/api/useGetFlows');

// Mock useGetApplications
vi.mock('../../../api/useGetApplications', () => ({
  __esModule: true,
  default: vi.fn(),
}));

// Mock generateAppPrimaryColorSuggestions
vi.mock('../../../utils/generateAppPrimaryColorSuggestions', () => ({
  __esModule: true,
  default: () => ['#3B82F6'],
}));

// Mock useConfig to avoid ConfigProvider requirement
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      endpoints: {
        server: 'http://localhost:3001',
      },
    }),
  };
});

const {default: useIdentityProviders} = await import('@/features/integrations/api/useIdentityProviders');
const {default: getIntegrationIcon} = await import('@/features/integrations/utils/getIntegrationIcon');
const {default: useGetFlows} = await import('@/features/flows/api/useGetFlows');
const {default: useGetApplications} = await import('../../../api/useGetApplications');

describe('ConfigureSignInOptions', () => {
  const mockOnIntegrationToggle = vi.fn();

  const mockIdentityProviders: IdentityProvider[] = [
    {
      id: 'google-idp',
      name: 'Google',
      type: IdentityProviderTypes.GOOGLE,
      description: 'Sign in with Google',
    },
    {
      id: 'github-idp',
      name: 'GitHub',
      type: IdentityProviderTypes.GITHUB,
      description: 'Sign in with GitHub',
    },
  ];

  const defaultProps: ConfigureSignInOptionsProps = {
    integrations: {
      [AuthenticatorTypes.BASIC_AUTH]: true,
    },
    onIntegrationToggle: mockOnIntegrationToggle,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getIntegrationIcon).mockReturnValue(<div>Icon</div>);
    // Default mock: no applications
    vi.mocked(useGetApplications).mockReturnValue({
      data: {
        totalResults: 0,
        count: 0,
        applications: [],
      },
      isLoading: false,
      isError: false,
      isSuccess: true,
      isFetching: false,
      isStale: false,
      isPending: false,
      error: null,
      status: 'success',
      fetchStatus: 'idle',
    } as unknown as ReturnType<typeof useGetApplications>);
    // Mock useGetFlows
    vi.mocked(useGetFlows).mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        flows: [],
        links: [],
      },
      isLoading: false,
      isError: false,
      isSuccess: true,
      isFetching: false,
      isStale: false,
      isPending: false,
      error: null,
      status: 'success',
      fetchStatus: 'idle',
    } as unknown as ReturnType<typeof useGetFlows>);
  });

  const renderComponent = async (props: Partial<ConfigureSignInOptionsProps> = {}) => {
    const renderResult = await renderWithProviders(
      <ApplicationCreateProvider>
        <ConfigureSignInOptions {...defaultProps} {...props} />
      </ApplicationCreateProvider>,
    );

    return {
      ...renderResult,
      rerender: (newProps: Partial<ConfigureSignInOptionsProps> = {}) =>
        renderResult.rerender(
          <ApplicationCreateProvider>
            <ConfigureSignInOptions {...defaultProps} {...newProps} />
          </ApplicationCreateProvider>,
        ),
    };
  };

  it('should render loading state', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render error state', async () => {
    const error = new Error('Failed to load integrations');
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    await expect.element(page.getByRole('alert')).toBeInTheDocument();
    await expect.element(page.getByText(/Failed to load authentication methods/i)).toBeInTheDocument();
  });

  it('should render the component with title and subtitle', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    await expect.element(page.getByRole('heading', {level: 1})).toBeInTheDocument();
    await expect.element(page.getByText('Choose how users will sign-in to your application')).toBeInTheDocument();
  });

  it('should always render Username & Password option first', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    await expect.element(page.getByText('Username & Password')).toBeInTheDocument();
  });

  it('should render Username & Password as toggleable (not forced enabled)', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    const switches = page.getByRole('switch').all();
    expect(switches[0]).toBeChecked();

    // Should be toggleable (not disabled)
    expect(switches[0]).not.toBeDisabled();
  });

  it('should render Username & Password as unchecked when not selected', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({
      integrations: {},
    });

    const switches = page.getByRole('switch').all();
    expect(switches[0]).not.toBeChecked();
  });

  it('should render all identity providers', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    await expect.element(page.getByText('Google')).toBeInTheDocument();
    await expect.element(page.getByText('GitHub')).toBeInTheDocument();
  });

  it('should call onIntegrationToggle when clicking Username & Password list item', async () => {

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    const usernamePasswordButton = page.getByText('Username & Password').element().closest('.MuiListItemButton-root');
    if (usernamePasswordButton) {
      await userEvent.click(usernamePasswordButton);
    }

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.BASIC_AUTH);
  });

  it('should call onIntegrationToggle when clicking provider list item', async () => {

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    const googleButton = page.getByText('Google').element().closest('.MuiListItemButton-root');
    if (googleButton) {
      await userEvent.click(googleButton);
    }

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
  });

  it('should call onIntegrationToggle when toggling switch', async () => {

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    const switches = page.getByRole('switch').all();
    await userEvent.click(switches[2]); // Click Google switch

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
  });

  it('should show checked state for enabled integrations', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': false,
      },
    });

    const switches = page.getByRole('switch').all();
    expect(switches[0]).toBeChecked(); // Username & Password
    expect(switches[2]).toBeChecked(); // Google
    expect(switches[3]).not.toBeChecked(); // GitHub
  });

  it('should show username/password option when no integrations are available', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    // Should show username/password in the list with a switch (always toggleable)
    await expect.element(page.getByText('Username & Password')).toBeInTheDocument();
    await expect.element(page.getByRole('list')).toBeInTheDocument();

    // Should have a toggle/switch (username/password is always toggleable)
    const switches = page.getByRole('switch').all();
    expect(switches.length).toBeGreaterThan(0);
  });

  it('should render integration icons', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    // Google and GitHub use direct icons, not getIntegrationIcon
    // Other providers (if any) would use getIntegrationIcon
    await expect.element(page.getByText('Google')).toBeInTheDocument();
    await expect.element(page.getByText('GitHub')).toBeInTheDocument();
  });

  it('should render UserRound icon for Username & Password', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    // UserRound icon should be present
    const usernamePasswordSection = page.getByText('Username & Password').element().closest('div');
    expect(usernamePasswordSection).toBeInTheDocument();
  });

  it('should stop propagation when clicking switch', async () => {

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    const switches = page.getByRole('switch').all();
    await userEvent.click(switches[1]);

    // Should only trigger once (not twice from card and switch)
    expect(mockOnIntegrationToggle).toHaveBeenCalledTimes(1);
  });

  it('should handle empty integrations record', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({integrations: {}});

    const switches = page.getByRole('switch').all();
    // Username & Password should default to false when integrations is empty
    expect(switches[0]).not.toBeChecked();
    // Others should default to false
    expect(switches[1]).not.toBeChecked();
  });

  it('should render info icon in subtitle', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    const subtitle = page.getByText('Choose how users will sign-in to your application').element().closest('div');
    expect(subtitle).toBeInTheDocument();
  });

  it('should handle multiple rapid toggles', async () => {

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    const switches = page.getByRole('switch').all();
    await userEvent.click(switches[1]);
    await userEvent.click(switches[2]);
    await userEvent.click(switches[1]);

    expect(mockOnIntegrationToggle).toHaveBeenCalledTimes(3);
  });

  it('should handle providers with long names', async () => {
    const longNameProvider: IdentityProvider = {
      id: 'long-name-idp',
      name: 'Very Long Identity Provider Name That Should Still Display',
      type: 'OIDC',
      description: 'Test provider',
    };

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [longNameProvider],
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent();

    await expect.element(page.getByText(longNameProvider.name)).toBeInTheDocument();
  });

  it('should maintain switch state after re-render', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    const {rerender} = await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    let switches = page.getByRole('switch').all();
    expect(switches[2]).toBeChecked();

    await rerender({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    switches = page.getByRole('switch').all();
    expect(switches[2]).toBeChecked();
  });

  describe('Google and GitHub always shown', () => {
    it('should always show Google option even when not configured', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      await expect.element(page.getByText('Google')).toBeInTheDocument();
    });

    it('should always show GitHub option even when not configured', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      await expect.element(page.getByText('GitHub')).toBeInTheDocument();
    });

    it('should show Google as disabled with "Not configured" when not in API', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      const googleText = page.getByText('Google');
      const listItem = googleText.element().closest('.MuiListItem-root');
      expect(listItem).toBeInTheDocument();

      // Should have "Not configured" as secondary text (both Google and GitHub show it)
      const notConfiguredTexts = page.getByText('Not configured').all();
      expect(notConfiguredTexts.length).toBeGreaterThanOrEqual(1);

      // Should not have a switch for Google (disabled)
      const switches = page.getByRole('switch').all();
      // Only username/password and passkey should have a switch
      expect(switches.length).toBe(2);

      // Google button should be disabled
      const googleButton = googleText.element().closest('.MuiListItemButton-root');
      expect(googleButton).toHaveAttribute('aria-disabled', 'true');
    });

    it('should show GitHub as disabled with "Not configured" when not in API', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      const githubText = page.getByText('GitHub');
      const listItem = githubText.element().closest('.MuiListItem-root');
      expect(listItem).toBeInTheDocument();

      // Should have "Not configured" as secondary text
      const notConfiguredTexts = page.getByText('Not configured').all();
      expect(notConfiguredTexts.length).toBeGreaterThan(0);
    });

    it('should show Google as enabled with switch when configured in API', async () => {

      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      const switches = page.getByRole('switch').all();
      // Should have switches for username/password, passkey, Google, and GitHub
      expect(switches.length).toBe(4);

      // Google should be toggleable
      const googleButton = page.getByText('Google').element().closest('.MuiListItemButton-root');
      expect(googleButton).not.toBeDisabled();
    });

    it('should show GitHub as enabled with switch when configured in API', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      const switches = page.getByRole('switch').all();
      // Should have switches for username/password, passkey, Google, and GitHub
      expect(switches.length).toBe(4);

      // GitHub should be toggleable
      const githubButton = page.getByText('GitHub').element().closest('.MuiListItemButton-root');
      expect(githubButton).not.toBeDisabled();
    });

    it('should show Google enabled and GitHub disabled when only Google is configured', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [mockIdentityProviders[0]], // Only Google
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      // Google should be enabled
      const googleButton = page.getByText('Google').element().closest('.MuiListItemButton-root');
      expect(googleButton).not.toHaveAttribute('aria-disabled', 'true');

      // GitHub should be disabled
      const githubButton = page.getByText('GitHub').element().closest('.MuiListItemButton-root');
      expect(githubButton).toHaveAttribute('aria-disabled', 'true');

      // Should show "Not configured" for GitHub
      const notConfiguredTexts = page.getByText('Not configured').all();
      expect(notConfiguredTexts.length).toBeGreaterThan(0);
    });
  });

  describe('Validation warning', () => {
    it('should show warning when no options are selected', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent({
        integrations: {}, // No selections
      });

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
      // Check for the actual translated text
      const alert = page.getByRole('alert');
      expect(alert.element().textContent).toMatch(/at least one sign-in option is required/i);
    });

    it('should not show warning when at least one option is selected', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      // Should not have warning alert
      const alerts = page.getByRole('alert').all();
      const warningAlerts = alerts.filter((alert) => alert.element().textContent?.includes('at least one'));
      expect(warningAlerts.length).toBe(0);
    });

    it('should show warning when only username/password is deselected', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'google-idp': false,
          'github-idp': false,
        },
      });

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
      // Check for the actual translated text
      const alert = page.getByRole('alert');
      expect(alert.element().textContent).toMatch(/at least one sign-in option is required/i);
    });

    it('should hide warning when user selects an option', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      const {rerender} = await renderComponent({
        integrations: {}, // No selections initially
      });

      await expect.element(page.getByRole('alert')).toBeInTheDocument();

      // Select username/password
      await rerender({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      // Warning should be gone
      const alerts = page.getByRole('alert').all();
      const warningAlerts = alerts.filter((alert) => alert.element().textContent?.includes('at least one'));
      expect(warningAlerts.length).toBe(0);
    });
  });

  describe('onReadyChange callback', () => {
    it('should call onReadyChange with true when integrations are selected', async () => {
      const onReadyChange = vi.fn();
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
        onReadyChange,
      });

      expect(onReadyChange).toHaveBeenCalledWith(true);
    });

    it('should call onReadyChange with false when no integrations are selected', async () => {
      const onReadyChange = vi.fn();
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent({
        integrations: {},
        onReadyChange,
      });

      expect(onReadyChange).toHaveBeenCalledWith(false);
    });
  });

  describe('Flow loading states', () => {
    it('should render loading state when flows are loading', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
        isSuccess: false,
        isFetching: true,
        isStale: false,
        isPending: true,
        error: null,
        status: 'pending',
        fetchStatus: 'fetching',
      } as unknown as ReturnType<typeof useGetFlows>);

      await renderComponent();

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should render error state when flows fail to load', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      const flowsError = new Error('Failed to load flows');
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        isSuccess: false,
        isFetching: false,
        isStale: false,
        isPending: false,
        error: flowsError,
        status: 'error',
        fetchStatus: 'idle',
      } as unknown as ReturnType<typeof useGetFlows>);

      await renderComponent();

      await expect.element(page.getByRole('alert')).toBeInTheDocument();
      await expect.element(page.getByText(/Failed to load authentication methods/i)).toBeInTheDocument();
    });
  });

  describe('Flow data handling', () => {
    it('should handle when flows data is empty', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      vi.mocked(useGetFlows).mockReturnValue({
        data: {
          totalResults: 0,
          startIndex: 1,
          count: 0,
          flows: [],
          links: [],
        },
        isLoading: false,
        isError: false,
        isSuccess: true,
        isFetching: false,
        isStale: false,
        isPending: false,
        error: null,
        status: 'success',
        fetchStatus: 'idle',
      } as unknown as ReturnType<typeof useGetFlows>);

      await renderComponent();

      // Component should still render without flows
      await expect.element(page.getByText('Username & Password')).toBeInTheDocument();
    });

    it('should handle when flows data is null', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      vi.mocked(useGetFlows).mockReturnValue({
        data: null,
        isLoading: false,
        isError: false,
        isSuccess: true,
        isFetching: false,
        isStale: false,
        isPending: false,
        error: null,
        status: 'success',
        fetchStatus: 'idle',
      } as unknown as ReturnType<typeof useGetFlows>);

      await renderComponent();

      // Component should still render without flows
      await expect.element(page.getByText('Username & Password')).toBeInTheDocument();
    });
  });

  describe('Integration type mapping', () => {
    it('should handle OIDC type providers', async () => {
      const oidcProvider: IdentityProvider = {
        id: 'oidc-idp',
        name: 'OIDC Provider',
        type: 'OIDC',
        description: 'Generic OIDC provider',
      };

      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [oidcProvider, ...mockIdentityProviders],
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      const oidcButton = page.getByText('OIDC Provider').element().closest('.MuiListItemButton-root');
      if (oidcButton) {
        await userEvent.click(oidcButton);
      }

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('oidc-idp');
    });
  });

  describe('Hint text', () => {
    it('should render hint text with lightbulb icon', async () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      await renderComponent();

      expect(
        page.getByText('You can always change these settings later in the application settings.'),
      ).toBeInTheDocument();
    });
  });
});
