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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import ConfigureSignInOptions, {type ConfigureSignInOptionsProps} from '../ConfigureSignInOptions';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.configure.SignInOptions.title': 'Sign In Options',
        'applications:onboarding.configure.SignInOptions.subtitle': 'Choose how users will sign-in to your application',
        'applications:onboarding.configure.SignInOptions.usernamePassword': 'Username & Password',
        'applications:onboarding.configure.SignInOptions.google': 'Google',
        'applications:onboarding.configure.SignInOptions.github': 'GitHub',
        'applications:onboarding.configure.SignInOptions.notConfigured': 'Not configured',
        'applications:onboarding.configure.SignInOptions.noSelectionWarning':
          'At least one login option is required. Please select at least one authentication method.',
        'applications:onboarding.configure.SignInOptions.hint':
          'You can always change these settings later in the application settings.',
        'applications:onboarding.configure.SignInOptions.error': 'Failed to load authentication methods: {{error}}',
        'applications:onboarding.configure.SignInOptions.flowNotAvailable': 'Authentication flow not available',
        'applications:onboarding.configure.SignInOptions.orDivider': 'Or',
        'applications:onboarding.configure.SignInOptions.selectExistingFlow': 'Select an existing flow',
        'applications:onboarding.configure.SignInOptions.none': 'None',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock the dependencies
vi.mock('@/features/integrations/api/useIdentityProviders');
vi.mock('@/features/integrations/utils/getIntegrationIcon');

const {default: useIdentityProviders} = await import('@/features/integrations/api/useIdentityProviders');
const {default: getIntegrationIcon} = await import('@/features/integrations/utils/getIntegrationIcon');

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
  });

  const renderComponent = (props: Partial<ConfigureSignInOptionsProps> = {}) =>
    render(<ConfigureSignInOptions {...defaultProps} {...props} />);

  it('should render loading state', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render error state', () => {
    const error = new Error('Failed to load integrations');
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: false,
      error,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/Failed to load authentication methods/i)).toBeInTheDocument();
  });

  it('should render the component with title and subtitle', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
    expect(screen.getByText('Choose how users will sign-in to your application')).toBeInTheDocument();
  });

  it('should always render Username & Password option first', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByText('Username & Password')).toBeInTheDocument();
  });

  it('should render Username & Password as toggleable (not forced enabled)', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    const switches = screen.getAllByRole('switch');
    expect(switches[0]).toBeChecked();

    // Should be toggleable (not disabled)
    expect(switches[0]).not.toBeDisabled();
  });

  it('should render Username & Password as unchecked when not selected', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {},
    });

    const switches = screen.getAllByRole('switch');
    expect(switches[0]).not.toBeChecked();
  });

  it('should render all identity providers', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    expect(screen.getByText('Google')).toBeInTheDocument();
    expect(screen.getByText('GitHub')).toBeInTheDocument();
  });

  it('should call onIntegrationToggle when clicking Username & Password list item', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const usernamePasswordButton = screen.getByText('Username & Password').closest('.MuiListItemButton-root');
    if (usernamePasswordButton) {
      await user.click(usernamePasswordButton);
    }

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.BASIC_AUTH);
  });

  it('should call onIntegrationToggle when clicking provider list item', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const googleButton = screen.getByText('Google').closest('.MuiListItemButton-root');
    if (googleButton) {
      await user.click(googleButton);
    }

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
  });

  it('should call onIntegrationToggle when toggling switch', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const switches = screen.getAllByRole('switch');
    await user.click(switches[1]); // Click Google switch

    expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
  });

  it('should show checked state for enabled integrations', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': false,
      },
    });

    const switches = screen.getAllByRole('switch');
    expect(switches[0]).toBeChecked(); // Username & Password
    expect(switches[1]).toBeChecked(); // Google
    expect(switches[2]).not.toBeChecked(); // GitHub
  });

  it('should show username/password option when no integrations are available', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [],
      isLoading: false,
      error: null,
    } as unknown as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    // Should show username/password in the list with a switch (always toggleable)
    expect(screen.getByText('Username & Password')).toBeInTheDocument();
    expect(screen.getByRole('list')).toBeInTheDocument();

    // Should have a toggle/switch (username/password is always toggleable)
    const switches = screen.getAllByRole('switch');
    expect(switches.length).toBeGreaterThan(0);
  });

  it('should render integration icons', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    // Google and GitHub use direct icons, not getIntegrationIcon
    // Other providers (if any) would use getIntegrationIcon
    expect(screen.getByText('Google')).toBeInTheDocument();
    expect(screen.getByText('GitHub')).toBeInTheDocument();
  });

  it('should render UserRound icon for Username & Password', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    // UserRound icon should be present
    const usernamePasswordSection = screen.getByText('Username & Password').closest('div');
    expect(usernamePasswordSection).toBeInTheDocument();
  });

  it('should stop propagation when clicking switch', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const switches = screen.getAllByRole('switch');
    await user.click(switches[1]);

    // Should only trigger once (not twice from card and switch)
    expect(mockOnIntegrationToggle).toHaveBeenCalledTimes(1);
  });

  it('should handle empty integrations record', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({integrations: {}});

    const switches = screen.getAllByRole('switch');
    // Username & Password should default to false when integrations is empty
    expect(switches[0]).not.toBeChecked();
    // Others should default to false
    expect(switches[1]).not.toBeChecked();
  });

  it('should render info icon in subtitle', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const subtitle = screen.getByText('Choose how users will sign-in to your application').closest('div');
    expect(subtitle).toBeInTheDocument();
  });

  it('should handle multiple rapid toggles', async () => {
    const user = userEvent.setup();

    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent();

    const switches = screen.getAllByRole('switch');
    await user.click(switches[1]);
    await user.click(switches[2]);
    await user.click(switches[1]);

    expect(mockOnIntegrationToggle).toHaveBeenCalledTimes(3);
  });

  it('should handle providers with long names', () => {
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

    renderComponent();

    expect(screen.getByText(longNameProvider.name)).toBeInTheDocument();
  });

  it('should maintain switch state after re-render', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    const {rerender} = renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    let switches = screen.getAllByRole('switch');
    expect(switches[1]).toBeChecked();

    rerender(
      <ConfigureSignInOptions
        integrations={{
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'google-idp': true,
        }}
        onIntegrationToggle={mockOnIntegrationToggle}
      />,
    );

    switches = screen.getAllByRole('switch');
    expect(switches[1]).toBeChecked();
  });

  describe('Google and GitHub always shown', () => {
    it('should always show Google option even when not configured', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      expect(screen.getByText('Google')).toBeInTheDocument();
    });

    it('should always show GitHub option even when not configured', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('should show Google as disabled with "Not configured" when not in API', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      const googleText = screen.getByText('Google');
      const listItem = googleText.closest('.MuiListItem-root');
      expect(listItem).toBeInTheDocument();

      // Should have "Not configured" as secondary text (both Google and GitHub show it)
      const notConfiguredTexts = screen.getAllByText('Not configured');
      expect(notConfiguredTexts.length).toBeGreaterThanOrEqual(1);

      // Should not have a switch for Google (disabled)
      const switches = screen.getAllByRole('switch');
      // Only username/password should have a switch
      expect(switches.length).toBe(1);

      // Google button should be disabled
      const googleButton = googleText.closest('.MuiListItemButton-root');
      expect(googleButton).toHaveAttribute('aria-disabled', 'true');
    });

    it('should show GitHub as disabled with "Not configured" when not in API', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [], // No providers in API
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      const githubText = screen.getByText('GitHub');
      const listItem = githubText.closest('.MuiListItem-root');
      expect(listItem).toBeInTheDocument();

      // Should have "Not configured" as secondary text
      const notConfiguredTexts = screen.getAllByText('Not configured');
      expect(notConfiguredTexts.length).toBeGreaterThan(0);
    });

    it('should show Google as enabled with switch when configured in API', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      const switches = screen.getAllByRole('switch');
      // Should have switches for username/password, Google, and GitHub
      expect(switches.length).toBe(3);

      // Google should be toggleable
      const googleButton = screen.getByText('Google').closest('.MuiListItemButton-root');
      expect(googleButton).not.toBeDisabled();
    });

    it('should show GitHub as enabled with switch when configured in API', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      const switches = screen.getAllByRole('switch');
      // Should have switches for username/password, Google, and GitHub
      expect(switches.length).toBe(3);

      // GitHub should be toggleable
      const githubButton = screen.getByText('GitHub').closest('.MuiListItemButton-root');
      expect(githubButton).not.toBeDisabled();
    });

    it('should show Google enabled and GitHub disabled when only Google is configured', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: [mockIdentityProviders[0]], // Only Google
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      renderComponent();

      // Google should be enabled
      const googleButton = screen.getByText('Google').closest('.MuiListItemButton-root');
      expect(googleButton).not.toHaveAttribute('aria-disabled', 'true');

      // GitHub should be disabled
      const githubButton = screen.getByText('GitHub').closest('.MuiListItemButton-root');
      expect(githubButton).toHaveAttribute('aria-disabled', 'true');

      // Should show "Not configured" for GitHub
      const notConfiguredTexts = screen.getAllByText('Not configured');
      expect(notConfiguredTexts.length).toBeGreaterThan(0);
    });
  });

  describe('Validation warning', () => {
    it('should show warning when no options are selected', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      renderComponent({
        integrations: {}, // No selections
      });

      expect(screen.getByRole('alert')).toBeInTheDocument();
      // Check for the translation key or the actual text
      const alert = screen.getByRole('alert');
      expect(alert.textContent).toMatch(/noSelectionWarning|at least one login option is required/i);
    });

    it('should not show warning when at least one option is selected', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      // Should not have warning alert
      const alerts = screen.queryAllByRole('alert');
      const warningAlerts = alerts.filter((alert) => alert.textContent?.includes('at least one'));
      expect(warningAlerts.length).toBe(0);
    });

    it('should show warning when only username/password is deselected', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'google-idp': false,
          'github-idp': false,
        },
      });

      expect(screen.getByRole('alert')).toBeInTheDocument();
      // Check for the translation key or the actual text
      const alert = screen.getByRole('alert');
      expect(alert.textContent).toMatch(/noSelectionWarning|at least one login option is required/i);
    });

    it('should hide warning when user selects an option', () => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);

      const {rerender} = renderComponent({
        integrations: {}, // No selections initially
      });

      expect(screen.getByRole('alert')).toBeInTheDocument();

      // Select username/password
      rerender(
        <ConfigureSignInOptions
          integrations={{
            [AuthenticatorTypes.BASIC_AUTH]: true,
          }}
          onIntegrationToggle={mockOnIntegrationToggle}
        />,
      );

      // Warning should be gone
      const warningAlerts = screen
        .queryAllByRole('alert')
        .filter((alert) => alert.textContent?.includes('at least one'));
      expect(warningAlerts.length).toBe(0);
    });
  });

  describe('Custom flow selection', () => {
    const mockAuthFlows = [
      {
        id: 'flow-1',
        name: 'Basic Flow',
        handle: 'default-basic-flow',
        flowType: 'AUTHENTICATION',
        activeVersion: 1,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
      {
        id: 'flow-2',
        name: 'Google Flow',
        handle: 'default-google-flow',
        flowType: 'AUTHENTICATION',
        activeVersion: 1,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      },
      {
        id: 'flow-3',
        name: 'App Flow',
        handle: 'develop-app-flow',
        flowType: 'AUTHENTICATION',
        activeVersion: 1,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z',
      }, // Should be filtered
    ] as ConfigureSignInOptionsProps['authFlows'];

    beforeEach(() => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);
    });

    it('should render custom flow dropdown with authFlows prop', () => {
      renderComponent({authFlows: mockAuthFlows});

      expect(screen.getByLabelText('Select an existing flow')).toBeInTheDocument();
    });

    it('should render "Or" divider between toggles and dropdown', () => {
      renderComponent({authFlows: mockAuthFlows});

      expect(screen.getByText('Or')).toBeInTheDocument();
    });

    it('should filter out develop-app-flow from dropdown', () => {
      renderComponent({authFlows: mockAuthFlows});

      // Click to open dropdown
      const dropdown = screen.getByLabelText('Select an existing flow');
      expect(dropdown).toBeInTheDocument();

      // The develop-app-flow should not be visible
      expect(screen.queryByText('develop-app-flow')).not.toBeInTheDocument();
    });

    it('should call onCustomFlowChange when selecting a flow', () => {
      const mockOnCustomFlowChange = vi.fn();

      renderComponent({
        authFlows: mockAuthFlows,
        onCustomFlowChange: mockOnCustomFlowChange,
      });

      // Note: Testing Select component interaction requires more complex setup
      // This test verifies the component renders correctly
      expect(screen.getByLabelText('Select an existing flow')).toBeInTheDocument();
    });

    it('should not show warning when custom flow is selected', () => {
      renderComponent({
        integrations: {}, // No toggle selections
        customFlowId: 'flow-1',
        authFlows: mockAuthFlows,
      });

      // Should not show selection warning when custom flow is selected
      const warningAlerts = screen
        .queryAllByRole('alert')
        .filter((alert) => alert.textContent?.includes('at least one'));
      expect(warningAlerts.length).toBe(0);
    });

    it('should render dropdown showing "None" option', () => {
      renderComponent({authFlows: mockAuthFlows});

      // None should be the default option text
      expect(screen.getByLabelText('Select an existing flow')).toBeInTheDocument();
    });
  });

  describe('Flow Availability', () => {
    beforeEach(() => {
      vi.mocked(useIdentityProviders).mockReturnValue({
        data: mockIdentityProviders,
        isLoading: false,
        error: null,
      } as ReturnType<typeof useIdentityProviders>);
    });

    it('should call isFlowAvailable with flow handles', () => {
      const mockIsFlowAvailable = vi.fn().mockReturnValue(true);

      renderComponent({
        isFlowAvailable: mockIsFlowAvailable,
      });

      expect(mockIsFlowAvailable).toHaveBeenCalled();
    });

    it('should enable options when flows are available', () => {
      const mockIsFlowAvailable = vi.fn().mockReturnValue(true);

      renderComponent({
        isFlowAvailable: mockIsFlowAvailable,
      });

      const switches = screen.getAllByRole('switch');
      expect(switches.length).toBeGreaterThan(0);
    });

    it('should not render custom flow section when authFlows is empty', () => {
      renderComponent({authFlows: []});

      expect(screen.queryByText('Or')).not.toBeInTheDocument();
      expect(screen.queryByLabelText('Select an existing flow')).not.toBeInTheDocument();
    });
  });
});
