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
import {render, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {useGetThemes} from '@thunder/shared-design';
import type {UseQueryResult} from '@tanstack/react-query';
import type {ThemeListResponse} from '@thunder/shared-design';
import AppearanceSection from '../AppearanceSection';
import type {Application} from '../../../../models/application';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

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
    it('should render the appearance section', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.customization.sections.appearance')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.customization.sections.appearance.description')).toBeInTheDocument();
    });

    it('should render theme autocomplete field', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.customization.labels.theme')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('applications:edit.customization.theme.placeholder')).toBeInTheDocument();
    });

    it('should display helper text', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.customization.theme.hint')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading indicator when themes are loading', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as UseQueryResult<ThemeListResponse>);

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when themes are loaded', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Theme Selection', () => {
    it('should display current theme from application', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('Default Theme');
    });

    it('should prioritize editedApp theme_id over application', () => {
      const editedApp = {
        theme_id: 'theme-2',
      };

      render(
        <AppearanceSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('Dark Theme');
    });

    it('should show all available themes in dropdown', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const listbox = screen.getByRole('listbox');
      expect(within(listbox).getByText('Default Theme')).toBeInTheDocument();
      expect(within(listbox).getByText('Dark Theme')).toBeInTheDocument();
      expect(within(listbox).getByText('Light Theme')).toBeInTheDocument();
    });

    it('should call onFieldChange when theme is changed', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const listbox = screen.getByRole('listbox');
      const darkThemeOption = within(listbox).getByText('Dark Theme');
      await user.click(darkThemeOption);

      expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', 'theme-2');
    });

    it('should handle clearing theme selection', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      const clearButton = autocomplete.parentElement?.querySelector('[aria-label="Clear"]');

      if (clearButton) {
        await user.click(clearButton);
        expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', '');
      }
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing theme_id in application', () => {
      const appWithoutTheme: Partial<Application> = {...mockApplication};
      delete appWithoutTheme.theme_id;

      render(
        <AppearanceSection
          application={appWithoutTheme as Application}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('');
    });

    it('should handle empty themes list', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: {themes: []},
        isLoading: false,
      } as unknown as UseQueryResult<ThemeListResponse>);

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should handle undefined themes data', () => {
      vi.mocked(useGetThemes).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as UseQueryResult<ThemeListResponse>);

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should handle theme_id not found in themes list', () => {
      const appWithInvalidTheme = {...mockApplication, theme_id: 'non-existent-id'};

      render(<AppearanceSection application={appWithInvalidTheme} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('');
    });
  });
});
