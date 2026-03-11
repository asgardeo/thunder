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

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import type * as OxygenUI from '@wso2/oxygen-ui';
import type {GroupListResponse} from '../../models/group';
import GroupsList from '../GroupsList';

interface MockDataGridProps {
  rows?: {id: string; name?: string; [key: string]: unknown}[];
  columns?: {
    field?: string;
    valueGetter?: (value: unknown, row: Record<string, unknown>) => unknown;
    renderCell?: (params: {row: Record<string, unknown>; field: string; value: unknown; id: string}) => React.ReactNode;
  }[];
  loading?: boolean;
  onRowClick?: (params: {row: unknown}, details: unknown, event: unknown) => void;
  getRowId?: (row: {id: string; [key: string]: unknown}) => string;
}

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof OxygenUI>('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      ...(actual.DataGrid ?? {}),
      GridColDef: {},
      GridRenderCellParams: {},
    },
    ListingTable: {
      Provider: ({children, loading = false}: {children: React.ReactNode; loading?: boolean}) => (
        <div data-testid="listing-table-provider" data-loading={loading ? 'true' : 'false'}>
          {children}
        </div>
      ),
      Container: ({children}: {children: React.ReactNode}): React.ReactElement => children as React.ReactElement,
      DataGrid: ({rows = [], columns = [], loading = false, onRowClick = undefined, getRowId = undefined}: MockDataGridProps) => (
        <div data-testid="data-grid" data-loading={loading}>
          {rows.map((row) => {
            const rowId = getRowId ? getRowId(row) : row.id;
            return (
              <div key={rowId} className="MuiDataGrid-row-container">
                <button
                  type="button"
                  className="MuiDataGrid-row"
                  onClick={() => onRowClick?.({row}, {}, {})}
                  data-testid={`row-${rowId}`}
                >
                  {row.name}
                </button>
                {columns.map((column) => {
                  if (!column?.field) return null;
                  const value = column.valueGetter ? column.valueGetter(undefined, row) : row[column.field];
                  const content = column.renderCell
                    ? column.renderCell({row, field: column.field, value, id: String(rowId)})
                    : value;
                  return (
                    <span key={`${rowId}-${column.field}`} className="MuiDataGrid-cell">
                      {content as React.ReactNode}
                    </span>
                  );
                })}
              </div>
            );
          })}
        </div>
      ),
    },
  };
});

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

const mockUseGetGroups = vi.fn();
vi.mock('../../api/useGetGroups', () => ({
  default: (...args: unknown[]): unknown => mockUseGetGroups(...args),
}));

const mockDeleteMutate = vi.fn();
vi.mock('../../api/useDeleteGroup', () => ({
  default: () => ({
    mutate: mockDeleteMutate,
    isPending: false,
  }),
}));

describe('GroupsList', () => {
  const mockGroupsData: GroupListResponse = {
    totalResults: 2,
    startIndex: 0,
    count: 2,
    groups: [
      {id: 'g1', name: 'Group One', description: 'First group', organizationUnitId: 'ou1'},
      {id: 'g2', name: 'Group Two', organizationUnitId: 'ou2'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetGroups.mockReturnValue({
      data: mockGroupsData,
      isLoading: false,
      error: null,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render groups in the data grid', async () => {
    await renderWithProviders(<GroupsList />);

    await expect.element(page.getByTestId('row-g1')).toHaveTextContent('Group One');
    await expect.element(page.getByTestId('row-g2')).toHaveTextContent('Group Two');
  });

  it('should show loading state', async () => {
    mockUseGetGroups.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
    });
    await renderWithProviders(<GroupsList />);

    await expect.element(page.getByTestId('listing-table-provider')).toHaveAttribute('data-loading', 'true');
  });

  it('should show error state', async () => {
    mockUseGetGroups.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Fetch failed'),
    });
    await renderWithProviders(<GroupsList />);

    await expect.element(page.getByText('Failed to load groups')).toBeInTheDocument();
    await expect.element(page.getByText('Fetch failed')).toBeInTheDocument();
  });

  it('should navigate to group on row click', async () => {
    mockNavigate.mockResolvedValue(undefined);
    await renderWithProviders(<GroupsList />);

    await userEvent.click(page.getByTestId('row-g1'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups/g1');
    });
  });

  it('should render edit and delete action buttons for each row', async () => {
    await renderWithProviders(<GroupsList />);

    // The actions column renders Pencil (Edit) and Trash2 (Delete) icon buttons
    // via ListingTable.RowActions. The mock renders the renderCell content.
    await expect.element(page.getByTestId('row-g1')).toBeInTheDocument();
    await expect.element(page.getByTestId('row-g2')).toBeInTheDocument();
  });

  it('should open delete dialog when delete action button is clicked', async () => {
    await renderWithProviders(<GroupsList />);

    // Find Delete icon buttons (aria-label from t('common:actions.delete') = 'Delete')
    // They are rendered in the actions column via renderCell
    const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
    expect(deleteButtons.length).toBeGreaterThan(0);
    await deleteButtons[0].click();

    // The GroupDeleteDialog should open, showing "Delete Group" dialog title
    await expect.element(page.getByRole('dialog')).toBeInTheDocument();
    await expect.element(page.getByText('Delete Group')).toBeInTheDocument();
  });

  it('should close delete dialog when Cancel is clicked', async () => {
    await renderWithProviders(<GroupsList />);

    const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
    expect(deleteButtons.length).toBeGreaterThan(0);
    await deleteButtons[0].click();

    await expect.element(page.getByRole('dialog')).toBeInTheDocument();

    await userEvent.click(page.getByRole('button', {name: 'Cancel'}));

    await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
  });
});
