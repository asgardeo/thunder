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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import ManageChildOrganizationUnitSection from '../ManageChildOrganizationUnitSection';
import type {OrganizationUnit} from '../../../../models/organization-unit';

// Mock the useGetChildOrganizationUnits hook
const mockUseGetChildOrganizationUnits = vi.fn();
vi.mock('../../../../api/useGetChildOrganizationUnits', () => ({
  default: (id: string): unknown => mockUseGetChildOrganizationUnits(id),
}));

// Mock useDataGridLocaleText hook
vi.mock('../../../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
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

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    error: vi.fn(),
  }),
}));


describe('ManageChildOrganizationUnitSection', () => {
  const mockChildOUs: OrganizationUnit[] = [
    {
      id: 'ou-child-1',
      handle: 'frontend',
      name: 'Frontend Team',
      description: 'Frontend development team',
      parent: 'ou-parent',
    },
    {
      id: 'ou-child-2',
      handle: 'backend',
      name: 'Backend Team',
      description: 'Backend development team',
      parent: 'ou-parent',
    },
    {
      id: 'ou-child-3',
      handle: 'devops',
      name: 'DevOps Team',
      description: null,
      parent: 'ou-parent',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);
  });

  it('should render the manage child OUs section', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByText('Child Organization Units')).toBeInTheDocument();
    await expect.element(page.getByText('View and manage child organization units under this OU')).toBeInTheDocument();
  });

  it('should render data grid with child OUs', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    await expect.element(page.getByText('Frontend Team')).toBeInTheDocument();
    await expect.element(page.getByText('Backend Team')).toBeInTheDocument();
    await expect.element(page.getByText('DevOps Team')).toBeInTheDocument();
  });

  it('should render column headers', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByRole('columnheader', {name: 'Name'})).toBeInTheDocument();
    await expect.element(page.getByRole('columnheader', {name: 'Handle'})).toBeInTheDocument();
    await expect.element(page.getByRole('columnheader', {name: 'Description'})).toBeInTheDocument();
  });

  it('should show loading state', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: null,
      isLoading: true,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    // DataGrid shows loading overlay when isLoading is true
  });

  it('should handle empty child OUs list', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: []},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
    // Grid should show "No rows" message
  });

  it('should handle null child OUs data', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: null,
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByRole('grid')).toBeInTheDocument();
  });

  it('should call useGetChildOrganizationUnits with correct ID', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-456" organizationUnitName="Engineering" />,
    );

    expect(mockUseGetChildOrganizationUnits).toHaveBeenCalledWith('ou-456');
  });

  it('should render handles correctly', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByText('frontend')).toBeInTheDocument();
    await expect.element(page.getByText('backend')).toBeInTheDocument();
    await expect.element(page.getByText('devops')).toBeInTheDocument();
  });

  it('should render descriptions correctly', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await expect.element(page.getByText('Frontend development team')).toBeInTheDocument();
    await expect.element(page.getByText('Backend development team')).toBeInTheDocument();
  });

  it('should show "-" for null description', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    // The third OU has null description, should show "-"
    const cells = page.getByText('-').all();
    expect((cells).length).toBeGreaterThan(0);
  });

  it('should navigate to child OU when row is clicked', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    // Get the grid cell with the text "Frontend Team"
    await userEvent.click(page.getByText('Frontend Team'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/organization-units/ou-child-1',
        expect.objectContaining({
          state: {
            fromOU: {
              id: 'ou-parent',
              name: 'Engineering',
            },
          },
        }),
      );
    });
  });

  it('should include navigation state when navigating to child OU', async () => {
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Product Team" />,
    );

    await userEvent.click(page.getByText('Backend Team'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(
        '/organization-units/ou-child-2',
        expect.objectContaining({
          state: {
            fromOU: {
              id: 'ou-parent',
              name: 'Product Team',
            },
          },
        }),
      );
    });
  });

  it('should handle navigation errors gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));
    mockUseGetChildOrganizationUnits.mockReturnValue({
      data: {organizationUnits: mockChildOUs},
      isLoading: false,
    });

    await renderWithProviders(
      <ManageChildOrganizationUnitSection organizationUnitId="ou-parent" organizationUnitName="Engineering" />,
    );

    await userEvent.click(page.getByText('Frontend Team'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });

    // Should not throw error - error is logged
  });
});
