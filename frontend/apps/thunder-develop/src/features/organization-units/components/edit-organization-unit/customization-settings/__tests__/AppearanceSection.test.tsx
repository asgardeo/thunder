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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import AppearanceSection from '../AppearanceSection';
import type {OrganizationUnit} from '../../../../models/organization-unit';

// Mock useGetThemes hook
const mockUseGetThemes = vi.fn();
vi.mock('@thunder/shared-design', () => ({
  useGetThemes: (): unknown => mockUseGetThemes(),
}));

describe('AppearanceSection', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-123',
    handle: 'engineering',
    name: 'Engineering',
    description: 'Engineering department',
    parent: null,
    theme_id: 'default-theme',
  };

  const mockThemes = [
    {id: 'default-theme', displayName: 'Default Theme'},
    {id: 'dark-theme', displayName: 'Dark Theme'},
    {id: 'light-theme', displayName: 'Light Theme'},
  ];

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the appearance section', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByText('Appearance')).toBeInTheDocument();
    await expect.element(page.getByText('Customize the look and feel of this organization unit.')).toBeInTheDocument();
  });

  it('should render theme label', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByText('Theme').first()).toBeInTheDocument();
  });

  it('should show loading spinner when themes are loading', async () => {
    mockUseGetThemes.mockReturnValue({
      data: null,
      isLoading: true,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render autocomplete with theme options', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByPlaceholder('Select a theme')).toBeInTheDocument();
  });

  it('should display current theme from organizationUnit', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByRole('combobox')).toHaveValue('Default Theme');
  });

  it('should display edited theme when available', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    const editedOU: Partial<OrganizationUnit> = {
      theme_id: 'dark-theme',
    };

    await renderWithProviders(
      <AppearanceSection
        organizationUnit={mockOrganizationUnit}
        editedOU={editedOU}
        onFieldChange={mockOnFieldChange}
      />,
    );

    await expect.element(page.getByRole('combobox')).toHaveValue('Dark Theme');
  });

  it('should call onFieldChange when theme is selected', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    const autocomplete = page.getByRole('combobox');
    await userEvent.click(autocomplete);

    await expect.element(page.getByText('Light Theme')).toBeInTheDocument();

    await userEvent.click(page.getByText('Light Theme'));

    await vi.waitFor(() => {
      expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', 'light-theme');
    });
  });

  it('should call onFieldChange with empty string when theme is cleared', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    // Hover over the autocomplete to make the clear button visible
    await userEvent.hover(page.getByRole('combobox'));
    const clearButton = page.getByRole('button', {name: 'Clear'});

    await userEvent.click(clearButton);

    await vi.waitFor(() => {
      expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', '');
    });
  });

  it('should handle empty themes list', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: []},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByPlaceholder('Select a theme')).toBeInTheDocument();
  });

  it('should handle null themes data', async () => {
    mockUseGetThemes.mockReturnValue({
      data: null,
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByPlaceholder('Select a theme')).toBeInTheDocument();
  });

  it('should render helper text', async () => {
    mockUseGetThemes.mockReturnValue({
      data: {themes: mockThemes},
      isLoading: false,
    });

    await renderWithProviders(
      <AppearanceSection organizationUnit={mockOrganizationUnit} editedOU={{}} onFieldChange={mockOnFieldChange} />,
    );

    await expect.element(page.getByText('The theme applied to this organization unit.')).toBeInTheDocument();
  });
});
