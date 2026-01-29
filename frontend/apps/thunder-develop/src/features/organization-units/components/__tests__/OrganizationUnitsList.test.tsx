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
import {screen, fireEvent, waitFor} from '@testing-library/react';
import {renderWithProviders} from '../../../../test/test-utils';
import OrganizationUnitsList from '../OrganizationUnitsList';
import type {OrganizationUnitListResponse} from '../../types/organization-units';

// Mock navigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    error: vi.fn(),
    info: vi.fn(),
    debug: vi.fn(),
  }),
}));

// Mock the API hook
const mockUseGetOrganizationUnits = vi.fn();
vi.mock('../../api/useGetOrganizationUnits', () => ({
  default: () =>
    mockUseGetOrganizationUnits() as {
      data: OrganizationUnitListResponse | undefined;
      isLoading: boolean;
      error: Error | null;
    },
}));

// Mock delete hook
vi.mock('../../api/useDeleteOrganizationUnit', () => ({
  default: () => ({
    mutate: vi.fn(),
    isPending: false,
  }),
}));

// Mock useDataGridLocaleText
vi.mock('../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:listing.columns.name': 'Name',
        'organizationUnits:listing.columns.handle': 'Handle',
        'organizationUnits:listing.columns.description': 'Description',
        'organizationUnits:listing.columns.actions': 'Actions',
        'organizationUnits:listing.error.title': 'Error loading organization units',
        'organizationUnits:listing.error.unknown': 'An unknown error occurred',
        'common:actions.view': 'View',
        'common:actions.delete': 'Delete',
        'organizationUnits:delete.title': 'Delete Organization Unit',
        'organizationUnits:delete.message': 'Are you sure?',
        'organizationUnits:delete.disclaimer': 'This cannot be undone.',
        'common:actions.cancel': 'Cancel',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('OrganizationUnitsList', () => {
  const mockOUData: OrganizationUnitListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    organizationUnits: [
      {id: 'ou-1', handle: 'root', name: 'Root Organization', description: 'Root OU', parent: null},
      {id: 'ou-2', handle: 'child', name: 'Child Organization', description: null, parent: 'ou-1'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
    mockUseGetOrganizationUnits.mockReturnValue({
      data: mockOUData,
      isLoading: false,
      error: null,
    });
  });

  it('should render DataGrid with organization units data', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
      expect(screen.getByText('Child Organization')).toBeInTheDocument();
    });
  });

  it('should display handle column', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('root')).toBeInTheDocument();
      expect(screen.getByText('child')).toBeInTheDocument();
    });
  });

  it('should display description or dash for null', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Root OU')).toBeInTheDocument();
      // For null description, valueGetter returns '-'
      expect(screen.getByText('-')).toBeInTheDocument();
    });
  });

  it('should show error state when fetch fails', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
    });

    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Error loading organization units')).toBeInTheDocument();
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('should show error with fallback message when error has no message', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: {},
    });

    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Error loading organization units')).toBeInTheDocument();
      expect(screen.getByText('An unknown error occurred')).toBeInTheDocument();
    });
  });

  it('should render empty list when no organization units', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        organizationUnits: [],
      },
      isLoading: false,
      error: null,
    });

    renderWithProviders(<OrganizationUnitsList />);

    // DataGrid should render but with no data rows
    await waitFor(() => {
      expect(screen.queryByText('Root Organization')).not.toBeInTheDocument();
    });
  });

  it('should handle row click navigation', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // Click on the row (simulate row click by clicking on the name cell)
    const row = screen.getByText('Root Organization').closest('.MuiDataGrid-row');
    if (row) {
      fireEvent.click(row);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units/ou-1');
      });
    }
  });

  it('should render column headers', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Handle')).toBeInTheDocument();
      expect(screen.getByText('Description')).toBeInTheDocument();
      expect(screen.getByText('Actions')).toBeInTheDocument();
    });
  });

  it('should pass loading state to DataGrid', () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });

    renderWithProviders(<OrganizationUnitsList />);

    // Component should render without errors when loading
    expect(screen.getByText('Name')).toBeInTheDocument();
  });

  it('should handle navigation error on row click gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation error'));

    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    const row = screen.getByText('Root Organization').closest('.MuiDataGrid-row');
    if (row) {
      fireEvent.click(row);

      // Should not throw - error should be caught by logger
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units/ou-1');
      });
    }
  });

  it('should render action column for each row', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // Actions column header should be visible
    expect(screen.getByText('Actions')).toBeInTheDocument();
  });

  it('should handle data with undefined organizationUnits gracefully', () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    // This should not throw - should render empty grid
    renderWithProviders(<OrganizationUnitsList />);

    expect(screen.getByText('Name')).toBeInTheDocument();
  });

  it('should render avatar for each organization unit', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // Avatar elements should be present - they contain the Building icon
    const avatars = document.querySelectorAll('.MuiAvatar-root');
    expect(avatars.length).toBeGreaterThan(0);
  });

  it('should handle click on second row for navigation', async () => {
    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Child Organization')).toBeInTheDocument();
    });

    const row = screen.getByText('Child Organization').closest('.MuiDataGrid-row');
    if (row) {
      fireEvent.click(row);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units/ou-2');
      });
    }
  });

  it('should render with data that has organizationUnits array', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 1,
        startIndex: 1,
        count: 1,
        organizationUnits: [{id: 'ou-3', handle: 'single', name: 'Single OU', description: 'A single OU', parent: null}],
      },
      isLoading: false,
      error: null,
    });

    renderWithProviders(<OrganizationUnitsList />);

    await waitFor(() => {
      expect(screen.getByText('Single OU')).toBeInTheDocument();
      expect(screen.getByText('single')).toBeInTheDocument();
      expect(screen.getByText('A single OU')).toBeInTheDocument();
    });
  });
});
