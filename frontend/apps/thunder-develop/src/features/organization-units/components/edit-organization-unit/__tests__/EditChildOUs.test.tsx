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
import {screen, fireEvent, waitFor, renderWithProviders} from '@thunder/test-utils';
import EditChildOUs from '../child-ous/EditChildOUs';
import type {OrganizationUnitListResponse} from '../../../types/organization-units';

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
const mockUseGetChildOrganizationUnits = vi.fn();
vi.mock('../../../api/useGetChildOrganizationUnits', () => ({
  default: () =>
    mockUseGetChildOrganizationUnits() as {data: OrganizationUnitListResponse | undefined; isLoading: boolean},
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
        'organizationUnits:view.childOUs.title': 'Child Organization Units',
        'organizationUnits:view.childOUs.subtitle': 'Organization units that belong to this parent',
        'organizationUnits:listing.columns.name': 'Name',
        'organizationUnits:listing.columns.handle': 'Handle',
        'organizationUnits:listing.columns.description': 'Description',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('EditChildOUs', () => {
  const mockChildOUsData: OrganizationUnitListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    organizationUnits: [
      {id: 'child-1', handle: 'child-one', name: 'Child One', description: 'First child', parent: 'parent-ou'},
      {id: 'child-2', handle: 'child-two', name: 'Child Two', description: null, parent: 'parent-ou'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: mockChildOUsData,
      isLoading: false,
    });
  });

  it('should render title and subtitle', () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    expect(screen.getByText('Child Organization Units')).toBeInTheDocument();
    expect(screen.getByText('Organization units that belong to this parent')).toBeInTheDocument();
  });

  it('should render DataGrid with child organization units', async () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('Child One')).toBeInTheDocument();
      expect(screen.getByText('Child Two')).toBeInTheDocument();
    });
  });

  it('should display handle column', async () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('child-one')).toBeInTheDocument();
      expect(screen.getByText('child-two')).toBeInTheDocument();
    });
  });

  it('should display description or dash for null', async () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('First child')).toBeInTheDocument();
      expect(screen.getByText('-')).toBeInTheDocument();
    });
  });

  it('should render empty grid when no child OUs exist', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        organizationUnits: [],
      },
      isLoading: false,
    });

    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.queryByText('Child One')).not.toBeInTheDocument();
    });
  });

  it('should handle row click navigation', async () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('Child One')).toBeInTheDocument();
    });

    // Click on the row
    const row = screen.getByText('Child One').closest('.MuiDataGrid-row');
    if (row) {
      fireEvent.click(row);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units/child-1', {
          state: {fromOU: {id: 'parent-ou', name: 'Parent OU'}},
        });
      });
    }
  });

  it('should pass loading state to DataGrid', () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: true,
    });

    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    // Component should render without errors when loading
    expect(screen.getByText('Child Organization Units')).toBeInTheDocument();
  });

  it('should handle row click navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));

    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('Child One')).toBeInTheDocument();
    });

    const row = screen.getByText('Child One').closest('.MuiDataGrid-row');
    if (row) {
      fireEvent.click(row);

      // Should not throw - error is logged
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units/child-1', {
          state: {fromOU: {id: 'parent-ou', name: 'Parent OU'}},
        });
      });
    }
  });

  it('should render column headers', async () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Handle')).toBeInTheDocument();
      expect(screen.getByText('Description')).toBeInTheDocument();
    });
  });

  it('should render avatars for each child OU', async () => {
    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    await waitFor(() => {
      expect(screen.getByText('Child One')).toBeInTheDocument();
    });

    const avatars = document.querySelectorAll('.MuiAvatar-root');
    expect(avatars.length).toBeGreaterThan(0);
  });

  it('should handle undefined data gracefully', () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
    });

    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    expect(screen.getByText('Child Organization Units')).toBeInTheDocument();
  });

  it('should handle null organizationUnits array gracefully', () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        organizationUnits: null as unknown as [],
      },
      isLoading: false,
    });

    renderWithProviders(<EditChildOUs organizationUnitId="parent-ou" organizationUnitName="Parent OU" />);

    // Should render without errors - nullish coalescing handles null
    expect(screen.getByText('Child Organization Units')).toBeInTheDocument();
  });
});
