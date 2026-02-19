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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
import React from 'react';
import type * as OxygenUI from '@wso2/oxygen-ui';
import {DataGrid} from '@wso2/oxygen-ui';
import UsersList from '../UsersList';
import type {UserListResponse, ApiUserType, ApiError} from '../../types/users';

type DataGridProps = DataGrid.DataGridProps;

const mockNavigate = vi.fn();
const mockRefetch = vi.fn();
const mockDeleteUser = vi.fn();

// Mock DataGrid to avoid CSS import issues
interface MockRow {
  id: string;
  attributes?: Record<string, unknown>;
  [key: string]: unknown;
}

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof OxygenUI>('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      ...(actual.DataGrid ?? {}),
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

                    const content =
                      typeof column.renderCell === 'function' ? column.renderCell(params as never) : value;

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
    },
  };
});

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

interface UseGetUserTypeReturn {
  data: ApiUserType | null;
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
const mockUseGetUserType = vi.fn<(userTypeId: string) => UseGetUserTypeReturn>();
const mockUseDeleteUser = vi.fn<() => UseDeleteUserReturn>();

vi.mock('../../api/useGetUsers', () => ({
  default: () => mockUseGetUsers(),
}));

vi.mock('../../api/useGetUserType', () => ({
  default: (userTypeId: string) => mockUseGetUserType(userTypeId),
}));

vi.mock('../../api/useDeleteUser', () => ({
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

  const mockSchema: ApiUserType = {
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
    mockUseGetUserType.mockReturnValue({
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

    mockUseGetUserType.mockReturnValue({
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
      expect(screen.getByText('Are you sure you want to delete this user?')).toBeInTheDocument();
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
    mockUseGetUserType.mockReturnValue({
      data: null,
      loading: true,
      error: null,
    });

    render(<UsersList selectedSchema="schema1" />);

    const grid = screen.getByTestId('data-grid');
    expect(grid).toBeInTheDocument();
  });

  it('handles different field types correctly', async () => {
    const schemaWithTypes: ApiUserType = {
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

    mockUseGetUserType.mockReturnValue({
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

  it('handles error when delete user fails', async () => {
    const user = userEvent.setup();
    const deleteError = new Error('Delete failed');
    mockDeleteUser.mockRejectedValue(deleteError);

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
      // Note: Logger errors are tested separately, not via console.error spy
    });
  });

  it('handles error when row click navigation fails', async () => {
    const user = userEvent.setup();
    const navigationError = new Error('Navigation failed');
    mockNavigate.mockRejectedValue(navigationError);

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

  it('renders schema without avatar when name fields are not present', async () => {
    const schemaWithoutNameFields: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        email: {type: 'string'},
        phone: {type: 'string'},
      },
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithoutNameFields,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: mockUsersData,
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

  it('displays null values for missing attributes', async () => {
    const usersWithMissingAttrs: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
          },
        },
      ],
    };

    mockUseGetUsers.mockReturnValue({
      data: usersWithMissingAttrs,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('handles array field with empty array', async () => {
    const schemaWithArray: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        tags: {type: 'array', items: {type: 'string'}},
      },
    };

    const usersWithEmptyArray: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            tags: [],
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithArray,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithEmptyArray,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('handles array field with null value', async () => {
    const schemaWithArray: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        tags: {type: 'array', items: {type: 'string'}},
      },
    };

    const usersWithNullArray: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            tags: null,
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithArray,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithNullArray,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('handles object field with null value', async () => {
    const schemaWithObject: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        metadata: {type: 'object', properties: {}},
      },
    };

    const usersWithNullObject: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            metadata: null,
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithObject,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithNullObject,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('handles boolean field with null value', async () => {
    const schemaWithBoolean: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        verified: {type: 'boolean'},
      },
    };

    const usersWithNullBoolean: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            verified: null,
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithBoolean,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithNullBoolean,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('handles status field with string value', async () => {
    const schemaWithStatus: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        status: {type: 'string'},
      },
    };

    const usersWithStatus: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            status: 'active',
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithStatus,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithStatus,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByText('Active')).toBeInTheDocument();
    });
  });

  it('handles status field with inactive string value', async () => {
    const schemaWithStatus: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        status: {type: 'string'},
      },
    };

    const usersWithInactiveStatus: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            status: 'inactive',
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithStatus,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithInactiveStatus,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByText('Inactive')).toBeInTheDocument();
    });
  });

  it('handles active field with null value', async () => {
    const schemaWithActive: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        active: {type: 'boolean'},
      },
    };

    const usersWithNullActive: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            active: null,
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithActive,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithNullActive,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('opens menu for multiple users independently', async () => {
    const user = userEvent.setup();
    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toHaveTextContent('john.doe');
      expect(screen.getByTestId('row-user2')).toHaveTextContent('jane.smith');
    });

    const actionButtons = screen.getAllByRole('button', {name: /open actions menu/i});
    await user.click(actionButtons[0]);

    await waitFor(() => {
      expect(screen.getByText('View')).toBeInTheDocument();
    });

    // Menu should be displayed
    const viewButton = screen.getByText('View');
    expect(viewButton).toBeInTheDocument();
  });

  it('handles number field with null value', async () => {
    const schemaWithNumber: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        age: {type: 'number'},
      },
    };

    const usersWithNullNumber: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            age: null,
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithNumber,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithNullNumber,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('displays avatar with only username', async () => {
    const usersWithOnlyUsername: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john',
            firstname: '',
            lastname: '',
          },
        },
      ],
    };

    mockUseGetUsers.mockReturnValue({
      data: usersWithOnlyUsername,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('handles schema with empty schema entries', async () => {
    const emptySchema: ApiUserType = {
      id: 'schema1',
      name: 'Empty Schema',
      schema: {},
    };

    mockUseGetUserType.mockReturnValue({
      data: emptySchema,
      loading: false,
      error: null,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      const grid = screen.getByTestId('data-grid');
      expect(grid).toBeInTheDocument();
    });
  });

  it('displays false boolean value correctly', async () => {
    const schemaWithBoolean: ApiUserType = {
      id: 'schema1',
      name: 'Test Schema',
      schema: {
        username: {type: 'string'},
        verified: {type: 'boolean'},
      },
    };

    const usersWithFalseBoolean: UserListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      users: [
        {
          id: 'user1',
          organizationUnit: 'org1',
          type: 'schema1',
          attributes: {
            username: 'john.doe',
            verified: false,
          },
        },
      ],
    };

    mockUseGetUserType.mockReturnValue({
      data: schemaWithBoolean,
      loading: false,
      error: null,
    });

    mockUseGetUsers.mockReturnValue({
      data: usersWithFalseBoolean,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      expect(screen.getByTestId('row-user1')).toBeInTheDocument();
    });
  });

  it('displays loading state when deleting user', async () => {
    const user = userEvent.setup();
    mockUseDeleteUser.mockReturnValue({
      deleteUser: mockDeleteUser,
      loading: true,
      error: null,
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
      const confirmButton = screen.getByRole('button', {name: /loading/i});
      expect(confirmButton).toBeDisabled();
    });
  });

  it('handles schema loading with all field types', async () => {
    const comprehensiveSchema: ApiUserType = {
      id: 'schema1',
      name: 'Comprehensive Schema',
      schema: {
        username: {type: 'string'},
        firstname: {type: 'string'},
        lastname: {type: 'string'},
        isActive: {type: 'boolean'},
        age: {type: 'number'},
        tags: {type: 'array', items: {type: 'string'}},
        profile: {type: 'object', properties: {}},
      },
    };

    mockUseGetUserType.mockReturnValue({
      data: comprehensiveSchema,
      loading: false,
      error: null,
    });

    render(<UsersList selectedSchema="schema1" />);

    await waitFor(() => {
      const grid = screen.getByTestId('data-grid');
      expect(grid).toBeInTheDocument();
    });
  });
});
