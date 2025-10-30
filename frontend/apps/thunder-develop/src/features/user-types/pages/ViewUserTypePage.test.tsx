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
import ViewUserTypePage from './ViewUserTypePage';
import type {ApiUserSchema, ApiError} from '../types/user-types';

const mockNavigate = vi.fn();
const mockRefetch = vi.fn();
const mockUpdateUserType = vi.fn();
const mockResetUpdateError = vi.fn();
const mockDeleteUserType = vi.fn();

// Mock react-router
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({id: 'schema-123'}),
  };
});

// Mock hooks
interface UseGetUserTypeReturn {
  data: ApiUserSchema | null;
  loading: boolean;
  error: ApiError | null;
  refetch: (id: string) => void;
}

interface UseUpdateUserTypeReturn {
  updateUserType: (id: string, data: {name: string; schema: Record<string, unknown>}) => Promise<void>;
  error: ApiError | null;
  reset: () => void;
}

interface UseDeleteUserTypeReturn {
  deleteUserType: (id: string) => Promise<void>;
  loading: boolean;
  error: ApiError | null;
}

const mockUseGetUserType = vi.fn<(id?: string) => UseGetUserTypeReturn>();
const mockUseUpdateUserType = vi.fn<() => UseUpdateUserTypeReturn>();
const mockUseDeleteUserType = vi.fn<() => UseDeleteUserTypeReturn>();

vi.mock('../api/useGetUserType', () => ({
  default: (id?: string) => mockUseGetUserType(id),
}));

vi.mock('../api/useUpdateUserType', () => ({
  default: () => mockUseUpdateUserType(),
}));

vi.mock('../api/useDeleteUserType', () => ({
  default: () => mockUseDeleteUserType(),
}));

describe('ViewUserTypePage', () => {
  const mockUserType: ApiUserSchema = {
    id: 'schema-123',
    name: 'Employee Schema',
    schema: {
      email: {
        type: 'string',
        required: true,
        unique: true,
      },
      age: {
        type: 'number',
        required: false,
      },
      isActive: {
        type: 'boolean',
        required: true,
      },
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetUserType.mockReturnValue({
      data: mockUserType,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });
    mockUseUpdateUserType.mockReturnValue({
      updateUserType: mockUpdateUserType,
      error: null,
      reset: mockResetUpdateError,
    });
    mockUseDeleteUserType.mockReturnValue({
      deleteUserType: mockDeleteUserType,
      loading: false,
      error: null,
    });
  });

  describe('Loading and Error States', () => {
    it('displays loading state', () => {
      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays error state with error message', () => {
      const error: ApiError = {
        code: 'LOAD_ERROR',
        message: 'Failed to load user type',
        description: 'Network error',
      };

      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      expect(screen.getByText('Failed to load user type')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /back to user types/i})).toBeInTheDocument();
    });

    it('displays warning when user type not found', () => {
      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      expect(screen.getByText('User type not found')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /back to user types/i})).toBeInTheDocument();
    });

    it('navigates back from error state', async () => {
      const user = userEvent.setup();
      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error: {code: 'ERROR', message: 'Error', description: ''},
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      const backButton = screen.getByRole('button', {name: /back to user types/i});
      await user.click(backButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });
  });

  describe('View Mode', () => {
    it('renders user type details in view mode', () => {
      render(<ViewUserTypePage />);

      expect(screen.getByText('User Type Details')).toBeInTheDocument();
      expect(screen.getByText('View and manage user type schema')).toBeInTheDocument();
      expect(screen.getByText('schema-123')).toBeInTheDocument();
      expect(screen.getByText('Employee Schema')).toBeInTheDocument();
    });

    it('displays schema properties in table', () => {
      render(<ViewUserTypePage />);

      expect(screen.getByText('Property Name')).toBeInTheDocument();
      expect(screen.getByText('Type')).toBeInTheDocument();
      expect(screen.getByText('Required')).toBeInTheDocument();
      expect(screen.getByText('Unique')).toBeInTheDocument();

      expect(screen.getByText('email')).toBeInTheDocument();
      expect(screen.getByText('age')).toBeInTheDocument();
      expect(screen.getByText('isActive')).toBeInTheDocument();
    });

    it('displays edit and delete buttons in view mode', () => {
      render(<ViewUserTypePage />);

      expect(screen.getByRole('button', {name: /edit/i})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /delete/i})).toBeInTheDocument();
    });

    it('navigates back when back button is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      const backButton = screen.getByRole('button', {name: /go back/i});
      await user.click(backButton);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });

    it('displays enum values in view mode', () => {
      const userTypeWithEnum: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        schema: {
          status: {
            type: 'string',
            required: true,
            enum: ['ACTIVE', 'INACTIVE', 'PENDING'],
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithEnum,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      expect(screen.getByText(/ACTIVE, INACTIVE, PENDING/i)).toBeInTheDocument();
    });

    it('displays regex pattern in view mode', () => {
      const userTypeWithRegex: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        schema: {
          username: {
            type: 'string',
            required: true,
            regex: '^[a-zA-Z0-9]+$',
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithRegex,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      expect(screen.getByText('Pattern:')).toBeInTheDocument();
      expect(screen.getByText('^[a-zA-Z0-9]+$')).toBeInTheDocument();
    });
  });

  describe('Edit Mode', () => {
    it('enters edit mode when edit button is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      const editButton = screen.getByRole('button', {name: /edit/i});
      await user.click(editButton);

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
        expect(screen.getByRole('button', {name: /cancel/i})).toBeInTheDocument();
        expect(screen.queryByRole('button', {name: /edit/i})).not.toBeInTheDocument();
      });
    });

    it('displays editable form fields in edit mode', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      await waitFor(() => {
        expect(screen.getByPlaceholderText(/user type name/i)).toBeInTheDocument();
      });
    });

    it('allows editing user type name', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const nameInput = screen.getByPlaceholderText(/user type name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Schema Name');

      expect(nameInput).toHaveValue('Updated Schema Name');
    });

    it('allows toggling required checkbox', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const requiredCheckboxes = screen.getAllByRole('checkbox', {name: /required/i});
      const firstCheckbox = requiredCheckboxes[0];

      const isInitiallyChecked = firstCheckbox.getAttribute('checked') !== null;
      await user.click(firstCheckbox);

      await waitFor(() => {
        if (isInitiallyChecked) {
          expect(firstCheckbox).not.toBeChecked();
        } else {
          expect(firstCheckbox).toBeChecked();
        }
      });
    });

    it('allows changing property type', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const typeSelects = screen.getAllByRole('combobox');
      await user.click(typeSelects[0]);

      // Click on Number option instead to avoid duplicate "String" text
      const numberOption = await screen.findByRole('option', {name: 'Number'});
      await user.click(numberOption);

      await waitFor(() => {
        expect(typeSelects[0]).toHaveTextContent('Number');
      });
    });

    it('cancels edit mode and reverts changes', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const nameInput = screen.getByPlaceholderText(/user type name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'Changed Name');

      await user.click(screen.getByRole('button', {name: /cancel/i}));

      await waitFor(() => {
        expect(screen.getByText('Employee Schema')).toBeInTheDocument();
        expect(screen.queryByPlaceholderText(/user type name/i)).not.toBeInTheDocument();
        expect(mockResetUpdateError).toHaveBeenCalled();
      });
    });

    it('saves changes successfully', async () => {
      const user = userEvent.setup();
      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const nameInput = screen.getByPlaceholderText(/user type name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Schema');

      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith('schema-123', {
          name: 'Updated Schema',
          schema: expect.any(Object) as Record<string, unknown>,
        });
        expect(mockRefetch).toHaveBeenCalledWith('schema-123');
      });
    });

    it('displays saving state', async () => {
      const user = userEvent.setup();
      mockUpdateUserType.mockImplementation(
        () =>
          new Promise((resolve) => {
            setTimeout(resolve, 100);
          }),
      );

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      expect(screen.getByText('Saving...')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /saving.../i})).toBeDisabled();
    });

    it('displays update error', async () => {
      const user = userEvent.setup();
      const error: ApiError = {
        code: 'UPDATE_ERROR',
        message: 'Failed to update',
        description: 'Validation failed',
      };

      mockUseUpdateUserType.mockReturnValue({
        updateUserType: mockUpdateUserType,
        error,
        reset: mockResetUpdateError,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      expect(screen.getByText('Failed to update')).toBeInTheDocument();
      expect(screen.getByText('Validation failed')).toBeInTheDocument();
    });

    it('allows adding enum values in edit mode', async () => {
      const user = userEvent.setup();
      const userTypeWithString: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        schema: {
          status: {
            type: 'string',
            required: true,
            enum: [],
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithString,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const enumInput = screen.getByPlaceholderText(/add value and press enter/i);
      await user.type(enumInput, 'ACTIVE');

      const addButton = screen.getByRole('button', {name: /^add$/i});
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByText('ACTIVE')).toBeInTheDocument();
      });
    });

    it('allows removing enum values in edit mode', async () => {
      const user = userEvent.setup();
      const userTypeWithEnum: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        schema: {
          status: {
            type: 'string',
            required: true,
            enum: ['ACTIVE', 'INACTIVE'],
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithEnum,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const activeChip = screen.getByText('ACTIVE').closest('.MuiChip-root');
      const deleteButton = within(activeChip as HTMLElement).getByTestId('CancelIcon');

      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.queryByText('ACTIVE')).not.toBeInTheDocument();
        expect(screen.getByText('INACTIVE')).toBeInTheDocument();
      });
    });

    it('allows editing regex pattern', async () => {
      const user = userEvent.setup();
      const userTypeWithString: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        schema: {
          username: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithString,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const regexInput = screen.getByPlaceholderText(/e.g., \^/i);
      await user.click(regexInput);
      await user.paste('^[a-z]+$');

      expect(regexInput).toHaveValue('^[a-z]+$');
    });

    it('property name field is disabled in edit mode', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const propertyNameInputs = screen.getAllByPlaceholderText(/e.g., email, age, address/i);
      propertyNameInputs.forEach((input) => {
        expect(input).toBeDisabled();
      });
    });
  });

  describe('Delete Functionality', () => {
    it('opens delete confirmation dialog', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      const deleteButton = screen.getByRole('button', {name: /delete/i});
      await user.click(deleteButton);

      await waitFor(() => {
        expect(screen.getByText('Delete User Type')).toBeInTheDocument();
        expect(screen.getByText(/are you sure you want to delete this user type/i)).toBeInTheDocument();
      });
    });

    it('closes delete dialog when cancel is clicked', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /delete/i}));

      await waitFor(() => {
        expect(screen.getByText('Delete User Type')).toBeInTheDocument();
      });

      const cancelButton = screen.getAllByRole('button', {name: /cancel/i})[0];
      await user.click(cancelButton);

      await waitFor(() => {
        expect(screen.queryByText('Delete User Type')).not.toBeInTheDocument();
      });
    });

    it('deletes user type and navigates back', async () => {
      const user = userEvent.setup();
      mockDeleteUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /delete/i}));

      await waitFor(() => {
        expect(screen.getByText('Delete User Type')).toBeInTheDocument();
      });

      const dialogDeleteButton = screen.getByRole('button', {name: /^delete$/i});
      await user.click(dialogDeleteButton);

      await waitFor(() => {
        expect(mockDeleteUserType).toHaveBeenCalledWith('schema-123');
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });

    it('displays deleting state', async () => {
      const user = userEvent.setup();
      mockUseDeleteUserType.mockReturnValue({
        deleteUserType: mockDeleteUserType,
        loading: true,
        error: null,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /delete/i}));

      await waitFor(() => {
        expect(screen.getByText('Deleting...')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: /deleting.../i})).toBeDisabled();
      });
    });

    it('displays delete error in dialog', async () => {
      const user = userEvent.setup();
      const error: ApiError = {
        code: 'DELETE_ERROR',
        message: 'Cannot delete user type',
        description: 'User type is in use',
      };

      mockUseDeleteUserType.mockReturnValue({
        deleteUserType: mockDeleteUserType,
        loading: false,
        error,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /delete/i}));

      await waitFor(() => {
        expect(screen.getByText('Cannot delete user type')).toBeInTheDocument();
        expect(screen.getByText('User type is in use')).toBeInTheDocument();
      });
    });
  });
});
