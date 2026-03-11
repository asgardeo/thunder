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
import {useGetThemes} from '@thunder/shared-design';
import type {UseQueryResult} from '@tanstack/react-query';
import type {ThemeListResponse} from '@thunder/shared-design';
import AppearanceSection from '../AppearanceSection';
import type {Application} from '../../../../models/application';

vi.mock('@thunder/shared-design', () => ({
  useGetThemes: vi.fn(),
}));

describe('AppearanceSection', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    theme_id: 'theme-1',
  } as Application;

  const mockThemes = [
    {id: 'theme-1', displayName: 'Default Theme'},
    {id: 'theme-2', displayName: 'Dark Theme'},
    {id: 'theme-3', displayName: 'Light Theme'},
  ];

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    mockOnFieldChange.mockClear();
    vi.mocked(useGetThemes).mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    } as UseQueryResult<ThemeListResponse>);
  });

  describe('Rendering', () => {
    it('should render the appearance section', async () => {
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('heading', {name: 'Appearance'})).toBeInTheDocument();
      await expect.element(page.getByText('Customize the visual appearance of your application.')).toBeInTheDocument();
    });

    it('should render theme autocomplete field', async () => {
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Theme', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('Select a Theme')).toBeInTheDocument();
    });

    it('should display helper text', async () => {
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Choose a theme to customize authentication pages. Select the Default Theme (shared across all applications) or pick an app-specific theme.')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading indicator when themes are loading', async () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as UseQueryResult<ThemeListResponse>);

      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when themes are loaded', async () => {
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Theme Selection', () => {
    it('should display current theme from application', async () => {
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('Default Theme');
    });

    it('should prioritize editedApp theme_id over application', async () => {
      const editedApp = {
        theme_id: 'theme-2',
      };

      await render(
        <AppearanceSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />,
      );

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('Dark Theme');
    });

    it('should show all available themes in dropdown', async () => {
      
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      const listbox = page.getByRole('listbox');
      expect(listbox.getByText('Default Theme')).toBeInTheDocument();
      expect(listbox.getByText('Dark Theme')).toBeInTheDocument();
      expect(listbox.getByText('Light Theme')).toBeInTheDocument();
    });

    it('should call onFieldChange when theme is changed', async () => {
      
      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      const listbox = page.getByRole('listbox');
      const darkThemeOption = listbox.getByText('Dark Theme');
      await userEvent.click(darkThemeOption);

      expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', 'theme-2');
    });

    it('should handle clearing theme selection', async () => {

      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = page.getByRole('combobox');
      await userEvent.hover(autocomplete);

      const clearButton = page.getByRole('button', {name: 'Clear'});
      await userEvent.click(clearButton);
      expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', '');
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing theme_id in application', async () => {
      const appWithoutTheme: Partial<Application> = {...mockApplication};
      delete appWithoutTheme.theme_id;

      await render(
        <AppearanceSection
          application={appWithoutTheme as Application}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('');
    });

    it('should handle empty themes list', async () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: []},
        isLoading: false,
      } as unknown as UseQueryResult<ThemeListResponse>);

      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });

    it('should handle undefined themes data', async () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as UseQueryResult<ThemeListResponse>);

      await render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('combobox')).toBeInTheDocument();
    });

    it('should handle theme_id not found in themes list', async () => {
      const appWithInvalidTheme = {...mockApplication, theme_id: 'non-existent-id'};

      await render(<AppearanceSection application={appWithInvalidTheme} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('');
    });
  });
});
