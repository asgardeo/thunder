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
import userEvent from '@testing-library/user-event';
import {type ReactNode} from 'react';
import render from '@/test/test-utils';
import UserTypesList from './UserTypesList';
import type useGetUserTypesHook from '../api/useGetUserTypes';
import type useDeleteUserTypeHook from '../api/useDeleteUserType';
import type {UserSchemaListResponse, ApiError, UserSchemaListItem} from '../types/user-types';

const mockNavigate = vi.fn();
const mockRefetch = vi.fn<() => Promise<void>>();
const mockDeleteUserType = vi.fn();
const mockResetDeleteUserType = vi.fn();

type MockDataGridRow = UserSchemaListItem & Record<string, unknown>;

type MockDataGridColumn =
  | {
      field?: string;
      valueGetter?: (_value: unknown, row: MockDataGridRow) => unknown;
      renderCell?: (params: {row: MockDataGridRow; field: string; value: unknown; id: string}) => ReactNode;
    }
  | null
  | undefined;

interface MockDataGridProps {
  rows?: MockDataGridRow[];
  columns?: MockDataGridColumn[];
  loading?: boolean;
  onRowClick?: (params: {row: MockDataGridRow}, event: unknown, details: unknown) => void;
}

// Mock DataGrid to avoid CSS import issues
vi.mock('@mui/x-data-grid', () => ({
  DataGrid: ({rows = [], columns = [], loading = false, onRowClick = () => undefined}: MockDataGridProps) => (
    <div data-testid="data-grid" data-loading={loading}>
      {rows.map((row: MockDataGridRow) => {
        const rowId = String(row.id ?? '');
        return (
          <div key={row.id} className="MuiDataGrid-row-container">
            <button
              type="button"
              className="MuiDataGrid-row"
              onClick={() => onRowClick?.({row}, {} as unknown, {} as unknown)}
              data-testid={`row-${rowId}`}
            >
              {row.name}
            </button>
            {columns.map((column) => {
              if (!column?.field) return null;

              const fallbackValue = (row as Record<string, unknown>)[column.field];
              const value =
                typeof column.valueGetter === 'function' ? column.valueGetter(undefined, row) : fallbackValue;

              const params = {
                row,
                field: column.field,
                value,
                id: rowId,
              };

              const content = (typeof column.renderCell === 'function' ? column.renderCell(params) : value) as
                | ReactNode
                | null
                | undefined;

              if (content === null || content === undefined) {
                return null;
              }

              return (
                <span key={`${rowId}-${column.field}`} className="MuiDataGrid-cell">
                  {content}
                </span>
              );
            })}
          </div>
        );
      })}
    </div>
  ),
  GridColDef: {},
}));
/* eslint-enable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-unsafe-return, react/destructuring-assignment */

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock hooks
type UseGetUserTypesReturn = ReturnType<typeof useGetUserTypesHook>;
type UseDeleteUserTypeReturn = ReturnType<typeof useDeleteUserTypeHook>;

const mockUseGetUserTypes = vi.fn<() => UseGetUserTypesReturn>();
const mockUseDeleteUserType = vi.fn<() => UseDeleteUserTypeReturn>();

vi.mock('../api/useGetUserTypes', () => ({
  default: () => mockUseGetUserTypes(),
}));

vi.mock('../api/useDeleteUserType', () => ({
  default: () => mockUseDeleteUserType(),
}));

describe('UserTypesList', () => {
  const mockUserTypesData: UserSchemaListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    schemas: [
      {id: 'schema1', name: 'Employee Schema'},
      {id: 'schema2', name: 'Contractor Schema'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUserTypes.mockReturnValue({
      data: mockUserTypesData,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });
    mockUseDeleteUserType.mockReturnValue({
      deleteUserType: mockDeleteUserType,
      loading: false,
      error: null,
      reset: mockResetDeleteUserType,
    });
  });

  it('renders DataGrid with user types', () => {
    render(<UserTypesList />);

    expect(screen.getByTestId('row-schema1')).toHaveTextContent('Employee Schema');
    expect(screen.getByTestId('row-schema2')).toHaveTextContent('Contractor Schema');
  });

  it('displays loading state', () => {
    mockUseGetUserTypes.mockReturnValue({
      data: null,
      loading: true,
      error: null,
      refetch: mockRefetch,
    });

    render(<UserTypesList />);

    const grid = screen.getByTestId('data-grid');
    expect(grid).toHaveAttribute('data-loading', 'true');
  });

  it('displays error in snackbar', async () => {
    const error: ApiError = {
      code: 'ERROR_CODE',
      message: 'Failed to load user types',
      description: 'Error description',
    };

    mockUseGetUserTypes.mockReturnValue({
      data: null,
      loading: false,
      error,
      refetch: mockRefetch,
    });

    render(<UserTypesList />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load user types')).toBeInTheDocument();
    });
  });

  it('opens menu when actions button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
      expect(screen.getByText('Delete')).toBeInTheDocument();
    });
  });

  it('navigates to view page when View is clicked', async () => {
    const user = userEvent.setup();
    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
    });

    const viewButton = screen.getByText('View');
    await user.click(viewButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/user-types/schema1');
    });
  });

  it('opens delete dialog when Delete is clicked', async () => {
    const user = userEvent.setup();
    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('Delete')).toBeInTheDocument();
    });

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Delete User Type')).toBeInTheDocument();
      expect(
        screen.getByText('Are you sure you want to delete this user type? This action cannot be undone.'),
      ).toBeInTheDocument();
    });
  });

  it('cancels delete when Cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Delete User Type')).toBeInTheDocument();
    });

    const cancelButton = screen.getByRole('button', {name: /cancel/i});
    await user.click(cancelButton);

    await waitFor(() => {
      expect(screen.queryByText('Delete User Type')).not.toBeInTheDocument();
    });
  });

  it('deletes user type when confirmed', async () => {
    const user = userEvent.setup();
    mockDeleteUserType.mockResolvedValue(undefined);

    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Delete User Type')).toBeInTheDocument();
    });

    const confirmButton = screen.getByRole('button', {name: /delete/i});
    await user.click(confirmButton);

    await waitFor(() => {
      expect(mockDeleteUserType).toHaveBeenCalledWith('schema1');
      expect(mockRefetch).toHaveBeenCalled();
    });
  });

  it('displays delete error in dialog', async () => {
    const user = userEvent.setup();
    const deleteError: ApiError = {
      code: 'DELETE_ERROR',
      message: 'Failed to delete',
      description: 'Cannot delete user type',
    };

    mockUseDeleteUserType.mockReturnValue({
      deleteUserType: mockDeleteUserType,
      loading: false,
      error: deleteError,
      reset: mockResetDeleteUserType,
    });

    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Failed to delete')).toBeInTheDocument();
      expect(screen.getByText('Cannot delete user type')).toBeInTheDocument();
    });
  });

  it('navigates when row is clicked', async () => {
    const user = userEvent.setup();
    render(<UserTypesList />);

    const row = screen.getByTestId('row-schema1');
    if (row) {
      await user.click(row);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/user-types/schema1');
      });
    }
  });

  it('closes snackbar when close button is clicked', async () => {
    const user = userEvent.setup();
    const error: ApiError = {
      code: 'ERROR_CODE',
      message: 'Failed to load user types',
      description: 'Error description',
    };

    mockUseGetUserTypes.mockReturnValue({
      data: null,
      loading: false,
      error,
      refetch: mockRefetch,
    });

    render(<UserTypesList />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load user types')).toBeInTheDocument();
    });

    const closeButton = screen.getByLabelText(/close/i);
    await user.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByText('Failed to load user types')).not.toBeInTheDocument();
    });
  });

  it('displays deleting state on confirm button', async () => {
    const user = userEvent.setup();
    mockUseDeleteUserType.mockReturnValue({
      deleteUserType: mockDeleteUserType,
      loading: true,
      error: null,
      reset: mockResetDeleteUserType,
    });

    render(<UserTypesList />);

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Deleting...')).toBeInTheDocument();
    });
  });

  it('renders empty grid when no data', () => {
    mockUseGetUserTypes.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        schemas: [],
      },
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UserTypesList />);

    const grid = screen.getByTestId('data-grid');
    expect(grid).toBeInTheDocument();
  });
});
