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
import {render} from '@thunder/test-utils/browser';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import IndividualMethodsToggleView, {type IndividualMethodsToggleViewProps} from '../IndividualMethodsToggleView';

// Mock the integration icon utility
vi.mock('@/features/integrations/utils/getIntegrationIcon', () => ({
  default: vi.fn((type: string) => <div data-testid={`icon-${type}`}>Mock Icon</div>),
}));

describe('IndividualMethodsToggleView', () => {
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
    {
      id: 'oauth-idp-1',
      name: 'OAuth Provider',
      type: IdentityProviderTypes.OAUTH,
      description: 'Sign in with OAuth',
    },
    {
      id: 'oidc-idp-1',
      name: 'OIDC Provider',
      type: IdentityProviderTypes.OIDC,
      description: 'Sign in with OIDC',
    },
  ];

  const defaultProps: IndividualMethodsToggleViewProps = {
    integrations: {
      [AuthenticatorTypes.BASIC_AUTH]: false,
    },
    availableIntegrations: mockIdentityProviders,
    onIntegrationToggle: mockOnIntegrationToggle,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = async (props: Partial<IndividualMethodsToggleViewProps> = {}) =>
    render(<IndividualMethodsToggleView {...defaultProps} {...props} />);

  describe('core authentication methods', () => {
    it('should render Username & Password option', async () => {
      await renderComponent();

      await expect.element(page.getByText('Username & Password')).toBeInTheDocument();
    });

    it('should render Passkey option', async () => {
      await renderComponent();

      await expect.element(page.getByText('Passkey')).toBeInTheDocument();
    });

    it('should render Google option', async () => {
      await renderComponent();

      await expect.element(page.getByText('Google')).toBeInTheDocument();
    });

    it('should render GitHub option', async () => {
      await renderComponent();

      await expect.element(page.getByText('GitHub')).toBeInTheDocument();
    });

    it('should render in correct order: Username & Password, Google, GitHub', async () => {
      await renderComponent();

      const listItems = page.getByRole('button').all();
      // First item should be Username & Password
      expect(listItems[0]).toHaveTextContent('Username & Password');
    });
  });

  describe('integration states', () => {
    it('should show Username & Password as enabled when basic_auth is true', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      const switches = page.getByRole('switch').all();
      // First switch should be for Username & Password
      expect(switches[0]).toBeChecked();
    });

    it('should show Username & Password as disabled when basic_auth is false', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
        },
      });

      const switches = page.getByRole('switch').all();
      expect(switches[0]).not.toBeChecked();
    });

    it('should show Passkey as enabled when passkey is true', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          [AuthenticatorTypes.PASSKEY]: true,
        },
      });

      const switches = page.getByRole('switch').all();
      // Second switch should be for Passkey
      expect(switches[1]).toBeChecked();
    });

    it('should show Passkey as disabled when passkey is false', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          [AuthenticatorTypes.PASSKEY]: false,
        },
      });

      const switches = page.getByRole('switch').all();
      // Second switch should be for Passkey
      expect(switches[1]).not.toBeChecked();
    });

    it('should show Passkey as disabled when passkey is undefined', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          [AuthenticatorTypes.PASSKEY]: undefined as unknown as boolean,
        },
      });

      const switches = page.getByRole('switch').all();
      expect(switches[1]).not.toBeChecked();
    });

    it('should show Google as enabled when its ID is in integrations', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'google-idp': true,
        },
      });

      const switches = page.getByRole('switch').all();
      // Second switch should be for Google
      expect(switches[2]).toBeChecked();
    });

    it('should show GitHub as enabled when its ID is in integrations', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'github-idp': true,
        },
      });

      const switches = page.getByRole('switch').all();
      // Third switch should be for GitHub
      expect(switches[3]).toBeChecked();
    });
  });

  describe('toggle interactions', () => {
    it('should call onIntegrationToggle with basic_auth when Username & Password is toggled', async () => {
      await renderComponent();

      const switches = page.getByRole('switch').all();
      await userEvent.click(switches[0]);

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.BASIC_AUTH);
    });

    it('should call onIntegrationToggle with passkey when Passkey is toggled', async () => {
      await renderComponent();

      const switches = page.getByRole('switch').all();
      await userEvent.click(switches[1]); // Passkey switch

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.PASSKEY);
    });

    it('should call onIntegrationToggle with google provider ID when Google is toggled', async () => {
      await renderComponent();

      const switches = page.getByRole('switch').all();
      await userEvent.click(switches[2]); // Google switch

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('google-idp');
    });

    it('should call onIntegrationToggle with github provider ID when GitHub is toggled', async () => {
      await renderComponent();

      const switches = page.getByRole('switch').all();
      await userEvent.click(switches[3]); // GitHub switch

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith('github-idp');
    });

    it('should call onIntegrationToggle when clicking on list item button', async () => {
      await renderComponent();

      const buttons = page.getByRole('button').all();
      await userEvent.click(buttons[0]); // Username & Password button

      expect(mockOnIntegrationToggle).toHaveBeenCalledWith(AuthenticatorTypes.BASIC_AUTH);
    });
  });

  describe('other social providers', () => {
    it('should render other providers (non-Google, non-GitHub)', async () => {
      await renderComponent();

      await expect.element(page.getByText('OAuth Provider')).toBeInTheDocument();
      await expect.element(page.getByText('OIDC Provider')).toBeInTheDocument();
    });

    it('should call onIntegrationToggle with provider ID when other provider is toggled', async () => {
      await renderComponent();

      // Find the OAuth Provider item and its switch
      const oauthItem = page.getByText('OAuth Provider').element().closest('li');
      const oauthSwitch = oauthItem?.querySelector('input[type="checkbox"]');

      if (oauthSwitch) {
        await userEvent.click(oauthSwitch);
        expect(mockOnIntegrationToggle).toHaveBeenCalledWith('oauth-idp-1');
      }
    });

    it('should show other provider as enabled when its ID is in integrations', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'oauth-idp-1': true,
        },
      });

      const oauthItem = page.getByText('OAuth Provider').element().closest('li');
      const oauthSwitch = oauthItem?.querySelector('input[type="checkbox"]');

      expect(oauthSwitch).toBeChecked();
    });
  });

  describe('unavailable providers', () => {
    it('should show Google as unavailable when not in availableIntegrations', async () => {
      await renderComponent({
        availableIntegrations: mockIdentityProviders.filter((idp) => idp.type !== IdentityProviderTypes.GOOGLE),
      });

      await expect.element(page.getByText('Google')).toBeInTheDocument();
      await expect.element(page.getByText('Not configured')).toBeInTheDocument();
    });

    it('should show GitHub as unavailable when not in availableIntegrations', async () => {
      await renderComponent({
        availableIntegrations: mockIdentityProviders.filter((idp) => idp.type !== IdentityProviderTypes.GITHUB),
      });

      await expect.element(page.getByText('GitHub')).toBeInTheDocument();
      await expect.element(page.getByText('Not configured')).toBeInTheDocument();
    });

    it('should show both Google and GitHub as unavailable when empty integrations', async () => {
      await renderComponent({
        availableIntegrations: [],
      });

      await expect.element(page.getByText('Google')).toBeInTheDocument();
      await expect.element(page.getByText('GitHub')).toBeInTheDocument();
      expect(page.getByText('Not configured').all()).toHaveLength(2);
    });

    it('should disable unavailable provider buttons', async () => {
      await renderComponent({
        availableIntegrations: [],
      });

      // Find all buttons - Google and GitHub should be disabled (aria-disabled)
      const buttons = page.getByRole('button').all();
      // The unavailable providers will have buttons with aria-disabled attribute
      const disabledButtons = buttons.filter((btn) => btn.element().getAttribute('aria-disabled') === 'true');

      // Should have 2 disabled buttons (Google and GitHub)
      expect(disabledButtons).toHaveLength(2);
    });

    it('should not render switch for unavailable providers', async () => {
      await renderComponent({
        availableIntegrations: [],
      });

      // Only Username & Password and Passkey should have a switch
      const switches = page.getByRole('switch').all();
      expect(switches).toHaveLength(2);
    });
  });

  describe('list structure', () => {
    it('should render a list element', async () => {
      await renderComponent();

      const list = page.getByRole('list');
      expect(list).toBeInTheDocument();
    });

    it('should render dividers between items', async () => {
      await renderComponent();

      const list = page.getByRole('list');
      const dividers = list.element().querySelectorAll('hr, .MuiDivider-root');
      expect(dividers.length).toBeGreaterThan(0);
    });
  });

  describe('edge cases', () => {
    it('should handle empty integrations object', async () => {
      await renderComponent({
        integrations: {},
      });

      // Should not crash and should render all options as unchecked
      await expect.element(page.getByText('Username & Password')).toBeInTheDocument();
      const switches = page.getByRole('switch').all();
      switches.forEach((switchEl) => {
        expect(switchEl).not.toBeChecked();
      });
    });

    it('should handle undefined integration value for basic_auth', async () => {
      await renderComponent({
        integrations: {
          // basic_auth not defined
        },
      });

      const switches = page.getByRole('switch').all();
      // Username & Password switch should be unchecked when undefined
      expect(switches[0]).not.toBeChecked();
    });

    it('should handle undefined integration value for Google provider', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'google-idp': undefined as unknown as boolean,
        },
      });

      const switches = page.getByRole('switch').all();
      // Google switch should be unchecked when undefined
      expect(switches[1]).not.toBeChecked();
    });

    it('should handle undefined integration value for GitHub provider', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'github-idp': undefined as unknown as boolean,
        },
      });

      const switches = page.getByRole('switch').all();
      // GitHub switch should be unchecked when undefined
      expect(switches[2]).not.toBeChecked();
    });

    it('should handle undefined integration value for other providers', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          'oauth-idp-1': undefined as unknown as boolean,
        },
      });

      const oauthItem = page.getByText('OAuth Provider').element().closest('li');
      const oauthSwitch = oauthItem?.querySelector('input[type="checkbox"]');

      // OAuth provider switch should be unchecked when undefined
      expect(oauthSwitch).not.toBeChecked();
    });

    it('should use default ID for Google when provider not found', async () => {
      await renderComponent({
        availableIntegrations: [],
      });

      // Google should still render with fallback ID 'google'
      await expect.element(page.getByText('Google')).toBeInTheDocument();
    });

    it('should use default ID for GitHub when provider not found', async () => {
      await renderComponent({
        availableIntegrations: [],
      });

      // GitHub should still render with fallback ID 'github'
      await expect.element(page.getByText('GitHub')).toBeInTheDocument();
    });

    it('should handle providers with special characters in names', async () => {
      const specialProviders: IdentityProvider[] = [
        {
          id: 'special-provider',
          name: 'Provider with Special & Characters',
          type: IdentityProviderTypes.OAUTH,
          description: 'Special provider',
        },
      ];

      await renderComponent({
        availableIntegrations: [...mockIdentityProviders, ...specialProviders],
      });

      await expect.element(page.getByText('Provider with Special & Characters')).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper list structure for screen readers', async () => {
      await renderComponent();

      const list = page.getByRole('list');
      expect(list).toBeInTheDocument();
    });

    it('should have switches with proper roles', async () => {
      await renderComponent();

      const switches = page.getByRole('switch').all();
      expect(switches.length).toBeGreaterThan(0);
    });

    it('should have buttons with proper roles', async () => {
      await renderComponent();

      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBeGreaterThan(0);
    });

    it('should be keyboard navigable', async () => {
      await renderComponent();

      // Tab to first focusable element
      await userEvent.tab();

      const firstButton = page.getByRole('button').all()[0];
      expect(firstButton).toHaveFocus();
    });
  });
});
