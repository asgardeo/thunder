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
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import LogoUpdateModal from '../LogoUpdateModal';

// Mock the utils
vi.mock('../../features/applications/utils/generateAppLogoSuggestion', () => ({
  default: vi.fn((count: number) => Array.from({length: count}, (_, i) => `https://logo${i + 1}.com/logo.png`)),
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:logoModal.title': 'Update Logo',
        'applications:logoModal.preview.title': 'Preview',
        'applications:logoModal.customUrl.title': 'Custom URL',
        'applications:logoModal.customUrl.placeholder': 'Enter image URL',
        'applications:logoModal.customUrl.hint': 'Provide a URL to an image',
        'applications:logoModal.logos.title': 'Suggested Logos',
        'applications:logoModal.logos.shuffle': 'Shuffle',
        'applications:logoModal.cancel': 'Cancel',
        'applications:logoModal.update': 'Update',
      };
      return translations[key] || key;
    },
  }),
}));

describe('LogoUpdateModal', () => {
  const mockOnClose = vi.fn();
  const mockOnLogoUpdate = vi.fn();
  const currentLogoUrl = 'https://current-logo.com/logo.png';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render modal when open is true', () => {
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      expect(screen.getByText('Update Logo')).toBeInTheDocument();
    });

    it('should not render modal content when open is false', () => {
      render(<LogoUpdateModal open={false} onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      expect(screen.queryByText('Update Logo')).not.toBeInTheDocument();
    });

    it('should display preview section', () => {
      render(
        <LogoUpdateModal open onClose={mockOnClose} currentLogoUrl={currentLogoUrl} onLogoUpdate={mockOnLogoUpdate} />,
      );

      expect(screen.getByText('Preview')).toBeInTheDocument();
    });

    it('should display custom URL input section', () => {
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      expect(screen.getByText('Custom URL')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Enter image URL')).toBeInTheDocument();
    });

    it('should display suggested logos section', () => {
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      expect(screen.getByText('Suggested Logos')).toBeInTheDocument();
      expect(screen.getByText('Shuffle')).toBeInTheDocument();
    });

    it('should render 12 logo suggestions', () => {
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const images = screen.getAllByRole('img');
      // Filter out the preview image
      const suggestionImages = images.filter((img) => img.getAttribute('src')?.startsWith('https://logo'));

      expect(suggestionImages.length).toBe(12);
    });
  });

  describe('User Interactions', () => {
    it('should call onClose when close button is clicked', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const closeButton = screen.getAllByRole('button')[0]; // X icon button
      await user.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when cancel button is clicked', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const cancelButton = screen.getByRole('button', {name: 'Cancel'});
      await user.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should update custom URL when typing in input', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const input = screen.getByPlaceholderText('Enter image URL');
      await user.type(input, 'https://custom-logo.com/logo.png');

      expect(input).toHaveValue('https://custom-logo.com/logo.png');
    });

    it('should generate new suggestions when shuffle button is clicked', async () => {
      const user = userEvent.setup();
      const generateAppLogoSuggestions = await import('../../features/applications/utils/generateAppLogoSuggestion');

      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const shuffleButton = screen.getByText('Shuffle');
      await user.click(shuffleButton);

      await waitFor(() => {
        expect(generateAppLogoSuggestions.default).toHaveBeenCalledWith(12);
      });
    });

    it('should call onLogoUpdate with custom URL when update button is clicked', async () => {
      const user = userEvent.setup();
      const customUrl = 'https://custom-logo.com/logo.png';
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const input = screen.getByPlaceholderText('Enter image URL');
      await user.type(input, customUrl);

      const updateButton = screen.getByRole('button', {name: 'Update'});
      await user.click(updateButton);

      expect(mockOnLogoUpdate).toHaveBeenCalledWith(customUrl);
    });

    it('should select logo from suggestions when clicked', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const images = screen.getAllByRole('img');
      const firstSuggestion = images.find((img) => img.getAttribute('src') === 'https://logo1.com/logo.png');

      if (firstSuggestion?.parentElement) {
        await user.click(firstSuggestion.parentElement);
      }

      const updateButton = screen.getByRole('button', {name: 'Update'});
      await user.click(updateButton);

      expect(mockOnLogoUpdate).toHaveBeenCalledWith('https://logo1.com/logo.png');
    });

    it('should clear selected logo when entering custom URL', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      // First select a logo
      const images = screen.getAllByRole('img');
      const firstSuggestion = images.find((img) => img.getAttribute('src') === 'https://logo1.com/logo.png');

      if (firstSuggestion?.parentElement) {
        await user.click(firstSuggestion.parentElement);
      }

      // Then type in custom URL
      const input = screen.getByPlaceholderText('Enter image URL');
      await user.type(input, 'https://custom.com/logo.png');

      const updateButton = screen.getByRole('button', {name: 'Update'});
      await user.click(updateButton);

      // Should use custom URL, not selected logo
      expect(mockOnLogoUpdate).toHaveBeenCalledWith('https://custom.com/logo.png');
    });

    it('should clear custom URL when selecting from suggestions', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      // First type custom URL
      const input = screen.getByPlaceholderText('Enter image URL');
      await user.type(input, 'https://custom.com/logo.png');

      // Then select a logo
      const images = screen.getAllByRole('img');
      const firstSuggestion = images.find((img) => img.getAttribute('src') === 'https://logo1.com/logo.png');

      if (firstSuggestion?.parentElement) {
        await user.click(firstSuggestion.parentElement);
      }

      expect(input).toHaveValue('');
    });
  });

  describe('Update Button State', () => {
    it('should enable update button when custom URL is provided', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const updateButton = screen.getByRole('button', {name: 'Update'});
      expect(updateButton).toBeDisabled();

      const input = screen.getByPlaceholderText('Enter image URL');
      await user.type(input, 'https://custom-logo.com/logo.png');

      expect(updateButton).not.toBeDisabled();
    });

    it('should enable update button when logo is selected', async () => {
      const user = userEvent.setup();
      render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const images = screen.getAllByRole('img');
      const firstSuggestion = images.find((img) => img.getAttribute('src') === 'https://logo1.com/logo.png');

      if (firstSuggestion?.parentElement) {
        await user.click(firstSuggestion.parentElement);
      }

      const updateButton = screen.getByRole('button', {name: 'Update'});

      expect(updateButton).not.toBeDisabled();
    });

    it('should enable update button when current logo URL is provided', () => {
      render(
        <LogoUpdateModal open onClose={mockOnClose} currentLogoUrl={currentLogoUrl} onLogoUpdate={mockOnLogoUpdate} />,
      );

      const updateButton = screen.getByRole('button', {name: 'Update'});

      expect(updateButton).not.toBeDisabled();
    });
  });

  describe('Initial State', () => {
    it('should populate custom URL with current logo URL on mount', () => {
      render(
        <LogoUpdateModal open onClose={mockOnClose} currentLogoUrl={currentLogoUrl} onLogoUpdate={mockOnLogoUpdate} />,
      );

      const input = screen.getByPlaceholderText('Enter image URL');

      expect(input).toHaveValue(currentLogoUrl);
    });

    it('should reset state when modal is reopened', async () => {
      const {rerender} = render(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      const user = userEvent.setup();
      const input = screen.getByPlaceholderText('Enter image URL');
      await user.type(input, 'https://test.com/logo.png');

      // Close modal
      rerender(<LogoUpdateModal open={false} onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      // Reopen modal
      rerender(<LogoUpdateModal open onClose={mockOnClose} onLogoUpdate={mockOnLogoUpdate} />);

      await waitFor(() => {
        const reopenedInput = screen.getByPlaceholderText('Enter image URL');
        expect(reopenedInput).toHaveValue('');
      });
    });
  });
});
