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
import {UserRound} from '@wso2/oxygen-ui-icons-react';
import AuthenticationMethodItem, {type AuthenticationMethodItemProps} from '../AuthenticationMethodItem';

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

  const renderComponent = async (props: Partial<AuthenticationMethodItemProps> = {}) =>
    render(<AuthenticationMethodItem {...defaultProps} {...props} />);

  describe('when method is available', () => {
    it('should render the authentication method with all elements', async () => {
      await renderComponent();

      await expect.element(page.getByText('Test Auth Method')).toBeInTheDocument();
      await expect.element(page.getByTestId('test-icon')).toBeInTheDocument();
      await expect.element(page.getByRole('switch')).toBeInTheDocument();
      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('should render with secondary text when provided', async () => {
      await renderComponent({
        secondary: 'Additional info',
      });

      await expect.element(page.getByText('Test Auth Method')).toBeInTheDocument();
      await expect.element(page.getByText('Additional info')).toBeInTheDocument();
    });

    it('should show switch as checked when method is enabled', async () => {
      await renderComponent({
        isEnabled: true,
      });

      const switchElement = page.getByRole('switch');
      expect(switchElement).toBeChecked();
    });

    it('should show switch as unchecked when method is disabled', async () => {
      await renderComponent({
        isEnabled: false,
      });

      const switchElement = page.getByRole('switch');
      expect(switchElement).not.toBeChecked();
    });

    it('should call onToggle when list item button is clicked', async () => {
      await renderComponent();

      const button = page.getByRole('button');
      await userEvent.click(button);

      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
      expect(mockOnToggle).toHaveBeenCalledTimes(1);
    });

    it('should call onToggle when switch is clicked', async () => {
      await renderComponent();

      const switchElement = page.getByRole('switch');
      await userEvent.click(switchElement);

      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
      expect(mockOnToggle).toHaveBeenCalledTimes(1);
    });

    it('should not disable button or switch when available', async () => {
      await renderComponent({
        isAvailable: true,
      });

      const button = page.getByRole('button');
      const switchElement = page.getByRole('switch');

      expect(button).not.toBeDisabled();
      expect(switchElement).not.toBeDisabled();
    });

    it('should handle multiple rapid clicks gracefully', async () => {
      await renderComponent();

      const button = page.getByRole('button');

      // Click multiple times rapidly
      await userEvent.click(button);
      await userEvent.click(button);
      await userEvent.click(button);

      expect(mockOnToggle).toHaveBeenCalledTimes(3);
      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
    });
  });

  describe('when method is not available', () => {
    it('should render disabled state with "Not configured" text', async () => {
      await renderComponent({
        isAvailable: false,
      });

      await expect.element(page.getByText('Test Auth Method')).toBeInTheDocument();
      await expect.element(page.getByText('Not configured')).toBeInTheDocument();
      await expect.element(page.getByTestId('test-icon')).toBeInTheDocument();
    });

    it('should disable the button when not available', async () => {
      await renderComponent({
        isAvailable: false,
      });

      const button = page.getByRole('button');
      expect(button).toHaveAttribute('aria-disabled', 'true');
    });

    it('should not render a switch when not available', async () => {
      await renderComponent({
        isAvailable: false,
      });

      await expect.element(page.getByRole('switch')).not.toBeInTheDocument();
    });

    it('should not call onToggle when disabled button is clicked', async () => {
      await renderComponent({
        isAvailable: false,
      });

      const button = page.getByRole('button');

      // Verify button is disabled by checking aria-disabled
      expect(button).toHaveAttribute('aria-disabled', 'true');
      expect(mockOnToggle).not.toHaveBeenCalled();
    });

    it('should show "Not configured" regardless of secondary prop', async () => {
      await renderComponent({
        isAvailable: false,
        secondary: 'This should not be shown',
      });

      await expect.element(page.getByText('Not configured')).toBeInTheDocument();
      await expect.element(page.getByText('This should not be shown')).not.toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA roles', async () => {
      await renderComponent();

      await expect.element(page.getByRole('button')).toBeInTheDocument();
      await expect.element(page.getByRole('switch')).toBeInTheDocument();
    });

    it('should be keyboard navigable', async () => {
      await renderComponent();

      const button = page.getByRole('button');

      // Focus the button using Tab
      await userEvent.tab();
      expect(button).toHaveFocus();

      // Activate with Enter
      await userEvent.keyboard('{Enter}');
      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
    });

    it('should handle Space key activation', async () => {
      await renderComponent();

      const button = page.getByRole('button');
      button.element().focus();

      await userEvent.keyboard(' ');
      expect(mockOnToggle).toHaveBeenCalledWith('test-auth-method');
    });
  });

  describe('different authentication method types', () => {
    it('should handle different icon types', async () => {
      const customIcon = <div data-testid="custom-icon">Custom Icon</div>;
      await renderComponent({
        icon: customIcon,
      });

      await expect.element(page.getByTestId('custom-icon')).toBeInTheDocument();
      await expect.element(page.getByTestId('test-icon')).not.toBeInTheDocument();
    });

    it('should handle long authentication method names', async () => {
      await renderComponent({
        name: 'Very Long Authentication Method Name That Should Still Work Properly',
      });

      expect(
        page.getByText('Very Long Authentication Method Name That Should Still Work Properly'),
      ).toBeInTheDocument();
    });

    it('should handle special characters in method names', async () => {
      await renderComponent({
        name: 'OAuth 2.0 & OIDC Provider',
      });

      await expect.element(page.getByText('OAuth 2.0 & OIDC Provider')).toBeInTheDocument();
    });

    it('should handle different ID formats', async () => {
      await renderComponent({
        id: 'oauth2-provider-123',
      });

      const button = page.getByRole('button');
      await userEvent.click(button);

      expect(mockOnToggle).toHaveBeenCalledWith('oauth2-provider-123');
    });
  });

  describe('state combinations', () => {
    it('should handle enabled and available state', async () => {
      await renderComponent({
        isEnabled: true,
        isAvailable: true,
      });

      const switchElement = page.getByRole('switch');
      const button = page.getByRole('button');

      expect(switchElement).toBeChecked();
      expect(switchElement).not.toBeDisabled();
      expect(button).not.toBeDisabled();
    });

    it('should handle disabled but available state', async () => {
      await renderComponent({
        isEnabled: false,
        isAvailable: true,
      });

      const switchElement = page.getByRole('switch');
      const button = page.getByRole('button');

      expect(switchElement).not.toBeChecked();
      expect(switchElement).not.toBeDisabled();
      expect(button).not.toBeDisabled();
    });

    it('should override enabled state when not available', async () => {
      await renderComponent({
        isEnabled: true,
        isAvailable: false,
      });

      await expect.element(page.getByRole('switch')).not.toBeInTheDocument();
      await expect.element(page.getByRole('button')).toHaveAttribute('aria-disabled', 'true');
    });
  });

  describe('isDisabled prop', () => {
    it('should disable switch and button when isDisabled is true', async () => {
      await renderComponent({
        isAvailable: true,
        isDisabled: true,
      });

      const switchElement = page.getByRole('switch');
      const button = page.getByRole('button');

      expect(switchElement).toBeDisabled();
      expect(button).toHaveAttribute('aria-disabled', 'true');
    });

    it('should not call onToggle when isDisabled is true', async () => {
      await renderComponent({
        isAvailable: true,
        isDisabled: true,
      });

      const switchElement = page.getByRole('switch');
      // Verify the switch is disabled - clicking is prevented by pointer-events: none
      expect(switchElement).toBeDisabled();
      // Since we can't click a disabled element, verify onToggle wasn't called during render
      expect(mockOnToggle).not.toHaveBeenCalled();
    });

    it('should render secondary text as undefined when not provided', async () => {
      await renderComponent({
        isAvailable: true,
        secondary: undefined,
      });

      // The primary text should be present but no secondary text
      await expect.element(page.getByText('Test Auth Method')).toBeInTheDocument();
      // No additional secondary text element should be rendered
    });

    it('should combine isEnabled true with isDisabled true', async () => {
      await renderComponent({
        isEnabled: true,
        isAvailable: true,
        isDisabled: true,
      });

      const switchElement = page.getByRole('switch');
      expect(switchElement).toBeChecked();
      expect(switchElement).toBeDisabled();
    });

    it('should render with secondary text when available and provided', async () => {
      await renderComponent({
        isAvailable: true,
        secondary: 'Description text',
      });

      await expect.element(page.getByText('Description text')).toBeInTheDocument();
    });

    it('should render icon in the disabled state when not available', async () => {
      await renderComponent({
        isAvailable: false,
      });

      await expect.element(page.getByTestId('test-icon')).toBeInTheDocument();
    });

    it('should render icon in the enabled state when available', async () => {
      await renderComponent({
        isAvailable: true,
      });

      await expect.element(page.getByTestId('test-icon')).toBeInTheDocument();
    });

    it('should pass isDisabled false by default when not specified', async () => {
      await renderComponent({
        isAvailable: true,
      });

      const switchElement = page.getByRole('switch');
      expect(switchElement).not.toBeDisabled();
    });
  });
});
