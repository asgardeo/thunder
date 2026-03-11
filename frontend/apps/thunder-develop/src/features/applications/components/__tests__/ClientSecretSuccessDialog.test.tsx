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

import {render} from '@thunder/test-utils/browser';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import ClientSecretSuccessDialog from '../ClientSecretSuccessDialog';

describe('ClientSecretSuccessDialog', () => {
  const mockOnClose = vi.fn();
  const testClientSecret = 'test-client-secret-abc123xyz789';

  const defaultProps = {
    open: true,
    clientSecret: testClientSecret,
    onClose: mockOnClose,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the dialog when open is true', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
      await expect.element(page.getByText('Save Your New Client Secret')).toBeInTheDocument();
    });

    it('should display the subtitle message', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await expect.element(
        page.getByText("This is the only time you'll see this secret. Store it somewhere safe."),
      ).toBeInTheDocument();
    });

    it('should display the client secret label', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await expect.element(page.getByText('New Client Secret', {exact: true})).toBeInTheDocument();
    });

    it('should have the client secret as a masked password field by default', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      // The dialog renders in a portal - query the document for the password input
      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
      const dialog = page.getByRole('dialog').element();
      const passwordInput = dialog.querySelector('input[type="password"]');
      expect(passwordInput).not.toBeNull();
      expect(passwordInput).toHaveValue(testClientSecret);
    });

    it('should display the security reminder', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await expect.element(page.getByText('Security Reminder')).toBeInTheDocument();
      await expect.element(
        page.getByText(
          'Never share your client secret publicly or store it in version control. If you believe your secret has been compromised, regenerate it immediately.',
        ),
      ).toBeInTheDocument();
    });

    it('should not render dialog content when open is false', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} open={false} />);

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
    });

    it('should render Done and Copy Secret buttons', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await expect.element(page.getByRole('button', {name: 'Done'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Copy Secret'})).toBeInTheDocument();
    });

    it('should render the warning icon', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      // The AlertTriangle icon is rendered inside the dialog
      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
      // Verify the dialog has SVG icons by checking presence of SVG elements inside dialog
      const dialog = page.getByRole('dialog').element();
      const svgIcons = dialog.querySelectorAll('svg');
      expect(svgIcons.length).toBeGreaterThan(0);
    });
  });

  describe('User Interactions', () => {
    it('should call onClose when Done button is clicked', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Done'}));

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when Escape key is pressed', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      await userEvent.keyboard('{Escape}');

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should copy client secret to clipboard when Copy Secret button is clicked', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);
      const writeTextSpy = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined);

      await userEvent.click(page.getByRole('button', {name: 'Copy Secret'}));

      expect(writeTextSpy).toHaveBeenCalledWith(testClientSecret);
    });

    it('should copy client secret when inline copy icon is clicked', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);
      const writeTextSpy = vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined);

      await userEvent.click(page.getByRole('button', {name: 'Copy to clipboard'}));

      expect(writeTextSpy).toHaveBeenCalledWith(testClientSecret);
    });

    it('should show "Copied to clipboard" text after successful copy', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);
      vi.spyOn(navigator.clipboard, 'writeText').mockResolvedValue(undefined);

      await userEvent.click(page.getByRole('button', {name: 'Copy Secret'}));

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Copied to clipboard')).toBeInTheDocument();
      });
    });

    it('should toggle secret visibility when visibility button is clicked', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      // Initially should be password type (masked)
      const dialog = page.getByRole('dialog').element();
      expect(dialog.querySelector('input[type="password"]')).not.toBeNull();

      // Click toggle visibility button
      await userEvent.click(page.getByRole('button', {name: 'Toggle secret visibility'}));

      // Should now be text type (visible)
      await vi.waitFor(() => {
        const textInput = dialog.querySelector('input[type="text"]');
        expect(textInput).not.toBeNull();
        expect(textInput).toHaveValue(testClientSecret);
      });
    });
  });

  describe('Text Field Properties', () => {
    it('should have a readonly text field', async () => {
      await render(<ClientSecretSuccessDialog {...defaultProps} />);

      // The dialog renders in a portal - query the document for the input
      const dialog = page.getByRole('dialog').element();
      const input = dialog.querySelector('input');
      expect(input).toHaveAttribute('readonly');
    });

    it('should display the full client secret when visible', async () => {
      const longSecret = 'a'.repeat(64);
      await render(<ClientSecretSuccessDialog {...defaultProps} clientSecret={longSecret} />);

      const dialog = page.getByRole('dialog').element();

      // Toggle visibility to see the full secret
      await userEvent.click(page.getByRole('button', {name: 'Toggle secret visibility'}));

      await vi.waitFor(() => {
        const textInput = dialog.querySelector('input[type="text"]');
        expect(textInput).not.toBeNull();
        expect(textInput).toHaveValue(longSecret);
      });
    });

    it('should reset visibility state when dialog is closed and reopened', async () => {
      const {rerender} = await render(<ClientSecretSuccessDialog {...defaultProps} />);

      const dialog = page.getByRole('dialog').element();

      // Toggle visibility
      await userEvent.click(page.getByRole('button', {name: 'Toggle secret visibility'}));

      // Verify toggled to text type
      await vi.waitFor(() => {
        expect(dialog.querySelector('input[type="text"]')).not.toBeNull();
      });

      // Click Done to close
      await userEvent.click(page.getByRole('button', {name: 'Done'}));

      // Rerender with open: false then open: true
      await rerender(<ClientSecretSuccessDialog {...defaultProps} open={false} />);
      await rerender(<ClientSecretSuccessDialog {...defaultProps} open />);

      // Dialog should reopen with masked field (password type)
      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
      await vi.waitFor(() => {
        const reopenedDialog = page.getByRole('dialog').element();
        expect(reopenedDialog.querySelector('input[type="password"]')).not.toBeNull();
      });
    });
  });
});
