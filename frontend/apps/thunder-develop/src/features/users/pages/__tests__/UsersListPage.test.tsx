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

  it('renders page title', () => {
    render(<UsersListPage />);

    expect(screen.getByText('User Management')).toBeInTheDocument();
  });

  it('renders page description', () => {
    render(<UsersListPage />);

    expect(screen.getByText('Manage users, roles, and permissions across your organization')).toBeInTheDocument();
  });

  it('renders create user button', () => {
    render(<UsersListPage />);

    const createButton = screen.getByRole('button', {name: /add user/i});
    expect(createButton).toBeInTheDocument();
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

  it('navigates to create user page when create button is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    const createButton = screen.getByRole('button', {name: /add user/i});
    await user.click(createButton);

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

  it('renders plus icon in create user button', () => {
    render(<UsersListPage />);

    const createButton = screen.getByRole('button', {name: /add user/i});
    // Check that button has an icon by checking for svg within the button
    const icon = createButton.querySelector('svg');
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

  it('create user button has contained variant', () => {
    render(<UsersListPage />);

    const createButton = screen.getByRole('button', {name: /add user/i});
    expect(createButton).toHaveClass('MuiButton-contained');
  });

  it('opens invite dialog when invite user button is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    const inviteButton = screen.getByRole('button', {name: /invite user/i});
    await user.click(inviteButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
  });

  it('closes invite dialog when onClose is triggered', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    // Open dialog
    const inviteButton = screen.getByRole('button', {name: /invite user/i});
    await user.click(inviteButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    // Close dialog by clicking the close button
    const closeButton = screen.getByRole('button', {name: /close/i});
    await user.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });

  it('calls onSuccess handler when invite is successful', async () => {
    const user = userEvent.setup();
    render(<UsersListPage />);

    // Open dialog
    const inviteButton = screen.getByRole('button', {name: /invite user/i});
    await user.click(inviteButton);

    await waitFor(() => {
      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    // Trigger the success callback
    const triggerSuccessButton = screen.getByTestId('trigger-success');
    await user.click(triggerSuccessButton);

    // Verify the onSuccess callback was captured and can be called
    expect(capturedOnSuccess).toBeDefined();
  });

  it('handles navigation error gracefully', async () => {
    const navigationError = new Error('Navigation failed');
    mockNavigate.mockRejectedValueOnce(navigationError);

    const user = userEvent.setup();
    render(<UsersListPage />);

    const createButton = screen.getByRole('button', {name: /add user/i});
    await user.click(createButton);

    // Verify navigate was called even though it will fail
    expect(mockNavigate).toHaveBeenCalledWith('/users/create');

    // Wait a bit for the error handler to be called
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalled();
    });
  });
});
