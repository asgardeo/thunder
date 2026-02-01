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
import {render, screen, waitFor, within, userEvent} from '@thunder/test-utils';
import type {OrganizationUnitListParams} from '@/features/organization-units/types/organization-units';
import ViewUserTypePage from '../ViewUserTypePage';
import type {ApiUserSchema, ApiError, UpdateUserSchemaRequest} from '../../types/user-types';

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
  updateUserType: (id: string, data: UpdateUserSchemaRequest) => Promise<void>;
  error: ApiError | null;
  reset: () => void;
}

interface UseDeleteUserTypeReturn {
  deleteUserType: (id: string) => Promise<void>;
  loading: boolean;
  error: ApiError | null;
}

interface UseGetOrganizationUnitsReturn {
  data: {
    totalResults: number;
    startIndex: number;
    count: number;
    organizationUnits: {
      id: string;
      name: string;
      handle: string;
      description?: string | null;
      parent?: string | null;
    }[];
  } | null;
  isLoading: boolean;
  error: ApiError | null;
  refetch: (newParams?: OrganizationUnitListParams) => Promise<void>;
}

const mockUseGetUserType = vi.fn<(id?: string) => UseGetUserTypeReturn>();
const mockUseUpdateUserType = vi.fn<() => UseUpdateUserTypeReturn>();
const mockUseDeleteUserType = vi.fn<() => UseDeleteUserTypeReturn>();
const mockUseGetOrganizationUnits = vi.fn<() => UseGetOrganizationUnitsReturn>();
const mockRefetchOrganizationUnits = vi.fn();

vi.mock('../../api/useGetUserType', () => ({
  default: (id?: string) => mockUseGetUserType(id),
}));

vi.mock('../../api/useUpdateUserType', () => ({
  default: () => mockUseUpdateUserType(),
}));

vi.mock('../../api/useDeleteUserType', () => ({
  default: () => mockUseDeleteUserType(),
}));

vi.mock('../../../organization-units/api/useGetOrganizationUnits', () => ({
  default: () => mockUseGetOrganizationUnits(),
}));

const getOrganizationUnitSelect = () => screen.getAllByRole('combobox')[0];
const getPropertyTypeSelect = (index = 0) => screen.getAllByRole('combobox')[index + 1];
const getPropertyTypeSelects = () => screen.getAllByRole('combobox').slice(1);

describe('ViewUserTypePage', () => {
  const mockUserType: ApiUserSchema = {
    id: 'schema-123',
    name: 'Employee Schema',
    ouId: 'root-ou',
    allowSelfRegistration: false,
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

  const mockOrganizationUnitsResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    organizationUnits: [
      {id: 'root-ou', name: 'Root Organization', handle: 'root', description: null, parent: null},
      {id: 'child-ou', name: 'Child Organization', handle: 'child', description: null, parent: 'root-ou'},
    ],
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
    mockUseGetOrganizationUnits.mockReturnValue({
      data: mockOrganizationUnitsResponse,
      isLoading: false,
      error: null,
      refetch: mockRefetchOrganizationUnits,
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
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
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

    it('falls back to organization unit id when lookup data is missing', () => {
      mockUseGetOrganizationUnits.mockReturnValue({
        data: {...mockOrganizationUnitsResponse, organizationUnits: []},
        isLoading: false,
        error: null,
        refetch: mockRefetchOrganizationUnits,
      });

      render(<ViewUserTypePage />);

      expect(screen.getByText('root-ou')).toBeInTheDocument();
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
        ouId: 'root-ou',
        allowSelfRegistration: false,
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
        ouId: 'root-ou',
        allowSelfRegistration: false,
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

      const typeSelects = getPropertyTypeSelects();
      await user.click(typeSelects[0]);

      // Click on Number option instead to avoid duplicate "String" text
      const numberOption = await screen.findByRole('option', {name: 'Number'});
      await user.click(numberOption);

      await waitFor(() => {
        expect(typeSelects[0]).toHaveTextContent('Number');
      });
    });

    it('allows selecting organization unit by name', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const ouSelect = getOrganizationUnitSelect();
      await user.click(ouSelect);

      const childOption = await screen.findByText('Child Organization');
      await user.click(childOption);

      await waitFor(() => {
        expect(ouSelect).toHaveTextContent('Child Organization');
      });

      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            ouId: 'child-ou',
          }),
        );
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
          ouId: 'root-ou',
          allowSelfRegistration: false,
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

      await waitFor(() => {
        expect(screen.getByText('Saving...')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: /saving.../i})).toBeDisabled();
      });
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
        ouId: 'root-ou',
        allowSelfRegistration: false,
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
        ouId: 'root-ou',
        allowSelfRegistration: false,
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
        ouId: 'root-ou',
        allowSelfRegistration: false,
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

    it('closes dialog on delete error', async () => {
      const user = userEvent.setup();
      mockDeleteUserType.mockRejectedValue(new Error('Delete failed'));

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /delete/i}));

      await waitFor(() => {
        expect(screen.getByText('Delete User Type')).toBeInTheDocument();
      });

      const dialogDeleteButton = screen.getByRole('button', {name: /^delete$/i});
      await user.click(dialogDeleteButton);

      await waitFor(() => {
        expect(mockDeleteUserType).toHaveBeenCalled();
        expect(screen.queryByText('Delete User Type')).not.toBeInTheDocument();
      });
    });
  });

  describe('Advanced Edit Mode Features', () => {
    it('allows adding enum value with Enter key', async () => {
      const user = userEvent.setup();
      const userTypeWithString: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
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
      await user.type(enumInput, 'ACTIVE{Enter}');

      await waitFor(() => {
        expect(screen.getByText('ACTIVE')).toBeInTheDocument();
      });
    });

    it('allows toggling unique checkbox for number type', async () => {
      const user = userEvent.setup();
      const userTypeWithNumber: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          employeeId: {
            type: 'number',
            required: true,
            unique: false,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithNumber,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const uniqueCheckbox = screen.getByRole('checkbox', {name: /unique/i});
      await user.click(uniqueCheckbox);

      await waitFor(() => {
        expect(uniqueCheckbox).toBeChecked();
      });
    });

    it('saves schema with array type properties', async () => {
      const user = userEvent.setup();
      const userTypeWithArray: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          tags: {
            type: 'array',
            required: false,
            items: {type: 'string'},
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithArray,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            schema: expect.objectContaining({
              tags: expect.objectContaining({
                type: 'array',
                items: {type: 'string'},
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }),
        );
      });
    });

    it('saves schema with object type properties', async () => {
      const user = userEvent.setup();
      const userTypeWithObject: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          address: {
            type: 'object',
            required: false,
            properties: {},
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithObject,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            schema: expect.objectContaining({
              address: expect.objectContaining({
                type: 'object',
                properties: {},
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }),
        );
      });
    });

    it('handles save error and keeps form in edit mode', async () => {
      const user = userEvent.setup();
      mockUpdateUserType.mockRejectedValue(new Error('Save failed'));

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const nameInput = screen.getByPlaceholderText(/user type name/i);
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Schema');

      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalled();
      });

      // Form should still be in edit mode
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
      });
    });

    it('saves schema with enum values for string type', async () => {
      const user = userEvent.setup();
      const userTypeWithString: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          status: {
            type: 'string',
            required: true,
            enum: ['ACTIVE', 'INACTIVE'],
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithString,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            schema: expect.objectContaining({
              status: expect.objectContaining({
                type: 'string',
                enum: ['ACTIVE', 'INACTIVE'],
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }),
        );
      });
    });

    it('saves schema with regex pattern for string type', async () => {
      const user = userEvent.setup();
      const userTypeWithRegex: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          username: {
            type: 'string',
            required: true,
            regex: '^[a-zA-Z]+$',
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithRegex,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            schema: expect.objectContaining({
              username: expect.objectContaining({
                type: 'string',
                regex: '^[a-zA-Z]+$',
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }),
        );
      });
    });

    it('resets enum and regex when changing type from string to boolean', async () => {
      const user = userEvent.setup();
      const userTypeWithString: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          status: {
            type: 'string',
            required: true,
            unique: true,
            enum: ['ACTIVE', 'INACTIVE'],
            regex: '^[A-Z]+$',
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

      const typeSelect = getPropertyTypeSelect();
      await user.click(typeSelect);

      const booleanOption = await screen.findByRole('option', {name: 'Boolean'});
      await user.click(booleanOption);

      await waitFor(() => {
        expect(typeSelect).toHaveTextContent('Boolean');
        // Unique checkbox should not be visible for boolean type
        expect(screen.queryByRole('checkbox', {name: /unique/i})).not.toBeInTheDocument();
      });
    });

    it('does not add empty enum value', async () => {
      const user = userEvent.setup();
      const userTypeWithString: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
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

      const addButton = screen.getByRole('button', {name: /^add$/i});
      await user.click(addButton);

      // No chip should be added
      await waitFor(() => {
        expect(screen.queryByRole('button', {name: /cancel/i})).toBeInTheDocument();
      });
    });
  });

  describe('Navigation Error Handling', () => {
    it('handles navigation error from user type not found state', async () => {
      const user = userEvent.setup();
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error: null,
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

  describe('Schema Conversion and Save', () => {
    it('saves schema with unique flag for number type', async () => {
      const user = userEvent.setup();
      const userTypeWithUniqueNumber: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
        schema: {
          employeeId: {
            type: 'number',
            required: true,
            unique: true,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithUniqueNumber,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            schema: expect.objectContaining({
              employeeId: expect.objectContaining({
                type: 'number',
                unique: true,
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }),
        );
      });
    });
  });

  describe('Validation and Snackbar', () => {
    it('shows validation error when saving with empty organization unit', async () => {
      const user = userEvent.setup();
      const userTypeWithEmptyOu: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: '',
        allowSelfRegistration: false,
        schema: {
          email: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithEmptyOu,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(screen.getByText('Please provide an organization unit ID')).toBeInTheDocument();
      });

      expect(mockUpdateUserType).not.toHaveBeenCalled();
    });

    it('closes validation error snackbar when close button is clicked', async () => {
      const user = userEvent.setup();
      const userTypeWithEmptyOu: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: '',
        allowSelfRegistration: false,
        schema: {
          email: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithEmptyOu,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));
      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(screen.getByText('Please provide an organization unit ID')).toBeInTheDocument();
      });

      const closeButton = screen.getByLabelText(/close/i);
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('Please provide an organization unit ID')).not.toBeInTheDocument();
      });
    });

    it('allows toggling allowSelfRegistration checkbox in edit mode', async () => {
      const user = userEvent.setup();
      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const selfRegCheckbox = screen.getByRole('checkbox', {name: /allow self registration/i});
      expect(selfRegCheckbox).not.toBeChecked();

      await user.click(selfRegCheckbox);

      await waitFor(() => {
        expect(selfRegCheckbox).toBeChecked();
      });
    });

    it('displays organization unit placeholder when value is empty in edit mode', async () => {
      const user = userEvent.setup();
      const userTypeWithEmptyOu: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: '',
        allowSelfRegistration: false,
        schema: {
          email: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithEmptyOu,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const ouSelect = getOrganizationUnitSelect();
      expect(ouSelect).toHaveTextContent('Select an organization unit');
    });

    it('displays organization unit id in select when unit is not found in lookup', async () => {
      const user = userEvent.setup();
      const userTypeWithUnknownOu: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'unknown-ou-id',
        allowSelfRegistration: false,
        schema: {
          email: {
            type: 'string',
            required: true,
          },
        },
      };

      mockUseGetUserType.mockReturnValue({
        data: userTypeWithUnknownOu,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const ouSelect = getOrganizationUnitSelect();
      expect(ouSelect).toHaveTextContent('unknown-ou-id');
    });

    it('shows loading state in organization unit dropdown', async () => {
      const user = userEvent.setup();

      mockUseGetOrganizationUnits.mockReturnValue({
        data: null,
        isLoading: true,
        error: null,
        refetch: mockRefetchOrganizationUnits,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const ouSelect = getOrganizationUnitSelect();
      await user.click(ouSelect);

      await waitFor(() => {
        expect(screen.getByText('Loading...')).toBeInTheDocument();
      });
    });

    it('shows error message in organization unit dropdown when fetch fails', async () => {
      const user = userEvent.setup();
      const orgError: ApiError = {
        code: 'ORG_ERROR',
        message: 'Failed to load organization units',
        description: 'Error',
      };

      mockUseGetOrganizationUnits.mockReturnValue({
        data: null,
        isLoading: false,
        error: orgError,
        refetch: mockRefetchOrganizationUnits,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const ouSelect = getOrganizationUnitSelect();
      await user.click(ouSelect);

      await waitFor(() => {
        expect(screen.getByText('Failed to load organization units')).toBeInTheDocument();
      });
    });

    it('shows no organization units message when list is empty', async () => {
      const user = userEvent.setup();

      mockUseGetOrganizationUnits.mockReturnValue({
        data: {...mockOrganizationUnitsResponse, organizationUnits: []},
        isLoading: false,
        error: null,
        refetch: mockRefetchOrganizationUnits,
      });

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      const ouSelect = getOrganizationUnitSelect();
      await user.click(ouSelect);

      await waitFor(() => {
        expect(screen.getByText('No organization units available')).toBeInTheDocument();
      });
    });
  });

  describe('Schema Property Handling with Enum Type', () => {
    it('saves schema with enum type converted to string', async () => {
      const user = userEvent.setup();
      const userTypeWithEnum: ApiUserSchema = {
        id: 'schema-123',
        name: 'Test Schema',
        ouId: 'root-ou',
        allowSelfRegistration: false,
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

      mockUpdateUserType.mockResolvedValue(undefined);

      render(<ViewUserTypePage />);

      await user.click(screen.getByRole('button', {name: /edit/i}));

      // Add a new enum value
      const enumInput = screen.getByPlaceholderText(/add value and press enter/i);
      await user.type(enumInput, 'PENDING');

      const addButton = screen.getByRole('button', {name: /^add$/i});
      await user.click(addButton);

      await waitFor(() => {
        expect(screen.getByText('PENDING')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', {name: /save changes/i}));

      await waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            schema: expect.objectContaining({
              status: expect.objectContaining({
                type: 'string',
                enum: ['ACTIVE', 'INACTIVE', 'PENDING'],
              }) as Record<string, unknown>,
            }) as Record<string, unknown>,
          }),
        );
      });
    });
  });
});
