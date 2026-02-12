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
import {screen, waitFor, renderWithProviders} from '@thunder/test-utils';
import EditGroups from '../groups/EditGroups';
import type {GroupListResponse} from '../../../types/organization-units';

// Mock the API hook
const mockUseGetOrganizationUnitGroups = vi.fn();
vi.mock('../../../api/useGetOrganizationUnitGroups', () => ({
  default: () => mockUseGetOrganizationUnitGroups() as {data: GroupListResponse | undefined; isLoading: boolean},
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
        'organizationUnits:view.groups.title': 'Groups',
        'organizationUnits:view.groups.subtitle': 'Groups associated with this organization unit',
        'organizationUnits:view.groups.columns.name': 'Group Name',
        'organizationUnits:view.groups.columns.id': 'Group ID',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('EditGroups', () => {
  const mockGroupsData: GroupListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    groups: [
      {id: 'group-1', name: 'Admin Group', organizationUnit: 'ou-123'},
      {id: 'group-2', name: 'User Group', organizationUnit: 'ou-123'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: mockGroupsData,
      isLoading: false,
    });
  });

  it('should render title and subtitle', () => {
    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    expect(screen.getByText('Groups')).toBeInTheDocument();
    expect(screen.getByText('Groups associated with this organization unit')).toBeInTheDocument();
  });

  it('should render DataGrid with groups', async () => {
    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('Admin Group')).toBeInTheDocument();
      expect(screen.getByText('User Group')).toBeInTheDocument();
    });
  });

  it('should display group IDs', async () => {
    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('group-1')).toBeInTheDocument();
      expect(screen.getByText('group-2')).toBeInTheDocument();
    });
  });

  it('should render empty grid when no groups exist', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        groups: [],
      },
      isLoading: false,
    });

    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.queryByText('Admin Group')).not.toBeInTheDocument();
    });
  });

  it('should pass loading state to DataGrid', () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: undefined,
      isLoading: true,
    });

    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    // Component should render without errors when loading
    expect(screen.getByText('Groups')).toBeInTheDocument();
  });

  it('should handle undefined data gracefully', async () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: undefined,
      isLoading: false,
    });

    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    // Should render without errors
    expect(screen.getByText('Groups')).toBeInTheDocument();
  });

  it('should render column headers', async () => {
    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('Group Name')).toBeInTheDocument();
      expect(screen.getByText('Group ID')).toBeInTheDocument();
    });
  });

  it('should render avatars for each group', async () => {
    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    await waitFor(() => {
      expect(screen.getByText('Admin Group')).toBeInTheDocument();
    });

    const avatars = document.querySelectorAll('.MuiAvatar-root');
    expect(avatars.length).toBeGreaterThan(0);
  });

  it('should handle null groups array gracefully', () => {
    mockUseGetOrganizationUnitGroups.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        groups: null as unknown as [],
      },
      isLoading: false,
    });

    renderWithProviders(<EditGroups organizationUnitId="ou-123" />);

    // Should render without errors - nullish coalescing handles null
    expect(screen.getByText('Groups')).toBeInTheDocument();
  });
});
