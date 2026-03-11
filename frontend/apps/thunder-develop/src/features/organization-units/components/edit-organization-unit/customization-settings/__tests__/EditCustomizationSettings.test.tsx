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
import EditCustomizationSettings from '../EditCustomizationSettings';
import type {OrganizationUnit} from '../../../../models/organization-unit';

// Mock child components
vi.mock('../AppearanceSection', () => ({
  default: ({
    organizationUnit,
    editedOU,
    onFieldChange,
  }: {
    organizationUnit: OrganizationUnit;
    editedOU: Partial<OrganizationUnit>;
    onFieldChange: (field: keyof OrganizationUnit, value: unknown) => void;
  }) => (
    <div data-testid="appearance-section">
      AppearanceSection - {organizationUnit.name}
      <button type="button" onClick={() => onFieldChange('theme_id', 'new-theme')}>
        Change Theme
      </button>
      <span>Edited Theme: {editedOU.theme_id ?? 'none'}</span>
    </div>
  ),
}));

describe('EditCustomizationSettings', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-123',
    handle: 'engineering',
    name: 'Engineering',
    description: 'Engineering department',
    parent: null,
  };

  const mockEditedOU: Partial<OrganizationUnit> = {};
  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render AppearanceSection', async () => {
    await renderWithProviders(
      <EditCustomizationSettings
        organizationUnit={mockOrganizationUnit}
        editedOU={mockEditedOU}
        onFieldChange={mockOnFieldChange}
      />,
    );

    await expect.element(page.getByTestId('appearance-section')).toBeInTheDocument();
  });

  it('should pass organizationUnit to AppearanceSection', async () => {
    await renderWithProviders(
      <EditCustomizationSettings
        organizationUnit={mockOrganizationUnit}
        editedOU={mockEditedOU}
        onFieldChange={mockOnFieldChange}
      />,
    );

    await expect.element(page.getByText(/AppearanceSection - Engineering/)).toBeInTheDocument();
  });

  it('should pass editedOU to AppearanceSection', async () => {
    const editedOU: Partial<OrganizationUnit> = {
      theme_id: 'custom-theme',
    };

    await renderWithProviders(
      <EditCustomizationSettings
        organizationUnit={mockOrganizationUnit}
        editedOU={editedOU}
        onFieldChange={mockOnFieldChange}
      />,
    );

    await expect.element(page.getByText('Edited Theme: custom-theme')).toBeInTheDocument();
  });

  it('should pass onFieldChange to AppearanceSection', async () => {
    await renderWithProviders(
      <EditCustomizationSettings
        organizationUnit={mockOrganizationUnit}
        editedOU={mockEditedOU}
        onFieldChange={mockOnFieldChange}
      />,
    );

    await userEvent.click(page.getByText('Change Theme'));

    expect(mockOnFieldChange).toHaveBeenCalledWith('theme_id', 'new-theme');
  });

  it('should handle empty editedOU', async () => {
    await renderWithProviders(
      <EditCustomizationSettings
        organizationUnit={mockOrganizationUnit}
        editedOU={{}}
        onFieldChange={mockOnFieldChange}
      />,
    );

    await expect.element(page.getByText('Edited Theme: none')).toBeInTheDocument();
  });
});
