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
import SelectCountry from '../SelectCountry';

const mockCountries = [
  {name: 'France', regionCode: 'FR', flag: '🇫🇷'},
  {name: 'Germany', regionCode: 'DE', flag: '🇩🇪'},
  {name: 'Japan', regionCode: 'JP', flag: '🇯🇵'},
];

vi.mock('@thunder/i18n', () => ({
  buildCountryOptions: () => mockCountries,
}));

const defaultProps = {
  selectedCountry: null,
  onCountryChange: vi.fn(),
  onReadyChange: vi.fn(),
};

describe('SelectCountry', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders the step title and subtitle', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await expect.element(page.getByText('Choose a Country')).toBeInTheDocument();
      await expect.element(
        page.getByText('Select the country for the language you want to add.'),
      ).toBeInTheDocument();
    });

    it('renders the country autocomplete label', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await expect.element(page.getByText('Country')).toBeInTheDocument();
    });

    it('renders the helper tip', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await expect.element(
        page.getByText(
          'Country name will be used to derive a BCP 47 compliant locale code for the language.',
        ),
      ).toBeInTheDocument();
    });

    it('renders the autocomplete combobox', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });
  });

  describe('Options', () => {
    it('shows all country options when the dropdown is opened', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await userEvent.click(page.getByRole('combobox'));

      await expect.element(page.getByText('France')).toBeInTheDocument();
      await expect.element(page.getByText('Germany')).toBeInTheDocument();
      await expect.element(page.getByText('Japan')).toBeInTheDocument();
    });

    it('shows the region code chip for each option', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await userEvent.click(page.getByRole('combobox'));

      await expect.element(page.getByText('FR')).toBeInTheDocument();
      await expect.element(page.getByText('DE')).toBeInTheDocument();
    });

    it('filters options by country name', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await userEvent.type(page.getByRole('combobox'), 'Ger');

      await expect.element(page.getByText('Germany')).toBeInTheDocument();
      await expect.element(page.getByText('France')).not.toBeInTheDocument();
    });

    it('filters options by region code', async () => {
      await renderWithProviders(<SelectCountry {...defaultProps} />);

      await userEvent.type(page.getByRole('combobox'), 'JP');

      await expect.element(page.getByText('Japan')).toBeInTheDocument();
      await expect.element(page.getByText('France')).not.toBeInTheDocument();
    });
  });

  describe('onReadyChange', () => {
    it('calls onReadyChange(false) on mount when no country is selected', async () => {
      const onReadyChange = vi.fn();

      await renderWithProviders(<SelectCountry {...defaultProps} onReadyChange={onReadyChange} selectedCountry={null} />);

      expect(onReadyChange).toHaveBeenCalledWith(false);
    });

    it('calls onReadyChange(true) on mount when a country is already selected', async () => {
      const onReadyChange = vi.fn();

      await renderWithProviders(
        <SelectCountry {...defaultProps} onReadyChange={onReadyChange} selectedCountry={mockCountries[0]} />,
      );

      expect(onReadyChange).toHaveBeenCalledWith(true);
    });
  });

  describe('Interaction', () => {
    it('calls onCountryChange with the selected country when an option is clicked', async () => {
      const onCountryChange = vi.fn();

      await renderWithProviders(<SelectCountry {...defaultProps} onCountryChange={onCountryChange} />);

      await userEvent.click(page.getByRole('combobox'));
      await userEvent.click(page.getByText('France'));

      expect(onCountryChange).toHaveBeenCalledWith(mockCountries[0]);
    });

    it('calls onCountryChange(null) when the selection is cleared', async () => {
      const onCountryChange = vi.fn();

      await renderWithProviders(
        <SelectCountry
          {...defaultProps}
          selectedCountry={mockCountries[0]}
          onCountryChange={onCountryChange}
        />,
      );

      await userEvent.clear(page.getByRole('combobox'));

      expect(onCountryChange).toHaveBeenCalledWith(null);
    });
  });
});
