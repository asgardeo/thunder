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
import ManageGroupsSection from '../ManageGroupsSection';
import type {Group} from '../../../../models/group';

// Mock the useGetOrganizationUnitGroups hook
const mockUseGetOrganizationUnitGroups = vi.fn();
vi.mock('../../../../api/useGetOrganizationUnitGroups', () => ({
  default: (id: string): unknown => mockUseGetOrganizationUnitGroups(id),
}));

// Mock useDataGridLocaleText hook
vi.mock('../../../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

describe('ManageGroupsSection', () => {
  const mockGroups: Group[] = [
    {id: 'group-1', name: 'Developers', organizationUnit: 'ou-123'},
    {id: 'group-2', name: 'Designers', organizationUnit: 'ou-123'},
    {id: 'group-3', name: 'Product Managers', organizationUnit: 'ou-123'},
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the manage groups section', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {groups: mockGroups},
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    await expect.element(page.getByText('Groups')).toBeInTheDocument();
    await expect.element(page.getByText('View groups belonging to this organization unit')).toBeInTheDocument();
  });

  it('should render data grid with groups', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {groups: mockGroups},
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    await expect.element(page.getByText('Developers')).toBeInTheDocument();
    await expect.element(page.getByText('Designers')).toBeInTheDocument();
    await expect.element(page.getByText('Product Managers')).toBeInTheDocument();
  });

  it('should render column headers', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {groups: mockGroups},
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('columnheader', {name: 'Group Name'})).toBeInTheDocument();
    await expect.element(page.getByRole('columnheader', {name: 'Group ID'})).toBeInTheDocument();
  });

  it('should show loading state', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: null,
      isLoading: true,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    // DataGrid shows loading overlay when isLoading is true
  });

  it('should handle empty groups list', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {groups: []},
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    // Grid should show "No rows" message
  });

  it('should handle null groups data', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: null,
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
  });

  it('should call useGetOrganizationUnitGroups with correct ID', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {groups: mockGroups},
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-456" />);

    expect(mockUseGetOrganizationUnitGroups).toHaveBeenCalledWith('ou-456');
  });

  it('should render group IDs correctly', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {groups: mockGroups},
      isLoading: false,
    });

    await renderWithProviders(<ManageGroupsSection organizationUnitId="ou-123" />);

    expect((page.getByText('group-1').all()).length).toBeGreaterThan(0);
    expect((page.getByText('group-2').all()).length).toBeGreaterThan(0);
    expect((page.getByText('group-3').all()).length).toBeGreaterThan(0);
  });
});
