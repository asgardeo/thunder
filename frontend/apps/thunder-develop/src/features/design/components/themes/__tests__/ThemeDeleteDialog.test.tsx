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
import {render, page, userEvent} from '@thunder/test-utils/browser';
import ThemeDeleteDialog from '../ThemeDeleteDialog';

const mockMutate = vi.fn();

vi.mock('@thunder/shared-design', () => ({
  useDeleteTheme: vi.fn(() => ({
    mutate: mockMutate,
    isPending: false,
  })),
}));

describe('ThemeDeleteDialog', () => {
  beforeEach(() => {
    mockMutate.mockReset();
  });

  describe('Rendering', () => {
    it('renders Dialog when open is true', async () => {
      render(<ThemeDeleteDialog themeName="Ocean Blue" open themeId="theme-1" onClose={vi.fn()} />);
      // Dialog is open — some content is visible
      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
    });

    it('does not render dialog content when open is false', async () => {
      render(<ThemeDeleteDialog themeName="Ocean Blue" open={false} themeId="theme-1" onClose={vi.fn()} />);
      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
    });

    it('renders the theme name in the dialog', async () => {
      render(<ThemeDeleteDialog themeName="Ocean Blue" open themeId="theme-1" onClose={vi.fn()} />);
      await expect.element(page.getByText(/Ocean Blue/)).toBeInTheDocument();
    });

    it('renders without crashing when themeName is undefined', async () => {
      render(<ThemeDeleteDialog themeName={null} open themeId="theme-1" onClose={vi.fn()} />);
      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
    });

    it('renders a delete button', async () => {
      render(<ThemeDeleteDialog open themeId="theme-1" themeName="Test" onClose={vi.fn()} />);
      // Should contain a destructive/delete action button (label comes from i18n key)
      expect(page.getByRole('button').all().length).toBeGreaterThan(0);
    });
  });

  describe('Cancel behaviour', () => {
    it('calls onClose when the cancel button is clicked', async () => {
      const onClose = vi.fn();
      render(<ThemeDeleteDialog open themeId="theme-1" themeName="My Theme" onClose={onClose} />);

      // Find cancel button by its translation key text
      const cancelBtn = page.getByText('common:actions.cancel');
      await userEvent.click(cancelBtn);

      expect(onClose).toHaveBeenCalledOnce();
    });

    it('does not call mutate when cancel is clicked', async () => {
      render(<ThemeDeleteDialog open themeId="theme-1" themeName="My Theme" onClose={vi.fn()} />);

      const cancelBtn = page.getByText('common:actions.cancel');
      await userEvent.click(cancelBtn);

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });

  describe('Delete behaviour', () => {
    it('calls mutate with the themeId when delete is confirmed', async () => {
      render(<ThemeDeleteDialog open themeId="theme-abc" themeName="My Theme" onClose={vi.fn()} />);

      const deleteBtn = page.getByText('common:actions.delete');
      await userEvent.click(deleteBtn);

      expect(mockMutate).toHaveBeenCalledWith('theme-abc', expect.any(Object));
    });

    it('does not call mutate when themeId is undefined', async () => {
      render(<ThemeDeleteDialog themeId={null} themeName="My Theme" open onClose={vi.fn()} />);

      // When themeId is null the delete button is disabled, preventing any click
      const deleteBtn = page.getByRole('button', {name: 'common:actions.delete'});
      await expect.element(deleteBtn).toBeDisabled();
      expect(mockMutate).not.toHaveBeenCalled();
    });
  });

  describe('Loading state', () => {
    it('disables buttons when isPending is true', async () => {
      const {useDeleteTheme} = await import('@thunder/shared-design');
      (useDeleteTheme as ReturnType<typeof vi.fn>).mockReturnValueOnce({
        mutate: mockMutate,
        isPending: true,
      });

      render(<ThemeDeleteDialog open themeId="theme-1" themeName="Test" onClose={vi.fn()} />);

      const buttons = page.getByRole('button').all();
      const disabledButtons: typeof buttons = [];
      for (const btn of buttons) {
        const el = await btn.element();
        if (el.hasAttribute('disabled')) {
          disabledButtons.push(btn);
        }
      }
      expect(disabledButtons.length).toBeGreaterThan(0);
    });
  });

  describe('Success callback', () => {
    it('calls onSuccess after successful deletion', async () => {
      const onSuccess = vi.fn();
      const onClose = vi.fn();

      // Simulate mutate calling onSuccess callback
      mockMutate.mockImplementation((_: unknown, callbacks: {onSuccess?: () => void}) => {
        callbacks?.onSuccess?.();
      });

      render(<ThemeDeleteDialog open themeId="theme-1" themeName="My Theme" onClose={onClose} onSuccess={onSuccess} />);

      const deleteBtn = page.getByText('common:actions.delete');
      await userEvent.click(deleteBtn);

      await vi.waitFor(() => {
        expect(onSuccess).toHaveBeenCalledOnce();
        expect(onClose).toHaveBeenCalledOnce();
      });
    });
  });
});
