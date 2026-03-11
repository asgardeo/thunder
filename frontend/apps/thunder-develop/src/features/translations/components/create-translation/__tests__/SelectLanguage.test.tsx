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
import SelectLanguage from '../SelectLanguage';

const mockLocales = [
  {code: 'fr-FR', displayName: 'French (France)', flag: '🇫🇷'},
  {code: 'fr-BE', displayName: 'French (Belgium)', flag: '🇧🇪'},
  {code: 'fr-CA', displayName: 'French (Canada)', flag: '🇨🇦'},
];

vi.mock('@thunder/i18n', () => ({
  buildLocaleOptions: vi.fn(() => mockLocales),
}));

const selectedCountry = {name: 'France', regionCode: 'FR', flag: '🇫🇷'};

const defaultProps = {
  selectedCountry,
  selectedLocale: null,
  onLocaleChange: vi.fn(),
  onReadyChange: vi.fn(),
};

describe('SelectLanguage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the step title and subtitle', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await expect.element(page.getByText('Choose a Language')).toBeInTheDocument();
      // subtitle has interpolation: 'Select the language variant spoken in {{country}}.'
      // with country = 'France' it becomes 'Select the language variant spoken in France.'
      await expect.element(
        page.getByText('Select the language variant spoken in France.'),
      ).toBeInTheDocument();
    });

    it('renders the language autocomplete label', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await expect.element(page.getByText('Language')).toBeInTheDocument();
    });

    it('renders the helper tip', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await expect.element(
        page.getByText(
          'Language picked here together with the country selection will determine the BCP 47 compliant locale code.',
        ),
      ).toBeInTheDocument();
    });

    it('renders the autocomplete combobox', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('Options', () => {
    it('shows all locale options when the dropdown is opened', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await userEvent.click(page.getByRole('combobox'));

      await expect.element(page.getByText('French (France)')).toBeInTheDocument();
      await expect.element(page.getByText('French (Belgium)')).toBeInTheDocument();
      await expect.element(page.getByText('French (Canada)')).toBeInTheDocument();
    });

    it('shows the BCP 47 code chip for each option', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await userEvent.click(page.getByRole('combobox'));

      await expect.element(page.getByText('fr-FR')).toBeInTheDocument();
      await expect.element(page.getByText('fr-BE')).toBeInTheDocument();
    });

    it('filters options by display name', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await userEvent.type(page.getByRole('combobox'), 'Belgium');

      await expect.element(page.getByText('French (Belgium)')).toBeInTheDocument();
      await expect.element(page.getByText('French (France)')).not.toBeInTheDocument();
    });

    it('filters options by locale code', async () => {
      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      await userEvent.type(page.getByRole('combobox'), 'fr-CA');

      await expect.element(page.getByText('French (Canada)')).toBeInTheDocument();
      await expect.element(page.getByText('French (Belgium)')).not.toBeInTheDocument();
    });
  });

  describe('buildLocaleOptions', () => {
    it('calls buildLocaleOptions with the selected country regionCode', async () => {
      const {buildLocaleOptions} = await import('@thunder/i18n');

      await renderWithProviders(<SelectLanguage {...defaultProps} />);

      expect(buildLocaleOptions).toHaveBeenCalledWith('FR');
    });
  });

  describe('onReadyChange', () => {
    it('calls onReadyChange(false) on mount when no locale is selected', async () => {
      const onReadyChange = vi.fn();

      await renderWithProviders(<SelectLanguage {...defaultProps} onReadyChange={onReadyChange} selectedLocale={null} />);

      expect(onReadyChange).toHaveBeenCalledWith(false);
    });

    it('calls onReadyChange(true) on mount when a locale is already selected', async () => {
      const onReadyChange = vi.fn();

      await renderWithProviders(
        <SelectLanguage {...defaultProps} onReadyChange={onReadyChange} selectedLocale={mockLocales[0]} />,
      );

      expect(onReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('Interaction', () => {
    it('calls onLocaleChange with the selected locale when an option is clicked', async () => {
      const onLocaleChange = vi.fn();

      await renderWithProviders(<SelectLanguage {...defaultProps} onLocaleChange={onLocaleChange} />);

      await userEvent.click(page.getByRole('combobox'));
      await userEvent.click(page.getByText('French (France)'));

      expect(onLocaleChange).toHaveBeenCalledWith(mockLocales[0]);
    });

    it('calls onLocaleChange(null) when the selection is cleared', async () => {
      const onLocaleChange = vi.fn();

      await renderWithProviders(
        <SelectLanguage {...defaultProps} selectedLocale={mockLocales[0]} onLocaleChange={onLocaleChange} />,
      );

      await userEvent.clear(page.getByRole('combobox'));

      expect(onLocaleChange).toHaveBeenCalledWith(null);
    });
  });
});
