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
import type {JSX} from 'react';
import type {InviteUserRenderProps} from '@asgardeo/react';
import UsersListPage from '../UsersListPage';
import type {UserSchemaListResponse} from '../../types/users';

const mockNavigate = vi.fn();

// Mock InviteUser component
const mockHandleInputChange = vi.fn();
const mockHandleInputBlur = vi.fn();
const mockHandleSubmit = vi.fn();
const mockCopyInviteLink = vi.fn();
const mockResetFlow = vi.fn();

const mockInviteUserRenderProps: InviteUserRenderProps = {
  values: {},
  fieldErrors: {},
  touched: {},
  error: null,
  isLoading: false,
  components: [],
  handleInputChange: mockHandleInputChange,
  handleInputBlur: mockHandleInputBlur,
  handleSubmit: mockHandleSubmit,
  isInviteGenerated: false,
  inviteLink: undefined,
  copyInviteLink: mockCopyInviteLink,
  inviteLinkCopied: false,
  resetFlow: mockResetFlow,
  isValid: false,
  meta: null,
};

vi.mock('@asgardeo/react', async () => {
  const actual = await vi.importActual<typeof import('@asgardeo/react')>('@asgardeo/react');
  return {
    ...actual,
    InviteUser: ({
      children,
    }: {
      children: (props: InviteUserRenderProps) => JSX.Element;
      onInviteLinkGenerated?: (link: string) => void;
      onError?: (error: Error) => void;
    }) => children(mockInviteUserRenderProps),
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

// Mock the UsersList component
vi.mock('../../components/UsersList', () => ({
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
vi.mock('../../api/useGetUserSchemas', () => ({
  default: () => mockUseGetUserSchemas(),
}));

// Mock useTemplateLiteralResolver
vi.mock('@thunder/shared-hooks', () => ({
  useTemplateLiteralResolver: () => ({
    resolve: (key: string) => key,
  }),
}));

// Store onSuccess callback for testing
let capturedOnSuccess: ((inviteLink: string) => void) | undefined;

// Mock InviteUserDialog to capture callbacks
vi.mock('../../components/InviteUserDialog', () => ({
  default: ({
    open,
    onClose,
    onSuccess,
  }: {
    open: boolean;
    onClose: () => void;
    onSuccess?: (inviteLink: string) => void;
  }) => {
    capturedOnSuccess = onSuccess;
    if (!open) return null;
    return (
      <div role="dialog" data-testid="invite-dialog">
        <button type="button" onClick={onClose} aria-label="close">
          Close
        </button>
        <button type="button" onClick={() => onSuccess?.('https://invite.link/123')} data-testid="trigger-success">
          Trigger Success
        </button>
      </div>
    );
  },
}));

describe('UsersListPage', () => {
  const mockSchemas: UserSchemaListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    schemas: [
      {id: 'schema1', name: 'Employee Schema', ouId: 'root-ou'},
      {id: 'schema2', name: 'Contractor Schema', ouId: 'child-ou'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemas,
    });
  });

  it('renders page title', async () => {
    await renderWithProviders(<UsersListPage />);

    await expect.element(page.getByText('User Management')).toBeInTheDocument();
  });

  it('renders page description', async () => {
    await renderWithProviders(<UsersListPage />);

    await expect.element(page.getByText('Manage users, roles, and permissions across your organization')).toBeInTheDocument();
  });

  it('renders create user button', async () => {
    await renderWithProviders(<UsersListPage />);

    const createButton = page.getByRole('button', {name: /add user/i});
    await expect.element(createButton).toBeInTheDocument();
  });

  it('renders search input', async () => {
    await renderWithProviders(<UsersListPage />);

    const searchInput = page.getByPlaceholder('Search users...');
    await expect.element(searchInput).toBeInTheDocument();
  });

  it('renders search icon', async () => {
    await renderWithProviders(<UsersListPage />);

    // Check for lucide-react Search icon
    const searchIcon = document.querySelector('svg');
    expect(searchIcon).toBeInTheDocument();
  });

  it('allows typing in search input', async () => {
    await renderWithProviders(<UsersListPage />);

    const searchInput = page.getByPlaceholder('Search users...');
    await userEvent.fill(searchInput, 'john doe');

    await expect.element(searchInput).toHaveValue('john doe');
  });

  it('navigates to create user page when create button is clicked', async () => {
    await renderWithProviders(<UsersListPage />);

    const createButton = page.getByRole('button', {name: /add user/i});
    await userEvent.click(createButton);

    expect(mockNavigate).toHaveBeenCalledWith('/users/create');
  });

  it('renders schema select dropdown', async () => {
    await renderWithProviders(<UsersListPage />);

    const select = page.getByRole('combobox');
    await expect.element(select).toBeInTheDocument();
  });

  it('displays schema options from API', async () => {
    await renderWithProviders(<UsersListPage />);

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    await vi.waitFor(() => {
      const employeeOptions = page.getByText('Employee Schema');
      const contractorOptions = page.getByText('Contractor Schema');
      expect(employeeOptions.length).toBeGreaterThan(0);
      expect(contractorOptions.length).toBeGreaterThan(0);
    });
  });

  it('selects first schema by default', async () => {
    await renderWithProviders(<UsersListPage />);

    const usersList = page.getByTestId('users-list');
    await expect.element(usersList).toHaveAttribute('data-schema', 'schema1');
  });

  it('changes selected schema when dropdown value changes', async () => {
    await renderWithProviders(<UsersListPage />);

    const select = page.getByRole('combobox');
    await userEvent.click(select);

    await expect.element(page.getByText('Contractor Schema')).toBeInTheDocument();


    await userEvent.click(page.getByText('Contractor Schema'));

    await vi.waitFor(async () => {
      const usersList = page.getByTestId('users-list');
      await expect.element(usersList).toHaveAttribute('data-schema', 'schema2');
    });
  });

  it('renders UsersList component', async () => {
    await renderWithProviders(<UsersListPage />);

    await expect.element(page.getByTestId('users-list')).toBeInTheDocument();
  });

  it('passes selected schema to UsersList', async () => {
    await renderWithProviders(<UsersListPage />);

    const usersList = page.getByTestId('users-list');
    await expect.element(usersList).toHaveAttribute('data-schema');
  });

  it('renders plus icon in create user button', async () => {
    await renderWithProviders(<UsersListPage />);

    const createButton = page.getByRole('button', {name: /add user/i});
    // Check that button has an icon by checking for svg within the button
    const icon = createButton.element().querySelector('svg');
    expect(icon).toBeInTheDocument();
  });

  it('handles empty schemas list', async () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: {totalResults: 0, startIndex: 1, count: 0, schemas: []},
    });

    await renderWithProviders(<UsersListPage />);

    const usersList = page.getByTestId('users-list');
    await expect.element(usersList).toHaveAttribute('data-schema', '');
  });

  it('handles undefined schemas data', async () => {
    mockUseGetUserSchemas.mockReturnValue({
      data: undefined,
    });

    await renderWithProviders(<UsersListPage />);

    await expect.element(page.getByText('User Management')).toBeInTheDocument();
    await expect.element(page.getByTestId('users-list')).toBeInTheDocument();
  });

  it('has correct heading level', async () => {
    await renderWithProviders(<UsersListPage />);

    const heading = page.getByRole('heading', {level: 1, name: /user management/i});
    await expect.element(heading).toBeInTheDocument();
  });

  it('create user button has contained variant', async () => {
    await renderWithProviders(<UsersListPage />);

    const createButton = page.getByRole('button', {name: /add user/i});
    await expect.element(createButton).toHaveClass('MuiButton-contained');
  });

  it('opens invite dialog when invite user button is clicked', async () => {
    await renderWithProviders(<UsersListPage />);

    const inviteButton = page.getByRole('button', {name: /invite user/i});
    await userEvent.click(inviteButton);

    await expect.element(page.getByRole('dialog')).toBeInTheDocument();

  });

  it('closes invite dialog when onClose is triggered', async () => {
    await renderWithProviders(<UsersListPage />);

    // Open dialog
    const inviteButton = page.getByRole('button', {name: /invite user/i});
    await userEvent.click(inviteButton);

    await expect.element(page.getByRole('dialog')).toBeInTheDocument();


    // Close dialog by clicking the close button
    const closeButton = page.getByRole('button', {name: /close/i});
    await userEvent.click(closeButton);

    await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();

  });

  it('calls onSuccess handler when invite is successful', async () => {
    await renderWithProviders(<UsersListPage />);

    // Open dialog
    const inviteButton = page.getByRole('button', {name: /invite user/i});
    await userEvent.click(inviteButton);

    await expect.element(page.getByRole('dialog')).toBeInTheDocument();


    // Trigger the success callback
    const triggerSuccessButton = page.getByTestId('trigger-success');
    await userEvent.click(triggerSuccessButton);

    // Verify the onSuccess callback was captured and can be called
    expect(capturedOnSuccess).toBeDefined();
  });

  it('handles navigation error gracefully', async () => {
    const navigationError = new Error('Navigation failed');
    mockNavigate.mockRejectedValueOnce(navigationError);

    await renderWithProviders(<UsersListPage />);

    const createButton = page.getByRole('button', {name: /add user/i});
    await userEvent.click(createButton);

    // Verify navigate was called even though it will fail
    expect(mockNavigate).toHaveBeenCalledWith('/users/create');

    // Wait a bit for the error handler to be called
    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });
  });
});
