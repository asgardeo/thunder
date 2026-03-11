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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import LogoUpdateModal from '../LogoUpdateModal';

// Mock the utils
vi.mock('../../features/applications/utils/generateAppLogoSuggestion', () => ({
  default: vi.fn((count: number) => Array.from({length: count}, (_, i) => `https://logo${i + 1}.com/logo.png`)),
}));

describe('LogoUpdateModal', () => {
  const mockOnClose = vi.fn();
  const mockOnLogoUpdate = vi.fn();
  const currentLogoUrl = 'https://current-logo.com/logo.png';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render modal when open is true', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await expect.element(page.getByText('Update Application Logo')).toBeInTheDocument();
    });

    it('should not render modal content when open is false', async () => {
      await render(<LogoUpdateModal open={false} onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await expect.element(page.getByText('Update Application Logo')).not.toBeInTheDocument();
    });

    it('should display preview section', async () => {
      await render(
        <LogoUpdateModal open onClose={mockOnClose} currentLogoUrl={currentLogoUrl} onLogoUpdate={mockOnLogoUpdate} />,
      );

      await expect.element(page.getByText('Preview')).toBeInTheDocument();
    });

    it('should display custom URL input section', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await expect.element(page.getByText('Custom Logo URL')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('https://example.com/logo.png')).toBeInTheDocument();
    });

    it('should display suggested logos section', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await expect.element(page.getByText('Application Logo')).toBeInTheDocument();
      await expect.element(page.getByText('Shuffle')).toBeInTheDocument();
    });

    it('should render 12 logo suggestions', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const images = page.getByRole('img').all();
      // Filter out the preview image
      const suggestionImages: typeof images = [];
      // eslint-disable-next-line no-restricted-syntax
      for (const img of images) {
        const src = img.element().getAttribute('src');
        if (src?.startsWith('https://logo')) {
          suggestionImages.push(img);
        }
      }

      expect(suggestionImages.length).toBe(12);
    });
  });

  describe('User Interactions', () => {
    it('should call onClose when close button is clicked', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const buttons = page.getByRole('button').all();
      await userEvent.click(buttons[0]); // X icon button

      await vi.waitFor(() => {
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });

    it('should call onClose when cancel button is clicked', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await userEvent.click(page.getByRole('button', {name: 'Cancel'}));

      await vi.waitFor(() => {
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });

    it('should update custom URL when typing in input', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const input = page.getByPlaceholder('https://example.com/logo.png');
      await userEvent.fill(input, 'https://custom-logo.com/logo.png');

      await expect.element(input).toHaveValue('https://custom-logo.com/logo.png');
    });

    it('should generate new suggestions when shuffle button is clicked', async () => {
      const generateAppLogoSuggestions = await import('../../features/applications/utils/generateAppLogoSuggestion');

      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await userEvent.click(page.getByText('Shuffle'));

      await vi.waitFor(() => {
        expect(generateAppLogoSuggestions.default).toHaveBeenCalledWith(12);
      });
    });

    it('should call onLogoUpdate with custom URL when update button is clicked', async () => {
      const customUrl = 'https://custom-logo.com/logo.png';
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const input = page.getByPlaceholder('https://example.com/logo.png');
      await userEvent.fill(input, customUrl);

      await userEvent.click(page.getByRole('button', {name: 'Update Logo'}));

      await vi.waitFor(() => {
        expect(mockOnLogoUpdate).toHaveBeenCalledWith(customUrl);
      });
    });

    it('should select logo from suggestions when clicked', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const images = page.getByRole('img').all();
      let firstSuggestion: (typeof images)[0] | undefined;
      // eslint-disable-next-line no-restricted-syntax
      for (const img of images) {
        const src = img.element().getAttribute('src');
        if (src === 'https://logo1.com/logo.png') {
          firstSuggestion = img;
          break;
        }
      }

      if (firstSuggestion) {
        const {parentElement} = firstSuggestion.element();
        if (parentElement) {
          await userEvent.click(page.elementLocator(parentElement));
        }
      }

      await userEvent.click(page.getByRole('button', {name: 'Update Logo'}));

      await vi.waitFor(() => {
        expect(mockOnLogoUpdate).toHaveBeenCalledWith('https://logo1.com/logo.png');
      });
    });

    it('should clear selected logo when entering custom URL', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      // First select a logo
      const images = page.getByRole('img').all();
      let firstSuggestion: (typeof images)[0] | undefined;
      // eslint-disable-next-line no-restricted-syntax
      for (const img of images) {
        const src = img.element().getAttribute('src');
        if (src === 'https://logo1.com/logo.png') {
          firstSuggestion = img;
          break;
        }
      }

      if (firstSuggestion) {
        const {parentElement} = firstSuggestion.element();
        if (parentElement) {
          await userEvent.click(page.elementLocator(parentElement));
        }
      }

      // Then type in custom URL
      const input = page.getByPlaceholder('https://example.com/logo.png');
      await userEvent.fill(input, 'https://custom.com/logo.png');

      await userEvent.click(page.getByRole('button', {name: 'Update Logo'}));

      // Should use custom URL, not selected logo
      await vi.waitFor(() => {
        expect(mockOnLogoUpdate).toHaveBeenCalledWith('https://custom.com/logo.png');
      });
    });

    it('should clear custom URL when selecting from suggestions', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      // First type custom URL
      const input = page.getByPlaceholder('https://example.com/logo.png');
      await userEvent.fill(input, 'https://custom.com/logo.png');

      // Then select a logo
      const images = page.getByRole('img').all();
      let firstSuggestion: (typeof images)[0] | undefined;
      // eslint-disable-next-line no-restricted-syntax
      for (const img of images) {
        const src = img.element().getAttribute('src');
        if (src === 'https://logo1.com/logo.png') {
          firstSuggestion = img;
          break;
        }
      }

      if (firstSuggestion) {
        const {parentElement} = firstSuggestion.element();
        if (parentElement) {
          await userEvent.click(page.elementLocator(parentElement));
        }
      }

      await expect.element(input).toHaveValue('');
    });
  });

  describe('Update Button State', () => {
    it('should enable update button when custom URL is provided', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const updateButton = page.getByRole('button', {name: 'Update Logo'});
      await expect.element(updateButton).toBeDisabled();

      const input = page.getByPlaceholder('https://example.com/logo.png');
      await userEvent.fill(input, 'https://custom-logo.com/logo.png');

      await expect.element(updateButton).not.toBeDisabled();
    });

    it('should enable update button when logo is selected', async () => {
      await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const images = page.getByRole('img').all();
      let firstSuggestion: (typeof images)[0] | undefined;
      // eslint-disable-next-line no-restricted-syntax
      for (const img of images) {
        const src = img.element().getAttribute('src');
        if (src === 'https://logo1.com/logo.png') {
          firstSuggestion = img;
          break;
        }
      }

      if (firstSuggestion) {
        const {parentElement} = firstSuggestion.element();
        if (parentElement) {
          await userEvent.click(page.elementLocator(parentElement));
        }
      }

      const updateButton = page.getByRole('button', {name: 'Update Logo'});

      await expect.element(updateButton).not.toBeDisabled();
    });

    it('should enable update button when current logo URL is provided', async () => {
      await render(
        <LogoUpdateModal open onClose={mockOnClose} currentLogoUrl={currentLogoUrl} onLogoUpdate={mockOnLogoUpdate} />,
      );

      const updateButton = page.getByRole('button', {name: 'Update Logo'});

      await expect.element(updateButton).not.toBeDisabled();
    });
  });

  describe('Initial State', () => {
    it('should populate custom URL with current logo URL on mount', async () => {
      await render(
        <LogoUpdateModal open onClose={mockOnClose} currentLogoUrl={currentLogoUrl} onLogoUpdate={mockOnLogoUpdate} />,
      );

      const input = page.getByPlaceholder('https://example.com/logo.png');

      await expect.element(input).toHaveValue(currentLogoUrl);
    });

    it('should reset state when modal is reopened', async () => {
      const {rerender} = await render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const input = page.getByPlaceholder('https://example.com/logo.png');
      await userEvent.fill(input, 'https://test.com/logo.png');

      // Close modal
      await rerender(<LogoUpdateModal open={false} onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      // Reopen modal
      await rerender(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await expect.element(page.getByPlaceholder('https://example.com/logo.png')).toHaveValue('');
    });
  });
});
