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
import OrganizationUnitsListPage from '../OrganizationUnitsListPage';
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
const mockOUData: OrganizationUnitListResponse = {
  totalResults: 2,
  startIndex: 1,
  count: 2,
  organizationUnits: [
    {id: 'ou-1', handle: 'root', name: 'Root Organization', description: 'Root OU', parent: null},
    {id: 'ou-2', handle: 'child', name: 'Child Organization', description: null, parent: 'ou-1'},
  ],
};

vi.mock('../../api/useGetOrganizationUnits', () => ({
  default: () => ({
    data: mockOUData,
    isLoading: false,
    error: null,
  }),
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
        'organizationUnits:listing.title': 'Organization Units',
        'organizationUnits:listing.subtitle': 'Manage your organization units',
        'organizationUnits:listing.addOrganizationUnit': 'Add Organization Unit',
        'organizationUnits:listing.columns.name': 'Name',
        'organizationUnits:listing.columns.handle': 'Handle',
        'organizationUnits:listing.columns.description': 'Description',
        'organizationUnits:listing.columns.actions': 'Actions',
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

describe('OrganizationUnitsListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
  });

  it('should render page title', () => {
    renderWithProviders(<OrganizationUnitsListPage />);

    expect(screen.getByText('Organization Units')).toBeInTheDocument();
  });

  it('should render page subtitle', () => {
    renderWithProviders(<OrganizationUnitsListPage />);

    expect(screen.getByText('Manage your organization units')).toBeInTheDocument();
  });

  it('should render add organization unit button', () => {
    renderWithProviders(<OrganizationUnitsListPage />);

    expect(screen.getByText('Add Organization Unit')).toBeInTheDocument();
  });

  it('should navigate to create page when add button is clicked', async () => {
    renderWithProviders(<OrganizationUnitsListPage />);

    fireEvent.click(screen.getByText('Add Organization Unit'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });

  it('should render OrganizationUnitsList component', async () => {
    renderWithProviders(<OrganizationUnitsListPage />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
      expect(screen.getByText('Child Organization')).toBeInTheDocument();
    });
  });

  it('should have add button with Plus icon', () => {
    renderWithProviders(<OrganizationUnitsListPage />);

    const addButton = screen.getByText('Add Organization Unit').closest('button');
    expect(addButton).toBeInTheDocument();
    // Button should have contained variant (primary action style)
    expect(addButton).toHaveClass('MuiButton-contained');
  });

  it('should handle navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation error'));

    renderWithProviders(<OrganizationUnitsListPage />);

    fireEvent.click(screen.getByText('Add Organization Unit'));

    // Should not throw - error is logged internally
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });
});
