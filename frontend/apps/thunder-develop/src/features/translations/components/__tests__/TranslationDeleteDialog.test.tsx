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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {page, userEvent, renderWithProviders} from '@thunder/test-utils/browser';
import TranslationDeleteDialog from '../TranslationDeleteDialog';

const mockMutate = vi.fn();
vi.mock('@thunder/i18n', () => ({
  useDeleteTranslations: () => ({mutate: mockMutate, isPending: false}),
  getDisplayNameForCode: (code: string) => `DisplayName(${code})`,
}));

const defaultProps = {
  open: true,
  language: 'fr-FR',
  onClose: vi.fn(),
  onSuccess: vi.fn(),
};

describe('TranslationDeleteDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the dialog title', async () => {
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} />);

      await expect.element(page.getByText('Delete Language')).toBeInTheDocument();
    });

    it('renders the warning disclaimer', async () => {
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} />);

      await expect.element(
        page.getByText('All custom translations for this language will be permanently removed and reset to defaults.'),
      ).toBeInTheDocument();
    });

    it('renders cancel and delete buttons', async () => {
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} />);

      await expect.element(page.getByRole('button', {name: 'Cancel'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Delete'})).toBeInTheDocument();
    });
  });

  describe('Cancel', () => {
    it('calls onClose when cancel is clicked', async () => {
      const onClose = vi.fn();
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} onClose={onClose} />);

      await userEvent.click(page.getByRole('button', {name: 'Cancel'}));

      expect(onClose).toHaveBeenCalled();
    });
  });

  describe('Delete', () => {
    it('calls deleteTranslations.mutate with the language when delete is clicked', async () => {
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Delete'}));

      expect(mockMutate).toHaveBeenCalledWith(
        'fr-FR',
        expect.objectContaining({
          onSuccess: expect.any(Function) as unknown as () => void,
          onError: expect.any(Function) as unknown as () => void,
        }),
      );
    });

    it('does not call mutate when language is null', async () => {
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} language={null} />);

      await userEvent.click(page.getByRole('button', {name: 'Delete'}));

      expect(mockMutate).not.toHaveBeenCalled();
    });

    it('calls onClose and onSuccess on successful deletion', async () => {
      const onClose = vi.fn();
      const onSuccess = vi.fn();
      mockMutate.mockImplementation((_lang: string, opts: {onSuccess: () => void}) => {
        opts.onSuccess();
      });
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} onClose={onClose} onSuccess={onSuccess} />);

      await userEvent.click(page.getByRole('button', {name: 'Delete'}));

      expect(onClose).toHaveBeenCalled();
      expect(onSuccess).toHaveBeenCalled();
    });

    it('shows an error alert on deletion failure', async () => {
      mockMutate.mockImplementation((_lang: string, opts: {onError: () => void}) => {
        opts.onError();
      });
      await renderWithProviders(<TranslationDeleteDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Delete'}));

      await expect.element(page.getByText('Failed to delete translations. Please try again.')).toBeInTheDocument();
    });
  });
});
