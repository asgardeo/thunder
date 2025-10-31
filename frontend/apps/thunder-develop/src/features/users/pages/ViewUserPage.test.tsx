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
import {screen, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import render from '@/test/test-utils';
import ViewUserPage from './ViewUserPage';
import type {ApiError, ApiUser, ApiUserSchema, UserSchemaListResponse} from '../types/users';

const mockNavigate = vi.fn();
const mockUpdateUser = vi.fn();
const mockDeleteUser = vi.fn();
const mockResetUpdateError = vi.fn();
const mockResetDeleteError = vi.fn();
const mockRefetchUser = vi.fn();
const mockRefetchSchema = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({userId: 'user123'}),
  };
});

// Mock hooks
interface UseGetUserReturn {
  data: ApiUser | null;
  loading: boolean;
  error: ApiError | null;
  refetch: () => void;
}

interface UseGetUserSchemasReturn {
  data: UserSchemaListResponse | null;
  loading: boolean;
  error: ApiError | null;
  refetch: () => void;
}

interface UseGetUserSchemaReturn {
  data: ApiUserSchema | null;
  loading: boolean;
  error: ApiError | null;
  refetch: (id?: string) => void;
}

interface UseUpdateUserReturn {
  updateUser: (userId: string, data: {organizationUnit: string; type: string; attributes: Record<string, unknown>}) => Promise<ApiUser>;
  data: ApiUser | null;
  loading: boolean;
  error: ApiError | null;
  reset: () => void;
}

interface UseDeleteUserReturn {
  deleteUser: (userId: string) => Promise<boolean>;
  loading: boolean;
  error: ApiError | null;
  reset: () => void;
}

const mockUseGetUser = vi.fn<() => UseGetUserReturn>();
const mockUseGetUserSchemas = vi.fn<() => UseGetUserSchemasReturn>();
const mockUseGetUserSchema = vi.fn<() => UseGetUserSchemaReturn>();
const mockUseUpdateUser = vi.fn<() => UseUpdateUserReturn>();
const mockUseDeleteUser = vi.fn<() => UseDeleteUserReturn>();

vi.mock('../api/useGetUser', () => ({
  default: () => mockUseGetUser(),
}));

vi.mock('../api/useGetUserSchemas', () => ({
  default: () => mockUseGetUserSchemas(),
}));

vi.mock('../api/useGetUserSchema', () => ({
  default: () => mockUseGetUserSchema(),
}));

vi.mock('../api/useUpdateUser', () => ({
  default: () => mockUseUpdateUser(),
}));

vi.mock('../api/useDeleteUser', () => ({
  default: () => mockUseDeleteUser(),
}));

describe('ViewUserPage', () => {
  const mockUserData: ApiUser = {
    id: 'user123',
    organizationUnit: 'test-ou',
    type: 'Employee',
    attributes: {
      username: 'john_doe',
      email: 'john@example.com',
      age: 30,
      active: true,
    },
  };

  const mockSchemasData: UserSchemaListResponse = {
    totalResults: 1,
    startIndex: 1,
    count: 1,
    schemas: [
      {id: 'employee', name: 'Employee'},
    ],
  };

  const mockSchemaData: ApiUserSchema = {
    id: 'employee',
    name: 'Employee',
    schema: {
      username: {
        type: 'string',
        required: true,
      },
      email: {
        type: 'string',
        required: true,
      },
      age: {
        type: 'number',
        required: false,
      },
      active: {
        type: 'boolean',
        required: false,
      },
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUser.mockReturnValue({
      data: mockUserData,
      loading: false,
      error: null,
      refetch: mockRefetchUser,
    });
    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemasData,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });
    mockUseGetUserSchema.mockReturnValue({
      data: mockSchemaData,
      loading: false,
      error: null,
      refetch: mockRefetchSchema,
    });
    mockUseUpdateUser.mockReturnValue({
      updateUser: mockUpdateUser,
      data: null,
      loading: false,
      error: null,
      reset: mockResetUpdateError,
    });
    mockUseDeleteUser.mockReturnValue({
      deleteUser: mockDeleteUser,
      loading: false,
      error: null,
      reset: mockResetDeleteError,
    });
  });

  describe('Loading and Error States', () => {
    it('displays loading spinner when user data is loading', () => {
      mockUseGetUser.mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: mockRefetchUser,
      });

      render(<ViewUserPage />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays loading spinner when schema is loading', () => {
      mockUseGetUserSchema.mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: mockRefetchSchema,
      });

      render(<ViewUserPage />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays error alert when user fails to load', () => {
      const error: ApiError = {
        code: 'USER_NOT_FOUND',
        message: 'User not found',
        description: 'The requested user does not exist',
      };

      mockUseGetUser.mockReturnValue({
        data: null,
        loading: false,
        error,
        refetch: mockRefetchUser,
      });

      render(<ViewUserPage />);

      expect(screen.getByRole('alert')).toHaveTextContent('User not found');
      expect(screen.getByRole('button', {name: /back to users/i})).toBeInTheDocument();
    });

    it('displays error alert when schema fails to load', () => {
      const error: ApiError = {
        code: 'SCHEMA_NOT_FOUND',
        message: 'Schema not found',
        description: 'The requested schema does not exist',
      };

      mockUseGetUserSchema.mockReturnValue({
        data: null,
        loading: false,
        error,
        refetch: mockRefetchSchema,
      });

      render(<ViewUserPage />);

      expect(screen.getByRole('alert')).toHaveTextContent('Schema not found');
    });

    it('displays warning when user is null but no error', () => {
      mockUseGetUser.mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: mockRefetchUser,
      });

      render(<ViewUserPage />);

      expect(screen.getByRole('alert')).toHaveTextContent('User not found');
    });
  });

  describe('View Mode', () => {
    it('renders user profile page with title', () => {
      render(<ViewUserPage />);

      expect(screen.getByRole('heading', {name: 'User Profile'})).toBeInTheDocument();
      expect(screen.getByText('View and manage user information')).toBeInTheDocument();
    });

    it('displays basic user information', () => {
      render(<ViewUserPage />);

      expect(screen.getByText('User ID')).toBeInTheDocument();
      expect(screen.getByText('user123')).toBeInTheDocument();

      expect(screen.getByText('Organization Unit')).toBeInTheDocument();
      expect(screen.getByText('test-ou')).toBeInTheDocument();

      expect(screen.getByText('User Type')).toBeInTheDocument();
      expect(screen.getByText('Employee')).toBeInTheDocument();
    });

    it('displays user attributes in view mode', () => {
      render(<ViewUserPage />);

      expect(screen.getByText('username')).toBeInTheDocument();
      expect(screen.getByText('john_doe')).toBeInTheDocument();

      expect(screen.getByText('email')).toBeInTheDocument();
      expect(screen.getByText('john@example.com')).toBeInTheDocument();

      expect(screen.getByText('age')).toBeInTheDocument();
      expect(screen.getByText('30')).toBeInTheDocument();

      expect(screen.getByText('active')).toBeInTheDocument();
      expect(screen.getByText('Yes')).toBeInTheDocument();
    });

    it('displays "No" for false boolean values', () => {
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, attributes: {active: false}},
        loading: false,
        error: null,
        refetch: mockRefetchUser,
      });

      render(<ViewUserPage />);

      expect(screen.getByText('No')).toBeInTheDocument();
    });

    it('displays array values as comma-separated list', () => {
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, attributes: {tags: ['admin', 'developer', 'manager']}},
        loading: false,
        error: null,
        refetch: mockRefetchUser,
      });

      render(<ViewUserPage />);

      expect(screen.getByText('admin, developer, manager')).toBeInTheDocument();
    });

    it('displays "No attributes available" when user has no attributes', () => {
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, attributes: {}},
        loading: false,
        error: null,
        refetch: mockRefetchUser,
      });

      render(<ViewUserPage />);

      expect(screen.getByText('No attributes available')).toBeInTheDocument();
    });

    it('renders Edit and Delete buttons in view mode', () => {
      render(<ViewUserPage />);

      expect(screen.getByRole('button', {name: /edit/i})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /delete/i})).toBeInTheDocument();
    });

    it('navigates back when Back button is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      const backButton = screen.getByRole('button', {name: /go back/i});
      await user.click(backButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });
    });
  });

  describe('Edit Mode', () => {
    it('enters edit mode when Edit button is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      const editButton = screen.getByRole('button', {name: /edit/i});
      await user.click(editButton);

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
        expect(screen.getByRole('button', {name: /cancel/i})).toBeInTheDocument();
      });

      // Edit and Delete buttons should not be visible in edit mode
      expect(screen.queryByRole('button', {name: /^edit$/i})).not.toBeInTheDocument();
      expect(screen.queryByRole('button', {name: /^delete$/i})).not.toBeInTheDocument();
    });

    it('displays form fields in edit mode', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/Enter username/i)).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/Enter email/i)).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/Enter age/i)).toBeInTheDocument();
        expect(screen.getByRole('checkbox')).toBeInTheDocument();
      });
    });

    it('populates form fields with current user data', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/Enter username/i)).toHaveValue('john_doe');
        expect(screen.getByPlaceholderText(/Enter email/i)).toHaveValue('john@example.com');
        expect(screen.getByPlaceholderText(/Enter age/i)).toHaveValue(30);
        expect(screen.getByRole('checkbox')).toBeChecked();
      });
    });

    it('allows editing form fields', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const emailInput = await screen.findByPlaceholderText(/Enter email/i);
      await user.clear(emailInput);
      await user.type(emailInput, 'newemail@example.com');

      expect(emailInput).toHaveValue('newemail@example.com');
    });

    it('successfully updates user and refetches data', async () => {
      const user = userEvent.setup();
      const updatedUser: ApiUser = {
        ...mockUserData,
        attributes: {...mockUserData.attributes, email: 'updated@example.com'},
      };
      mockUpdateUser.mockResolvedValue(updatedUser);
      mockRefetchUser.mockResolvedValue(undefined);

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const emailInput = await screen.findByPlaceholderText(/Enter email/i);
      await user.clear(emailInput);
      await user.type(emailInput, 'updated@example.com');

      const saveButton = screen.getByRole('button', {name: /save changes/i});
      await user.click(saveButton);

      await waitFor(() => {
        expect(mockUpdateUser).toHaveBeenCalledWith('user123', {
          organizationUnit: 'test-ou',
          type: 'Employee',
          attributes: {
            username: 'john_doe',
            email: 'updated@example.com',
            age: 30,
            active: true,
          },
        });
        expect(mockRefetchUser).toHaveBeenCalled();
      });
    });

    it('exits edit mode after successful save', async () => {
      const user = userEvent.setup();
      mockUpdateUser.mockResolvedValue(mockUserData);
      mockRefetchUser.mockResolvedValue(undefined);

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /edit/i})).toBeInTheDocument();
        expect(screen.queryByRole('button', {name: /save changes/i})).not.toBeInTheDocument();
      });
    });

    it('displays update error when save fails', async () => {
      const user = userEvent.setup();
      const error: ApiError = {
        code: 'UPDATE_ERROR',
        message: 'Failed to update user',
        description: 'Validation failed',
      };
      mockUpdateUser.mockRejectedValue(new Error('Failed to update user'));
      mockUseUpdateUser.mockReturnValue({
        updateUser: mockUpdateUser,
        data: null,
        loading: false,
        error,
        reset: mockResetUpdateError,
      });

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(screen.getByRole('alert')).toHaveTextContent('Failed to update user');
        expect(screen.getByText('Validation failed')).toBeInTheDocument();
      });
    });

    it('displays special message for duplicate field error', async () => {
      const user = userEvent.setup();
      const error: ApiError = {
        code: 'USR-1014',
        message: 'Duplicate field value',
        description: 'Email already exists',
      };
      mockUpdateUser.mockRejectedValue(new Error('Duplicate field value'));
      mockUseUpdateUser.mockReturnValue({
        updateUser: mockUpdateUser,
        data: null,
        loading: false,
        error,
        reset: mockResetUpdateError,
      });

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(screen.getByText(/Please check the unique fields/i)).toBeInTheDocument();
      });
    });

    it('cancels edit mode and resets form', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const emailInput = await screen.findByPlaceholderText(/Enter email/i);
      await user.clear(emailInput);
      await user.type(emailInput, 'changed@example.com');

      const cancelButton = screen.getByRole('button', {name: /cancel/i});
      await user.click(cancelButton);

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /edit/i})).toBeInTheDocument();
        expect(mockResetUpdateError).toHaveBeenCalled();
      });
    });

    it('disables buttons during submission', async () => {
      const user = userEvent.setup();
      mockUpdateUser.mockImplementation(() => new Promise(() => {})); // Never resolves

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const saveButton = screen.getByRole('button', {name: /save changes/i});
      await user.click(saveButton);

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /saving.../i})).toBeDisabled();
        expect(screen.getByRole('button', {name: /cancel/i})).toBeDisabled();
      });
    });

    it('logs error when update fails', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
      const user = userEvent.setup();
      const error = new Error('Update failed');
      mockUpdateUser.mockRejectedValue(error);

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith('Failed to update user:', error);
      });

      consoleSpy.mockRestore();
    });
  });

  describe('Delete Functionality', () => {
    it('opens delete confirmation dialog when Delete button is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      const deleteButton = screen.getByRole('button', {name: /^delete$/i});
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByRole('dialog')).toBeInTheDocument();
        expect(screen.getByText('Delete User')).toBeInTheDocument();
        expect(screen.getByText(/Are you sure you want to delete this user/i)).toBeInTheDocument();
      });
    });

    it('closes delete dialog when Cancel is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /^delete$/i}));

      const dialog = screen.getByRole('dialog');
      const cancelButton = within(dialog).getByRole('button', {name: /cancel/i});
      await user.click(cancelButton);

      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
      });
    });

    it('successfully deletes user and navigates to users list', async () => {
      const user = userEvent.setup();
      mockDeleteUser.mockResolvedValue(true);

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /^delete$/i}));

      const dialog = screen.getByRole('dialog');
      const confirmButton = within(dialog).getByRole('button', {name: /^delete$/i});
      await user.click(confirmButton);

      await waitFor(() => {
        expect(mockDeleteUser).toHaveBeenCalledWith('user123');
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });
    });

    it('displays delete error in dialog', async () => {
      const user = userEvent.setup();
      const error: ApiError = {
        code: 'DELETE_ERROR',
        message: 'Failed to delete user',
        description: 'User has dependencies',
      };
      mockDeleteUser.mockRejectedValue(new Error('Failed to delete user'));
      mockUseDeleteUser.mockReturnValue({
        deleteUser: mockDeleteUser,
        loading: false,
        error,
        reset: mockResetDeleteError,
      });

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /^delete$/i}));

      const dialog = screen.getByRole('dialog');
      expect(within(dialog).getByText('Failed to delete user')).toBeInTheDocument();
      expect(within(dialog).getByText('User has dependencies')).toBeInTheDocument();
    });

    it('disables buttons during deletion', async () => {
      const user = userEvent.setup();
      mockDeleteUser.mockImplementation(() => new Promise(() => {})); // Never resolves
      mockUseDeleteUser.mockReturnValue({
        deleteUser: mockDeleteUser,
        loading: true,
        error: null,
        reset: mockResetDeleteError,
      });

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /^delete$/i}));

      const dialog = screen.getByRole('dialog');
      expect(within(dialog).getByRole('button', {name: /deleting.../i})).toBeDisabled();
      expect(within(dialog).getByRole('button', {name: /cancel/i})).toBeDisabled();
    });

    it('logs error when delete fails', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
      const user = userEvent.setup();
      const error = new Error('Delete failed');
      mockDeleteUser.mockRejectedValue(error);

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /^delete$/i}));

      const dialog = screen.getByRole('dialog');
      const confirmButton = within(dialog).getByRole('button', {name: /^delete$/i});
      await user.click(confirmButton);

      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith('Failed to delete user:', error);
      });

      consoleSpy.mockRestore();
    });

    it('closes dialog after delete error', async () => {
      const user = userEvent.setup();
      mockDeleteUser.mockRejectedValue(new Error('Delete failed'));

      render(<ViewUserPage />);

      await user.click(screen.getByRole('button', {name: /^delete$/i}));

      const dialog = screen.getByRole('dialog');
      const confirmButton = within(dialog).getByRole('button', {name: /^delete$/i});
      await user.click(confirmButton);

      await waitFor(() => {
        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
      });
    });
  });
});
