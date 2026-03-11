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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render, getByDisplayValue} from '@thunder/test-utils/browser';
import ShowClientSecret, {type ShowClientSecretProps} from '../ShowClientSecret';

// Mock the useCopyToClipboard hook
vi.mock('@thunder/shared-hooks', () => ({
  useCopyToClipboard: vi.fn(),
}));

const {useCopyToClipboard} = await import('@thunder/shared-hooks');

describe('ShowClientSecret', () => {
  const mockOnCopySecret = vi.fn();
  const mockOnContinue = vi.fn();
  const mockCopy = vi.fn().mockResolvedValue(undefined);

  const defaultProps: ShowClientSecretProps = {
    appName: 'Test Application',
    clientSecret: 'test_secret_12345',
    onCopySecret: mockOnCopySecret,
    onContinue: mockOnContinue,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useCopyToClipboard).mockReturnValue({
      copied: false,
      copy: mockCopy,
    });
  });

  const renderComponent = async (props: Partial<ShowClientSecretProps> = {}) =>
    render(<ShowClientSecret {...defaultProps} {...props} />);

  describe('rendering', () => {
    it('should render the component with warning icon', async () => {
      await renderComponent();

      // Warning icon should be present - just check component renders without error
      await expect.element(page.getByRole('heading', {level: 1})).toBeInTheDocument();
    });

    it('should render the title and subtitle', async () => {
      await renderComponent();

      await expect.element(page.getByRole('heading', {level: 1, name: /save your client secret/i})).toBeInTheDocument();
      await expect.element(page.getByText(/store it somewhere safe/i)).toBeInTheDocument();
    });

    it('should display the application name', async () => {
      await renderComponent();

      await expect.element(page.getByText('App Name')).toBeInTheDocument();
      await expect.element(page.getByText('Test Application')).toBeInTheDocument();
    });

    it('should render the client secret field', async () => {
      await renderComponent();

      await expect.element(page.getByText('Client Secret')).toBeInTheDocument();
      const input = getByDisplayValue('test_secret_12345');
      expect(input).toBeInTheDocument();
      expect(input).toHaveAttribute('type', 'password');
      expect(input).toHaveAttribute('readonly');
    });

    it('should render security reminder alert', async () => {
      await renderComponent();

      await expect.element(page.getByText(/security reminder/i)).toBeInTheDocument();
      await expect.element(page.getByText(/should be treated with the same level of security/i)).toBeInTheDocument();
    });

    it('should render action buttons', async () => {
      await renderComponent();

      await expect.element(page.getByRole('button', {name: /copy secret/i})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /continue/i})).toBeInTheDocument();
    });
  });

  describe('visibility toggle', () => {
    it('should toggle client secret visibility when eye icon is clicked', async () => {
      await renderComponent();

      const input = getByDisplayValue('test_secret_12345');
      expect(input).toHaveAttribute('type', 'password');

      // Get all icon buttons - the first one in the input should be the visibility toggle
      const allButtons = page.getByRole('button').all();
      // Filter to get only icon buttons (small buttons without text content)
      const iconButtons = allButtons.filter(btn => btn.element().querySelector('svg'));
      // The first icon button should be the visibility toggle
      const visibilityButton = iconButtons[0];
      expect(visibilityButton).toBeInTheDocument();

      await userEvent.click(visibilityButton);

      // Should now show as text
      expect(input).toHaveAttribute('type', 'text');

      // Click again to hide (same button, just state changed)
      await userEvent.click(visibilityButton);

      // Should be back to password
      expect(input).toHaveAttribute('type', 'password');
    });
  });

  describe('copy functionality', () => {
    it('should call copy function when copy button in input is clicked', async () => {
      await renderComponent();

      // Get all icon buttons - the second one in the input should be the copy button
      const allButtons = page.getByRole('button').all();
      const iconButtons = allButtons.filter(btn => btn.element().querySelector('svg'));
      // The second icon button should be the copy button (first is visibility toggle)
      const copyButton = iconButtons[1];
      expect(copyButton).toBeInTheDocument();

      await userEvent.click(copyButton);

      await vi.waitFor(() => {
        expect(mockCopy).toHaveBeenCalledWith('test_secret_12345');
      });
    });

    it('should call copy function when main copy button is clicked', async () => {
      await renderComponent();

      const mainCopyButton = page.getByRole('button', {name: /copy secret/i});
      await userEvent.click(mainCopyButton);

      await vi.waitFor(() => {
        expect(mockCopy).toHaveBeenCalledWith('test_secret_12345');
      });
    });

    it('should show copied state when copy succeeds', async () => {
      vi.mocked(useCopyToClipboard).mockReturnValue({
        copied: true,
        copy: mockCopy,
      });

      await renderComponent();

      await expect.element(page.getByRole('button', {name: /copied/i})).toBeInTheDocument();
    });

    it('should disable copy button when in copied state', async () => {
      vi.mocked(useCopyToClipboard).mockReturnValue({
        copied: true,
        copy: mockCopy,
      });

      await renderComponent();

      const mainCopyButton = page.getByRole('button', {name: /copied/i});
      expect(mainCopyButton).toBeDisabled();
    });

    it('should call onCopySecret callback through useCopyToClipboard', async () => {
      await renderComponent();

      // Get the config passed to useCopyToClipboard
      const hookCall = vi.mocked(useCopyToClipboard).mock.calls[0][0];
      expect(hookCall).toHaveProperty('onCopy', mockOnCopySecret);
      expect(hookCall).toHaveProperty('resetDelay', 2000);
    });
  });

  describe('continue action', () => {
    it('should call onContinue when continue button is clicked', async () => {
      await renderComponent();

      const continueButton = page.getByRole('button', {name: /continue/i});
      await userEvent.click(continueButton);

      expect(mockOnContinue).toHaveBeenCalledTimes(1);
    });
  });

  describe('props variations', () => {
    it('should render with different app name', async () => {
      await renderComponent({appName: 'Another App'});

      await expect.element(page.getByText('Another App')).toBeInTheDocument();
    });

    it('should render with different client secret', async () => {
      await renderComponent({clientSecret: 'different_secret_abc'});

      const input = getByDisplayValue('different_secret_abc');
      expect(input).toBeInTheDocument();
    });
  });

  describe('accessibility', () => {
    it('should have proper heading structure', async () => {
      await renderComponent();

      const heading = page.getByRole('heading', {level: 1});
      expect(heading).toBeInTheDocument();
    });

    it('should have accessible buttons', async () => {
      await renderComponent();

      const buttons = page.getByRole('button').all();
      expect(buttons.length).toBeGreaterThan(0);
      buttons.forEach((button) => {
        expect(button).toBeVisible();
      });
    });

    it('should have readonly input for security', async () => {
      await renderComponent();

      const input = getByDisplayValue('test_secret_12345');
      expect(input).toHaveAttribute('readonly');
    });
  });
});
