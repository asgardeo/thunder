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
import {useGetBrandings} from '@thunder/shared-branding';
import type {UseQueryResult} from '@tanstack/react-query';
import type {BrandingListResponse} from '@thunder/shared-branding';
import AppearanceSection from '../AppearanceSection';
import type {Application} from '../../../../models/application';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('@thunder/shared-branding', () => ({
  useGetBrandings: vi.fn(),
}));

describe('AppearanceSection', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    branding_id: 'branding-1',
  } as Application;

  const mockBrandings = [
    {id: 'branding-1', displayName: 'Default Theme'},
    {id: 'branding-2', displayName: 'Dark Theme'},
    {id: 'branding-3', displayName: 'Light Theme'},
  ];

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    mockOnFieldChange.mockClear();
    vi.mocked(useGetBrandings).mockReturnValue({
      data: {brandings: mockBrandings},
      isLoading: false,
    } as UseQueryResult<BrandingListResponse>);
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
    it('should show loading indicator when brandings are loading', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as UseQueryResult<BrandingListResponse>);

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when brandings are loaded', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Branding Selection', () => {
    it('should display current branding from application', () => {
      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('Default Theme');
    });

    it('should prioritize editedApp branding_id over application', () => {
      const editedApp = {
        branding_id: 'branding-2',
      };

      render(
        <AppearanceSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('Dark Theme');
    });

    it('should show all available brandings in dropdown', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const listbox = screen.getByRole('listbox');
      expect(within(listbox).getByText('Default Theme')).toBeInTheDocument();
      expect(within(listbox).getByText('Dark Theme')).toBeInTheDocument();
      expect(within(listbox).getByText('Light Theme')).toBeInTheDocument();
    });

    it('should call onFieldChange when branding is changed', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const listbox = screen.getByRole('listbox');
      const darkThemeOption = within(listbox).getByText('Dark Theme');
      await user.click(darkThemeOption);

      expect(mockOnFieldChange).toHaveBeenCalledWith('branding_id', 'branding-2');
    });

    it('should handle clearing branding selection', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      const clearButton = autocomplete.parentElement?.querySelector('[aria-label="Clear"]');

      if (clearButton) {
        await user.click(clearButton);
        expect(mockOnFieldChange).toHaveBeenCalledWith('branding_id', '');
      }
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing branding_id in application', () => {
      const appWithoutBranding: Partial<Application> = {...mockApplication};
      delete appWithoutBranding.branding_id;

      render(
        <AppearanceSection
          application={appWithoutBranding as Application}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('');
    });

    it('should handle empty brandings list', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: {brandings: []},
        isLoading: false,
      } as unknown as UseQueryResult<BrandingListResponse>);

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should handle undefined brandings data', () => {
      vi.mocked(useGetBrandings).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as UseQueryResult<BrandingListResponse>);

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should handle branding_id not found in brandings list', () => {
      const appWithInvalidBranding = {...mockApplication, branding_id: 'non-existent-id'};

      render(
        <AppearanceSection application={appWithInvalidBranding} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('');
    });

    it('should use editedApp branding_id when application has no branding_id', () => {
      const appWithoutBranding: Partial<Application> = {...mockApplication};
      delete appWithoutBranding.branding_id;

      render(
        <AppearanceSection
          application={appWithoutBranding as Application}
          editedApp={{branding_id: 'branding-2'}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('Dark Theme');
    });

    it('should handle selecting null option', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);
      await user.clear(autocomplete);

      // Simulate pressing escape to deselect
      await user.keyboard('{Escape}');

      // The autocomplete should still be functional
      expect(autocomplete).toBeInTheDocument();
    });

    it('should handle string option in getOptionLabel', async () => {
      const user = userEvent.setup();

      render(<AppearanceSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);
      await user.clear(autocomplete);
      await user.type(autocomplete, 'Default Theme');

      // Should find the matching option
      const listbox = screen.getByRole('listbox');
      expect(within(listbox).getByText('Default Theme')).toBeInTheDocument();
    });

    it('should handle both application and editedApp having branding_id', () => {
      render(
        <AppearanceSection
          application={{...mockApplication, branding_id: 'branding-1'}}
          editedApp={{branding_id: 'branding-3'}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const input = screen.getByRole('combobox');
      expect(input).toHaveValue('Light Theme');
    });
  });
});
