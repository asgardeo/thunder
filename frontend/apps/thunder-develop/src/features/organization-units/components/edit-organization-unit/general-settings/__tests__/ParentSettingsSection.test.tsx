/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
import {page} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import ParentSettingsSection from '../ParentSettingsSection';
import type {OrganizationUnit} from '../../../../models/organization-unit';

// Mock the useGetOrganizationUnit hook
const mockUseGetOrganizationUnit = vi.fn();
vi.mock('../../../../api/useGetOrganizationUnit', () => ({
  default: (id?: string, enabled?: boolean): unknown => mockUseGetOrganizationUnit(id, enabled),
}));

// Mock navigate function
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('ParentSettingsSection', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-child-123',
    handle: 'engineering-frontend',
    name: 'Frontend Engineering',
    description: 'Frontend team',
    parent: 'ou-parent-123',
  };

  const mockParentOU: OrganizationUnit = {
    id: 'ou-parent-123',
    handle: 'engineering',
    name: 'Engineering',
    description: 'Engineering department',
    parent: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the parent settings section', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockParentOU,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    await expect.element(page.getByText('Parent Organization Unit').first()).toBeInTheDocument();
    await expect.element(page.getByText('The parent organization unit in the hierarchy.')).toBeInTheDocument();
  });

  it('should show "Root Organization Unit" when no parent exists', async () => {
    const rootOU: OrganizationUnit = {
      ...mockOrganizationUnit,
      parent: null,
    };

    mockUseGetOrganizationUnit.mockReturnValue({
      data: null,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={rootOU} />);

    await expect.element(page.getByRole('textbox')).toHaveValue('Root Organization Unit');
    await expect.element(page.getByRole('textbox')).toHaveAttribute('readonly');
  });

  it('should show loading spinner while fetching parent', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: null,
      isLoading: true,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render parent name as link when parent is loaded', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockParentOU,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    const link = page.getByText('Engineering');
    await expect.element(link).toBeInTheDocument();
    await expect.element(link).toHaveAttribute('href', '/organization-units/ou-parent-123');
  });

  it('should render parent ID alongside parent name', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockParentOU,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    await expect.element(page.getByText('Engineering')).toBeInTheDocument();
    await expect.element(page.getByText('(ou-parent-123)')).toBeInTheDocument();
  });

  it('should include navigation state in parent link', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockParentOU,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    const link = page.getByText('Engineering');
    await expect.element(link).toHaveAttribute('data-state', JSON.stringify({
      fromOU: {
        id: 'ou-child-123',
        name: 'Frontend Engineering',
      },
    }));
  });

  it('should show raw parent ID when parent cannot be loaded', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: null,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    await expect.element(page.getByRole('textbox')).toHaveValue('ou-parent-123');
    await expect.element(page.getByRole('textbox')).toHaveAttribute('readonly');
  });

  it('should not fetch parent when parent is null', async () => {
    const rootOU: OrganizationUnit = {
      ...mockOrganizationUnit,
      parent: null,
    };

    await renderWithProviders(<ParentSettingsSection organizationUnit={rootOU} />);

    expect(mockUseGetOrganizationUnit).toHaveBeenCalledWith(undefined, false);
  });

  it('should fetch parent when parent ID exists', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockParentOU,
      isLoading: false,
    });

    await renderWithProviders(<ParentSettingsSection organizationUnit={mockOrganizationUnit} />);

    expect(mockUseGetOrganizationUnit).toHaveBeenCalledWith('ou-parent-123', true);
  });
});
