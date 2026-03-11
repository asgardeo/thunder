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
import ManageUsersSection from '../ManageUsersSection';
import type {ApiUser} from '../../../../../users/types/users';

// Mock the useGetOrganizationUnitUsers hook
const mockUseGetOrganizationUnitUsers = vi.fn();
vi.mock('../../../../api/useGetOrganizationUnitUsers', () => ({
  default: (id: string): unknown => mockUseGetOrganizationUnitUsers(id),
}));

// Mock useDataGridLocaleText hook
vi.mock('../../../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

describe('ManageUsersSection', () => {
  const mockUsers: ApiUser[] = [
    {id: 'user-1', type: 'internal', organizationUnit: 'ou-123'},
    {id: 'user-2', type: 'external', organizationUnit: 'ou-123'},
    {id: 'user-3', type: 'internal', organizationUnit: 'ou-123'},
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the manage users section', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {users: mockUsers},
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    await expect.element(page.getByText('Users')).toBeInTheDocument();
    await expect.element(page.getByText('View users belonging to this organization unit')).toBeInTheDocument();
  });

  it('should render data grid with users', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {users: mockUsers},
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    await expect.element(page.getByText('user-1')).toBeInTheDocument();
    await expect.element(page.getByText('user-2')).toBeInTheDocument();
    await expect.element(page.getByText('user-3')).toBeInTheDocument();
  });

  it('should render column headers', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {users: mockUsers},
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('columnheader', {name: 'User ID'})).toBeInTheDocument();
    await expect.element(page.getByRole('columnheader', {name: 'User Type'})).toBeInTheDocument();
  });

  it('should show loading state', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: null,
      isLoading: true,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    // DataGrid shows loading overlay when isLoading is true
  });

  it('should handle empty users list', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {users: []},
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    // Grid should show "No rows" message
  });

  it('should handle null users data', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: null,
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
  });

  it('should call useGetOrganizationUnitUsers with correct ID', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {users: mockUsers},
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-456" />);

    expect(mockUseGetOrganizationUnitUsers).toHaveBeenCalledWith('ou-456');
  });

  it('should render user type correctly', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {users: mockUsers},
      isLoading: false,
    });

    await renderWithProviders(<ManageUsersSection organizationUnitId="ou-123" />);

    expect((page.getByText('internal').all()).length).toBeGreaterThan(0);
    expect((page.getByText('external').all()).length).toBeGreaterThan(0);
  });
});
