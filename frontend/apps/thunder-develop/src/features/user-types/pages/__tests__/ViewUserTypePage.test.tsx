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

import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import type {OrganizationUnitListParams} from '@/features/organization-units/models/requests';
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

const getOrganizationUnitSelect = async () => (page.getByRole('combobox').all())[0];
const getPropertyTypeSelect = async (index = 0) => (page.getByRole('combobox').all())[index + 1];
const getPropertyTypeSelects = async () => (page.getByRole('combobox').all()).slice(1);

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
    it('displays loading state', async () => {
      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: true,
        error: null,
        refetch: mockRefetch,
      });

      await render(<ViewUserTypePage />);

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('displays error state with error message', async () => {
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

      await render(<ViewUserTypePage />);

      await expect.element(page.getByText('Failed to load user type')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /back to user types/i})).toBeInTheDocument();
    });

    it('displays warning when user type not found', async () => {
      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      await render(<ViewUserTypePage />);

      await expect.element(page.getByText('User type not found')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /back to user types/i})).toBeInTheDocument();
    });

    it('navigates back from error state', async () => {
      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error: {code: 'ERROR', message: 'Error', description: ''},
        refetch: mockRefetch,
      });

      await render(<ViewUserTypePage />);

      const backButton = page.getByRole('button', {name: /back to user types/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });
  });

  describe('View Mode', () => {
    it('renders user type details in view mode', async () => {
      await render(<ViewUserTypePage />);

      await expect.element(page.getByText('Manage User Type')).toBeInTheDocument();
      await expect.element(page.getByText('View and manage user type information')).toBeInTheDocument();
      await expect.element(page.getByText('schema-123')).toBeInTheDocument();
      await expect.element(page.getByText('Employee Schema')).toBeInTheDocument();
      await expect.element(page.getByText('Root Organization')).toBeInTheDocument();
    });

    it('displays schema properties in table', async () => {
      await render(<ViewUserTypePage />);

      await expect.element(page.getByText('Property Name')).toBeInTheDocument();
      await expect.element(page.getByText('Type')).toBeInTheDocument();
      await expect.element(page.getByText('Required')).toBeInTheDocument();
      await expect.element(page.getByText('Unique')).toBeInTheDocument();

      await expect.element(page.getByText('email')).toBeInTheDocument();
      await expect.element(page.getByText('age')).toBeInTheDocument();
      await expect.element(page.getByText('isActive')).toBeInTheDocument();
    });

    it('falls back to organization unit id when lookup data is missing', async () => {
      mockUseGetOrganizationUnits.mockReturnValue({
        data: {...mockOrganizationUnitsResponse, organizationUnits: []},
        isLoading: false,
        error: null,
        refetch: mockRefetchOrganizationUnits,
      });

      await render(<ViewUserTypePage />);

      await expect.element(page.getByText('root-ou')).toBeInTheDocument();
    });

    it('displays edit and delete buttons in view mode', async () => {
      await render(<ViewUserTypePage />);

      await expect.element(page.getByRole('button', {name: /edit/i})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /delete/i})).toBeInTheDocument();
    });

    it('navigates back when back button is clicked', async () => {
      await render(<ViewUserTypePage />);

      const backButton = page.getByRole('button', {name: /^back$/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });

    it('displays enum values in view mode', async () => {
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

      await render(<ViewUserTypePage />);

      await expect.element(page.getByText(/ACTIVE, INACTIVE, PENDING/i)).toBeInTheDocument();
    });

    it('displays regex pattern in view mode', async () => {
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

      await render(<ViewUserTypePage />);

      await expect.element(page.getByText('Pattern:')).toBeInTheDocument();
      await expect.element(page.getByText('^[a-zA-Z0-9]+$')).toBeInTheDocument();
    });
  });

  describe('Edit Mode', () => {
    it('enters edit mode when edit button is clicked', async () => {
      await render(<ViewUserTypePage />);

      const editButton = page.getByRole('button', {name: /edit/i});
      await userEvent.click(editButton);

      await expect.element(page.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /cancel/i})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /edit/i})).not.toBeInTheDocument();
    });

    it('displays editable form fields in edit mode', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByPlaceholder(/user type name/i)).toBeInTheDocument();
    });

    it('allows editing user type name', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const nameInput = page.getByPlaceholder(/user type name/i);
      await userEvent.clear(nameInput);
      await userEvent.fill(nameInput, 'Updated Schema Name');

      await expect.element(nameInput).toHaveValue('Updated Schema Name');
    });

    it('allows toggling required checkbox', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByRole('button', {name: /save changes/i})).toBeInTheDocument();

      const requiredCheckboxes = page.getByRole('checkbox', {name: /required/i}).all();
      const firstCheckbox = requiredCheckboxes[0];

      const isInitiallyChecked = firstCheckbox.element().getAttribute('checked') !== null;
      await userEvent.click(firstCheckbox);

      if (isInitiallyChecked) {
        await expect.element(firstCheckbox).not.toBeChecked();
      } else {
        await expect.element(firstCheckbox).toBeChecked();
      }
    });

    it('allows changing property type', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const typeSelects = await getPropertyTypeSelects();
      await userEvent.click(typeSelects[0]);

      // Click on Number option instead to avoid duplicate "String" text
      const numberOption = page.getByRole('option', {name: 'Number'});
      await userEvent.click(numberOption);

      await expect.element(typeSelects[0]).toHaveTextContent('Number');
    });

    it('allows selecting organization unit by name', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const ouSelect = await getOrganizationUnitSelect();
      await userEvent.click(ouSelect);

      const childOption = page.getByText('Child Organization');
      await userEvent.click(childOption);

      await expect.element(ouSelect).toHaveTextContent('Child Organization');

      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalledWith(
          'schema-123',
          expect.objectContaining({
            ouId: 'child-ou',
          }),
        );
      });
    });

    it('cancels edit mode and reverts changes', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const nameInput = page.getByPlaceholder(/user type name/i);
      await userEvent.clear(nameInput);
      await userEvent.fill(nameInput, 'Changed Name');

      await userEvent.click(page.getByRole('button', {name: /cancel/i}));

      await expect.element(page.getByText('Employee Schema')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder(/user type name/i)).not.toBeInTheDocument();
      expect(mockResetUpdateError).toHaveBeenCalled();
    });

    it('saves changes successfully', async () => {
      mockUpdateUserType.mockResolvedValue(undefined);

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const nameInput = page.getByPlaceholder(/user type name/i);
      await userEvent.clear(nameInput);
      await userEvent.fill(nameInput, 'Updated Schema');

      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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
      // Use a promise that never resolves so the saving state persists
      mockUpdateUserType.mockImplementation(() => new Promise(() => {}));

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await expect.element(page.getByText('Saving...')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /saving.../i})).toBeDisabled();
    });

    it('displays update error', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      await expect.element(page.getByText('Failed to update')).toBeInTheDocument();
      await expect.element(page.getByText('Validation failed')).toBeInTheDocument();
    });

    it('allows adding enum values in edit mode', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const enumInput = page.getByPlaceholder(/add value and press enter/i);
      await userEvent.fill(enumInput, 'ACTIVE');

      const addButton = page.getByRole('button', {name: /^add$/i});
      await userEvent.click(addButton);

      await expect.element(page.getByText('ACTIVE')).toBeInTheDocument();
    });

    it('allows removing enum values in edit mode', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const activeChip = page.getByText('ACTIVE').element().closest('.MuiChip-root');
      const deleteButton = (activeChip as HTMLElement).querySelector('[data-testid="CancelIcon"]')!;

      await userEvent.click(deleteButton as HTMLElement);

      await expect.element(page.getByText('ACTIVE')).not.toBeInTheDocument();
      await expect.element(page.getByText('INACTIVE')).toBeInTheDocument();
    });

    it('allows editing regex pattern', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const regexInput = page.getByPlaceholder(/e.g., \^/i);
      await userEvent.click(regexInput);
      await userEvent.fill(regexInput, '^[a-z]+$');

      await expect.element(regexInput).toHaveValue('^[a-z]+$');
    });

    it('property name field is disabled in edit mode', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const propertyNameInputs = page.getByPlaceholder(/e.g., email, age, address/i).all();
      // eslint-disable-next-line no-restricted-syntax, no-await-in-loop
      for (const input of propertyNameInputs) {
        // eslint-disable-next-line no-await-in-loop
        await expect.element(input).toBeDisabled();
      }
    });
  });

  describe('Delete Functionality', () => {
    it('opens delete confirmation dialog', async () => {
      await render(<ViewUserTypePage />);

      const deleteButton = page.getByRole('button', {name: /delete/i});
      await userEvent.click(deleteButton);

      await expect.element(page.getByText('Delete User Type')).toBeInTheDocument();
      await expect.element(page.getByText(/are you sure you want to delete this user type/i)).toBeInTheDocument();
    });

    it('closes delete dialog when cancel is clicked', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /delete/i}));

      await expect.element(page.getByText('Delete User Type')).toBeInTheDocument();

      const cancelButtons = page.getByRole('button', {name: /cancel/i}).all();
      await userEvent.click(cancelButtons[0]);

      await expect.element(page.getByText('Delete User Type')).not.toBeInTheDocument();
    });

    it('deletes user type and navigates back', async () => {
      mockDeleteUserType.mockResolvedValue(undefined);

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /delete/i}));

      await expect.element(page.getByText('Delete User Type')).toBeInTheDocument();

      const dialogDeleteButton = page.getByRole('button', {name: /^delete$/i});
      await userEvent.click(dialogDeleteButton);

      await vi.waitFor(() => {
        expect(mockDeleteUserType).toHaveBeenCalledWith('schema-123');
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });

    it('displays deleting state', async () => {
      mockUseDeleteUserType.mockReturnValue({
        deleteUserType: mockDeleteUserType,
        loading: true,
        error: null,
      });

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /delete/i}));

      await expect.element(page.getByText('Deleting...')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /deleting.../i})).toBeDisabled();
    });

    it('displays delete error in dialog', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /delete/i}));

      await expect.element(page.getByText('Cannot delete user type')).toBeInTheDocument();
      await expect.element(page.getByText('User type is in use')).toBeInTheDocument();
    });

    it('closes dialog on delete error', async () => {
      mockDeleteUserType.mockRejectedValue(new Error('Delete failed'));

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /delete/i}));

      await expect.element(page.getByText('Delete User Type')).toBeInTheDocument();

      const dialogDeleteButton = page.getByRole('button', {name: /^delete$/i});
      await userEvent.click(dialogDeleteButton);

      await vi.waitFor(() => {
        expect(mockDeleteUserType).toHaveBeenCalled();
      });
      await expect.element(page.getByText('Delete User Type')).not.toBeInTheDocument();
    });
  });

  describe('Advanced Edit Mode Features', () => {
    it('allows adding enum value with Enter key', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const enumInput = page.getByPlaceholder(/add value and press enter/i);
      await userEvent.fill(enumInput, 'ACTIVE');
      await userEvent.keyboard('{Enter}');

      await expect.element(page.getByText('ACTIVE')).toBeInTheDocument();
    });

    it('allows toggling unique checkbox for number type', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const uniqueCheckbox = page.getByRole('checkbox', {name: /unique/i});
      await userEvent.click(uniqueCheckbox);

      await expect.element(uniqueCheckbox).toBeChecked();
    });

    it('saves schema with array type properties', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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
      mockUpdateUserType.mockRejectedValue(new Error('Save failed'));

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const nameInput = page.getByPlaceholder(/user type name/i);
      await userEvent.clear(nameInput);
      await userEvent.fill(nameInput, 'Updated Schema');

      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
        expect(mockUpdateUserType).toHaveBeenCalled();
      });

      // Form should still be in edit mode
      await expect.element(page.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
    });

    it('saves schema with enum values for string type', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const typeSelect = await getPropertyTypeSelect();
      await userEvent.click(typeSelect);

      const booleanOption = page.getByRole('option', {name: 'Boolean'});
      await userEvent.click(booleanOption);

      await expect.element(typeSelect).toHaveTextContent('Boolean');
      // Unique checkbox should not be visible for boolean type
      await expect.element(page.getByRole('checkbox', {name: /unique/i})).not.toBeInTheDocument();
    });

    it('does not add empty enum value', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const addButton = page.getByRole('button', {name: /^add$/i});
      await userEvent.click(addButton);

      // No chip should be added
      await expect.element(page.getByRole('button', {name: /cancel/i})).toBeInTheDocument();
    });
  });

  describe('Navigation Error Handling', () => {
    it('handles navigation error from user type not found state', async () => {
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      mockUseGetUserType.mockReturnValue({
        data: null,
        loading: false,
        error: null,
        refetch: mockRefetch,
      });

      await render(<ViewUserTypePage />);

      const backButton = page.getByRole('button', {name: /back to user types/i});
      await userEvent.click(backButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/user-types');
      });
    });
  });

  describe('Schema Conversion and Save', () => {
    it('saves schema with unique flag for number type', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await expect.element(page.getByText('Please provide an organization unit ID')).toBeInTheDocument();

      expect(mockUpdateUserType).not.toHaveBeenCalled();
    });

    it('closes validation error snackbar when close button is clicked', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));
      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await expect.element(page.getByText('Please provide an organization unit ID')).toBeInTheDocument();

      const closeButton = page.getByLabelText(/close/i);
      await userEvent.click(closeButton);

      await expect.element(page.getByText('Please provide an organization unit ID')).not.toBeInTheDocument();
    });

    it('allows toggling allowSelfRegistration checkbox in edit mode', async () => {
      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const selfRegCheckbox = page.getByRole('checkbox', {name: /allow self registration/i});
      await expect.element(selfRegCheckbox).not.toBeChecked();

      await userEvent.click(selfRegCheckbox);

      await expect.element(selfRegCheckbox).toBeChecked();
    });

    it('displays organization unit placeholder when value is empty in edit mode', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const ouSelect = await getOrganizationUnitSelect();
      await expect.element(ouSelect).toHaveTextContent('Select an organization unit');
    });

    it('displays organization unit id in select when unit is not found in lookup', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const ouSelect = await getOrganizationUnitSelect();
      await expect.element(ouSelect).toHaveTextContent('unknown-ou-id');
    });

    it('shows loading state in organization unit dropdown', async () => {
      mockUseGetOrganizationUnits.mockReturnValue({
        data: null,
        isLoading: true,
        error: null,
        refetch: mockRefetchOrganizationUnits,
      });

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const ouSelect = await getOrganizationUnitSelect();
      await userEvent.click(ouSelect);

      await expect.element(page.getByText('Loading...')).toBeInTheDocument();
    });

    it('shows error message in organization unit dropdown when fetch fails', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const ouSelect = await getOrganizationUnitSelect();
      await userEvent.click(ouSelect);

      await expect.element(page.getByText('Failed to load organization units')).toBeInTheDocument();
    });

    it('shows no organization units message when list is empty', async () => {
      mockUseGetOrganizationUnits.mockReturnValue({
        data: {...mockOrganizationUnitsResponse, organizationUnits: []},
        isLoading: false,
        error: null,
        refetch: mockRefetchOrganizationUnits,
      });

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      const ouSelect = await getOrganizationUnitSelect();
      await userEvent.click(ouSelect);

      await expect.element(page.getByText('No organization units available')).toBeInTheDocument();
    });
  });

  describe('Schema Property Handling with Enum Type', () => {
    it('saves schema with enum type converted to string', async () => {
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

      await render(<ViewUserTypePage />);

      await userEvent.click(page.getByRole('button', {name: /edit/i}));

      // Add a new enum value
      const enumInput = page.getByPlaceholder(/add value and press enter/i);
      await userEvent.fill(enumInput, 'PENDING');

      const addButton = page.getByRole('button', {name: /^add$/i});
      await userEvent.click(addButton);

      await expect.element(page.getByText('PENDING')).toBeInTheDocument();

      await userEvent.click(page.getByRole('button', {name: /save changes/i}));

      await vi.waitFor(() => {
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
