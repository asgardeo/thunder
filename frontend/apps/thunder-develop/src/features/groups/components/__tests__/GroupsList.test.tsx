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
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunder/test-utils';
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
}

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof OxygenUI>('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      ...(actual.DataGrid ?? {}),
      DataGrid: ({rows = [], columns = [], loading = false, onRowClick = undefined}: MockDataGridProps) => (
        <div data-testid="data-grid" data-loading={loading}>
          {rows.map((row) => (
            <div key={row.id} className="MuiDataGrid-row-container">
              <button
                type="button"
                className="MuiDataGrid-row"
                onClick={() => onRowClick?.({row}, {}, {})}
                data-testid={`row-${row.id}`}
              >
                {row.name}
              </button>
              {columns.map((column) => {
                if (!column?.field) return null;
                const value = column.valueGetter ? column.valueGetter(undefined, row) : row[column.field];
                const content = column.renderCell
                  ? column.renderCell({row, field: column.field, value, id: String(row.id)})
                  : value;
                return (
                  <span key={`${row.id}-${column.field}`} className="MuiDataGrid-cell">
                    {content as React.ReactNode}
                  </span>
                );
              })}
            </div>
          ))}
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

  it('should render groups in the data grid', () => {
    renderWithProviders(<GroupsList />);

    expect(screen.getByTestId('row-g1')).toHaveTextContent('Group One');
    expect(screen.getByTestId('row-g2')).toHaveTextContent('Group Two');
  });

  it('should show loading state', () => {
    mockUseGetGroups.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
    });
    renderWithProviders(<GroupsList />);

    expect(screen.getByTestId('data-grid')).toHaveAttribute('data-loading', 'true');
  });

  it('should show error state', () => {
    mockUseGetGroups.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Fetch failed'),
    });
    renderWithProviders(<GroupsList />);

    expect(screen.getByText('listing.error')).toBeInTheDocument();
    expect(screen.getByText('Fetch failed')).toBeInTheDocument();
  });

  it('should navigate to group on row click', async () => {
    const user = userEvent.setup();
    mockNavigate.mockResolvedValue(undefined);
    renderWithProviders(<GroupsList />);

    await user.click(screen.getByTestId('row-g1'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups/g1');
    });
  });

  it('should open actions menu on button click', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupsList />);

    const actionButtons = screen.getAllByRole('button', {name: /Open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
      expect(screen.getByText('Delete')).toBeInTheDocument();
    });
  });

  it('should navigate to group on view click', async () => {
    const user = userEvent.setup();
    mockNavigate.mockResolvedValue(undefined);
    renderWithProviders(<GroupsList />);

    const actionButtons = screen.getAllByRole('button', {name: /Open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
    });

    await user.click(screen.getByText('View'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups/g1');
    });
  });

  it('should open delete dialog on delete click', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupsList />);

    const actionButtons = screen.getAllByRole('button', {name: /Open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('Delete')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(screen.getByText('delete.title')).toBeInTheDocument();
    });
  });

  it('should close delete dialog when cancel is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupsList />);

    const actionButtons = screen.getAllByRole('button', {name: /Open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('Delete')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Delete'));

    await waitFor(() => {
      expect(screen.getByText('delete.title')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Cancel'));

    await waitFor(() => {
      expect(screen.queryByText('delete.title')).not.toBeInTheDocument();
    });
  });
});
