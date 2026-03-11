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
import InitializeLanguage from '../InitializeLanguage';

const defaultProps = {
  populateFromEnglish: true,
  onPopulateChange: vi.fn(),
  isCreating: false,
  progress: 0,
};

describe('InitializeLanguage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the step title and subtitle', async () => {
      await renderWithProviders(<InitializeLanguage {...defaultProps} />);

      await expect.element(page.getByText('Initialize Translations')).toBeInTheDocument();
      await expect.element(
        page.getByText('Choose how to populate the translation keys for this language.'),
      ).toBeInTheDocument();
    });

    it('renders both strategy card labels', async () => {
      await renderWithProviders(<InitializeLanguage {...defaultProps} />);

      await expect.element(page.getByText('Copy from English')).toBeInTheDocument();
      await expect.element(page.getByText('Start empty')).toBeInTheDocument();
    });

    it('renders both strategy card descriptions', async () => {
      await renderWithProviders(<InitializeLanguage {...defaultProps} />);

      await expect.element(
        page.getByText(
          'All keys will be pre-filled with English (en-US) text as a starting point. You can edit them afterwards.',
        ),
      ).toBeInTheDocument();
      await expect.element(
        page.getByText(
          'All keys will be created with empty values. Useful when you have your own translations ready to paste in.',
        ),
      ).toBeInTheDocument();
    });
  });

  describe('Progress indicator', () => {
    it('does not show the progress bar when not creating', async () => {
      await renderWithProviders(<InitializeLanguage {...defaultProps} isCreating={false} />);

      await expect.element(page.getByRole('progressbar')).not.toBeInTheDocument();
    });

    it('shows a spinner and progress bar while creating', async () => {
      await renderWithProviders(<InitializeLanguage {...defaultProps} isCreating progress={42} />);

      // LinearProgress and CircularProgress both use role="progressbar"
      expect(page.getByRole('progressbar').all()).toHaveLength(2);
    });

    it('displays the current progress percentage while creating', async () => {
      await renderWithProviders(<InitializeLanguage {...defaultProps} isCreating progress={65} />);

      await expect.element(page.getByText(/65%/)).toBeInTheDocument();
    });
  });

  describe('Strategy selection', () => {
    it('calls onPopulateChange(true) when the Copy from English card is clicked', async () => {
      const onPopulateChange = vi.fn();

      await renderWithProviders(
        <InitializeLanguage {...defaultProps} populateFromEnglish={false} onPopulateChange={onPopulateChange} />,
      );

      await userEvent.click(page.getByText('Copy from English'));

      expect(onPopulateChange).toHaveBeenCalledWith(true);
    });

    it('calls onPopulateChange(false) when the Start Empty card is clicked', async () => {
      const onPopulateChange = vi.fn();

      await renderWithProviders(
        <InitializeLanguage {...defaultProps} populateFromEnglish onPopulateChange={onPopulateChange} />,
      );

      await userEvent.click(page.getByText('Start empty'));

      expect(onPopulateChange).toHaveBeenCalledWith(false);
    });

    it('does not call onPopulateChange when a card is clicked while creating', async () => {
      const onPopulateChange = vi.fn();

      await renderWithProviders(<InitializeLanguage {...defaultProps} isCreating onPopulateChange={onPopulateChange} />);

      await userEvent.click(page.getByText('Start empty'));

      expect(onPopulateChange).not.toHaveBeenCalled();
    });
  });
});
