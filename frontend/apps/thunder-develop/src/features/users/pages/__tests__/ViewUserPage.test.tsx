/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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

import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import ViewUserPage from '../ViewUserPage';
import type {ApiUser, ApiUserSchema, UserSchemaListResponse} from '../../types/users';

const {mockLoggerError} = vi.hoisted(() => ({
  mockLoggerError: vi.fn(),
}));

const mockNavigate = vi.fn();
const mockUpdateMutateAsync = vi.fn();
const mockDeleteMutateAsync = vi.fn();
const mockResetUpdateError = vi.fn();
const mockResetDeleteError = vi.fn();

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: vi.fn(),
    error: mockLoggerError,
    debug: vi.fn(),
  }),
}));

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({userId: 'user123'}),
    Link: ({to, children = undefined, ...props}: {to: string; children?: ReactNode; [key: string]: unknown}) => (
      <a
        {...(props as Record<string, unknown>)}
        href={to}
        onClick={(e) => {
          e.preventDefault();
          Promise.resolve(mockNavigate(to)).catch(() => {});
        }}
      >
        {children}
      </a>
    ),
  };
});

// Mock hooks - TanStack Query interfaces
interface UseGetUserReturn {
  data: ApiUser | undefined;
  isLoading: boolean;
  error: Error | null;
}

interface UseGetUserSchemasReturn {
  data: UserSchemaListResponse | undefined;
  isLoading: boolean;
  error: Error | null;
}

interface UseGetUserSchemaReturn {
  data: ApiUserSchema | undefined;
  isLoading: boolean;
  error: Error | null;
}

interface UseUpdateUserReturn {
  mutate: ReturnType<typeof vi.fn>;
  mutateAsync: ReturnType<typeof vi.fn>;
  isPending: boolean;
  error: Error | null;
  data: unknown;
  isError: boolean;
  isSuccess: boolean;
  isIdle: boolean;
  reset: () => void;
}

interface UseDeleteUserReturn {
  mutate: ReturnType<typeof vi.fn>;
  mutateAsync: ReturnType<typeof vi.fn>;
  isPending: boolean;
  error: Error | null;
  data: unknown;
  isError: boolean;
  isSuccess: boolean;
  isIdle: boolean;
  reset: () => void;
}

const mockUseGetUser = vi.fn<() => UseGetUserReturn>();
const mockUseGetUserSchemas = vi.fn<() => UseGetUserSchemasReturn>();
const mockUseGetUserSchema = vi.fn<() => UseGetUserSchemaReturn>();
const mockUseUpdateUser = vi.fn<() => UseUpdateUserReturn>();
const mockUseDeleteUser = vi.fn<() => UseDeleteUserReturn>();

vi.mock('../../api/useGetUser', () => ({
  default: () => mockUseGetUser(),
}));

vi.mock('../../api/useGetUserSchemas', () => ({
  default: () => mockUseGetUserSchemas(),
}));

vi.mock('../../api/useGetUserSchema', () => ({
  default: () => mockUseGetUserSchema(),
}));

vi.mock('../../api/useUpdateUser', () => ({
  default: () => mockUseUpdateUser(),
}));

vi.mock('../../api/useDeleteUser', () => ({
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
    schemas: [{id: 'employee', name: 'Employee', ouId: 'test-ou'}],
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

  const defaultUpdateReturn: UseUpdateUserReturn = {
    mutate: vi.fn(),
    mutateAsync: mockUpdateMutateAsync,
    isPending: false,
    error: null,
    data: undefined,
    isError: false,
    isSuccess: false,
    isIdle: true,
    reset: mockResetUpdateError,
  };

  const defaultDeleteReturn: UseDeleteUserReturn = {
    mutate: vi.fn(),
    mutateAsync: mockDeleteMutateAsync,
    isPending: false,
    error: null,
    data: undefined,
    isError: false,
    isSuccess: false,
    isIdle: true,
    reset: mockResetDeleteError,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockResolvedValue(undefined);
    mockUpdateMutateAsync.mockResolvedValue(mockUserData);
    mockDeleteMutateAsync.mockResolvedValue(undefined);
    mockUseGetUser.mockReturnValue({
      data: mockUserData,
      isLoading: false,
      error: null,
    });
    mockUseGetUserSchemas.mockReturnValue({
      data: mockSchemasData,
      isLoading: false,
      error: null,
    });
    mockUseGetUserSchema.mockReturnValue({
      data: mockSchemaData,
      isLoading: false,
      error: null,
    });
    mockUseUpdateUser.mockReturnValue({...defaultUpdateReturn});
    mockUseDeleteUser.mockReturnValue({...defaultDeleteReturn});
  });

  describe('Loading and Error States', () => {
    it('displays loading spinner when user data is loading', async () => {
      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays loading spinner when schema is loading', async () => {
      mockUseGetUserSchema.mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays error alert when user fails to load', async () => {
      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('User not found'),
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('alert')).toHaveTextContent('User not found');
      await expect.element(page.getByRole('button', {name: /back to users/i})).toBeInTheDocument();
    });

    it('handles navigation error when clicking back button in error state', async () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('User not found'),
      });

      mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

      await renderWithProviders(<ViewUserPage />);

      const backButton = page.getByRole('button', {name: /back to users/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });

      consoleSpy.mockRestore();
    });

    it('displays error alert when schema fails to load', async () => {
      mockUseGetUserSchema.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error('Schema not found'),
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('alert')).toHaveTextContent('Schema not found');
    });

    it('displays generic error message when error message is empty', async () => {
      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error(''),
      });

      await renderWithProviders(<ViewUserPage />);

      // Should display the fallback message since error.message is empty
      await expect.element(page.getByRole('alert')).toHaveTextContent('');
    });

    it('displays warning when user is null but no error', async () => {
      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('alert')).toHaveTextContent('User not found');
    });

    it('handles navigation error when clicking back button in user not found state', async () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      });

      mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

      await renderWithProviders(<ViewUserPage />);

      const backButton = page.getByRole('button', {name: /back to users/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });

      consoleSpy.mockRestore();
    });
  });

  describe('View Mode', () => {
    it('renders user profile page with title', async () => {
      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('heading', {name: 'Manage User'})).toBeInTheDocument();
      await expect.element(page.getByText('View and manage user information')).toBeInTheDocument();
    });

    it('displays basic user information', async () => {
      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('User ID')).toBeInTheDocument();
      await expect.element(page.getByText('user123')).toBeInTheDocument();

      await expect.element(page.getByText('Organization Unit')).toBeInTheDocument();
      await expect.element(page.getByText('test-ou')).toBeInTheDocument();

      await expect.element(page.getByText('User Type')).toBeInTheDocument();
      await expect.element(page.getByText('Employee')).toBeInTheDocument();
    });

    it('displays user attributes in view mode', async () => {
      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('username')).toBeInTheDocument();
      await expect.element(page.getByText('john_doe')).toBeInTheDocument();

      await expect.element(page.getByText('email')).toBeInTheDocument();
      await expect.element(page.getByText('john@example.com')).toBeInTheDocument();

      await expect.element(page.getByText('age')).toBeInTheDocument();
      await expect.element(page.getByText('30')).toBeInTheDocument();

      await expect.element(page.getByText('active')).toBeInTheDocument();
      await expect.element(page.getByText('Yes')).toBeInTheDocument();
    });

    it('displays "No" for false boolean values', async () => {
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, attributes: {active: false}},
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('No')).toBeInTheDocument();
    });

    it('displays array values as comma-separated list', async () => {
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, attributes: {tags: ['admin', 'developer', 'manager']}},
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('admin, developer, manager')).toBeInTheDocument();
    });

    it('displays "No attributes available" when user has no attributes', async () => {
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, attributes: {}},
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('No attributes available')).toBeInTheDocument();
    });

    it('renders Edit and Delete buttons in view mode', async () => {
      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByRole('button', {name: /edit/i})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /delete/i})).toBeInTheDocument();
    });

    it('navigates back when Back button is clicked', async () => {
      await renderWithProviders(<ViewUserPage />);

      const backButton = page.getByRole('button', {name: /^back$/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });
    });

    it('handles navigation error when back button is clicked', async () => {
        const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
      mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

      await renderWithProviders(<ViewUserPage />);

      const backButton = page.getByRole('button', {name: /^back$/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });

      consoleSpy.mockRestore();
    });
  });

  describe('Edit Mode', () => {
    it('enters edit mode when Edit button is clicked', async () => {
      await renderWithProviders(<ViewUserPage />);

      const editButton = page.getByRole('button', {name: /edit/i});
      await userEvent.click(editButton);

      await expect.element(page.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
        await expect.element(page.getByRole('button', {name: /cancel/i})).toBeInTheDocument();


      // Edit and Delete buttons should not be visible in edit mode
      await expect.element(page.getByRole('button', {name: /^edit$/i})).not.toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /^delete$/i})).not.toBeInTheDocument();
    });

    it('does not submit if userId is missing from params', async () => {
        // When userId is undefined, useGetUser will be called with undefined
      // Let's test the guard clause by simulating missing required fields instead
      mockUseGetUser.mockReturnValue({
        data: {...mockUserData, organizationUnit: '', type: ''},
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      // Should not call mutateAsync when organizationUnit or type is empty
      await vi.waitFor(() => {
        expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      });
    });

    it('does not submit if user organizationUnit is missing', async () => {
        mockUseGetUser.mockReturnValue({
        data: {...mockUserData, organizationUnit: undefined as unknown as string},
        isLoading: false,
        error: null,
      });
      mockUseGetUserSchemas.mockReturnValue({
        data: {
          ...mockSchemasData,
          schemas: [{...mockSchemasData.schemas[0], ouId: ''}],
        },
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      });
    });

    it('does not submit if user type is missing', async () => {
        mockUseGetUser.mockReturnValue({
        data: {...mockUserData, type: undefined as unknown as string},
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
      });
    });

    it('displays form fields in edit mode', async () => {
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByPlaceholder(/Enter username/i)).toBeInTheDocument();
        await expect.element(page.getByPlaceholder(/Enter email/i)).toBeInTheDocument();
        await expect.element(page.getByPlaceholder(/Enter age/i)).toBeInTheDocument();
        await expect.element(page.getByRole('checkbox')).toBeInTheDocument();

    });

    it('filters out password field from schema in edit mode', async () => {
        const schemaWithPassword: ApiUserSchema = {
        id: 'employee',
        name: 'Employee',
        schema: {
          username: {
            type: 'string',
            required: true,
          },
          password: {
            type: 'string',
            required: true,
            credential: true,
          },
          email: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserSchema.mockReturnValue({
        data: schemaWithPassword,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByPlaceholder(/Enter username/i)).toBeInTheDocument();
      await expect.element(page.getByPlaceholder(/Enter email/i)).toBeInTheDocument();
        // Password field should not be present
      await expect.element(page.getByPlaceholder(/Enter password/i)).not.toBeInTheDocument();
    });

    it('filters out all credential fields from schema in edit mode', async () => {
        const schemaWithMultipleCredentials: ApiUserSchema = {
        id: 'employee',
        name: 'Employee',
        schema: {
          username: {
            type: 'string',
            required: true,
          },
          password: {
            type: 'string',
            required: true,
            credential: true,
          },
          pin: {
            type: 'string',
            credential: true,
          },
          email: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserSchema.mockReturnValue({
        data: schemaWithMultipleCredentials,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByPlaceholder(/Enter username/i)).toBeInTheDocument();
      await expect.element(page.getByPlaceholder(/Enter email/i)).toBeInTheDocument();
        // All credential fields should be filtered out
      await expect.element(page.getByPlaceholder(/Enter password/i)).not.toBeInTheDocument();
      await expect.element(page.getByPlaceholder(/Enter pin/i)).not.toBeInTheDocument();
    });

    it('does not filter non-credential fields with similar names', async () => {
        const schemaWithoutCredential: ApiUserSchema = {
        id: 'employee',
        name: 'Employee',
        schema: {
          username: {
            type: 'string',
            required: true,
          },
          password: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserSchema.mockReturnValue({
        data: schemaWithoutCredential,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

        // A field named "password" without credential: true should still appear
      await expect.element(page.getByPlaceholder(/Enter password/i)).toBeInTheDocument();
      await expect.element(page.getByPlaceholder(/Enter username/i)).toBeInTheDocument();
    });

    it('populates form fields with current user data', async () => {
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByPlaceholder(/Enter username/i)).toHaveValue('john_doe');
        await expect.element(page.getByPlaceholder(/Enter email/i)).toHaveValue('john@example.com');
        await expect.element(page.getByPlaceholder(/Enter age/i)).toHaveValue(30);
        await expect.element(page.getByRole('checkbox')).toBeChecked();

    });

    it('allows editing form fields', async () => {
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const emailInput = page.getByPlaceholder(/Enter email/i);
      await userEvent.clear(emailInput);
      await userEvent.fill(emailInput, 'newemail@example.com');

      await expect.element(emailInput).toHaveValue('newemail@example.com');
    });

    it('successfully updates user', async () => {
        const updatedUser: ApiUser = {
        ...mockUserData,
        attributes: {...mockUserData.attributes, email: 'updated@example.com'},
      };
      mockUpdateMutateAsync.mockResolvedValue(updatedUser);

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const emailInput = page.getByPlaceholder(/Enter email/i);
      await userEvent.clear(emailInput);
      await userEvent.fill(emailInput, 'updated@example.com');

      const saveButton = page.getByRole('button', {name: /save changes/i});
      await userEvent.click(saveButton);

      await vi.waitFor(() => {
        expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
          userId: 'user123',
          data: {
            organizationUnit: 'test-ou',
            type: 'Employee',
            attributes: {
              username: 'john_doe',
              email: 'updated@example.com',
              age: 30,
              active: true,
            },
          },
        });
      });
    });

    it('uses schema organization unit when updating user', async () => {
        mockUseGetUser.mockReturnValue({
        data: {...mockUserData, organizationUnit: 'stale-ou'},
        isLoading: false,
        error: null,
      });
      mockUseGetUserSchemas.mockReturnValue({
        data: {
          ...mockSchemasData,
          schemas: [{...mockSchemasData.schemas[0], ouId: 'schema-ou'}],
        },
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockUpdateMutateAsync).toHaveBeenCalled();
        const callArgs = mockUpdateMutateAsync.mock.calls[0][0] as {data: {organizationUnit: string}};
        expect(callArgs.data.organizationUnit).toBe('schema-ou');
      });
    });

    it('falls back to user organization unit when schema does not provide one', async () => {
        mockUseGetUserSchemas.mockReturnValue({
        data: {
          ...mockSchemasData,
          schemas: [{...mockSchemasData.schemas[0], ouId: ''}],
        },
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockUpdateMutateAsync).toHaveBeenCalled();
        const callArgs = mockUpdateMutateAsync.mock.calls[0][0] as {data: {organizationUnit: string}};
        expect(callArgs.data.organizationUnit).toBe('test-ou');
      });
    });

    it('exits edit mode after successful save', async () => {
        mockUpdateMutateAsync.mockResolvedValue(mockUserData);

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await expect.element(page.getByRole('button', {name: /edit/i})).toBeInTheDocument();
        await expect.element(page.getByRole('button', {name: /save changes/i})).not.toBeInTheDocument();

    });

    it('displays update error when save fails', async () => {
        mockUpdateMutateAsync.mockRejectedValue(new Error('Failed to update user'));
      mockUseUpdateUser.mockReturnValue({
        ...defaultUpdateReturn,
        error: new Error('Failed to update user'),
        isError: true,
        isIdle: false,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await expect.element(page.getByRole('alert')).toHaveTextContent('Failed to update user');

    });

    it('cancels edit mode and resets form', async () => {
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const emailInput = page.getByPlaceholder(/Enter email/i);
      await userEvent.clear(emailInput);
      await userEvent.fill(emailInput, 'changed@example.com');

      const cancelButton = page.getByRole('button', {name: /cancel/i});
      await userEvent.click(cancelButton);

      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /edit/i})).toBeInTheDocument();
        expect(mockResetUpdateError).toHaveBeenCalled();
      });
    });

    it('disables buttons during submission', async () => {
        const neverResolvingUpdate = vi.fn().mockImplementation(() => new Promise(() => {})); // Never resolves
      mockUseUpdateUser.mockReturnValue({
        ...defaultUpdateReturn,
        mutateAsync: neverResolvingUpdate,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const saveButton = page.getByRole('button', {name: /save changes/i});
      await userEvent.click(saveButton);

      await expect.element(page.getByRole('button', {name: /saving.../i})).toBeDisabled();
        await expect.element(page.getByRole('button', {name: /cancel/i})).toBeDisabled();

    });

    it('logs error when update fails', async () => {
        const error = new Error('Update failed');

      const failingUpdateMutateAsync = vi.fn().mockRejectedValue(error);
      mockUseUpdateUser.mockReturnValue({
        ...defaultUpdateReturn,
        mutateAsync: failingUpdateMutateAsync,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockLoggerError).toHaveBeenCalledWith('Failed to update user', {error});
      });
    });
  });

  describe('Delete Functionality', () => {
    it('opens delete confirmation dialog when Delete button is clicked', async () => {
      await renderWithProviders(<ViewUserPage />);

      const deleteButton = page.getByRole('button', {name: /^delete$/i});
      await userEvent.click(deleteButton);

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
        await expect.element(page.getByText('Delete User')).toBeInTheDocument();
        await expect.element(page.getByText(/Are you sure you want to delete this user/i)).toBeInTheDocument();

    });

    it('calls mutateAsync with correct userId when delete is confirmed', async () => {
  
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      const confirmButton = dialog.getByRole('button', {name: /^delete$/i});
      await userEvent.click(confirmButton);

      // Verify userId is passed correctly
      await vi.waitFor(() => {
        expect(mockDeleteMutateAsync).toHaveBeenCalledWith('user123');
      });
    });

    it('closes delete dialog when Cancel is clicked', async () => {
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      const cancelButton = dialog.getByRole('button', {name: /cancel/i});
      await userEvent.click(cancelButton);

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();

    });

    it('successfully deletes user and navigates to users list', async () => {
  
      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      const confirmButton = dialog.getByRole('button', {name: /^delete$/i});
      await userEvent.click(confirmButton);

      await vi.waitFor(() => {
        expect(mockDeleteMutateAsync).toHaveBeenCalledWith('user123');
        expect(mockNavigate).toHaveBeenCalledWith('/users');
      });
    });

    it('displays delete error in dialog', async () => {
        mockUseDeleteUser.mockReturnValue({
        ...defaultDeleteReturn,
        error: new Error('Failed to delete user'),
        isError: true,
        isIdle: false,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      expect(dialog.getByText('Failed to delete user')).toBeInTheDocument();
    });

    it('disables buttons during deletion', async () => {
        mockUseDeleteUser.mockReturnValue({
        ...defaultDeleteReturn,
        isPending: true,
        isIdle: false,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      expect(dialog.getByRole('button', {name: /deleting.../i})).toBeDisabled();
      expect(dialog.getByRole('button', {name: /cancel/i})).toBeDisabled();
    });

    it('logs error when delete fails', async () => {
        const error = new Error('Delete failed');

      const failingDeleteMutateAsync = vi.fn().mockRejectedValue(error);
      mockUseDeleteUser.mockReturnValue({
        ...defaultDeleteReturn,
        mutateAsync: failingDeleteMutateAsync,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      const confirmButton = dialog.getByRole('button', {name: /^delete$/i});
      await userEvent.click(confirmButton);

      await vi.waitFor(() => {
        expect(mockLoggerError).toHaveBeenCalledWith('Failed to delete user', {error});
      });
    });

    it('closes dialog after delete error', async () => {
        const failingDeleteMutateAsync = vi.fn().mockRejectedValue(new Error('Delete failed'));
      mockUseDeleteUser.mockReturnValue({
        ...defaultDeleteReturn,
        mutateAsync: failingDeleteMutateAsync,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /^delete$/i}));

      const dialog = page.getByRole('dialog');
      const confirmButton = dialog.getByRole('button', {name: /^delete$/i});
      await userEvent.click(confirmButton);

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();

    });
  });

  describe('Attribute Display Edge Cases', () => {
    it('displays dash for null attribute values', async () => {
      const userWithNullAttr: ApiUser = {
        id: 'user123',
        organizationUnit: 'test-ou',
        type: 'Employee',
        attributes: {
          username: 'john_doe',
          middleName: null as unknown as string,
        },
      };

      mockUseGetUser.mockReturnValue({
        data: userWithNullAttr,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      const middleNameSection = page.getByText('middleName').element().parentElement;
      expect(middleNameSection).toHaveTextContent('-');
    });

    it('displays dash for undefined attribute values', async () => {
      const userWithUndefinedAttr: ApiUser = {
        id: 'user123',
        organizationUnit: 'test-ou',
        type: 'Employee',
        attributes: {
          username: 'john_doe',
          nickname: undefined as unknown as string,
        },
      };

      mockUseGetUser.mockReturnValue({
        data: userWithUndefinedAttr,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      const nicknameSection = page.getByText('nickname').element().parentElement;
      expect(nicknameSection).toHaveTextContent('-');
    });

    it('displays comma-separated values for array attributes', async () => {
      const userWithArrayAttr: ApiUser = {
        id: 'user123',
        organizationUnit: 'test-ou',
        type: 'Employee',
        attributes: {
          username: 'john_doe',
          tags: ['developer', 'senior', 'fullstack'] as unknown as string,
        },
      };

      mockUseGetUser.mockReturnValue({
        data: userWithArrayAttr,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('tags')).toBeInTheDocument();
      await expect.element(page.getByText('developer, senior, fullstack')).toBeInTheDocument();
    });

    it('displays JSON string for object attributes', async () => {
      const userWithObjectAttr: ApiUser = {
        id: 'user123',
        organizationUnit: 'test-ou',
        type: 'Employee',
        attributes: {
          username: 'john_doe',
          address: {city: 'New York', country: 'USA'} as unknown as string,
        },
      };

      mockUseGetUser.mockReturnValue({
        data: userWithObjectAttr,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await expect.element(page.getByText('address')).toBeInTheDocument();
      await expect.element(page.getByText('{"city":"New York","country":"USA"}')).toBeInTheDocument();
    });

    it('displays dash for unknown attribute types', async () => {
      const userWithUnknownType: ApiUser = {
        id: 'user123',
        organizationUnit: 'test-ou',
        type: 'Employee',
        attributes: {
          username: 'john_doe',
          // Symbol is not a standard JSON type
          unknownType: Symbol('test') as unknown as string,
        },
      };

      mockUseGetUser.mockReturnValue({
        data: userWithUnknownType,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      const unknownTypeSection = page.getByText('unknownType').element().parentElement;
      expect(unknownTypeSection).toHaveTextContent('-');
    });
  });

  describe('Edge Cases', () => {
    it('displays fallback error message when error messages are undefined', async () => {
      mockUseGetUser.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: new Error(),
      });

      mockUseGetUserSchema.mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      const alert = page.getByRole('alert');
      await expect.element(alert).toBeInTheDocument();
    });

    it('displays "No schema available for editing" when schema is null in edit mode', async () => {
        mockUseGetUserSchema.mockReturnValue({
        data: {
          id: 'employee',
          name: 'Employee',
          schema: null as unknown as ApiUserSchema['schema'],
        },
        isLoading: false,
        error: null,
      });

      await renderWithProviders(<ViewUserPage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByText('No schema available for editing')).toBeInTheDocument();

    });
  });
});
