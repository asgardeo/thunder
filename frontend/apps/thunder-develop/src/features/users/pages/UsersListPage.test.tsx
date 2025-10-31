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
import render from '@/test/test-utils';
import UsersListPage from './UsersListPage';
import type {UserSchemaListResponse} from '../types/users';

const mockNavigate = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock the UsersList component
vi.mock('../components/UsersList', () => ({
  default: ({selectedSchema}: {selectedSchema: string}) => (
    <div data-testid="users-list" data-schema={selectedSchema}>
      Users List Component
    </div>
  ),
}));

// Define the return type for the hook
interface UseGetUserSchemasReturn {
  data: UserSchemaListResponse | undefined;
}

// Mock the useGetUserSchemas hook
const mockUseGetUserSchemas = vi.fn<() => UseGetUserSchemasReturn>();
vi.mock('../api/useGetUserSchemas', () => ({
  default: () => mockUseGetUserSchemas(),
}));

describe('UsersListPage', () => {
  const mockSchemas: UserSchemaListResponse = {
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
    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemas,
    });
  });

  it('renders page title', () => {
    render(<UsersListPage />);

    expect(screen.getByText('User Management')).toBeInTheDocument();
  });

  it('renders page description', () => {
    render(<UsersListPage />);

    expect(screen.getByText('Manage users, roles, and permissions across your organization')).toBeInTheDocument();
  });

  it('renders refresh button', () => {
    render(<UsersListPage />);

    const refreshButton = screen.getByRole('button', {name: /refresh/i});
    expect(refreshButton).toBeInTheDocument();
  });

  it('renders add user button', () => {
    render(<UsersListPage />);

    const addButton = screen.getByRole('button', {name: /add user/i});
    expect(addButton).toBeInTheDocument();
  });

  it('renders search input', () => {
    render(<UsersListPage />);

    const searchInput = screen.getByPlaceholderText('Search users...');
    expect(searchInput).toBeInTheDocument();
  });

  it('renders search icon', () => {
    const {container} = render(<UsersListPage />);

    // Check for lucide-react Search icon
    const searchIcon = container.querySelector('svg');
    expect(searchIcon).toBeInTheDocument();
  });

  it('allows typing in search input', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    const searchInput = screen.getByPlaceholderText('Search users...');
    await user.type(searchInput, 'john doe');

    expect(searchInput).toHaveValue('john doe');
  });

  it('navigates to add user page when add button is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    const addButton = screen.getByRole('button', {name: /add user/i});
    await user.click(addButton);

    expect(mockNavigate).toHaveBeenCalledWith('/users/create');
  });

  it('renders schema select dropdown', () => {
    render(<UsersListPage />);

    const select = screen.getByRole('combobox');
    expect(select).toBeInTheDocument();
  });

  it('displays schema options from API', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    const select = screen.getByRole('combobox');
    await user.click(select);

    await waitFor(() => {
      const employeeOptions = screen.getAllByText('Employee Schema');
      const contractorOptions = screen.getAllByText('Contractor Schema');
      expect(employeeOptions.length).toBeGreaterThan(0);
      expect(contractorOptions.length).toBeGreaterThan(0);
    });
  });

  it('selects first schema by default', () => {
    render(<UsersListPage />);

    const usersList = screen.getByTestId('users-list');
    expect(usersList).toHaveAttribute('data-schema', 'schema1');
  });

  it('changes selected schema when dropdown value changes', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    const select = screen.getByRole('combobox');
    await user.click(select);

    await waitFor(() => {
      expect(screen.getByText('Contractor Schema')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Contractor Schema'));

    await waitFor(() => {
      const usersList = screen.getByTestId('users-list');
      expect(usersList).toHaveAttribute('data-schema', 'schema2');
    });
  });

  it('renders UsersList component', () => {
    render(<UsersListPage />);

    expect(screen.getByTestId('users-list')).toBeInTheDocument();
  });

  it('passes selected schema to UsersList', () => {
    render(<UsersListPage />);

    const usersList = screen.getByTestId('users-list');
    expect(usersList).toHaveAttribute('data-schema');
  });

  it('renders refresh icon in refresh button', () => {
    render(<UsersListPage />);

    const refreshButton = screen.getByRole('button', {name: /refresh/i});
    // Check that button has an icon by checking for svg within the button
    const icon = refreshButton.querySelector('svg');
    expect(icon).toBeInTheDocument();
  });

  it('renders add icon in add user button', () => {
    render(<UsersListPage />);

    const addButton = screen.getByRole('button', {name: /add user/i});
    // Check that button has an icon by checking for svg within the button
    const icon = addButton.querySelector('svg');
    expect(icon).toBeInTheDocument();
  });

  it('handles empty schemas list', () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: {totalResults: 0, startIndex: 1, count: 0, schemas: []},
    });

    render(<UsersListPage />);

    const usersList = screen.getByTestId('users-list');
    expect(usersList).toHaveAttribute('data-schema', '');
  });

  it('handles undefined schemas data', () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: undefined,
    });

    render(<UsersListPage />);

    expect(screen.getByText('User Management')).toBeInTheDocument();
    expect(screen.getByTestId('users-list')).toBeInTheDocument();
  });

  it('has correct heading level', () => {
    render(<UsersListPage />);

    const heading = screen.getByRole('heading', {level: 1, name: /user management/i});
    expect(heading).toBeInTheDocument();
  });

  it('refresh button has outlined variant', () => {
    render(<UsersListPage />);

    const refreshButton = screen.getByRole('button', {name: /refresh/i});
    expect(refreshButton).toHaveClass('MuiButton-outlined');
  });

  it('add user button has contained variant', () => {
    render(<UsersListPage />);

    const addButton = screen.getByRole('button', {name: /add user/i});
    expect(addButton).toHaveClass('MuiButton-contained');
  });
});
