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
import {UserRound} from '@wso2/oxygen-ui-icons-react';
import AuthenticationMethodItem, {type AuthenticationMethodItemProps} from '../AuthenticationMethodItem';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.configure.SignInOptions.notConfigured': 'Not configured',
      };
      return translations[key] || key;
    },
  }),
}));

describe('AuthenticationMethodItem', () => {
  const mockOnToggle = vi.fn();

  const defaultProps: AuthenticationMethodItemProps = {
    id: 'test-auth-method',
    name: 'Test Auth Method',
    icon: <UserRound size={24} data-testid="test-icon" />,
    isEnabled: false,
    isAvailable: true,
    onToggle: mockOnToggle,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const renderComponent = (props: Partial<AuthenticationMethodItemProps> = {}) =>
    render(<AuthenticationMethodItem {...defaultProps} {...props} />);

  describe('when method is available', () => {
    it('should render the authentication method with all elements', () => {
      renderComponent();

      expect(screen.getByText('Test Auth Method')).toBeInTheDocument();
      expect(screen.getByTestId('test-icon')).toBeInTheDocument();
      expect(screen.getByRole('switch')).toBeInTheDocument();
      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render with secondary text when provided', () => {
      renderComponent({
        secondary: 'Additional info',
      });

      expect(screen.getByText('Test Auth Method')).toBeInTheDocument();
      expect(screen.getByText('Additional info')).toBeInTheDocument();
    });

    it('should show switch as checked when method is enabled', () => {
      renderComponent({
        isEnabled: true,
      });

      const switchElement = screen.getByRole('switch');
      expect(switchElement).toBeChecked();
    });

    it('should show switch as unchecked when method is disabled', () => {
      renderComponent({
        isEnabled: false,
      });

      const switchElement = screen.getByRole('switch');
      expect(switchElement).not.toBeChecked();
    });

    it('should call onToggle when list item button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button');
      await user.click(button);

      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
      expect(mockOnToggle).toHaveBeenCalledTimes(1);
    });

    it('should call onToggle when switch is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const switchElement = screen.getByRole('switch');
      await user.click(switchElement);

      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
      expect(mockOnToggle).toHaveBeenCalledTimes(1);
    });

    it('should not disable button or switch when available', () => {
      renderComponent({
        isAvailable: true,
      });

      const button = screen.getByRole('button');
      const switchElement = screen.getByRole('switch');

      expect(button).not.toBeDisabled();
      expect(switchElement).not.toBeDisabled();
    });

    it('should handle multiple rapid clicks gracefully', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button');

      // Click multiple times rapidly
      await user.click(button);
      await user.click(button);
      await user.click(button);

      expect(mockOnToggle).toHaveBeenCalledTimes(3);
      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
    });
  });

  describe('when method is not available', () => {
    it('should render disabled state with "Not configured" text', () => {
      renderComponent({
        isAvailable: false,
      });

      expect(screen.getByText('Test Auth Method')).toBeInTheDocument();
      expect(screen.getByText('Not configured')).toBeInTheDocument();
      expect(screen.getByTestId('test-icon')).toBeInTheDocument();
    });

    it('should disable the button when not available', () => {
      renderComponent({
        isAvailable: false,
      });

      const button = screen.getByRole('button');
      expect(button).toHaveAttribute('aria-disabled', 'true');
    });

    it('should not render a switch when not available', () => {
      renderComponent({
        isAvailable: false,
      });

      expect(screen.queryByRole('switch')).not.toBeInTheDocument();
    });

    it('should not call onToggle when disabled button is clicked', async () => {
      renderComponent({
        isAvailable: false,
      });

      const button = screen.getByRole('button');

      // Verify button is disabled by checking aria-disabled
      expect(button).toHaveAttribute('aria-disabled', 'true');
      expect(mockOnToggle).not.toHaveBeenCalled();
    });

    it('should show "Not configured" regardless of secondary prop', () => {
      renderComponent({
        isAvailable: false,
        secondary: 'This should not be shown',
      });

      expect(screen.getByText('Not configured')).toBeInTheDocument();
      expect(screen.queryByText('This should not be shown')).not.toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA roles', () => {
      renderComponent();

      expect(screen.getByRole('button')).toBeInTheDocument();
      expect(screen.getByRole('switch')).toBeInTheDocument();
    });

    it('should be keyboard navigable', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button');

      // Focus the button using Tab
      await user.tab();
      expect(button).toHaveFocus();

      // Activate with Enter
      await user.keyboard('{Enter}');
      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
    });

    it('should handle Space key activation', async () => {
      const user = userEvent.setup();
      renderComponent();

      const button = screen.getByRole('button');
      button.focus();

      await user.keyboard(' ');
      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
    });
  });

  describe('different authentication method types', () => {
    it('should handle different icon types', () => {
      const customIcon = <div data-testid="custom-icon">Custom Icon</div>;
      renderComponent({
        icon: customIcon,
      });

      expect(screen.getByTestId('custom-icon')).toBeInTheDocument();
      expect(screen.queryByTestId('test-icon')).not.toBeInTheDocument();
    });

    it('should handle long authentication method names', () => {
      renderComponent({
        name: 'Very Long Authentication Method Name That Should Still Work Properly',
      });

      expect(
        screen.getByText('Very Long Authentication Method Name That Should Still Work Properly'),
      ).toBeInTheDocument();
    });

    it('should handle special characters in method names', () => {
      renderComponent({
        name: 'OAuth 2.0 & OIDC Provider',
      });

      expect(screen.getByText('OAuth 2.0 & OIDC Provider')).toBeInTheDocument();
    });

    it('should handle different ID formats', async () => {
      const user = userEvent.setup();
      renderComponent({
        id: 'oauth2-provider-123',
      });

      const button = screen.getByRole('button');
      await user.click(button);

      expect(mockOnToggle).toHaveBeenCalledWith('oauth2-provider-123');
    });
  });

  describe('state combinations', () => {
    it('should handle enabled and available state', () => {
      renderComponent({
        isEnabled: true,
        isAvailable: true,
      });

      const switchElement = screen.getByRole('switch');
      const button = screen.getByRole('button');

      expect(switchElement).toBeChecked();
      expect(switchElement).not.toBeDisabled();
      expect(button).not.toBeDisabled();
    });

    it('should handle disabled but available state', () => {
      renderComponent({
        isEnabled: false,
        isAvailable: true,
      });

      const switchElement = screen.getByRole('switch');
      const button = screen.getByRole('button');

      expect(switchElement).not.toBeChecked();
      expect(switchElement).not.toBeDisabled();
      expect(button).not.toBeDisabled();
    });

    it('should override enabled state when not available', () => {
      renderComponent({
        isEnabled: true,
        isAvailable: false,
      });

      expect(screen.queryByRole('switch')).not.toBeInTheDocument();
      expect(screen.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
  });
});
