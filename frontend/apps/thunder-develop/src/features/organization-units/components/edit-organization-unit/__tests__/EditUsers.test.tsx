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
import {screen, waitFor} from '@testing-library/react';
import {renderWithProviders} from '../../../../../test/test-utils';
import EditUsers from '../users/EditUsers';
import type {UserListResponse} from '../../../../users/types/users';

// Mock the API hook
const mockUseGetOrganizationUnitUsers = vi.fn();
vi.mock('../../../api/useGetOrganizationUnitUsers', () => ({
  default: () => mockUseGetOrganizationUnitUsers() as {data: UserListResponse | undefined; isLoading: boolean},
}));

// Mock useDataGridLocaleText
vi.mock('../../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:view.users.title': 'Users',
        'organizationUnits:view.users.subtitle': 'Users belonging to this organization unit',
        'organizationUnits:view.users.columns.id': 'User ID',
        'organizationUnits:view.users.columns.type': 'User Type',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('EditUsers', () => {
  const mockUsersData: UserListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    users: [
      {id: 'user-1', organizationUnit: 'ou-123', type: 'employee'},
      {id: 'user-2', organizationUnit: 'ou-123', type: 'contractor'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: mockUsersData,
      isLoading: false,
    });
  });

  it('should render title and subtitle', () => {
    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('Users belonging to this organization unit')).toBeInTheDocument();
  });

  it('should render DataGrid with users', async () => {
    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('user-1')).toBeInTheDocument();
      expect(screen.getByText('user-2')).toBeInTheDocument();
    });
  });

  it('should display user types', async () => {
    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('employee')).toBeInTheDocument();
      expect(screen.getByText('contractor')).toBeInTheDocument();
    });
  });

  it('should render empty grid when no users exist', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        users: [],
      },
      isLoading: false,
    });

    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.queryByText('user-1')).not.toBeInTheDocument();
    });
  });

  it('should pass loading state to DataGrid', () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: undefined,
      isLoading: true,
    });

    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    // Component should render without errors when loading
    expect(screen.getByText('Users')).toBeInTheDocument();
  });

  it('should handle undefined data gracefully', async () => {
    mockUseGetOrganizationUnitUsers.mockReturnValue({
      data: undefined,
      isLoading: false,
    });

    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    // Should render without errors
    expect(screen.getByText('Users')).toBeInTheDocument();
  });

  it('should render column headers', async () => {
    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('User ID')).toBeInTheDocument();
      expect(screen.getByText('User Type')).toBeInTheDocument();
    });
  });

  it('should render avatars for each user', async () => {
    renderWithProviders(<EditUsers organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('user-1')).toBeInTheDocument();
    });

    const avatars = document.querySelectorAll('.MuiAvatar-root');
    expect(avatars.length).toBeGreaterThan(0);
  });
});
