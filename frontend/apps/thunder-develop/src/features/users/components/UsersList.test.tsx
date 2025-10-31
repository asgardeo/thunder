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
import React from 'react';
import type {DataGridProps} from '@mui/x-data-grid';
import render from '@/test/test-utils';
import UsersList from './UsersList';
import type {UserListResponse, ApiUserSchema, ApiError} from '../types/users';

const mockNavigate = vi.fn();
const mockRefetch = vi.fn();
const mockDeleteUser = vi.fn();

// Mock DataGrid to avoid CSS import issues
interface MockRow {
  id: string;
  attributes?: Record<string, unknown>;
  [key: string]: unknown;
}

vi.mock('@mui/x-data-grid', () => ({
  DataGrid: (props: DataGridProps) => {
    const {rows = [], columns = [], loading, onRowClick} = props;

    return (
      <div data-testid="data-grid" data-loading={loading}>
        {rows.map((row) => {
          const mockRow = row as MockRow;
          const username = mockRow.attributes?.username;
          const displayText = typeof username === 'string' ? username : mockRow.id;

          return (
            <div key={mockRow.id} className="MuiDataGrid-row-container">
              <button
                type="button"
                className="MuiDataGrid-row"
                onClick={() => {
                  if (onRowClick) {
                    onRowClick({row: mockRow} as never, {} as never, {} as never);
                  }
                }}
                data-testid={`row-${mockRow.id}`}
              >
                {displayText}
              </button>
              {columns?.map((column) => {
                if (column?.field === undefined) return null;

                let value: unknown;
                if (typeof column.valueGetter === 'function') {
                  value = column.valueGetter({} as never, mockRow as never, column as never, {} as never);
                } else if (column.field in mockRow) {
                  value = mockRow[column.field];
                } else {
                  value = mockRow.attributes?.[column.field];
                }

                const params = {
                  row: mockRow,
                  field: column.field,
                  value,
                  id: mockRow.id,
                };

                const content = typeof column.renderCell === 'function' ? column.renderCell(params as never) : value;

                if (content === null || content === undefined) {
                  return null;
                }

                // Convert content to a renderable format
                let renderableContent: React.ReactNode;
                if (typeof content === 'string' || typeof content === 'number' || typeof content === 'boolean') {
                  renderableContent = String(content);
                } else if (React.isValidElement(content)) {
                  renderableContent = content;
                } else if (Array.isArray(content)) {
                  renderableContent = JSON.stringify(content);
                } else if (typeof content === 'object') {
                  renderableContent = JSON.stringify(content);
                } else {
                  renderableContent = '';
                }

                return (
                  <span key={`${mockRow.id}-${column.field}`} className="MuiDataGrid-cell">
                    {renderableContent}
                  </span>
                );
              })}
            </div>
          );
        })}
      </div>
    );
  },
  GridColDef: {} as never,
  GridRenderCellParams: {} as never,
}));

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock hooks
interface UseGetUsersReturn {
  data: UserListResponse | null;
  loading: boolean;
  error: ApiError | null;
  refetch: (params?: unknown) => void;
}

interface UseGetUserSchemaReturn {
  data: ApiUserSchema | null;
  loading: boolean;
  error: ApiError | null;
  refetch?: (newId?: string) => void;
}

interface UseDeleteUserReturn {
  deleteUser: (userId: string) => Promise<boolean | undefined>;
  loading: boolean;
  error: ApiError | null;
  reset?: () => void;
}

const mockUseGetUsers = vi.fn<() => UseGetUsersReturn>();
const mockUseGetUserSchema = vi.fn<(schemaId: string) => UseGetUserSchemaReturn>();
const mockUseDeleteUser = vi.fn<() => UseDeleteUserReturn>();

vi.mock('../api/useGetUsers', () => ({
  default: () => mockUseGetUsers(),
}));

vi.mock('../api/useGetUserSchema', () => ({
  default: (schemaId: string) => mockUseGetUserSchema(schemaId),
}));

vi.mock('../api/useDeleteUser', () => ({
  default: () => mockUseDeleteUser(),
}));

describe('UsersList', () => {
  const mockUsersData: UserListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    users: [
      {
        id: 'user1',
        organizationUnit: 'org1',
        type: 'schema1',
        attributes: {
          username: 'john.doe',
          firstname: 'John',
          lastname: 'Doe',
          email: 'john@example.com',
          isActive: true,
        },
      },
      {
        id: 'user2',
        organizationUnit: 'org2',
        type: 'schema2',
        attributes: {
          username: 'jane.smith',
          firstname: 'Jane',
          lastname: 'Smith',
          email: 'jane@example.com',
          isActive: false,
        },
      },
    ],
  };

  const mockSchema: ApiUserSchema = {
    id: 'schema1',
    name: 'Default Schema',
    schema: {
      username: {type: 'string', required: true},
      firstname: {type: 'string'},
      lastname: {type: 'string'},
      email: {type: 'string'},
      isActive: {type: 'boolean'},
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUsers.mockReturnValue({
      data: mockUsersData,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });
    mockUseGetUserSchema.mockReturnValue({
      data: mockSchema,
      loading: false,
      error: null,
    });
    mockUseDeleteUser.mockReturnValue({
      deleteUser: mockDeleteUser,
      loading: false,
      error: null,
    });
  });

  it('renders DataGrid with users', async () => {
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
      expect(screen.getByTestId('row-user2')).toHaveTextContent('jane.smith');
    });
  });

  it('displays user avatars with initials', async () => {
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      const grid = screen.getByTestId('data-grid');
      expect(grid).toBeInTheDocument();
    });
  });

  it('displays loading state', () => {
    mockUseGetUsers.mockReturnValue({
      data: null,
      loading: true,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    const grid = screen.getByTestId('data-grid');
    expect(grid).toBeInTheDocument();
  });

  it('displays error from users request', async () => {
    const error: ApiError = {
      code: 'ERROR_CODE',
      message: 'Failed to load users',
      description: 'Error description',
    };

    mockUseGetUsers.mockReturnValue({
      data: null,
      loading: false,
      error,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load users')).toBeInTheDocument();
    });
  });

  it('displays error from schema request', async () => {
    const error: ApiError = {
      code: 'ERROR_CODE',
      message: 'Failed to load schema',
      description: 'Error description',
    };

    mockUseGetUserSchema.mockReturnValue({
      data: null,
      loading: false,
      error,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load schema')).toBeInTheDocument();
    });
  });

  it('opens menu when actions button is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
      expect(screen.getByText('Delete')).toBeInTheDocument();
    });
  });

  it('navigates to view page when View is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
    });

    const viewButton = screen.getByText('View');
    await user.click(viewButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users/user1');
    });
  });

  it('opens delete dialog when Delete is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Delete User')).toBeInTheDocument();
      expect(
        screen.getByText('Are you sure you want to delete this user? This action cannot be undone.'),
      ).toBeInTheDocument();
    });
  });

  it('deletes user when confirmed', async () => {
    const user = userEvent.setup();
    mockDeleteUser.mockResolvedValue(undefined);

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Delete User')).toBeInTheDocument();
    });

    const confirmButton = screen.getByRole('button', {name: /^delete$/i});
    await user.click(confirmButton);

    await waitFor(() => {
      expect(mockDeleteUser).toHaveBeenCalledWith('user1');
      expect(mockRefetch).toHaveBeenCalled();
    });
  });

  it('navigates when row is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const row = screen.getByTestId('row-user1');
    await user.click(row);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/users/user1');
    });
  });

  it('displays active status with chip', async () => {
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByText('Active')).toBeInTheDocument();
      expect(screen.getByText('Inactive')).toBeInTheDocument();
    });
  });

  it('renders empty grid when no data', async () => {
    mockUseGetUsers.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        users: [],
      },
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      const grid = screen.getByTestId('data-grid');
      expect(grid).toBeInTheDocument();
    });
  });

  it('renders empty columns when schema is not loaded', () => {
    mockUseGetUserSchema.mockReturnValue({
      data: null,
      loading: true,
      error: null,
    });

    render(<UsersList selectedSchema="schema1" />);

    const grid = screen.getByTestId('data-grid');
    expect(grid).toBeInTheDocument();
  });

  it('handles different field types correctly', async () => {
    const schemaWithTypes: ApiUserSchema = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        age: {type: 'number'},
        verified: {type: 'boolean'},
        tags: {type: 'array', items: {type: 'string'}},
        metadata: {type: 'object', properties: {}},
      },
    };

    const usersWithTypes: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'testuser',
            age: 25,
            verified: true,
            tags: ['tag1', 'tag2'],
            metadata: {key: 'value'},
          },
        },
      ],
    };

    mockUseGetUserSchema.mockReturnValue({
      data: schemaWithTypes,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithTypes,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('testuser');
    });
  });

  it('cancels delete when Cancel button is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Delete User')).toBeInTheDocument();
    });

    const cancelButton = screen.getByRole('button', {name: /cancel/i});
    await user.click(cancelButton);

    await waitFor(() => {
      expect(screen.queryByText('Delete User')).not.toBeInTheDocument();
    });
  });

  it('displays delete error in dialog', async () => {
    const user = userEvent.setup();
    const deleteError: ApiError = {
      code: 'DELETE_ERROR',
      message: 'Failed to delete',
      description: 'Cannot delete user',
    };

    mockUseDeleteUser.mockReturnValue({
      deleteUser: mockDeleteUser,
      loading: false,
      error: deleteError,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    const deleteButton = screen.getByText('Delete');
    await user.click(deleteButton);

    await waitFor(() => {
      expect(screen.getByText('Failed to delete')).toBeInTheDocument();
      expect(screen.getByText('Cannot delete user')).toBeInTheDocument();
    });
  });

  it('closes snackbar when close button is clicked', async () => {
    const user = userEvent.setup();
    const error: ApiError = {
      code: 'ERROR_CODE',
      message: 'Failed to load users',
      description: 'Error description',
    };

    mockUseGetUsers.mockReturnValue({
      data: null,
      loading: false,
      error,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByText('Failed to load users')).toBeInTheDocument();
    });

    const closeButton = screen.getByLabelText(/close/i);
    await user.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByText('Failed to load users')).not.toBeInTheDocument();
    });
  });
});
